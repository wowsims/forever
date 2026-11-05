import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { EncounterPickerView, type EncounterPickerViewProps } from './EncounterPicker';

const renderView = (props: Partial<EncounterPickerViewProps> = {}) =>
	render(
		<EncounterPickerView
			duration={<span>duration</span>}
			zone={<span>zone</span>}
			areaTypes={<span>area types</span>}
			onDesignClick={() => {}}
			{...props}
		/>,
	);

describe('EncounterPickerView', () => {
	it('lays out the encounter choice first, then the designer button and the fields, and leaves out the ones not given', () => {
		const { getByTestId } = renderView({ preset: <span>preset</span>, execute: <span>execute</span>, allies: <span>allies</span> });
		const root = getByTestId('encounter-picker-root');
		expect([...root.children].map(child => child.textContent)).toEqual([
			'preset',
			'settings_tab.encounter.designer.open',
			'duration',
			'execute',
			'zone',
			'area types',
			'allies',
		]);
	});

	it('reports the Design encounter press', () => {
		const onDesignClick = vi.fn();
		const { getByTestId } = renderView({ onDesignClick });
		fireEvent.click(getByTestId('design-encounter-button'));
		expect(onDesignClick).toHaveBeenCalledTimes(1);
	});
});
