import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
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

// The rating points one stored EP weight unit covers. A percent pseudo-stat's weight is stored per percentage point,
// since enchants hold it as a percent and item EP is a dot product, and the stats table shows it per rating point beside
// the ratings. The school hit weights come from the sim per hit rating point already, and have no root rating.
export const epRatingScale = (stat: UnitStat): number => (stat.isPseudoStat() && stat.isPercentRating() ? (stat.convertPercentToRating(1) ?? 1) : 1);

export const epFilterScope = (statSet: EpStatSet): StatFilterScope => ({
	defaults: EP_UNIT_STATS.filter(stat => isEpStat(stat, statSet)),
	// Every Stat, but only the pseudo-stats the spec weights: a row's Current EP cell writes straight into the
	// player's EP weights, so an off-spec pseudo-stat row (Main Hand DPS on a caster) would let one entry skew
	// every item and enchant EP in the gear picker.
	universe: EP_UNIT_STATS.filter(stat => stat.isStat() || isEpStat(stat, statSet)),
});
