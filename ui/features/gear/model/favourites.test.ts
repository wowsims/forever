import { DatabaseFilters, UIEnchant as Enchant } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import { ItemData, ItemListType, SelectorModalTabs } from '../types';
import { applyFavourite, isItemFavourited } from './favourites';

const row = <T extends ItemListType>(id: number, item: T): ItemData<T> =>
	({
		item,
		id,
		name: String(id),
		searchText: String(id),
		nameDescription: '',
		phase: 1,
		quality: 0,
		ignoreEPFilter: false,
		actionId: null,
		onEquip: () => {},
	}) as unknown as ItemData<T>;

const filters = (over: Partial<DatabaseFilters> = {}): DatabaseFilters =>
	({
		favoriteItems: [],
		favoriteEnchants: [],
		favoriteRandomSuffixes: [],
		favoriteReforges: [],
		...over,
	}) as unknown as DatabaseFilters;

const enchant = (effectId: number, type: number) => ({ effectId, type }) as unknown as Enchant;

describe('the list a favourite goes to', () => {
	it('has its own list for enchants, keyed by the unique enchant string', () => {
		const f = filters();
		applyFavourite(f, SelectorModalTabs.Enchants, row(1, enchant(4444, 2)), true);
		expect(f.favoriteEnchants).toEqual(['4444-2']);
		expect(f.favoriteItems).toEqual([]);
	});

	it('keys every other tab by the row id', () => {
		const f = filters();
		applyFavourite(f, SelectorModalTabs.Items, row(4321, {} as never), true);
		expect(f.favoriteItems).toEqual([4321]);
	});
});

describe('isItemFavourited', () => {
	it('reads the tab’s own list', () => {
		expect(isItemFavourited(filters({ favoriteItems: [9] }), SelectorModalTabs.Items, row(9, {} as never))).toBe(true);
		expect(isItemFavourited(filters({ favoriteRandomSuffixes: [9] }), SelectorModalTabs.Items, row(9, {} as never))).toBe(false);
	});

	it('is always false on a tab with no favourites list', () => {
		expect(isItemFavourited(filters({ favoriteItems: [9] }), 'Nothing' as SelectorModalTabs, row(9, {} as never))).toBe(false);
	});
});

describe('applyFavourite', () => {
	it('adds and removes, and reports that it did', () => {
		const f = filters();
		expect(applyFavourite(f, SelectorModalTabs.Items, row(7, {} as never), true)).toBe(true);
		expect(f.favoriteItems).toEqual([7]);
		expect(applyFavourite(f, SelectorModalTabs.Items, row(7, {} as never), false)).toBe(true);
		expect(f.favoriteItems).toEqual([]);
	});

	it('leaves a missing entry alone rather than splicing the last one out', () => {
		const f = filters({ favoriteItems: [1, 2] });
		applyFavourite(f, SelectorModalTabs.Items, row(3, {} as never), false);
		expect(f.favoriteItems).toEqual([1, 2]);
	});

	it('refuses a tab with no favourites list, so the caller can skip its DOM toggle', () => {
		const f = filters();
		expect(applyFavourite(f, 'Nothing' as SelectorModalTabs, row(1, {} as never), true)).toBe(false);
		expect(f.favoriteItems).toEqual([]);
	});

	it('round-trips against isItemFavourited on the enchant string', () => {
		const f = filters();
		const entry = row(1, enchant(55, 1));
		applyFavourite(f, SelectorModalTabs.Enchants, entry, true);
		expect(isItemFavourited(f, SelectorModalTabs.Enchants, entry)).toBe(true);
	});
});
