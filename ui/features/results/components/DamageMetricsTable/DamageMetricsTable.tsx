import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { ActionMetrics } from '@sim/proto/sim_result';
import { Tooltip } from '@ui-kit/Tooltip';
import { useMemo } from 'react';

import { useActionRows } from '../../hooks/useActionRows';
import { filterMetricRows, indexMetricRows } from '../../model/grouping';
import { groupActionsWithPets } from '../../model/pet_groups';
import type { SimResultData } from '../../model/result_data';
import {
	attackFormat,
	attackMetricsColumns,
	castsTooltip,
	damageBreakdownTooltip,
	hitsTooltip,
	missTooltip,
	threatTooltip,
	useMetricMax,
} from '../AttackMetricsColumns';
import { createMetricsColumnHelper, metricForAnchor, MetricsTable, MetricTooltip } from '../MetricsTable';

const helper = createMetricsColumnHelper<ActionMetrics>();

const TOOLTIP = {
	damage: 'damage-metrics-damage',
	casts: 'damage-metrics-casts',
	avgCast: 'damage-metrics-avg-cast',
	avgCastHeader: 'damage-metrics-avg-cast-header',
	hits: 'damage-metrics-hits',
	avgHit: 'damage-metrics-avg-hit',
	missPercent: 'damage-metrics-miss-percent',
	dps: 'damage-metrics-dps',
};

const damageGroups = (resultData: SimResultData): Array<Array<ActionMetrics>> => {
	const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
	if (!players.length) return [];

	return groupActionsWithPets(players[0], unit => unit.getDamageActions().map(action => action.forTarget(resultData.filter)));
};

const isThreatOnly = (metric: ActionMetrics) => metric.hitAttempts == 0 && metric.dps == 0;

export const DamageMetricsTable = () => {
	const sim = useSim();
	const { threat: showThreatMetrics } = useDisplayMetrics(sim);
	const { resultData, rows: built } = useActionRows(damageGroups);

	const rows = useMemo(() => (showThreatMetrics ? built : filterMetricRows(built, isThreatOnly)), [built, showThreatMetrics]);
	const metricsByRowId = useMemo(() => indexMetricRows(rows), [rows]);
	const maxDamage = useMetricMax(rows, metric => metric.damage);

	const columns = useMemo(
		() =>
			helper.columns([
				attackMetricsColumns.name(),
				attackMetricsColumns.primary({
					id: 'damage-done',
					header: i18n.t('results_tab.details.columns.damage_done'),
					total: metric => metric.avgDamage,
					value: metric => metric.damage,
					percentage: metric => metric.totalDamagePercent,
					max: maxDamage,
					tooltipId: TOOLTIP.damage,
				}),
				attackMetricsColumns.casts({ tooltipId: TOOLTIP.casts }),
				attackMetricsColumns.withTicks({
					id: 'avg-cast',
					header: i18n.t('results_tab.details.columns.avg_cast'),
					value: metric => metric.avgCastHit,
					tick: metric => metric.avgCastTick,
					format: attackFormat.compact,
					zeroWhen: metric => metric.isPassiveAction,
					dashWhen: metric => metric.isPassiveAction,
					meta: {
						tooltipId: TOOLTIP.avgCast,
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
				attackMetricsColumns.withTicks({
					id: 'avg-hit',
					header: i18n.t('results_tab.details.columns.avg_hit'),
					value: metric => metric.avgHit,
					tick: metric => metric.avgTick,
					format: attackFormat.compact,
					meta: { tooltipId: TOOLTIP.avgHit },
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
				helper.accessor(row => row.metric.damageThroughput, {
					id: 'dpet',
					header: i18n.t('results_tab.details.columns.dpet'),
					cell: info => attackFormat.compact(info.getValue()),
				}),
				attackMetricsColumns.rate({
					id: 'dps',
					header: i18n.t('results_tab.details.columns.dps'),
					value: metric => metric.dps,
					tooltipId: TOOLTIP.dps,
				}),
			]),
		[maxDamage],
	);

	const forAnchor = metricForAnchor(metricsByRowId);

	return (
		<>
			<MetricsTable
				rootTestId="damage-metrics-root"
				columns={columns}
				rows={rows}
				sortColumnId="dps"
				hasResult={!!resultData}
				rowThreatOnly={isThreatOnly}
			/>
			<Tooltip id={TOOLTIP.avgCastHeader} />
			<MetricTooltip id={TOOLTIP.damage} forAnchor={forAnchor} body={damageBreakdownTooltip} />
			<MetricTooltip id={TOOLTIP.casts} forAnchor={forAnchor} body={castsTooltip} />
			<MetricTooltip
				id={TOOLTIP.avgCast}
				forAnchor={forAnchor}
				body={metric => (!metric.avgCastHit && !metric.avgCastTick ? null : threatTooltip(metric, metric.avgCastThreat, showThreatMetrics))}
			/>
			<MetricTooltip id={TOOLTIP.hits} forAnchor={forAnchor} body={hitsTooltip} />
			<MetricTooltip id={TOOLTIP.avgHit} forAnchor={forAnchor} body={metric => threatTooltip(metric, metric.avgHitThreat, showThreatMetrics)} />
			<MetricTooltip id={TOOLTIP.missPercent} forAnchor={forAnchor} body={missTooltip} />
			<MetricTooltip id={TOOLTIP.dps} forAnchor={forAnchor} body={metric => (metric.dps ? threatTooltip(metric, metric.tps, showThreatMetrics) : null)} />
		</>
	);
};
