import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import { SelectorModalTabs } from '../types';
import { getItemIdByItemType } from './item_ids';

describe('getItemIdByItemType', () => {
	it('reads effectId on the enchant tab, not id', () => {
		const enchant = { id: 1, effectId: 4444 } as unknown as Enchant;
		expect(getItemIdByItemType(SelectorModalTabs.Enchants, enchant)).toBe(4444);
	});

	it('reads id on the item and random-suffix tabs', () => {
		expect(getItemIdByItemType(SelectorModalTabs.Items, { id: 5 } as unknown as Item)).toBe(5);
		expect(getItemIdByItemType(SelectorModalTabs.RandomSuffixes, { id: 7 } as never)).toBe(7);
	});

	it('is null on a tab that carries no id', () => {
		expect(getItemIdByItemType('Nothing' as SelectorModalTabs, { id: 5 } as unknown as Item)).toBeNull();
	});

	it('survives a null item on every tab', () => {
		expect(getItemIdByItemType(SelectorModalTabs.Enchants, null)).toBeUndefined();
		expect(getItemIdByItemType(SelectorModalTabs.Items, undefined)).toBeUndefined();
	});
});
