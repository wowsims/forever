import { PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { SCHOOL_HIT_PERCENT_PSEUDO_STATS, UnitStat } from '@sim/proto/stats';
import type { StatTooltipContent } from '@sim/spec_config';
import { LocaleHtml } from '@ui-kit/Tooltip';
import type { ReactNode } from 'react';

/** The stats the optimizer offers a cap for. Held as UnitStats because several are pseudo-stats. */
export const INCLUDED_STATS: UnitStat[] = [
	UnitStat.fromStat(Stat.StatHitRating),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent),
	...SCHOOL_HIT_PERCENT_PSEUDO_STATS.map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat)),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent),
	UnitStat.fromStat(Stat.StatCritRating),
	UnitStat.fromStat(Stat.StatHasteRating),
	UnitStat.fromStat(Stat.StatExpertiseRating),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent),
	UnitStat.fromStat(Stat.StatArmorPenetration),
	UnitStat.fromStat(Stat.StatDodgeRating),
	UnitStat.fromStat(Stat.StatParryRating),
	UnitStat.fromStat(Stat.StatDefenseRating),
];

export const isIncludedStat = (unitStat: UnitStat): boolean => INCLUDED_STATS.some(included => included.equals(unitStat));

const HasteCapTooltip = () => <LocaleHtml html={i18n.t('sidebar.buttons.suggest_reforges.haste_cap_tooltip')} />;

const DEFAULT_STAT_TOOLTIPS: Partial<Record<Stat, ReactNode>> = {
	[Stat.StatHasteRating]: <HasteCapTooltip />,
};

/** The panel's own entries, overridden per stat by the spec's. Evaluated once per popover open. */
export const buildStatTooltips = (override: StatTooltipContent | undefined): Partial<Record<Stat, ReactNode>> => {
	const tooltips: Partial<Record<Stat, ReactNode>> = { ...DEFAULT_STAT_TOOLTIPS };
	for (const [stat, make] of Object.entries(override ?? {})) {
		const content = make?.();
		if (content !== undefined) tooltips[Number(stat) as Stat] = content;
	}
	return tooltips;
};
