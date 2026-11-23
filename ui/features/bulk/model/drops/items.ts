import { Faction, HandType, ItemType } from '@generated/proto/common';
import { UIItem, UIItem_FactionRestriction, type UIItemSource } from '@generated/proto/ui';
import { type BulkSimItemSlot, getBulkItemSlotFromSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import type { Player } from '@sim/player/player';
import type { PlayerSpec } from '@sim/player/player_spec';
import { canEquipItem, getEligibleItemSlots } from '@sim/proto/items';

import { dropsItemSourceKeys, type DropsNames, type DropsSourceLabel, dropsSourceLabel, dropsSourceTree } from './sources';

export interface DropsItemContext {
	playerSpec: PlayerSpec<any>;
	canDualWield: boolean;
	phase: number;
	faction: Faction;
	wornIds: ReadonlySet<number>;
}

export const dropsItemContext = (player: Player<any>): DropsItemContext => ({
	playerSpec: player.getPlayerSpec(),
	canDualWield: getBulkPlayerCanDualWield(player),
	phase: player.sim.getPhase(),
	faction: player.getFaction(),
	wornIds: new Set(
		player
			.getGear()
			.asArray()
			.flatMap(equipped => (equipped ? [equipped.id] : [])),
	),
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
		kept.push(sources.length === item.sources.length ? item : UIItem.create({ ...item, sources }));
	}
	return kept;
};

export type DropsLeftOutReason = 'class' | 'random_suffix';
export type DropsSlotTag = 'mh' | 'oh' | '2h' | 'r1' | 'r2' | 't1' | 't2';

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
	leftOut: ReadonlyArray<{ item: UIItem; reason: DropsLeftOutReason }>;
}

const tagsOf = (item: UIItem, canDualWield: boolean): DropsSlotTag[] => {
	switch (item.type) {
		case ItemType.ItemTypeFinger:
			return ['r1', 'r2'];
		case ItemType.ItemTypeTrinket:
			return ['t1', 't2'];
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

export const dropsItemList = (
	items: readonly UIItem[],
	ticked: ReadonlySet<string>,
	excluded: ReadonlySet<number>,
	context: DropsItemContext,
	names: DropsNames,
): DropsItemList => {
	const listed: DropsItem[] = [];
	const leftOut: Array<{ item: UIItem; reason: DropsLeftOutReason }> = [];
	for (const item of items) {
		const sourceKeys = dropsItemSourceKeys(item, names).filter(key => ticked.has(key));
		if (sourceKeys.length === 0) continue;
		if (!canEquipItem(item, context.playerSpec, undefined)) {
			leftOut.push({ item, reason: 'class' });
		} else if (item.randomSuffixOptions.length > 0) {
			leftOut.push({ item, reason: 'random_suffix' });
		} else {
			listed.push({
				item,
				sourceKeys,
				worn: context.wornIds.has(item.id),
				included: !excluded.has(item.id),
				tags: tagsOf(item, context.canDualWield),
				bulkSlot: getBulkItemSlotFromSlot(getEligibleItemSlots(item)[0], context.canDualWield),
			});
		}
	}
	listed.sort((a, b) => byName(a.item, b.item));
	leftOut.sort((a, b) => byName(a.item, b.item));
	return { items: listed, leftOut };
};

export const dropsPickedItemIds = (list: DropsItemList): number[] =>
	list.items
		.filter(entry => entry.included && !entry.worn)
		.map(entry => entry.item.id)
		.sort((a, b) => a - b);

export const dropsItemsBySlot = (list: DropsItemList): Array<{ slot: BulkSimItemSlot; items: DropsItem[] }> => {
	const bySlot = new Map<BulkSimItemSlot, DropsItem[]>();
	for (const entry of list.items) bySlot.set(entry.bulkSlot, [...(bySlot.get(entry.bulkSlot) ?? []), entry]);
	return [...bySlot].map(([slot, items]) => ({ slot, items })).sort((a, b) => a.slot - b.slot);
};

const leafKeysInOrder = (list: DropsItemList, names: DropsNames): string[] =>
	dropsSourceTree(
		list.items.map(entry => entry.item),
		names,
	).flatMap(node => node.leafKeys);

export const dropsItemsBySource = (list: DropsItemList, names: DropsNames): Array<{ key: string; label: DropsSourceLabel; items: DropsItem[] }> => {
	const bySource = new Map<string, DropsItem[]>();
	for (const entry of list.items) {
		for (const key of entry.sourceKeys) bySource.set(key, [...(bySource.get(key) ?? []), entry]);
	}
	return leafKeysInOrder(list, names)
		.filter(key => bySource.has(key))
		.map(key => ({ key, label: dropsSourceLabel(key, names), items: bySource.get(key)! }));
};

export const dropsLeftOutCounts = (list: DropsItemList): Record<DropsLeftOutReason, number> => {
	const counts: Record<DropsLeftOutReason, number> = { class: 0, random_suffix: 0 };
	for (const { reason } of list.leftOut) counts[reason]++;
	return counts;
};
