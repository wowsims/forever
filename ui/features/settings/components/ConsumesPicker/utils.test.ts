import { iconEnumPickerShown, iconEnumValues } from '@features/inputs/components/ConnectedIconEnumPicker/utils';
import { shownStats, withSelection } from '@features/stat-filters/model/stat_filters';
import { RaidBuffs } from '@generated/proto/buffs';
import {
	Class,
	ConsumableType,
	ConsumesSpec,
	HandType,
	ItemSlot,
	MobType,
	MobTypeStats,
	PotionClass,
	Profession,
	PseudoStat,
	Race,
	Spec,
	Stat,
	WeaponType,
	ZoneStats,
} from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { DruidTalents } from '@generated/proto/druid';
import { ItemEffect } from '@generated/proto/spell';
import { DatabaseFilters, UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { Player, SpecConfigData } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import { Stats, UnitStat } from '@sim/proto/stats';
import { ConsumableSlot, relevantConsumableIds } from '@sim/settings/consumables';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { consumableFilterScope } from '../../model/consumables';
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
const food = grouped(ConsumableType.ConsumableTypeFood);
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
		potionClasses: [PotionClass.PotionClassResource, PotionClass.PotionClassOffensive],
	}),
	Consumable.create({ id: 9421, name: 'Major Healthstone', type: ConsumableType.ConsumableTypeConjured, categoryId: 1153, itemEffects: [onUse(1153)] }),
	Consumable.create({ id: 13510, name: 'Flask of the Titans', type: ConsumableType.ConsumableTypeFlask, stats: strength(10) }),
	elixir(9206, 'Elixir of Greater Strength', strength(25), '29/0', 'Strength'),
	elixir(6662, 'Elixir of Giant Growth', strength(8), '29/0', 'Strength'),
	elixir(3383, 'Elixir of Wisdom', statArray(Stat.StatIntellect, 6), '29/3', 'Intellect'),
	elixir(12820, 'Winterfall Firewater', statArray(Stat.StatAttackPower, 35), '99/0', 'Attack Power'),
	Consumable.create({ id: 12451, name: 'Juju Power', type: ConsumableType.ConsumableTypeUnknown, stats: strength(30), countsWithElixirs: true }),
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
	scroll(2289, 'Scroll of Strength II', strength(9), 'scroll 29/0', 'Strength'),
	scroll(10310, 'Scroll of Strength IV', strength(17), 'scroll 29/0', 'Strength'),
	scroll(1180, 'Scroll of Stamina', statArray(Stat.StatStamina, 4), 'scroll 29/2', 'Stamina'),
	Consumable.create({ id: 12217, name: 'Dragonbreath Chili', type: ConsumableType.ConsumableTypeFood }),
	food(21151, 'Rumsey Rum Black Label', statArray(Stat.StatStamina, 15), 'Rum', 'Rum'),
	other(18269, 'Gordok Green Grog', statArray(Stat.StatStamina, 10), 'Rum', 'Rum'),
	other(8410, 'R.O.I.D.S.', strength(25), 'Zanza', 'Zanza'),
	other(20079, 'Spirit of Zanza', statArray(Stat.StatSpirit, 50), 'Zanza', 'Zanza'),
	during('Midsummer Fire Festival', meleeHit(other(23327, 'Fire-toasted Bun', [], 'Hit food', 'Hit food'))),
	during('Love is in the Air', meleeHit(other(22237, 'Dark Desire', [], 'Hit food', 'Hit food'))),
	during(
		'Midsummer Fire Festival',
		Consumable.create({ id: 23326, name: 'Midsummer Sausage', type: ConsumableType.ConsumableTypeOther, stats: statArray(Stat.StatSpellDamage, 23) }),
	),
	Consumable.create({ id: 4376, name: 'Flame Deflector', type: ConsumableType.ConsumableTypeDevice, potionClasses: [PotionClass.PotionClassDefensive] }),
	Consumable.create({ id: 1400, name: 'Swiftfeather Token', type: ConsumableType.ConsumableTypeOther, potionClasses: [PotionClass.PotionClassOffensive] }),
];

const db = { getConsumables: () => CONSUMABLES, getSpellEffect: () => undefined } as unknown as Database;

interface PlayerState {
	cls: Class;
	spec: Spec;
	professions: Array<Profession>;
	epStats: Array<Stat>;
	displayStats: Array<UnitStat>;
	specConfig?: SpecConfigData<any>;
	mainHand?: UIItem;
	offHand?: UIItem;
	swapEnabled: boolean;
	swapMainHand?: UIItem;
	swapOffHand?: UIItem;
	consumes: ConsumesSpec;
	minLvl: Partial<Record<ConsumableType, number>>;
	filters: DatabaseFilters;
	epWeights: Stats;
	mobType: MobType;
	meleeDps: boolean;
	tank: boolean;
	raidBuffs: RaidBuffs;
	talents?: DruidTalents;
}

const playerOf = (state: Partial<PlayerState> = {}) => {
	const full: PlayerState = {
		cls: Class.ClassWarrior,
		spec: Spec.SpecDpsWarrior,
		professions: [],
		epStats: [Stat.StatStrength, Stat.StatAttackPower],
		displayStats: [],
		swapEnabled: false,
		consumes: ConsumesSpec.create(),
		minLvl: {},
		filters: DatabaseFilters.create(),
		epWeights: new Stats(),
		mobType: MobType.MobTypeHumanoid,
		meleeDps: true,
		tank: false,
		raidBuffs: RaidBuffs.create(),
		...state,
	};
	const inHand = (slot: ItemSlot, mainHand?: UIItem, offHand?: UIItem) => {
		const item = slot === ItemSlot.ItemSlotMainHand ? mainHand : slot === ItemSlot.ItemSlotOffHand ? offHand : undefined;
		return item ? { item } : null;
	};
	const player = {
		sim: {
			getMinConsumableLvl: (type: ConsumableType) => full.minLvl[type] ?? 0,
			store: { getState: () => ({ sim: { filters: full.filters } }) },
			encounter: {
				getTarget: (index: number) => (index === 0 ? { mobType: full.mobType } : undefined),
			},
		},
		getClass: () => full.cls,
		getSpec: () => full.spec,
		getPlayerSpec: () => ({ isMeleeDpsSpec: full.meleeDps, isTankSpec: full.tank }),
		getRace: () => Race.RaceOrc,
		hasProfession: (profession: Profession) => full.professions.includes(profession),
		getEquippedItem: (slot: ItemSlot) => inHand(slot, full.mainHand, full.offHand),
		itemSwapSettings: {
			getEnableItemSwap: () => full.swapEnabled,
			getItem: (slot: ItemSlot) => inHand(slot, full.swapMainHand, full.swapOffHand),
		},
		getSpecConfig: () => full.specConfig ?? { epStats: full.epStats, displayStats: full.displayStats },
		getEpWeights: () => full.epWeights,
		getConsumes: () => ConsumesSpec.clone(full.consumes),
		setConsumes: (consumes: ConsumesSpec) => (full.consumes = ConsumesSpec.clone(consumes)),
		getRaid: () => ({ getBuffs: () => RaidBuffs.clone(full.raidBuffs) }),
		getTalents: () => full.talents ?? {},
	} as unknown as Player<any>;
	return { player, state: full };
};

const offered = (config: ConsumeConfig, player: Player<any>) =>
	iconEnumValues(config, player)
		.filter(value => value.value !== 0 && (!value.showWhen || value.showWhen(player)))
		.map(value => value.value);
const toggled = (toggles: ReadonlyArray<ConsumeToggle>, player: Player<any>) =>
	toggles.filter(toggle => !toggle.showWhen || toggle.showWhen(player)).map(toggle => toggle.actionId.itemId);
// Both halves of the miscellaneous row.
const misc = (configs: ReturnType<typeof consumeConfigs>) => [...configs.miscOffensive, ...configs.miscDefensive];
const marked = (config: ConsumeConfig | ConsumeToggle, player: Player<any>) =>
	config.type === 'icon' ? config.keptBy!(player, config.getValue(player)) : config.keptBy!(player, config.getValue(player));

describe('consumeConfigs', () => {
	it('offers the Mighty Rage Potion to a class the game files allow it to', () => {
		const warrior = playerOf().player;
		expect(offered(consumeConfigs(warrior, db).resourcePotion, warrior)).toEqual([13442]);
		const paladin = playerOf({ cls: Class.ClassPaladin, spec: Spec.SpecRetributionPaladin }).player;
		expect(offered(consumeConfigs(paladin, db).resourcePotion, paladin)).toEqual([]);
	});

	it('offers a potion in the picker of each of its classes only', () => {
		const { player } = playerOf();
		const configs = consumeConfigs(player, db);
		expect(offered(configs.resourcePotion, player)).toEqual([13442]);
		expect(offered(configs.defensivePotion, player)).toEqual([]);
		expect(offered(configs.offensivePotion, player)).toEqual([13442]);
	});

	it('puts each consumable in its picker', () => {
		const { player } = playerOf({ professions: [Profession.Engineering] });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.conjured, player)).toEqual([]);
		expect(offered(configs.flask, player)).toEqual([13510]);
		expect(configs.elixirs.map(config => offered(config, player))).toEqual([[6662, 9206], [], [12820]]);
		expect(configs.scrolls.map(config => offered(config, player))).toEqual([[2289, 10310], []]);
		expect(offered(configs.food, player)).toEqual([238637]);
		expect(configs.foodGroups.map(config => offered(config, player))).toEqual([[]]);
		expect(configs.miscGroups.map(config => offered(config, player))).toEqual([[8410]]);
		expect(configs.holidayGroups.map(config => offered(config, player))).toEqual([[]]);
		expect(toggled(configs.holiday, player)).toEqual([]);
		expect(offered(configs.explosive, player)).toEqual([18588]);
		expect(toggled(configs.explosiveExtras, player)).toEqual([10646]);
		expect(toggled(misc(configs), player)).toEqual([12451, 4376]);
	});

	it('hides a consumable below the minimum level set for its type only', () => {
		const { player } = playerOf({ professions: [Profession.Engineering], minLvl: { [ConsumableType.ConsumableTypeExplosive]: 41 } });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.explosive, player)).toEqual([]);
		expect(toggled(configs.explosiveExtras, player)).toEqual([10646]);
		expect(offered(configs.food, player)).toEqual([238637]);
	});

	describe('order', () => {
		const potion = (id: number, name: string, lvl: number, stats: Array<number> = []) =>
			Consumable.create({ id, name, type: ConsumableType.ConsumableTypePotion, lvl, stats, potionClasses: [PotionClass.PotionClassOffensive] });
		const dbOf = (consumables: Array<Consumable>) => ({ getConsumables: () => consumables, getSpellEffect: () => undefined }) as unknown as Database;
		const agility = (value: number) => statArray(Stat.StatAgility, value);
		const epStats = [Stat.StatStrength, Stat.StatAgility];

		it('lists a picker by the EP each adds at the player weights, none first and one adding no stat last', () => {
			const potions = dbOf([
				potion(1, 'Proc', 60),
				potion(2, 'Low', 10, strength(5)),
				potion(3, 'Agile', 50, agility(10)),
				potion(4, 'Strong', 20, strength(10)),
			]);
			const { player } = playerOf({ epStats, epWeights: Stats.fromMap({ [Stat.StatStrength]: 2, [Stat.StatAgility]: 1 }) });
			const config = consumeConfigs(player, potions).offensivePotion;
			expect(iconEnumValues(config, player)[0].value).toBe(0);
			expect(offered(config, player)).toEqual([4, 3, 2, 1]);
		});

		it('orders an equal EP by level, highest first, then by name', () => {
			const potions = dbOf([potion(1, 'Bravo', 30, strength(10)), potion(2, 'Alpha', 30, strength(10)), potion(3, 'Charlie', 40, strength(10))]);
			const { player } = playerOf();
			expect(offered(consumeConfigs(player, potions).offensivePotion, player)).toEqual([3, 2, 1]);
		});

		it('orders again when the weights change, on the configs already built', () => {
			const potions = dbOf([potion(1, 'Strong', 0, strength(10)), potion(2, 'Agile', 0, agility(10))]);
			const { player, state } = playerOf({ epStats, epWeights: Stats.fromMap({ [Stat.StatStrength]: 1 }) });
			const config = consumeConfigs(player, potions).offensivePotion;
			expect(offered(config, player)).toEqual([1, 2]);
			state.epWeights = Stats.fromMap({ [Stat.StatAgility]: 1 });
			expect(offered(config, player)).toEqual([2, 1]);
			expect(iconEnumValues(config, player)[0].value).toBe(0);
		});

		it("orders again when the target's mob type changes, on the configs already built, and counts a zone stat in every fight", () => {
			const slaying = Consumable.create({
				...potion(1, 'Slaying', 0),
				mobTypeStats: [MobTypeStats.create({ mobTypes: [MobType.MobTypeDemon], stats: { [Stat.StatStrength]: 20 } })],
			});
			const natural = Consumable.create({
				...potion(2, 'Natural', 0, strength(5)),
				zoneStats: [ZoneStats.create({ zoneIds: [101, 102], stats: { [Stat.StatStrength]: 20 } })],
			});
			const potions = dbOf([slaying, natural, potion(3, 'Strong', 0, strength(10))]);
			const { player, state } = playerOf({ epWeights: Stats.fromMap({ [Stat.StatStrength]: 1 }) });
			const config = consumeConfigs(player, potions).offensivePotion;
			expect(offered(config, player)).toEqual([2, 3, 1]);
			state.mobType = MobType.MobTypeDemon;
			expect(offered(config, player)).toEqual([2, 1, 3]);
			state.mobType = MobType.MobTypeMechanical;
			expect(offered(config, player)).toEqual([2, 3, 1]);
		});
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

	it('hides the swap imbues while item swap is off, even with a swap weapon that takes them', () => {
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const { player } = playerOf({ mainHand: sword, swapMainHand: sword });
		const configs = consumeConfigs(player, db);
		expect(iconEnumPickerShown(configs.mhImbue, player)).toBe(true);
		expect(iconEnumPickerShown(configs.swapMhImbue, player)).toBe(false);
		expect(iconEnumPickerShown(configs.swapOhImbue, player)).toBe(false);
		expect(iconEnumPickerShown(configs.swapMhWeaponImbue, player)).toBe(false);
		expect(iconEnumPickerShown(configs.swapOhWeaponImbue, player)).toBe(false);
	});

	it('offers each swap hand the imbues its swap weapon takes, and picks them into the swap fields', () => {
		const mace = UIItem.create({ weaponType: WeaponType.WeaponTypeMace, handType: HandType.HandTypeOneHand });
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const shield = UIItem.create({ weaponType: WeaponType.WeaponTypeShield, handType: HandType.HandTypeOffHand });
		const { player, state } = playerOf({ mainHand: mace, swapEnabled: true, swapMainHand: sword, swapOffHand: shield });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.mhImbue, player)).toEqual([]);
		expect(offered(configs.swapMhImbue, player)).toEqual([12404]);
		expect(offered(configs.swapOhImbue, player)).toEqual([228980]);

		configs.swapMhImbue.setValue(player, 12404);
		configs.swapOhImbue.setValue(player, 228980);
		expect([state.consumes.mhImbueId, state.consumes.swapMhImbueId, state.consumes.swapOhImbueId]).toEqual([0, 12404, 228980]);

		state.swapMainHand = undefined;
		expect(iconEnumPickerShown(configs.swapMhImbue, player)).toBe(false);
		expect(iconEnumPickerShown(configs.swapOhImbue, player)).toBe(true);
	});

	it('offers the swap weapon imbues by the swap weapon, apart from the main hand', () => {
		const poison = Consumable.create({
			id: 8928,
			name: 'Instant Poison VI',
			type: ConsumableType.ConsumableTypeWeaponImbue,
			weaponSubclassMask: 1 << 15,
			classAllowlist: [Class.ClassRogue],
		});
		const poisonDb = { getConsumables: () => [poison], getSpellEffect: () => undefined } as unknown as Database;
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const dagger = UIItem.create({ weaponType: WeaponType.WeaponTypeDagger, handType: HandType.HandTypeOneHand });
		const { player, state } = playerOf({ cls: Class.ClassRogue, spec: Spec.SpecRogue, mainHand: sword, swapEnabled: true, swapMainHand: dagger });
		const configs = consumeConfigs(player, poisonDb);
		expect(offered(configs.mhWeaponImbue, player)).toEqual([]);
		expect(offered(configs.swapMhWeaponImbue, player)).toEqual([8928]);

		configs.swapMhWeaponImbue.setValue(player, 8928);
		expect([state.consumes.mhWeaponImbueId, state.consumes.swapMhWeaponImbueId]).toEqual([0, 8928]);
	});

	it('labels each group as db.json does and hides one that offers the player nothing', () => {
		const { player } = playerOf();
		const { elixirs, scrolls, foodGroups, miscGroups } = consumeConfigs(player, db);
		expect(elixirs.map(config => config.tooltip)).toEqual(['Strength', 'Intellect', 'Attack Power']);
		expect(elixirs.map(config => iconEnumPickerShown(config, player))).toEqual([true, false, true]);
		expect(scrolls.map(config => config.tooltip)).toEqual(['Strength', 'Stamina']);
		expect(scrolls.map(config => iconEnumPickerShown(config, player))).toEqual([true, false]);
		expect(foodGroups.map(config => config.tooltip)).toEqual(['Rum']);
		expect(foodGroups.map(config => iconEnumPickerShown(config, player))).toEqual([false]);
		expect(miscGroups.map(config => config.tooltip)).toEqual(['Zanza']);
		expect(miscGroups.map(config => iconEnumPickerShown(config, player))).toEqual([true]);
	});

	it('picks one rum, a food or not, beside the food and the misc toggles', () => {
		const { player, state } = playerOf({ epStats: [Stat.StatStrength, Stat.StatStamina] });
		const configs = consumeConfigs(player, db);
		const [rum] = configs.foodGroups;
		expect(offered(rum, player)).toEqual([18269, 21151]);
		expect(offered(configs.food, player)).toEqual([238637]);
		expect(toggled(misc(configs), player)).not.toContain(18269);
		configs.food.setValue(player, 238637);
		rum.setValue(player, 21151);
		rum.setValue(player, 18269);
		expect([state.consumes.foodId, state.consumes.extraIds]).toEqual([238637, [18269]]);
	});

	it('offers what grants a stat the spec displays, as the buffs do, though its EP leaves the stat out', () => {
		const weighed = playerOf().player;
		expect(iconEnumPickerShown(consumeConfigs(weighed, db).foodGroups[0], weighed)).toBe(false);
		const { player } = playerOf({ displayStats: [UnitStat.fromStat(Stat.StatStamina)] });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.foodGroups[0], player)).toEqual([18269, 21151]);
		expect(configs.scrolls.map(config => offered(config, player))).toEqual([[2289, 10310], [1180]]);
	});

	it('splits the miscellaneous row as the generator classes each consumable', () => {
		const { player } = playerOf();
		const { miscOffensive, miscDefensive } = consumeConfigs(player, db);
		expect(miscDefensive.map(toggle => toggle.consumable.id)).toEqual([4376]);
		expect(miscOffensive.map(toggle => toggle.consumable.id)).toEqual([12451, 1400]);
	});

	it('offers Dragonbreath Chili beside the food to the specs that hit in melee', () => {
		const chiliOf = (player: Player<any>) => consumeConfigs(player, db).dragonbreathChili!;
		const melee = playerOf();
		expect(offered(consumeConfigs(melee.player, db).food, melee.player)).not.toContain(12217);
		expect(toggled([chiliOf(melee.player)], melee.player)).toEqual([12217]);
		const tank = playerOf({ meleeDps: false, tank: true }).player;
		expect(toggled([chiliOf(tank)], tank)).toEqual([12217]);
		const hunter = playerOf({ cls: Class.ClassHunter, spec: Spec.SpecHunter, meleeDps: false }).player;
		expect(toggled([chiliOf(hunter)], hunter)).toEqual([12217]);
		const caster = playerOf({ cls: Class.ClassMage, spec: Spec.SpecMage, meleeDps: false }).player;
		expect(toggled([chiliOf(caster)], caster)).toEqual([]);

		chiliOf(melee.player).setValue(melee.player, true);
		expect(melee.state.consumes.extraIds).toEqual([12217]);
	});

	it('lists the holiday consumables apart: a group holding one as one pick, the rest each as a toggle', () => {
		const caster = { epStats: [Stat.StatSpellDamage], displayStats: [] } as unknown as SpecConfigData<any>;
		const { player } = playerOf({ cls: Class.ClassMage, spec: Spec.SpecMage, specConfig: caster });
		const configs = consumeConfigs(player, db);
		expect(configs.holidayGroups.map(config => config.tooltip)).toEqual(['Hit food']);
		expect(toggled(configs.holiday, player)).toEqual([23326]);
		expect(toggled(misc(configs), player)).not.toContain(23326);
	});

	it('shows the Hit food to a spec that shows melee hit', () => {
		const { player } = playerOf({ displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent)] });
		const [hitFood] = consumeConfigs(player, db).holidayGroups;
		expect(hitFood.tooltip).toBe('Hit food');
		expect(iconEnumPickerShown(hitFood, player)).toBe(true);
		expect(offered(hitFood, player)).toEqual([22237, 23327]);
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

	it('picks an elixir group member into the elixir ids, and Juju Power beside it from the miscellaneous row into the extra ids', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ extraIds: [4376] }) });
		const configs = consumeConfigs(player, db);
		const [strengthGroup] = configs.elixirs;
		const juju = misc(configs).find(toggle => toggle.consumable.id === 12451)!;
		strengthGroup.setValue(player, 9206);
		juju.setValue(player, true);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [4376, 12451]]);
		juju.setValue(player, false);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [4376]]);

		state.consumes = ConsumesSpec.create({ elixirIds: [9206, 12451] });
		expect(juju.getValue(player)).toBe(true);
		juju.setValue(player, false);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], []]);
	});

	it('picks a scroll beside the elixirs of its stat', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds: [9206] }) });
		const [strengthScrolls] = consumeConfigs(player, db).scrolls;
		strengthScrolls.setValue(player, 10310);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [10310]]);
		strengthScrolls.setValue(player, 2289);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [2289]]);
	});

	it('picks a scroll of each stat and one of the Zanza group into the extra ids', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds: [9206] }) });
		const configs = consumeConfigs(player, db);
		configs.scrolls[1].setValue(player, 1180);
		configs.miscGroups[0].setValue(player, 8410);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [1180, 8410]]);
		configs.miscGroups[0].setValue(player, 20079);
		configs.scrolls[1].setValue(player, 0);
		expect(state.consumes.extraIds).toEqual([20079]);
	});

	it('offers every rank the player can use, lower ranks included', () => {
		const ranks = [
			scroll(2289, 'Scroll of Strength II', strength(9), 'scroll 29/0', 'Strength'),
			Consumable.create({ ...scroll(10310, 'Scroll of Strength IV', strength(17), 'scroll 29/0', 'Strength'), classAllowlist: [Class.ClassRogue] }),
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

describe('consumeConfigs under the consumables stat filter', () => {
	const agility = statArray(Stat.StatAgility, 25);
	const agilityDb = {
		getConsumables: () => [
			...CONSUMABLES,
			elixir(9187, 'Elixir of Greater Agility', agility, '29/1', 'Agility'),
			Consumable.create({ id: 21023, name: 'Agility Snack', type: ConsumableType.ConsumableTypeOther, stats: agility }),
		],
		getSpellEffect: () => undefined,
	} as unknown as Database;
	const agilityWarrior = (state: Partial<PlayerState> = {}) => playerOf({ epStats: [Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower], ...state });
	const hiding = (player: Player<any>, database: Database, stat: Stat) => {
		const unset = DatabaseFilters.create();
		const shown = shownStats(unset, 'consumables', consumableFilterScope(player, database));
		return withSelection(
			unset,
			'consumables',
			shown.filter(unitStat => !unitStat.equals(UnitStat.fromStat(stat))),
		);
	};
	const agilityElixirs = (player: Player<any>) => offered(consumeConfigs(player, agilityDb).elixirs[1], player);

	it('offers with the slot set to the spec defaults what it offers with the slot unset', () => {
		const { player, state } = playerOf({ professions: [Profession.Engineering] });
		const configs = consumeConfigs(player, db);
		const everyOffer = () =>
			Object.values(configs)
				.flat()
				.map((config: ConsumeConfig | ConsumeToggle) => ('values' in config ? offered(config, player) : !config.showWhen || config.showWhen(player)));
		const unset = everyOffer();

		state.filters = withSelection(DatabaseFilters.create(), 'consumables', consumableFilterScope(player, db).defaults);

		expect(unset.flat().filter(id => typeof id === 'number')).not.toHaveLength(0);
		expect(everyOffer()).toEqual(unset);
	});

	it('drops a consumable whose stats the filter hides, read live without rebuilding the configs', () => {
		const { player, state } = agilityWarrior();
		const configs = consumeConfigs(player, agilityDb);
		expect(offered(configs.elixirs[1], player)).toEqual([9187]);
		expect(toggled(misc(configs), player)).toContain(21023);

		state.filters = hiding(player, agilityDb, Stat.StatAgility);
		expect(offered(configs.elixirs[1], player)).toEqual([]);
		expect(toggled(misc(configs), player)).not.toContain(21023);
		expect(offered(configs.elixirs[0], player)).toEqual([6662, 9206]);
	});

	it('reads the filters without copying them on each offer check', () => {
		const { player } = agilityWarrior();
		const configs = consumeConfigs(player, agilityDb);
		const clone = vi.spyOn(DatabaseFilters, 'clone');

		offered(configs.elixirs[1], player);
		toggled(misc(configs), player);

		expect(clone).not.toHaveBeenCalled();
		clone.mockRestore();
	});

	it('keeps a picked consumable whose stats the filter hides offered and marked, and in the consumables the sim gets', () => {
		const { player, state } = agilityWarrior({ consumes: ConsumesSpec.create({ elixirIds: [9187], extraIds: [21023] }) });
		const configs = consumeConfigs(player, agilityDb);
		expect(marked(configs.elixirs[1], player)).toBe(false);

		state.filters = hiding(player, agilityDb, Stat.StatAgility);
		expect(offered(configs.elixirs[1], player)).toEqual([9187]);
		expect(marked(configs.elixirs[1], player)).toBe(true);
		const snack = misc(configs).find(toggle => toggle.actionId.itemId === 21023)!;
		expect(snack.showWhen!(player)).toBe(true);
		expect(marked(snack, player)).toBe(true);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9187], [21023]]);

		configs.elixirs[1].setValue(player, 0);
		expect(agilityElixirs(player)).toEqual([]);
		expect(marked(configs.elixirs[1], player)).toBe(false);
	});

	it('keeps a picked consumable the class may not use hidden, whatever the filter shows', () => {
		const { player, state } = playerOf({
			cls: Class.ClassPaladin,
			spec: Spec.SpecRetributionPaladin,
			consumes: ConsumesSpec.create({ resourcePotId: 13442 }),
		});
		const configs = consumeConfigs(player, db);
		expect(offered(configs.resourcePotion, player)).toEqual([]);
		state.filters = withSelection(DatabaseFilters.create(), 'consumables', consumableFilterScope(player, db).universe);
		expect(offered(configs.resourcePotion, player)).toEqual([]);
	});

	it('marks a pick the level filter keeps, and unmarks it once the level shows it', () => {
		const { player, state } = playerOf({ minLvl: { [ConsumableType.ConsumableTypeExplosive]: 41 }, consumes: ConsumesSpec.create({ explosiveId: 18588 }) });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.explosive, player)).toEqual([18588]);
		expect(marked(configs.explosive, player)).toBe(true);

		state.minLvl = {};
		expect(marked(configs.explosive, player)).toBe(false);
	});

	it('marks a toggle the level filter keeps', () => {
		const { player, state } = playerOf({ professions: [Profession.Engineering], consumes: ConsumesSpec.create({ extraIds: [10646] }) });
		const sapper = consumeConfigs(player, db).explosiveExtras.find(toggle => toggle.actionId.itemId === 10646)!;
		expect(marked(sapper, player)).toBe(false);

		state.minLvl = { [ConsumableType.ConsumableTypeExplosive]: 42 };
		expect(sapper.showWhen!(player)).toBe(true);
		expect(marked(sapper, player)).toBe(true);
	});

	it('leaves the potions a sim request lists to the spec stats', () => {
		const { player, state } = playerOf();
		const configs = consumeConfigs(player, db);
		expect(relevantConsumableIds(CONSUMABLES, ConsumableSlot.Potion, player, db)).toEqual([13442]);

		state.filters = hiding(player, db, Stat.StatStrength);
		expect(offered(configs.offensivePotion, player)).toEqual([]);
		expect(relevantConsumableIds(CONSUMABLES, ConsumableSlot.Potion, player, db)).toEqual([13442]);
	});

	it('lets the filter show a stat the spec does not weigh, and offers its consumables', () => {
		const { player, state } = playerOf();
		expect(agilityElixirs(player)).toEqual([]);
		state.filters = withSelection(DatabaseFilters.create(), 'consumables', [UnitStat.fromStat(Stat.StatAgility)]);
		expect(agilityElixirs(player)).toEqual([9187]);
	});
});

describe('consumeConfigs on picks that add nothing', () => {
	const stats = (amounts: Partial<Record<Stat, number>>) => Object.assign(new Array<number>(42).fill(0), amounts);
	const crit = Object.assign(new Array<number>(36).fill(0), {
		[PseudoStat.PseudoStatMeleeCritPercent]: 2,
		[PseudoStat.PseudoStatSpellCritPercent]: 2,
		[PseudoStat.PseudoStatRangedCritPercent]: 2,
	});
	const critElixir = (id: number, name: string, granted: Array<number>) =>
		Consumable.create({ ...elixir(id, name, granted, '290/0', 'Critical Strike'), pseudoStats: crit });
	const stackingDb = {
		getConsumables: () => [
			elixir(250350, 'Elixir of Ferocity', stats({ [Stat.StatStrength]: 18, [Stat.StatAgility]: 18 }), '29/0', 'Strength'),
			critElixir(13452, 'Elixir of the Mongoose', stats({ [Stat.StatAgility]: 25 })),
			Consumable.create({ id: 12451, name: 'Juju Power', type: ConsumableType.ConsumableTypeUnknown, stats: strength(30), countsWithElixirs: true }),
			critElixir(250351, 'Elixir of the Grizzly', strength(25)),
			elixir(9206, 'Elixir of Greater Strength', strength(25), '29/0', 'Strength'),
			elixir(6662, 'Elixir of Giant Growth', strength(8), '29/0', 'Strength'),
			elixir(9187, 'Elixir of Greater Agility', stats({ [Stat.StatAgility]: 25 }), '29/1', 'Agility'),
			elixir(247755, 'Elixir of Minor Force', stats({ [Stat.StatPhysicalDamage]: 1 }), '13/1', 'Physical Damage'),
			Consumable.create({
				id: 5206,
				name: 'Bogling Root',
				type: ConsumableType.ConsumableTypeOther,
				stats: stats({ [Stat.StatPhysicalDamage]: 1 }),
				countsWithElixirs: true,
			}),
			elixir(22192, 'Bloodkelp Elixir of Dodging', [], '0/0', 'Other'),
			elixir(3383, 'Elixir of Wisdom', stats({ [Stat.StatIntellect]: 6 }), '29/3', 'Intellect'),
			elixir(13447, 'Elixir of the Sages', stats({ [Stat.StatIntellect]: 18, [Stat.StatSpirit]: 18 }), '29/3', 'Intellect'),
			elixir(13453, 'Elixir of Brute Force', stats({ [Stat.StatStrength]: 18, [Stat.StatStamina]: 18 }), '29/0', 'Strength'),
			elixir(250329, 'Elixir of the Phalanx', stats({ [Stat.StatHealth]: 400, [Stat.StatArmorNoMods]: 500 }), '34/0', 'Health'),
			elixir(250335, 'Elixir of Greater Fortitude', stats({ [Stat.StatHealth]: 400 }), '34/0', 'Health'),
			elixir(13454, 'Greater Arcane Elixir', stats({ [Stat.StatSpellDamage]: 35 }), '13/126', 'Spell Power'),
			elixir(13445, 'Elixir of Greater Defense', stats({ [Stat.StatArmorNoMods]: 450 }), '674/1', 'Armor'),
			scroll(10310, 'Scroll of Strength IV', strength(17), 'scroll 29/0', 'Strength'),
			scroll(10308, 'Scroll of Intellect IV', stats({ [Stat.StatIntellect]: 16 }), 'scroll 29/3', 'Intellect'),
		],
		getSpellEffect: () => undefined,
	} as unknown as Database;
	const picking = (
		elixirIds: Array<number>,
		extraIds: Array<number> = [],
		raidBuffs = RaidBuffs.create(),
		cls = Class.ClassWarrior,
		more: Partial<PlayerState> = {},
	) => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds, extraIds }), raidBuffs, cls, ...more });
		const configs = consumeConfigs(player, stackingDb);
		const groupNamed = (name: string) => configs.elixirs.find(config => config.tooltip === name)!;
		// The badge of the consumable's dropdown, undefined unless it is the pick its group shows, or of its toggle in the
		// miscellaneous row, undefined unless it is picked.
		const shownAs = (id: number) => {
			const toggle = misc(configs).find(config => config.consumable.id === id);
			if (toggle) return toggle.getValue(player) ? toggle.overwritten!(player, true) : undefined;
			const config = configs.elixirs.find(group => iconEnumValues(group, player).some(value => value.value === id))!;
			const value = config.getValue(player);
			return value === id ? config.overwritten!(player, value) : undefined;
		};
		const overwrittenBy = (id: number) => shownAs(id)?.tooltip;
		const partly = (id: number) => shownAs(id)?.partly ?? false;
		const [strengthScrolls, intellectScrolls] = configs.scrolls;
		const scrollOverwrittenBy = (scrolls = strengthScrolls) => scrolls.overwritten!(player, scrolls.getValue(player))?.tooltip;
		const intellectScrollOverwrittenBy = () => scrollOverwrittenBy(intellectScrolls);
		return { player, state, configs, groupNamed, overwrittenBy, partly, scrollOverwrittenBy, intellectScrollOverwrittenBy };
	};

	// The test setup loads no locale, so the tooltip would read as its key without the names.
	beforeAll(() => {
		i18n.addResourceBundle('en', 'translation', english, true, true);
	});
	afterAll(() => {
		i18n.removeResourceBundle('en', 'translation');
	});

	it("orders a warrior's elixir pickers as its character stats panel shows the stat each group shares, the rest last in key order", () => {
		const displayStats = [
			...[Stat.StatHealth, Stat.StatStrength, Stat.StatAgility, Stat.StatStamina, Stat.StatAttackPower, Stat.StatArmor].map(stat =>
				UnitStat.fromStat(stat),
			),
			...[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent].map(stat => UnitStat.fromPseudoStat(stat)),
		];
		const { configs } = picking([], [], RaidBuffs.create(), Class.ClassWarrior, { displayStats });
		expect(configs.elixirs.map(config => config.tooltip)).toEqual([
			'Strength',
			'Agility',
			'Critical Strike',
			'Armor',
			'Health',
			'Other',
			'Physical Damage',
			'Spell Power',
			'Intellect',
		]);
	});

	it("orders a mage's elixir pickers with Intellect, spell power and crit first and Strength late", () => {
		const displayStats = [
			...[Stat.StatHealth, Stat.StatMana, Stat.StatStamina, Stat.StatIntellect, Stat.StatSpirit, Stat.StatSpellDamage].map(stat =>
				UnitStat.fromStat(stat),
			),
			...[PseudoStat.PseudoStatSpellHitPercent, PseudoStat.PseudoStatSpellCritPercent].map(stat => UnitStat.fromPseudoStat(stat)),
		];
		const { configs } = picking([], [], RaidBuffs.create(), Class.ClassMage, { spec: Spec.SpecMage, displayStats });
		expect(configs.elixirs.map(config => config.tooltip)).toEqual([
			'Intellect',
			'Spell Power',
			'Critical Strike',
			'Health',
			'Other',
			'Physical Damage',
			'Strength',
			'Agility',
			'Armor',
		]);
	});

	it('marks Elixir of Ferocity beside Elixir of Greater Agility and Elixir of the Grizzly, naming both', () => {
		const { overwrittenBy, partly } = picking([250350, 9187, 250351]);
		expect(overwrittenBy(250350)).toBe('Strength overwritten by Elixir of the Grizzly\nAgility overwritten by Elixir of Greater Agility');
		expect(partly(250350)).toBe(false);
		expect(overwrittenBy(9187)).toBeUndefined();
		expect(overwrittenBy(250351)).toBeUndefined();
	});

	it('marks Elixir of Ferocity beside Elixir of the Mongoose as partly overwritten, naming the Agility it loses', () => {
		const { overwrittenBy, partly } = picking([250350, 13452]);
		expect(overwrittenBy(250350)).toBe('Agility overwritten by Elixir of the Mongoose');
		expect(partly(250350)).toBe(true);
		expect(overwrittenBy(13452)).toBeUndefined();
		expect(partly(13452)).toBe(false);
	});

	it('marks Elixir of Greater Strength beside Elixir of the Grizzly, which grants as much Strength and a crit besides', () => {
		const { overwrittenBy } = picking([250351, 9206]);
		expect(overwrittenBy(9206)).toBe('Strength overwritten by Elixir of the Grizzly');
		expect(overwrittenBy(250351)).toBeUndefined();
	});

	it('marks Elixir of the Grizzly beside Juju Power as partly overwritten, its crit still counting', () => {
		const { overwrittenBy, partly } = picking([250351], [12451]);
		expect(overwrittenBy(250351)).toBe('Strength overwritten by Juju Power');
		expect(partly(250351)).toBe(true);
		expect(overwrittenBy(12451)).toBeUndefined();
	});

	it("answers for each option of a group as if it were picked in place of the group's pick", () => {
		const { player, state, groupNamed } = picking([250350, 13452]);
		const [strength, agility, crit] = ['Strength', 'Agility', 'Critical Strike'].map(groupNamed);
		expect(agility.overwritten!(player, 9187)).toEqual({ tooltip: 'Agility overwritten by Elixir of the Mongoose', partly: false });
		expect(strength.overwritten!(player, 9206)).toBeUndefined();
		expect(strength.overwritten!(player, 250350)).toEqual({ tooltip: 'Agility overwritten by Elixir of the Mongoose', partly: true });
		expect(crit.overwritten!(player, 250351)).toBeUndefined();
		expect(strength.overwritten!(player, 0)).toBeUndefined();
		expect(state.consumes.elixirIds).toEqual([250350, 13452]);

		const grizzly = picking([250351]);
		expect(grizzly.groupNamed('Strength').overwritten!(grizzly.player, 9206)?.tooltip).toBe('Strength overwritten by Elixir of the Grizzly');
		expect(grizzly.groupNamed('Strength').overwritten!(grizzly.player, 12451)).toBeUndefined();
		expect(grizzly.groupNamed('Strength').overwritten!(grizzly.player, 250350)?.tooltip).toBe('Strength overwritten by Elixir of the Grizzly');
	});

	it('badges a miscellaneous toggle that counts with the elixirs as if it were picked, before it is', () => {
		const { player, configs } = picking([247755, 250351]);
		const toggleOf = (id: number) => misc(configs).find(toggle => toggle.consumable.id === id)!;
		expect(toggleOf(5206).getValue(player)).toBe(false);
		expect(toggleOf(5206).overwritten!(player, false)).toEqual({ tooltip: 'Bonus Weapon Damage overwritten by Elixir of Minor Force', partly: false });
		expect(toggleOf(12451).overwritten!(player, false)).toBeUndefined();
		expect(
			misc(configs)
				.filter(toggle => toggle.overwritten)
				.map(toggle => toggle.consumable.id),
		).toEqual([12451, 5206]);
	});

	it('reads only the first pick of each group, as the sim selects them', () => {
		const twoCrits = picking([250351, 13452, 9187]);
		expect(twoCrits.groupNamed('Critical Strike').getValue(twoCrits.player)).toBe(250351);
		expect(twoCrits.overwrittenBy(9187)).toBeUndefined();
		expect(twoCrits.overwrittenBy(250351)).toBeUndefined();
		expect(picking([247755], [5206]).overwrittenBy(247755)).toBeUndefined();
	});

	it('marks no elixir that is not picked, nor one granting nothing the picker reads', () => {
		const { overwrittenBy } = picking([22192, 9206, 250351]);
		expect(overwrittenBy(6662)).toBeUndefined();
		expect(overwrittenBy(22192)).toBeUndefined();
		expect(overwrittenBy(9206)).toBeDefined();
	});

	it('follows the picks as they change, on the configs already built', () => {
		const { player, groupNamed, overwrittenBy, partly } = picking([250350, 13452]);
		expect(partly(250350)).toBe(true);
		groupNamed('Agility').setValue(player, 9187);
		expect(overwrittenBy(9187)).toBe('Agility overwritten by Elixir of the Mongoose');
		groupNamed('Critical Strike').setValue(player, 250351);
		expect(overwrittenBy(250350)).toBe('Strength overwritten by Elixir of the Grizzly\nAgility overwritten by Elixir of Greater Agility');
		expect(partly(250350)).toBe(false);
		groupNamed('Critical Strike').setValue(player, 0);
		expect(overwrittenBy(250350)).toBe('Agility overwritten by Elixir of Greater Agility');
	});

	it('leaves a scroll and the elixirs of its stat unmarked, since they stack', () => {
		const strength = picking([], [10310, 12451]);
		expect(strength.scrollOverwrittenBy()).toBeUndefined();
		expect(strength.overwrittenBy(12451)).toBeUndefined();
		const intellect = picking([3383], [10308]);
		expect(intellect.intellectScrollOverwrittenBy()).toBeUndefined();
		expect(intellect.overwrittenBy(3383)).toBeUndefined();
		expect(picking([], []).scrollOverwrittenBy()).toBeUndefined();
	});

	it('marks Scroll of Intellect IV under Arcane Intellect, naming the buff', () => {
		expect(picking([], [10308]).intellectScrollOverwrittenBy()).toBeUndefined();
		expect(picking([], [10308], RaidBuffs.create({ arcaneIntellect: true })).intellectScrollOverwrittenBy()).toBe(
			'Intellect overwritten by Arcane Intellect',
		);
	});

	it('marks Scroll of Intellect IV and an Intellect elixir for a mage with the Arcane Intellect input off, since the sim turns it on for every mage', () => {
		const mage = picking([3383], [10308], RaidBuffs.create(), Class.ClassMage);
		expect(mage.intellectScrollOverwrittenBy()).toBe('Intellect overwritten by Arcane Intellect');
		expect(mage.overwrittenBy(3383)).toBe('Intellect overwritten by Arcane Intellect');
	});

	it('leaves a scroll unmarked under Mark of the Wild, which stacks with the scrolls', () => {
		const motw = RaidBuffs.create({ markOfTheWild: true });
		expect(picking([], [10308], motw).intellectScrollOverwrittenBy()).toBeUndefined();
		expect(picking([], [10310], motw).scrollOverwrittenBy()).toBeUndefined();
	});

	it('follows the raid buffs as they change, on the configs already built', () => {
		const { state, intellectScrollOverwrittenBy } = picking([], [10308]);
		expect(intellectScrollOverwrittenBy()).toBeUndefined();
		state.raidBuffs = RaidBuffs.create({ arcaneIntellect: true });
		expect(intellectScrollOverwrittenBy()).toBeDefined();
		state.raidBuffs = RaidBuffs.create();
		expect(intellectScrollOverwrittenBy()).toBeUndefined();
	});

	it('marks an Intellect elixir and Scroll of Intellect IV alike under Arcane Intellect, naming the buff', () => {
		const underArcaneIntellect = picking([3383], [10308], RaidBuffs.create({ arcaneIntellect: true }));
		expect(underArcaneIntellect.overwrittenBy(3383)).toBe('Intellect overwritten by Arcane Intellect');
		expect(underArcaneIntellect.intellectScrollOverwrittenBy()).toBe('Intellect overwritten by Arcane Intellect');
	});

	it('marks Elixir of the Sages under Prayer of Spirit as partly overwritten, its Intellect still counting', () => {
		const { overwrittenBy, partly } = picking([13447], [], RaidBuffs.create({ prayerOfSpirit: true }));
		expect(overwrittenBy(13447)).toBe('Spirit overwritten by Prayer of Spirit');
		expect(partly(13447)).toBe(true);
		expect(picking([13447]).overwrittenBy(13447)).toBeUndefined();
	});

	it('marks an elixir whose crit a raid crit matches, naming the stats it loses and to whom', () => {
		const lotp = RaidBuffs.create({ leaderOfThePack: true });
		expect(picking([250351], [12451], lotp).overwrittenBy(250351)).toBe('Strength overwritten by Juju Power\nCrit overwritten by Leader of the Pack');
		expect(picking([250351], [], lotp).overwrittenBy(250351)).toBe('Crit overwritten by Leader of the Pack');
		expect(picking([13452], [], RaidBuffs.create({ moonkinAura: true })).overwrittenBy(13452)).toBe('Crit overwritten by Moonkin Aura');
	});

	it("counts a druid's own Moonkin Aura or Leader of the Pack as on, whatever the input says, where the sim turns it on", () => {
		const druid = (spec: Spec, talents: Partial<DruidTalents>) =>
			picking([250351], [12451], RaidBuffs.create(), Class.ClassDruid, { spec, talents: DruidTalents.create(talents) }).overwrittenBy(250351);
		const strengthLost = 'Strength overwritten by Juju Power';
		expect(druid(Spec.SpecBalanceDruid, { moonkinForm: true })).toBe('Strength overwritten by Juju Power\nCrit overwritten by Moonkin Aura');
		expect(druid(Spec.SpecBalanceDruid, {})).toBe(strengthLost);
		expect(druid(Spec.SpecFeralCatDruid, { leaderOfThePack: true })).toBe('Strength overwritten by Juju Power\nCrit overwritten by Leader of the Pack');
		expect(druid(Spec.SpecFeralBearDruid, { leaderOfThePack: true })).toBe('Strength overwritten by Juju Power\nCrit overwritten by Leader of the Pack');
		expect(druid(Spec.SpecFeralCatDruid, {})).toBe(strengthLost);
		expect(druid(Spec.SpecRestorationDruid, { moonkinForm: true, leaderOfThePack: true })).toBe(strengthLost);
	});

	it('marks no elixir under Mark of the Wild, which stacks with the elixirs', () => {
		const motw = RaidBuffs.create({ markOfTheWild: true });
		expect(picking([6662, 3383], [], motw).overwrittenBy(6662)).toBeUndefined();
		expect(picking([6662, 3383], [], motw).overwrittenBy(3383)).toBeUndefined();
	});

	it('marks Elixir of Brute Force under Power Word: Fortitude as partly overwritten, and fully where another pick matches its Strength', () => {
		const fortitude = RaidBuffs.create({ powerWordFortitude: true });
		expect(picking([13453], [], fortitude).overwrittenBy(13453)).toBe('Stamina overwritten by Power Word: Fortitude');
		expect(picking([13453, 250351], [], fortitude).overwrittenBy(13453)).toBe(
			'Strength overwritten by Elixir of the Grizzly\nStamina overwritten by Power Word: Fortitude',
		);
		expect(picking([13453, 250351]).overwrittenBy(13453)).toBe('Strength overwritten by Elixir of the Grizzly');
	});
});
