import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SidebarDisabledContext } from './context/sidebar_disabled';
import { SidebarActionButton } from './SidebarActionButton';

describe('SidebarActionButton', () => {
	it('renders a sidebar action carrying its testId, and fires onClick', () => {
		const onClick = vi.fn();
		render(
			<SidebarActionButton testId="dps-action" onClick={onClick}>
				Simulate
			</SidebarActionButton>,
		);
		const button = screen.getByRole('button', { name: 'Simulate' });
		expect(button.getAttribute('data-testid')).toBe('dps-action');
		expect(button.hasAttribute('data-sidebar-action')).toBe(true);

		fireEvent.click(button);
		expect(onClick).toHaveBeenCalledTimes(1);
	});

	it.each([
		[undefined, 'primary'],
		['cancel', 'cancel'],
	] as const)('renders the %s variant as %s', (variant, expected) => {
		render(
			<SidebarActionButton variant={variant} onClick={() => {}}>
				Stop
			</SidebarActionButton>,
		);

		expect(screen.getByRole('button', { name: 'Stop' }).getAttribute('data-variant')).toBe(expected);
	});

	it('shows no spinner while idle', () => {
		render(<SidebarActionButton onClick={() => {}}>Simulate</SidebarActionButton>);
		const button = screen.getByRole('button', { name: 'Simulate' });
		expect(button.hasAttribute('data-loading')).toBe(false);
		expect(within(button).queryByTestId('loader')).toBeNull();
	});

	it('is disabled and overlays a spinner on the label while loading', () => {
		render(
			<SidebarActionButton loading onClick={() => {}}>
				Stat Weights
			</SidebarActionButton>,
		);
		const button = screen.getByRole('button', { name: 'Stat Weights' }) as HTMLButtonElement;
		expect(button.disabled).toBe(true);
		expect(button.hasAttribute('data-loading')).toBe(true);
		expect(button.getAttribute('aria-busy')).toBe('true');
		expect(button.textContent).toContain('Stat Weights');
		expect(within(button).queryByTestId('loader')).not.toBeNull();
	});

	it('is disabled by its own prop', () => {
		render(
			<SidebarActionButton disabled onClick={() => {}}>
				Simulate
			</SidebarActionButton>,
		);
		expect((screen.getByRole('button', { name: 'Simulate' }) as HTMLButtonElement).disabled).toBe(true);
	});

	it('is disabled by the sidebar context even when its own prop is false', () => {
		render(
			<SidebarDisabledContext value>
				<SidebarActionButton disabled={false} onClick={() => {}}>
					Simulate
				</SidebarActionButton>
			</SidebarDisabledContext>,
		);
		expect((screen.getByRole('button', { name: 'Simulate' }) as HTMLButtonElement).disabled).toBe(true);
	});
});
