import i18n from '@i18n/config';
import type { ActionMetrics } from '@sim/proto/sim_result';
import { Tooltip } from '@ui-kit/Tooltip';
import { useMemo } from 'react';

import { useActionRows } from '../../hooks/useActionRows';
import { indexMetricRows } from '../../model/grouping';
import { groupActionsWithPets } from '../../model/pet_groups';
import type { SimResultData } from '../../model/result_data';
import { amountHeader, attackFormat, attackMetricsColumns, castsTooltip, hitsTooltip, missTooltip, useMetricMax } from '../AttackMetricsColumns';
import { MetricsCombinedTooltip } from '../MetricsCombinedTooltip';
import { createMetricsColumnHelper, metricForAnchor, MetricsTable, MetricTooltip } from '../MetricsTable';

const helper = createMetricsColumnHelper<ActionMetrics>();

const TOOLTIP = {
	casts: 'threat-metrics-casts',
	avgCastHeader: 'threat-metrics-avg-cast-header',
	hits: 'threat-metrics-hits',
	missPercent: 'threat-metrics-miss-percent',
	tps: 'threat-metrics-tps',
};

const threatGroups = (resultData: SimResultData): Array<Array<ActionMetrics>> => {
	const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
	if (!players.length) return [];

	return groupActionsWithPets(players[0], unit => unit.getThreatActions().map(action => action.forTarget(resultData.filter)));
};

/**
 * The per-action threat table. Threat is a first-class metric here — TPS decides who tanks — so it
 * gets its own tab rather than the per-column threat tooltips the damage table carries. The table is
 * mounted only while `showThreatMetrics` is on; nothing inside it re-checks that.
 */
export const ThreatMetricsTable = () => {
	const { resultData, rows } = useActionRows(threatGroups);
	const metricsByRowId = useMemo(() => indexMetricRows(rows), [rows]);
	const maxThreat = useMetricMax(rows, metric => metric.threat);

	const columns = useMemo(
		() =>
			helper.columns([
				attackMetricsColumns.name(),
				attackMetricsColumns.primary({
					id: 'threat-done',
					header: i18n.t('results_tab.details.columns.threat_done'),
					total: metric => metric.avgThreat,
					value: metric => metric.threat,
					percentage: metric => metric.totalThreatPercent,
					max: maxThreat,
					// The bar carries the whole breakdown; there is no second tooltip on this column.
					tooltipId: 'threat-metrics-threat-done',
				}),
				attackMetricsColumns.casts({ tooltipId: TOOLTIP.casts }),
				attackMetricsColumns.withTicks({
					id: 'avg-cast',
					header: i18n.t('results_tab.details.columns.avg_cast'),
					value: metric => metric.avgCastThreat,
					tick: () => 0,
					format: attackFormat.compact,
					zeroWhen: metric => metric.isPassiveAction,
					dashWhen: metric => metric.isPassiveAction,
					meta: {
						headerTooltipId: TOOLTIP.avgCastHeader,
						headerTooltip: i18n.t('results_tab.details.tooltips.damage_avg_cast_tooltip'),
					},
				}),
				attackMetricsColumns.withTicks({
					id: 'hits',
					header: i18n.t('results_tab.details.columns.hits'),
					value: metric => metric.landedHits,
					tick: metric => metric.landedTicks,
					format: attackFormat.number,
					meta: { tooltipId: TOOLTIP.hits },
				}),
				helper.accessor(row => row.metric.avgHitThreat, {
					id: 'avg-hit',
					header: i18n.t('results_tab.details.columns.avg_hit'),
					cell: info => attackFormat.compact(info.getValue()),
				}),
				attackMetricsColumns.withTicks({
					id: 'crit-percent',
					header: i18n.t('results_tab.details.columns.crit_percent'),
					value: metric => metric.critPercent + metric.blockedCritPercent,
					tick: metric => metric.critTickPercent,
					format: attackFormat.percent,
				}),
				helper.accessor(row => row.metric.totalMissesPercent, {
					id: 'miss-percent',
					header: i18n.t('results_tab.details.columns.miss_percent'),
					meta: { tooltipId: TOOLTIP.missPercent },
					cell: info => attackFormat.percent(info.getValue()),
				}),
				helper.accessor(row => row.metric.threatThroughput, {
					id: 'tpet',
					header: i18n.t('results_tab.details.columns.tpet'),
					cell: info => attackFormat.compact(info.getValue()),
				}),
				attackMetricsColumns.rate({
					id: 'tps',
					header: i18n.t('results_tab.details.columns.tps'),
					value: metric => metric.tps,
					tooltipId: TOOLTIP.tps,
				}),
			]),
		[maxThreat],
	);

	const forAnchor = metricForAnchor(metricsByRowId);

	return (
		<>
			<MetricsTable rootTestId="threat-metrics-root" columns={columns} rows={rows} sortColumnId="tps" hasResult={!!resultData} />
			<Tooltip id={TOOLTIP.avgCastHeader} />
			<MetricTooltip id={TOOLTIP.casts} forAnchor={forAnchor} body={castsTooltip} />
			<MetricTooltip id={TOOLTIP.hits} forAnchor={forAnchor} body={hitsTooltip} />
			<MetricTooltip id={TOOLTIP.missPercent} forAnchor={forAnchor} body={missTooltip} />
			<MetricTooltip
				id={TOOLTIP.tps}
				forAnchor={forAnchor}
				body={metric => {
					if (!metric.tps) return null;
					const peak = Math.max(metric.avgCastThreat, metric.avgHitThreat);
					return (
						<MetricsCombinedTooltip
							headerValues={amountHeader()}
							groups={[
								{
									spellSchool: metric.spellSchool,
									totalPercentage: 100,
									data: [
										{
											name: i18n.t('results_tab.details.tooltip_table.per_cast'),
											value: metric.avgCastThreat,
											percentage: (metric.avgCastThreat / peak) * 100,
										},
										{
											name: i18n.t('results_tab.details.tooltip_table.per_hit'),
											value: metric.avgHitThreat,
											percentage: (metric.avgHitThreat / peak) * 100,
										},
									],
								},
							]}
						/>
					);
				}}
			/>
		</>
	);
};
