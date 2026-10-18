import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { MetricsCombinedTooltip, type MetricsCombinedTooltipGroup } from './MetricsCombinedTooltip';

const group = (name: string | undefined, data: MetricsCombinedTooltipGroup['data']): MetricsCombinedTooltipGroup => ({
	name,
	totalPercentage: 100,
	spellSchool: null,
	data,
});

// The count cell holds a total bar: its share, the bar, then the amount; the amount is the bar's last child.
const cellText = (cell: HTMLTableCellElement) =>
	cell.querySelector('[data-testid="metrics-total-bar-fill"]') ? cell.firstElementChild?.lastElementChild?.textContent : cell.textContent;

const cells = (container: HTMLElement) => [...container.querySelectorAll<HTMLTableRowElement>('tbody tr')].map(row => [...row.cells].map(cellText));

describe('MetricsCombinedTooltip', () => {
	it('drops zero-valued entries and orders the rest by value, descending', () => {
		const { container } = render(
			<MetricsCombinedTooltip
				groups={[
					group(undefined, [
						{ name: 'Hit', value: 10, percentage: 10 },
						{ name: 'Dodge', value: 0, percentage: 0 },
						{ name: 'Crit', value: 40, percentage: 40 },
					]),
				]}
			/>,
		);
		expect(cells(container)).toEqual([
			['Crit', '40'],
			['Hit', '10'],
		]);
	});

	it('does not sort the array it was given', () => {
		const data = [
			{ name: 'Hit', value: 10, percentage: 10 },
			{ name: 'Crit', value: 40, percentage: 40 },
		];
		render(<MetricsCombinedTooltip groups={[group(undefined, data)]} />);
		expect(data.map(entry => entry.name)).toEqual(['Hit', 'Crit']);
	});

	it('adds the average column only when an entry carries one', () => {
		const withoutAverage = render(<MetricsCombinedTooltip groups={[group(undefined, [{ name: 'Hit', value: 10, percentage: 10 }])]} />);
		expect(withoutAverage.container.querySelectorAll('thead th')).toHaveLength(2);

		const withAverage = render(<MetricsCombinedTooltip groups={[group(undefined, [{ name: 'Hit', value: 10, percentage: 10, average: 5 }])]} />);
		expect([...withAverage.container.querySelectorAll('thead th')].map(th => th.textContent)).toEqual([
			'results_tab.details.tooltip_table.type',
			'results_tab.details.tooltip_table.count',
			'results_tab.details.tooltip_table.average',
		]);
		expect(cells(withAverage.container)).toEqual([['Hit', '10', '5']]);
	});

	it('adds the min and max columns around the average only when an entry carries them', () => {
		const withRange = render(
			<MetricsCombinedTooltip groups={[group(undefined, [{ name: 'Hit', value: 10, percentage: 10, average: 10, min: 8, max: 12 }])]} />,
		);
		expect([...withRange.container.querySelectorAll('thead th')].map(th => th.textContent)).toEqual([
			'results_tab.details.tooltip_table.type',
			'results_tab.details.tooltip_table.count',
			'results_tab.details.tooltip_table.min',
			'results_tab.details.tooltip_table.average',
			'results_tab.details.tooltip_table.max',
		]);
		expect(cells(withRange.container)).toEqual([['Hit', '10', '8', '10', '12']]);

		const withoutRange = render(<MetricsCombinedTooltip groups={[group(undefined, [{ name: 'Hit', value: 10, percentage: 10, average: 10 }])]} />);
		expect(withoutRange.container.querySelectorAll('thead th')).toHaveLength(3);
	});

	it('leaves the min and max cells empty on a row without a range while another row has one', () => {
		const { container } = render(
			<MetricsCombinedTooltip
				groups={[
					group(undefined, [
						{ name: 'Hit', value: 40, percentage: 40, min: 8, max: 12 },
						{ name: 'Crit', value: 10, percentage: 10 },
					]),
				]}
			/>,
		);
		expect(cells(container)).toEqual([
			['Hit', '40', '8', '12'],
			['Crit', '10', '', ''],
		]);
	});

	it('takes the header overrides by position and leaves the holes at their defaults', () => {
		const { container } = render(
			<MetricsCombinedTooltip headerValues={[undefined, 'Amount']} groups={[group(undefined, [{ name: 'Hit', value: 10, percentage: 10 }])]} />,
		);
		expect([...container.querySelectorAll('thead th')].map(th => th.textContent)).toEqual(['results_tab.details.tooltip_table.type', 'Amount']);
	});

	it('writes a group header row only when there is more than one named group', () => {
		const one = render(<MetricsCombinedTooltip groups={[group('Hits', [{ name: 'Hit', value: 10, percentage: 10 }])]} />);
		expect(one.container.querySelectorAll('[data-testid="metrics-table-group-header"]')).toHaveLength(0);

		const two = render(
			<MetricsCombinedTooltip
				groups={[group('Hits', [{ name: 'Hit', value: 10, percentage: 10 }]), group('Ticks', [{ name: 'Tick', value: 5, percentage: 5 }])]}
			/>,
		);
		expect([...two.container.querySelectorAll('[data-testid="metrics-table-group-header"]')].map(row => row.textContent)).toEqual(['Hits', 'Ticks']);
	});

	it('renders a total bar per row', () => {
		const { container } = render(<MetricsCombinedTooltip groups={[group(undefined, [{ name: 'Hit', value: 10, percentage: 10 }])]} />);
		expect(container.querySelectorAll('[data-testid="metrics-total-bar-fill"]')).toHaveLength(1);
	});
});
