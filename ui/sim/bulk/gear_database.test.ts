import { BulkEnchantSlotOptions, BulkSettings } from '@generated/proto/api';
import { EnchantType, ItemSlot, ItemType } from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import type { Database } from '../proto/database';
import { EquippedItem } from '../proto/equipped_item';
import { Gear } from '../proto/gear';
import { makeBulkItemDatabaseFromSpecs } from './gear_database';

const cloakAgility = UIEnchant.create({ effectId: 849, itemId: 1, name: 'Enchant Cloak - Minor Agility', type: ItemType.ItemTypeBack });
const lesserImpact = UIEnchant.create({
	effectId: 1897,
	itemId: 2,
	name: 'Enchant 2H Weapon - Lesser Impact',
	type: ItemType.ItemTypeWeapon,
	enchantType: EnchantType.EnchantTypeTwoHand,
});
const superiorStriking = UIEnchant.create({ effectId: 1897, itemId: 3, name: 'Enchant Weapon - Superior Striking', type: ItemType.ItemTypeWeapon });

const db = {
	lookupItemSpec: () => null,
	getItemEffectRandPropPoints: () => undefined,
	getEnchantsByEffectId: (effectId: number) => [cloakAgility, lesserImpact, superiorStriking].filter(enchant => enchant.effectId === effectId),
} as unknown as Database;

const shippedRows = (database: { enchants: Array<{ effectId: number; name: string }> }) =>
	database.enchants.map(enchant => [enchant.effectId, enchant.name]).sort();

describe('makeBulkItemDatabaseFromSpecs', () => {
	it('ships the pooled enchants no item carries yet', () => {
		const database = makeBulkItemDatabaseFromSpecs(
			db,
			new Gear({}),
			BulkSettings.create({ enchants: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [849, 999] })] }),
		);

		expect(shippedRows(database)).toEqual([[849, cloakAgility.name]]);
	});

	it('ships every row of a pooled effect id', () => {
		const database = makeBulkItemDatabaseFromSpecs(
			db,
			new Gear({}),
			BulkSettings.create({ enchants: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotMainHand, effectIds: [1897] })] }),
		);

		expect(shippedRows(database)).toEqual([
			[1897, lesserImpact.name],
			[1897, superiorStriking.name],
		]);
	});

	it('ships every row of an equipped enchant’s effect id', () => {
		const sword = new EquippedItem({
			item: UIItem.create({ id: 10, type: ItemType.ItemTypeWeapon, scalingOptions: { 0: { ilvl: 60 } } }),
			enchant: superiorStriking,
		});
		const database = makeBulkItemDatabaseFromSpecs(db, new Gear({ [ItemSlot.ItemSlotMainHand]: sword }), BulkSettings.create());

		expect(shippedRows(database)).toEqual([
			[1897, lesserImpact.name],
			[1897, superiorStriking.name],
		]);
	});
});
