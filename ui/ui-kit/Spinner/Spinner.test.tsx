import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Spinner } from './Spinner';

describe('Spinner', () => {
	it('renders a div with the md fallback size by default', () => {
		const { container } = render(<Spinner />);
		const el = container.firstElementChild as HTMLElement;
		expect(el.tagName).toBe('DIV');
		expect(el.getAttribute('data-testid')).toBe('loader');
		expect(el.getAttribute('data-size')).toBe('md');
	});

	it.each(['xs', 'sm', 'md', 'lg'] as const)('applies data-size for size %s', size => {
		const { container } = render(<Spinner size={size} />);
		expect((container.firstElementChild as HTMLElement).getAttribute('data-size')).toBe(size);
	});

	it('carries the testId it is given in place of the loader default', () => {
		const { container } = render(<Spinner testId="bulk-spinner" />);
		expect((container.firstElementChild as HTMLElement).getAttribute('data-testid')).toBe('bulk-spinner');
	});

	it('is an image named by its label, and a plain element without one', () => {
		const { getByRole, container, rerender } = render(<Spinner label="Running" />);
		expect(getByRole('img', { name: 'Running' })).toBe(container.firstElementChild);

		rerender(<Spinner />);
		expect((container.firstElementChild as HTMLElement).hasAttribute('role')).toBe(false);
	});
});
