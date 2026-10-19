import type { ActionMetrics } from '@sim/proto/sim_result';
import {
	actionMetric as metric,
	metricRows as rows,
	metricsTooltipRowNames,
	metricsTooltipTable as tooltipTable,
	mockActionMetricsMerge,
} from '@testing/results-metrics';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimResultData } from '../../model/result_data';
import { DtpsMetricsTable } from './DtpsMetricsTable';

let result: SimResultData | null = null;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));

// A target's actions. `makeNewTarget` gives every target a null `petActionId`, so `isPet` is false on
// every row this table can build — which is the whole of defect 12.
const targetResult = (...targets: Array<Array<ActionMetrics>>) =>
	({
		filter: {},
		result: {
			getRaidIndexedPlayers: () => [{ unitIndex: 0, totalDamageTaken: 400 }],
			getTargets: () => targets.map(actions => ({ getDamageActions: () => actions })),
			getOwnSideDamageOn: () => [],
		},
	}) as unknown as SimResultData;

describe('DtpsMetricsTable', () => {
	beforeEach(() => {
		result = null;
		mockActionMetricsMerge('damage', 'dps');
	});

	it('builds the nine-column shell before any result', () => {
		const { container } = render(<DtpsMetricsTable />);

		expect(container.querySelector('[data-testid="dtps-metrics-root"]')).toBeTruthy();
		expect(container.querySelectorAll('thead th')).toHaveLength(9);
		expect(rows(container)).toHaveLength(0);
	});

	it('renders a single-entry target group as a plain row, with no parent above it', () => {
		result = targetResult([metric('Melee')]);
		const { container } = render(<DtpsMetricsTable />);

		expect(rows(container).map(row => [row.cells[0].textContent, row.hasAttribute('data-parent'), row.hasAttribute('data-child')])).toEqual([
			['Melee', false, false],
		]);
	});

	it("gives each row its share of the player's damage taken", () => {
		result = targetResult([metric('Melee', { avgDamage: 100 })]);
		const { container } = render(<DtpsMetricsTable />);

		expect(rows(container)[0].cells[1].textContent).toContain('25%');
	});

	it('merges the same ability across targets into one parent with a child per target', () => {
		result = targetResult([metric('Melee', { dps: 30 })], [metric('Melee', { dps: 10 })]);
		const { container } = render(<DtpsMetricsTable />);

		expect(rows(container).map(row => [row.cells[0].textContent, row.hasAttribute('data-parent'), row.hasAttribute('data-child')])).toEqual([
			['Melee', true, false],
			['Melee', false, true],
			['Melee', false, true],
		]);
	});

	it('sorts a passive action as 0 on Avg Cast while still showing its value', () => {
		result = targetResult([metric('Melee', { dps: 30 }), metric('Cleave', { isPassiveAction: true, avgCastHit: 5, dps: 20 })]);
		const { container } = render(<DtpsMetricsTable />);

		const passive = rows(container).find(row => row.cells[0].textContent === 'Cleave')!;
		expect(passive.cells[3].getAttribute('data-text')).toBe('0');
		expect(passive.cells[3].textContent).toBe('5');
	});

	it('opens the damage-taken breakdown from the primary-metric cell', async () => {
		result = targetResult([metric('Melee')]);
		const { container } = render(<DtpsMetricsTable />);

		fireEvent.mouseEnter(container.querySelector('td[data-primary-metric]')!);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual(['results_tab.details.attack_types.hit', 'results_tab.details.attack_types.critical_hit']);
	});

	it('counts blocked hits and crushing blows out of the hit row of the casts breakdown', async () => {
		result = targetResult([
			metric('Melee', { landedHits: 100, landedRolls: 100, casts: 112, totalMisses: 12, misses: 12, blocks: 20, blockedCrits: 5, crushes: 10 }),
		]);
		const { container } = render(<DtpsMetricsTable />);

		fireEvent.mouseEnter(rows(container)[0].cells[2]);
		await waitFor(() => expect(tooltipTable()).toBeTruthy());
		expect(metricsTooltipRowNames()).toEqual([
			'results_tab.details.attack_types.hit',
			'results_tab.details.attack_types.blocked_hit',
			'results_tab.details.attack_types.miss',
			'results_tab.details.attack_types.crushing_blow',
		]);
		expect(rows(tooltipTable() as HTMLElement)[0].cells[1].textContent?.endsWith('65')).toBe(true);
	});

	it('carries no tooltip on Avg Hit or on the rate column, so a target reports no threat', () => {
		result = targetResult([metric('Melee')]);
		const { container } = render(<DtpsMetricsTable />);

		const cells = rows(container)[0].cells;
		expect(cells[5].getAttribute('data-tooltip-id')).toBeNull();
		expect(cells[8].getAttribute('data-tooltip-id')).toBeNull();
		expect(container.querySelector('td.text-success')?.getAttribute('data-tooltip-id')).toBeNull();
	});

	it('anchors both header tooltips on their own header cell', () => {
		const { container } = render(<DtpsMetricsTable />);
		const headers = [...container.querySelectorAll('thead th')];

		expect(headers[3].getAttribute('data-tooltip-id')).toBe('dtps-metrics-avg-cast-header');
		expect(headers[3].getAttribute('data-tooltip-content')).toBeTruthy();
		expect(headers[6].getAttribute('data-tooltip-id')).toBe('dtps-metrics-miss-percent-header');
		expect(headers[6].getAttribute('data-tooltip-content')).toBeTruthy();
	});

	it('renders no bar denominator at all before a result, rather than -Infinity', () => {
		const { container } = render(<DtpsMetricsTable />);
		expect(container.querySelectorAll('[data-testid="metrics-total-bar-fill"]')).toHaveLength(0);
	});
});
