import { iconEnumPickerShown, iconEnumValues } from '@features/inputs/components/ConnectedIconEnumPicker/utils';
import { shownStats, withSelection } from '@features/stat-filters/model/stat_filters';
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
import { ItemEffect } from '@generated/proto/spell';
import { DatabaseFilters, UIItem } from '@generated/proto/ui';
import type { Player, SpecConfigData } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import { Stats, UnitStat } from '@sim/proto/stats';
import { ConsumableSlot, relevantConsumableIds } from '@sim/settings/consumables';
import { describe, expect, it, vi } from 'vitest';

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
	zoneId: number;
	meleeDps: boolean;
	tank: boolean;
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
		zoneId: 0,
		meleeDps: true,
		tank: false,
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
				getZoneId: () => full.zoneId,
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
		expect(configs.elixirs.map(config => offered(config, player))).toEqual([[6662, 9206, 12451, 2289, 10310], [], [12820]]);
		expect(configs.scrolls.map(config => offered(config, player))).toEqual([[]]);
		expect(offered(configs.food, player)).toEqual([238637]);
		expect(configs.foodGroups.map(config => offered(config, player))).toEqual([[]]);
		expect(configs.miscGroups.map(config => offered(config, player))).toEqual([[8410]]);
		expect(configs.holidayGroups.map(config => offered(config, player))).toEqual([[]]);
		expect(toggled(configs.holiday, player)).toEqual([]);
		expect(offered(configs.explosive, player)).toEqual([18588]);
		expect(toggled(configs.explosiveExtras, player)).toEqual([10646]);
		expect(toggled(misc(configs), player)).toEqual([4376]);
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

		it("orders again when the target's mob type or the zone changes, on the configs already built", () => {
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
			expect(offered(config, player)).toEqual([3, 2, 1]);
			state.mobType = MobType.MobTypeDemon;
			expect(offered(config, player)).toEqual([1, 3, 2]);
			state.zoneId = 102;
			expect(offered(config, player)).toEqual([2, 1, 3]);
			state.mobType = MobType.MobTypeMechanical;
			expect(offered(config, player)).toEqual([2, 3, 1]);
			state.zoneId = 0;
			expect(offered(config, player)).toEqual([3, 2, 1]);
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
		expect(scrolls.map(config => config.tooltip)).toEqual(['Stamina']);
		expect(scrolls.map(config => iconEnumPickerShown(config, player))).toEqual([false]);
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
		expect(configs.scrolls.map(config => offered(config, player))).toEqual([[1180]]);
	});

	it('splits the miscellaneous row as the generator classes each consumable', () => {
		const { player } = playerOf();
		const { miscOffensive, miscDefensive } = consumeConfigs(player, db);
		expect(miscDefensive.map(toggle => toggle.consumable.id)).toEqual([4376]);
		expect(miscOffensive.map(toggle => toggle.consumable.id)).toEqual([1400]);
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
		configs.miscGroups[0].setValue(player, 8410);
		expect([state.consumes.elixirIds, state.consumes.extraIds]).toEqual([[9206], [1180, 8410]]);
		configs.miscGroups[0].setValue(player, 20079);
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
		expect(offered(configs.elixirs[0], player)).toEqual([6662, 9206, 12451, 2289, 10310]);
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
