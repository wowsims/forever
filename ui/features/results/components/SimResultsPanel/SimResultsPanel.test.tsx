// The two halves this panel is built on, and one race that only a unit test can reach.
//
// Progress renders once per tick from the store's progress listener. A browser-level check waits for
// text to appear, so it passes on tick two and can never see that tick one's numbers were dropped —
// that gap is what the "same commit" case below covers.
import { ProgressMetrics } from '@generated/proto/api';
import { formatToNumber } from '@sim/utils/format';
import { renderWithHost } from '@testing/react';
import { mockSubscriptions } from '@testing/sim/subscriptions';
import { act } from '@testing-library/react';
import { Profiler } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultsManager } from '../../model/results_manager';
import { ResultsPanelStore } from './results_panel_store';
import { SimResultsPanel } from './SimResultsPanel';

vi.mock('@sim/state/subscriptions', async () => mockSubscriptions());

// The summary needs a whole SimResult; SimResultSummary.test.tsx is where it is asserted. What is
// under test here is that the content zone renders it and nothing else.
vi.mock('./SimResultSummary', () => ({ SimResultSummary: () => <div data-testid="sim-result-summary-root" /> }));

const progress = (dps: number, hps: number, completed: number, total: number, presimRunning = false) =>
	ProgressMetrics.create({ dps, hps, completedIterations: completed, totalIterations: total, presimRunning });

// The sim reports ready a microtask after mount, and every case here is a loaded sim, so the mount
// is not finished until that has landed.
const mount = async (panel: ResultsPanelStore, disabled = false, results: SimResultsManager | null = null, simDisabled = disabled) => {
	const commits = vi.fn();
	const view = renderWithHost(
		<Profiler id="panel" onRender={commits}>
			<SimResultsPanel panel={panel} results={results} />
		</Profiler>,
		{
			disabled,
			simDisabled,
			sim: {
				waitForInit: () => Promise.resolve(),
				getShowDamageMetrics: () => true,
				getShowThreatMetrics: () => true,
				getShowHealingMetrics: () => true,
			},
		},
	);
	await act(async () => {});
	return { commits, view };
};

const zone = (view: ReturnType<typeof renderWithHost>, selector: string) => view.container.querySelector<HTMLElement>(selector)!;
const zones = (view: ReturnType<typeof renderWithHost>) => ({
	pending: !!view.container.querySelector('[data-testid="results-sim"], [data-testid="loader"]'),
	content: !!view.container.querySelector('[data-testid="sim-result-summary-root"]'),
});

let panel: ResultsPanelStore;

// The test i18n stub echoes the key, so these assertions pin the key rather than the copy.
const READOUT = '[data-testid="results-sim-progress"] [data-testid="progress-tracker-modal-progress-text"]';

describe('SimResultsPanel', () => {
	beforeEach(() => {
		panel = new ResultsPanelStore();
	});

	it('renders nothing before a run starts', async () => {
		const { view } = await mount(panel);

		expect(view.container.querySelector('[data-testid="results-viewer"]')).toBeNull();
		expect(zones(view)).toEqual({ pending: false, content: false });
	});

	it('follows the visibility table through the handle', async () => {
		const { view } = await mount(panel, false, {} as SimResultsManager);

		act(() => panel.setPending());
		expect(zones(view)).toEqual({ pending: true, content: false });

		act(() => panel.setProgress(progress(1, 2, 3, 4)));
		expect(zones(view)).toEqual({ pending: true, content: false });
		expect(view.container.querySelector('[data-testid="results-sim"]')).not.toBeNull();
		expect(view.container.querySelector('[data-testid="loader"]')).toBeNull();

		act(() => panel.showResult());
		expect(zones(view)).toEqual({ pending: false, content: true });

		act(() => panel.hideAll());
		expect(zones(view)).toEqual({ pending: false, content: false });
	});

	it('renders the finished run into the content zone once the stage says so, and leaves it out without a manager', async () => {
		const { view } = await mount(panel);
		act(() => panel.showResult());
		expect(view.container.querySelector('[data-testid="sim-result-summary-root"]')).toBeNull();

		const { view: withResults } = await mount(panel, false, {} as SimResultsManager);
		expect(withResults.container.querySelector('[data-testid="sim-result-summary-root"]')).not.toBeNull();
	});

	it('puts the first tick on screen in the same commit that mounts the block', async () => {
		const { view } = await mount(panel);
		act(() => panel.setPending());

		// One call, carrying both the stage change and the first numbers — the shape every transport
		// delivers. Mutation check: drop `SimProgress`'s `latestProgress` read and this reads empty.
		act(() => panel.setProgress(progress(1234.5678, 22.5, 10, 1000)));

		expect(zone(view, '[data-testid="results-sim-dps"] [data-testid="topline-result-avg"]').textContent).toBe('1234.57');
		expect(zone(view, '[data-testid="results-sim-hps"] [data-testid="topline-result-avg"]').textContent).toBe('22.50');
		expect(zone(view, READOUT).textContent).toBe(`${formatToNumber(10)} / ${formatToNumber(1000)}`);
		expect(zone(view, '[data-testid="results-sim-progress"] [role="progressbar"]').getAttribute('aria-valuenow')).toBe('0.01');
	});

	it('writes later ticks with one commit each, the progress bar', async () => {
		const { view, commits } = await mount(panel);
		act(() => panel.setPending());
		act(() => panel.setProgress(progress(1, 1, 1, 100)));
		const afterMount = commits.mock.calls.length;

		for (let i = 2; i <= 100; i++) act(() => panel.setProgress(progress(i, i * 2, i, 100)));

		expect(zone(view, '[data-testid="results-sim-dps"] [data-testid="topline-result-avg"]').textContent).toBe('100.00');
		expect(zone(view, '[data-testid="results-sim-hps"] [data-testid="topline-result-avg"]').textContent).toBe('200.00');
		expect(zone(view, '[data-testid="results-sim-progress"] [role="progressbar"]').getAttribute('aria-valuenow')).toBe('1');
		expect(commits.mock.calls.length).toBe(afterMount + 99);

		// And a stage change is still exactly one commit.
		act(() => panel.showResult());
		expect(commits.mock.calls.length).toBe(afterMount + 100);
	});

	it('shows the presim caption instead of the counter', async () => {
		const { view } = await mount(panel);
		act(() => panel.setPending());
		act(() => panel.setProgress(progress(0, 0, 0, 0, true)));

		expect(zone(view, READOUT).textContent).toBe('sidebar.results.progress.presim_running');
		expect(zone(view, '[data-testid="results-sim-progress"] [role="progressbar"]').hasAttribute('aria-valuenow')).toBe(false);
	});

	it('shows the time remaining once a second tick gives it a pace', async () => {
		const { view } = await mount(panel);
		act(() => panel.setPending());
		act(() => panel.setProgress(progress(1, 1, 10, 100)));
		expect(view.container.querySelector('[data-testid="results-sim-time-remaining"]')).toBeNull();

		act(() => panel.setProgress(progress(1, 1, 20, 100)));
		expect(zone(view, '[data-testid="results-sim-time-remaining"]').textContent).toBe('sidebar.results.progress.time_remaining');
	});

	it('commits the stage before the caller returns, so a same-task read sees it', async () => {
		const { view } = await mount(panel);
		// The read happens inside the same batch as the write, like the run action's own click
		// handler reading the DOM back before its first await. Without `flushSync` in the store the
		// commit would wait for the end of the batch and this would read the stage it replaced.
		let pendingOnReturn = false;
		act(() => {
			panel.setPending();
			pendingOnReturn = zones(view).pending;
		});
		expect(pendingOnReturn).toBe(true);
	});

	it('renders the unlaunched notice last, and only for a disabled spec', async () => {
		const launched = await mount(panel, false);
		expect(launched.view.container.querySelector('[data-testid="sim-ui-unlaunched-container"]')).toBeNull();
		launched.view.unmount();

		const { view } = await mount(new ResultsPanelStore(), true);
		const viewer = zone(view, '[data-testid="results-viewer"]');
		expect(viewer.lastElementChild!.getAttribute('data-testid')).toBe('sim-ui-unlaunched-container');
		// One paragraph: there is no external healing sim to point at.
		expect(viewer.querySelectorAll('[data-testid="sim-ui-unlaunched-container"] p').length).toBe(1);
		expect(viewer.querySelector('a[href*="questionablyepic"]')).toBeNull();
	});

	it('renders the gear planner notice for a sim that never runs, and the unlaunched notice instead when the sim is disabled', async () => {
		const launched = await mount(panel, false);
		expect(launched.view.container.querySelector('[data-testid="sim-ui-gear-planner-container"]')).toBeNull();
		launched.view.unmount();

		const planner = await mount(new ResultsPanelStore(), false, null, true);
		const viewer = zone(planner.view, '[data-testid="results-viewer"]');
		expect(viewer.lastElementChild!.getAttribute('data-testid')).toBe('sim-ui-gear-planner-container');
		// One paragraph for every gear planner: there is no external healing sim to point at.
		expect(viewer.querySelectorAll('[data-testid="sim-ui-gear-planner-container"] p').length).toBe(1);
		expect(viewer.querySelector('a[href*="questionablyepic"]')).toBeNull();
		expect(viewer.querySelector('[data-testid="sim-ui-unlaunched-container"]')).toBeNull();
		planner.view.unmount();

		// Unlaunched outside dev mode: both flags are set and only the unlaunched notice shows.
		const { view } = await mount(new ResultsPanelStore(), true, null, true);
		expect(view.container.querySelector('[data-testid="sim-ui-unlaunched-container"]')).not.toBeNull();
		expect(view.container.querySelector('[data-testid="sim-ui-gear-planner-container"]')).toBeNull();
	});

	it('drops its subscriptions on unmount', async () => {
		const { view } = await mount(panel);
		view.unmount();

		// A late tick from a run that outlived the panel must not throw.
		expect(() => panel.setProgress(progress(1, 1, 1, 1))).not.toThrow();
	});
});
