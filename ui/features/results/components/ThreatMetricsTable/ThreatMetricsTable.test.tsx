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
import { ThreatMetricsTable } from './ThreatMetricsTable';

let result: SimResultData | null = null;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));

const playerResult = playerMetricsResult('getThreatActions');

describe('ThreatMetricsTable', () => {
	beforeEach(() => {
		result = null;
		mockActionMetricsMerge('threat', 'tps');
	});

	it('builds the ten-column shell before any result', () => {
		const { container } = render(<ThreatMetricsTable />);

		expect(container.querySelector('[data-testid="threat-metrics-root"]')).toBeTruthy();
		expect(container.querySelectorAll('thead th')).toHaveLength(10);
		expect(rows(container)).toHaveLength(0);
	});

	// getThreatActions, not getDamageActions: a taunt has no dps and would never reach this table
	// through the damage list.
	it('opens sorted by TPS descending and keeps a pet group as a parent with its children', () => {
		result = playerResult(
			[metric('Shield Slam', { tps: 300, threat: 6000 }), metric('Taunt', { tps: 5, threat: 100 })],
			[pet('Claw', { tps: 4, threat: 80 }), pet('Bite', { tps: 6, threat: 120 })],
		);
		const { container } = render(<ThreatMetricsTable />);

		expect(rows(container).map(row => [row.cells[0].textContent, row.hasAttribute('data-child')])).toEqual([
			['Shield Slam', false],
			['Claw', false],
			['Bite', true],
			['Claw', true],
			['Taunt', false],
		]);
	});

	it('splits TPS into per-cast and per-hit threat', async () => {
		result = playerResult([metric('Shield Slam')]);
		const { container } = render(<ThreatMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td.text-success')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual(['results_tab.details.tooltip_table.per_cast', 'results_tab.details.tooltip_table.per_hit']);
	});

	it('breaks the hits cell down by outcome, resisted hits included', async () => {
		result = playerResult([metric('Shield Slam', { hits: 8, resistedHits: 2, crits: 2, blockedCrits: 1 })]);
		const { container } = render(<ThreatMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[4]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual([
			'results_tab.details.attack_types.hit',
			'results_tab.details.attack_types.resisted_hit',
			'results_tab.details.attack_types.critical_hit',
			'results_tab.details.attack_types.blocked_critical_hit',
		]);
	});

	it('breaks the miss cell into miss, parry and dodge', async () => {
		result = playerResult([
			metric('Shield Slam', {
				misses: 5,
				missPercent: 5,
				parries: 3,
				parryPercent: 3,
				dodges: 1,
				dodgePercent: 1,
				totalMisses: 9,
				totalMissesPercent: 9,
			}),
		]);
		const { container } = render(<ThreatMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[7]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual([
			'results_tab.details.attack_types.miss',
			'results_tab.details.attack_types.parry',
			'results_tab.details.attack_types.dodge',
		]);
	});

	it('dashes a passive action on Avg Cast and sorts it as 0', () => {
		result = playerResult([metric('Shield Slam', { tps: 300 }), metric('Thunderfury', { isPassiveAction: true, tps: 20 })]);
		const { container } = render(<ThreatMetricsTable />);

		const passive = rows(container).find(row => row.cells[0].textContent === 'Thunderfury')!;
		expect(passive.cells[3].getAttribute('data-text')).toBe('0');
		expect(passive.cells[3].textContent).toBe('-');
	});

	it('anchors the avg-cast header tooltip on the header cell itself', () => {
		const { container } = render(<ThreatMetricsTable />);
		const header = [...container.querySelectorAll('thead th')][3];

		expect(header.getAttribute('data-tooltip-id')).toBe('threat-metrics-avg-cast-header');
		expect(header.getAttribute('data-tooltip-content')).toBeTruthy();
	});
});
