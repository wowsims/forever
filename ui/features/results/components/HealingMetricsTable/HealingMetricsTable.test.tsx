import {
	actionMetric as metric,
	metricRows as rows,
	metricsTooltipRowNames,
	metricsTooltipTable as tooltipTable,
	mockActionMetricsMerge,
	playerMetricsResult,
} from '@testing/results-metrics';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { HealingMetricsTable } from './HealingMetricsTable';

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

const playerResult = playerMetricsResult('getHealingActions');

describe('HealingMetricsTable', () => {
	beforeEach(() => {
		result = null;
		showThreatMetrics = true;
		mockActionMetricsMerge('healing', 'hps');
	});

	it('builds the twelve-column shell before any result', () => {
		const { container } = render(<HealingMetricsTable />);

		expect(container.querySelector('[data-testid="healing-metrics-root"]')).toBeTruthy();
		expect(container.querySelectorAll('thead th')).toHaveLength(12);
		expect(rows(container)).toHaveLength(0);
	});

	it('drops an action with no healing and opens sorted by HPS descending', () => {
		result = playerResult([metric('Healing Touch', { hps: 4 }), metric('Rejuvenation', { hps: 12 }), metric('Melee', { hps: 0 })]);
		const { container } = render(<HealingMetricsTable />);

		expect(rows(container).map(row => row.cells[0].textContent)).toEqual(['Rejuvenation', 'Healing Touch']);
	});

	it('draws the shielding overlay as a second bar fill', () => {
		result = playerResult([metric('Healing Touch')]);
		const { container } = render(<HealingMetricsTable />);

		expect(container.querySelectorAll('td[data-primary-metric] [data-testid="metrics-total-bar-fill"]')).toHaveLength(2);
	});

	it('drops the overlay bar when nothing is shielded', () => {
		result = playerResult([metric('Healing Touch', { shielding: 0 })]);
		const { container } = render(<HealingMetricsTable />);

		expect(container.querySelectorAll('td[data-primary-metric] [data-testid="metrics-total-bar-fill"]')).toHaveLength(1);
	});

	it('opens the healing breakdown with its average column from the primary-metric cell', async () => {
		result = playerResult([metric('Healing Touch')]);
		const { container } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect([...tooltipTable()!.querySelectorAll('thead th')]).toHaveLength(3);
		expect(metricsTooltipRowNames()).toEqual(['results_tab.details.attack_types.hit', 'results_tab.details.attack_types.critical_hit']);
	});

	it('keeps a threat tooltip shut while threat metrics are hidden, and opens it once they are shown', async () => {
		result = playerResult([metric('Healing Touch')]);
		showThreatMetrics = false;
		const hidden = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(hidden.container.querySelector('td.text-success')!);
		await waitFor(() => expect(hidden.container.querySelector('td.text-success')?.getAttribute('data-tooltip-id')).toBeTruthy());
		expect(tooltipTable()).toBeNull();
		hidden.unmount();

		showThreatMetrics = true;
		const shown = render(<HealingMetricsTable />);
		fireEvent.mouseEnter(shown.container.querySelector('td.text-success')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(tooltipTable()!.textContent).toContain('results_tab.details.attack_types.threat');
	});

	it('names the two groups of the hits breakdown', async () => {
		result = playerResult([metric('Healing Touch', { landedTicks: 4, ticks: 4 })]);
		const { container } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[6]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect([...tooltipTable()!.querySelectorAll('tbody th')].map(th => th.textContent)).toEqual([
			'results_tab.details.columns.hits',
			'results_tab.details.tooltip_table.ticks',
		]);
	});

	it('anchors all three header tooltips on their own header cell', () => {
		const { container } = render(<HealingMetricsTable />);
		const headers = [...container.querySelectorAll('thead th')];

		expect(headers[5].getAttribute('data-tooltip-id')).toBe('healing-metrics-avg-cast-header');
		expect(headers[6].getAttribute('data-tooltip-id')).toBe('healing-metrics-hits-header');
		expect(headers[7].getAttribute('data-tooltip-id')).toBe('healing-metrics-avg-hit-header');
		expect(headers.map(th => th.getAttribute('data-tooltip-content')).filter(Boolean)).toHaveLength(3);
	});

	it('leaves no tooltip behind when the table unmounts while one is open', async () => {
		result = playerResult([metric('Healing Touch')]);
		const { container, unmount } = render(<HealingMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());

		unmount();
		expect(document.querySelectorAll('.sim-tooltip')).toHaveLength(0);
	});

	it('renders no bar denominator at all before a result, rather than -Infinity', () => {
		const { container } = render(<HealingMetricsTable />);
		expect(container.querySelectorAll('[data-testid="metrics-total-bar-fill"]')).toHaveLength(0);
	});
});
