import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TooltipButton } from './TooltipButton';

describe('TooltipButton', () => {
	it('renders an icon button with the default testid and the regular question glyph', () => {
		render(<TooltipButton tooltip="What this does" />);
		const button = screen.getByRole('button');
		expect(button.getAttribute('data-testid')).toBe('tooltip-button');
		expect(button.getAttribute('type')).toBe('button');
		const icon = button.querySelector('svg')!;
		expect(icon.getAttribute('data-icon')).toBe('circle-question');
		expect(icon.getAttribute('data-prefix')).toBe('far');
	});

	it('takes its accessible name from a string tooltip, or from the label a rich tooltip comes with', () => {
		render(
			<>
				<TooltipButton tooltip="What this does" />
				<TooltipButton tooltip={<b>Rich</b>} label="Help" />
			</>,
		);
		expect(screen.getByRole('button', { name: 'What this does' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Help' })).toBeTruthy();
	});

	it('carries the testId it is given in place of the tooltip-button default', () => {
		render(<TooltipButton tooltip="What this does" testId="consumes-help" />);
		expect(screen.getByRole('button').getAttribute('data-testid')).toBe('consumes-help');
	});

	it('gives each button its own tooltip', () => {
		render(
			<>
				<TooltipButton tooltip="First" />
				<TooltipButton tooltip="Second" />
			</>,
		);
		const [first, second] = screen.getAllByRole('button');
		expect(first.dataset.tooltipId).not.toBe(second.dataset.tooltipId);

		fireEvent.mouseEnter(second);
		expect(screen.getByText('Second')).toBeTruthy();
		expect(screen.queryByText('First')).toBeNull();
	});
});
