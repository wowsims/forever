import type { StatWeightsResult } from '@generated/proto/api';
import type { Stat, UnitStats } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateStat } from '@i18n/localization';
import type { UnitStat } from '@sim/proto/stats';
import { formatToNumber } from '@sim/utils/format';

import type { StatsTableColumn } from '../../model/stats_table';
import type { EpColumn } from './types';

/** One tooltip serves every anchor in the dialog; each anchor carries its own `data-tooltip-content`. */
export const EP_TOOLTIP_ID = 'ep-weights-tooltip';

export const statName = (stat: Stat | undefined): string => (stat !== undefined ? translateStat(stat) : '??');

/** How far the sim moved the stat to measure it, with + for a stat run above the baseline only; undefined without a run. */
export const simulatedStepText = (stat: UnitStat, result: StatWeightsResult | null): string | undefined => {
	const stepAbove = result?.stepAbove ? stat.getProtoValue(result.stepAbove) : 0;
	if (stepAbove === 0) return undefined;
	const sign = result?.stepBelow && stat.getProtoValue(result.stepBelow) !== 0 ? '±' : '+';
	if (stat.isRatingBackedPercent()) {
		return i18n.t('sidebar.buttons.stat_weights.modal.simulated_step_rating_percent', {
			sign,
			rating: formatToNumber(stepAbove * stat.epRatingScale()),
			percent: formatToNumber(stepAbove),
		});
	}
	return stat.isPercentRating()
		? i18n.t('sidebar.buttons.stat_weights.modal.simulated_step_rating', { sign, rating: formatToNumber(stepAbove) })
		: `${sign}${formatToNumber(stepAbove)}`;
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
