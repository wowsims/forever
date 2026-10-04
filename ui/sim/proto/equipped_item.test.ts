import { GemColor, ItemType } from '@generated/proto/common';
import { UIEnchant as Enchant, UIGem as Gem, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { EquippedItem } from './equipped_item';

const socketedItem = () =>
	Item.create({
		id: 30107,
		name: 'Vestments of the Sea-Witch',
		type: ItemType.ItemTypeChest,
		gemSockets: [GemColor.GemColorRed],
		scalingOptions: { 0: { ilvl: 141, randPropPoints: 0, weaponDamageMin: 0, weaponDamageMax: 0, stats: [] } },
	});

const gem = (id: number) => Gem.create({ id, color: GemColor.GemColorRed });
const enchant = (effectId: number) => Enchant.create({ effectId, itemId: effectId });

// This tree's signature is equals(other, ignoreEnchants?, ignoreGems?) where MoP's is
// equals(other, ignoreReforge?, ignoreEnchants?, ignoreGems?, ignoreUpgrades?). A call written
// against one silently means something else against the other, so pin the positions.
describe('EquippedItem.equals', () => {
	const base = new EquippedItem({ item: socketedItem(), enchant: enchant(2661), gems: [gem(32196)] });

	it('takes enchants and gems into account by default', () => {
		expect(base.equals(new EquippedItem({ item: socketedItem(), enchant: enchant(2661), gems: [gem(32196)] }))).toBe(true);
		expect(base.equals(new EquippedItem({ item: socketedItem(), enchant: enchant(2669), gems: [gem(32196)] }))).toBe(false);
		expect(base.equals(new EquippedItem({ item: socketedItem(), enchant: enchant(2661), gems: [gem(32215)] }))).toBe(false);
	});

	it('reads the second parameter as ignoreEnchants', () => {
		const otherEnchant = new EquippedItem({ item: socketedItem(), enchant: enchant(2669), gems: [gem(32196)] });
		expect(base.equals(otherEnchant, true)).toBe(true);
		expect(base.equals(otherEnchant, false)).toBe(false);
	});

	it('reads the third parameter as ignoreGems, and the second does not cover gems', () => {
		const otherGem = new EquippedItem({ item: socketedItem(), enchant: enchant(2661), gems: [gem(32215)] });
		expect(base.equals(otherGem, true)).toBe(false);
		expect(base.equals(otherGem, false, true)).toBe(true);
	});

	it('ignores both when both flags are set', () => {
		const other = new EquippedItem({ item: socketedItem(), enchant: enchant(2669), gems: [gem(32215)] });
		expect(base.equals(other, true, true)).toBe(true);
	});
});
