import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { EpCell } from './EpCell';

const value = () => screen.getByTestId('selector-modal-list-item-ep-value').textContent;
const delta = () => screen.getByTestId('selector-modal-list-item-ep-delta');

describe('EpCell', () => {
	it('shows one decimal below the rounding boundary and a whole number at or above it', () => {
		const { unmount } = render(<EpCell ep={9.94} equippedEP={null} />);
		expect(value()).toBe('9.9');
		unmount();

		render(<EpCell ep={9.95} equippedEP={null} />);
		expect(value()).toBe('10');
	});

	it('leaves the delta empty with nothing equipped or an equal EP', () => {
		const { unmount } = render(<EpCell ep={5} equippedEP={null} />);
		expect(delta().textContent).toBe('');
		unmount();

		render(<EpCell ep={5} equippedEP={5} />);
		expect(delta().textContent).toBe('');
		expect(delta().hasAttribute('data-sign')).toBe(false);
	});

	it('signs the delta against the equipped EP', () => {
		const { unmount } = render(<EpCell ep={8} equippedEP={5} />);
		expect(delta().textContent).toBe('+3');
		expect(delta().getAttribute('data-sign')).toBe('positive');
		unmount();

		render(<EpCell ep={5} equippedEP={8} />);
		expect(delta().textContent).toBe('-3');
		expect(delta().getAttribute('data-sign')).toBe('negative');
	});
});
