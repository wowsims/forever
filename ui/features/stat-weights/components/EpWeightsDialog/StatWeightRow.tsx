import type { StatWeightsResult, StatWeightValues } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import type { DisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { scaledEpValue, type Stats, type UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { subscribeStatWeightsChange } from '@sim/state/subscriptions';
import { sanitizeId } from '@sim/utils/format';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { TableCell, TableRow } from '@ui-kit/Table';
import { NumberField } from '@ui-kit/TextField';
import clsx from 'clsx';

import { epRatingScale } from '../../model/ep_unit_stats';
import { StatWeightCells } from './StatWeightCells';

/** One shown metric's pair of cells: the weights it computed (none before a run, or for an excluded stat) and its EP ratio. */
export interface StatWeightRowMetric {
	ratioIndex: number;
	statWeights?: StatWeightValues;
	epRatio: number;
}

export interface StatWeightRowInclude {
	value: boolean;
	disabled: boolean;
	onChange: (included: boolean) => void;
}

export interface StatWeightRowViewProps {
	stat: UnitStat;
	fullName: string;
	shortName: string;
	metrics: StatWeightRowMetric[];
	iterations: number;
	/** The row's weighted EP minus its current EP weight, which colours the EP cells. */
	epDelta: number;
	currentEp: number;
	onCurrentEpChange: (value: number) => void;
	/** Absent for a stat outside the spec's EP set: the Update cell stays empty. */
	include?: StatWeightRowInclude;
	showThreatMetrics: boolean;
	showUpdateColumn?: boolean;
}

export const StatWeightRowView = ({
	stat,
	fullName,
	shortName,
	metrics,
	iterations,
	epDelta,
	currentEp,
	onCurrentEpChange,
	include,
	showThreatMetrics,
	showUpdateColumn = true,
}: StatWeightRowViewProps) => {
	const cellClassName = clsx('ui-ep-weights-table-cell', showThreatMetrics && 'max-lg:pl-0');

	return (
		<TableRow>
			<TableCell className={cellClassName}>{fullName}</TableCell>
			{showUpdateColumn && (
				<TableCell testId="swcalc-include-toggle" className={cellClassName}>
					{include && (
						<BooleanPicker
							id={`sw-stat-toggle-${sanitizeId(fullName)}`}
							value={include.value}
							disabled={include.disabled}
							onChange={include.onChange}
						/>
					)}
				</TableCell>
			)}
			{metrics.map(({ statWeights, ratioIndex, epRatio }) => (
				<StatWeightCells
					key={ratioIndex}
					stat={stat}
					statWeights={statWeights}
					iterations={iterations}
					epRatio={epRatio}
					epDelta={epDelta}
					cellClassName={cellClassName}
				/>
			))}
			<TableCell testId="current-ep" className={cellClassName}>
				<NumberField
					id={`ep-weight-stat-${sanitizeId(shortName)}`}
					float
					value={currentEp}
					onChange={onCurrentEpChange}
					inputClassName={showThreatMetrics ? 'ui-ep-weights-compact-input' : undefined}
				/>
			</TableCell>
		</TableRow>
	);
};

export interface StatWeightRowProps {
	stat: UnitStat;
	result: StatWeightsResult | null;
	iterations: number;
	epRatios: number[];
	epWeights: Stats;
	settings: StatWeightActionSettings;
	player: Player<any>;
	epReferenceStat: Stat;
	includable: boolean;
	showThreatMetrics: boolean;
	displayMetrics: DisplayMetrics;
	showUpdateColumn?: boolean;
}

export const StatWeightRow = ({
	stat,
	result,
	iterations,
	epRatios,
	epWeights,
	settings,
	player,
	epReferenceStat,
	includable,
	showThreatMetrics,
	displayMetrics,
	showUpdateColumn = true,
}: StatWeightRowProps) => {
	const excluded = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.isUnitStatExcludedFromCalc(stat));
	const rowResult = excluded ? null : result;
	const allMetrics: { statWeights: StatWeightValues | undefined; metric: keyof DisplayMetrics }[] = [
		{ statWeights: rowResult?.dps, metric: 'damage' },
		{ statWeights: rowResult?.hps, metric: 'healing' },
		{ statWeights: rowResult?.tps, metric: 'threat' },
		{ statWeights: rowResult?.dtps, metric: 'threat' },
		{ statWeights: rowResult?.tmi, metric: 'threat' },
		{ statWeights: rowResult?.pDeath, metric: 'threat' },
	];
	const metrics = allMetrics
		.map(({ statWeights, metric }, ratioIndex) => ({ statWeights, metric, ratioIndex, epRatio: epRatios[ratioIndex] }))
		.filter(entry => displayMetrics[entry.metric]);
	const scale = epRatingScale(stat);
	const currentEp = epWeights.getUnitStat(stat);
	// A result without per-stat counts falls back to the run's iteration count.
	const samples = (rowResult?.sampleCounts && stat.getProtoValue(rowResult.sampleCounts)) || iterations;

	return (
		<StatWeightRowView
			stat={stat}
			fullName={stat.getFullName(player.getClass())}
			shortName={stat.getShortName(player.playerClass.classID)}
			metrics={metrics}
			iterations={samples}
			epDelta={(scaledEpValue(stat, epRatios, rowResult) - currentEp) / scale}
			currentEp={currentEp / scale}
			onCurrentEpChange={value => player.setEpWeights(player.getEpWeights().withUnitStat(stat, value * scale))}
			include={
				includable
					? {
							value: !excluded,
							disabled: stat.isStat() && epReferenceStat === stat.getStat(),
							onChange: included => settings.setStatExcluded(stat, !included),
						}
					: undefined
			}
			showThreatMetrics={showThreatMetrics}
			showUpdateColumn={showUpdateColumn}
		/>
	);
};
