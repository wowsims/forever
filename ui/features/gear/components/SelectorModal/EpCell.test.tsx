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

	it('shows one decimal for a delta below the rounding boundary and a whole number from it', () => {
		const { unmount } = render(<EpCell ep={4.6} equippedEP={5} />);
		expect(delta().textContent).toBe('-0.4');
		unmount();

		const small = render(<EpCell ep={15} equippedEP={5} />);
		expect(delta().textContent).toBe('+10');
		small.unmount();

		render(<EpCell ep={14.94} equippedEP={5} />);
		expect(delta().textContent).toBe('+9.9');
	});

	it('signs the delta against the equipped EP', () => {
		const { unmount } = render(<EpCell ep={8} equippedEP={5} />);
		expect(delta().textContent).toBe('+3.0');
		expect(delta().getAttribute('data-sign')).toBe('positive');
		unmount();

		render(<EpCell ep={5} equippedEP={8} />);
		expect(delta().textContent).toBe('-3.0');
		expect(delta().getAttribute('data-sign')).toBe('negative');
	});
});
