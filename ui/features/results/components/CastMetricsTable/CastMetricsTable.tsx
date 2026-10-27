import i18n from '@i18n/config';
import { ActionMetrics, type UnitMetrics } from '@sim/proto/sim_result';
import { useMemo } from 'react';

import { useShowApmBreakdown } from '../../context/ShowApmBreakdownContext';
import { useActionRows } from '../../hooks/useActionRows';
import type { SimResultData } from '../../model/result_data';
import { createMetricsColumnHelper, MetricsActionCell, MetricsTable } from '../MetricsTable';

const helper = createMetricsColumnHelper<ActionMetrics>();

const playerOf = (resultData: SimResultData): UnitMetrics | undefined => {
	const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
	return players.find(player => player.spec?.isHealingSpec) ?? players[0];
};

const castsOf = (unit: UnitMetrics, resultData: SimResultData) =>
	unit.actions.map(action => action.forTarget(resultData.filter)).filter(action => action.casts != 0);

// One mode per toggle state, built once, so `useActionRows` sees stable functions.
const ALL_CASTS = {
	count: (metric: ActionMetrics) => metric.casts,
	perMinute: (metric: ActionMetrics) => metric.castsPerMinute,
	player: (resultData: SimResultData): Array<Array<ActionMetrics>> => {
		const player = playerOf(resultData);
		return player ? ActionMetrics.groupById(castsOf(player, resultData)) : [];
	},
	pets: (resultData: SimResultData): Array<Array<ActionMetrics>> =>
		(playerOf(resultData)?.pets ?? []).map(pet => castsOf(pet, resultData)).filter(group => group.length),
};

const APM_BREAKDOWN: typeof ALL_CASTS = {
	count: metric => metric.manualCasts,
	perMinute: metric => metric.manualCastsPerMinute,
	player: resultData => {
		const player = playerOf(resultData);
		return player ? ActionMetrics.groupById(player.apmActions(resultData.filter)) : [];
	},
	// APM counts only the player's presses, so the breakdown has no pet table.
	pets: () => [],
};

export const CastMetricsTable = () => {
	const mode = useShowApmBreakdown() ? APM_BREAKDOWN : ALL_CASTS;
	const { resultData, rows } = useActionRows(mode.player);
	const { rows: petRows } = useActionRows(mode.pets);

	const columns = useMemo(
		() =>
			helper.columns([
				helper.accessor(row => row.metric.name, {
					id: 'name',
					header: i18n.t('results_tab.details.columns.name'),
					cell: info => (
						<MetricsActionCell
							name={info.row.original.metric.name}
							actionId={info.row.original.metric.actionId}
							expandable={info.row.getCanExpand()}
							expanded={info.row.getIsExpanded()}
							onToggle={info.row.getToggleExpandedHandler()}
						/>
					),
				}),
				helper.accessor(row => mode.count(row.metric), {
					id: 'casts',
					header: i18n.t('results_tab.details.columns.casts'),
					cell: info => info.getValue().toFixed(1),
				}),
				helper.accessor(row => mode.perMinute(row.metric), {
					id: 'cpm',
					header: i18n.t('results_tab.details.columns.cpm'),
					cell: info => info.getValue().toFixed(1),
				}),
			]),
		[mode],
	);

	return (
		<>
			<MetricsTable rootTestId="cast-metrics-root" columns={columns} rows={rows} sortColumnId="casts" hasResult={!!resultData} />
			{petRows.length > 0 && (
				<div data-testid="pet-cast-metrics-container" className="mt-4">
					<span data-testid="pet-cast-metrics-title" className="text-lg font-bold text-white">
						{i18n.t('results_tab.details.pets')}
					</span>
					<MetricsTable rootTestId="pet-cast-metrics-root" columns={columns} rows={petRows} sortColumnId="casts" hasResult={!!resultData} />
				</div>
			)}
		</>
	);
};
