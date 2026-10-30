import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { formatDurationSeconds } from '@sim/utils/format';
import { act, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { BulkProgressDialog } from './BulkProgressDialog';

type Listener = (progress: unknown) => void;

const run = vi.hoisted(() => ({ progress: null as unknown, listeners: new Set<Listener>() }));

vi.mock('../../model/run', () => ({
	bulkProgress: () => run.progress,
	subscribeBulkProgress: (_player: unknown, listener: Listener) => {
		run.listeners.add(listener);
		return () => {
			run.listeners.delete(listener);
		};
	},
	cancelBulkBatch: vi.fn(),
}));

const CURRENT = {
	encounter: 'Current encounter',
	encounterSummary: '3:00 ±5s · 1 target',
	settings: 'Current settings',
	talents: 'Current talents',
	rotation: 'Current rotation',
};
const FURY = { ...CURRENT, talents: 'Fury', rotation: 'Fury' };
const LONG_FIGHT = {
	encounter: 'Default - 60',
	encounterSummary: '5:00 ±30s · 3 targets',
	settings: 'Defaults',
	talents: 'Fury',
	rotation: 'Fury - Autocast on use',
};

// One bundle group, so each row is an encounter.
const progressAt = (rows: object[], running: number, progress = 0, seconds: number[] = []) => ({
	stage: 'preparing',
	title: '',
	steps: { rows, groups: rows.map(() => 0), weights: rows.map(() => 10_000), seconds, running, progress },
});

const mount = (rows: object[]) => {
	run.progress = progressAt(rows, 0);
	render(
		<SimHostProvider host={{ player: {} } as never}>
			<BulkProgressDialog />
		</SimHostProvider>,
	);
};

const tick = (progress: object) =>
	act(() => {
		run.progress = progress;
		for (const listener of run.listeners) listener(progress);
	});

const runStep = (rows: object[], running: number, progress?: number) => tick(progressAt(rows, running, progress));

const bodyRows = () => within(screen.getByRole('table')).getAllByRole('row').slice(1);
// The step columns' text; the status cell is read by role.
const cells = (row: HTMLElement) =>
	within(row)
		.getAllByRole('cell')
		.slice(0, -1)
		.map(cell => cell.textContent);
const stepLines = () => bodyRows().map(row => within(row).getByRole('progressbar'));
const timeRemaining = () => screen.queryByTestId('progress-tracker-time-remaining');

describe('BulkProgressDialog', () => {
	beforeEach(() => run.listeners.clear());

	it('lists a Builds batch as a table of encounter with its length and targets, settings, talents and rotation, and marks the running step', () => {
		const rows = [CURRENT, FURY, LONG_FIGHT];
		mount(rows);

		const table = screen.getByRole('table');
		expect(
			within(table)
				.getAllByRole('columnheader')
				.map(header => header.textContent),
		).toEqual([
			i18n.t('bulk_tab.builds.parts.encounter'),
			i18n.t('bulk_tab.builds.parts.settings'),
			i18n.t('bulk_tab.builds.parts.talents'),
			i18n.t('bulk_tab.builds.parts.rotation'),
			i18n.t('bulk_tab.progress.status'),
		]);
		expect(bodyRows().map(cells)).toEqual([
			['Current encounter3:00 ±5s · 1 target', 'Current settings', 'Current talents', 'Current rotation'],
			['Current encounter3:00 ±5s · 1 target', 'Current settings', 'Fury', 'Fury'],
			['Default - 605:00 ±30s · 3 targets', 'Defaults', 'Fury', 'Fury - Autocast on use'],
		]);
		expect(bodyRows().map(row => row.getAttribute('aria-current'))).toEqual(['step', null, null]);

		runStep(rows, 1);

		expect(bodyRows().map(row => row.getAttribute('data-state'))).toEqual(['done', 'running', 'pending']);
		expect(bodyRows().map(row => row.getAttribute('aria-current'))).toEqual([null, 'step', null]);
		expect(within(bodyRows()[1]).getByTestId('progress-step-spinner')).toBeTruthy();
		expect(within(bodyRows()[0]).getByRole('img', { name: i18n.t('bulk_tab.progress.done') })).toBeTruthy();
		expect(within(bodyRows()[1]).getByRole('img', { name: i18n.t('bulk_tab.progress.running') })).toBeTruthy();
		expect(within(bodyRows()[2]).queryByRole('img')).toBeNull();
		expect(within(bodyRows()[2]).getByText(i18n.t('bulk_tab.progress.pending'))).toBeTruthy();
	});

	it('gives each step row its own progress bar: done rows full, the running row as far as its step is, pending rows empty', () => {
		const rows = [CURRENT, FURY, LONG_FIGHT];
		mount(rows);
		expect(stepLines().map(line => line.getAttribute('aria-valuenow'))).toEqual(['0', '0', '0']);

		runStep(rows, 1, 0.4);
		expect(stepLines().map(line => line.getAttribute('aria-valuenow'))).toEqual(['1', '0.4', '0']);
		expect(stepLines().map(line => line.getAttribute('aria-valuetext'))).toEqual(['100%', '40%', '0%']);

		runStep(rows, 2, 0.1);
		expect(stepLines().map(line => line.getAttribute('aria-valuenow'))).toEqual(['1', '1', '0.1']);
	});

	it('names each row’s progress bar by its encounter and talents', () => {
		i18n.addResourceBundle('en', 'translation', english, true, true);
		try {
			mount([CURRENT, FURY, LONG_FIGHT]);
			expect(stepLines().map(line => line.getAttribute('aria-label'))).toEqual([
				'Run progress: Current encounter, Current talents',
				'Run progress: Current encounter, Fury',
				'Run progress: Default - 60, Fury',
			]);
		} finally {
			i18n.removeResourceBundle('en', 'translation');
		}
	});

	it('shows the batch’s time left beside the elapsed time once the first encounter is done', () => {
		vi.useFakeTimers();
		try {
			const rows = [CURRENT, LONG_FIGHT];
			mount(rows);
			const elapsedLine = () => screen.getByTestId('time-elapsed').parentElement!.textContent;
			act(() => vi.advanceTimersByTime(1000));
			expect(elapsedLine()).toBe(`${i18n.t('common.elapsed_time')}: ${formatDurationSeconds(1)}`);

			// The first encounter took 30 seconds, and the second, with the same plan, is a fifth through: 24 s left.
			run.progress = progressAt(rows, 1, 0.2, [30]);
			const timeLeft = (seconds: number) => ` · ${i18n.t('bulk_tab.progress.time_left', { time: formatDurationSeconds(seconds) })}`;
			act(() => vi.advanceTimersByTime(1000));
			expect(elapsedLine()).toBe(`${i18n.t('common.elapsed_time')}: ${formatDurationSeconds(2)}${timeLeft(24)}`);

			// With no newer estimate it counts down on its own, a second a second.
			act(() => vi.advanceTimersByTime(3000));
			expect(elapsedLine()).toBe(`${i18n.t('common.elapsed_time')}: ${formatDurationSeconds(5)}${timeLeft(21)}`);
		} finally {
			vi.useRealTimers();
		}
	});

	it('leaves the bar’s own time left off in a batch of several steps', () => {
		const rows = [CURRENT, LONG_FIGHT];
		mount(rows);
		tick({ ...progressAt(rows, 0), stage: 'sim', current: 3, total: 8, secondsRemaining: 30 });

		expect(timeRemaining()).toBeNull();
	});

	it('counts the iterations done under the bar in a batch of one step, as they tick', () => {
		i18n.addResourceBundle('en', 'translation', english, true, true);
		try {
			mount([CURRENT]);
			const sim = { ...progressAt([CURRENT], 0), stage: 'sim', current: 3, total: 8, secondsRemaining: 30 };
			expect(screen.queryByTestId('results-sim')).toBeNull();

			tick({ ...sim, iterations: { completed: 400, total: 1000 } });
			const message = screen.getByTestId('progress-tracker-modal-message');
			expect(message.hidden).toBe(false);
			expect(message.textContent).toBe('400 / 1000iterations complete');

			tick({ ...sim, iterations: { completed: 650, total: 1000 } });
			expect(screen.getByTestId('results-sim').textContent).toBe('650 / 1000iterations complete');

			tick({ ...progressAt([CURRENT], 0), stage: 'preparing', current: 1, total: 4, secondsRemaining: 30 });
			expect(screen.queryByTestId('results-sim')).toBeNull();
		} finally {
			i18n.removeResourceBundle('en', 'translation');
		}
	});

	it('shows no iterations line in a batch of several steps', () => {
		const rows = [CURRENT, LONG_FIGHT];
		mount(rows);
		tick({ ...progressAt(rows, 0), stage: 'sim', current: 3, total: 8, secondsRemaining: 30, iterations: { completed: 400, total: 1000 } });

		expect(screen.queryByTestId('results-sim')).toBeNull();
		expect(screen.getByTestId('progress-tracker-modal-message').hidden).toBe(true);
	});

	it('shows no step table for a batch of one step, which a Gear-mode batch always is, and the stage’s time left under its bar', () => {
		mount([CURRENT]);
		tick({ ...progressAt([CURRENT], 0), stage: 'sim', current: 3, total: 8, secondsRemaining: 30 });

		expect(screen.getByRole('progressbar')).toBeTruthy();
		expect(screen.queryByRole('table')).toBeNull();
		expect(timeRemaining()?.textContent).toBe(i18n.t('sidebar.results.progress.time_remaining', { time: formatDurationSeconds(30) }));
		expect(screen.getByTestId('progress-tracker-modal-message').hidden).toBe(true);
	});
});
