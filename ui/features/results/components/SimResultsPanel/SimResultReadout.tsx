import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { formatToPercent } from '@sim/utils/format';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import { useSyncExternalStore } from 'react';

import { useResultFreshness } from '../../hooks/useResultFreshness';
import type { ResultFreshness } from '../../model/result_freshness';
import type { SimResultsManager } from '../../model/results_manager';
import { type ResultMetric, toplineResultMetrics } from '../../model/topline_metrics';
import { formatAverage, formatStdev } from '../ResultMetricList/utils';
import { isRunStage, ResultsPanelStage, type ResultsPanelStore } from './results_panel_store';
import { useRunProgress } from './SimProgress';
import { FRESHNESS_TEXT } from './SimResultSummary';

export interface SimResultReadoutRun {
	metric: 'dps' | 'hps';
	/** Undefined until the first iterations report. */
	value?: number;
	/** Undefined while the presims run. */
	fraction?: number;
}

export interface SimResultReadoutViewProps {
	/** The finished run's primary metric. */
	metric?: ResultMetric;
	freshness?: ResultFreshness;
	/** A run in progress, which the readout shows over any finished one. */
	run?: SimResultReadoutRun;
}

/** Two short lines for the bottom bar: the run in progress, else the finished run's primary metric, else the label of what the bar opens. */
export const SimResultReadoutView = ({ metric, freshness, run }: SimResultReadoutViewProps) => (
	<span className="flex min-w-0 items-center gap-2 text-left font-bold" data-testid="sim-result-readout">
		<Icon name="chevron-up" size="xs" className="shrink-0 text-link" />
		{run ? (
			<span className="flex min-w-0 flex-col tabular-nums" data-testid="sim-result-readout-run">
				<span className="truncate">
					{run.value === undefined ? '—' : run.value.toFixed(2)}
					<span className="text-ui font-normal"> {i18n.t(`sidebar.results.metrics.${run.metric}.label`)}</span>
				</span>
				<span className="text-xs font-normal text-muted">
					{run.fraction === undefined
						? i18n.t('sidebar.results.progress.running')
						: formatToPercent(run.fraction * 100, { maximumFractionDigits: 0 })}
				</span>
			</span>
		) : metric ? (
			<span className="flex min-w-0 flex-col" data-testid="sim-result-readout-metric">
				<span className="truncate">
					{formatAverage(metric, 'list')}
					<span className="text-ui font-normal"> {metric.name}</span>
				</span>
				<span className="flex items-center gap-1 text-xs font-normal text-muted">
					<Icon name="plus-minus" size="xs" />
					{formatStdev(metric, 'list')}
					{freshness && (
						<span
							className={clsx('ml-1 size-2 shrink-0 rounded-full bg-current', FRESHNESS_TEXT[freshness])}
							data-freshness={freshness}
							role="img"
							aria-label={i18n.t(`sidebar.results.freshness.${freshness}`)}
						/>
					)}
				</span>
			</span>
		) : (
			<span className="truncate">{i18n.t('sidebar.character_stats.title')}</span>
		)}
	</span>
);

export interface SimResultReadoutProps {
	panel: ResultsPanelStore;
	results: SimResultsManager;
}

// Mounted with the run, so it starts from the store's progress for that run rather than from the last one's.
const RunReadout = ({ panel, metric }: { panel: ResultsPanelStore; metric: SimResultReadoutRun['metric'] }) => {
	const { progress, fraction } = useRunProgress(panel);
	return <SimResultReadoutView run={{ metric, value: progress?.completedIterations ? progress[metric] : undefined, fraction }} />;
};

export const SimResultReadout = ({ panel, results }: SimResultReadoutProps) => {
	const displayMetrics = useDisplayMetrics(useSim());
	const stage = useSyncExternalStore(panel.subscribe, panel.getStage);
	const { current } = useStoreSubscribe(results.subscribe, results.getData);
	const freshness = useResultFreshness();

	if (isRunStage(stage)) return <RunReadout panel={panel} metric={displayMetrics.damage || !displayMetrics.healing ? 'dps' : 'hps'} />;
	const metric = stage === ResultsPanelStage.Result && current ? toplineResultMetrics(current.simResult, undefined, { displayMetrics })[0] : undefined;
	return <SimResultReadoutView metric={metric} freshness={freshness} />;
};
