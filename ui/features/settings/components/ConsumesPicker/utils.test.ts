import { iconEnumPickerShown } from '@features/inputs/components/ConnectedIconEnumPicker';
import { Class, ConsumableType, ConsumesSpec, HandType, ItemSlot, Profession, PseudoStat, Race, Spec, Stat, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ItemEffect } from '@generated/proto/spell';
import { UIItem } from '@generated/proto/ui';
import type { Player, SpecConfigData } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { type ConsumeConfig, consumeConfigs, type ConsumeToggle } from './utils';

const onUse = (categoryId: number, cooldownMs = 0) =>
	ItemEffect.create({ effect: { oneofKind: 'onUse', onUse: { categoryId, cooldownMs, categoryCooldownMs: 0 } } });
const statArray = (stat: Stat, value: number) => Object.assign(new Array<number>(42).fill(0), { [stat]: value });
const strength = (value: number) => statArray(Stat.StatStrength, value);
const grouped = (type: ConsumableType) => (id: number, name: string, stats: Array<number>, groupKey: string, groupName: string) =>
	Consumable.create({ id, name, type, stats, groupKey, groupName });
const elixir = grouped(ConsumableType.ConsumableTypeElixir);
const scroll = grouped(ConsumableType.ConsumableTypeScroll);
const other = grouped(ConsumableType.ConsumableTypeOther);
const meleeHit = (food: Consumable) =>
	Consumable.create({
		...food,
		pseudoStats: Object.assign(new Array<number>(36).fill(0), { [PseudoStat.PseudoStatMeleeHitPercent]: 2, [PseudoStat.PseudoStatRangedHitPercent]: 2 }),
	});
const during = (holiday: string, consumable: Consumable) => Consumable.create({ ...consumable, holiday });

const CONSUMABLES = [
	Consumable.create({
		id: 13442,
		name: 'Mighty Rage Potion',
		type: ConsumableType.ConsumableTypePotion,
		categoryId: 4,
		itemEffects: [onUse(4)],
		classAllowlist: [Class.ClassWarrior, Class.ClassDruid],
		stats: strength(60),
	}),
	Consumable.create({ id: 9421, name: 'Major Healthstone', type: ConsumableType.ConsumableTypeConjured, categoryId: 1153, itemEffects: [onUse(1153)] }),
	Consumable.create({ id: 13510, name: 'Flask of the Titans', type: ConsumableType.ConsumableTypeFlask, stats: strength(10) }),
	elixir(9206, 'Elixir of Greater Strength', strength(25), '29/0', 'Strength'),
	elixir(6662, 'Elixir of Giant Growth', strength(8), '29/0', 'Strength'),
	elixir(3383, 'Elixir of Wisdom', statArray(Stat.StatIntellect, 6), '29/3', 'Intellect'),
	elixir(12820, 'Winterfall Firewater', statArray(Stat.StatAttackPower, 35), '99/0', 'Attack Power'),
	grouped(ConsumableType.ConsumableTypeUnknown)(12451, 'Juju Power', strength(30), '29/0', 'Strength'),
	Consumable.create({
		id: 238637,
		name: 'Prowler Steak',
		type: ConsumableType.ConsumableTypeFood,
		categoryId: 11,
		itemEffects: [onUse(11)],
		stats: strength(20),
	}),
	Consumable.create({
		id: 18588,
		name: 'Ez-Thro Dynamite II',
		type: ConsumableType.ConsumableTypeExplosive,
		lvl: 30,
		categoryId: 24,
		itemEffects: [onUse(24)],
	}),
	Consumable.create({
		id: 10646,
		name: 'Goblin Sapper Charge',
		type: ConsumableType.ConsumableTypeExplosive,
		lvl: 41,
		categoryId: 24,
		itemEffects: [onUse(24, 300000)],
		requiredProfession: Profession.Engineering,
	}),
	Consumable.create({ id: 12404, name: 'Dense Sharpening Stone', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 33219 }),
	Consumable.create({ id: 228980, name: 'Conductive Shield Coating', type: ConsumableType.ConsumableTypeImbue, weaponInventoryTypeMask: 16384 }),
	scroll(2289, 'Scroll of Strength II', strength(9), '29/0', 'Strength'),
	scroll(10310, 'Scroll of Strength IV', strength(17), '29/0', 'Strength'),
	scroll(1180, 'Scroll of Stamina', statArray(Stat.StatStamina, 4), '29/2', 'Stamina'),
	other(8410, 'R.O.I.D.S.', strength(25), 'Zanza', 'Zanza'),
	other(20079, 'Spirit of Zanza', statArray(Stat.StatSpirit, 50), 'Zanza', 'Zanza'),
	during('Midsummer Fire Festival', meleeHit(other(23327, 'Fire-toasted Bun', [], 'Hit food', 'Hit food'))),
	during('Love is in the Air', meleeHit(other(22237, 'Dark Desire', [], 'Hit food', 'Hit food'))),
	during(
		'Midsummer Fire Festival',
		Consumable.create({ id: 23326, name: 'Midsummer Sausage', type: ConsumableType.ConsumableTypeOther, stats: statArray(Stat.StatSpellDamage, 23) }),
	),
	Consumable.create({ id: 4376, name: 'Flame Deflector', type: ConsumableType.ConsumableTypeDevice }),
];

const db = { getConsumables: () => CONSUMABLES, getSpellEffect: () => undefined } as unknown as Database;

interface PlayerState {
	cls: Class;
	spec: Spec;
	professions: Array<Profession>;
	epStats: Array<Stat>;
	displayStats: Array<UnitStat>;
	specConfig?: SpecConfigData<any>;
	windfury: boolean;
	flametongue: boolean;
	mainHand?: UIItem;
	offHand?: UIItem;
	consumes: ConsumesSpec;
	minLvl: Partial<Record<ConsumableType, number>>;
}

const playerOf = (state: Partial<PlayerState> = {}) => {
	const full: PlayerState = {
		cls: Class.ClassWarrior,
		spec: Spec.SpecDpsWarrior,
		professions: [],
		epStats: [Stat.StatStrength, Stat.StatAttackPower],
		displayStats: [],
		windfury: false,
		flametongue: false,
		consumes: ConsumesSpec.create(),
		minLvl: {},
		...state,
	};
	const player = {
		sim: { getMinConsumableLvl: (type: ConsumableType) => full.minLvl[type] ?? 0 },
		getClass: () => full.cls,
		getSpec: () => full.spec,
		getRace: () => Race.RaceOrc,
		hasProfession: (profession: Profession) => full.professions.includes(profession),
		getEquippedItem: (slot: ItemSlot) => {
			const item = slot === ItemSlot.ItemSlotMainHand ? full.mainHand : slot === ItemSlot.ItemSlotOffHand ? full.offHand : undefined;
			return item ? { item } : null;
		},
		getSpecConfig: () => full.specConfig ?? { epStats: full.epStats, displayStats: full.displayStats },
		getParty: () => ({ getBuffs: () => ({ windfuryTotem: full.windfury, flametongueTotem: full.flametongue }) }),
		getConsumes: () => ConsumesSpec.clone(full.consumes),
		setConsumes: (consumes: ConsumesSpec) => (full.consumes = ConsumesSpec.clone(consumes)),
	} as unknown as Player<any>;
	return { player, state: full };
};

const offered = (config: ConsumeConfig, player: Player<any>) =>
	config.values.filter(value => value.value !== 0 && (!value.showWhen || value.showWhen(player))).map(value => value.value);
const toggled = (toggles: ReadonlyArray<ConsumeToggle>, player: Player<any>) =>
	toggles.filter(toggle => !toggle.showWhen || toggle.showWhen(player)).map(toggle => toggle.actionId.itemId);

describe('consumeConfigs', () => {
	it('offers the Mighty Rage Potion to a class the game files allow it to', () => {
		const warrior = playerOf().player;
		expect(offered(consumeConfigs(warrior, db).potion, warrior)).toEqual([13442]);
		const paladin = playerOf({ cls: Class.ClassPaladin, spec: Spec.SpecRetributionPaladin }).player;
		expect(offered(consumeConfigs(paladin, db).potion, paladin)).toEqual([]);
	});

	it('puts each consumable in its picker', () => {
		const { player } = playerOf({ professions: [Profession.Engineering] });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.conjured, player)).toEqual([]);
		expect(offered(configs.flask, player)).toEqual([13510]);
		expect(configs.elixirs.map(config => offered(config, player))).toEqual([[9206, 6662, 12451, 2289, 10310], [], [12820]]);
		expect(configs.scrolls.map(config => offered(config, player))).toEqual([[]]);
		expect(configs.groups.map(config => offered(config, player))).toEqual([[8410]]);
		expect(offered(configs.food, player)).toEqual([238637]);
		expect(configs.holidayGroups.map(config => offered(config, player))).toEqual([[]]);
		expect(toggled(configs.holiday, player)).toEqual([]);
		expect(offered(configs.explosive, player)).toEqual([18588]);
		expect(toggled(configs.explosiveExtras, player)).toEqual([10646]);
		expect(toggled(configs.miscellaneous, player)).toEqual([4376]);
	});

	it('hides a consumable below the minimum level set for its type only', () => {
		const { player } = playerOf({ professions: [Profession.Engineering], minLvl: { [ConsumableType.ConsumableTypeExplosive]: 41 } });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.explosive, player)).toEqual([]);
		expect(toggled(configs.explosiveExtras, player)).toEqual([10646]);
		expect(offered(configs.food, player)).toEqual([238637]);
	});

	it('lists a picker by level, highest first', () => {
		const potion = (id: number, lvl: number) =>
			Consumable.create({ id, name: `Potion ${id}`, type: ConsumableType.ConsumableTypePotion, lvl, stats: strength(10) });
		const potions = { getConsumables: () => [potion(1, 20), potion(2, 50), potion(3, 35)], getSpellEffect: () => undefined } as unknown as Database;
		const { player } = playerOf();
		expect(offered(consumeConfigs(player, potions).potion, player)).toEqual([2, 3, 1]);
	});

	it('offers the Goblin Sapper Charge to an engineer only', () => {
		const { player } = playerOf();
		expect(toggled(consumeConfigs(player, db).explosiveExtras, player)).toEqual([]);
	});

	it('offers each hand the imbues its item takes, a shield coating on a shield', () => {
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const shield = UIItem.create({ weaponType: WeaponType.WeaponTypeShield, handType: HandType.HandTypeOffHand });
		const { player } = playerOf({ spec: Spec.SpecProtectionWarrior, mainHand: sword, offHand: shield });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.mhImbue, player)).toEqual([12404]);
		expect(offered(configs.ohImbue, player)).toEqual([228980]);
	});

	it('gives each hand a stone and a weapon imbue side by side, each picker offering only its kind', () => {
		const imbues = [
			Consumable.create({ id: 12404, name: 'Dense Sharpening Stone', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 33219 }),
			Consumable.create({
				id: 8928,
				name: 'Instant Poison VI',
				type: ConsumableType.ConsumableTypeWeaponImbue,
				weaponSubclassMask: 173555,
				classAllowlist: [Class.ClassRogue],
			}),
		];
		const imbueDb = { getConsumables: () => imbues, getSpellEffect: () => undefined } as unknown as Database;
		const dagger = UIItem.create({ weaponType: WeaponType.WeaponTypeDagger, handType: HandType.HandTypeOneHand });
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const { player, state } = playerOf({ cls: Class.ClassRogue, spec: Spec.SpecRogue, mainHand: dagger, offHand: sword });
		const configs = consumeConfigs(player, imbueDb);
		expect(offered(configs.mhImbue, player)).toEqual([12404]);
		expect(offered(configs.ohImbue, player)).toEqual([12404]);
		expect(offered(configs.mhWeaponImbue, player)).toEqual([8928]);
		expect(offered(configs.ohWeaponImbue, player)).toEqual([8928]);

		configs.mhImbue.setValue(player, 12404);
		configs.mhWeaponImbue.setValue(player, 8928);
		expect([state.consumes.mhImbueId, state.consumes.mhWeaponImbueId]).toEqual([12404, 8928]);
	});

	it('shows the main-hand imbue under Windfury Totem and Flametongue Totem', () => {
		const { player } = playerOf({ windfury: true, flametongue: true });
		const mhImbue = consumeConfigs(player, db).mhImbue;
		expect(mhImbue.showWhen?.(player) ?? true).toBe(true);
	});

	it('labels each group as db.json does and hides one that offers the player nothing', () => {
		const { player } = playerOf();
		const { elixirs, scrolls, groups } = consumeConfigs(player, db);
		expect(elixirs.map(config => config.tooltip)).toEqual(['Strength', 'Intellect', 'Attack Power']);
		expect(elixirs.map(config => iconEnumPickerShown(config, player))).toEqual([true, false, true]);
		expect(scrolls.map(config => config.tooltip)).toEqual(['Stamina']);
		expect(scrolls.map(config => iconEnumPickerShown(config, player))).toEqual([false]);
		expect(groups.map(config => config.tooltip)).toEqual(['Zanza']);
		expect(groups.map(config => iconEnumPickerShown(config, player))).toEqual([true]);
	});

	it('lists the holiday consumables apart: a group holding one as one pick, the rest each as a toggle', () => {
		const caster = { epStats: [Stat.StatSpellDamage], displayStats: [] } as unknown as SpecConfigData<any>;
		const { player } = playerOf({ cls: Class.ClassMage, spec: Spec.SpecMage, specConfig: caster });
		const configs = consumeConfigs(player, db);
		expect(configs.holidayGroups.map(config => config.tooltip)).toEqual(['Hit food']);
		expect(toggled(configs.holiday, player)).toEqual([23326]);
		expect(toggled(configs.miscellaneous, player)).not.toContain(23326);
	});

	it('shows the Hit food to a spec that shows melee hit', () => {
		const { player } = playerOf({ displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent)] });
		const [hitFood] = consumeConfigs(player, db).holidayGroups;
		expect(hitFood.tooltip).toBe('Hit food');
		expect(iconEnumPickerShown(hitFood, player)).toBe(true);
		expect(offered(hitFood, player)).toEqual([23327, 22237]);
	});

	it('offers Dark Desire to a spec that shows melee hit, not to one that weighs hit rating for spell hit alone', () => {
		const specPlayer = (cls: Class, spec: Spec, specConfig: SpecConfigData<any>) => playerOf({ cls, spec, specConfig }).player;
		const hitFoodOf = (player: Player<any>) => offered(consumeConfigs(player, db).holidayGroups[0], player);
		const caster = {
			epStats: [Stat.StatSpellDamage, Stat.StatHitRating],
			displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent)],
		} as SpecConfigData<any>;
		const melee = {
			epStats: [Stat.StatAttackPower, Stat.StatHitRating],
			displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent)],
		} as SpecConfigData<any>;
		expect(hitFoodOf(specPlayer(Class.ClassMage, Spec.SpecMage, caster))).toEqual([]);
		expect(hitFoodOf(specPlayer(Class.ClassRogue, Spec.SpecRogue, melee))).toContain(22237);
		expect(hitFoodOf(specPlayer(Class.ClassWarrior, Spec.SpecDpsWarrior, melee))).toContain(22237);
	});

	it('picks one elixir of each group, and any number of groups', () => {
		const { player, state } = playerOf();
		const [strengthGroup, , attackPowerGroup] = consumeConfigs(player, db).elixirs;
		strengthGroup.setValue(player, 9206);
		attackPowerGroup.setValue(player, 12820);
		expect(state.consumes.elixirIds).toEqual([9206, 12820]);
		strengthGroup.setValue(player, 6662);
		expect(state.consumes.elixirIds).toEqual([12820, 6662]);
		expect(strengthGroup.getValue(player)).toBe(6662);
		strengthGroup.setValue(player, 0);
		expect(state.consumes.elixirIds).toEqual([12820]);
	});

	it('keeps the elixirs beside a flask', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds: [9206] }) });
		const configs = consumeConfigs(player, db);
		configs.flask.setValue(player, 13510);
		configs.elixirs[2].setValue(player, 12820);
		expect([state.consumes.flaskId, state.consumes.elixirIds]).toEqual([13510, [9206, 12820]]);
	});

	it('picks a group member into the elixir ids or the extra ids by its type', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ extraIds: [4376] }) });
		const [strengthGroup] = consumeConfigs(player, db).elixirs;
		strengthGroup.setValue(player, 9206);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [4376]]);
		strengthGroup.setValue(player, 12451);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[], [4376, 12451]]);
		expect(strengthGroup.getValue(player)).toBe(12451);
		strengthGroup.setValue(player, 0);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[], [4376]]);
	});

	it('picks a scroll in the picker of its elixir group, in place of the elixir', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds: [9206] }) });
		const [strengthGroup] = consumeConfigs(player, db).elixirs;
		strengthGroup.setValue(player, 10310);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[], [10310]]);
		strengthGroup.setValue(player, 9206);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], []]);
	});

	it('picks a scroll no elixir shares and one of the Zanza group into the extra ids', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds: [9206] }) });
		const configs = consumeConfigs(player, db);
		configs.scrolls[0].setValue(player, 1180);
		configs.groups[0].setValue(player, 8410);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [1180, 8410]]);
		configs.groups[0].setValue(player, 20079);
		configs.scrolls[0].setValue(player, 0);
		expect(state.consumes.extraIds).toEqual([20079]);
	});

	it('offers every rank the player can use, lower ranks included', () => {
		const ranks = [
			scroll(2289, 'Scroll of Strength II', strength(9), '29/0', 'Strength'),
			Consumable.create({ ...scroll(10310, 'Scroll of Strength IV', strength(17), '29/0', 'Strength'), classAllowlist: [Class.ClassRogue] }),
			Consumable.create({ id: 12404, name: 'Dense Sharpening Stone II', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 1 << 7 }),
			Consumable.create({ id: 12643, name: 'Dense Sharpening Stone III', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 1 << 15 }),
			Consumable.create({ id: 8928, name: 'Instant Poison V', type: ConsumableType.ConsumableTypeWeaponImbue, lvl: 52, weaponSubclassMask: 1 << 15 }),
			Consumable.create({ id: 8927, name: 'Instant Poison IV', type: ConsumableType.ConsumableTypeWeaponImbue, lvl: 44, weaponSubclassMask: 1 << 15 }),
		];
		const rankDb = { getConsumables: () => ranks, getSpellEffect: () => undefined } as unknown as Database;
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const dagger = UIItem.create({ weaponType: WeaponType.WeaponTypeDagger, handType: HandType.HandTypeOneHand });

		const warrior = playerOf({ mainHand: sword }).player;
		const configs = consumeConfigs(warrior, rankDb);
		expect(offered(configs.scrolls[0], warrior)).toEqual([2289]);
		expect(offered(configs.mhImbue, warrior)).toEqual([12404]);

		const rogue = playerOf({ cls: Class.ClassRogue, mainHand: dagger }).player;
		const rogueConfigs = consumeConfigs(rogue, rankDb);
		expect(offered(rogueConfigs.scrolls[0], rogue)).toEqual([2289, 10310]);
		expect(offered(rogueConfigs.mhImbue, rogue)).toEqual([12643]);
		expect(offered(rogueConfigs.mhWeaponImbue, rogue)).toEqual([8928, 8927]);
	});

	it('keeps offering a picked consumable below the minimum level, and hides it once unpicked', () => {
		const { player, state } = playerOf({ minLvl: { [ConsumableType.ConsumableTypeExplosive]: 41 } });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.explosive, player)).toEqual([]);
		state.consumes = ConsumesSpec.create({ explosiveId: 18588 });
		expect(offered(configs.explosive, player)).toEqual([18588]);
		configs.explosive.setValue(player, 0);
		expect(offered(configs.explosive, player)).toEqual([]);
	});
});
