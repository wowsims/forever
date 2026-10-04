import { PseudoStat } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { StatAttribution, Stats, UnitStat } from '@sim/proto/stats';

import { TooltipRow } from './TooltipRow';

export type ShowStat = (deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean) => string;

export interface StatTooltipBodyProps {
	hand: UnitStat;
	handLabel?: string;
	attribution: StatAttribution;
	show: ShowStat;
	bonusStats: Stats;
	bonusStatValue: number;
	total: string;
	/** Given only where the effective-weapon-speed row applies: a hunter's bow, crossbow or gun on the ranged haste row. */
	rangedWeaponSpeed?: number;
}

// One hand's stage breakdown. `handLabel`, given only on a combined row while dual wielding, heads the
// list so the two tooltips read apart. As a component rather than a plain function, its `show()` calls
// run only once react-tooltip actually opens it, not on every StatRow render.
export const StatTooltipBody = ({ hand, handLabel, attribution, show, bonusStats, bonusStatValue, total, rangedWeaponSpeed }: StatTooltipBodyProps) => (
	<>
		{handLabel && (
			<div data-testid="character-stats-tooltip-hand" className="mb-1 text-sm font-semibold text-white">
				{handLabel}
			</div>
		)}
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.base')} value={show(attribution.base, hand, true)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.gear')} value={show(attribution.gear, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.talents')} value={show(attribution.talents, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.buffs')} value={show(attribution.buffs, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.consumes')} value={show(attribution.consumes, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.debuffs')} value={show(attribution.debuffs, hand)} />
		{bonusStatValue !== 0 && <TooltipRow label={i18n.t('sidebar.character_stats.tooltip.bonus')} value={show(bonusStats, hand)} />}
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.total')} value={total} />
		{rangedWeaponSpeed !== undefined && (
			<TooltipRow
				label={i18n.t('sidebar.character_stats.tooltip.eWS')}
				value={`${(rangedWeaponSpeed / (1 + attribution.final.getPseudoStat(PseudoStat.PseudoStatRangedHastePercent) / 100)).toFixed(2)}s`}
			/>
		)}
	</>
);
