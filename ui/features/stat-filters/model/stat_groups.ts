import { type Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';

/** In the character stats panel, `primary` (health and mana) shows without a header and the rest are collapsible. */
export type StatGroupKey = 'primary' | 'attributes' | 'physical' | 'spell' | 'defense' | 'resistance';

/** The groups the character stats panel and the stat filters sort stats into, each in display order. */
export const STAT_GROUPS: ReadonlyArray<{ key: StatGroupKey; stats: ReadonlyArray<UnitStat> }> = [
	{ key: 'primary', stats: [UnitStat.fromStat(Stat.StatHealth), UnitStat.fromStat(Stat.StatHP5), UnitStat.fromStat(Stat.StatMana)] },
	{
		key: 'attributes',
		stats: [
			UnitStat.fromStat(Stat.StatStrength),
			UnitStat.fromStat(Stat.StatAgility),
			UnitStat.fromStat(Stat.StatStamina),
			UnitStat.fromStat(Stat.StatIntellect),
			UnitStat.fromStat(Stat.StatSpirit),
		],
	},
	{
		key: 'physical',
		stats: [
			UnitStat.fromStat(Stat.StatAttackPower),
			UnitStat.fromStat(Stat.StatFeralAttackPower),
			UnitStat.fromStat(Stat.StatRangedAttackPower),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHastePercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedSpeedMultiplier),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent),
			UnitStat.fromStat(Stat.StatArmorPenetration),
		],
	},
	{
		key: 'spell',
		stats: [
			UnitStat.fromStat(Stat.StatSpellDamage),
			UnitStat.fromStat(Stat.StatHealingPower),
			UnitStat.fromStat(Stat.StatArcaneDamage),
			UnitStat.fromStat(Stat.StatFireDamage),
			UnitStat.fromStat(Stat.StatFrostDamage),
			UnitStat.fromStat(Stat.StatHolyDamage),
			UnitStat.fromStat(Stat.StatNatureDamage),
			UnitStat.fromStat(Stat.StatShadowDamage),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFrost),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentHoly),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentNature),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellCritPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHastePercent),
			UnitStat.fromStat(Stat.StatSpellPiercing),
			UnitStat.fromStat(Stat.StatMP5),
		],
	},
	{
		key: 'defense',
		stats: [
			UnitStat.fromStat(Stat.StatArmor),
			UnitStat.fromStat(Stat.StatBonusArmor),
			UnitStat.fromStat(Stat.StatArmorNoMods),
			UnitStat.fromStat(Stat.StatDefenseRating),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatParryPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent),
			UnitStat.fromStat(Stat.StatBlockValue),
		],
	},
	{
		key: 'resistance',
		stats: [
			UnitStat.fromStat(Stat.StatArcaneResistance),
			UnitStat.fromStat(Stat.StatFireResistance),
			UnitStat.fromStat(Stat.StatFrostResistance),
			UnitStat.fromStat(Stat.StatNatureResistance),
			UnitStat.fromStat(Stat.StatShadowResistance),
		],
	},
];

export type FilterGroupKey = StatGroupKey | 'ratings';

/** The off-hand percent of each melee percent, by the melee stat's key: the character stats panel shows both in one row. */
export const OFF_HAND_PERCENTS: ReadonlyMap<string, UnitStat> = new Map([
	[UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent).getKey(), UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent)],
	[UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent).getKey(), UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent)],
]);

/**
 * The groups the stat filter popovers sort stats into, each in display order: the character stats groups with the
 * off-hand hit and crit percents in Physical, then ratings and weapon DPS, which the panel has no rows for.
 */
const FILTER_GROUPS: ReadonlyArray<{ key: FilterGroupKey; stats: ReadonlyArray<UnitStat> }> = [
	...STAT_GROUPS.map(({ key, stats }) => ({
		key,
		stats: stats.flatMap(stat => {
			const offHand = OFF_HAND_PERCENTS.get(stat.getKey());
			return offHand ? [stat, offHand] : [stat];
		}),
	})),
	{
		key: 'ratings',
		stats: [
			UnitStat.fromStat(Stat.StatHitRating),
			UnitStat.fromStat(Stat.StatCritRating),
			UnitStat.fromStat(Stat.StatHasteRating),
			UnitStat.fromStat(Stat.StatExpertiseRating),
			UnitStat.fromStat(Stat.StatDodgeRating),
			UnitStat.fromStat(Stat.StatParryRating),
			UnitStat.fromStat(Stat.StatBlockRating),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandDps),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedDps),
		],
	},
];

const groupByKey = new Map(FILTER_GROUPS.flatMap(({ key, stats }) => stats.map(stat => [stat.getKey(), key] as const)));

const statGroupOf = (stat: UnitStat): FilterGroupKey | 'other' => groupByKey.get(stat.getKey()) ?? 'other';

/** The groups a stat filter popover lists for `universe`, each with only its stats in it, Other last and empty groups left out. */
export const filterGroupsOf = (universe: ReadonlyArray<UnitStat>): Array<{ key: FilterGroupKey | 'other'; stats: ReadonlyArray<UnitStat> }> => {
	const inUniverse = new Set(universe.map(stat => stat.getKey()));
	const grouped = FILTER_GROUPS.map(({ key, stats }) => ({ key, stats: stats.filter(stat => inUniverse.has(stat.getKey())) }));
	const other = { key: 'other' as const, stats: universe.filter(stat => statGroupOf(stat) === 'other') };
	return [...grouped, other].filter(group => group.stats.length > 0);
};

/**
 * Each stat's label in a stat filter popover for `universe`, by stat key: its short name, or, where two stats share
 * one (Dodge rating and Dodge percent), its rating or percent form.
 */
export const filterLabelsOf = (universe: ReadonlyArray<UnitStat>, playerClass: Class): Map<string, string> => {
	const shortNames = universe.map(stat => stat.getShortName(playerClass));
	return new Map(
		universe.map((stat, index) => {
			const name = shortNames[index];
			if (shortNames.indexOf(name) === shortNames.lastIndexOf(name)) return [stat.getKey(), name];
			return [stat.getKey(), i18n.t(stat.isStat() ? 'stat_filters.rating_label' : 'stat_filters.percent_label', { stat: name })];
		}),
	);
};

/** The groups that hold at least one shown stat: the ones a popover opens with when nothing is stored for its panel. */
export const defaultOpenFilterGroups = (universe: ReadonlyArray<UnitStat>, shown: ReadonlyArray<UnitStat>): Array<FilterGroupKey | 'other'> => {
	const shownKeys = new Set(shown.map(stat => stat.getKey()));
	return filterGroupsOf(universe)
		.filter(group => group.stats.some(stat => shownKeys.has(stat.getKey())))
		.map(group => group.key);
};
