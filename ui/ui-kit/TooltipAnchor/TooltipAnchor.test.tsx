import { fireEvent, render, screen } from '@testing-library/react';
import { Tooltip } from '@ui-kit/Tooltip';
import { describe, expect, it } from 'vitest';

import { TooltipAnchor } from './TooltipAnchor';

describe('TooltipAnchor', () => {
	it('renders a focusable span with no role, carrying the tooltip id', () => {
		render(
			<TooltipAnchor tooltipId="crit-cap" testId="stat-value-link">
				13.00%
			</TooltipAnchor>,
		);
		const anchor = screen.getByTestId('stat-value-link');
		expect(anchor.tagName).toBe('SPAN');
		expect(anchor.tabIndex).toBe(0);
		expect(anchor.getAttribute('role')).toBeNull();
		expect(anchor.getAttribute('data-tooltip-id')).toBe('crit-cap');
		expect(anchor.hasAttribute('data-tooltip-content')).toBe(false);
		expect(anchor.textContent).toBe('13.00%');
		expect(screen.queryByRole('button')).toBeNull();
	});

	it('opens its tooltip on hover and on keyboard focus', async () => {
		const { unmount } = render(
			<>
				<TooltipAnchor tooltipId="avoidance" testId="anchor">
					102.40%
				</TooltipAnchor>
				<Tooltip id="avoidance" content="Crush capped" />
			</>,
		);
		fireEvent.mouseEnter(screen.getByTestId('anchor'));
		expect(await screen.findByText('Crush capped')).toBeTruthy();
		unmount();

		render(
			<>
				<TooltipAnchor tooltipId="miss" testId="anchor">
					5.00%
				</TooltipAnchor>
				<Tooltip id="miss" content="Base miss" />
			</>,
		);
		fireEvent.focus(screen.getByTestId('anchor'));
		expect(await screen.findByText('Base miss')).toBeTruthy();
	});
});
