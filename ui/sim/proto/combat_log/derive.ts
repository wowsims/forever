import { ResourceType } from '@generated/proto/spell';

import { bucket, getEnumValues, stringComparator } from '../../utils/collections';
import { sum } from '../../utils/math';
import type { ActionId } from '../action_id';
import { computeActionIdAsString } from './parse';
import {
	AuraLog,
	AuraStacksLog,
	AuraUptimeLog,
	AutoDelayLog,
	BaseLog,
	CastBeganLog,
	CastCancelledLog,
	CastCompletedLog,
	CastLog,
	CombatLog,
	DamageLog,
	DpsLog,
	Entity,
	isAura,
	isAuraStacks,
	isAutoDelay,
	isCastBegan,
	isCastCancelled,
	isCastCompleted,
	isDamage,
	isResource,
	ResourceGroupLog,
	ResourceLog,
	ThreatLogGroup,
} from './types';

const DPS_WINDOW = 15; // Window over which to calculate DPS.

function groupDuplicateTimestamps<T extends BaseLog>(logs: Array<T>): Array<Array<T>> {
	const grouped: Array<Array<T>> = [];
	let curGroup: Array<T> = [];

	logs.forEach(log => {
		if (curGroup.length == 0 || log.timestamp == curGroup[0].timestamp) {
			curGroup.push(log);
		} else {
			grouped.push(curGroup);
			curGroup = [log];
		}
	});
	if (curGroup.length > 0) {
		grouped.push(curGroup);
	}

	return grouped;
}

export function buildDpsLogs(damageLogs: Array<DamageLog>): Array<DpsLog> {
	const groupedDamageLogs = groupDuplicateTimestamps(damageLogs);

	let curDamageLogs: Array<DamageLog> = [];
	let curDamageTotal = 0;

	return groupedDamageLogs.map(ddLogGroup => {
		ddLogGroup.forEach(ddLog => {
			curDamageLogs.push(ddLog);
			curDamageTotal += ddLog.amount;
		});

		const newStartIdx = curDamageLogs.findIndex(curLog => {
			const inWindow = curLog.timestamp > ddLogGroup[0].timestamp - DPS_WINDOW;
			if (!inWindow) {
				curDamageTotal -= curLog.amount;
			}
			return inWindow;
		});
		if (newStartIdx == -1) {
			curDamageLogs = [];
		} else {
			curDamageLogs = curDamageLogs.slice(newStartIdx);
		}

		const dps = curDamageTotal / DPS_WINDOW;
		if (isNaN(dps)) {
			console.warn('NaN dps!');
		}

		const dpsLog: DpsLog = {
			raw: '',
			logIndex: ddLogGroup[0].logIndex,
			timestamp: ddLogGroup[0].timestamp,
			source: ddLogGroup[0].source,
			target: null,
			actionId: null,
			actionIdAsString: null,
			spellSchool: ddLogGroup[0].spellSchool,
			threat: 0,
			activeAuras: [],
			kind: 'dps',
			dps,
			damageLogs: ddLogGroup,
		};
		return dpsLog;
	});
}

export function buildThreatGroups(logs: Array<CombatLog>): Array<ThreatLogGroup> {
	const groupedLogs = groupDuplicateTimestamps(logs.filter(log => log.threat != 0));
	let curThreat = 0;
	return groupedLogs.map(logGroup => {
		const newThreat = sum(logGroup.map(log => log.threat));
		const threatLog: ThreatLogGroup = {
			raw: '',
			logIndex: logGroup[0].logIndex,
			timestamp: logGroup[0].timestamp,
			source: logGroup[0].source,
			target: logGroup[0].target,
			actionId: null,
			actionIdAsString: null,
			spellSchool: logGroup[0].spellSchool,
			threat: newThreat,
			activeAuras: [],
			kind: 'threat-group',
			threatBefore: curThreat,
			threatAfter: curThreat + newThreat,
			logs: logGroup,
		};

		curThreat += newThreat;
		return threatLog;
	});
}

const RESOURCE_TYPES = (getEnumValues(ResourceType) as Array<ResourceType>).filter(val => val != ResourceType.ResourceTypeNone);

export function buildResourceGroups(logs: Array<CombatLog>): Record<ResourceType, Array<ResourceGroupLog>> {
	// Bucket by resource type in one pass rather than filtering the whole log array once per
	// resource type.
	const byResourceType = new Map<ResourceType, Array<ResourceLog>>();
	for (const log of logs) {
		if (!isResource(log)) continue;
		const existing = byResourceType.get(log.resourceType);
		if (existing) {
			existing.push(log);
		} else {
			byResourceType.set(log.resourceType, [log]);
		}
	}

	const maxResource = (resourceLogs: Array<ResourceLog>) => {
		let max = 0;
		resourceLogs.forEach(l => {
			if (l.total > max) max = l.total;
		});
		return max;
	};

	const results: Partial<Record<ResourceType, Array<ResourceGroupLog>>> = {};
	RESOURCE_TYPES.forEach(resourceType => {
		const resourceLogs = byResourceType.get(resourceType);
		// Skip grouping entirely for the ~15 resource types a given fight never touches, rather
		// than running an empty bucket through groupDuplicateTimestamps().map() for nothing.
		if (!resourceLogs) {
			results[resourceType] = [];
			return;
		}
		results[resourceType] = groupDuplicateTimestamps(resourceLogs).map(logGroup => {
			const groupLog: ResourceGroupLog = {
				raw: '',
				logIndex: logGroup[0].logIndex,
				timestamp: logGroup[0].timestamp,
				source: logGroup[0].source,
				target: logGroup[0].target,
				actionId: null,
				actionIdAsString: null,
				spellSchool: logGroup[0].spellSchool,
				threat: 0,
				activeAuras: [],
				kind: 'resource-group',
				resourceType,
				valueBefore: logGroup[0].valueBefore,
				valueAfter: logGroup[logGroup.length - 1].valueAfter,
				maxValue: maxResource(logGroup),
				logs: logGroup,
			};
			return groupLog;
		});
	});

	return results as Record<ResourceType, Array<ResourceGroupLog>>;
}

type UnmatchedGained = { gained: AuraLog; stacks: Array<AuraStacksLog> };

export function buildAuraUptimes(logs: Array<CombatLog>, entity: Entity, encounterDuration: number): Array<AuraUptimeLog> {
	// openEntries is append-only, in log order; openByKey buckets its indices by equalsIgnoringTag identity so a fade/stacks event
	// finds its match in O(1) instead of rescanning every still-open aura. Within a bucket an
	// exact-tag match wins over the earliest open instance: the same spell id can be open under
	// two tags at once (Steady Focus and its counter share 53224), which an
	// `equals(x) || equalsIgnoringTag(x)` predicate would collapse into the tag-blind half.
	const openEntries: Array<UnmatchedGained | null> = [];
	const openByKey = new Map<string, Array<number>>();
	const uptimeLogs: Array<AuraUptimeLog> = [];

	const keyFor = (actionId: ActionId) => actionId.equalityKeyIgnoringTag();

	const pushOpen = (entry: UnmatchedGained) => {
		const index = openEntries.push(entry) - 1;
		const key = keyFor(entry.gained.actionId!);
		const indices = openByKey.get(key);
		if (indices) {
			indices.push(index);
		} else {
			openByKey.set(key, [index]);
		}
	};
	const findOpen = (actionId: ActionId): { indices: Array<number>; pos: number } | null => {
		const indices = openByKey.get(keyFor(actionId));
		if (!indices || indices.length == 0) return null;
		const exact = indices.findIndex(index => openEntries[index]!.gained.actionId!.equals(actionId));
		return { indices, pos: exact >= 0 ? exact : 0 };
	};
	// Mutates the matched entry in place without removing it - a stacks-change log doesn't close
	// the aura.
	const peekOpen = (actionId: ActionId): UnmatchedGained | null => {
		const found = findOpen(actionId);
		return found ? openEntries[found.indices[found.pos]] : null;
	};
	const removeOpen = (actionId: ActionId): UnmatchedGained | null => {
		const found = findOpen(actionId);
		if (!found) return null;
		const index = found.indices.splice(found.pos, 1)[0];
		const entry = openEntries[index];
		openEntries[index] = null;
		return entry;
	};

	logs.forEach(log => {
		if (!log.source || !log.source.equals(entity)) {
			return;
		}

		if (isAuraStacks(log)) {
			if (log.newStacks <= 0) {
				return;
			}
			const matched = peekOpen(log.actionId!);
			if (!matched) {
				console.warn('Unmatched aura stacks change log: ' + log.actionId!.name);
				return;
			}
			matched.stacks.push(log);
			return;
		}

		if (!isAura(log)) {
			return;
		}

		if (log.isGained) {
			pushOpen({ gained: log, stacks: [] });
			return;
		}

		const matched = removeOpen(log.actionId!);
		if (!matched) {
			console.warn('Unmatched aura faded log: ' + log.actionId!.name);
			return;
		}
		const gainedLog = matched.gained;

		uptimeLogs.push({
			raw: log.raw,
			logIndex: gainedLog.logIndex,
			timestamp: gainedLog.timestamp,
			source: log.source,
			target: log.target,
			actionId: gainedLog.actionId,
			actionIdAsString: gainedLog.actionIdAsString,
			spellSchool: log.spellSchool,
			threat: gainedLog.threat,
			activeAuras: [],
			kind: 'aura-uptime',
			gainedAt: gainedLog.timestamp,
			fadedAt: log.timestamp,
			stacksChange: matched.stacks,
		});

		if (log.isRefreshed) {
			pushOpen({ gained: log, stacks: [] });
		}
	});

	// Auras active at the end won't have a faded log, so need to add them separately.
	for (const entry of openEntries) {
		if (!entry) continue;
		const gainedLog = entry.gained;
		uptimeLogs.push({
			raw: gainedLog.raw,
			logIndex: gainedLog.logIndex,
			timestamp: gainedLog.timestamp,
			source: gainedLog.source,
			target: gainedLog.target,
			actionId: gainedLog.actionId,
			actionIdAsString: gainedLog.actionIdAsString,
			spellSchool: gainedLog.spellSchool,
			threat: gainedLog.threat,
			activeAuras: [],
			kind: 'aura-uptime',
			gainedAt: gainedLog.timestamp,
			fadedAt: encounterDuration,
			stacksChange: entry.stacks,
		});
	}

	uptimeLogs.sort((a, b) => a.gainedAt - b.gainedAt);
	return uptimeLogs;
}

// Populates the activeAuras field for all logs using the provided auras.
export function populateActiveAuras(logs: Array<CombatLog>, auraLogs: Array<AuraUptimeLog>): void {
	// `curAuras` is kept sorted by name as auras come and go, rather than being filtered,
	// copied and re-sorted for every single log. Insertion goes after any equal name so
	// the result matches what a stable sort of the arrival order produced before.
	const curAuras: Array<AuraUptimeLog> = [];
	let auraLogsIndex = 0;
	let changed = true;
	// Shared by every log until the active set changes (contract invariant #2) - mutate this
	// array and every earlier log still pointing at it silently corrupts too.
	let activeAuras: Array<AuraUptimeLog> = [];

	for (const log of logs) {
		while (auraLogsIndex < auraLogs.length && auraLogs[auraLogsIndex].gainedAt <= log.timestamp) {
			const gained = auraLogs[auraLogsIndex];
			let insertAt = curAuras.length;
			while (insertAt > 0 && stringComparator(curAuras[insertAt - 1].actionId!.name, gained.actionId!.name) > 0) {
				insertAt--;
			}
			curAuras.splice(insertAt, 0, gained);
			auraLogsIndex++;
			changed = true;
		}
		for (let i = curAuras.length - 1; i >= 0; i--) {
			if (curAuras[i].fadedAt <= log.timestamp) {
				curAuras.splice(i, 1);
				changed = true;
			}
		}

		if (changed) {
			activeAuras = curAuras.slice();
			changed = false;
		}
		log.activeAuras = activeAuras;
	}
}

function buildCastLog(
	castBeganLog: CastBeganLog,
	castCompletedLog: CastCompletedLog | null,
	castCancelledLog: CastCancelledLog | null,
	autoDelayLog: AutoDelayLog | null,
	damageDealtLogs: Array<DamageLog>,
): CastLog {
	const actionId = castCompletedLog?.actionId || castCancelledLog?.actionId || castBeganLog.actionId; // Use completed log because of arcane blast
	const spellSchool = castCompletedLog?.spellSchool || castCancelledLog?.spellSchool || castBeganLog.spellSchool;
	const threat = castCompletedLog?.threat || castCancelledLog?.threat || castBeganLog.threat;

	let castTime = castBeganLog.castTime;
	// effectiveTime stays the began line's: it is the GCD time the tooltip and DPET report, not
	// the cast's measured length.
	const effectiveTime = castBeganLog.effectiveTime;
	if (castCompletedLog) {
		castTime = castCompletedLog.timestamp - castBeganLog.timestamp;
	}
	const cancelTime = castCancelledLog?.cancelTime ?? 0;

	let travelTime = 0;
	if (castCompletedLog && damageDealtLogs.length >= 1 && castCompletedLog.timestamp < damageDealtLogs[0].timestamp && !damageDealtLogs[0].tick) {
		travelTime = damageDealtLogs[0].timestamp - castCompletedLog.timestamp;
	}

	return {
		raw: castBeganLog.raw,
		logIndex: castBeganLog.logIndex,
		timestamp: castBeganLog.timestamp,
		source: castBeganLog.source,
		target: castBeganLog.target,
		actionId,
		actionIdAsString: computeActionIdAsString(actionId),
		spellSchool,
		threat,
		activeAuras: [],
		kind: 'cast',
		castTime,
		gcd: castBeganLog.gcd,
		effectiveTime,
		travelTime,
		cancelTime,
		castBeganLog,
		castCancelledLog,
		castCompletedLog,
		autoDelayLog,
		delay: autoDelayLog?.delay ?? 0,
		delayText: autoDelayLog?.delayText ?? '',
		readyAtText: autoDelayLog?.readyAtTooltip ?? '',
		damageDealtLogs,
	};
}

// Log timestamps have two decimals, so two travel times of one flight can read up to 0.02 sec apart.
const TRAVEL_TIME_TOLERANCE = 0.025;

// A spell still in flight when its next cast completes lands in that next cast's window. Arcane
// Missiles under Missile Barrage fires every 0.5 sec, and at 20 yards each missile flies 1 sec, so
// the first two missiles of a channel show no hit and the last shows three. The hits that share a
// timestamp are one landing, and here the k-th landing goes to the k-th completed cast.
//
// Pairing by order is used only where the windows fail, when one window holds landings at two
// timestamps. It holds only for a steady flight time, so this returns null unless every cast
// completes at its own timestamp and every landing comes the same time after its cast. That keeps
// the windows for two casts at one timestamp (Thorns on two attackers) and for one action from
// several sources (the melee of several targets). It is also null when a cast is left without a
// landing it would have made before the log ends.
function pairHitsInFlight(castsCompleted: Array<CastCompletedLog>, damageLogs: Array<DamageLog>, endTime: number): Array<Array<DamageLog>> | null {
	const landings = groupDuplicateTimestamps(damageLogs.filter(log => !log.tick));
	if (landings.length === 0 || landings.length > castsCompleted.length) return null;

	for (let i = 1; i < castsCompleted.length; i++) {
		if (castsCompleted[i].timestamp <= castsCompleted[i - 1].timestamp) return null;
	}

	// Counted as buildCastLogs' windows take them: a landing goes to the last cast completed at or
	// before it, and the first cast also takes every landing before it.
	let crowdedWindow = false;
	let windowIdx = 0;
	let landingsInWindow = 0;
	for (const landing of landings) {
		const landedAt = landing[0].timestamp;
		while (windowIdx + 1 < castsCompleted.length && landedAt >= castsCompleted[windowIdx + 1].timestamp) {
			windowIdx++;
			landingsInWindow = 0;
		}
		if (++landingsInWindow > 1) crowdedWindow = true;
	}
	if (!crowdedWindow) return null;

	let minTravel = Infinity;
	let maxTravel = -Infinity;
	for (let i = 0; i < landings.length; i++) {
		minTravel = Math.min(minTravel, landings[i][0].timestamp - castsCompleted[i].timestamp);
		maxTravel = Math.max(maxTravel, landings[i][0].timestamp - castsCompleted[i].timestamp);
	}
	if (minTravel <= 0 || maxTravel - minTravel > TRAVEL_TIME_TOLERANCE) return null;

	for (let i = landings.length; i < castsCompleted.length; i++) {
		if (castsCompleted[i].timestamp + minTravel < endTime) return null;
	}

	return castsCompleted.map((_, i) => landings[i] ?? []);
}

export function buildCastLogs(logs: Array<CombatLog>): Array<CastLog> {
	// One classification pass instead of five full scans of the same array.
	const castBeganLogs: Array<CastBeganLog> = [];
	const castCompletedLogs: Array<CastCompletedLog> = [];
	const castCancelledLogs: Array<CastCancelledLog> = [];
	const damageLogs: Array<DamageLog> = [];
	const autoDelayLogs: Array<AutoDelayLog> = [];
	for (const log of logs) {
		if (isCastBegan(log)) castBeganLogs.push(log);
		else if (isCastCompleted(log)) castCompletedLogs.push(log);
		else if (isCastCancelled(log)) castCancelledLogs.push(log);
		else if (isDamage(log)) damageLogs.push(log);
		else if (isAutoDelay(log)) autoDelayLogs.push(log);
	}

	const toBucketKey = (actionId: ActionId) => {
		if (actionId.spellId == 30451) {
			// Arcane Blast is unique because it can finish its cast as a different spell than it
			// started (if stacks drop).
			return actionId.toStringIgnoringTag();
		} else if (actionId.spellId === 27014) {
			// Raptor Strike should be grouped with regular melee swings
			return 'other-3-1';
		} else {
			return actionId.toString();
		}
	};
	const castBeganLogsByAbility = bucket(castBeganLogs, log => toBucketKey(log.actionId!));
	const castCompletedLogsByAbility = bucket(castCompletedLogs, log => toBucketKey(log.actionId!));
	const castCancelledLogsByAbility = bucket(castCancelledLogs, log => toBucketKey(log.actionId!));
	const damageLogsByAbility = bucket(damageLogs, log => toBucketKey(log.actionId!));
	const autoDelayLogsByAbility = bucket(autoDelayLogs, log => toBucketKey(log.actionId!));

	const endTime = logs.length > 0 ? logs[logs.length - 1].timestamp : 0;

	const castLogs: Array<CastLog> = [];
	Object.keys(castBeganLogsByAbility).forEach(bucketKey => {
		const abilityCastsBegan = castBeganLogsByAbility[bucketKey]!;
		const abilityCastsCompleted = castCompletedLogsByAbility[bucketKey];
		const abilityCastsCancelled = castCancelledLogsByAbility[bucketKey];
		const abilityDamage = damageLogsByAbility[bucketKey];
		const abilityAutoDelay = autoDelayLogsByAbility[bucketKey];
		const hitsInFlight = abilityCastsCompleted && abilityDamage ? pairHitsInFlight(abilityCastsCompleted, abilityDamage, endTime) : null;

		let ddIdx = 0;
		let adIdx = 0;
		let castSkipIdx = 0;
		// abilityCastsCancelled is in the same log order as abilityCastsBegan, and each cbLog's
		// cancel window starts where the previous one's ended - a monotonic cursor never needs to
		// revisit a candidate an earlier cbLog already passed.
		let cancelCursor = 0;
		for (let cbIdx = 0; cbIdx < abilityCastsBegan.length; cbIdx++) {
			const cbLog = abilityCastsBegan[cbIdx];

			let cCancelLog: CastCancelledLog | null = null;
			if (abilityCastsCancelled) {
				while (cancelCursor < abilityCastsCancelled.length && abilityCastsCancelled[cancelCursor].timestamp < cbLog.timestamp) {
					cancelCursor++;
				}
				const candidate = abilityCastsCancelled[cancelCursor];
				const nextBegan = abilityCastsBegan[cbIdx + 1];
				if (candidate && (!nextBegan || candidate.timestamp <= nextBegan.timestamp)) {
					cCancelLog = candidate;
					cancelCursor++;
				}
			}

			// The k-th completed log belongs to the k-th began that was not cancelled, so a cancelled
			// began takes none and shifts every later pairing by one. The final began may have no
			// completed log at all.
			let ccLog: CastCompletedLog | null = null;
			let nextCcLog: CastCompletedLog | null = null;
			const ccIdx = cbIdx + castSkipIdx;
			if (cCancelLog) {
				castSkipIdx--;
			} else if (abilityCastsCompleted) {
				if (ccIdx < abilityCastsCompleted.length) ccLog = abilityCastsCompleted[ccIdx];
				if (ccIdx + 1 < abilityCastsCompleted.length) nextCcLog = abilityCastsCompleted[ccIdx + 1];
			}

			// Find all damage dealt logs between the cur and next cast completed logs: only the ticks
			// when the hits are paired in flight.
			const ddLogs: Array<DamageLog> = [];
			while (!cCancelLog && abilityDamage && ddIdx < abilityDamage.length && (!nextCcLog || abilityDamage[ddIdx].timestamp < nextCcLog.timestamp)) {
				if (!hitsInFlight || abilityDamage[ddIdx].tick) ddLogs.push(abilityDamage[ddIdx]);
				ddIdx++;
			}
			if (hitsInFlight && ccLog) {
				ddLogs.push(...hitsInFlight[ccIdx]);
				ddLogs.sort((a, b) => a.timestamp - b.timestamp);
			}
			// At most one auto-delay log per cast; the next one inside this cast's window.
			const nextCbLog = abilityCastsBegan[cbIdx + 1];
			let adLog: AutoDelayLog | null = null;
			if (
				abilityAutoDelay &&
				adIdx < abilityAutoDelay.length &&
				abilityAutoDelay[adIdx].timestamp >= cbLog.timestamp &&
				(!nextCbLog || abilityAutoDelay[adIdx].timestamp < nextCbLog.timestamp)
			) {
				adLog = abilityAutoDelay[adIdx];
				adIdx++;
			}
			castLogs.push(buildCastLog(cbLog, ccLog, cCancelLog, adLog, ddLogs));
		}
	});

	castLogs.sort((a, b) => a.timestamp - b.timestamp);
	return castLogs;
}
