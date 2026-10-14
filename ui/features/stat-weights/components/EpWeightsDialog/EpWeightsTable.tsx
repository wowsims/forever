import type { StatWeightsResult } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import type { DisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { Player } from '@sim/player/player';
import type { Stats, UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { Table } from '@ui-kit/Table';
import { type ReactNode, useCallback } from 'react';

import type { EpStatSet } from '../../model/ep_unit_stats';
import { isEpStat } from '../../model/ep_unit_stats';
import { EpRatiosRow } from './EpRatiosRow';
import { EpWeightsHeader } from './EpWeightsHeader';
import { StatWeightRow } from './StatWeightRow';
import type { EpColumn, StatsType } from './types';

export interface EpWeightsTableViewProps {
	columns: EpColumn[];
	statsType: StatsType;
	epRatios: number[];
	onEpRatioChange: (ratioIndex: number, value: number) => void;
	onComputeEp: () => void;
	showThreatMetrics: boolean;
	showEpRatios: boolean;
	/** The Update column picks which stats a calculation writes; a gear planner has no calculation. */
	showUpdateColumn?: boolean;
	statAction?: ReactNode;
	children: ReactNode;
}

export const EpWeightsTableView = ({
	columns,
	statsType,
	epRatios,
	onEpRatioChange,
	onComputeEp,
	showThreatMetrics,
	showEpRatios,
	showUpdateColumn = true,
	statAction,
	children,
}: EpWeightsTableViewProps) => (
	<div data-testid="results-ep-table-container" className="relative flex-1 overflow-y-auto">
		<Table
			testId="results-ep-table"
			striped="solid"
			className={[showThreatMetrics && 'max-lg:pr-0', `stats-type-${statsType}`]}
			data-stats-type={statsType}>
			<thead>
				<EpWeightsHeader columns={columns} showThreatMetrics={showThreatMetrics} showUpdateColumn={showUpdateColumn} statAction={statAction} />
				{showEpRatios && (
					<EpRatiosRow
						columns={columns}
						epRatios={epRatios}
						onEpRatioChange={onEpRatioChange}
						onComputeEp={onComputeEp}
						showThreatMetrics={showThreatMetrics}
					/>
				)}
			</thead>
			<tbody>{children}</tbody>
		</Table>
	</div>
);

export interface EpWeightsTableProps {
	columns: EpColumn[];
	stats: UnitStat[];
	statsType: StatsType;
	result: StatWeightsResult | null;
	iterations: number;
	epRatios: number[];
	epWeights: Stats;
	settings: StatWeightActionSettings;
	player: Player<any>;
	epStatSet: EpStatSet;
	epReferenceStat: Stat;
	onComputeEp: () => void;
	showThreatMetrics: boolean;
	showEpRatios: boolean;
	displayMetrics: DisplayMetrics;
	showUpdateColumn?: boolean;
	statAction?: ReactNode;
}

export const EpWeightsTable = ({
	columns,
	stats,
	statsType,
	result,
	iterations,
	epRatios,
	epWeights,
	settings,
	player,
	epStatSet,
	epReferenceStat,
	onComputeEp,
	showThreatMetrics,
	showEpRatios,
	displayMetrics,
	showUpdateColumn = true,
	statAction,
}: EpWeightsTableProps) => {
	const onEpRatioChange = useCallback(
		(ratioIndex: number, value: number) => {
			const next = player.getEpRatios();
			next[ratioIndex] = value;
			player.setEpRatios(next);
		},
		[player],
	);

	return (
		<EpWeightsTableView
			columns={columns}
			statsType={statsType}
			epRatios={epRatios}
			onEpRatioChange={onEpRatioChange}
			onComputeEp={onComputeEp}
			showThreatMetrics={showThreatMetrics}
			showEpRatios={showEpRatios}
			showUpdateColumn={showUpdateColumn}
			statAction={statAction}>
			{stats.map(stat => (
				<StatWeightRow
					key={stat.getFullName(player.getClass())}
					stat={stat}
					result={result}
					iterations={iterations}
					epRatios={epRatios}
					epWeights={epWeights}
					settings={settings}
					player={player}
					epReferenceStat={epReferenceStat}
					includable={isEpStat(stat, epStatSet)}
					showThreatMetrics={showThreatMetrics}
					displayMetrics={displayMetrics}
					showUpdateColumn={showUpdateColumn}
				/>
			))}
		</EpWeightsTableView>
	);
};
