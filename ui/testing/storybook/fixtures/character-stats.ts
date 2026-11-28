import type { VersusBonusSet } from '@features/character-stats/components/CharacterStats/VersusBonuses';
import { type MeleeCritCapInfo, VersusEffectKind } from '@generated/proto/api';
import { ActionID, AreaType, CreatureFamily, MobType, PseudoStat, RangedWeaponType, Stat, Target } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import { type StatAttribution, Stats, type UnitStat } from '@sim/proto/stats';

const stage = (strength: number, attackPower: number, meleeHit: number, offHandHit: number, rangedHaste: number) =>
	new Stats()
		.withStat(Stat.StatStrength, strength)
		.withStat(Stat.StatAttackPower, attackPower)
		.withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, meleeHit)
		.withPseudoStat(PseudoStat.PseudoStatOffHandHitPercent, offHandHit)
		.withPseudoStat(PseudoStat.PseudoStatRangedHastePercent, rangedHaste);

/** Each stage's own contribution, and their sum as `final`. */
export const makeStatAttribution = (): StatAttribution => {
	const base = stage(120, 160, 0, 0, 0);
	const gear = stage(180, 240, 3, 2, 0);
	const talents = stage(0, 0, 3, 3, 0);
	const buffs = stage(45, 190, 0, 0, 15);
	const consumes = stage(25, 0, 0, 0, 0);
	const debuffs = stage(0, 0, 2, 2, 0);
	return { base, gear, talents, buffs, consumes, debuffs, final: base.add(gear).add(talents).add(buffs).add(consumes).add(debuffs) };
};

/** A stand-in for the sheet's formatter: a pseudo stat as a percentage, a stat as a whole number. */
export const showStat = (deltaStats: Stats, unitStat: UnitStat): string => {
	const value = deltaStats.getUnitStat(unitStat);
	return unitStat.isPseudoStat() ? `${value.toFixed(2)}%` : value.toFixed(0);
};

export const makeRangedWeapon = (): UIItem => UIItem.create({ rangedWeaponType: RangedWeaponType.RangedWeaponTypeBow, weaponSpeed: 2.9 });

export const MISS_INFO = { base: 5, defense: 2.4, debuffs: 1, total: 8.4 };

export const AVOIDANCE_INFO = { miss: 8.4, dodge: 18.2, parry: 14.1, block: 42.3, total: 103, shear: 60.7 };

export const CRIT_IMMUNITY_INFO = { total: 5.6, delta: 0, defense: 4.6, talents: 1 };

export const CRIT_CAP_INFO: MeleeCritCapInfo = {
	glancing: 40,
	suppression: 4.8,
	remainingHit: 1.2,
	remainingExpertise: 2.5,
	block: 0,
	specOffset: 0,
	critCap: 61.1,
	delta: -8.4,
};

export const VERSUS_TARGETS = [
	Target.create({ mobType: MobType.MobTypeHumanoid, creatureFamily: CreatureFamily.CreatureFamilyDwarf }),
	Target.create({ mobType: MobType.MobTypeBeast, creatureFamily: CreatureFamily.CreatureFamilySerpentine }),
];

export const VERSUS_AREA_TYPES = [AreaType.AreaTypeForestGrassland];

const itemSource = (itemId: number) => ActionID.create({ rawId: { oneofKind: 'itemId', itemId } });
const spellSource = (spellId: number) => ActionID.create({ rawId: { oneofKind: 'spellId', spellId } });
const humanoidStats = { [Stat.StatAttackPower]: 6, [Stat.StatRangedAttackPower]: 6 };
const undeadStats = { [Stat.StatAttackPower]: 18, [Stat.StatRangedAttackPower]: 18 };

/** A Troll hunter with Beast Slaying, The Jaw Breaker and Humanoid Slaying against Humanoids, Gargoyle Slashers against Undead, an area ring, items whose spells deal, restore or last more against some targets, the Chain of the Scarlet Crusade 4-piece's damage against Undead, and area factors: Golden Banana's in Forest and Grassland, Darkmoon Card: Heroism's and the Krol'dok Battlegear 4-piece's in Strongholds and Cities. */
export const VERSUS_BONUSES: VersusBonusSet = {
	mobTypeBonuses: [
		{
			mobType: MobType.MobTypeBeast,
			stats: {},
			damageMultiplier: 1.05,
			critDamageMultiplier: 1,
			sources: [{ actionId: spellSource(20557), stats: {}, damageMultiplier: 1.05, critDamageMultiplier: 1 }],
		},
		{
			mobType: MobType.MobTypeHumanoid,
			stats: humanoidStats,
			damageMultiplier: 1.03,
			critDamageMultiplier: 1.03,
			sources: [
				{ actionId: itemSource(22322), stats: humanoidStats, damageMultiplier: 1, critDamageMultiplier: 1 },
				{ actionId: spellSource(19153), stats: {}, damageMultiplier: 1.03, critDamageMultiplier: 1.03 },
			],
		},
		{
			mobType: MobType.MobTypeUndead,
			stats: undeadStats,
			damageMultiplier: 1,
			critDamageMultiplier: 1,
			sources: [{ actionId: itemSource(13957), stats: undeadStats, damageMultiplier: 1, critDamageMultiplier: 1 }],
		},
	],
	areaStats: [{ areaType: AreaType.AreaTypeForestGrassland, stats: { [Stat.StatAttackPower]: 28, [Stat.StatHitRating]: 10 } }],
	versusEffects: [
		{
			itemId: 280032,
			source: itemSource(280032),
			itemName: "Bolt's Charged Eyeball",
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilyDinosaur],
			multiplier: 3,
			kind: VersusEffectKind.VersusEffectKindDamage,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 274609,
			source: itemSource(274609),
			itemName: 'Serpentbane Crown',
			mobTypes: [MobType.MobTypeBeast],
			creatureFamilies: [CreatureFamily.CreatureFamilySerpentine],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindDamage,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 17943,
			source: itemSource(17943),
			itemName: 'Fist of Stone',
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilyEarthElemental],
			multiplier: 3,
			kind: VersusEffectKind.VersusEffectKindRestore,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 11815,
			source: itemSource(11815),
			itemName: 'Hand of Justice',
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilyDwarf],
			multiplier: 3,
			kind: VersusEffectKind.VersusEffectKindProcChance,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 13965,
			source: itemSource(13965),
			itemName: "Blackhand's Breadth",
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilyOrc, CreatureFamily.CreatureFamilyOgre],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindCritChance,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 10760,
			source: itemSource(10760),
			itemName: 'Swine Fists',
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilySwine],
			multiplier: 0,
			kind: VersusEffectKind.VersusEffectKindProcOnly,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 9477,
			source: itemSource(9477),
			itemName: "The Chief's Enforcer",
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilyTroll],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindDamage,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 9477,
			source: itemSource(9477),
			itemName: "The Chief's Enforcer",
			mobTypes: [],
			creatureFamilies: [CreatureFamily.CreatureFamilyTroll],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindDuration,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 0,
			source: spellSource(1293674),
			itemName: '',
			mobTypes: [MobType.MobTypeUndead],
			creatureFamilies: [],
			multiplier: 3,
			kind: VersusEffectKind.VersusEffectKindDamage,
			areaTypes: [],
			amount: 0,
		},
		{
			itemId: 270226,
			source: itemSource(270226),
			itemName: 'Golden Banana',
			mobTypes: [],
			creatureFamilies: [],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindEffect,
			areaTypes: [AreaType.AreaTypeForestGrassland],
			amount: 0,
		},
		{
			itemId: 270226,
			source: itemSource(270226),
			itemName: 'Golden Banana',
			mobTypes: [],
			creatureFamilies: [],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindDuration,
			areaTypes: [AreaType.AreaTypeForestGrassland],
			amount: 0,
		},
		{
			itemId: 274760,
			source: itemSource(274760),
			itemName: 'Everlook Delivery Bot',
			mobTypes: [],
			creatureFamilies: [],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindHitsNeeded,
			areaTypes: [AreaType.AreaTypeSnowy],
			amount: 0,
		},
		{
			itemId: 19287,
			source: itemSource(19287),
			itemName: 'Darkmoon Card: Heroism',
			mobTypes: [],
			creatureFamilies: [],
			multiplier: 2,
			kind: VersusEffectKind.VersusEffectKindProcChance,
			areaTypes: [AreaType.AreaTypeStrongholdsCities],
			amount: 0,
		},
		{
			itemId: 0,
			source: spellSource(1324298),
			itemName: '',
			mobTypes: [],
			creatureFamilies: [],
			multiplier: 0,
			kind: VersusEffectKind.VersusEffectKindFlatHeal,
			areaTypes: [AreaType.AreaTypeStrongholdsCities],
			amount: 44,
		},
	],
};
