import type { ReactNode } from 'react';

/** The discrete half: a handful of transitions per run, so React state. */
export interface ProgressTrackerState {
	stage: 'initializing' | 'complete' | 'error' | string;
	message?: ReactNode;
}

/** The continuous half: pushed through `ProgressTrackerHandle` so a tick renders `HandleProgressBar` alone, never the dialog. */
export interface ProgressTrackerProgress {
	title?: string;
	current?: number;
	total?: number;
	/** Shows the time left on a line of its own; null keeps the line empty until there is an estimate, so the dialog does not grow. */
	secondsRemaining?: number | null;
}

export interface ProgressTrackerHandle {
	setProgress: (progress: ProgressTrackerProgress) => void;
}
