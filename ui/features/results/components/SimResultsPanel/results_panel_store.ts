import type { ResultsPanelHandle } from '@features/results/model/results_panel_handle';
import type { ProgressMetrics } from '@generated/proto/api';
import { timeRemainingAtRate } from '@sim/utils/math';
import { flushSync } from 'react-dom';

export enum ResultsPanelStage {
	Idle = 'idle',
	Pending = 'pending',
	Running = 'running',
	Result = 'result',
}

/** Stage and the Stop button are React state; the progress numbers are DOM writes through `onProgress`, and `latestProgress` sits outside the snapshot so the block tick one mounts can read the values that mounted it. `notify` is `flushSync` because the run action reads the panel back in the click's own task. */
export class ResultsPanelStore implements ResultsPanelHandle {
	latestProgress: ProgressMetrics | null = null;

	private stage: ResultsPanelStage = ResultsPanelStage.Idle;
	private abortHandler: ((event: MouseEvent) => void) | null = null;
	private buttonsVisible = false;
	private readonly listeners = new Set<() => void>();
	private readonly progressListeners = new Set<(progress: ProgressMetrics) => void>();
	private estimateFrom: { at: number; completed: number } | null = null;

	readonly subscribe = (listener: () => void): (() => void) => {
		this.listeners.add(listener);
		return () => {
			this.listeners.delete(listener);
		};
	};

	readonly getStage = (): ResultsPanelStage => this.stage;

	readonly getAbortHandler = (): ((event: MouseEvent) => void) | null => this.abortHandler;

	readonly getButtonsVisible = (): boolean => this.buttonsVisible;

	onProgress(listener: (progress: ProgressMetrics) => void): () => void {
		this.progressListeners.add(listener);
		return () => {
			this.progressListeners.delete(listener);
		};
	}

	setPending() {
		this.latestProgress = null;
		this.estimateFrom = null;
		this.setStage(ResultsPanelStage.Pending);
	}

	/** Measured from the first tick after the presim, so the presim's time does not count against the iterations. */
	secondsRemaining(progress: ProgressMetrics): number | undefined {
		const from = this.estimateFrom;
		if (!from || progress.presimRunning) return undefined;
		return timeRemainingAtRate((Date.now() - from.at) / 1000, progress.completedIterations - from.completed, progress.totalIterations - from.completed);
	}

	setProgress(progress: ProgressMetrics) {
		this.latestProgress = progress;
		if (!progress.presimRunning && !this.estimateFrom) this.estimateFrom = { at: Date.now(), completed: progress.completedIterations };
		// A block that mounts on this stage change reads `latestProgress` instead; one already mounted while pending hears it here.
		if (this.stage !== ResultsPanelStage.Running) this.setStage(ResultsPanelStage.Running);
		this.progressListeners.forEach(listener => listener(progress));
	}

	showResult() {
		this.setStage(ResultsPanelStage.Result);
	}

	hideAll() {
		this.buttonsVisible = false;
		this.setStage(ResultsPanelStage.Idle);
	}

	addAbortButton(abortClicked: (event: MouseEvent) => void) {
		this.abortHandler = abortClicked;
		this.buttonsVisible = true;
		this.notify();
	}

	removeAbortButton() {
		this.abortHandler = null;
		this.buttonsVisible = false;
		this.notify();
	}

	private setStage(stage: ResultsPanelStage) {
		this.stage = stage;
		this.notify();
	}

	private notify() {
		flushSync(() => this.listeners.forEach(listener => listener()));
	}
}
