import { fireEvent, render, screen, within } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { Dialog } from './Dialog';

describe('Dialog', () => {
	it('builds the header/body/footer stack', () => {
		render(
			<Dialog open onOpenChange={() => {}} size="xl" title="Encounter" footer={<button type="button">Done</button>}>
				<p>contents</p>
			</Dialog>,
		);

		const popup = screen.getByRole('dialog');
		expect(popup.getAttribute('data-testid')).toBe('sim-dialog-popup');
		expect(popup.getAttribute('data-size')).toBe('xl');
		expect(Array.from(popup.children).map(el => el.getAttribute('data-testid'))).toEqual(['sim-dialog-header', 'sim-dialog-body', 'sim-dialog-footer']);
		expect(within(popup).getByTestId('sim-dialog-body').textContent).toBe('contents');
	});

	// The default portal target is `<body>`, which is outside `.sim-ui` and so outside the spec theme
	// — measured in a browser, a `.btn-primary` in a body-portaled dialog is Bootstrap's
	// rgb(13, 110, 253) while the same markup under `.sim-ui` is the warrior brown rgb(199, 156, 110).
	// The portal container from context is how a consumer keeps the theme; that its geometry is unaffected was measured too.
	it('portals into the container from context', () => {
		const host = document.createElement('div');
		host.className = 'sim-ui';
		document.body.appendChild(host);

		render(
			<PortalContainerContext value={host}>
				<Dialog open onOpenChange={() => {}} title="Options">
					body
				</Dialog>
			</PortalContainerContext>,
		);
		expect(host.contains(screen.getByRole('dialog'))).toBe(true);

		host.remove();
	});

	it('closes on the close button and on Escape', () => {
		const onOpenChange = vi.fn();
		render(
			<Dialog open onOpenChange={onOpenChange} title="Options">
				body
			</Dialog>,
		);

		const close = screen.getByRole('button', { name: 'common.close' });
		// The tree has several <button>s inside forms with no type, so this one needs `type="button"` set explicitly.
		expect(close.getAttribute('type')).toBe('button');
		expect(close.tagName).toBe('BUTTON');
		expect(close.getAttribute('aria-label')).toBe('common.close');

		fireEvent.click(close);
		expect(onOpenChange).toHaveBeenCalledWith(false);

		onOpenChange.mockClear();
		fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' });
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it('removes the close button and refuses Escape when preventClose is set', () => {
		const onOpenChange = vi.fn();
		render(
			<Dialog open onOpenChange={onOpenChange} preventClose title="Simulating">
				body
			</Dialog>,
		);

		expect(screen.queryByRole('button', { name: 'common.close' })).toBeNull();
		fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
		expect(onOpenChange).not.toHaveBeenCalled();
		expect(screen.getByRole('dialog')).toBeTruthy();
	});

	// With `preventClose` there is nothing to put in the header at all, so there is no header.
	it('drops the header bar rather than emptying it', () => {
		render(
			<Dialog open onOpenChange={() => {}} preventClose>
				body
			</Dialog>,
		);
		expect(within(screen.getByRole('dialog')).queryByTestId('sim-dialog-header')).toBeNull();
	});

	// Base UI suppresses the backdrop of a *nested* dialog, so this only reproduces with one dialog
	// inside another — a standalone elevated dialog gets a backdrop either way and proves nothing.
	// Without it the inner dialog has nothing to dim the outer one with, and no z-index can help
	// because there is no element to raise.
	it('gives a dialog nested inside another its own raised backdrop', () => {
		render(
			<Dialog open onOpenChange={() => {}} title="Outer">
				<Dialog open onOpenChange={() => {}} elevated title="Working">
					body
				</Dialog>
			</Dialog>,
		);

		expect(screen.getAllByTestId('sim-dialog-backdrop').filter(el => el.getAttribute('data-elevated') === 'true')).toHaveLength(1);
		expect(screen.getAllByTestId('sim-dialog-viewport').filter(el => el.getAttribute('data-elevated') === 'true')).toHaveLength(1);
	});

	it('leaves a nested dialog without one when it is not elevated', () => {
		render(
			<Dialog open onOpenChange={() => {}} title="Outer">
				<Dialog open onOpenChange={() => {}} title="Working">
					body
				</Dialog>
			</Dialog>,
		);

		expect(screen.getAllByTestId('sim-dialog-backdrop').filter(el => el.getAttribute('data-elevated') === 'true')).toHaveLength(0);
	});

	it('closes only the innermost of two nested dialogs on Escape, and the outer one on a second Escape', () => {
		const outerChange = vi.fn();
		const innerChange = vi.fn();
		const NestedDialogs = () => {
			const [innerOpen, setInnerOpen] = useState(true);
			return (
				<Dialog open onOpenChange={outerChange} title="Outer">
					<Dialog
						open={innerOpen}
						onOpenChange={next => {
							innerChange(next);
							setInnerOpen(next);
						}}
						elevated
						title="Inner">
						body
					</Dialog>
				</Dialog>
			);
		};
		render(<NestedDialogs />);

		fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' });
		expect(innerChange).toHaveBeenCalledWith(false);
		expect(outerChange).not.toHaveBeenCalled();

		fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' });
		expect(outerChange).toHaveBeenCalledWith(false);
	});
});
