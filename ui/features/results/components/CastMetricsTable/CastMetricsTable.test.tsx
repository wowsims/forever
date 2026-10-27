import { actionMetric as metric, castMetricsResult, metricRows, mockActionMetricsMerge, petActionMetric as pet } from '@testing/results-metrics';
import { render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ShowApmBreakdownContext } from '../../context/ShowApmBreakdownContext';
import type { SimResultData } from '../../model/result_data';
import { CastMetricsTable } from './CastMetricsTable';

let result: SimResultData | null = null;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));

const rowsOf = (container: HTMLElement, testId: string) => metricRows(container.querySelector<HTMLElement>(`[data-testid="${testId}"]`)!);
const tableRows = (container: HTMLElement, testId: string) => rowsOf(container, testId).map(row => [row.cells[0].textContent, row.hasAttribute('data-child')]);

const playerRows = (container: HTMLElement) => tableRows(container, 'cast-metrics-root');
const playerCells = (container: HTMLElement) => rowsOf(container, 'cast-metrics-root').map(row => [...row.cells].map(cell => cell.textContent));
const petTable = (container: HTMLElement) => container.querySelector('[data-testid="pet-cast-metrics-root"]');

const renderBreakdown = () =>
	render(
		<ShowApmBreakdownContext.Provider value={true}>
			<CastMetricsTable />
		</ShowApmBreakdownContext.Provider>,
	);

describe('CastMetricsTable', () => {
	beforeEach(() => {
		result = null;
		mockActionMetricsMerge('casts', 'castsPerMinute', 'manualCasts', 'manualCastsPerMinute');
	});

	it('lists every action with casts while the APM breakdown is off, which it is by default', () => {
		result = castMetricsResult([
			metric('Mortal Strike', { casts: 20 }),
			metric('Melee', { manualCasts: 0 }),
			metric('Judgement', { casts: 0, manualCasts: 5 }),
		]);
		const { container } = render(<CastMetricsTable />);

		expect(playerRows(container)).toEqual([
			['Mortal Strike', false],
			['Melee', false],
		]);
	});

	it('lists only the presses APM counts once the breakdown is shown, with their press counts', () => {
		result = castMetricsResult([
			metric('Mortal Strike', { casts: 20, manualCasts: 15, manualCastsPerMinute: 7.5 }),
			metric('Melee', { manualCasts: 0 }),
			metric('Judgement', { casts: 0, manualCasts: 5, manualCastsPerMinute: 2.5 }),
		]);
		const { container } = renderBreakdown();

		expect(playerCells(container)).toEqual([
			['Mortal Strike', '15.0', '7.5'],
			['Judgement', '5.0', '2.5'],
		]);
	});

	it('puts the pet under its own table, which the APM breakdown leaves out', () => {
		result = castMetricsResult([metric('Steady Shot')], [pet('Claw'), pet('Melee')]);
		const allCasts = render(<CastMetricsTable />);
		expect(tableRows(allCasts.container, 'pet-cast-metrics-root')).toEqual([
			['Claw', false],
			['Claw', true],
			['Melee', true],
		]);
		allCasts.unmount();

		const { container } = renderBreakdown();
		expect(playerRows(container)).toEqual([['Steady Shot', false]]);
		expect(petTable(container)).toBeNull();
	});

	it('draws no pet table without a pet action to list', () => {
		result = castMetricsResult([metric('Steady Shot')]);
		const withoutPets = render(<CastMetricsTable />);
		expect(petTable(withoutPets.container)).toBeNull();
		withoutPets.unmount();

		result = castMetricsResult([metric('Steady Shot')], [pet('Melee', { casts: 0 })]);
		const { container } = render(<CastMetricsTable />);
		expect(petTable(container)).toBeNull();
	});
});
