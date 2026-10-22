import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { ActionId } from '@sim/proto/action_id';
import { ActionMetrics } from '@sim/proto/sim_result';
import { formatToNumber } from '@sim/utils/format';
import { percentOf } from '@sim/utils/math';
import { Tooltip } from '@ui-kit/Tooltip';
import { useMemo } from 'react';

import { useActionRows } from '../../hooks/useActionRows';
import { filterMetricRows, indexMetricRows } from '../../model/grouping';
import type { SimResultData } from '../../model/result_data';
import { amountHeader, attackFormat, attackMetricsColumns, customMetricsGroups, threatTooltip, useMetricMax } from '../AttackMetricsColumns';
import { MetricsCombinedTooltip, type MetricsCombinedTooltipGroup } from '../MetricsCombinedTooltip';
import { createMetricsColumnHelper, metricForAnchor, MetricsActionCell, MetricsTable, MetricTooltip } from '../MetricsTable';

const helper = createMetricsColumnHelper<ActionMetrics>();

const TOOLTIP = {
	healing: 'healing-metrics-healing',
	avgCast: 'healing-metrics-avg-cast',
	avgCastHeader: 'healing-metrics-avg-cast-header',
	hits: 'healing-metrics-hits',
	hitsHeader: 'healing-metrics-hits-header',
	avgHit: 'healing-metrics-avg-hit',
	avgHitHeader: 'healing-metrics-avg-hit-header',
	hps: 'healing-metrics-hps',
};

const sourceActionTag = (actionId: ActionId) => (actionId.itemId ? -actionId.itemId : actionId.spellId);

// Tags can also identify internal actions, so treat a tag as a proc source only when it resolves to an action metric.
const sourceActionIdsFor = (resultData: SimResultData): Set<number> => {
	const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
	if (!players.length) return new Set();
	const player = players[0];
	const availableSourceActionTags = new Set([...player.getDamageActions(), ...player.getHealingActions()].map(action => sourceActionTag(action.actionId)));
	return new Set(
		player
			.getHealingActions()
			.map(action => action.forTarget(resultData.filter))
			.filter(action => action.actionId.tag !== 0 && availableSourceActionTags.has(action.actionId.tag))
			.map(action => action.actionId.spellId),
	);
};

const isSourceActionMetric = (action: ActionMetrics, sourceActionIds: ReadonlySet<number>) => sourceActionIds.has(action.actionId.spellId);

const healingGroups = (resultData: SimResultData): Array<Array<ActionMetrics>> => {
	const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
	if (!players.length) return [];
	const sourceActionIds = sourceActionIdsFor(resultData);

	const actions = players[0]
		.getHealingActions()
		.map(action => action.forTarget(resultData.filter))
		.filter(action => action.hps > 0);
	return ActionMetrics.groupById(actions.filter(action => !isSourceActionMetric(action, sourceActionIds)));
};

const healingSourceMetrics = (resultData: SimResultData | null) => {
	if (!resultData) return { byProcId: new Map<number, Array<ActionMetrics>>(), sourceActionIds: new Set<number>() };
	const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
	if (!players.length) return { byProcId: new Map<number, Array<ActionMetrics>>(), sourceActionIds: new Set<number>() };
	const sourceActionIds = sourceActionIdsFor(resultData);

	const sourceActions = players[0]
		.getHealingActions()
		.map(action => action.forTarget(resultData.filter))
		.filter(action => action.hps > 0 && isSourceActionMetric(action, sourceActionIds));
	return {
		sourceActionIds,
		byProcId: new Map(
			[...sourceActionIds].map(
				sourceActionId =>
					[
						sourceActionId,
						ActionMetrics.groupById(
							sourceActions.filter(action => action.actionId.spellId === sourceActionId),
							true,
						).map(group => ActionMetrics.merge(group)),
					] as const,
			),
		),
	};
};

const healingGroup = (metric: ActionMetrics): MetricsCombinedTooltipGroup => {
	const done = metric.healingDone;
	return {
		spellSchool: metric.spellSchool,
		totalPercentage: 100,
		data: [
			{ name: i18n.t('results_tab.details.attack_types.hit'), ...done.heal },
			{ name: i18n.t('results_tab.details.attack_types.critical_hit'), ...done.critHeal },
			{ name: i18n.t('results_tab.details.attack_types.tick'), ...done.tick },
			{ name: i18n.t('results_tab.details.attack_types.critical_tick'), ...done.critTick },
			{ name: i18n.t('results_tab.details.attack_types.shield'), ...done.shield },
		],
	};
};

const healingHitGroups = (metric: ActionMetrics): Array<MetricsCombinedTooltipGroup> => [
	{
		spellSchool: metric.spellSchool,
		totalPercentage: 100,
		name: i18n.t('results_tab.details.columns.hits'),
		data: [
			{ name: i18n.t('results_tab.details.attack_types.hit'), value: metric.hits, percentage: percentOf(metric.hits, metric.landedHits) },
			{ name: i18n.t('results_tab.details.attack_types.critical_hit'), value: metric.crits, percentage: percentOf(metric.crits, metric.landedHits) },
			{ name: i18n.t('results_tab.details.attack_types.glancing_blow'), value: metric.glances, percentage: percentOf(metric.glances, metric.landedHits) },
			{ name: i18n.t('results_tab.details.attack_types.blocked_hit'), value: metric.blocks, percentage: percentOf(metric.blocks, metric.landedHits) },
			{
				name: i18n.t('results_tab.details.attack_types.blocked_critical_hit'),
				value: metric.blockedCrits,
				percentage: percentOf(metric.blockedCrits, metric.landedHits),
			},
		],
	},
	{
		spellSchool: metric.spellSchool,
		totalPercentage: 100,
		name: i18n.t('results_tab.details.tooltip_table.ticks'),
		data: [
			{ name: i18n.t('results_tab.details.attack_types.tick'), value: metric.ticks, percentage: percentOf(metric.ticks, metric.landedTicks) },
			{
				name: i18n.t('results_tab.details.attack_types.critical_tick'),
				value: metric.critTicks,
				percentage: percentOf(metric.critTicks, metric.landedTicks),
			},
		],
	},
];

const isThreatOnly = (metric: ActionMetrics) => metric.hitAttempts == 0 && metric.hps == 0;

export const HealingMetricsTable = () => {
	const sim = useSim();
	const { threat: showThreatMetrics } = useDisplayMetrics(sim);
	const groupsOf = useMemo(() => (resultData: SimResultData) => healingGroups(resultData), []);
	const { resultData, rows: actionRows } = useActionRows(groupsOf);
	const sourceMetrics = useMemo(() => healingSourceMetrics(resultData), [resultData]);
	const sourceMetricsByProcId = sourceMetrics.byProcId;
	const sourceRows = useMemo(
		() =>
			[...sourceMetricsByProcId.values()].flatMap(metrics =>
				metrics.length
					? [
							{
								metric: ActionMetrics.merge(metrics, { removeTag: true }),
								subRows: metrics.map(metric => ({ metric })),
							},
						]
					: [],
			),
		[sourceMetricsByProcId],
	);
	const built = useMemo(() => [...actionRows, ...sourceRows], [actionRows, sourceRows]);
	const sourceActionsByType = useMemo(() => {
		const players = resultData?.result.getRaidIndexedPlayers(resultData.filter) ?? [];
		const player = players[0];
		if (!player || !resultData) return { damage: new Map(), healing: new Map() };
		const byTag = (actions: Array<ActionMetrics>) =>
			new Map(
				actions
					.map(action => action.forTarget(resultData.filter))
					.filter(action => !sourceMetrics.sourceActionIds.has(action.actionId.spellId))
					.map(action => [sourceActionTag(action.actionId), action] as const)
					.filter(([tag]) => tag !== 0),
			);
		return {
			damage: byTag(player?.getDamageActions() ?? []),
			healing: byTag(player?.getHealingActions() ?? []),
		};
	}, [resultData, sourceMetrics.sourceActionIds]);

	const rows = useMemo(() => (showThreatMetrics ? built : filterMetricRows(built, isThreatOnly)), [built, showThreatMetrics]);
	const metricsByRowId = useMemo(() => indexMetricRows(rows), [rows]);
	const maxHealing = useMetricMax(rows, metric => metric.healing);

	const columns = useMemo(
		() =>
			helper.columns([
				helper.accessor(row => row.metric.name, {
					id: 'name',
					header: i18n.t('results_tab.details.columns.name'),
					cell: info => {
						const metric = info.row.original.metric;
						const isProcParent = isSourceActionMetric(metric, sourceMetrics.sourceActionIds) && metric.actionId.tag === 0;
						const sourceAction = isProcParent
							? undefined
							: (sourceActionsByType.damage.get(metric.actionId.tag) ?? sourceActionsByType.healing.get(metric.actionId.tag));
						return (
							<MetricsActionCell
								name={isProcParent ? metric.actionId.withoutTag().name : (sourceAction?.name ?? metric.name)}
								actionId={sourceAction?.actionId ?? metric.actionId}
								expandable={info.row.getCanExpand()}
								expanded={info.row.getIsExpanded()}
								onToggle={info.row.getToggleExpandedHandler()}
							/>
						);
					},
				}),
				attackMetricsColumns.primary({
					id: 'healing-done',
					header: i18n.t('results_tab.details.columns.healing_done'),
					total: metric => metric.avgHealing,
					value: metric => metric.healing,
					percentage: metric => metric.totalHealingPercent,
					max: maxHealing,
					overlay: metric => metric.shielding,
					tooltipId: TOOLTIP.healing,
				}),
				attackMetricsColumns.casts(),
				helper.accessor(row => row.metric.castsPerMinute, {
					id: 'cpm',
					header: i18n.t('results_tab.details.columns.cpm'),
					cell: info => attackFormat.number(info.getValue()),
				}),
				helper.accessor(row => row.metric.avgCastTimeMs, {
					id: 'cast-time',
					header: i18n.t('results_tab.details.columns.cast_time'),
					cell: info => formatToNumber(info.getValue() / 1000, { minimumFractionDigits: 2, fallbackString: '-' }),
				}),
				helper.accessor(row => row.metric.avgCastHealing, {
					id: 'avg-cast',
					header: i18n.t('results_tab.details.columns.avg_cast'),
					meta: {
						tooltipId: TOOLTIP.avgCast,
						headerTooltipId: TOOLTIP.avgCastHeader,
						headerTooltip: i18n.t('results_tab.details.tooltips.healing_avg_cast_tooltip'),
					},
					cell: info => attackFormat.compact(info.getValue()),
				}),
				helper.accessor(row => row.metric.landedHits, {
					id: 'hits',
					header: i18n.t('results_tab.details.columns.hits'),
					meta: {
						tooltipId: TOOLTIP.hits,
						headerTooltipId: TOOLTIP.hitsHeader,
						headerTooltip: i18n.t('results_tab.details.tooltips.healing_hits_tooltip'),
					},
					cell: info => {
						const metric = info.row.original.metric;
						return (
							<>
								{attackFormat.number(metric.landedHits)}
								{!!metric.landedTicks && <> ({attackFormat.number(metric.landedTicks)})</>}{' '}
							</>
						);
					},
				}),
				helper.accessor(row => row.metric.avgHitHealing, {
					id: 'avg-hit',
					header: i18n.t('results_tab.details.columns.avg_hit'),
					meta: {
						tooltipId: TOOLTIP.avgHit,
						headerTooltipId: TOOLTIP.avgHitHeader,
						headerTooltip: i18n.t('results_tab.details.tooltips.healing_avg_hit_tooltip'),
					},
					cell: info => attackFormat.compact(info.getValue()),
				}),
				helper.accessor(row => row.metric.hpm, {
					id: 'hpm',
					header: i18n.t('results_tab.details.columns.hpm'),
					cell: info => attackFormat.compact(info.getValue()),
				}),
				helper.accessor(row => row.metric.critPercent || row.metric.critTickPercent, {
					id: 'crit-percent',
					header: i18n.t('results_tab.details.columns.crit_percent'),
					cell: info => attackFormat.percent(info.getValue()),
				}),
				helper.accessor(row => row.metric.healingThroughput, {
					id: 'hpet',
					header: i18n.t('results_tab.details.columns.hpet'),
					cell: info => attackFormat.compact(info.getValue()),
				}),
				attackMetricsColumns.rate({
					id: 'hps',
					header: i18n.t('results_tab.details.columns.hps'),
					value: metric => metric.hps,
					tooltipId: TOOLTIP.hps,
				}),
			]),
		[maxHealing, sourceMetrics.sourceActionIds, sourceActionsByType],
	);

	const forAnchor = metricForAnchor(metricsByRowId);

	return (
		<>
			<MetricsTable
				rootTestId="healing-metrics-root"
				columns={columns}
				rows={rows}
				sortColumnId="hps"
				hasResult={!!resultData}
				rowThreatOnly={isThreatOnly}
			/>
			<Tooltip id={TOOLTIP.avgCastHeader} />
			<Tooltip id={TOOLTIP.hitsHeader} />
			<Tooltip id={TOOLTIP.avgHitHeader} />
			<MetricTooltip
				id={TOOLTIP.healing}
				forAnchor={forAnchor}
				body={metric => <MetricsCombinedTooltip headerValues={amountHeader()} groups={[healingGroup(metric), ...customMetricsGroups(metric)]} />}
			/>
			<MetricTooltip
				id={TOOLTIP.avgCast}
				forAnchor={forAnchor}
				body={metric => (metric.avgCastHealing ? threatTooltip(metric, metric.avgCastThreat, showThreatMetrics) : null)}
			/>
			<MetricTooltip
				id={TOOLTIP.hits}
				forAnchor={forAnchor}
				body={metric => (!metric.landedHits && !metric.landedTicks ? null : <MetricsCombinedTooltip groups={healingHitGroups(metric)} />)}
			/>
			<MetricTooltip
				id={TOOLTIP.avgHit}
				forAnchor={forAnchor}
				body={metric => (metric.avgHitHealing ? threatTooltip(metric, metric.avgHitThreat, showThreatMetrics) : null)}
			/>
			<MetricTooltip id={TOOLTIP.hps} forAnchor={forAnchor} body={metric => threatTooltip(metric, metric.tps, showThreatMetrics)} />
		</>
	);
};
