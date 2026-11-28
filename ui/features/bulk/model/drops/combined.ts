import type { Faction, ItemSlot } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import type { BulkRankedResult, BulkResults } from '@sim/bulk/types';
import type { Database } from '@sim/proto/database';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';

import { dropsGap, dropsItemLookup, dropsResultItems, type DropsSourceGroup, dropsSourceGroups } from './results';
import { dropsLevelItemById } from './settings';
import { dropsNamesOf } from './sources';

export interface DropsCombinedSet {
	gear: Gear;
	/** Best gain first. */
	itemIds: number[];
	leftOutIds: number[];
}

// By item id: without Enchanting the sim returns the rings without their enchants, so an untouched ring can differ from the worn one.
const changedSlots = (gear: Gear, current: Gear): ItemSlot[] =>
	gear.getItemSlots().filter(slot => (gear.getEquippedItem(slot)?.id ?? 0) !== (current.getEquippedItem(slot)?.id ?? 0));

const excludesEachOther = (item: UIItem, other: UIItem) =>
	(item.unique && item.id === other.id) || (item.limitCategory !== 0 && item.limitCategory === other.limitCategory);

/**
 * The card's upgrades worn together, null when fewer than two fit. Each upgrade, largest gain first, takes its row with the
 * largest gain whose changed slots no earlier pick took, so a ring that upgrades in both slots can take the other, and whose
 * items no item left in the set excludes by Unique-Equipped or limit category. A row brings every slot it changed, so a
 * two-hander brings its empty off hand and a weapon moved to the other hand comes along. The set starts from the first
 * pick's gear, which keys as the sim will return it.
 */
const dropsCombinedSet = (group: DropsSourceGroup, results: BulkResults): DropsCombinedSet | null => {
	const { currentGear, originalGearResults: current } = results;
	const byGain = (rows: readonly BulkRankedResult[]) => [...rows].sort((a, b) => dropsGap(b, current) - dropsGap(a, current));
	let slots: Partial<Record<ItemSlot, EquippedItem | null>> | null = null;
	const taken = new Set<ItemSlot>();
	const itemIds: number[] = [];
	const leftOutIds: number[] = [];
	const fits = (row: BulkRankedResult, changed: readonly ItemSlot[]) => {
		if (!slots) return true;
		if (changed.some(slot => taken.has(slot))) return false;
		const kept = Object.entries(slots).flatMap(([slot, equipped]) => (equipped && !changed.includes(Number(slot)) ? [equipped.item] : []));
		const incoming = changed.flatMap(slot => row.gear.getEquippedItem(slot)?.item ?? []);
		return !incoming.some(item => kept.some(other => excludesEachOther(item, other)));
	};
	for (const entry of group.items.filter(item => item.upgrade).sort((a, b) => b.gain - a.gain)) {
		const pick = byGain(entry.upgradeRows)
			.map(row => ({ row, changed: changedSlots(row.gear, currentGear) }))
			.find(({ row, changed }) => fits(row, changed));
		if (!pick) {
			leftOutIds.push(entry.itemId);
			continue;
		}
		slots ??= pick.row.gear.asMap();
		for (const slot of pick.changed) {
			slots[slot] = pick.row.gear.getEquippedItem(slot);
			taken.add(slot);
		}
		itemIds.push(entry.itemId);
	}
	return slots && itemIds.length > 1 ? { gear: new Gear(slots), itemIds, leftOutIds } : null;
};

/** By source card key, each card of a Drops run with more than one upgrade and a set of two or more, the cards built as `DropsResults` builds them. */
export const dropsCombinedSets = (db: Database, results: BulkResults, runSources: readonly string[], runFaction: Faction): Map<string, DropsCombinedSet> => {
	const groups = dropsSourceGroups(
		dropsResultItems(results),
		new Set(runSources),
		results.currentGear,
		dropsItemLookup(dropsLevelItemById(db), runFaction),
		dropsNamesOf(db),
	);
	return new Map(
		groups.flatMap(group => {
			const set = group.upgrades > 1 ? dropsCombinedSet(group, results) : null;
			return set ? [[group.key, set] as const] : [];
		}),
	);
};
