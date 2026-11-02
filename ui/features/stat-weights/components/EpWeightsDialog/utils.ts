import type { StatWeightsResult } from '@generated/proto/api';
import type { Stat, UnitStats } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateStat } from '@i18n/localization';
import type { UnitStat } from '@sim/proto/stats';

import type { StatsTableColumn } from '../../model/stats_table';
import type { EpColumn } from './types';

/** One tooltip serves every anchor in the dialog; each anchor carries its own `data-tooltip-content`. */
export const EP_TOOLTIP_ID = 'ep-weights-tooltip';

export const statName = (stat: Stat | undefined): string => (stat !== undefined ? translateStat(stat) : '??');

const stepAmount = (value: number): number => Number(value.toFixed(2));

/** What the stat's runs changed it by, with + for a run above the baseline only; null before a run. */
export const statStepText = (stat: UnitStat, result: StatWeightsResult | null): string | null => {
	const modHigh = result?.modHigh ? stat.getProtoValue(result.modHigh) : 0;
	if (modHigh === 0) return null;
	const sign = result?.modLow && stat.getProtoValue(result.modLow) !== 0 ? '±' : '+';
	const ratingScale = stat.epRatingScale();
	if (ratingScale !== 1) {
		return i18n.t('sidebar.buttons.stat_weights.modal.step_rating_percent', {
			sign,
			rating: stepAmount(modHigh * ratingScale),
			percent: stepAmount(modHigh),
		});
	}
	return stat.isStat() && stat.isPercentRating()
		? i18n.t('sidebar.buttons.stat_weights.modal.step_rating', { sign, rating: stepAmount(modHigh) })
		: `${sign}${stepAmount(modHigh)}`;
};

/** The reference stat is read at render, so the tooltip follows a change of it. */
const columnLabelTooltip = (column: StatsTableColumn): string => {
	if (!column.getEpRefStat) return column.labelTooltip;
	const refStatName = statName(column.getEpRefStat());
	return `${column.labelTooltip} ${i18n.t('sidebar.buttons.stat_weights.modal.tooltips.normalized_by', { refStatName })}`;
};

export const buildEpColumns = (columns: StatsTableColumn[], onCopy: (weights: UnitStats | undefined) => void): EpColumn[] => {
	const ratioIndexes = { ep: 0, weight: 0 };
	return columns.map((column, index) => ({
		...column,
		id: `ep-col-${index}`,
		ratioIndex: column.type === 'action' ? undefined : ratioIndexes[column.type]++,
		labelTooltip: columnLabelTooltip(column),
		onCopy: () => onCopy(column.getWeights()),
	}));
};
