import type { ProgressMetrics } from '@generated/proto/api';
import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { ProgressTrackerBar } from '@ui-kit/ProgressTrackerDialog';
import { Spinner } from '@ui-kit/Spinner';
import clsx, { ClassValue } from 'clsx';
import { useLayoutEffect, useState } from 'react';

import { formatDurationSeconds, formatToNumber } from '../../../../sim/utils/format';
import type { ResultsPanelStore } from './results_panel_store';

export interface SimProgressProps {
	panel: ResultsPanelStore;
	className?: ClassValue;
}

const fractionDone = (progress: ProgressMetrics | null): number | undefined =>
	progress && !progress.presimRunning && progress.totalIterations > 0 ? progress.completedIterations / progress.totalIterations : undefined;

const readoutOf = (progress: ProgressMetrics | null): string => {
	if (progress?.presimRunning) return i18n.t('sidebar.results.progress.presim_running');
	if (!progress?.completedIterations) return i18n.t('sidebar.results.progress.starting');
	return `${formatToNumber(progress.completedIterations)} / ${formatToNumber(progress.totalIterations)}`;
};

interface ProgressMetricProps {
	metric: 'dps' | 'hps';
	value: number | undefined;
	compact: boolean;
}

const ProgressMetric = ({ metric, value, compact }: ProgressMetricProps) => (
	<div data-testid={`results-sim-${metric}`} className="text-left font-bold" data-size={compact ? 'sm' : 'md'}>
		<span data-testid="topline-result-avg" className={compact ? 'text-base' : 'text-2xl leading-none'}>
			{value === undefined ? '—' : value.toFixed(2)}
		</span>
		<span className="text-ui font-normal"> {i18n.t(`sidebar.results.metrics.${metric}.label`)}</span>
	</div>
);

/** Laid out like the finished run's summary, so the sidebar barely moves when the result replaces it. */
export const SimProgress = ({ panel, className }: SimProgressProps) => {
	const { damage: showDamage, healing: showHealing } = useDisplayMetrics(useSim());
	const [{ progress, secondsRemaining }, setTick] = useState(() => ({ progress: panel.latestProgress, secondsRemaining: undefined as number | undefined }));
	useLayoutEffect(() => panel.onProgress(next => setTick({ progress: next, secondsRemaining: panel.secondsRemaining(next) })), [panel]);
	const fraction = fractionDone(progress);

	return (
		<div data-testid="results-sim" className={clsx('flex w-full flex-col gap-2', className)}>
			<div className="flex items-center justify-between gap-2 text-sm" data-testid="results-progress-header">
				<span>{i18n.t('sidebar.results.progress.running')}</span>
				{!fraction ? (
					<Spinner size="xs" />
				) : (
					secondsRemaining !== undefined && (
						<span className="tabular-nums" data-testid="results-sim-time-remaining">
							{i18n.t('sidebar.results.progress.time_remaining', {
								time: formatDurationSeconds(secondsRemaining, { separatorStyle: 'colon', minimumUnit: 'minutes' }).padStart(5, '0'),
							})}
						</span>
					)
				)}
			</div>
			{showDamage && <ProgressMetric metric="dps" value={progress?.dps} compact={false} />}
			{showHealing && <ProgressMetric metric="hps" value={progress?.hps} compact={showDamage} />}
			<ProgressTrackerBar running progress={fraction} readout={readoutOf(progress)} className="w-full max-w-full" testId="results-sim-progress" />
		</div>
	);
};
