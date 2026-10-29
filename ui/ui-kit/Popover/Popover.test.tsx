import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { Button } from '@ui-kit/Button';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { describe, expect, it, vi } from 'vitest';

import type { PopoverProps } from './Popover';
import { Popover } from './Popover';

// `outsidePressEvent.mouse` is `'intentional'` for a non-modal popover: the press is the whole
// sequence, and only the closing `click` dismisses.
const pressOutside = (node: Element) => {
	fireEvent.pointerDown(node);
	fireEvent.mouseDown(node);
	fireEvent.pointerUp(node);
	fireEvent.mouseUp(node);
	fireEvent.click(node);
};

const renderPopover = (props: Partial<PopoverProps> = {}) =>
	render(
		<Popover trigger="Settings" {...props}>
			<button type="button">Reset</button>
		</Popover>,
	);

const openPopover = (props: Partial<PopoverProps> = {}, triggerName = 'Settings') => {
	const result = renderPopover(props);
	const trigger = screen.getByRole('button', { name: triggerName });
	// A real press focuses the button it lands on; `fireEvent.click` does not.
	trigger.focus();
	fireEvent.click(trigger);
	return { ...result, trigger };
};

describe('Popover', () => {
	it('portals into the container it is given', () => {
		const host = document.createElement('div');
		host.className = 'sim-ui';
		document.body.appendChild(host);

		openPopover({ container: host });
		expect(host.contains(screen.getByRole('dialog'))).toBe(true);

		host.remove();
	});

	it('portals into the container from context when it is given none', () => {
		const host = document.createElement('div');
		document.body.appendChild(host);

		render(
			<PortalContainerContext value={host}>
				<Popover trigger="Settings">
					<button type="button">Reset</button>
				</Popover>
			</PortalContainerContext>,
		);
		fireEvent.click(screen.getByRole('button', { name: 'Settings' }));
		expect(host.contains(screen.getByRole('dialog'))).toBe(true);

		host.remove();
	});

	it('forwards trigger props, so an icon-only trigger can be named and anchor a tooltip', () => {
		renderPopover({ triggerProps: { 'aria-label': 'Reforge settings', 'data-tooltip-id': 'reforge-settings-tip' } });

		const trigger = screen.getByRole('button', { name: 'Reforge settings' });
		expect(trigger.getAttribute('data-tooltip-id')).toBe('reforge-settings-tip');
		// A trigger inside a form must not submit it.
		expect(trigger.getAttribute('type')).toBe('button');
	});

	// The pickers this popover is for commit on their own events, so it leaves focus on the trigger.
	// Escape still closes, because the dismiss listener is on the document rather than on the popup.
	it('leaves focus on the trigger, and still closes on Escape', () => {
		const onOpenChange = vi.fn();
		const { trigger } = openPopover({ onOpenChange });
		expect(document.activeElement).toBe(trigger);

		onOpenChange.mockClear();
		fireEvent.keyDown(trigger, { key: 'Escape' });
		expect(onOpenChange).toHaveBeenCalledWith(false);
		expect(screen.getByRole('dialog').hasAttribute('data-open')).toBe(false);
	});

	it('focuses the first tabbable element in finalFocus on close, not the trigger', async () => {
		const section = document.createElement('div');
		section.innerHTML = '<h6>Picks</h6><button type="button">First pick</button>';
		document.body.appendChild(section);
		const { trigger } = openPopover({ finalFocus: { current: section } });

		fireEvent.keyDown(trigger, { key: 'Escape' });
		await waitFor(() => expect(document.activeElement).toBe(within(section).getByRole('button', { name: 'First pick' })));
		section.remove();
	});

	it('names the popup after its title, set as a heading above the contents', () => {
		openPopover({ title: 'Reforge settings' });

		const popup = screen.getByRole('dialog', { name: 'Reforge settings' });
		const heading = within(popup).getByRole('heading', { name: 'Reforge settings' });
		expect(popup.firstElementChild).toBe(heading);
	});

	it('renders no title unless given one', () => {
		openPopover();
		expect(screen.queryByRole('heading')).toBeNull();
		expect(screen.getByRole('dialog').hasAttribute('aria-labelledby')).toBe(false);
	});

	it('renders no backdrop unless asked for one', () => {
		openPopover();
		expect(screen.queryByTestId('sim-popover-backdrop')).toBeNull();
	});

	it('dims the page behind the popup when asked, and closes on a press on the backdrop', () => {
		const onOpenChange = vi.fn();
		openPopover({ backdrop: true, onOpenChange });
		onOpenChange.mockClear();

		const backdrop = screen.getByTestId('sim-popover-backdrop');
		expect(backdrop.contains(screen.getByRole('dialog'))).toBe(false);

		pressOutside(backdrop);
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	// The bug class this pins: a popup whose portal node outlives the React tree that opened it. The
	// portal is a body child no ancestor unmount reaches, so nothing but Base UI's own cleanup removes it.
	it.each([
		['', false],
		[' from a container', true],
	])('takes its portal node with it when it unmounts%s', (_, inContainer) => {
		const host = document.createElement('div');
		document.body.appendChild(host);
		const root = inContainer ? host : document.body;

		const { unmount } = openPopover(inContainer ? { container: host } : {});
		expect(within(root).getAllByTestId('sim-popover-portal')).toHaveLength(1);

		unmount();
		expect(within(root).queryByTestId('sim-popover-portal')).toBeNull();
		expect(screen.queryByRole('dialog')).toBeNull();
		if (inContainer) expect(host.innerHTML).toBe('');
		host.remove();
	});

	it('renders the trigger as the kit Button it is given, and opens from it', () => {
		const error = vi.spyOn(console, 'error');
		const onOpenChange = vi.fn();
		const { trigger } = openPopover(
			{
				onOpenChange,
				triggerRender: <Button variant="primary" />,
				triggerTestId: 'settings-trigger',
				triggerProps: { 'aria-label': 'Reforge settings' },
				trigger: <span>Cog</span>,
			},
			'Reforge settings',
		);

		expect(trigger).toBe(screen.getByTestId('settings-trigger'));
		expect(trigger.tagName).toBe('BUTTON');
		expect(trigger.textContent).toBe('Cog');
		expect(onOpenChange).toHaveBeenCalledWith(true);
		expect(trigger.hasAttribute('data-popup-open')).toBe(true);
		expect(error).not.toHaveBeenCalled();
		error.mockRestore();
	});

	it('renders a kit Button trigger with an href as an anchor with an explicit role="button"', () => {
		const { trigger } = openPopover({ triggerRender: <Button href="https://x" />, triggerTestId: 'settings-trigger', trigger: <span>Cog</span> }, 'Cog');

		expect(trigger.tagName).toBe('A');
		expect(trigger.getAttribute('role')).toBe('button');
	});
});
