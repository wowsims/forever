import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TargetPickerView } from './TargetPicker';

describe('TargetPickerView', () => {
	it('lays the identity, stat and threat fields out as three sections', () => {
		const { getAllByTestId, queryByTestId } = render(
			<TargetPickerView identity={<span>npc</span>} stats={<span>armor</span>} threat={<span>swing speed</span>} />,
		);
		expect(getAllByTestId('target-picker-section').map(section => section.textContent)).toEqual(['npc', 'armor', 'swing speed']);
		expect(queryByTestId('target-picker-section-placeholder')).toBeNull();
	});

	it('holds the third column with a hidden placeholder without threat fields', () => {
		const { getAllByTestId, getByTestId } = render(<TargetPickerView identity={<span>npc</span>} stats={<span>armor</span>} threat={false} />);
		expect(getAllByTestId('target-picker-section')).toHaveLength(2);
		expect(getByTestId('target-picker-section-placeholder').getAttribute('aria-hidden')).toBe('true');
	});
});
