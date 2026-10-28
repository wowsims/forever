import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { EncounterPickerView, type EncounterPickerViewProps } from './EncounterPicker';

const renderView = (props: Partial<EncounterPickerViewProps> = {}) =>
	render(
		<EncounterPickerView
			duration={<span>duration</span>}
			zone={<span>zone</span>}
			areaTypes={<span>area types</span>}
			targetInputs={<span>target inputs</span>}
			onAdvancedClick={() => {}}
			onDesignClick={() => {}}
			{...props}
		/>,
	);

describe('EncounterPickerView', () => {
	it('lays out the fields in order, and leaves out the ones not given', () => {
		const { getByTestId } = renderView({ execute: <span>execute</span>, allies: <span>allies</span> });
		const root = getByTestId('encounter-picker-root');
		expect([...root.children].map(child => child.textContent)).toEqual([
			'duration',
			'execute',
			'zone',
			'area types',
			'allies',
			'target inputs',
			'settings_tab.encounter.advancedsettings_tab.encounter.designer.open',
		]);
	});

	it('reports the Advanced and Design encounter presses', () => {
		const onAdvancedClick = vi.fn();
		const onDesignClick = vi.fn();
		const { getByTestId } = renderView({ onAdvancedClick, onDesignClick });
		fireEvent.click(getByTestId('advanced-button'));
		expect(onAdvancedClick).toHaveBeenCalledTimes(1);
		expect(onDesignClick).not.toHaveBeenCalled();
		fireEvent.click(getByTestId('design-encounter-button'));
		expect(onDesignClick).toHaveBeenCalledTimes(1);
	});
});
