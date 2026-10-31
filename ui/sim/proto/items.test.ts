import { EnchantType, HandType, ItemType, WeaponType } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { enchantAppliesToItem, isPVPItem } from './items';

const offHandEnchant = Enchant.create({ effectId: 7659, type: ItemType.ItemTypeWeapon, enchantType: EnchantType.EnchantTypeOffHand });
const shieldEnchant = Enchant.create({ effectId: 7663, type: ItemType.ItemTypeWeapon, enchantType: EnchantType.EnchantTypeShield });
const weaponEnchant = Enchant.create({ effectId: 1897, type: ItemType.ItemTypeWeapon });

const holdable = Item.create({ id: 1, type: ItemType.ItemTypeWeapon, weaponType: WeaponType.WeaponTypeOffHand, handType: HandType.HandTypeOffHand });
const shield = Item.create({ id: 2, type: ItemType.ItemTypeWeapon, weaponType: WeaponType.WeaponTypeShield, handType: HandType.HandTypeOffHand });
const sword = Item.create({ id: 3, type: ItemType.ItemTypeWeapon, weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });

describe('enchantAppliesToItem on the off hand', () => {
	it('puts an off-hand enchant on a held-in-off-hand item only', () => {
		expect(enchantAppliesToItem(offHandEnchant, holdable)).toBe(true);
		expect(enchantAppliesToItem(offHandEnchant, shield)).toBe(false);
		expect(enchantAppliesToItem(offHandEnchant, sword)).toBe(false);
	});

	it('puts a shield enchant on a shield only', () => {
		expect(enchantAppliesToItem(shieldEnchant, shield)).toBe(true);
		expect(enchantAppliesToItem(shieldEnchant, holdable)).toBe(false);
		expect(enchantAppliesToItem(shieldEnchant, sword)).toBe(false);
	});

	it('keeps a weapon enchant off shields and held-in-off-hand items', () => {
		expect(enchantAppliesToItem(weaponEnchant, sword)).toBe(true);
		expect(enchantAppliesToItem(weaponEnchant, shield)).toBe(false);
		expect(enchantAppliesToItem(weaponEnchant, holdable)).toBe(false);
	});
});

describe('isPVPItem', () => {
	it.each(["Knight-Captain's Plate Helm", "Premier Warlord's Linked Helm", 'Premier Chain Boots', "Highlander's Plate Girdle", 'Insignia of the Horde'])(
		'counts %s as a PvP item',
		name => expect(isPVPItem(Item.create({ id: 10, name }))).toBe(true),
	);

	it.each(['Thunderfury, Blessed Blade of the Windseeker', "Commander's Boots", "Knight's Lance", "Sentinel's Guard"])(
		'keeps %s out of the PvP items',
		name => expect(isPVPItem(Item.create({ id: 11, name }))).toBe(false),
	);
});
