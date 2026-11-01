import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { Class, EnchantType, HandType, ItemSlot, ItemType, Profession, Stat, WeaponType } from '@generated/proto/common';
import { BulkEnchantPreset, type UIEnchant, type UIItem } from '@generated/proto/ui';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { describe, expect, it, vi } from 'vitest';

import type { BulkEnchantPick } from './enchant_picks';
import type { BulkGearSetPick } from './gear_sets';
import { bulkGearSetups, gearSetupLabel, gearSetupsSentWith } from './gear_setups';

vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string) => (({ 'bulk_tab.results.current_gear': 'Current gear' }) as Record<string, string>)[key] ?? key,
	},
}));

const enchant = (effectId: number, type: ItemType, overrides: Partial<UIEnchant> = {}) =>
	makeUIEnchant({ effectId, name: `Enchant ${effectId}`, type, ...overrides });

const CLOAK_AGILITY = enchant(849, ItemType.ItemTypeBack);
const CLOAK_ARMOR = enchant(884, ItemType.ItemTypeBack);
const CHEST_STATS = enchant(1891, ItemType.ItemTypeChest);
const CRUSADER = enchant(1900, ItemType.ItemTypeWeapon);
const TWO_HAND_AGILITY = enchant(2646, ItemType.ItemTypeWeapon, { enchantType: EnchantType.EnchantTypeTwoHand });
const SHIELD_STAMINA = enchant(929, ItemType.ItemTypeWeapon, { enchantType: EnchantType.EnchantTypeShield });
const HELD_INTELLECT = enchant(7659, ItemType.ItemTypeWeapon, { enchantType: EnchantType.EnchantTypeOffHand });
const RING_STAMINA = enchant(7301, ItemType.ItemTypeFinger, { requiredProfession: Profession.Enchanting });
const RING_AGILITY = enchant(7302, ItemType.ItemTypeFinger, { requiredProfession: Profession.Enchanting });

const ALL_ENCHANTS = [CLOAK_AGILITY, CLOAK_ARMOR, CHEST_STATS, CRUSADER, TWO_HAND_AGILITY, SHIELD_STAMINA, HELD_INTELLECT, RING_STAMINA, RING_AGILITY];

type SlotEnchants = Partial<Record<ItemSlot, UIEnchant[]>>;

const SLOT_ENCHANTS: SlotEnchants = {
	[ItemSlot.ItemSlotBack]: [CLOAK_AGILITY, CLOAK_ARMOR],
	[ItemSlot.ItemSlotChest]: [CHEST_STATS],
	[ItemSlot.ItemSlotMainHand]: [CRUSADER, TWO_HAND_AGILITY],
	[ItemSlot.ItemSlotOffHand]: [CRUSADER, SHIELD_STAMINA, HELD_INTELLECT],
	[ItemSlot.ItemSlotFinger1]: [RING_STAMINA, RING_AGILITY],
	[ItemSlot.ItemSlotFinger2]: [RING_STAMINA, RING_AGILITY],
};

// Like the real database and player: the database lists every enchant of a slot, the player's list leaves out the
// enchants that need Enchanting when it lacks it.
const makePlayer = ({ enchanter = true, slotEnchants = SLOT_ENCHANTS, byEffectId = ALL_ENCHANTS } = {}) =>
	makeBulkPlayer({
		hasProfession: (profession: Profession) => enchanter && profession === Profession.Enchanting,
		getEnchants: (slot: ItemSlot) => (slotEnchants[slot] ?? []).filter(row => enchanter || row.requiredProfession !== Profession.Enchanting),
		sim: {
			db: {
				getEnchants: (slot: ItemSlot) => slotEnchants[slot] ?? [],
				getEnchantsByEffectId: (effectId: number) => byEffectId.filter(row => row.effectId === effectId),
			},
		},
	});

const armor = (id: number, type: ItemType): Partial<UIItem> => ({
	id,
	type,
	weaponType: WeaponType.WeaponTypeUnknown,
	handType: HandType.HandTypeUnknown,
	unique: false,
});

const weapon = (id: number, weaponType: WeaponType, handType: HandType): Partial<UIItem> => ({ id, type: ItemType.ItemTypeWeapon, weaponType, handType });

const CLOAK = armor(101, ItemType.ItemTypeBack);
const CHEST = armor(102, ItemType.ItemTypeChest);
const RING_A = armor(103, ItemType.ItemTypeFinger);
const RING_B = armor(104, ItemType.ItemTypeFinger);
const SWORD = weapon(105, WeaponType.WeaponTypeSword, HandType.HandTypeOneHand);
const OFF_HAND_SWORD = weapon(106, WeaponType.WeaponTypeSword, HandType.HandTypeOneHand);
const HELD = weapon(107, WeaponType.WeaponTypeOffHand, HandType.HandTypeOffHand);

const gearOf = (items: Array<[ItemSlot, Partial<UIItem>, UIEnchant?]>) =>
	new Gear(Object.fromEntries(items.map(([slot, item, worn]) => [slot, makeEquippedItem({ item, enchant: worn })])));

const enchantIn = (gear: Gear, slot: ItemSlot) => gear.getEquippedItem(slot)?.enchant?.effectId ?? null;

const enchantsIn = (gear: Gear, slots: ItemSlot[]) => slots.map(slot => enchantIn(gear, slot));

const currentSet = (gear: Gear, bonusStats = new Stats()): BulkGearSetPick => ({ name: '', gear, bonusStats });

const savedSet = (name: string, gear: Gear, bonusStats = new Stats()): BulkGearSetPick => ({ name, gear, bonusStats });

const presetPick = (name: string, slots: Array<[ItemSlot, number[]]>): BulkEnchantPick => ({
	name,
	preset: BulkEnchantPreset.create({ name, slots: slots.map(([slot, effectIds]) => BulkEnchantSlotOptions.create({ slot, effectIds })) }),
});

const AGILITY = presetPick('Agility', [[ItemSlot.ItemSlotBack, [CLOAK_AGILITY.effectId]]]);

const BACK_AND_CHEST = [ItemSlot.ItemSlotBack, ItemSlot.ItemSlotChest];

describe('bulkGearSetups', () => {
	it('gives each gear set one setup with its own enchants without an enchant preset', () => {
		const current = currentSet(gearOf([[ItemSlot.ItemSlotBack, CLOAK, CLOAK_ARMOR]]));
		const saved = savedSet('Raid set', gearOf([[ItemSlot.ItemSlotChest, CHEST]]));

		const setups = bulkGearSetups(makePlayer(), [current, saved], []);

		expect(setups.map(setup => [setup.gearSet, setup.gear, setup.gearSetGear])).toEqual([
			['', current.gear, current.gear],
			['Raid set', saved.gear, saved.gear],
		]);
	});

	it('pools the presets per slot: each slot keeps its own enchant or takes any preset’s, in every combination, own first', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotBack, CLOAK],
			[ItemSlot.ItemSlotChest, CHEST],
		]);
		const raid = presetPick('Raid', [
			[ItemSlot.ItemSlotBack, [CLOAK_AGILITY.effectId]],
			[ItemSlot.ItemSlotChest, [CHEST_STATS.effectId]],
		]);
		const tank = presetPick('Tank', [[ItemSlot.ItemSlotBack, [CLOAK_ARMOR.effectId]]]);

		const setups = bulkGearSetups(makePlayer(), [currentSet(gear)], [raid, tank]);

		expect(setups.map(setup => enchantsIn(setup.gear, BACK_AND_CHEST))).toEqual([
			[null, null],
			[null, CHEST_STATS.effectId],
			[CLOAK_AGILITY.effectId, null],
			[CLOAK_AGILITY.effectId, CHEST_STATS.effectId],
			[CLOAK_ARMOR.effectId, null],
			[CLOAK_ARMOR.effectId, CHEST_STATS.effectId],
		]);
		expect(new Set(setups.map(setup => setup.gearSetGear))).toEqual(new Set([gear]));
	});

	it('pools the same enchants whatever order the presets are ticked in', () => {
		const gear = gearOf([[ItemSlot.ItemSlotBack, CLOAK]]);
		const tank = presetPick('Tank', [[ItemSlot.ItemSlotBack, [CLOAK_ARMOR.effectId]]]);
		const backOf = (picks: BulkEnchantPick[]) =>
			bulkGearSetups(makePlayer(), [currentSet(gear)], picks).map(setup => enchantIn(setup.gear, ItemSlot.ItemSlotBack));

		expect(backOf([tank, AGILITY])).toEqual(backOf([AGILITY, tank]));
	});

	it('runs every gear set, the current gear first, each with its own enchants and then the preset’s', () => {
		const sets = [
			currentSet(gearOf([[ItemSlot.ItemSlotBack, CLOAK, CLOAK_ARMOR]])),
			savedSet(
				'Raid set',
				gearOf([
					[ItemSlot.ItemSlotBack, CLOAK],
					[ItemSlot.ItemSlotChest, CHEST],
				]),
			),
		];

		const setups = bulkGearSetups(makePlayer(), sets, [AGILITY]);

		expect(setups.map(setup => [setup.gearSet, ...enchantsIn(setup.gear, BACK_AND_CHEST)])).toEqual([
			['', CLOAK_ARMOR.effectId, null],
			['', CLOAK_AGILITY.effectId, null],
			['Raid set', null, null],
			['Raid set', CLOAK_AGILITY.effectId, null],
		]);
	});

	it('takes the enchant row from the slot’s own enchant list, as Gear mode does', () => {
		const elsewhere = enchant(849, ItemType.ItemTypeBack, { name: 'Not in the slot list' });
		const otherClass = enchant(849, ItemType.ItemTypeBack, { name: 'Mage row', classAllowlist: [Class.ClassMage] });
		const chestRow = enchant(849, ItemType.ItemTypeChest, { name: 'Chest row' });
		const cloakRow = enchant(849, ItemType.ItemTypeBack, { name: 'Cloak row' });
		const player = makePlayer({
			slotEnchants: { [ItemSlot.ItemSlotBack]: [otherClass, chestRow, cloakRow] },
			byEffectId: [elsewhere, otherClass, chestRow, cloakRow],
		});

		const [, setup] = bulkGearSetups(player, [currentSet(gearOf([[ItemSlot.ItemSlotBack, CLOAK]]))], [AGILITY]);

		expect(setup.gear.getEquippedItem(ItemSlot.ItemSlotBack)?.enchant?.name).toBe('Cloak row');
	});

	it('crosses the rings’ own enchants and two preset ring enchants into nine setups', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotFinger1, RING_A],
			[ItemSlot.ItemSlotFinger2, RING_B],
		]);
		const rings = presetPick('Rings', [[ItemSlot.ItemSlotFinger1, [RING_STAMINA.effectId, RING_AGILITY.effectId]]]);

		const setups = bulkGearSetups(makePlayer({ enchanter: false }), [currentSet(gear)], [rings]);

		const options = [null, RING_STAMINA.effectId, RING_AGILITY.effectId];
		expect(setups.map(setup => enchantsIn(setup.gear, [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2]))).toEqual(
			options.flatMap(first => options.map(second => [first, second])),
		);
	});

	it('offers a Main Hand enchant to both hands and a Finger 1 enchant to both rings', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotMainHand, SWORD],
			[ItemSlot.ItemSlotOffHand, OFF_HAND_SWORD],
			[ItemSlot.ItemSlotFinger1, RING_A],
			[ItemSlot.ItemSlotFinger2, RING_B],
		]);
		const preset = presetPick('Raid', [
			[ItemSlot.ItemSlotMainHand, [CRUSADER.effectId]],
			[ItemSlot.ItemSlotFinger1, [RING_STAMINA.effectId]],
		]);

		const setups = bulkGearSetups(makePlayer(), [currentSet(gear)], [preset]);

		expect(setups).toHaveLength(16);
		expect(enchantsIn(setups[15].gear, [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2])).toEqual([
			CRUSADER.effectId,
			CRUSADER.effectId,
			RING_STAMINA.effectId,
			RING_STAMINA.effectId,
		]);
	});

	it('skips an enchant that does not fit the item, keeping the set’s own there', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotBack, CLOAK],
			[ItemSlot.ItemSlotMainHand, SWORD, CRUSADER],
			[ItemSlot.ItemSlotOffHand, HELD, HELD_INTELLECT],
		]);
		const preset = presetPick('Mixed', [
			[ItemSlot.ItemSlotBack, [CLOAK_AGILITY.effectId]],
			[ItemSlot.ItemSlotMainHand, [TWO_HAND_AGILITY.effectId]],
			[ItemSlot.ItemSlotOffHand, [SHIELD_STAMINA.effectId]],
		]);

		const setups = bulkGearSetups(makePlayer(), [currentSet(gear)], [preset]);

		expect(setups.map(setup => enchantsIn(setup.gear, [ItemSlot.ItemSlotBack, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand]))).toEqual([
			[null, CRUSADER.effectId, HELD_INTELLECT.effectId],
			[CLOAK_AGILITY.effectId, CRUSADER.effectId, HELD_INTELLECT.effectId],
		]);
	});

	it('runs a set once with its own enchants when the presets fit nothing on it or hold only what it wears', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotBack, CLOAK, CLOAK_AGILITY],
			[ItemSlot.ItemSlotMainHand, SWORD, CRUSADER],
			[ItemSlot.ItemSlotOffHand, HELD],
		]);
		const shieldOnly = presetPick('Shield', [[ItemSlot.ItemSlotOffHand, [SHIELD_STAMINA.effectId]]]);
		const twoHandOnly = presetPick('Two-hand', [[ItemSlot.ItemSlotMainHand, [TWO_HAND_AGILITY.effectId]]]);

		const setups = bulkGearSetups(makePlayer(), [currentSet(gear)], [shieldOnly, twoHandOnly, AGILITY]);

		expect(setups.map(setup => setup.gear)).toEqual([gear]);
	});

	it('drops a setup equal to an earlier one, so a set the current gear’s variant already wears runs under the current gear', () => {
		const sets = [
			currentSet(gearOf([[ItemSlot.ItemSlotBack, CLOAK, CLOAK_ARMOR]])),
			savedSet('Agile', gearOf([[ItemSlot.ItemSlotBack, CLOAK, CLOAK_AGILITY]])),
		];

		const setups = bulkGearSetups(makePlayer(), sets, [AGILITY]);

		expect(setups.map(setup => [setup.gearSet, enchantIn(setup.gear, ItemSlot.ItemSlotBack)])).toEqual([
			['', CLOAK_ARMOR.effectId],
			['', CLOAK_AGILITY.effectId],
		]);
	});

	it('keeps two setups with the same gear and different bonus stats', () => {
		const gear = gearOf([[ItemSlot.ItemSlotBack, CLOAK]]);
		const buffed = savedSet('Buffed', gear, new Stats().withStat(Stat.StatStrength, 10));

		const setups = bulkGearSetups(makePlayer(), [currentSet(gear), buffed], [AGILITY]);

		expect(setups.map(setup => setup.gearSet)).toEqual(['', '', 'Buffed', 'Buffed']);
	});
});

describe('gearSetupsSentWith', () => {
	it('drops the ring enchants for a request without Enchanting, keeping the first of setups that become equal', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotBack, CLOAK],
			[ItemSlot.ItemSlotFinger1, RING_A],
			[ItemSlot.ItemSlotFinger2, RING_B],
		]);
		const raid = presetPick('Raid', [
			[ItemSlot.ItemSlotBack, [CLOAK_AGILITY.effectId]],
			[ItemSlot.ItemSlotFinger1, [RING_STAMINA.effectId]],
		]);
		const setups = bulkGearSetups(makePlayer(), [currentSet(gear)], [raid]);

		const sent = gearSetupsSentWith(setups, false);

		expect(setups).toHaveLength(8);
		expect(gearSetupsSentWith(setups, true)).toEqual(setups);
		expect(sent.map(setup => enchantsIn(setup.gear, [ItemSlot.ItemSlotBack, ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2]))).toEqual([
			[null, null, null],
			[CLOAK_AGILITY.effectId, null, null],
		]);
		expect(sent.map(setup => setup.ownGear)).toEqual([setups[0].gear, setups[4].gear]);
	});

	it('keeps both rings of one limit category, as gear loaded from a spec can wear them, while it drops their enchants', () => {
		const gear = gearOf([
			[ItemSlot.ItemSlotFinger1, { ...RING_A, limitCategory: 710 }, RING_STAMINA],
			[ItemSlot.ItemSlotFinger2, { ...RING_B, limitCategory: 710 }, RING_AGILITY],
		]);
		const [sent] = gearSetupsSentWith(bulkGearSetups(makePlayer(), [currentSet(gear)], []), false);

		const rings = [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2].map(slot => sent.gear.getEquippedItem(slot));
		expect(rings.map(ring => ring?.item.id)).toEqual([RING_A.id, RING_B.id]);
		expect(rings.map(ring => ring?.enchant ?? null)).toEqual([null, null]);
	});
});

describe('gearSetupLabel', () => {
	it('names a setup by its gear set, and the current gear’s as the current gear', () => {
		const sets = [currentSet(gearOf([[ItemSlot.ItemSlotBack, CLOAK]])), savedSet('Pre-BiS', gearOf([[ItemSlot.ItemSlotChest, CHEST]]))];

		const setups = bulkGearSetups(makePlayer(), sets, [AGILITY]);

		expect(setups.map(gearSetupLabel)).toEqual(['Current gear', 'Current gear', 'Pre-BiS']);
	});
});
