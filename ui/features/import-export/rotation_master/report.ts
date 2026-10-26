// What the converters say about what they did. Nothing is dropped silently: every part of the original that did not
// carry over exactly is listed here, with how far it is from the original.

import rawNames from './rm_condition_names.json';

/** How faithful a conversion is. exact: the same behaviour. approximate: a close substitute. lossy: a part was left out. unsupported: nothing to convert. */
export type Fidelity = 'exact' | 'approximate' | 'lossy' | 'unsupported';
export const FIDELITIES: ReadonlyArray<Fidelity> = ['exact', 'approximate', 'lossy', 'unsupported'];

export interface Reason {
	/** A stable code (`x-cond-dropped`, `cond-unsupported`, ...); `describeReason` words it. */
	code: string;
	level: Fidelity | 'check';
	params: Record<string, string | number | boolean>;
}

export const worse = (a: Fidelity, b: Fidelity): Fidelity => (FIDELITIES.indexOf(b) > FIDELITIES.indexOf(a) ? b : a);

/** The edition a rotation is for (Rotation Master's word) and which sim repositories read it. */
export type Edition = 'era' | 'tbc' | 'progression' | 'forever';
export type SimRepo = 'classic' | 'sod' | 'tbc' | 'tbc-new' | 'wotlk' | 'cata' | 'mop' | 'forever';
export type ClassKey = 'WARRIOR' | 'PALADIN' | 'HUNTER' | 'ROGUE' | 'PRIEST' | 'DEATHKNIGHT' | 'SHAMAN' | 'MAGE' | 'WARLOCK' | 'MONK' | 'DRUID';

/** What a simulator's APL protocol can express (the field names of its APLValue and APLAction messages). */
export interface SimKinds {
	values: ReadonlySet<string>;
	actions: ReadonlySet<string>;
}

/** One entry of the ladder of ranks of a spell: [spell id, rank]. */
export type RankLadder = ReadonlyArray<readonly [number, number]>;

/** The ladders of every spell that has ranks, by any of its ids. */
export type Ladders = ReadonlyMap<number, RankLadder>;

export const ladderIndex = (ladders: ReadonlyArray<RankLadder>): Ladders => {
	const index = new Map<number, RankLadder>();
	for (const ladder of ladders) for (const [id] of ladder) index.set(id, ladder);
	return index;
};

/**
 * The ladders of a spell list ({id, name, rank}), the way the simulator's database carries it: every name with several
 * different ranks, ids from rank 1 up. A name whose ranks repeat (two classes' spells sharing a name) is left out, never guessed.
 */
export const laddersOf = (rows: Iterable<{ id: number; name?: string; rank?: number }>): RankLadder[] => {
	const byName = new Map<string, Array<{ id: number; rank: number }>>();
	for (const row of rows) {
		if (typeof row.name !== 'string' || !Number.isInteger(row.rank) || (row.rank as number) < 1) continue;
		const list = byName.get(row.name);
		if (list) list.push({ id: row.id, rank: row.rank as number });
		else byName.set(row.name, [{ id: row.id, rank: row.rank as number }]);
	}
	const out: RankLadder[] = [];
	for (const list of byName.values()) {
		if (list.length < 2 || new Set(list.map(r => r.rank)).size !== list.length) continue;
		out.push(list.sort((a, b) => a.rank - b.rank).map(r => [r.id, r.rank] as const));
	}
	return out.sort((a, b) => a[0][0] - b[0][0]);
};

// ---- the words ------------------------------------------------------------------------------------------------------
const WHY: Record<string, string> = {
	'fight-length': 'the fight length, which the game does not know',
	arithmetic: 'arithmetic on values',
	'cast-time': 'cast times, travel times and the global cooldown',
	'proc-tracking': 'trinket and proc tracking',
	'sim-state': "the simulator's internal state",
	'multi-target': 'conditions over every target',
	'spec-logic': 'specialization-specific simulator logic',
	maximums: 'maximum values',
	regen: 'regeneration rates',
	boss: 'boss spell information',
	runes: 'rune details',
	distance: 'distances between units',
	sequences: 'action sequences',
	constant: 'a constant on its own',
	metadata: 'metadata',
	cooldowns: 'cooldown management by the simulator',
	'sim-only': 'a simulator-only action',
	timing: 'waiting and scheduling',
	targets: 'target changes',
	unknown: 'something Rotation Master does not know',
};
const WHYC: Record<string, string> = {
	'no-equivalent': 'the simulator has nothing like it',
	id: 'its spell is not given by id',
	unit: 'its unit cannot be written',
	value: 'its value cannot be written',
	class: 'this class has no such resource',
	incomplete: 'it is incomplete',
	operator: 'its comparison is incomplete',
	depth: 'it is nested too deeply',
	'sim-lacks': 'this simulator does not read it',
};
const NOTE: Record<string, string> = {
	'can-cast': 'can-cast checks more than readiness',
	'aura-known': 'an aura is not a spell the character knows',
	'generic-resource': 'a generic resource has no exact counterpart',
	'rune-count': 'runes recharge differently in the game',
	'target-count': "the simulator's number of targets became the enemies you have threat on",
	'execute-phase': 'the execute phase became a health threshold on the target',
	'friendly-spell': 'the target of a friendly spell is not kept',
	channel: 'channeling options are not kept',
	multidot: 'multi-dot became a single cast guarded by the debuff on the target',
	sequence: 'a sequence became separate steps',
	'sim-combat': 'the simulator is always in combat once the fight starts',
	'sim-target': "the simulator's target is always attackable",
	'power-form': 'the resource depends on the shapeshift form',
	'fight-length': 'the time until the target dies became the fight length',
	'rune-cooldown': 'rune cooldowns differ in the game',
	'swing-timer': 'the swing timer is read differently in the game',
	'autocast-cooldowns': 'Rotation Master shows cooldowns separately, in its Cooldowns list',
	'sim-idle': 'the simulator only chooses an action when you are free to act, so this is never true there',
	'sim-boss': "the simulator's target is a boss, so this is true for a boss or elite target and false for any other",
	'auto-attack': 'Rotation Master suggests Attack or Shoot by itself, unless a Stop step suppresses it',
	'cooldown-first': 'a Rotation Master cooldown, placed first',
};
const str = (v: unknown): string => String(v ?? '');
const whyCondition = (why: unknown): string => WHYC[str(why)] ?? str(why);
const NAMES = (rawNames as { names: Record<string, string> }).names;
/** The name the Rotation Master editor shows for a condition type ("WEAPON_IMBUE" -> "Weapon Enchant Type"). */
export const conditionName = (type: unknown): string => NAMES[str(type)] ?? str(type);

/** English words for one reason (the dialogs show these; the site words the same codes in its own languages). */
export function describeReason(reason: Reason): string {
	const p = reason.params;
	switch (reason.code) {
		// ---- Rotation Master -> APL
		case 'x-rank':
			return `Spell ${str(p.from)} became rank ${str(p.rank)} (${str(p.to)}), the highest the simulator knows.`;
		case 'x-prepull-cond':
			return 'Its other conditions were lost: a pre-pull action carries none.';
		case 'x-prepull-time':
			return 'It runs only out of combat, so it became a pre-pull action; the time before the pull is assumed and can be changed.';
		case 'x-cond-dropped':
			return `The "${conditionName(p.type)}" condition has no equivalent here (${whyCondition(p.why)}) and was left out; the action may fire more often.`;
		case 'x-cond-approx':
			return `The "${conditionName(p.type)}" condition was converted approximately: ${NOTE[str(p.note)] ?? str(p.note)}.`;
		case 'x-cond-never':
			return 'Its condition can never be true in the simulator, so it was not exported.';
		case 'x-cooldown-first':
			return 'It is a Rotation Master cooldown, which runs independently of the rotation; it was placed first in the priority list.';
		case 'x-potion':
			return "This edition's mana potions became the simulator's resource potion, which uses the potion chosen in the simulator's settings.";
		case 'x-item-list':
			return `It names ${str(p.count)} items; each became its own action in order.`;
		case 'x-disabled':
			return 'It is disabled in Rotation Master.';
		case 'x-hidden':
			return 'It is disabled in Rotation Master, so it was imported hidden.';
		case 'x-hidden-lossy':
			return 'A part of it was left out, so it was imported hidden; turn it on once you have checked it.';
		case 'x-unreachable':
			return p.why === 'stop' ? 'It comes after a Stop and can never run.' : 'It is in a branch arm that can never run.';
		case 'x-branch':
			return 'A branch: its condition was added to every step inside.';
		case 'x-stop':
			return 'A Stop: its negated condition was added to every later step.';
		case 'x-action-id':
			return 'Its spell id is not a valid id.';
		case 'x-action-itemset':
			return 'It names an item set, not items; the simulator has no item sets.';
		case 'x-action-pet':
			return 'Pet actions have no counterpart.';
		case 'x-action-assisted':
			return 'Assisted Combat suggestions have no counterpart.';
		case 'x-action-unknown':
			return 'Its action type is unknown.';
		// ---- APL -> Rotation Master
		case 'hidden':
			return 'It is hidden in the simulator.';
		case 'item-bad':
			return 'It is not a valid list item.';
		case 'action-empty':
			return 'It has no action.';
		case 'action-unsupported':
			return `The ${str(p.kind)} action has no equivalent (${WHY[str(p.why)] ?? str(p.why)}).`;
		case 'action-noop':
			return `The ${str(p.kind)} action: ${NOTE[str(p.note)] ?? str(p.note)}.`;
		case 'group-unknown':
			return `The action group ${str(p.name)} does not exist.`;
		case 'group-cycle':
			return `The action group ${str(p.name)} refers to itself or nests too deeply.`;
		case 'action-no-id':
			return 'Its spell has no id.';
		case 'action-other-id':
			return `It casts ${str(p.id)}, which is not a spell or item.`;
		case 'action-target':
			return `Its target (${str(p.unit)}) is not kept; Rotation Master casts at your current target.`;
		case 'action-friendly-spell':
		case 'action-channel':
		case 'action-multidot':
		case 'action-sequence':
			return `${NOTE[reason.code.slice(7)] ?? reason.code}.`;
		case 'cond-unsupported':
			return `A ${str(p.kind)} condition has no equivalent (${WHY[str(p.why)] ?? str(p.why)}) and was left out.`;
		case 'cond-approx':
			return `A ${str(p.kind)} condition was converted approximately: ${NOTE[str(p.note)] ?? str(p.note)}.`;
		case 'cond-empty':
			return 'An empty condition value was left out.';
		case 'cond-missing':
			return 'A comparison with the maximum became a comparison with what is missing from the maximum, which Rotation Master reads as a negative number.';
		case 'action-potion':
			return "The simulator's resource potion became an item step with this edition's mana potions. The simulator uses the potion chosen in its settings.";
		case 'cond-variable':
			return `The variable ${str(p.name)} could not be resolved.`;
		case 'cond-id':
			return `A condition names a spell with no usable id (${str(p.why)}).`;
		case 'cond-unit':
			return `A condition names a unit Rotation Master cannot (${str(p.unit)}).`;
		case 'cond-value':
			return `A condition has a value that cannot be converted (${str(p.what)}).`;
		case 'cond-cmp':
			return `A comparison could not be converted (${str(p.why)}).`;
		case 'cond-unavailable':
			return `"${conditionName(p.type)}" cannot be used on this edition (${str(p.reason)}).`;
		case 'cond-restricted':
			return `"${conditionName(p.type)}" is hidden from addons on this edition (${str(p.restricted)}).`;
		case 'always-false':
			return 'Its condition is always false.';
		case 'too-complex':
			return 'It is too complex to convert.';
		case 'nothing-to-import':
			return 'Nothing in it could be converted.';
		case 'step-limit':
			return `Rotation Master takes at most ${str(p.limit)} steps; the rest were left out.`;
		case 'item-limit':
			return `Too many entries (limit ${str(p.limit)}); the rest were left out.`;
		case 'prepull-timing': {
			const at = p.time ? ' at ' + str(p.time) : '';
			return `It is a pre-pull action${at}. Rotation Master has no timeline, so it became a step that applies out of combat, and the time before the pull is not kept.`;
		}
		default:
			return reason.code;
	}
}

/** The reasons of one item as sentences, identical sentences counted once ("... (3 times)"). */
export function describeReasons(reasons: ReadonlyArray<Reason>): string[] {
	const counts = new Map<string, number>();
	for (const r of reasons) {
		if (r.level === 'exact' || r.level === 'check') continue;
		const text = describeReason(r);
		counts.set(text, (counts.get(text) ?? 0) + 1);
	}
	return [...counts].map(([text, n]) => (n > 1 ? `${text} (${n} times)` : text));
}
