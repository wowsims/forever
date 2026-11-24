import { Faction, HandType, ItemType } from '@generated/proto/common';
import { type DatabaseFilters, type UIItem, UIItem_FactionRestriction, type UIItemSource } from '@generated/proto/ui';
import { type BulkSimItemSlot, getBulkItemSlotFromSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { Player } from '@sim/player/player';
import type { PlayerSpec } from '@sim/player/player_spec';
import type { Gear } from '@sim/proto/gear';
import { canEquipItem, getEligibleItemSlots } from '@sim/proto/items';
import { pushTo } from '@sim/utils/collections';

import { dropCategory, type DropsNames, type DropsSourceLabel, dropsSourceLabel, type DropsSourceNode } from './sources';

export interface DropsItemContext {
	playerSpec: PlayerSpec<any>;
	canDualWield: boolean;
	phase: number;
	faction: Faction;
	wornIds: ReadonlySet<number>;
}

export const wornIdSet = (gear: Gear): Set<number> => new Set(gear.asArray().flatMap(equipped => (equipped ? [equipped.id] : [])));

export const dropsItemContext = (player: Player<any>): DropsItemContext => ({
	playerSpec: player.getPlayerSpec(),
	canDualWield: getBulkPlayerCanDualWield(player),
	phase: player.sim.getPhase(),
	faction: player.getFaction(),
	wornIds: wornIdSet(player.getGear()),
});

export const dropsPhaseItems = (items: readonly UIItem[], phase: number): UIItem[] => items.filter(item => item.phase <= phase);

const opensTo = (side: Faction, faction: Faction) => side === Faction.Unknown || side === faction;

const sourceOpensTo = (source: UIItemSource, faction: Faction): boolean => {
	switch (source.source.oneofKind) {
		case 'quest':
			return opensTo(source.source.quest.faction, faction);
		case 'soldBy':
			return opensTo(source.source.soldBy.faction, faction);
		case 'rep':
			return opensTo(source.source.rep.factionId, faction);
		default:
			return true;
	}
};

/** An item kept for `faction` carries only the sources open to it; an item left with none is dropped. */
export const dropsFactionItems = (items: readonly UIItem[], faction: Faction): UIItem[] => {
	const closed = faction === Faction.Horde ? UIItem_FactionRestriction.ALLIANCE_ONLY : UIItem_FactionRestriction.HORDE_ONLY;
	const kept: UIItem[] = [];
	for (const item of items) {
		if (faction !== Faction.Unknown && item.factionRestriction === closed) continue;
		const sources = item.sources.filter(source => sourceOpensTo(source, faction));
		if (sources.length === 0) continue;
		kept.push(sources.length === item.sources.length ? item : { ...item, sources });
	}
	return kept;
};

// Drops lists what a level-60 character still wears; a dungeon's level is its loot's highest required level.
export const DROPS_MIN_LEVEL = 45;

/** The level of each dungeon zone: the highest `Player.requiredLevelOf` among all its drops. */
export const dropsDungeonLevels = (items: readonly UIItem[]): Map<number, number> => {
	const levels = new Map<number, number>();
	for (const item of items) {
		for (const { source } of item.sources) {
			if (source.oneofKind !== 'drop' || dropCategory(source.drop.difficulty) !== 'drop_dungeon') continue;
			levels.set(source.drop.zoneId, Math.max(levels.get(source.drop.zoneId) ?? 0, Player.requiredLevelOf(item)));
		}
	}
	return levels;
};

/**
 * Every raid drop, the drops of a dungeon at `DROPS_MIN_LEVEL` or above, and any other source only on an item that needs
 * that level; an item left with no source is dropped. `levels` is `dropsDungeonLevels` of the whole database.
 */
export const dropsLevelItems = (items: readonly UIItem[], levels: ReadonlyMap<number, number>): UIItem[] => {
	const kept: UIItem[] = [];
	for (const item of items) {
		const needsLevel = Player.requiredLevelOf(item) >= DROPS_MIN_LEVEL;
		const sources = item.sources.filter(({ source }) => {
			if (source.oneofKind !== 'drop') return needsLevel;
			switch (dropCategory(source.drop.difficulty)) {
				case 'drop_raid':
					return true;
				case 'drop_dungeon':
					return levels.get(source.drop.zoneId)! >= DROPS_MIN_LEVEL;
				default:
					return needsLevel;
			}
		});
		if (sources.length === 0) continue;
		kept.push(sources.length === item.sources.length ? item : { ...item, sources });
	}
	return kept;
};

/** The items that fit a gear slot, that the class can equip and that pass the gear picker's level filters; a random-suffix item is left out. */
export const dropsUsableItems = (
	items: readonly UIItem[],
	playerSpec: PlayerSpec<any>,
	filters: Pick<DatabaseFilters, 'minIlvl' | 'maxIlvl' | 'maxRequiredLevel'>,
): UIItem[] =>
	items.filter(
		item =>
			getEligibleItemSlots(item).length > 0 &&
			canEquipItem(item, playerSpec, undefined) &&
			item.randomSuffixOptions.length === 0 &&
			Player.passesLevelFilters(item, filters),
	);

export type DropsSlotTag = 'mh' | 'oh' | '2h';

export interface DropsItem {
	item: UIItem;
	/** The item's ticked leaf keys. */
	sourceKeys: readonly string[];
	worn: boolean;
	/** Not unticked by the user; a worn item is never sent whatever this says. */
	included: boolean;
	tags: readonly DropsSlotTag[];
	bulkSlot: BulkSimItemSlot;
}

export interface DropsItemList {
	items: readonly DropsItem[];
}

const tagsOf = (item: UIItem, canDualWield: boolean): DropsSlotTag[] => {
	switch (item.type) {
		case ItemType.ItemTypeWeapon:
			switch (item.handType) {
				case HandType.HandTypeTwoHand:
					return ['2h'];
				case HandType.HandTypeOffHand:
					return ['oh'];
				case HandType.HandTypeMainHand:
					return ['mh'];
				default:
					return canDualWield ? ['mh', 'oh'] : ['mh'];
			}
		default:
			return [];
	}
};

const byName = (a: UIItem, b: UIItem) => a.name.localeCompare(b.name);

/** The listed `items` with a ticked source; `sourceKeysOf` gives an item's `dropsItemSourceKeys`. */
export const dropsItemList = (
	items: readonly UIItem[],
	ticked: ReadonlySet<string>,
	excluded: ReadonlySet<number>,
	context: DropsItemContext,
	sourceKeysOf: (item: UIItem) => readonly string[],
): DropsItemList => {
	const listed: DropsItem[] = [];
	for (const item of items) {
		const sourceKeys = sourceKeysOf(item).filter(key => ticked.has(key));
		if (sourceKeys.length === 0) continue;
		listed.push({
			item,
			sourceKeys,
			worn: context.wornIds.has(item.id),
			included: !excluded.has(item.id),
			tags: tagsOf(item, context.canDualWield),
			bulkSlot: getBulkItemSlotFromSlot(getEligibleItemSlots(item)[0], context.canDualWield),
		});
	}
	listed.sort((a, b) => byName(a.item, b.item));
	return { items: listed };
};

export const dropsPickedItemIds = (list: DropsItemList): number[] =>
	list.items
		.filter(entry => entry.included && !entry.worn)
		.map(entry => entry.item.id)
		.sort((a, b) => a - b);

export const dropsItemsBySlot = (list: DropsItemList): Array<{ slot: BulkSimItemSlot; items: DropsItem[] }> => {
	const bySlot = new Map<BulkSimItemSlot, DropsItem[]>();
	for (const entry of list.items) pushTo(bySlot, entry.bulkSlot, entry);
	return [...bySlot].map(([slot, items]) => ({ slot, items })).sort((a, b) => a.slot - b.slot);
};

/** One group per source key the list reaches, in the order `tree`, the Sources tree the list was picked from, lists its leaves. */
export const dropsItemsBySource = (
	list: DropsItemList,
	tree: readonly DropsSourceNode[],
	names: DropsNames,
): Array<{ key: string; label: DropsSourceLabel; items: DropsItem[] }> => {
	const bySource = new Map<string, DropsItem[]>();
	for (const entry of list.items) {
		for (const key of entry.sourceKeys) pushTo(bySource, key, entry);
	}
	return tree
		.flatMap(node => node.leafKeys)
		.filter(key => bySource.has(key))
		.map(key => ({ key, label: dropsSourceLabel(key, names), items: bySource.get(key)! }));
};
