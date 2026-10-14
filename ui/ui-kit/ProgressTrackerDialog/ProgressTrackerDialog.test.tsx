// What this pins is the split the dialog is built on: the stage is dialog state and everything that
// moves with a worker message is state local to `HandleProgressBar`, so a tick renders that leaf and not the
// `keepMounted` dialog that outlives every run. The measured rate is ~10 progress callbacks a second
// on wasm (sim/core/sim.go:336 throttles each sim to one report per 100 ms, and
// ui/sim/wasm/sim.ts:118 decimates by worker count) and ~2/s on the native host, which is why the
// bar renders per message instead of coalescing across frames.
import i18n from '@i18n/config';
import { formatDurationSeconds } from '@sim/utils/format';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { Profiler, useState } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ElapsedTime } from './ElapsedTime';
import { HandleProgressBar } from './HandleProgressBar';
import { ProgressTrackerBar } from './ProgressTrackerBar';
import { ProgressTrackerDialog } from './ProgressTrackerDialog';
import type { ProgressTrackerHandle, ProgressTrackerState } from './types';

let handle: ProgressTrackerHandle | null = null;
let setStage: (state: ProgressTrackerState) => void = () => {};
let harnessRenders = 0;

const Harness = ({ open = true, onCancel }: { open?: boolean; onCancel?: () => void }) => {
	const [state, setState] = useState<ProgressTrackerState>({ stage: 'initializing' });
	setStage = setState;
	harnessRenders++;

	return (
		<ProgressTrackerDialog
			ref={value => {
				handle = value;
			}}
			open={open}
			title="Calculate Stat Weights"
			className="ep-weights-progress"
			state={state}
			hasProgressBar
			onCancel={onCancel}
		/>
	);
};

const renderDialog = (props: { open?: boolean; onCancel?: () => void } = {}) => {
	const commits = vi.fn();
	const result = render(
		<Profiler id="progress" onRender={commits}>
			<Harness {...props} />
		</Profiler>,
	);
	return { commits, result };
};

const bar = () => screen.getByRole('progressbar');
const barFill = () => screen.getByTestId('progress-tracker-bar-indicator');
const barText = () => screen.queryByTestId('progress-tracker-modal-progress-text');
const barTitle = () => screen.queryByTestId('progress-tracker-modal-progress-title');

describe('ProgressTrackerDialog', () => {
	beforeEach(() => {
		handle = null;
	});

	it('writes the caption, the bar and its text from setProgress', () => {
		renderDialog();
		act(() => handle!.setProgress({ title: '3 / 12 simulations complete', current: 30, total: 120 }));

		expect(barTitle()!.textContent).toBe('3 / 12 simulations complete');
		expect(barFill().style.width).toBe('25%');
		expect(bar().getAttribute('aria-valuenow')).toBe('0.25');
		expect(bar().getAttribute('aria-valuemax')).toBe('1');
		expect(bar().getAttribute('aria-valuemin')).toBe('0');
		expect(bar().getAttribute('aria-valuetext')).toBe('25%');
		expect(bar().hasAttribute('data-progressing')).toBe(true);
		expect(barText()!.textContent).toBe('30/120');
	});

	it('hides the bar and the caption while they have nothing to say', () => {
		renderDialog();
		act(() => handle!.setProgress({ title: '1 / 2', current: 1, total: 2 }));
		act(() => handle!.setProgress({}));

		expect(barTitle()).toBeNull();
		expect(barText()).toBeNull();
		// The track stays, empty: Base UI leaves the fill unsized and marks it unmeasured, which is what
		// the stylesheet hides on.
		expect(bar().hasAttribute('data-indeterminate')).toBe(true);
		expect(bar().hasAttribute('aria-valuenow')).toBe(false);
		expect(barFill().hasAttribute('data-indeterminate')).toBe(true);
		expect(barFill().style.width).toBe('');
	});

	// Sim progress bypasses the store. The bound is one commit of this leaf per batch of ticks — the
	// dialog around it never re-renders.
	it('renders once for a hundred progress ticks in one batch, and shows the last of them', () => {
		const { commits } = renderDialog();
		const atMount = commits.mock.calls.length;
		const dialogRenders = harnessRenders;

		act(() => {
			for (let i = 1; i <= 100; i++) handle!.setProgress({ title: `${i} / 100`, current: i, total: 100 });
		});

		expect(commits.mock.calls.length).toBe(atMount + 1);
		// The one commit is the bar alone: the component that owns the dialog never reconciled.
		expect(harnessRenders).toBe(dialogRenders);
		expect(barText()!.textContent).toBe('100/100');
		expect(barFill().style.width).toBe('100%');
		expect(bar().hasAttribute('data-complete')).toBe(true);
	});

	it('writes the time left on a line of its own, held empty until there is an estimate, and adds no line for a run that sends none', () => {
		renderDialog();
		const remaining = () => screen.queryByTestId('progress-tracker-time-remaining');

		act(() => handle!.setProgress({ current: 1, total: 4 }));
		expect(remaining()).toBeNull();

		act(() => handle!.setProgress({ current: 1, total: 4, secondsRemaining: null }));
		expect(remaining()!.textContent).toBe(' ');

		const t = vi.spyOn(i18n, 't');
		act(() => handle!.setProgress({ current: 2, total: 4, secondsRemaining: 95 }));
		expect(t).toHaveBeenCalledWith('sidebar.results.progress.time_remaining', { time: formatDurationSeconds(95) });
		expect(remaining()!.textContent).toBe(t.mock.results.at(-1)!.value);
		t.mockRestore();
	});

	it('keeps what is pushed the moment the handle attaches', () => {
		const pushOnAttach = (value: ProgressTrackerHandle | null) => value?.setProgress({ current: 2, total: 8 });
		render(<ProgressTrackerDialog ref={pushOnAttach} open title="Bulk Sim" state={{ stage: 'running' }} hasProgressBar />);

		expect(barText()!.textContent).toBe('2/8');
		expect(barFill().style.width).toBe('25%');
	});

	// A consumer that keeps the dialog mounted between runs would otherwise open the next one showing
	// the last one's numbers.
	it('clears the bar when the next run opens it', () => {
		const view = render(<Harness />);
		act(() => handle!.setProgress({ title: '5 / 5 simulations complete', current: 5, total: 5 }));
		expect(barText()!.textContent).toBe('5/5');

		view.rerender(<Harness open={false} />);
		view.rerender(<Harness />);

		expect(barText()).toBeNull();
		expect(barTitle()).toBeNull();
		expect(bar().hasAttribute('data-indeterminate')).toBe(true);
	});

	it('renders once for a stage transition', () => {
		const { commits } = renderDialog();
		const atMount = commits.mock.calls.length;

		act(() => setStage({ stage: 'complete', message: 'done' }));

		expect(commits.mock.calls.length).toBe(atMount + 1);
		expect(screen.getByTestId('progress-tracker-modal-content').getAttribute('data-stage')).toBe('complete');
		const message = screen.getByTestId('progress-tracker-modal-message');
		expect(message.textContent).toBe('done');
		expect(message.hidden).toBe(false);
	});

	it('hides the message while there is none', () => {
		renderDialog();
		expect(screen.getByTestId('progress-tracker-modal-message').hidden).toBe(true);
	});

	it('has no close button, and cancels through its own button', () => {
		const onCancel = vi.fn();
		renderDialog({ onCancel });

		expect(screen.queryByRole('button', { name: 'common.close' })).toBeNull();

		const cancel = screen.getByTestId('progress-tracker-modal-cancel-btn');
		expect(cancel.getAttribute('type')).toBe('button');
		fireEvent.click(cancel);
		expect(onCancel).toHaveBeenCalledTimes(1);
	});

	it('renders no cancel button when there is nothing to cancel', () => {
		renderDialog();
		expect(screen.queryByTestId('progress-tracker-modal-cancel-btn')).toBeNull();
	});

	it('names the dialog from its title, and keeps it mounted but hidden while closed', () => {
		renderDialog();
		expect(screen.getByRole('dialog', { name: 'Calculate Stat Weights' })).toBeTruthy();

		screen.getByRole('dialog').remove();
		renderDialog({ open: false });
		const popup = screen.getByRole('dialog', { hidden: true });
		expect(popup.getAttribute('data-size')).toBe('md');
		expect(popup.hasAttribute('hidden')).toBe(true);
	});

	it('restarts the elapsed readout when the next run opens it', () => {
		vi.useFakeTimers();
		const view = render(<Harness />);
		act(() => {
			vi.advanceTimersByTime(3200);
		});
		expect(screen.getByTestId('time-elapsed').textContent).toBe('3s');

		view.rerender(<Harness open={false} />);
		act(() => {
			vi.advanceTimersByTime(10_000);
		});
		view.rerender(<Harness />);

		expect(screen.getByTestId('time-elapsed').textContent).toBe('0s');
		vi.useRealTimers();
	});
});

describe('ProgressTrackerBar', () => {
	it('fills to the fraction it is given, with its caption and readout', () => {
		render(<ProgressTrackerBar running progress={0.5} title="Batch 2 of 4" readout="2/4" />);

		expect(barFill().style.width).toBe('50%');
		expect(bar().getAttribute('aria-valuetext')).toBe('50%');
		expect(barTitle()!.textContent).toBe('Batch 2 of 4');
		expect(barText()!.textContent).toBe('2/4');
	});

	it('marks the bar unmeasured while it has no fraction', () => {
		render(<ProgressTrackerBar running title="Preparing" />);

		expect(bar().hasAttribute('data-indeterminate')).toBe(true);
		expect(barFill().style.width).toBe('');
		expect(barTitle()!.textContent).toBe('Preparing');
		expect(barText()).toBeNull();
	});

	it('shows nothing measured while it is not running', () => {
		render(<ProgressTrackerBar running={false} progress={0.5} title="Batch 2 of 4" readout="2/4" />);

		expect(bar().hasAttribute('data-indeterminate')).toBe(true);
		expect(barTitle()).toBeNull();
		expect(barText()).toBeNull();
	});

	it('takes its root test id from testId, progress-tracker-bar by default', () => {
		const { unmount } = render(<ProgressTrackerBar running />);
		expect(screen.getByTestId('progress-tracker-bar').contains(barFill())).toBe(true);
		unmount();

		render(<ProgressTrackerBar running testId="batch-progress" />);
		expect(screen.getByTestId('batch-progress').contains(barFill())).toBe(true);
	});
});

describe('HandleProgressBar', () => {
	it('hands its testId to the bar it renders', () => {
		render(<HandleProgressBar running testId="reforge-progress" />);
		expect(screen.getByTestId('reforge-progress').contains(barFill())).toBe(true);
		expect(screen.queryByTestId('progress-tracker-bar')).toBeNull();
	});
});

describe('ElapsedTime', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	const text = () => screen.getByTestId('time-elapsed').textContent;

	it('ticks the readout without rendering', () => {
		const commits = vi.fn();
		render(
			<Profiler id="elapsed" onRender={commits}>
				<ElapsedTime running startedAt={Date.now()} />
			</Profiler>,
		);
		const atMount = commits.mock.calls.length;
		expect(text()).toBe('0s');

		act(() => {
			vi.advanceTimersByTime(2400);
		});

		expect(text()).toBe('2s');
		expect(commits.mock.calls.length).toBe(atMount);
	});

	// `keepMounted` is what makes this a leak rather than a tidy-up: without the cleanup the interval
	// outlives every run and keeps writing for the life of the page.
	it('stops its interval when the dialog closes and when it unmounts', () => {
		const firstRun = Date.now();
		const { rerender, unmount } = render(<ElapsedTime running startedAt={firstRun} />);
		act(() => {
			vi.advanceTimersByTime(1400);
		});
		expect(text()).toBe('1s');

		rerender(<ElapsedTime running={false} startedAt={firstRun} />);
		act(() => {
			vi.advanceTimersByTime(5000);
		});
		expect(text()).toBe('1s');
		expect(vi.getTimerCount()).toBe(0);

		rerender(<ElapsedTime running startedAt={Date.now()} />);
		expect(text()).toBe('0s');
		act(() => {
			vi.advanceTimersByTime(1400);
		});
		expect(text()).toBe('1s');
		unmount();
		expect(vi.getTimerCount()).toBe(0);
	});

	it('runs no timer while the dialog is closed', () => {
		render(<ElapsedTime running={false} startedAt={Date.now()} />);
		act(() => {
			vi.advanceTimersByTime(5000);
		});
		expect(text()).toBe('0s');
		expect(vi.getTimerCount()).toBe(0);
	});

	it('takes its test id from testId, time-elapsed by default', () => {
		render(<ElapsedTime running={false} startedAt={Date.now()} testId="bulk-elapsed" />);
		expect(screen.getByTestId('bulk-elapsed').textContent).toBe('0s');
		expect(screen.queryByTestId('time-elapsed')).toBeNull();
	});

	it('counts from the start it is given, not from when it mounted', () => {
		render(<ElapsedTime running startedAt={Date.now() - 65_000} />);
		expect(text()).toBe('1m 05s');

		act(() => {
			vi.advanceTimersByTime(1000);
		});
		expect(text()).toBe('1m 06s');
	});
});
