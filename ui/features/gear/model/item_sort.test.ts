import { ItemSlot } from '@generated/proto/common';
import { UIEnchant, UIItem as Item } from '@generated/proto/ui';
import { SortDirection } from '@sim/constants/other';
import { describe, expect, it } from 'vitest';

import { ItemData, SelectorModalTabs } from '../types';
import { defaultSortBy, ItemListSortBy, sortItemIdxs, tabHasLevel, tabSortsByEP } from './item_sort';

const row = (id: number, ilvl: number, scalingIlvl?: number): ItemData<Item> =>
	({
		item: { id, ilvl, scalingOptions: scalingIlvl === undefined ? undefined : { 0: { ilvl: scalingIlvl } } } as unknown as Item,
		id,
		ilvl,
		name: String(id),
		searchText: String(id),
		nameDescription: '',
		phase: 1,
		quality: 0,
		ignoreEPFilter: false,
		actionId: null,
		onEquip: () => {},
	}) as unknown as ItemData<Item>;

const sort = <T extends Item | UIEnchant>(rows: ItemData<T>[], over: Partial<Parameters<typeof sortItemIdxs<T>>[2]> = {}) =>
	sortItemIdxs(
		rows.map((_, i) => i),
		rows,
		{
			sortBy: ItemListSortBy.ILVL,
			sortDirection: SortDirection.DESC,
			computeEP: () => 0,
			isFavourited: () => false,
			...over,
		},
	);

describe('defaultSortBy', () => {
	it.each([
		[SelectorModalTabs.Items, ItemSlot.ItemSlotHead, ItemListSortBy.EP],
		[SelectorModalTabs.RandomSuffixes, ItemSlot.ItemSlotHead, ItemListSortBy.EP],
		[SelectorModalTabs.Items, ItemSlot.ItemSlotTrinket1, ItemListSortBy.ILVL],
		[SelectorModalTabs.Items, ItemSlot.ItemSlotTrinket2, ItemListSortBy.ILVL],
		[SelectorModalTabs.Enchants, ItemSlot.ItemSlotHead, ItemListSortBy.ILVL],
		[SelectorModalTabs.Enchants, ItemSlot.ItemSlotMainHand, ItemListSortBy.ILVL],
	])('sorts the %s tab of slot %i by %i: level on the trinket slots and the enchants tab, whose EP leaves procs out', (label, slot, sortBy) => {
		expect(defaultSortBy(label, slot)).toBe(sortBy);
	});
});

describe('tab columns', () => {
	it.each([
		[SelectorModalTabs.Items, ItemSlot.ItemSlotChest, true],
		[SelectorModalTabs.RandomSuffixes, ItemSlot.ItemSlotChest, true],
		[SelectorModalTabs.Enchants, ItemSlot.ItemSlotChest, false],
		[SelectorModalTabs.Items, ItemSlot.ItemSlotTrinket2, false],
	])('gives the %s tab of slot %i an EP column: %s', (label, slot, hasEP) => {
		expect(tabSortsByEP(label, slot)).toBe(hasEP);
	});

	it.each([
		[SelectorModalTabs.Items, true],
		[SelectorModalTabs.Enchants, true],
		[SelectorModalTabs.RandomSuffixes, false],
	])('gives the %s tab a level column: %s', (label, hasLevel) => {
		expect(tabHasLevel(label)).toBe(hasLevel);
	});
});

describe('sortItemIdxs', () => {
	it('sorts by ilvl descending, preferring the base scaling ilvl', () => {
		expect(sort([row(1, 100), row(2, 500, 480), row(3, 200)])).toEqual([1, 2, 0]);
	});

	it('reverses on ASC', () => {
		expect(sort([row(1, 100), row(2, 300), row(3, 200)], { sortDirection: SortDirection.ASC })).toEqual([0, 2, 1]);
	});

	it('sorts by EP and breaks near-ties on ilvl', () => {
		const rows = [row(1, 100), row(2, 200), row(3, 400)];
		const ep = new Map([
			[1, 10],
			[2, 10.005],
			[3, 5],
		]);
		expect(sort(rows, { sortBy: ItemListSortBy.EP, computeEP: item => ep.get(item.id)! })).toEqual([1, 0, 2]);
	});

	it('reads an enchant by its lvl, on the ilvl sort and on an EP near-tie', () => {
		const enchant = (effectId: number, lvl: number) => ({ item: UIEnchant.create({ effectId, lvl }), id: effectId }) as unknown as ItemData<UIEnchant>;
		const rows = [enchant(1, 20), enchant(2, 60), enchant(3, 40)];

		expect(sort(rows)).toEqual([1, 2, 0]);
		expect(sort(rows, { sortBy: ItemListSortBy.EP, computeEP: item => (item.effectId === 1 ? 10 : 1) })).toEqual([0, 1, 2]);
	});

	it('floats favourites above everything else', () => {
		const rows = [row(1, 100), row(2, 500), row(3, 200)];
		expect(sort(rows, { isFavourited: itemData => itemData.id === 1 })).toEqual([0, 1, 2]);
	});

	it('sorts the array it is given, in place', () => {
		const rows = [row(1, 100), row(2, 500)];
		const idxs = [0, 1];
		expect(
			sortItemIdxs(idxs, rows, { sortBy: ItemListSortBy.ILVL, sortDirection: SortDirection.DESC, computeEP: () => 0, isFavourited: () => false }),
		).toBe(idxs);
		expect(idxs).toEqual([1, 0]);
	});
});
