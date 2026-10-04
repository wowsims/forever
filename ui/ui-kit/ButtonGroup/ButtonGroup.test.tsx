import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ButtonGroup } from './ButtonGroup';

describe('ButtonGroup', () => {
	it('renders a group role around its children', () => {
		render(
			<ButtonGroup>
				<button type="button">One</button>
				<button type="button">Two</button>
			</ButtonGroup>,
		);
		const group = screen.getByRole('group');
		expect(screen.getAllByRole('button').map(button => button.parentElement)).toEqual([group, group]);
	});

	it('carries the small size as a data attribute and no size attribute otherwise', () => {
		render(
			<>
				<ButtonGroup size="sm" testId="small" />
				<ButtonGroup testId="default" />
			</>,
		);
		expect(screen.getByTestId('small').getAttribute('data-size')).toBe('sm');
		expect(screen.getByTestId('default').hasAttribute('data-size')).toBe(false);
	});

	it('marks a spaced group with data-gap, and a joined one without it', () => {
		render(
			<>
				<ButtonGroup gap testId="spaced" />
				<ButtonGroup testId="joined" />
			</>,
		);
		expect(screen.getByTestId('spaced').hasAttribute('data-gap')).toBe(true);
		expect(screen.getByTestId('joined').hasAttribute('data-gap')).toBe(false);
	});

	it('lets a caller override the role', () => {
		render(<ButtonGroup role="toolbar" testId="group" />);
		expect(screen.getByTestId('group').getAttribute('role')).toBe('toolbar');
	});
});
