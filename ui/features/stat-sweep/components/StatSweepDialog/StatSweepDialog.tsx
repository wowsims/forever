import { PrecisionOptions } from '@features/stat-weights/components/PrecisionOptions';
import { EP_UNIT_STATS, isEpStat } from '@features/stat-weights/model/ep_unit_stats';
import i18n from '@i18n/config';
import { usePlayer, useSim, useSpecConfig } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats } from '@sim/proto/stats';
import { precisionRunnable } from '@sim/settings/precision';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import { SimRunKind } from '@sim/state/sim_store';
import { subscribeStatSweepChange } from '@sim/state/subscriptions';
import { sum } from '@sim/utils/math';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import {
	HandleProgressStepList,
	type ProgressStepListHandle,
	ProgressTrackerDialog,
	type ProgressTrackerHandle,
	type ProgressTrackerProgress,
} from '@ui-kit/ProgressTrackerDialog';
import { downloadString } from '@ui-kit/utils/dom';
import { type Ref, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { useStatSweep } from '../../hooks/useStatSweep';
import { offeredSweepMetrics, shownSweepMetric, statSweepCsv, type StatSweepProgress, sweepStepPosition } from '../../model/stat_sweep';
import { StatSweepChart } from '../StatSweepChart';
import { StatSweepHistory } from '../StatSweepHistory';
import { StatSweepPanel } from '../StatSweepPanel';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);

const CSV_FILE_NAME = 'wowsims-stat-sweep.csv';

// The bar spans every stat's steps, read out in iterations at the running step's expected count; before a step has one it counts steps.
const trackerProgress = ({ stepCounts, running, completed, iterationsPerStep, secondsRemaining }: StatSweepProgress): ProgressTrackerProgress => {
	const { stat, step } = sweepStepPosition(stepCounts, running);
	const scale = iterationsPerStep || 1;
	return {
		title: t('progress', { current: step + 1, total: stepCounts[stat] }),
		current: completed * scale,
		total: sum(stepCounts) * scale,
		secondsRemaining: secondsRemaining ?? null,
	};
};

// The sweep the settings and the spec describe.
const useSweepSetup = (settings: StatSweepSettings) => {
	const player = usePlayer();
	const { epStats, epPseudoStats } = useSpecConfig();
	const displayMetrics = useDisplayMetrics(useSim());
	const metrics = useMemo(() => offeredSweepMetrics(displayMetrics), [displayMetrics]);

	const unitStats = useMemo(() => EP_UNIT_STATS.filter(stat => isEpStat(stat, { epStats, epPseudoStats: epPseudoStats ?? [] })), [epStats, epPseudoStats]);
	const statOptions = useMemo(() => unitStats.map(stat => ({ value: stat.getKey(), label: stat.getFullName(player.getClass()) })), [unitStats, player]);
	const values = useStoreSubscribe(subscribeStatSweepChange(settings), () => settings.values);
	const { stats: storedStats, range, step, showSpread, precisionMode, iterations, targetCi, maxIterationMultiplier } = values;
	const precision = { precisionMode, iterations, targetCi, maxIterationMultiplier };
	const metric = shownSweepMetric(values.metric, metrics);
	// The stored stats the spec offers, in the order offered; with none stored, or none of them offered, the first one offered. A cleared list stays empty.
	const selectedStats = useMemo(() => {
		if (storedStats?.length === 0) return [];
		const offered = unitStats.filter(stat => storedStats?.some(stored => stored.equals(stat)));
		return offered.length ? offered : unitStats.slice(0, 1);
	}, [unitStats, storedStats]);
	const selectedKeys = useMemo(() => selectedStats.map(stat => stat.getKey()), [selectedStats]);
	const labelOf = (key: string) => statOptions.find(option => option.value === key)?.label ?? '';
	const runnable = selectedStats.length > 0 && precisionRunnable(precision);
	return { player, unitStats, statOptions, metrics, metric, range, step, showSpread, precision, selectedStats, selectedKeys, labelOf, runnable };
};

interface StatSweepBodyHandle {
	run: () => void;
}

interface RunState {
	runnable: boolean;
	isRunning: boolean;
}

interface StatSweepBodyProps {
	settings: StatSweepSettings;
	ref: Ref<StatSweepBodyHandle>;
	onRunStateChange: (state: RunState) => void;
}

// The dialog's popup mounts this on open and drops it once the close transition ends, so the setup is read, the history
// loads and a run goes only while the dialog shows.
const StatSweepBody = ({ settings, ref, onRunStateChange }: StatSweepBodyProps) => {
	const { player, unitStats, statOptions, metrics, metric, range, step, showSpread, precision, selectedStats, selectedKeys, labelOf, runnable } =
		useSweepSetup(settings);
	const progressRef = useRef<ProgressTrackerHandle>(null);
	const statListRef = useRef<ProgressStepListHandle>(null);
	const sweep = useStatSweep();
	const defaultIterations = useSimStore('iterations');

	useEffect(() => onRunStateChange({ runnable, isRunning: sweep.isRunning }), [onRunStateChange, runnable, sweep.isRunning]);

	useImperativeHandle(ref, () => ({
		run: () => {
			if (!runnable) return;
			const finalStats = player.getCurrentStats()?.finalStats;
			const totals = finalStats ? Stats.fromProto(finalStats) : null;
			sweep
				.run(
					{
						stats: selectedStats.map(unitStat => ({
							unitStat,
							label: t('run_label', { stat: labelOf(unitStat.getKey()), range }),
							baseStatValue: totals?.getUnitStat(unitStat) ?? 0,
						})),
						range,
						step,
						metric,
						precision,
						batchLabel: t('batch_label', { stats: selectedKeys.map(labelOf).join(', '), range }),
					},
					progress => {
						progressRef.current?.setProgress(trackerProgress(progress));
						statListRef.current?.setRunning(sweepStepPosition(progress.stepCounts, progress.running).stat);
					},
				)
				.catch(console.error);
		},
	}));

	const onExportCsv = sweep.series.some(run => run.points.length) ? () => downloadString(statSweepCsv(sweep.series), CSV_FILE_NAME, 'text/csv') : undefined;
	const metricLabel = i18n.t(`sidebar.results.metrics.${metric}.label`);
	const precisionText = (key: string) => t(`precision.${key}`, { metric: metricLabel });

	return (
		<>
			<StatSweepPanel
				statOptions={statOptions}
				stats={selectedKeys}
				onStatsChange={keys => settings.set({ stats: unitStats.filter(stat => keys.includes(stat.getKey())) })}
				range={range}
				onRangeChange={range => settings.set({ range })}
				step={step}
				onStepChange={step => settings.set({ step })}
				precision={
					<PrecisionOptions
						value={precision}
						onChange={patch => settings.set(patch)}
						defaultIterations={defaultIterations}
						idPrefix="stat-sweep-precision"
						targetLabel={precisionText('target_ci')}
						targetTooltip={precisionText('target_ci_tooltip')}
						iterationsTooltip={precisionText('iterations_tooltip')}
						maxMultiplierTooltip={precisionText('max_multiplier_tooltip')}
						maxDecimalDigits={2}
						disabled={sweep.isRunning}
					/>
				}
				chart={
					<StatSweepChart
						series={sweep.series}
						metric={metric}
						metrics={metrics}
						onMetricChange={metric => settings.set({ metric })}
						showSpread={showSpread}
						onShowSpreadChange={showSpread => settings.set({ showSpread })}
					/>
				}
				history={
					<StatSweepHistory batches={sweep.historyBatches} onOverlayChange={sweep.setOverlaid} onDelete={sweep.remove} onExportCsv={onExportCsv} />
				}
				running={sweep.isRunning}
			/>
			{sweep.isRunning && (
				<ProgressTrackerDialog
					ref={progressRef}
					open
					testId="stat-sweep-progress-dialog"
					title={t('title')}
					state={{ stage: 'running' }}
					hasProgressBar
					belowBar={<HandleProgressStepList ref={statListRef} labels={selectedKeys.map(labelOf)} />}
					onCancel={sweep.cancel}
				/>
			)}
		</>
	);
};

export interface StatSweepDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	settings: StatSweepSettings;
}

/** Closing it cancels a run. */
export const StatSweepDialog = ({ open, onOpenChange, settings }: StatSweepDialogProps) => {
	const sim = useSim();
	const bodyRef = useRef<StatSweepBodyHandle>(null);
	const [runState, setRunState] = useState<RunState>({ runnable: false, isRunning: false });

	return (
		<Dialog
			open={open}
			onOpenChange={next => {
				if (!next) sim.runs.abort(SimRunKind.StatSweep).catch(console.error);
				onOpenChange(next);
			}}
			testId="stat-sweep-dialog"
			size="full"
			scrollContents
			title={t('title')}
			footer={
				<Button variant="primary" disabled={runState.isRunning || !runState.runnable} onClick={() => bodyRef.current?.run()}>
					{t('run')}
				</Button>
			}>
			<StatSweepBody ref={bodyRef} settings={settings} onRunStateChange={setRunState} />
		</Dialog>
	);
};
