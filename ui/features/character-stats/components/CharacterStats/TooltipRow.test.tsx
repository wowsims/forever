import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TooltipRow } from './TooltipRow';

describe('TooltipRow', () => {
	it('sets the label against the value', () => {
		const { getByTestId } = render(<TooltipRow label="Gear" value="+180" />);
		const row = getByTestId('character-stats-tooltip-row');
		expect([...row.children].map(cell => cell.textContent)).toEqual(['Gear', '+180']);
	});
});
