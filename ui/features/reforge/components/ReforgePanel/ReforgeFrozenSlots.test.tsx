import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ReforgeFrozenSlotsView } from './ReforgeFrozenSlots';

const PICKERS = ['head', 'neck', 'shoulder'].map(slot => <span key={slot} data-testid={`freeze-${slot}`} />);

describe('ReforgeFrozenSlotsView', () => {
	it('lays the slot toggles out in the order given', () => {
		render(<ReforgeFrozenSlotsView hidden={false} pickers={PICKERS} />);
		const grid = screen.getByTestId('reforge-frozen-slots');

		expect(Array.from(grid.children).map(child => child.getAttribute('data-testid'))).toEqual(['freeze-head', 'freeze-neck', 'freeze-shoulder']);
	});

	it('hides the toggles while slot freezing is off, keeping them mounted', () => {
		render(<ReforgeFrozenSlotsView hidden pickers={PICKERS} />);

		expect(screen.getByTestId('reforge-frozen-slots').hidden).toBe(true);
		expect(screen.getByTestId('freeze-head')).toBeTruthy();
	});
});
