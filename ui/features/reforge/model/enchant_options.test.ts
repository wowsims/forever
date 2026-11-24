import { enchantPresetPool, saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { Class, EnchantProcKind, HandType, ItemSlot, ItemType, PseudoStat, RangedWeaponType, Stat, WeaponType } from '@generated/proto/common';
import type { UIEnchant, UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { describe, expect, it, vi } from 'vitest';

import { buildEnchantOptions, classifyEnchantSlots, skippedEnchantSlots } from './enchant_options';

vi.mock('@sim/constants/missing_effects_auto_gen', () => ({
	MISSING_ITEM_EFFECTS: new Map(),
	MISSING_ENCHANT_EFFECTS: new Map([[7223, ['Deals Holy damage to an attacker.']]]),
}));

const WEIGHTS = new Stats().withStat(Stat.StatStrength, 2).withStat(Stat.StatAgility, 1).withPseudoStat(PseudoStat.PseudoStatRangedDps, 3);

const stats = (stat: Stat, value: number) => new Stats().withStat(stat, value).asProtoArray();

const armor = (type: ItemType): Partial<UIItem> => ({ type, weaponType: WeaponType.WeaponTypeUnknown, handType: HandType.HandTypeUnknown, weaponSpeed: 0 });

const BOW: Partial<UIItem> = { ...armor(ItemType.ItemTypeRanged), rangedWeaponType: RangedWeaponType.RangedWeaponTypeBow, weaponSpeed: 3 };
const WAND: Partial<UIItem> = { ...BOW, rangedWeaponType: RangedWeaponType.RangedWeaponTypeWand, weaponSpeed: 1.5 };

const enchant = (effectId: number, name: string, type: ItemType, overrides: Partial<UIEnchant> = {}) => makeUIEnchant({ effectId, name, type, ...overrides });

const WRIST_STRENGTH = enchant(1885, 'Enchant Bracer - Superior Strength', ItemType.ItemTypeWrist, { stats: stats(Stat.StatStrength, 9) });
const WRIST_AGILITY = enchant(247, 'Enchant Bracer - Minor Agility', ItemType.ItemTypeWrist, { stats: stats(Stat.StatAgility, 1) });
const WRIST_SPIRIT = enchant(1883, 'Enchant Bracer - Superior Spirit', ItemType.ItemTypeWrist, { stats: stats(Stat.StatSpirit, 9) });
const CRUSADER = makeUIEnchant({ procKind: EnchantProcKind.EnchantProcKindCombat });
const WEAPON_AGILITY = enchant(2564, 'Enchant Weapon - Agility', ItemType.ItemTypeWeapon, { stats: stats(Stat.StatAgility, 15) });
const CHEST_ABSORPTION = enchant(63, 'Enchant Chest - Lesser Absorption', ItemType.ItemTypeChest, { procKind: EnchantProcKind.EnchantProcKindSurvival });
const CHEST_STATS = enchant(1891, 'Enchant Chest - Greater Stats', ItemType.ItemTypeChest, {
	stats: new Stats().withStat(Stat.StatStrength, 4).withStat(Stat.StatAgility, 4).asProtoArray(),
});
const BACK_THREAT = enchant(2613, 'Enchant Cloak - Subtlety', ItemType.ItemTypeBack);
const BACK_AGILITY = enchant(849, 'Enchant Cloak - Lesser Agility', ItemType.ItemTypeBack, { stats: stats(Stat.StatAgility, 3) });
const SCOPE = enchant(664, 'Sniper Scope', ItemType.ItemTypeRanged, { weaponDamage: 7 });

const ALL_ENCHANTS = [WRIST_STRENGTH, WRIST_AGILITY, WRIST_SPIRIT, CRUSADER, WEAPON_AGILITY, CHEST_ABSORPTION, CHEST_STATS, BACK_THREAT, BACK_AGILITY, SCOPE];

// The rule reads two members of the player: its enchant list and whether its spec tanks.
const makePlayer = ({ isTankSpec = false, enchants = ALL_ENCHANTS }: { isTankSpec?: boolean; enchants?: UIEnchant[] } = {}) =>
	({ playerSpec: { isTankSpec }, getEnchants: () => enchants }) as unknown as Player<any>;

const makeGear = (items: Partial<Record<ItemSlot, Partial<UIItem>>>) =>
	new Gear(Object.fromEntries(Object.entries(items).map(([slot, item]) => [slot, makeEquippedItem({ item })])));

const RING_STAMINA = enchant(7301, 'Enchant Ring - Stamina', ItemType.ItemTypeFinger, { stats: stats(Stat.StatStamina, 4) });
const CHEST_UNIMPLEMENTED = enchant(7223, 'Enchant Chest - Retricutioner', ItemType.ItemTypeChest, { procKind: EnchantProcKind.EnchantProcKindCombat });

type PresetSlots = Array<[ItemSlot, number[]]>;

// The pool comes from the enchant preset model over a real store, so the slot mapping is the one the batch sim uses.
const makePresetPlayer = (checked: Record<string, PresetSlots>, { enchants = ALL_ENCHANTS, unchecked = {} as Record<string, PresetSlots> } = {}) => {
	const player = makeBulkPlayer({ playerSpec: { isTankSpec: false }, getEnchants: () => enchants });
	const save = (presets: Record<string, PresetSlots>) =>
		Object.entries(presets).forEach(([name, slots]) => saveEnchantPreset(player, { name, slots: slots.map(([slot, effectIds]) => ({ slot, effectIds })) }));
	save(checked);
	save(unchecked);
	Object.keys(checked).forEach(name => setEnchantPresetChecked(player, name, true));
	return player;
};

const optimize = (player: Player<any>, gear: Gear, frozenSlots: ItemSlot[], weights: Stats, presetPool?: BulkEnchantSlotOptions[]) => {
	const slots = classifyEnchantSlots(player, gear, frozenSlots, presetPool);
	return { options: buildEnchantOptions(slots, weights, Class.ClassWarrior, !!presetPool), skipped: skippedEnchantSlots(slots) };
};

const fromPresets = (player: Player<any>, gear: Gear) => optimize(player, gear, [], WEIGHTS, enchantPresetPool(player));

describe('buildEnchantOptions and skippedEnchantSlots', () => {
	it('lists the fitting static enchants with EP above 0, ids sorted and deduplicated', () => {
		const duplicate = makeUIEnchant({ ...WRIST_STRENGTH, name: 'Enchant Bracer - Superior Strength (second row)' });
		const player = makePlayer({ enchants: [WRIST_STRENGTH, WRIST_SPIRIT, WRIST_AGILITY, duplicate] });

		const result = optimize(player, makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist) }), [], WEIGHTS);

		expect(result.options).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotWrist, effectIds: [247, 1885] })]);
		expect(result.skipped).toEqual([]);
	});

	it('skips a weapon that can take a combat proc', () => {
		const result = optimize(makePlayer(), makeGear({ [ItemSlot.ItemSlotMainHand]: {} }), [], WEIGHTS);

		expect(result.options).toEqual([]);
		expect(result.skipped).toEqual([{ slot: ItemSlot.ItemSlotMainHand, reason: 'proc' }]);
	});

	it('skips a slot that can take a survival proc for a tank spec only', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotChest]: armor(ItemType.ItemTypeChest) });

		const tank = optimize(makePlayer({ isTankSpec: true }), gear, [], WEIGHTS);
		const dps = optimize(makePlayer(), gear, [], WEIGHTS);

		expect(tank.options).toEqual([]);
		expect(tank.skipped).toEqual([{ slot: ItemSlot.ItemSlotChest, reason: 'proc' }]);
		expect(dps.options).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotChest, effectIds: [1891] })]);
		expect(dps.skipped).toEqual([]);
	});

	it('blocks on a combat proc only when the sim implements it', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotChest]: armor(ItemType.ItemTypeChest) });
		const unimplemented = enchant(7223, 'Enchant Chest - Retricutioner', ItemType.ItemTypeChest, { procKind: EnchantProcKind.EnchantProcKindCombat });
		const implemented = enchant(7224, 'Enchant Chest - Implemented Proc', ItemType.ItemTypeChest, { procKind: EnchantProcKind.EnchantProcKindCombat });

		const missing = optimize(makePlayer({ enchants: [unimplemented, CHEST_STATS] }), gear, [], WEIGHTS);
		const simulated = optimize(makePlayer({ enchants: [implemented, CHEST_STATS] }), gear, [], WEIGHTS);

		expect(missing).toEqual({ options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotChest, effectIds: [1891] })], skipped: [] });
		expect(simulated).toEqual({ options: [], skipped: [{ slot: ItemSlot.ItemSlotChest, reason: 'proc' }] });
	});

	it('leaves frozen and empty slots out of both lists', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist), [ItemSlot.ItemSlotMainHand]: {} });

		const result = optimize(makePlayer(), gear, [ItemSlot.ItemSlotWrist, ItemSlot.ItemSlotMainHand], WEIGHTS);

		expect(result).toEqual({ options: [], skipped: [] });
	});

	it('neither blocks on nor picks an enchant with a permanent effect and no stats', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotBack]: armor(ItemType.ItemTypeBack) });

		const withAgility = optimize(makePlayer(), gear, [], WEIGHTS);
		const threatOnly = optimize(makePlayer({ enchants: [BACK_THREAT] }), gear, [], WEIGHTS);

		expect(withAgility).toEqual({ options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [849] })], skipped: [] });
		expect(threatOnly).toEqual({ options: [], skipped: [] });
	});

	it('values a scope on a bow through ranged DPS, and finds nothing for a wand', () => {
		const bow = makeGear({ [ItemSlot.ItemSlotRanged]: BOW });

		const scoped = optimize(makePlayer(), bow, [], WEIGHTS);
		const noRangedWeight = optimize(makePlayer(), bow, [], WEIGHTS.withPseudoStat(PseudoStat.PseudoStatRangedDps, 0));
		const wand = optimize(makePlayer(), makeGear({ [ItemSlot.ItemSlotRanged]: WAND }), [], WEIGHTS);

		expect(scoped).toEqual({ options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotRanged, effectIds: [664] })], skipped: [] });
		expect(noRangedWeight).toEqual({ options: [], skipped: [] });
		expect(wand).toEqual({ options: [], skipped: [] });
	});
});

describe('buildEnchantOptions and skippedEnchantSlots from the checked enchant presets', () => {
	it('leaves a slot no checked preset lists alone, and reports it', () => {
		const player = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotWrist, [WRIST_STRENGTH.effectId]]] });
		const gear = makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist), [ItemSlot.ItemSlotBack]: armor(ItemType.ItemTypeBack) });

		expect(fromPresets(player, gear)).toEqual({
			options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotWrist, effectIds: [1885] })],
			skipped: [{ slot: ItemSlot.ItemSlotBack, reason: 'unlisted' }],
		});
	});

	it('says nothing about a slot whose item takes no enchant, even one a preset lists', () => {
		const player = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotRanged, [SCOPE.effectId]]] });

		expect(fromPresets(player, makeGear({ [ItemSlot.ItemSlotRanged]: WAND }))).toEqual({ options: [], skipped: [] });
	});

	it('reports a slot that takes enchants when the presets list none of them', () => {
		const player = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotWrist, [WRIST_STRENGTH.effectId]]] });

		expect(fromPresets(player, makeGear({ [ItemSlot.ItemSlotRanged]: BOW }))).toEqual({
			options: [],
			skipped: [{ slot: ItemSlot.ItemSlotRanged, reason: 'unlisted' }],
		});
	});

	it('leaves a slot to the sim when a preset lists an implemented proc beside static enchants', () => {
		const player = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotMainHand, [CRUSADER.effectId, WEAPON_AGILITY.effectId]]] });

		expect(fromPresets(player, makeGear({ [ItemSlot.ItemSlotMainHand]: {} }))).toEqual({
			options: [],
			skipped: [{ slot: ItemSlot.ItemSlotMainHand, reason: 'proc' }],
		});
	});

	it('blocks on a listed survival proc for a spec that does not tank', () => {
		const player = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotChest, [CHEST_ABSORPTION.effectId, CHEST_STATS.effectId]]] });

		expect(fromPresets(player, makeGear({ [ItemSlot.ItemSlotChest]: armor(ItemType.ItemTypeChest) }))).toEqual({
			options: [],
			skipped: [{ slot: ItemSlot.ItemSlotChest, reason: 'proc' }],
		});
	});

	it('offers exactly the listed static enchants, one worth 0 EP included', () => {
		const player = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotWrist, [WRIST_STRENGTH.effectId, WRIST_SPIRIT.effectId]]] });

		expect(fromPresets(player, makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist) }))).toEqual({
			options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotWrist, effectIds: [1883, 1885] })],
			skipped: [],
		});
	});

	it('applies the Finger 1 list to both rings and the Weapons list to both hands', () => {
		const player = makePresetPlayer(
			{
				Raid: [
					[ItemSlot.ItemSlotFinger1, [RING_STAMINA.effectId]],
					[ItemSlot.ItemSlotMainHand, [WEAPON_AGILITY.effectId]],
				],
			},
			{ enchants: [RING_STAMINA, WEAPON_AGILITY] },
		);
		const gear = makeGear({
			[ItemSlot.ItemSlotFinger1]: armor(ItemType.ItemTypeFinger),
			[ItemSlot.ItemSlotFinger2]: armor(ItemType.ItemTypeFinger),
			[ItemSlot.ItemSlotMainHand]: {},
			[ItemSlot.ItemSlotOffHand]: {},
		});

		expect(fromPresets(player, gear)).toEqual({
			options: [
				BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotFinger1, effectIds: [7301] }),
				BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotFinger2, effectIds: [7301] }),
				BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotMainHand, effectIds: [2564] }),
				BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotOffHand, effectIds: [2564] }),
			],
			skipped: [],
		});
	});

	it('neither blocks on nor offers a listed proc the sim does not implement', () => {
		const chest = makeGear({ [ItemSlot.ItemSlotChest]: armor(ItemType.ItemTypeChest) });
		const enchants = [...ALL_ENCHANTS, CHEST_UNIMPLEMENTED];

		const withStats = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotChest, [CHEST_UNIMPLEMENTED.effectId, CHEST_STATS.effectId]]] }, { enchants });
		const procOnly = makePresetPlayer({ Raid: [[ItemSlot.ItemSlotChest, [CHEST_UNIMPLEMENTED.effectId]]] }, { enchants });

		expect(fromPresets(withStats, chest)).toEqual({
			options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotChest, effectIds: [1891] })],
			skipped: [],
		});
		expect(fromPresets(procOnly, chest)).toEqual({ options: [], skipped: [] });
	});

	it('offers nothing while no preset is checked', () => {
		const player = makePresetPlayer({}, { unchecked: { Raid: [[ItemSlot.ItemSlotWrist, [WRIST_STRENGTH.effectId]]] } });

		expect(fromPresets(player, makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist) }))).toEqual({
			options: [],
			skipped: [{ slot: ItemSlot.ItemSlotWrist, reason: 'unlisted' }],
		});
	});
});
