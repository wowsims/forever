import { PseudoStat, Stat } from '@generated/proto/common';
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
];

export const EP_UNIT_STATS: UnitStat[] = UnitStat.getAll().filter(stat => {
	if (stat.isStat()) {
		return true;
	} else {
		return EP_PSEUDO_STATS.includes(stat.getPseudoStat());
	}
});

export const isEpStat = (stat: UnitStat, { epStats, epPseudoStats }: EpStatSet): boolean => {
	if (stat.isStat()) return epStats.includes(stat.getStat());
	return epPseudoStats.includes(stat.getPseudoStat());
};

// A pseudo-stat the spec does not weight stays hidden even under "Show all stats": its Current EP
// cell writes straight into the player's EP weights, so an off-spec row there (Main Hand DPS on a
// caster) would skew every item and enchant EP in the gear picker from a row the checkbox then hides.
export const visibleEpUnitStats = (statSet: EpStatSet, showAllStats: boolean): UnitStat[] =>
	EP_UNIT_STATS.filter(stat => isEpStat(stat, statSet) || (showAllStats && stat.isStat()));
