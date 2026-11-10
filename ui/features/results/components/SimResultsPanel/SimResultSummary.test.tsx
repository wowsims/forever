import { Raid as RaidProto } from '@generated/proto/api';
import { Encounter as EncounterProto } from '@generated/proto/common';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { mockSubscriptions } from '@testing/sim/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { ReferenceDiffs } from '../../model/reference_diffs';
import { SimResultsManager } from '../../model/results_manager';
import type { ResultMetric } from '../../model/topline_metrics';
import { SimResultSummary, SimResultSummaryView, type SimResultSummaryViewProps } from './SimResultSummary';

vi.mock('@sim/state/subscriptions', async () => mockSubscriptions());

const dist = (avg: number, stdev = 0) => ({ avg, stdev });

// Both accessor sets the reference bar reads: the raid's for dps and hps, the first player's for the
// other five.
const result = (dps: number, iterations = 1000) =>
	({
		iterations,
		raidMetrics: { dps: dist(dps, 10), hps: dist(20, 2) },
		getFirstPlayer: () => player,
		getRaidIndexedPlayers: () => [player],
		getRaidIndexedActionMetrics: () => [],
		getTargets: () => [],
		request: { raid: RaidProto.create(), encounter: EncounterProto.create() },
		result: { avgIterationDuration: 300 },
		encounterMetrics: { durationSeconds: 300 },
		toProto: () => ({}),
	}) as never;

const player = {
	spec: { specID: -1 } as never,
	unitIndex: 0,
	getTargetIndex: () => null,
	dps: dist(1000, 10),
	tps: dist(500, 5),
	dtps: dist(50, 1),
	tmi: dist(2, 0.5),
	hps: dist(20, 2),
	tto: dist(30, 3),
	chanceOfDeath: dist(1, 0.1),
	secondsOomAvg: 0,
	actionsPerMinute: () => 42,
};

let store: ReturnType<typeof createSimStore>;

const sim = () =>
	({
		store,
		raid: { toProto: () => RaidProto.create(), fromProto: vi.fn() },
		encounter: { toProto: () => EncounterProto.create(), fromProto: vi.fn() },
	}) as never;

const mount = (results: SimResultsManager) =>
	renderWithHost(<SimResultSummary results={results} />, {
		config: { cssScheme: 'test' },
		sim: { store, getShowDamageMetrics: () => true, getShowThreatMetrics: () => true, getShowHealingMetrics: () => true },
	});

const diffs = (container: HTMLElement) => [...container.querySelectorAll('[data-testid="results-reference"] > [data-testid="results-reference-diff"]')];

let results: SimResultsManager;

const strip = (container: HTMLElement) => container.querySelector('[data-testid="results-sim-reference"]')!;
const metricIds = (container: HTMLElement) => [...container.querySelectorAll('[data-testid="results-metric"]')].map(row => row.getAttribute('data-metric'));

describe('SimResultSummary', () => {
	beforeEach(() => {
		store = createSimStore();
		results = new SimResultsManager(sim());
	});

	it('renders nothing before the first run', () => {
		expect([...mount(results).container.children].map(child => child.getAttribute('data-testid'))).toEqual(['test-portal']);
	});

	it('renders the stacked metric list and the reference bar once a run lands', () => {
		act(() => results.setSimResult(result(1000)));
		const { container } = mount(results);

		expect(metricIds(container)).toEqual(['dps', 'tps', 'apm', 'dtps', 'tmi', 'cod', 'tto', 'hps']);
		expect(strip(container).hasAttribute('data-has-reference')).toBe(false);
	});

	it('fills the deltas in and flags the bar the moment the reference is set', () => {
		act(() => results.setSimResult(result(1000)));
		const { container } = mount(results);

		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-set-reference"]')!));
		expect(strip(container).hasAttribute('data-has-reference')).toBe(true);

		// Same run on both sides: every rendered metric gets a delta, all of them zero and none coloured.
		expect(diffs(container)).toHaveLength(7);
		expect(new Set(diffs(container).map(diff => diff.textContent))).toEqual(new Set(['+0.00 (0.00%)']));
		expect(new Set(diffs(container).map(diff => diff.getAttribute('data-sign')))).toEqual(new Set([null]));
	});

	it('colours a significant gain positive and carries the significance on the anchor', () => {
		act(() => results.setSimResult(result(1000)));
		const { container } = mount(results);
		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-set-reference"]')!));
		act(() => results.setSimResult(result(1100)));

		const dps = container.querySelector('[data-metric="dps"] [data-testid="results-reference-diff"]')!;
		expect(dps.textContent).toBe('+100.00 (10.00%)');
		expect(dps.getAttribute('data-sign')).toBe('positive');
		expect(dps.getAttribute('data-tooltip-content')).toBe('common.significance.significant');
	});

	it('drops the reference again on delete', () => {
		act(() => results.setSimResult(result(1000)));
		const { container } = mount(results);
		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-set-reference"]')!));
		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-reference-delete"]')!));

		expect(strip(container).hasAttribute('data-has-reference')).toBe(false);
		expect(diffs(container)).toEqual([]);
	});

	it('swaps the two runs, so the delta changes sign', () => {
		act(() => results.setSimResult(result(1000)));
		const { container } = mount(results);
		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-set-reference"]')!));
		act(() => results.setSimResult(result(1100)));

		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-reference-swap"]')!));

		// In percent of the reference, now the 1100 run.
		const dps = container.querySelector('[data-metric="dps"] [data-testid="results-reference-diff"]')!;
		expect(dps.textContent).toBe('-100.00 (9.09%)');
		expect(dps.getAttribute('data-sign')).toBe('negative');
	});

	it('titles a swapped-in run as the reference run, and the last run again after swapping back or a new run', () => {
		const title = (container: HTMLElement) => container.querySelector('[data-testid="results-run-title"]')!.textContent;
		act(() => results.setSimResult(result(1000)));
		const { container } = mount(results);
		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-set-reference"]')!));
		act(() => results.setSimResult(result(1100)));
		expect(title(container)).toBe('sidebar.results.last_run');

		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-reference-swap"]')!));
		expect(title(container)).toBe('sidebar.results.reference_run');

		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-reference-swap"]')!));
		expect(title(container)).toBe('sidebar.results.last_run');

		act(() => fireEvent.click(container.querySelector('[data-testid="results-sim-reference-swap"]')!));
		act(() => results.setSimResult(result(1200)));
		expect(title(container)).toBe('sidebar.results.last_run');
	});
});

const METRICS: Array<ResultMetric> = [
	{ metric: 'dps', name: 'DPS', average: 1000, stdev: 10 },
	{ metric: 'tps', name: 'TPS', average: 800, stdev: 8 },
];

const DIFFS: ReferenceDiffs = {
	dps: { text: '+100.00 (10.00%)', tone: 'positive', significance: 'significant' },
	tps: { text: '+0.00 (0.00%)', tone: null, significance: 'not significant' },
};

const mountView = (props: Partial<SimResultSummaryViewProps> = {}) => {
	const handlers = { onSetReference: vi.fn(), onSwapReference: vi.fn(), onClearReference: vi.fn() };
	const view = render(<SimResultSummaryView metrics={METRICS} {...handlers} {...props} />);
	return { ...handlers, ...view };
};

describe('SimResultSummaryView', () => {
	it('lists every metric it is given and offers only the save button without a reference', () => {
		const { container } = mountView();

		expect(metricIds(container)).toEqual(['dps', 'tps']);
		expect(strip(container).hasAttribute('data-has-reference')).toBe(false);
		expect(screen.getByTestId('results-sim-set-reference').hidden).toBe(false);
		expect(screen.getByTestId('results-sim-reference-swap').parentElement!.hidden).toBe(true);
		expect(container.querySelectorAll('[data-testid="results-reference-diff"]')).toHaveLength(0);
		expect([...strip(container).querySelectorAll('svg')].every(icon => icon.getAttribute('aria-hidden') === 'true')).toBe(true);
	});

	it('flags the strip, swaps the buttons and fills the deltas in once a reference is set', () => {
		const { container } = mountView({ referenceDiffs: DIFFS });

		expect(strip(container).hasAttribute('data-has-reference')).toBe(true);
		expect(screen.getByTestId('results-sim-set-reference').hidden).toBe(true);
		expect(screen.getByTestId('results-sim-reference-swap').parentElement!.hidden).toBe(false);
		expect([...container.querySelectorAll('[data-testid="results-reference-diff"]')].map(diff => diff.textContent)).toEqual([
			'+100.00 (10.00%)',
			'+0.00 (0.00%)',
		]);
	});

	it('shows the first metric at full size and the rest compact', () => {
		const { container } = mountView();

		expect(
			[...container.querySelectorAll('[data-testid="results-metric"]')].map(row => [row.getAttribute('data-metric'), row.getAttribute('data-size')]),
		).toEqual([
			['dps', 'md'],
			['tps', 'sm'],
		]);
	});

	it.each(['fresh', 'stale'] as const)('marks a %s result in the header', freshness => {
		mountView({ freshness });

		expect(screen.getByTestId('results-freshness').getAttribute('data-freshness')).toBe(freshness);
		expect(screen.getByTestId('results-freshness').textContent).toBe(`sidebar.results.freshness.${freshness}`);
	});

	it('leaves the freshness out while nothing tracks it', () => {
		mountView({ freshness: null });

		expect(screen.queryByTestId('results-freshness')).toBeNull();
	});

	it('opens the breakdown through its handler, and has no link without one', () => {
		const onOpenBreakdown = vi.fn();
		const { unmount } = mountView({ onOpenBreakdown });
		fireEvent.click(screen.getByTestId('results-open-breakdown'));
		expect(onOpenBreakdown).toHaveBeenCalledOnce();
		unmount();

		mountView();
		expect(screen.queryByTestId('results-open-breakdown')).toBeNull();
	});

	it('reports each of the three reference actions', () => {
		const { onSetReference, onSwapReference, onClearReference } = mountView({ referenceDiffs: DIFFS });

		fireEvent.click(screen.getByTestId('results-sim-set-reference'));
		fireEvent.click(screen.getByTestId('results-sim-reference-swap'));
		fireEvent.click(screen.getByTestId('results-sim-reference-delete'));

		expect(onSetReference).toHaveBeenCalledOnce();
		expect(onSwapReference).toHaveBeenCalledOnce();
		expect(onClearReference).toHaveBeenCalledOnce();
	});
});
