import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { useRef, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import type { ConfirmPopoverProps } from './ConfirmPopover';
import { ConfirmPopover } from './ConfirmPopover';

const renderOpen = (props: Partial<ConfirmPopoverProps> = {}) => {
	const onOpenChange = vi.fn();
	render(
		<ConfirmPopover open onOpenChange={onOpenChange} {...props}>
			Delete saved gear &apos;Pre-raid&apos;?
		</ConfirmPopover>,
	);
	return { onOpenChange, popup: screen.getByRole('dialog') };
};

describe('ConfirmPopover', () => {
	it('shows its message with a cancel and a confirm button', () => {
		const { popup } = renderOpen({ onConfirm: () => {} });

		expect(popup.getAttribute('data-testid')).toBe('sim-confirm-popover');
		expect(within(popup).getByTestId('sim-confirm-popover-message').textContent).toBe("Delete saved gear 'Pre-raid'?");
		const buttons = within(within(popup).getByTestId('sim-confirm-popover-actions')).getAllByRole('button');
		// The test run loads no locale, so a default label reads as its key.
		expect(buttons.map(button => button.textContent)).toEqual(['common.cancel', 'common.confirm']);
	});

	it('dims the page behind it unless told not to', () => {
		renderOpen({ onConfirm: () => {} });
		expect(screen.getByTestId('sim-popover-backdrop')).toBeTruthy();
	});

	it('leaves the page undimmed with backdrop off', () => {
		renderOpen({ onConfirm: () => {}, backdrop: false });
		expect(screen.queryByTestId('sim-popover-backdrop')).toBeNull();
	});

	it('confirms, then closes', () => {
		const calls: string[] = [];
		const onConfirm = vi.fn(() => calls.push('confirm'));
		const { onOpenChange } = renderOpen({ onConfirm });
		onOpenChange.mockImplementation((open: boolean) => calls.push(`open:${open}`));

		fireEvent.click(screen.getByRole('button', { name: 'common.confirm' }));

		expect(onConfirm).toHaveBeenCalledTimes(1);
		expect(calls).toEqual(['confirm', 'open:false']);
	});

	it('closes on cancel without confirming', () => {
		const onConfirm = vi.fn();
		const { onOpenChange } = renderOpen({ onConfirm });

		fireEvent.click(screen.getByRole('button', { name: 'common.cancel' }));

		expect(onOpenChange).toHaveBeenCalledWith(false);
		expect(onConfirm).not.toHaveBeenCalled();
	});

	it('offers one dismissing button when there is nothing to confirm', () => {
		const { onOpenChange, popup } = renderOpen();

		const buttons = within(popup).getAllByRole('button');
		expect(buttons.map(button => button.textContent)).toEqual(['common.ok']);
		fireEvent.click(buttons[0]);
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it('takes its confirm label and testId from the caller', () => {
		const { popup } = renderOpen({ onConfirm: () => {}, confirmLabel: 'Delete', testId: 'saved-data-delete-confirm' });

		expect(popup.getAttribute('data-testid')).toBe('saved-data-delete-confirm');
		expect(
			within(popup)
				.getAllByRole('button')
				.map(button => button.textContent),
		).toEqual(['common.cancel', 'Delete']);
	});

	it('opens from its own trigger when the caller holds the state', () => {
		const Harness = () => {
			const [open, setOpen] = useState(false);
			return (
				<ConfirmPopover open={open} onOpenChange={setOpen} trigger="Delete" onConfirm={() => {}}>
					Sure?
				</ConfirmPopover>
			);
		};
		render(<Harness />);
		expect(screen.queryByRole('dialog')).toBeNull();

		fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
		expect(within(screen.getByRole('dialog')).getByTestId('sim-confirm-popover-message').textContent).toBe('Sure?');
	});

	it('hands focus to finalFocus once it closes', async () => {
		const Harness = () => {
			const [open, setOpen] = useState(false);
			const newRef = useRef<HTMLButtonElement>(null);
			return (
				<>
					<ConfirmPopover open={open} onOpenChange={setOpen} trigger="Delete" finalFocus={newRef} onConfirm={() => {}}>
						Sure?
					</ConfirmPopover>
					<button ref={newRef} type="button">
						New
					</button>
				</>
			);
		};
		render(<Harness />);
		const trigger = screen.getByRole('button', { name: 'Delete' });
		trigger.focus();
		fireEvent.click(trigger);

		fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'common.confirm' }));
		await waitFor(() => expect(document.activeElement).toBe(screen.getByRole('button', { name: 'New' })));
	});
});
