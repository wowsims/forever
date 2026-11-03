import { render, screen } from '@testing-library/react';
import { Button } from '@ui-kit/Button';
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

	it('carries its size as a data attribute and no size attribute otherwise', () => {
		render(
			<>
				<ButtonGroup size="xs" testId="xs" />
				<ButtonGroup size="sm" testId="sm" />
				<ButtonGroup size="md" testId="md" />
				<ButtonGroup testId="default" />
			</>,
		);
		expect(['xs', 'sm', 'md'].map(size => screen.getByTestId(size).getAttribute('data-size'))).toEqual(['xs', 'sm', 'md']);
		expect(screen.getByTestId('default').hasAttribute('data-size')).toBe(false);
	});

	it('sizes every Button in a sized group as the group, over the size the button sets itself', () => {
		render(
			<ButtonGroup size="xs">
				<Button>One</Button>
				<Button size="md">Two</Button>
			</ButtonGroup>,
		);
		expect(screen.getAllByRole('button').map(button => button.getAttribute('data-size'))).toEqual(['xs', 'xs']);
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
