import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { createContext, useContext } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { createToastManager, DEFAULT_TOAST_DELAY, toastManager } from './manager';
import { ToastArea } from './ToastArea';

const standardViewport = () => screen.getAllByTestId('sim-toast-viewport').find(el => el.getAttribute('data-inline') !== 'true')!;

afterEach(() => {
	vi.useRealTimers();
});

describe('Toast', () => {
	it('renders a toast added through the module-level manager in the standard area', () => {
		render(<ToastArea manager={toastManager} />);
		act(() => {
			toastManager.add({ variant: 'success', body: 'Import successful!' });
		});

		const toast = within(standardViewport()).getByTestId('sim-toast');
		expect(toast.getAttribute('data-variant')).toBe('success');
		expect(within(toast).getByTestId('sim-toast-title').textContent).toBe('WowSims');
		expect(within(toast).getByTestId('sim-toast-body').textContent).toBe('Import successful!');
		expect(toast.getAttribute('aria-describedby')).toBe(within(toast).getByTestId('sim-toast-body').id);
	});

	it('takes the title from the options', () => {
		render(<ToastArea manager={toastManager} />);
		act(() => {
			toastManager.add({ variant: 'warning', body: 'Download it', title: 'Native sim' });
		});

		const toast = within(standardViewport()).getByTestId('sim-toast');
		expect(toast.getAttribute('data-variant')).toBe('warning');
		expect(within(toast).getByTestId('sim-toast-title').textContent).toBe('Native sim');
	});

	// `body` is block content at several sites — `NoticeNativeSim` passes a `<div>` wrapping a `<p>`
	// and an `<a class="btn">`. Base UI's `Toast.Description` is a `<p>`, and a real parser reparents
	// block content out of one; happy-dom does not, so this asserts the tag rather than the nesting.
	it('renders the body in a div so block content is valid inside it', () => {
		render(<ToastArea manager={toastManager} />);
		act(() => {
			toastManager.add({
				variant: 'info',
				body: (
					<div>
						<p>Run it natively</p>
						<a href="https://example.invalid">Download</a>
					</div>
				),
			});
		});

		const body = within(standardViewport()).getByTestId('sim-toast-body');
		expect(body.tagName).toBe('DIV');
		expect(body.querySelector('p')?.textContent).toBe('Run it natively');
	});

	it('renders the body inside the tree the area is mounted in, so the body can read context', () => {
		const Marker = createContext('not reached');
		const Body = () => <span data-testid="body-context">{useContext(Marker)}</span>;

		render(
			<Marker value="reached">
				<ToastArea manager={toastManager} />
			</Marker>,
		);
		act(() => {
			toastManager.add({ variant: 'success', body: <Body /> });
		});

		expect(screen.getByTestId('body-context').textContent).toBe('reached');
	});

	it('gives a second area its own manager and keeps the two apart', () => {
		const noticeManager = createToastManager();

		render(
			<>
				<ToastArea manager={toastManager} />
				<div data-testid="notice-host">
					<ToastArea manager={noticeManager} inline />
				</div>
			</>,
		);
		act(() => {
			toastManager.add({ variant: 'info', body: 'standard body' });
			noticeManager.add({ variant: 'info', body: 'notice body' });
		});

		const standard = standardViewport();
		const notice = within(screen.getByTestId('notice-host')).getByTestId('sim-toast-viewport');

		expect(within(standard).getByText('standard body')).toBeTruthy();
		expect(within(standard).queryByText('notice body')).toBeNull();
		expect(within(notice).getByText('notice body')).toBeTruthy();
		expect(within(notice).queryByText('standard body')).toBeNull();
	});

	it('renders an inline area in place, not through a portal, and shows and closes its toasts there', async () => {
		const manager = createToastManager();
		const { container } = render(
			<div data-testid="toast-host">
				<ToastArea manager={manager} inline testId="in-place-area" />
			</div>,
		);
		const area = screen.getByTestId('in-place-area');
		expect(within(screen.getByTestId('toast-host')).getByTestId('in-place-area')).toBe(area);
		expect(container.querySelector('[data-testid="sim-toast-portal"]')).toBeNull();

		let id = '';
		act(() => {
			id = manager.add({ variant: 'info', body: 'in place body', canClose: true });
		});
		expect(within(area).getByText('in place body')).toBeTruthy();

		act(() => manager.close(id));
		await waitFor(() => expect(within(area).queryByText('in place body')).toBeNull());
	});

	it('renders no close button when canClose is false, and closes on it when it is true', async () => {
		const manager = createToastManager();
		render(<ToastArea manager={manager} />);
		act(() => {
			manager.add({ variant: 'warning', body: 'sealed', canClose: false, autohide: false });
		});
		expect(screen.queryByRole('button', { name: 'common.close' })).toBeNull();

		act(() => {
			manager.add({ variant: 'info', body: 'closable', autohide: false });
		});
		const close = screen.getByRole('button', { name: 'common.close' });
		expect(close.getAttribute('type')).toBe('button');
		expect(close.tagName).toBe('BUTTON');
		expect(close.getAttribute('aria-label')).toBe('common.close');

		fireEvent.click(close);
		await waitFor(() => expect(screen.queryByText('closable')).toBeNull());
		expect(screen.getByText('sealed')).toBeTruthy();
	});

	it('gives the close button one glyph and a real size', () => {
		render(<ToastArea manager={toastManager} />);
		act(() => {
			toastManager.add({ variant: 'error', body: 'boom' });
		});

		const close = screen.getByRole('button', { name: 'common.close' }).querySelectorAll('svg');
		expect(close).toHaveLength(1);
		expect(close[0].getAttribute('data-icon')).toBe('xmark');
		expect(close[0].getAttribute('data-size')).toBe('lg');
		const icon2xl = within(standardViewport()).getByTestId('sim-toast-icon');
		expect(icon2xl.getAttribute('data-prefix')).toBe('fas');
		expect(icon2xl.getAttribute('data-icon')).toBe('circle-exclamation');
		expect(icon2xl.getAttribute('data-size')).toBe('2xl');
	});

	// The negative half is the control: without it the positive half passes even if no timer ever runs.
	it('auto-dismisses on the default delay but keeps a non-autohiding toast', async () => {
		vi.useFakeTimers();
		const manager = createToastManager();
		render(<ToastArea manager={manager} />);
		act(() => {
			manager.add({ variant: 'info', body: 'transient' });
			manager.add({ variant: 'info', body: 'sticky', autohide: false });
			manager.add({ variant: 'info', body: 'patient', delay: DEFAULT_TOAST_DELAY * 4 });
		});

		await act(async () => {
			await vi.advanceTimersByTimeAsync(DEFAULT_TOAST_DELAY + 100);
		});

		expect(screen.queryByText('transient')).toBeNull();
		expect(screen.getByText('sticky')).toBeTruthy();
		expect(screen.getByText('patient')).toBeTruthy();
	});

	// The viewport is portaled outside the component subtree, so React unmounting the area is what has
	// to take the toasts with it, for any caller whose cleanup runs from a ref callback.
	it('removes its toasts from the document when the area unmounts', () => {
		const manager = createToastManager();

		const { unmount } = render(<ToastArea manager={manager} />);
		act(() => {
			manager.add({ variant: 'warning', body: 'orphan', autohide: false });
		});
		expect(screen.getAllByTestId('sim-toast')).toHaveLength(1);

		unmount();

		expect(screen.queryAllByTestId('sim-toast')).toHaveLength(0);
		expect(screen.queryAllByTestId('sim-toast-viewport')).toHaveLength(0);
		expect(screen.queryAllByTestId('sim-toast-portal')).toHaveLength(0);
	});
});
