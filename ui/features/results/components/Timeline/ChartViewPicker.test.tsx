import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ChartViewPicker } from './ChartViewPicker';

const mount = (value: 'rotation' | 'dps') => {
	const onChange = vi.fn();
	return { onChange, ...render(<ChartViewPicker value={value} onChange={onChange} />) };
};

const views = () => screen.getAllByRole('button');

describe('ChartViewPicker', () => {
	it('offers the two views as one group of buttons', () => {
		mount('rotation');

		expect(screen.getByRole('group').contains(views()[0])).toBe(true);
		expect(views().map(button => button.textContent)).toEqual([
			'results_tab.details.timeline.chart_types.rotation',
			'results_tab.details.timeline.chart_types.dps',
		]);
	});

	it('reports the picked view, and stays quiet for the one already showing', () => {
		const { onChange } = mount('rotation');

		fireEvent.click(views()[0]);
		expect(onChange).not.toHaveBeenCalled();
		fireEvent.click(views()[1]);
		expect(onChange).toHaveBeenCalledWith('dps');
	});
});
