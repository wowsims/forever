import { ResourceType } from '@generated/proto/spell';
import i18n from '@i18n/config';
import { ActionId, buffAuraToSpellIdMap, resourceTypeToIcon } from '@sim/proto/action_id';
import type { AuraUptimeLog, CastLog, ResourceGroupLog } from '@sim/proto/combat_log';
import { resourceCssName, resourceNames } from '@sim/proto/names';
import type { UnitMetrics } from '@sim/proto/sim_result';

import {
	actionBucketKey,
	AURA_AS_RESOURCE,
	FILL_RESOURCES,
	groupedAurasByAbility,
	IDS_TO_GROUP_FOR_ROTATION,
	makeRowKey,
	ORDERED_RESOURCE_TYPES,
	PERCENTAGE_RESOURCES,
	resourceBucketKey,
	sortedCastsByAbility,
} from './buckets';
import type {
	AuraItem,
	AuraStackSegment,
	BuildRotationModelParams,
	CastOutcome,
	ContentRow,
	ResourceDisplay,
	ResourceItem,
	RotationModel,
	Row,
	RowItem,
	Section,
	SectionId,
	SectionKind,
} from './types';
import { ROW_HEIGHTS } from './types';

function castOutcome(castLog: CastLog): CastOutcome {
	if (castLog.damageDealtLogs.length === 0) return castLog.castCancelledLog ? 'cancelled' : 'none';
	const ddl = castLog.damageDealtLogs[0];
	if (ddl.resist > 0) return 'partial';
	switch (ddl.outcome) {
		case 'miss':
		case 'dodge':
		case 'parry':
			return 'miss';
		// A blocked crit (`critical-block`) reads as partial rather than crit, the same as a plain
		// block or glance.
		case 'block':
		case 'glance':
		case 'critical-block':
			return 'partial';
		case 'crit':
			return 'crit';
		default:
			return 'hit';
	}
}

function sortAndPrefixMax(items: Array<RowItem>): Array<number> {
	items.sort((a, b) => a.start - b.start);
	const maxRightUpTo: Array<number> = [];
	let max = -Infinity;
	for (const item of items) {
		max = Math.max(max, item.end);
		maxRightUpTo.push(max);
	}
	return maxRightUpTo;
}

function auraItems(auraUptimeLogs: ReadonlyArray<AuraUptimeLog>, sharesRowWithCast: boolean): Array<AuraItem> {
	return auraUptimeLogs.map(aul => {
		const start = aul.gainedAt;
		const width = aul.fadedAt === aul.gainedAt ? 0.001 : aul.fadedAt - aul.gainedAt;
		const stacks: Array<AuraStackSegment> = [];
		aul.stacksChange.forEach((scl, i) => {
			if (scl.timestamp == aul.fadedAt) return;
			stacks.push({
				offset: scl.timestamp - aul.timestamp,
				duration: aul.stacksChange[i + 1] ? aul.stacksChange[i + 1].timestamp - scl.timestamp : aul.fadedAt - scl.timestamp,
				stacks: scl.newStacks,
			});
		});
		return { kind: 'aura', start, end: start + width, stacks, sharesRowWithCast, log: aul };
	});
}

/**
 * Casts of one row that start closer together than this, in seconds, run on as one stack. A cast icon
 * is 24px wide and the default zoom is 100px per second, so each cast inside the window draws over the
 * one before it.
 */
const STACK_WINDOW = 0.21;

function castRowItems(castLogs: ReadonlyArray<CastLog>, mergedAuras: ReadonlyArray<Array<AuraUptimeLog>>, duration: number, showGcd: boolean): Array<RowItem> {
	const items: Array<RowItem> = [];
	let stack: Array<CastLog> = [];
	castLogs.forEach(castLog => {
		const start = castLog.timestamp;
		const previous = stack[stack.length - 1];
		stack = previous && start - previous.timestamp < STACK_WINDOW ? [...stack, castLog] : [castLog];
		const cancelled = !!castLog.castCancelledLog;
		const width = cancelled ? castLog.cancelTime : castLog.castTime + castLog.travelTime;
		const hasTravelTime = !cancelled && castLog.travelTime != 0;
		if (castLog.delay > 0) {
			const delayStart = Math.max(0, start - castLog.delay);
			items.push({ kind: 'delay', start: delayStart, end: start, log: castLog });
		}
		items.push({
			kind: 'cast',
			start,
			end: start + width,
			outcome: castOutcome(castLog),
			cancelled,
			travelStart: hasTravelTime ? castLog.castTime : null,
			travelDuration: hasTravelTime ? castLog.travelTime : null,
			log: castLog,
			stack,
		});
		if (showGcd && !cancelled) {
			const extensionStart = start + castLog.castTime;
			const extensionEnd = Math.min(start + castLog.gcd, duration);
			if (extensionEnd > extensionStart) items.push({ kind: 'gcdExtension', start: extensionStart, end: extensionEnd });
		}
		castLog.damageDealtLogs.filter(ddl => ddl.tick).forEach(ddl => items.push({ kind: 'tick', start: ddl.timestamp, end: ddl.timestamp, log: ddl }));
	});
	mergedAuras.forEach(auraUptimeLogs => items.push(...auraItems(auraUptimeLogs, true)));
	return items;
}

function gcdRowItems(unit: UnitMetrics, duration: number): Array<RowItem> {
	return unit.castLogs
		.filter(castLog => castLog.gcd > 0 && !castLog.castCancelledLog && duration - castLog.timestamp > 0)
		.map(castLog => ({
			kind: 'gcdSegment' as const,
			start: castLog.timestamp,
			end: castLog.timestamp + Math.min(castLog.gcd, duration - castLog.timestamp),
			log: castLog,
		}));
}

// Regen ticks come 100 ms apart, with a little margin for the float timestamps.
const REGEN_STEP = 0.095;

// One group per regen step: a group that starts less than a step after the first group of the current
// step joins it, so a credit between two ticks adds no slice of its own.
function mergeRegenSteps(groups: Array<ResourceGroupLog>): Array<ResourceGroupLog> {
	const steps: Array<ResourceGroupLog> = [];
	for (const group of groups) {
		const step = steps[steps.length - 1];
		if (step && group.timestamp - step.timestamp < REGEN_STEP) {
			steps[steps.length - 1] = {
				...step,
				valueAfter: group.valueAfter,
				maxValue: Math.max(step.maxValue, group.maxValue),
				logs: [...step.logs, ...group.logs],
			};
		} else {
			steps.push(group);
		}
	}
	return steps;
}

function resourceItems(resourceType: ResourceType, logs: Array<ResourceGroupLog>, duration: number): Array<ResourceItem> {
	const display: ResourceDisplay = PERCENTAGE_RESOURCES.includes(resourceType) ? 'percent' : FILL_RESOURCES.includes(resourceType) ? 'fill' : 'number';
	const resourceLogs = display === 'fill' ? mergeRegenSteps(logs) : logs;
	const startValue = (group: ResourceGroupLog): number => (group.maxValue == null ? resourceLogs[0].valueBefore : group.maxValue);

	return resourceLogs.map((group, i) => {
		const start = group.timestamp;
		const width = (resourceLogs[i + 1]?.timestamp ?? duration) - start;
		const base = startValue(group);
		const percent = ((group.valueAfter / base) * 100).toFixed(0);
		return {
			kind: 'resource',
			start,
			end: start + width,
			startValue: base,
			display,
			fillPercent: Number(percent),
			text: display === 'percent' ? `${percent}%` : display === 'number' ? Math.floor(group.valueAfter).toFixed(0) : '',
			log: group,
		};
	});
}

export function buildRotationModel({ player, targets, duration, showGcd }: BuildRotationModelParams): RotationModel {
	const rows: Array<Row> = [];
	const sections: Array<Section> = [];
	const byKey = new Map<string, number>();
	const model: RotationModel = { duration, rows, sections, byKey };
	if (targets.length == 0) return model;

	const addRow = (row: Row) => {
		byKey.set(row.key, rows.length);
		rows.push(row);
	};

	const addContentRow = (section: Section, row: ContentRow) => {
		addRow(row);
		section.rowKeys.push(row.key);
	};

	const addSection = (
		id: SectionId,
		kind: SectionKind,
		label: string,
		separator: boolean,
		header: { label: string; actionId: ActionId | null } | null,
	): Section => {
		const index = sections.length;
		const section: Section = { id, kind, label, separatorKey: null, headerKey: null, rowKeys: [] };
		sections.push(section);
		if (separator) {
			section.separatorKey = `sep:${index}`;
			addRow({ kind: 'separator', key: section.separatorKey, section: id, height: ROW_HEIGHTS.separator });
		}
		if (header) {
			section.headerKey = `header:${id}`;
			addRow({
				kind: 'header',
				key: section.headerKey,
				section: id,
				height: ROW_HEIGHTS.header,
				label: header.label,
				actionId: header.actionId,
			});
		}
		return section;
	};

	const addResourceRow = (section: Section, resourceType: ResourceType, resourceLogs: Array<ResourceGroupLog>) => {
		if (resourceLogs.length == 0) return;

		const label = resourceNames.get(resourceType)!;
		const icon = resourceTypeToIcon[resourceType];
		const cssName = resourceCssName(resourceType);

		const items: Array<RowItem> = resourceItems(resourceType, resourceLogs, duration);
		addContentRow(section, {
			kind: 'resource',
			key: makeRowKey(section.id, 'resource', resourceBucketKey(resourceType)),
			section: section.id,
			height: ROW_HEIGHTS.resource,
			label,
			icon,
			cssName,
			items,
			maxRightUpTo: sortAndPrefixMax(items),
		});
	};

	const addGcdRow = (section: Section, unit: UnitMetrics) => {
		if (!showGcd) return;
		const items = gcdRowItems(unit, duration);
		if (items.length === 0) return;
		addContentRow(section, {
			kind: 'gcd',
			key: makeRowKey(section.id, 'gcd', 'gcd'),
			section: section.id,
			height: ROW_HEIGHTS.gcd,
			label: i18n.t('results_tab.details.timeline.gcd'),
			items,
			maxRightUpTo: sortAndPrefixMax(items),
		});
	};

	const addAuraRow = (section: Section, auraUptimeLogs: Array<AuraUptimeLog>) => {
		const actionId = auraUptimeLogs[0].actionId!;
		const items: Array<RowItem> = auraItems(auraUptimeLogs, false);
		addContentRow(section, {
			kind: 'aura',
			key: makeRowKey(section.id, 'aura', actionId.equalityKey()),
			section: section.id,
			height: ROW_HEIGHTS.aura,
			label: IDS_TO_GROUP_FOR_ROTATION.includes(actionId.spellId) ? actionId.baseName : actionId.name,
			actionId,
			items,
			maxRightUpTo: sortAndPrefixMax(items),
		});
	};

	const addCastRow = (section: Section, castLogs: Array<CastLog>, aurasById: Array<Array<AuraUptimeLog>>) => {
		const actionId = castLogs[0].actionId!;
		const grouped = IDS_TO_GROUP_FOR_ROTATION.includes(actionId.spellId);
		const mergedAuras = aurasById.filter(auraUptimeLogs => {
			const auraActionId = auraUptimeLogs[0].actionId!;
			const mapped = buffAuraToSpellIdMap[auraActionId.spellId] ?? auraActionId;
			return grouped ? actionId.equalityKeyIgnoringTag() === mapped.equalityKeyIgnoringTag() : actionId.equalityKey() === mapped.equalityKey();
		});

		const items = castRowItems(castLogs, mergedAuras, duration, showGcd);
		addContentRow(section, {
			kind: 'cast',
			key: makeRowKey(section.id, 'cast', actionBucketKey(actionId)),
			section: section.id,
			height: ROW_HEIGHTS.cast,
			label: grouped ? actionId.baseName : actionId.name,
			actionId,
			items,
			maxRightUpTo: sortAndPrefixMax(items),
		});
	};

	const playerSection = addSection('', 'player', '', false, null);
	ORDERED_RESOURCE_TYPES.forEach(resourceType => addResourceRow(playerSection, resourceType, player.groupedResourceLogs[resourceType]));

	const buffsById = groupedAurasByAbility(player.auraUptimeLogs);
	const debuffsByTargetById = targets.map(target => groupedAurasByAbility(target.auraUptimeLogs));
	const buffsAndDebuffsById = buffsById.concat(debuffsByTargetById[0]);

	AURA_AS_RESOURCE.forEach(actionId => {
		const auraUptimeLogs = buffsById.find(logs => logs[0].actionId!.equalityKey() === actionId.equalityKey());
		if (auraUptimeLogs) addAuraRow(playerSection, auraUptimeLogs);
	});

	addGcdRow(playerSection, player);

	const playerCastsByAbility = sortedCastsByAbility(player);
	playerCastsByAbility.forEach(castLogs => addCastRow(playerSection, castLogs, buffsAndDebuffsById));

	if (player.pets.length > 0) {
		const playerPets = new Map<string, { pet: UnitMetrics; castsByAbility: Array<Array<CastLog>> }>();
		player.pets.forEach(petsLog => {
			const petCastsByAbility = sortedCastsByAbility(petsLog);
			if (petCastsByAbility.length > 0 && !playerPets.has(petsLog.name)) {
				playerPets.set(petsLog.name, { pet: petsLog, castsByAbility: petCastsByAbility });
			}
		});

		playerPets.forEach(({ pet, castsByAbility }) => {
			const section = addSection(`pet:${pet.name}`, 'pet', pet.name, true, { label: pet.name, actionId: ActionId.fromPetName(pet.name) });
			ORDERED_RESOURCE_TYPES.forEach(resourceType => addResourceRow(section, resourceType, pet.groupedResourceLogs[resourceType]));
			addGcdRow(section, pet);
			castsByAbility.forEach(castLogs => addCastRow(section, castLogs, buffsAndDebuffsById));
		});
	}

	const auraAsResourceKeys = new Set(AURA_AS_RESOURCE.map(auraId => auraId.equalityKey()));
	const playerCastKeys = new Set(playerCastsByAbility.map(casts => casts[0].actionId!.equalityKeyIgnoringTag()));
	const buffsToShow = buffsById.filter(auraUptimeLogs => {
		const actionId = auraUptimeLogs[0].actionId;
		// The aura-as-resource test does not depend on the cast, but it sat inside the per-cast
		// scan, so with no player casts at all it never ran. Kept that way here: this is a
		// performance change, and correcting it would move rows.
		if (!actionId || playerCastKeys.size === 0) return true;
		const mapped = buffAuraToSpellIdMap[actionId.spellId] ?? actionId;
		return !playerCastKeys.has(mapped.equalityKeyIgnoringTag()) && !auraAsResourceKeys.has(actionId.equalityKey());
	});
	if (buffsToShow.length > 0) {
		const section = addSection('buffs', 'buffs', '', true, null);
		buffsToShow.forEach(auraUptimeLogs => addAuraRow(section, auraUptimeLogs));
	}

	targets.forEach(target => {
		const targetCastsByAbility = sortedCastsByAbility(target);
		if (targetCastsByAbility.length > 0) {
			const section = addSection(`target-casts:${target.label}`, 'targetCasts', target.label, true, { label: target.label, actionId: null });
			targetCastsByAbility.forEach(castLogs => addCastRow(section, castLogs, buffsAndDebuffsById));
		}
	});

	debuffsByTargetById.forEach((debuffsToShow, index) => {
		if (debuffsToShow.length > 0) {
			const label = targets?.[index]?.label ?? '';
			const section = addSection(`target-debuffs:${label}`, 'targetDebuffs', label, true, { label, actionId: null });
			debuffsToShow.forEach(auraUptimeLogs => addAuraRow(section, auraUptimeLogs));
		}
	});

	return model;
}
