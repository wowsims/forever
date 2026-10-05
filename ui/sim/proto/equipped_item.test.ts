import { ItemType } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { EquippedItem } from './equipped_item';

const chestItem = () =>
	Item.create({
		id: 30107,
		name: 'Vestments of the Sea-Witch',
		type: ItemType.ItemTypeChest,
		scalingOptions: { 0: { ilvl: 141, randPropPoints: 0, weaponDamageMin: 0, weaponDamageMax: 0, stats: [] } },
	});

const enchant = (effectId: number) => Enchant.create({ effectId, itemId: effectId });

// The second parameter of equals(other, ignoreEnchants?) is ignoreEnchants; pin its position.
describe('EquippedItem.equals', () => {
	const base = new EquippedItem({ item: chestItem(), enchant: enchant(2661) });

	it('takes enchants into account by default', () => {
		expect(base.equals(new EquippedItem({ item: chestItem(), enchant: enchant(2661) }))).toBe(true);
		expect(base.equals(new EquippedItem({ item: chestItem(), enchant: enchant(2669) }))).toBe(false);
	});

	it('reads the second parameter as ignoreEnchants', () => {
		const otherEnchant = new EquippedItem({ item: chestItem(), enchant: enchant(2669) });
		expect(base.equals(otherEnchant, true)).toBe(true);
		expect(base.equals(otherEnchant, false)).toBe(false);
	});
});
