import {
	actionMetric as metric,
	metricRows as rows,
	metricsTooltipRowNames,
	metricsTooltipTable as tooltipTable,
	mockActionMetricsMerge,
	petActionMetric as pet,
	playerMetricsResult,
} from '@testing/results-metrics';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { DamageMetricsTable } from './DamageMetricsTable';

let result: SimResultData | null = null;
let showThreatMetrics = true;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));
vi.mock('@sim/context/SimHostContext', async importOriginal => ({
	...(await importOriginal<typeof import('@sim/context/SimHostContext')>()),
	// `useDisplayMetrics` reads all three flags through one subscription; only threat varies per case here.
	useSim: () => ({ getShowThreatMetrics: () => showThreatMetrics, getShowDamageMetrics: () => true, getShowHealingMetrics: () => true }),
}));
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());

const playerResult = playerMetricsResult('getDamageActions');

describe('DamageMetricsTable', () => {
	beforeEach(() => {
		result = null;
		showThreatMetrics = true;
		mockActionMetricsMerge('damage', 'dps');
	});

	it('builds the ten-column shell before any result', () => {
		const { container } = render(<DamageMetricsTable />);

		expect(container.querySelector('[data-testid="damage-metrics-root"]')).toBeTruthy();
		expect(container.querySelectorAll('thead th')).toHaveLength(10);
		expect(rows(container)).toHaveLength(0);
	});

	it('opens sorted by DPS descending and keeps a pet group as a parent with its children', () => {
		result = playerResult([metric('Steady Shot', { dps: 30 }), metric('Auto Shot', { dps: 5 })], [pet('Claw', { dps: 4 }), pet('Bite', { dps: 6 })]);
		const { container } = render(<DamageMetricsTable />);

		expect(rows(container).map(row => [row.cells[0].textContent, row.hasAttribute('data-expanded'), row.hasAttribute('data-child')])).toEqual([
			['Steady Shot', false, false],
			['Claw', true, false],
			['Bite', false, true],
			['Claw', false, true],
			['Auto Shot', false, false],
		]);
	});

	it('marks a row with no hit attempts and no dps as threat-only', () => {
		result = playerResult([metric('Steady Shot'), metric('Vigilance', { hitAttempts: 0, dps: 0 })]);
		const { container } = render(<DamageMetricsTable />);

		expect(rows(container).map(row => [row.cells[0].textContent, row.getAttribute('data-testid') === 'threat-metrics'])).toEqual([
			['Steady Shot', false],
			['Vigilance', true],
		]);
	});

	it('drops a threat-only row entirely once threat metrics are hidden', () => {
		result = playerResult([metric('Steady Shot'), metric('Vigilance', { hitAttempts: 0, dps: 0 })]);
		showThreatMetrics = false;
		const { container } = render(<DamageMetricsTable />);

		expect(rows(container).map(row => row.cells[0].textContent)).toEqual(['Steady Shot']);
	});

	it('dashes a passive action on Avg Cast and sorts it as 0', () => {
		result = playerResult([metric('Steady Shot', { dps: 30 }), metric('Serpent Sting', { isPassiveAction: true, avgCastHit: 5, dps: 20 })]);
		const { container } = render(<DamageMetricsTable />);

		const passive = rows(container).find(row => row.cells[0].textContent === 'Serpent Sting')!;
		expect(passive.cells[3].getAttribute('data-text')).toBe('0');
		expect(passive.cells[3].textContent).toBe('-');
	});

	it('opens the damage breakdown from the primary-metric cell', async () => {
		result = playerResult([metric('Steady Shot')]);
		const { container } = render(<DamageMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual(['results_tab.details.attack_types.hit', 'results_tab.details.attack_types.critical_hit']);
	});

	it('counts blocked hits out of the hit row of the casts breakdown, and shows no crushing blows', async () => {
		result = playerResult([
			metric('Steady Shot', { landedHits: 100, landedRolls: 100, casts: 112, totalMisses: 12, misses: 12, blocks: 20, blockedCrits: 5, crushes: 10 }),
		]);
		const { container } = render(<DamageMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[2]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual([
			'results_tab.details.attack_types.hit',
			'results_tab.details.attack_types.blocked_hit',
			'results_tab.details.attack_types.miss',
		]);
		expect(rows(tooltipTable() as HTMLElement)[0].cells[1].textContent?.endsWith('75')).toBe(true);
	});

	it('takes the misses out of the casts of a DoT on the hit row of the casts breakdown', async () => {
		result = playerResult([metric('Corruption', { landedHits: 0, landedRolls: 0, casts: 10, totalMisses: 1, misses: 1, missPercent: 10 })]);
		const { container } = render(<DamageMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[2]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		const hit = rows(tooltipTable() as HTMLElement)[0].cells[1].textContent;
		expect(hit?.startsWith('90%')).toBe(true);
		expect(hit?.endsWith('9')).toBe(true);
	});

	it('resolves a child row against its own metric, not its parent', async () => {
		result = playerResult([], [pet('Claw', { dps: 4, hits: 40, crits: 0 }), pet('Bite', { dps: 6, hits: 60, crits: 0 })]);
		const { container } = render(<DamageMetricsTable />);

		const child = rows(container).find(row => row.hasAttribute('data-child') && row.cells[0].textContent === 'Claw')!;
		fireEvent.mouseEnter(child.cells[4]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual(['results_tab.details.attack_types.hit']);
		expect(tooltipTable()!.textContent).toContain('40');
	});

	it('keeps a threat tooltip shut while threat metrics are hidden, and opens it once they are shown', async () => {
		result = playerResult([metric('Steady Shot')]);
		showThreatMetrics = false;
		const hidden = render(<DamageMetricsTable />);

		fireEvent.mouseEnter(hidden.container.querySelector('td.text-success')!);
		await waitFor(() => expect(hidden.container.querySelector('td.text-success')?.getAttribute('data-tooltip-id')).toBeTruthy());
		expect(tooltipTable()).toBeNull();
		hidden.unmount();

		showThreatMetrics = true;
		const shown = render(<DamageMetricsTable />);
		fireEvent.mouseEnter(shown.container.querySelector('td.text-success')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(tooltipTable()!.textContent).toContain('results_tab.details.attack_types.threat');
	});

	it('anchors the avg-cast header tooltip on the header cell itself', () => {
		const { container } = render(<DamageMetricsTable />);
		const header = [...container.querySelectorAll('thead th')][3];

		expect(header.getAttribute('data-tooltip-id')).toBe('damage-metrics-avg-cast-header');
		expect(header.getAttribute('data-tooltip-content')).toBeTruthy();
	});

	it('leaves no tooltip behind when the table unmounts while one is open', async () => {
		result = playerResult([metric('Steady Shot')]);
		const { container, unmount } = render(<DamageMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());

		unmount();
		expect(document.querySelectorAll('.sim-tooltip')).toHaveLength(0);
	});

	it('renders no bar denominator at all before a result, rather than -Infinity', () => {
		const { container } = render(<DamageMetricsTable />);
		expect(container.querySelectorAll('[data-testid="metrics-total-bar-fill"]')).toHaveLength(0);
	});
});
