import { Progress } from '@base-ui/react/progress';
import i18n from '@i18n/config';
import { formatDurationSeconds } from '@sim/utils/format';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';

export interface ProgressTrackerBarProps {
	/** While `false` the bar shows nothing measured, whatever else it is given. */
	running: boolean;
	/** Fraction done, 0 to 1. Omit while nothing is measured: the track stays and the fill hides. */
	progress?: number;
	title?: string;
	readout?: string;
	/** The time left, on a line under the bar; null holds that line empty until there is an estimate. Omit for no line. */
	secondsRemaining?: number | null;
	className?: ClassValue;
	testId?: string;
}

export const ProgressTrackerBar = ({
	running,
	progress,
	title,
	readout,
	secondsRemaining,
	className,
	testId = 'progress-tracker-bar',
}: ProgressTrackerBarProps) => (
	<div className={clsx('ui-progress', className)} data-testid={testId}>
		{running && title && (
			<div className="mb-2" data-testid="progress-tracker-modal-progress-title">
				{title}
			</div>
		)}
		<Progress.Root value={running ? (progress ?? null) : null} max={1}>
			<Progress.Track className="ui-progress-track">
				<Progress.Indicator className="ui-progress-indicator" data-testid="progress-tracker-bar-indicator" />
			</Progress.Track>
		</Progress.Root>
		{running && readout && (
			<div className="text-right text-xs text-gray-400 tabular-nums" data-testid="progress-tracker-modal-progress-text">
				{readout}
			</div>
		)}
		{running && secondsRemaining !== undefined && (
			<div className="mt-2 tabular-nums" data-testid="progress-tracker-time-remaining">
				{secondsRemaining === null ? ' ' : i18n.t('sidebar.results.progress.time_remaining', { time: formatDurationSeconds(secondsRemaining) })}
			</div>
		)}
	</div>
);
