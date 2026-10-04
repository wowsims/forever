import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { FakeLink } from './FakeLink';

describe('FakeLink', () => {
	it('activates on click', () => {
		const onActivate = vi.fn();
		render(<FakeLink onActivate={onActivate} />);

		fireEvent.click(screen.getByTestId('fake-link'));

		expect(onActivate).toHaveBeenCalledOnce();
	});

	it('adds no tab stop and no accessible target', () => {
		render(
			<div>
				<FakeLink onActivate={vi.fn()} />
				<button type="button">Real</button>
			</div>,
		);

		const layer = screen.getByTestId('fake-link');
		expect(layer.getAttribute('aria-hidden')).toBe('true');
		expect(layer.hasAttribute('tabindex')).toBe(false);
		expect(layer.getAttribute('role')).toBeNull();
		expect(screen.getAllByRole('button').map(button => button.textContent)).toEqual(['Real']);
	});

	it('leaves a sibling click to the sibling', () => {
		const onActivate = vi.fn();
		const onSibling = vi.fn();
		render(
			<div>
				<FakeLink onActivate={onActivate} />
				<button type="button" onClick={onSibling}>
					Real
				</button>
			</div>,
		);

		fireEvent.click(screen.getByRole('button', { name: 'Real' }));

		expect(onSibling).toHaveBeenCalledOnce();
		expect(onActivate).not.toHaveBeenCalled();
	});
});
