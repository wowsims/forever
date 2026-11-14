import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { type ItemSlot, PseudoStat, Stat } from '@generated/proto/common';
import { MELEE_HANDS, WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { SCHOOL_HIT_PERCENT_PSEUDO_STATS, UnitStat } from '@sim/proto/stats';

export type EpStatSet = {
	epStats: Stat[];
	epPseudoStats: PseudoStat[];
};

const EP_PSEUDO_STATS = [
	PseudoStat.PseudoStatMainHandDps,
	PseudoStat.PseudoStatOffHandDps,
	PseudoStat.PseudoStatRangedDps,
	PseudoStat.PseudoStatMeleeHitPercent,
	PseudoStat.PseudoStatSpellHitPercent,
	...SCHOOL_HIT_PERCENT_PSEUDO_STATS,
	PseudoStat.PseudoStatMeleeCritPercent,
	PseudoStat.PseudoStatSpellCritPercent,
	PseudoStat.PseudoStatRangedHitPercent,
	PseudoStat.PseudoStatRangedCritPercent,
	PseudoStat.PseudoStatExpertisePercent,
	PseudoStat.PseudoStatMainHandWeaponSkill,
	PseudoStat.PseudoStatOffHandWeaponSkill,
];

export const EP_UNIT_STATS: UnitStat[] = UnitStat.getAll().filter(stat => {
	if (stat.isStat()) {
		return true;
	} else {
		return EP_PSEUDO_STATS.includes(stat.getPseudoStat());
	}
});

/**
 * The spec's EP pseudo stats with a weapon skill row for each hand whose weapon DPS it weighs and that holds a weapon,
 * when it shows melee weapon skill. A feral spec shows none, as its forms ignore weapon skill, and a hunter weighs no
 * melee weapon DPS. A hand with no weapon gets no row, as stat weights skip it and it would read 0.
 */
export const withWeaponSkillEpPseudoStats = (
	epPseudoStats: PseudoStat[],
	displayWeaponSkillGroups: ReadonlyArray<WeaponSkillGroup> | undefined,
	holdsWeaponIn: (slot: ItemSlot) => boolean,
): PseudoStat[] =>
	displayWeaponSkillGroups?.includes(WeaponSkillGroup.Melee)
		? [...epPseudoStats, ...MELEE_HANDS.filter(hand => epPseudoStats.includes(hand.weaponDps) && holdsWeaponIn(hand.slot)).map(hand => hand.weaponSkill)]
		: epPseudoStats;

export const isEpStat = (stat: UnitStat, { epStats, epPseudoStats }: EpStatSet): boolean => {
	if (stat.isStat()) return epStats.includes(stat.getStat());
	return epPseudoStats.includes(stat.getPseudoStat());
};

export const epFilterScope = (statSet: EpStatSet): StatFilterScope => ({
	defaults: EP_UNIT_STATS.filter(stat => isEpStat(stat, statSet)),
	// Every Stat, but only the pseudo-stats the spec weights: a row's Current EP cell writes straight into the
	// player's EP weights, so an off-spec pseudo-stat row (Main Hand DPS on a caster) would let one entry skew
	// every item and enchant EP in the gear picker.
	universe: EP_UNIT_STATS.filter(stat => stat.isStat() || isEpStat(stat, statSet)),
});
