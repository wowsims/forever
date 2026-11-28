import { ArmorType, Class, HandType, ItemType, RangedWeaponType, ScalingItemProperties, SpellSchool, Stat, WeaponType } from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Stats } from '@sim/proto/stats';
import type { ItemTooltipData, SetTooltipData } from '@sim/proto/tooltips';
import type { GameTooltipLine } from '@ui-kit/GameTooltip';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import character from '../../../../assets/locales/en/character.json';
import english from '../../../../assets/locales/en/translation.json';
import { type ItemTooltipInput, itemTooltipLines, otherItemTooltipLines } from './item_tooltip';

// The test setup loads no locale, and the lines are the game's English.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
	i18n.addResourceBundle('en', 'character', character, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
	i18n.removeResourceBundle('en', 'character');
});

const equippedItem = (item: Partial<UIItem>, stats: Partial<Record<Stat, number>>, enchant?: UIEnchant) =>
	new EquippedItem({
		item: UIItem.create({ scalingOptions: { 0: ScalingItemProperties.create({ ilvl: item.ilvl }) }, stats: Stats.fromMap(stats).asProtoArray(), ...item }),
		enchant,
	});

const lines = (input: Partial<ItemTooltipInput> & Pick<ItemTooltipInput, 'equipped'>) =>
	itemTooltipLines({ worn: new Set(), itemName: id => `Item ${id}`, spellText: id => `Spell ${id}.`, ...input });

// Each line as it reads, with the right-hand text after a bar.
const text = (tooltip: Array<GameTooltipLine>) => tooltip.map(line => (line.right ? `${line.left} | ${line.right}` : line.left));
const line = (tooltip: Array<GameTooltipLine>, left: string) => tooltip.find(entry => entry.left === left);

const arcaniteReaper = (enchant?: UIEnchant) =>
	equippedItem(
		{
			ilvl: 63,
			reqLvl: 58,
			type: ItemType.ItemTypeWeapon,
			handType: HandType.HandTypeTwoHand,
			weaponType: WeaponType.WeaponTypeAxe,
			weaponDamageMin: 153,
			weaponDamageMax: 256,
			weaponSpeed: 3.8,
		},
		{ [Stat.StatStamina]: 13, [Stat.StatAttackPower]: 62, [Stat.StatRangedAttackPower]: 62 },
		enchant,
	);

const crusader = UIEnchant.create({ effectId: 1900, name: 'Enchant Weapon - Crusader', spellId: 20007 });

describe('itemTooltipLines', () => {
	it('lays out an enchanted two-hander as the game does, the enchant after its stats and the price last', () => {
		const tooltip = lines({ equipped: arcaniteReaper(crusader), data: { b: 2, s: 73036 }, enchantName: 'Crusader', enchantIcon: 'crusader.jpg' });
		expect(text(tooltip)).toEqual([
			'Item Level 63',
			'Binds when equipped',
			'Two-Hand | Axe',
			'153 - 256 Damage | Speed 3.80',
			'(53.82 damage per second)',
			'+13 Stamina',
			'Enchanted: Crusader',
			'Requires Level 58',
			'Equip: +62 Attack Power.',
			'Sell Price:',
		]);
		expect(line(tooltip, 'Enchanted: Crusader')).toMatchObject({ tone: 'green', icon: 'crusader.jpg' });
		expect(line(tooltip, 'Sell Price:')?.money).toBe(73036);
	});

	it('names an enchant the tooltip files lack by what it adds', () => {
		expect(text(lines({ equipped: arcaniteReaper(crusader) }))).toContain('Enchanted: Crusader');
	});

	describe('a set piece', () => {
		const set: SetTooltipData = { n: 'Devilsaur Armor', p: [15062, 15063], b: [[2, 460230]] };
		const gauntlets = equippedItem(
			{ id: 15063, ilvl: 58, type: ItemType.ItemTypeHands, armorType: ArmorType.ArmorTypeLeather, setId: 143 },
			{ [Stat.StatArmor]: 103, [Stat.StatCritRating]: 14 },
		);
		const setLines = (worn: Array<number>) => {
			const tooltip = lines({
				equipped: gauntlets,
				set,
				worn: new Set(worn),
				itemName: id => (id === 15062 ? 'Devilsaur Leggings' : 'Devilsaur Gauntlets'),
			});
			return tooltip.slice(tooltip.findIndex(entry => entry.left?.startsWith('Devilsaur Armor')));
		};

		it('lights the pieces worn, and greys a bonus with its count until enough are', () => {
			expect(setLines([15063])).toEqual([
				{ left: 'Devilsaur Armor (1/2)', tone: 'gold', spaced: true },
				{ left: 'Devilsaur Leggings', tone: 'grey', indent: true },
				{ left: 'Devilsaur Gauntlets', tone: 'yellow', indent: true },
				{ left: '(2) Set: Spell 460230.', tone: 'grey', spaced: true },
			]);
			expect(setLines([])[0].left).toBe('Devilsaur Armor (0/2)');
		});

		it('writes an active bonus green without its count', () => {
			expect(setLines([15062, 15063]).at(-1)).toEqual({ left: 'Set: Spell 460230.', tone: 'green', spaced: true });
		});

		it('writes the crit it adds to one decimal, as the game does', () => {
			expect(text(lines({ equipped: gauntlets }))).toContain('Equip: Improves your chance to get a critical strike by 1.0%.');
		});
	});

	it('writes what the item asks of the player, in the game order', () => {
		const cap = equippedItem({ ilvl: 47, reqLvl: 50, type: ItemType.ItemTypeHead, classAllowlist: [Class.ClassHunter] }, {});
		const data: ItemTooltipData = { k: { n: 'Engineering', r: 215 }, g: { n: 'The League of Arathor', r: 6 }, u: -1 };
		expect(text(lines({ equipped: cap, data }))).toEqual([
			'Item Level 47',
			'Unique-Equipped',
			'Head',
			'Classes: Hunter',
			'Requires Level 50',
			'Requires Engineering (215)',
			'Requires The League of Arathor - Revered',
		]);
	});

	it("writes a wand's school in its damage", () => {
		const wand = equippedItem(
			{
				ilvl: 41,
				type: ItemType.ItemTypeRanged,
				rangedWeaponType: RangedWeaponType.RangedWeaponTypeWand,
				spellSchool: SpellSchool.SpellSchoolFire,
				weaponDamageMin: 35,
				weaponDamageMax: 66,
				weaponSpeed: 1.5,
			},
			{},
		);
		expect(text(lines({ equipped: wand }))).toContain('35 - 66 Fire Damage | Speed 1.50');
	});
});

describe('otherItemTooltipLines', () => {
	const spellText = (id: number) => `Spell ${id}.`;

	it('lays out a quiver with its slots, classes and equip effect', () => {
		const quiver: ItemTooltipData = { b: 1, u: 1, l: 75, r: 60, o: 18, t: 'Quiver', c: 18, a: [Class.ClassHunter], e: [[1, 29414, 0]] };
		expect(text(otherItemTooltipLines(quiver, spellText))).toEqual([
			'Item Level 75',
			'Binds when picked up',
			'Unique',
			'18 Slot Quiver',
			'Classes: Hunter',
			'Requires Level 60',
			'Equip: Spell 29414.',
		]);
	});

	it('writes ammo in its slot with its type beside it', () => {
		expect(text(otherItemTooltipLines({ l: 59, o: 24, t: 'Arrow', d: 20 }, spellText))).toEqual([
			'Item Level 59',
			'Ammo | Arrow',
			'Adds 20 damage per second',
		]);
	});

	it("writes a consumable's charges under its use, no item level, and its price", () => {
		const oil: ItemTooltipData = { l: 55, r: 45, e: [[0, 25122, 0, 5]], s: 1000 };
		expect(text(otherItemTooltipLines(oil, spellText))).toEqual(['Requires Level 45', 'Use: Spell 25122.', '5 Charges', 'Sell Price:']);
		expect(otherItemTooltipLines({ e: [[0, 17531, 120000]], u: 5 }, spellText).map(entry => entry.left)).toEqual([
			'Unique (5)',
			'Use: Spell 17531. (2 Min Cooldown)',
		]);
	});
});
