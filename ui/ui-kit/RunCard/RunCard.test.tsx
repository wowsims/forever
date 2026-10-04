import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { RunCard } from './RunCard';

const ids = (element: Element) => [...element.children].map(child => child.getAttribute('data-testid'));

describe('RunCard', () => {
	it('puts the controls ahead of the action, and the body under both', () => {
		render(
			<RunCard controls={<div data-testid="field" />} action={<button data-testid="run" />}>
				<div data-testid="result" />
			</RunCard>,
		);

		expect(ids(screen.getByTestId('run-card'))).toEqual(['run-card-controls', 'run-card-body']);
		expect(ids(screen.getByTestId('run-card-controls'))).toEqual(['field', 'run']);
		expect(ids(screen.getByTestId('run-card-body'))).toEqual(['result']);
	});

	it('leaves out an empty controls row and an empty body', () => {
		render(<RunCard testId="empty" />);

		expect(screen.getByTestId('empty').children).toHaveLength(0);
	});

	it('covers the card with a busy overlay only while loading', () => {
		const { rerender } = render(<RunCard action={<button />} loading />);
		expect(screen.getByTestId('run-card').getAttribute('aria-busy')).toBe('true');
		expect(screen.getByTestId('run-card-loading').querySelector('[data-testid="loader"]')).not.toBeNull();

		rerender(<RunCard action={<button />} />);
		expect(screen.getByTestId('run-card').hasAttribute('aria-busy')).toBe(false);
		expect(screen.queryByTestId('run-card-loading')).toBeNull();
	});
});
