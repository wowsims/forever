import { actionMetric as metric, castMetricsResult, metricRows, mockActionMetricsMerge, petActionMetric as pet } from '@testing/results-metrics';
import { render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { HidePassivesContext } from '../../context/HidePassivesContext';
import type { SimResultData } from '../../model/result_data';
import { CastMetricsTable } from './CastMetricsTable';

let result: SimResultData | null = null;

vi.mock('../../hooks/useSimResult', () => ({ useSimResult: () => result }));
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));

const tableRows = (container: HTMLElement, testId: string) =>
	metricRows(container.querySelector<HTMLElement>(`[data-testid="${testId}"]`)!).map(row => [row.cells[0].textContent, row.hasAttribute('data-child')]);

const playerRows = (container: HTMLElement) => tableRows(container, 'cast-metrics-root');
const petTable = (container: HTMLElement) => container.querySelector('[data-testid="pet-cast-metrics-root"]');

describe('CastMetricsTable', () => {
	beforeEach(() => {
		result = null;
		mockActionMetricsMerge('casts', 'castsPerMinute');
	});

	it('lists every action with casts while passives are shown, which they are by default', () => {
		result = castMetricsResult([metric('Mortal Strike', { casts: 20 }), metric('Melee', { isPressed: false })]);
		const { container } = render(<CastMetricsTable />);

		expect(playerRows(container)).toEqual([
			['Mortal Strike', false],
			['Melee', false],
		]);
	});

	it('leaves out an action the player does not press once passives are hidden', () => {
		result = castMetricsResult([metric('Mortal Strike', { casts: 20 }), metric('Melee', { isPressed: false })]);
		const { container } = render(
			<HidePassivesContext.Provider value={true}>
				<CastMetricsTable />
			</HidePassivesContext.Provider>,
		);

		expect(playerRows(container)).toEqual([['Mortal Strike', false]]);
	});

	it('puts the pet under its own table, without its melee while passives are hidden', () => {
		result = castMetricsResult([metric('Steady Shot')], [pet('Claw'), pet('Melee', { isPressed: false })]);
		const { container } = render(
			<HidePassivesContext.Provider value={true}>
				<CastMetricsTable />
			</HidePassivesContext.Provider>,
		);

		expect(playerRows(container)).toEqual([['Steady Shot', false]]);
		expect(tableRows(container, 'pet-cast-metrics-root')).toEqual([
			['Claw', false],
			['Claw', true],
		]);
	});

	it('draws no pet table without a pet action to list', () => {
		result = castMetricsResult([metric('Steady Shot')]);
		const withoutPets = render(<CastMetricsTable />);
		expect(petTable(withoutPets.container)).toBeNull();
		withoutPets.unmount();

		result = castMetricsResult([metric('Steady Shot')], [pet('Melee', { isPressed: false })]);
		const { container } = render(
			<HidePassivesContext.Provider value={true}>
				<CastMetricsTable />
			</HidePassivesContext.Provider>,
		);
		expect(petTable(container)).toBeNull();
	});
});
