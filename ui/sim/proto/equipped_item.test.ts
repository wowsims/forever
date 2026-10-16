import { ItemRandomSuffix, ItemSlot, ItemType, PseudoStat, Stat } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { enchantEP, EquippedItem } from './equipped_item';
import { Stats } from './stats';

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

describe('EquippedItem with a random suffix', () => {
	const ring = Item.create({
		id: 12017,
		name: 'Prismatic Band',
		type: ItemType.ItemTypeFinger,
		scalingOptions: { 0: { ilvl: 65, randPropPoints: 16, weaponDamageMin: 0, weaponDamageMax: 0, stats: { [Stat.StatIntellect]: 5 }, areaStats: [] } },
	});
	const suffixStats = new Array<number>(Stat.StatPhysicalDamage + 1).fill(0);
	suffixStats[Stat.StatIntellect] = 4106; // 6.57 at 16 points.
	suffixStats[Stat.StatStamina] = 7266; // 11.63.
	const suffixPseudoStats = new Array<number>(PseudoStat.PseudoStatDodgePercent + 1).fill(0);
	suffixPseudoStats[PseudoStat.PseudoStatDodgePercent] = 1;
	const suffix = ItemRandomSuffix.create({ id: 12748, name: 'of the Test', stats: suffixStats, pseudoStats: suffixPseudoStats });

	it('adds each rounded suffix stat on top of the item stats, and its pseudo stats unscaled', () => {
		const stats = new EquippedItem({ item: ring, randomSuffix: suffix }).withDynamicStats().calcStats();
		expect(stats.getStat(Stat.StatIntellect)).toBe(5 + 7);
		expect(stats.getStat(Stat.StatStamina)).toBe(12);
		expect(stats.getPseudoStat(PseudoStat.PseudoStatDodgePercent)).toBe(1);
	});

	it('gives the same stats when the dynamic item is worked out again', () => {
		const once = new EquippedItem({ item: ring, randomSuffix: suffix }).withDynamicStats();
		const twice = new EquippedItem({ item: once.item, randomSuffix: suffix }).withDynamicStats();
		expect(twice.calcStats().equals(once.calcStats())).toBe(true);
	});
});

describe('EquippedItem with a random suffix that reduces caster DPS', () => {
	const axe = () =>
		Item.create({
			id: 3210,
			name: 'Brutal War Axe',
			type: ItemType.ItemTypeWeapon,
			weaponSpeed: 3.6,
			scalingOptions: {
				0: {
					ilvl: 30,
					randPropPoints: 11,
					weaponDamageMin: 58,
					weaponDamageMax: 88,
					unroundedDps: 20.483225,
					stats: [],
					areaStats: [],
					reducedCasterWeaponDamage: { weaponDamageMin: 43, weaponDamageMax: 66, unroundedDps: 15.231116 },
				},
			},
		});
	const physician = ItemRandomSuffix.create({ id: 12911, name: 'of the Physician', stats: [], pseudoStats: [], reducedCasterDps: true });
	const bear = ItemRandomSuffix.create({ id: 12718, name: 'of the Bear', stats: [], pseudoStats: [] });

	it('swaps in the reduced damage range and its DPS', () => {
		const equipped = new EquippedItem({ item: axe(), randomSuffix: physician }).withDynamicStats();
		expect([equipped.item.weaponDamageMin, equipped.item.weaponDamageMax]).toEqual([43, 66]);
		expect(equipped.calcStats(ItemSlot.ItemSlotMainHand).getPseudoStat(PseudoStat.PseudoStatMainHandDps)).toBeCloseTo((43 + 66) / 2 / 3.6);
	});

	it('keeps the weapon damage with any other suffix, and with none', () => {
		for (const randomSuffix of [bear, null]) {
			const equipped = new EquippedItem({ item: axe(), randomSuffix }).withDynamicStats();
			expect([equipped.item.weaponDamageMin, equipped.item.weaponDamageMax]).toEqual([58, 88]);
			expect(equipped.calcStats(ItemSlot.ItemSlotMainHand).getPseudoStat(PseudoStat.PseudoStatMainHandDps)).toBeCloseTo((58 + 88) / 2 / 3.6);
		}
	});
});

// One point of ArmorNoMods is one point of sheet armor and has a weight of its own, which the gear optimizer's candidate
// filter reads with no scale; the Armor weight is per point of armor from items.
describe('enchantEP with ArmorNoMods', () => {
	const armor = 125;
	const armorEnchant = (stat: Stat) => Enchant.create({ effectId: 1504, stats: Stats.fromMap({ [stat]: armor, [Stat.StatStamina]: 3 }).asProtoArray() });

	it('values ArmorNoMods at its own weight and Armor at the Armor weight', () => {
		const weights = Stats.fromMap({ [Stat.StatArmor]: 0.5, [Stat.StatArmorNoMods]: 0.125, [Stat.StatStamina]: 2 });
		expect(enchantEP(armorEnchant(Stat.StatArmorNoMods), weights, 0, 0)).toBeCloseTo(armor * 0.125 + 3 * 2);
		expect(enchantEP(armorEnchant(Stat.StatArmor), weights, 0, 0)).toBeCloseTo(armor * 0.5 + 3 * 2);
	});

	it('gives ArmorNoMods nothing where only Armor has a weight', () => {
		const weights = Stats.fromMap({ [Stat.StatArmor]: 0.5 });
		expect(enchantEP(armorEnchant(Stat.StatArmorNoMods), weights, 0, 0)).toBe(0);
	});
});
