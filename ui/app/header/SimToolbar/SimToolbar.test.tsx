import { fireEvent, render, screen } from '@testing-library/react';
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

	it('confirms before updating and shows the progress', () => {
		const onUpdate = vi.fn();
		const nativeSimUpdate = { latest: 'v1.2.3', onUpdate };
		const { rerender } = render(
			<SimToolbarView knownIssues={[]} outdatedNativeSim nativeSimUpdate={{ ...nativeSimUpdate, status: { state: 'idle' } }} onOpenSettings={() => {}} />,
		);
		expect(downloadLinks()).toHaveLength(0);

		fireEvent.click(screen.getByTestId('native-sim-update'));
		expect(onUpdate).not.toHaveBeenCalled();
		fireEvent.click(screen.getByText('info.native_sim_update_action'));
		expect(onUpdate).toHaveBeenCalledTimes(1);

		rerender(
			<SimToolbarView
				knownIssues={[]}
				outdatedNativeSim
				nativeSimUpdate={{ ...nativeSimUpdate, status: { state: 'updating' } }}
				onOpenSettings={() => {}}
			/>,
		);
		expect(screen.getByTestId('native-sim-updating')).toBeTruthy();
	});
});
