import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { ToastArea, toastManager } from '@ui-kit/Toast';
import { describe, expect, it, vi } from 'vitest';

import { SimToolbarView } from './SimToolbar';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const downloadLinks = () => screen.queryAllByRole('link').filter(link => link.getAttribute('aria-label') === 'info.native_sim_outdated');

describe('SimToolbarView', () => {
	it('lists the known issues and hides the entry when there are none', () => {
		const { rerender } = render(<SimToolbarView knownIssues={[]} outdatedNativeSim={false} onOpenSettings={() => {}} />);
		expect(screen.queryByTestId('known-issues')).toBeNull();

		rerender(<SimToolbarView knownIssues={['Rage decay is off']} outdatedNativeSim={false} onOpenSettings={() => {}} />);
		expect(screen.getByTestId('known-issues').textContent).toBe('info.known_issues');
	});

	it('opens the settings from the gear', () => {
		const onOpenSettings = vi.fn();
		render(<SimToolbarView knownIssues={[]} outdatedNativeSim={false} onOpenSettings={onOpenSettings} />);

		fireEvent.click(screen.getByTestId('sim-options'));

		expect(onOpenSettings).toHaveBeenCalledTimes(1);
	});

	it('links the newer native build only when the running one is outdated', () => {
		const { rerender } = render(<SimToolbarView knownIssues={[]} outdatedNativeSim={false} onOpenSettings={() => {}} />);
		expect(downloadLinks()).toHaveLength(0);

		rerender(<SimToolbarView knownIssues={[]} outdatedNativeSim onOpenSettings={() => {}} />);
		expect(downloadLinks()).toHaveLength(1);
	});

	it('offers the update in a toast that opens the update dialog', async () => {
		const update = { latest: 'v1.2.3', status: { state: 'idle' as const }, onUpdate: () => {} };
		render(
			<>
				<ToastArea manager={toastManager} />
				<SimToolbarView knownIssues={[]} outdatedNativeSim nativeSimUpdate={update} onOpenSettings={() => {}} />
			</>,
		);

		const toastText = await screen.findByText('info.native_sim_update_toast');
		const toast = toastText.parentElement!;
		await act(async () => {
			fireEvent.click(within(toast).getByRole('button', { name: 'info.native_sim_update_action' }));
		});

		expect(await screen.findByTestId('native-sim-update-dialog')).toBeTruthy();
	});

	it('closes the toast when the dialog opens from the toolbar', async () => {
		const update = { latest: 'v1.2.3', status: { state: 'idle' as const }, onUpdate: () => {} };
		render(
			<>
				<ToastArea manager={toastManager} />
				<SimToolbarView knownIssues={[]} outdatedNativeSim nativeSimUpdate={update} onOpenSettings={() => {}} />
			</>,
		);
		await screen.findByText('info.native_sim_update_toast');

		await act(async () => {
			fireEvent.click(screen.getByTestId('native-sim-update'));
		});

		expect(await screen.findByTestId('native-sim-update-dialog')).toBeTruthy();
		await waitFor(() => expect(screen.queryByText('info.native_sim_update_toast')).toBeNull());
	});
});
