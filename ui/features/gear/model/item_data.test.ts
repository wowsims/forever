import { ItemRandomSuffix, ScalingItemProperties } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { EquippedItem } from '@sim/proto/equipped_item';
import { describe, expect, it, vi } from 'vitest';

import type { GearData } from '../types';
import { itemsTabData } from './item_data';

// `itemsTabData` builds a real `EquippedItem` per row, and that reads the base scaling entry.
const scalingOptions = { 0: ScalingItemProperties.create({ ilvl: 61 }) };
const OWL = ItemRandomSuffix.create({ id: 2, name: 'of the Owl' });
const EAGLE = ItemRandomSuffix.create({ id: 3, name: 'of the Eagle' });

const equipFromList = (item: Item, best: ItemRandomSuffix | null, current: EquippedItem | null = null) => {
	const equipItem = vi.fn();
	const gearData: GearData = { equipItem, getEquippedItem: () => current, subscribe: () => () => undefined };
	const player = { getBestRandomSuffix: () => best } as unknown as Player<any>;
	const [row] = itemsTabData(player, gearData, [item]);
	row.onEquip(row.item);
	return equipItem.mock.calls[0][0] as EquippedItem;
};

describe('itemsTabData', () => {
	it('equips a random suffix item with its best suffix, the one it ranks by', () => {
		const cape = Item.create({ id: 10267, scalingOptions, randomSuffixOptions: [1, 2] });
		expect(equipFromList(cape, OWL).randomSuffix).toEqual(OWL);
	});

	it('replaces the suffix of the item it swaps out', () => {
		const cape = Item.create({ id: 10267, scalingOptions, randomSuffixOptions: [1, 2, 3] });
		const worn = new EquippedItem({ item: Item.create({ id: 10266, scalingOptions, randomSuffixOptions: [3] }), randomSuffix: EAGLE });

		const equipped = equipFromList(cape, OWL, worn);
		expect(equipped.item.id).toBe(10267);
		expect(equipped.randomSuffix).toEqual(OWL);
	});

	it('equips an item without suffixes as it is', () => {
		expect(equipFromList(Item.create({ id: 1, scalingOptions }), null).randomSuffix).toBeNull();
	});
});
