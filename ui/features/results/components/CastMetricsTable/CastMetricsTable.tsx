import i18n from '@i18n/config';
import { ActionMetrics, type UnitMetrics } from '@sim/proto/sim_result';
import { useMemo } from 'react';

import { useHidePassives } from '../../context/HidePassivesContext';
import { useActionRows } from '../../hooks/useActionRows';
import type { SimResultData } from '../../model/result_data';
import { createMetricsColumnHelper, MetricsActionCell, MetricsTable } from '../MetricsTable';

const helper = createMetricsColumnHelper<ActionMetrics>();

// One pair of groups functions per toggle state, built once, so `useActionRows` sees stable functions.
const castGroupsOf = (hidePassives: boolean) => {
	const castsOf = (unit: UnitMetrics, resultData: SimResultData) =>
		unit.actions.filter(action => action.casts != 0 && (!hidePassives || action.isPressed)).map(action => action.forTarget(resultData.filter));
	const playerOf = (resultData: SimResultData): UnitMetrics | undefined => resultData.result.getRaidIndexedPlayers(resultData.filter)[0];

	return {
		player: (resultData: SimResultData): Array<Array<ActionMetrics>> => {
			const player = playerOf(resultData);
			return player ? ActionMetrics.groupById(castsOf(player, resultData)) : [];
		},
		pets: (resultData: SimResultData): Array<Array<ActionMetrics>> =>
			(playerOf(resultData)?.pets ?? []).map(pet => castsOf(pet, resultData)).filter(group => group.length),
	};
};

const WITHOUT_PASSIVES = castGroupsOf(true);
const WITH_PASSIVES = castGroupsOf(false);

export const CastMetricsTable = () => {
	const groups = useHidePassives() ? WITHOUT_PASSIVES : WITH_PASSIVES;
	const { resultData, rows } = useActionRows(groups.player);
	const { rows: petRows } = useActionRows(groups.pets);

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
				helper.accessor(row => row.metric.casts, {
					id: 'casts',
					header: i18n.t('results_tab.details.columns.casts'),
					cell: info => info.getValue().toFixed(1),
				}),
				helper.accessor(row => row.metric.castsPerMinute, {
					id: 'cpm',
					header: i18n.t('results_tab.details.columns.cpm'),
					cell: info => info.getValue().toFixed(1),
				}),
			]),
		[],
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
