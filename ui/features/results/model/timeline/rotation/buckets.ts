import { APLActionItemSwap_SwapSet } from '@generated/proto/apl';
import { OtherAction } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { ActionId } from '@sim/proto/action_id';
import type { AuraUptimeLog, CastLog } from '@sim/proto/combat_log';
import type { UnitMetrics } from '@sim/proto/sim_result';

import { actionCategory } from './categories';
import type { SectionId } from './types';

export const ROW_KEY_SEPARATOR = '\0';

export const ORDERED_RESOURCE_TYPES: Array<ResourceType> = [
	ResourceType.ResourceTypeHealth,
	ResourceType.ResourceTypeMana,
	ResourceType.ResourceTypeEnergy,
	ResourceType.ResourceTypeRage,
	ResourceType.ResourceTypeComboPoints,
	ResourceType.ResourceTypeFocus,
	ResourceType.ResourceTypeGenericResource,
];

export const PERCENTAGE_RESOURCES: Array<ResourceType> = [ResourceType.ResourceTypeHealth];

// Resources that regenerate in steps too small to read a number each: their row draws the level as a bar.
export const FILL_RESOURCES: Array<ResourceType> = [ResourceType.ResourceTypeMana, ResourceType.ResourceTypeEnergy, ResourceType.ResourceTypeFocus];

export const AURA_AS_RESOURCE: ActionId[] = [
	// Item Swap
	ActionId.fromOtherId(OtherAction.OtherActionItemSwap, APLActionItemSwap_SwapSet.Main),
	ActionId.fromOtherId(OtherAction.OtherActionItemSwap, APLActionItemSwap_SwapSet.Swap1),

	// Vengeance
	ActionId.fromSpellId(84840), // Druid
	ActionId.fromSpellId(84839), // Paladin
	ActionId.fromSpellId(93098), // Warrior
	ActionId.fromSpellId(93099), // Death Knight
	ActionId.fromSpellId(120267), // Monk

	// Monk
	ActionId.fromSpellId(124255, 1), // Stagger
	ActionId.fromSpellId(128938), // Elusive Brew - Stacks
	ActionId.fromSpellId(115308), // Elusive Brew - Active
	ActionId.fromSpellId(1247279), // Tiger Eye Brew - Stacks
	ActionId.fromSpellId(1247275), // Tiger Eye Brew - Active

	// Mage
	ActionId.fromSpellId(148022), // Icicle

	// Warlock
	ActionId.fromItemId(6265), // Soul Shards
];
export const IDS_TO_GROUP_FOR_ROTATION: Array<number> = [
	6774, // Rogue - Slice and Dice
	11300, // Rogue - Eviscerate
	11275, // Rogue - Rupture
	11198, // Rogue - Expose Armor
	5019, // Wand - Shoot (the channel and its shots)
	10438, // Shaman - Searing Totem (the totem and its attacks)
	10587, // Shaman - Magma Totem (the totem and its pulses)
	16387, // Shaman - Flametongue Totem (the totem and its strikes)
];

export function makeRowKey(section: SectionId, kind: string, bucketKey: string): string {
	return [section, kind, bucketKey].join(ROW_KEY_SEPARATOR);
}

export function actionBucketKey(actionId: ActionId): string {
	return IDS_TO_GROUP_FOR_ROTATION.includes(actionId.spellId) ? actionId.equalityKeyIgnoringTag() : actionId.equalityKey();
}

export function resourceBucketKey(resourceType: ResourceType): string {
	return `resource:${resourceType}`;
}

function compareNames(a: string, b: string): number {
	if (a < b) {
		return -1;
	} else if (b < a) {
		return 1;
	} else {
		return 0;
	}
}

function groupBy<T>(values: ReadonlyArray<T>, key: (value: T) => string): Array<Array<T>> {
	const groups = new Map<string, Array<T>>();
	for (const value of values) {
		const groupKey = key(value);
		const group = groups.get(groupKey);
		if (group) {
			group.push(value);
		} else {
			groups.set(groupKey, [value]);
		}
	}
	return [...groups.values()];
}

export function groupedAurasByAbility(auraUptimeLogs: ReadonlyArray<AuraUptimeLog>): Array<Array<AuraUptimeLog>> {
	const aurasByAbility = groupBy(auraUptimeLogs, log => log.actionId!.equalityKey());
	aurasByAbility.sort((a, b) => compareNames(a[0].actionId!.name, b[0].actionId!.name));
	return aurasByAbility;
}

export function sortedCastsByAbility(unit: UnitMetrics): Array<Array<CastLog>> {
	const meleeActionKeys = new Set(unit.getMeleeActions().map(action => action.actionId.equalityKey()));
	const spellActionKeys = new Set(unit.getSpellActions().map(action => action.actionId.equalityKey()));

	const castsByAbility = groupBy(unit.castLogs, log => actionBucketKey(log.actionId!));

	const categories = new Map<Array<CastLog>, number>();
	castsByAbility.forEach(casts => categories.set(casts, actionCategory(casts[0].actionId!, meleeActionKeys, spellActionKeys)));

	castsByAbility.sort((a, b) => {
		const categoryA = categories.get(a)!;
		const categoryB = categories.get(b)!;
		if (categoryA != categoryB) {
			return categoryA - categoryB;
		} else if (a[0].actionId!.anyId() == b[0].actionId!.anyId()) {
			return a[0].actionId!.tag - b[0].actionId!.tag;
		} else {
			return compareNames(a[0].actionId!.name, b[0].actionId!.name);
		}
	});

	return castsByAbility;
}
