import { RaidSimResult } from '@generated/proto/api';
import type { SpellSchool } from '@generated/proto/common';

import { ActionId } from '../action_id';
import { stringToResourceType } from '../names';
import { parseAplEntryPath } from './apl_entry_path';
import {
	AuraLog,
	AuraStacksLog,
	AutoDelayLog,
	BaseLog,
	CastBeganLog,
	CastCancelledLog,
	CastCompletedLog,
	CastFailedLog,
	CastPushbackLog,
	DamageLog,
	Entity,
	isCastFailed,
	LogKind,
	MajorCooldownLog,
	Outcome,
	ParsedLog,
	PartialResist,
	PlainLog,
	ResourceLog,
	SpellQueuedLog,
	StatChangeLog,
} from './types';

// Preamble patterns, hoisted so parseAll does not allocate a fresh RegExp per line.
const TIMESTAMP_PREFIX_REGEX = /(\[[0-9.-]+\]) (\[[0-9a-zA-Z\s\-()#]+\])?(.*)/;
const SPELL_SCHOOL_REGEX = / \(SpellSchool: (-?[0-9]+)\)/;
const THREAT_REGEX = / \(Threat: (-?[0-9]+\.[0-9]+)\)/;
const TIMESTAMP_REGEX = /\[(-?[0-9]+\.[0-9]+)\]\w*(.*)/;

// Groups: 1 = timestamp, 2 = entity bracket, 3 = the rest.
export function matchTimestampPrefix(raw: string): RegExpExecArray | null {
	const captureArr = TIMESTAMP_PREFIX_REGEX.exec(raw);
	return captureArr && captureArr.length == 4 ? captureArr : null;
}

export function rawWithoutTimestamp(raw: string): string {
	const captureArr = matchTimestampPrefix(raw);
	if (!captureArr) return raw;
	return `${captureArr[2] ?? ''}${captureArr[3]}`.trim();
}

export function computeActionIdAsString(actionId: ActionId | null): string | null {
	try {
		return actionId?.toString() || null;
	} catch {
		return null;
	}
}

const OUTCOME_BY_TOKEN: Record<string, Outcome> = {
	Miss: 'miss',
	Dodge: 'dodge',
	Parry: 'parry',
	BlockedCrit: 'critical-block',
	Block: 'block',
	Glance: 'glance',
	Crit: 'crit',
	Crush: 'crush',
	Hit: 'hit',
};

const RESIST_BY_TOKEN: Record<string, PartialResist> = {
	'10': 10,
	'20': 20,
	'30': 30,
	'40': 40,
	'50': 50,
	'60': 60,
	'70': 70,
	'80': 80,
	'90': 90,
	'100': 100,
};

// Go's time.Duration.String() can emit nanosecond-precision strings like "29.999999999s" or
// "33.217000001s". Round these to something readable without disturbing clean values.
const reformatGoDurations = (text: string): string =>
	text.replace(/(\d+m)?(\d+\.\d{3,})(m?s)/g, (_match, minute: string | undefined, value: string, unit: string) => {
		const seconds = parseFloat(value);
		if (unit === 'ms') {
			return `${minute ?? ''}${Math.round(seconds)}ms`;
		}
		return `${minute ?? ''}${seconds.toFixed(2)}s`;
	});

// The sim's cast-failure reasons all append ", curTime = <duration>", which is redundant with the
// line's own timestamp prefix.
const stripCurTime = (text: string): string => text.replace(/, curTime = [\dms.h]+$/, '');

function parseDuration(value: string, unit: string): number {
	const n = parseFloat(value);
	return unit == 'ms' ? n / 1000 : n;
}

function durationText(seconds: number): string {
	return seconds >= 1 ? `${seconds.toFixed(2)}s` : `${Math.round(seconds * 1000)}ms`;
}

// One object is allocated per line and each builder finishes it in place. The obvious shape -
// build a params literal, then spread it into a second literal per kind - allocates twice and
// copies nine fields per line, and measured ~30% slower than a single allocation per line.
type Mutable<T> = { -readonly [K in keyof T]: T[K] };

function newLog(
	raw: string,
	logIndex: number,
	timestamp: number,
	source: Entity | null,
	target: Entity | null,
	spellSchool: SpellSchool | null,
	threat: number,
): PendingLog {
	return { kind: 'plain', raw, logIndex, timestamp, source, target, actionId: null, actionIdAsString: null, spellSchool, threat, activeAuras: [] };
}

function buildDamageLog(log: PendingLog, match: RegExpExecArray): DamageLog {
	const out = log as Mutable<DamageLog>;
	out.kind = 'damage';
	out.outcome = OUTCOME_BY_TOKEN[match[3]] ?? 'hit';
	out.effect = match[17] ? (match[17] === 'healing' ? 'healing' : match[17] === 'shielding' ? 'shielding' : 'damage') : null;
	out.amount = match[16] ? parseFloat(match[16]) : 0;
	out.resist = RESIST_BY_TOKEN[match[14]] ?? 0;
	out.tick = Boolean(match[2]) && match[2].includes('tick');
	return out;
}

function buildResourceLog(log: PendingLog, match: RegExpExecArray): ResourceLog {
	const resourceType = stringToResourceType(match[3]);
	const out = log as Mutable<ResourceLog>;
	out.kind = 'resource';
	out.resourceType = resourceType;
	out.valueBefore = parseFloat(match[5]);
	out.valueAfter = parseFloat(match[6]);
	out.isSpend = match[1] == 'Spent';
	out.total = match[8] !== undefined ? parseFloat(match[8]) : 0;
	// Always assigned, including when undefined: the parity dump discovers fields with
	// Object.keys, so the key must exist either way.
	return out;
}

function buildAuraLog(log: PendingLog, match: RegExpExecArray): AuraLog {
	const event = match[1];
	const out = log as Mutable<AuraLog>;
	out.kind = 'aura';
	out.isGained = event == 'gained';
	out.isFaded = event == 'faded';
	out.isRefreshed = event == 'refreshed';
	return out;
}

function buildAuraStacksLog(log: PendingLog, match: RegExpExecArray): AuraStacksLog {
	const out = log as Mutable<AuraStacksLog>;
	out.kind = 'aura-stacks';
	out.oldStacks = parseInt(match[2]);
	out.newStacks = parseInt(match[3]);
	return out;
}

function buildMajorCooldownLog(log: PendingLog): MajorCooldownLog {
	const out = log as Mutable<MajorCooldownLog>;
	out.kind = 'major-cooldown';
	return out;
}

function buildCastBeganLog(log: PendingLog, match: RegExpExecArray): CastBeganLog {
	const out = log as Mutable<CastBeganLog>;
	out.kind = 'cast-began';
	out.manaCost = parseFloat(match[2]);
	out.castTime = parseDuration(match[3], match[4]);
	out.gcd = parseDuration(match[5], match[6]);
	out.effectiveTime = parseDuration(match[7], match[8]);
	out.aplEntryPath = match[9] ? parseAplEntryPath(match[9]) : null;
	return out;
}

// The "was ready at" stamp keeps the sim's own minute prefix in the log text ("1m2.50s") and is
// flattened to seconds for the tooltip.
function buildAutoDelayLog(log: PendingLog, match: RegExpExecArray): AutoDelayLog {
	const delay = parseDuration(match[2], match[3]);
	const readyAtMinute = match[4];
	const readyAtSeconds = parseFloat(match[5]);
	let readyAtTotal = readyAtSeconds;
	if (readyAtMinute && readyAtMinute.endsWith('m')) {
		readyAtTotal = parseFloat(readyAtMinute.slice(0, -1)) * 60 + readyAtSeconds;
	}
	const out = log as Mutable<AutoDelayLog>;
	out.kind = 'auto-delay';
	out.delay = delay;
	out.delayText = durationText(delay);
	out.readyAtLogText = readyAtTotal >= 1 ? `${readyAtMinute}${readyAtSeconds.toFixed(2)}s` : `${Math.round(readyAtTotal * 1000)}ms`;
	out.readyAtTooltip = durationText(readyAtTotal);
	return out;
}

function buildSpellQueuedLog(log: PendingLog, match: RegExpExecArray): SpellQueuedLog {
	const minutePart = match[2];
	const seconds = parseFloat(match[3]);
	let fireAt = seconds;
	if (minutePart && minutePart.endsWith('m')) {
		fireAt = parseFloat(minutePart.slice(0, -1)) * 60 + seconds;
	} else if (match[4] == 'ms') {
		fireAt = seconds / 1000;
	}
	const out = log as Mutable<SpellQueuedLog>;
	out.kind = 'spell-queued';
	out.fireAt = fireAt;
	out.fireAtText = fireAt >= 1 ? `${minutePart}${seconds.toFixed(2)}s` : `${Math.round(fireAt * 1000)}ms`;
	return out;
}

function buildCastFailedLog(log: PendingLog, match: RegExpExecArray): CastFailedLog {
	const out = log as Mutable<CastFailedLog>;
	out.kind = 'cast-failed';
	out.reason = reformatGoDurations(stripCurTime(match[2]));
	return out;
}

function buildCastPushbackLog(log: PendingLog, match: RegExpExecArray): CastPushbackLog {
	const pushback = parseDuration(match[2], match[3]);
	const out = log as Mutable<CastPushbackLog>;
	out.kind = 'cast-pushback';
	out.pushback = pushback;
	out.pushbackText = durationText(pushback);
	out.isChanneling = match[4] == 'channeling';
	return out;
}

function buildCastCancelledLog(log: PendingLog, match: RegExpExecArray): CastCancelledLog {
	let cancelTime = parseFloat(match[2]);
	if (match[3] == 'ms') cancelTime /= 1000;
	const out = log as Mutable<CastCancelledLog>;
	out.kind = 'cast-cancelled';
	out.cancelTime = cancelTime;
	return out;
}

function buildCastCompletedLog(log: PendingLog): CastCompletedLog {
	const out = log as Mutable<CastCompletedLog>;
	out.kind = 'cast-completed';
	return out;
}

function buildStatChangeLog(log: PendingLog, match: RegExpExecArray): StatChangeLog {
	const out = log as Mutable<StatChangeLog>;
	out.kind = 'stat-change';
	out.isGain = match[1] != 'Lost';
	out.stats = match[4];
	return out;
}

function buildPlainLog(log: PendingLog): PlainLog {
	const out = log as Mutable<PlainLog>;
	out.kind = 'plain';
	return out;
}

type LogMatcher = {
	// Literals the pattern requires: a line must contain at least one of them to be worth
	// testing. Skipping a matcher whose guard fails can therefore never skip a match the
	// pattern would have made, which is what keeps this dispatch equivalent to running every
	// pattern in order - it just stops paying for the ones that cannot match. The damage
	// pattern is both the first tried and the most expensive, so guarding it is most of the
	// win. assertGuardsAreNecessary() below checks each literal really is in its pattern.
	guard: Array<string>;
	regex: RegExp;
	// Some patterns match lines they cannot actually build from; those fall through to the
	// next matcher.
	valid?: (match: RegExpExecArray) => boolean;
	// The fragment naming this line's action, e.g. '{SpellID: 48707}'.
	idString: (match: RegExpExecArray) => string;
	build: (log: PendingLog, match: RegExpExecArray) => ParsedLog;
};

// Order is load-bearing: first match wins, and it runs most to least common.
const LOG_MATCHERS: Array<LogMatcher> = [
	{
		guard: ['Miss', 'Hit', 'Crit', 'Crush', 'Glance', 'Dodge', 'Parry', 'Block'],
		// BlockedCrit has to precede Block: the token is matched at one position, so `Block` would
		// take the front of `BlockedCrit` and the optional tail groups would all skip, which reads
		// back as a block with no amount. Crit is safe anywhere in the alternation because a space
		// is required before the token. The resist group is match[14], amount match[16] and effect
		// match[17].
		regex: /] (.*?) (tick )?((Miss)|(Hit)|(BlockedCrit)|(Crit)|(Crush)|(Glance)|(Dodge)|(Parry)|(Block))( \((\d+)% Resist\))?( for (\d+\.\d+) ((damage)|(healing)|(shielding)))?/,
		idString: match => match[1],
		build: (log, match) => buildDamageLog(log, match),
	},
	{
		guard: [' from '],
		regex: /(Gained|Spent) (\d+\.?\d*) (\S.+?\S) from (.*?) \((\d+\.?\d*) --> (\d+\.?\d*)\)( of (\d+\.?\d*) total)?/,
		idString: match => match[4],
		build: (log, match) => buildResourceLog(log, match),
	},
	{
		guard: ['Aura '],
		regex: /Aura ((gained)|(faded)|(refreshed)): (.*)/,
		valid: match => Boolean(match[5]),
		idString: match => match[5],
		build: (log, match) => buildAuraLog(log, match),
	},
	{
		guard: [' stacks: '],
		regex: /(.*) stacks: ([0-9]+) --> ([0-9]+)/,
		valid: match => Boolean(match[1]),
		idString: match => match[1],
		build: (log, match) => buildAuraStacksLog(log, match),
	},
	{
		guard: ['Major cooldown used: '],
		regex: /Major cooldown used: (.*)/,
		idString: match => match[1],
		build: log => buildMajorCooldownLog(log),
	},
	{
		guard: ['Casting '],
		regex: /Casting (.*) \(Cost = (\d+\.?\d*), Cast Time = (\d+\.?\d*)(m?s), GCD = (\d+\.?\d*)(m?s), Effective Time = (\d+\.?\d*)(m?s)\)(?: APL: (.+))?/,
		idString: match => match[1],
		build: (log, match) => buildCastBeganLog(log, match),
	},
	{
		guard: ['Cancelled '],
		regex: /Cancelled (.*) after (\d+\.?\d*)(m?s)/,
		idString: match => match[1],
		build: (log, match) => buildCastCancelledLog(log, match),
	},
	{
		guard: ['Completed cast '],
		regex: /Completed cast (.*)/,
		idString: match => match[1],
		build: log => buildCastCompletedLog(log),
	},
	{
		guard: [' delayed by '],
		regex: /] (.*?) delayed by (\d+\.?\d*)(m?s), was ready at (\d*?m?)(\d+\.?\d*)(m?s)/,
		idString: match => match[1],
		build: (log, match) => buildAutoDelayLog(log, match),
	},
	{
		guard: [' from '],
		regex: /((Gained)|(Lost)) ({.*}) from (fading )?(.*)/,
		idString: match => match[6],
		build: (log, match) => buildStatChangeLog(log, match),
	},
	{
		guard: ['Queueing up '],
		regex: /Queueing up (.*?) to cast at (\d*?m?)(\d+\.?\d*)(m?s)\./,
		idString: match => match[1],
		build: (log, match) => buildSpellQueuedLog(log, match),
	},
	{
		guard: [' failed to cast: '],
		regex: /] (.*?) failed to cast: (.*)/,
		idString: match => match[1],
		build: (log, match) => buildCastFailedLog(log, match),
	},
	{
		guard: [' pushed back '],
		regex: /] (.*?) pushed back (\d+\.?\d*)(m?s) while ((casting)|(channeling))/,
		idString: match => match[1],
		build: (log, match) => buildCastPushbackLog(log, match),
	},
];

type MatchedLine = { matcher: LogMatcher; match: RegExpExecArray };
type PendingLog = Mutable<Omit<BaseLog, 'kind'>> & { kind: LogKind };
type PendingEntry = { log: PendingLog; matcher: LogMatcher | null; match: RegExpExecArray | null; key: string };
type ActionIdRequest = { logString: string; playerIndex: number | undefined };

// The dispatch is only equivalent to trying every pattern in order if each guard literal is
// something its own pattern requires. That is a property of the table, so check it here
// rather than trusting the comment. Necessary, not sufficient: a literal inside an optional
// group would pass this and still not be required, so guards stay hand-reviewed too.
function assertGuardsAreNecessary() {
	for (const matcher of LOG_MATCHERS) {
		for (const literal of matcher.guard) {
			if (!matcher.regex.source.includes(literal)) {
				throw new Error(`Log matcher guard '${literal}' is absent from its pattern ${matcher.regex.source}`);
			}
		}
	}
}

function matchLogLine(raw: string): MatchedLine | null {
	for (const matcher of LOG_MATCHERS) {
		let possible = false;
		for (const literal of matcher.guard) {
			if (raw.includes(literal)) {
				possible = true;
				break;
			}
		}
		if (!possible) continue;
		const match = matcher.regex.exec(raw);
		if (match && (!matcher.valid || matcher.valid(match))) {
			return { matcher, match };
		}
	}
	return null;
}

assertGuardsAreNecessary();

export async function parseAll(result: RaidSimResult): Promise<Array<ParsedLog>> {
	const lines = result.logs.split('\n');
	const pending: Array<PendingEntry> = new Array(lines.length);
	// Resolving an ActionId is the only asynchronous part of parsing, and a fight names
	// on the order of a hundred of them across tens of thousands of lines. Collect the
	// distinct ones while classifying, resolve them in one batch, then build every log
	// synchronously.
	const distinctIds = new Map<string, ActionIdRequest>();

	for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
		let line = lines[lineIndex];
		const raw = line;
		let spellSchool: SpellSchool | null = null;
		let threat = 0;

		const spellSchoolMatch = SPELL_SCHOOL_REGEX.exec(line);
		if (spellSchoolMatch) {
			spellSchool = parseInt(spellSchoolMatch[1]);
		}

		const threatMatch = THREAT_REGEX.exec(line);
		if (threatMatch) {
			threat = parseFloat(threatMatch[1]);
			line = line.substring(0, threatMatch.index);
		}

		const match = TIMESTAMP_REGEX.exec(line);
		if (!match || !match[1]) {
			pending[lineIndex] = { log: newLog(raw, lineIndex, 0, null, null, spellSchool, threat), matcher: null, match: null, key: '' };
			continue;
		}

		const timestamp = parseFloat(match[1]);
		const remainder = match[2];

		const entities = Entity.parseAll(remainder);
		const source = entities[0] || null;
		const target = entities[1] || null;

		const log = newLog(raw, lineIndex, timestamp, source, target, spellSchool, threat);

		const matched = matchLogLine(raw);
		if (!matched) {
			pending[lineIndex] = { log, matcher: null, match: null, key: '' };
			continue;
		}
		// Keep the key: the build pass needs it, and recomputing it there means a second
		// idString() call and a second string per matched line.
		const logString = matched.matcher.idString(matched.match);
		const key = `${source?.index ?? -1}|${logString}`;
		pending[lineIndex] = { log, ...matched, key };
		if (!distinctIds.has(key)) {
			distinctIds.set(key, { logString, playerIndex: source?.index });
		}
	}

	const keys = [...distinctIds.keys()];
	const resolved = await Promise.all(
		keys.map(key => {
			const { logString, playerIndex } = distinctIds.get(key)!;
			return ActionId.fromLogString(logString).fill(playerIndex);
		}),
	);
	// ActionId is immutable (readonly fields, private constructor), so one resolved
	// instance can be shared by every line that names it.
	const actionIds = new Map<string, ActionId>();
	keys.forEach((key, i) => actionIds.set(key, resolved[i]));

	const logs = pending.map(entry => {
		const actionId = entry.matcher ? actionIds.get(entry.key)! : null;
		entry.log.actionId = actionId;
		entry.log.actionIdAsString = computeActionIdAsString(actionId);
		return entry.matcher ? entry.matcher.build(entry.log, entry.match!) : buildPlainLog(entry.log);
	});

	await Promise.all(
		logs.filter(isCastFailed).map(async log => {
			log.reason = (await ActionId.replaceAllInString(log.reason)).replace(/[.!?]$/, '');
		}),
	);

	return logs;
}
