import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Pill } from './Pill';

const attributes = (testId = 'pill') => {
	const pill = screen.getByTestId(testId);
	return { tone: pill.getAttribute('data-tone'), rounded: pill.getAttribute('data-rounded'), size: pill.getAttribute('data-size') };
};

describe('Pill', () => {
	it('defaults to a neutral, fully rounded, small pill around its children', () => {
		render(<Pill>+4.21</Pill>);

		expect(attributes()).toEqual({ tone: 'neutral', rounded: 'full', size: 'sm' });
		expect(screen.getByTestId('pill').textContent).toBe('+4.21');
	});

	it.each([
		['positive', 'xs'],
		['negative', 'md'],
		['warning', 'sm'],
		['primary', 'xs'],
	] as const)('carries tone %s and size %s as data attributes', (tone, size) => {
		render(<Pill tone={tone} size={size} />);

		expect(attributes()).toEqual({ tone, rounded: 'full', size });
	});

	it('passes other span props through, so it can anchor a tooltip', () => {
		render(<Pill testId="delta" data-tooltip-id="tip" aria-label="Delta" />);

		expect(screen.getByTestId('delta').getAttribute('data-tooltip-id')).toBe('tip');
		expect(screen.getByLabelText('Delta')).toBe(screen.getByTestId('delta'));
	});
});
