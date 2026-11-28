import { TabActivationContext } from '@features/hooks/useActivateTab';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimHostObject } from './individual_sim_ui';
import { SimDock } from './SimDock';
import { UpgradesSlotsProvider, useUpgradesSlotRefs } from './UpgradesSlots';

// The shell's Upgrades run panel slot, which the Upgrades tab body portals its settings into.
const SlotPanel = () => <div ref={useUpgradesSlotRefs().panel} data-testid="upgrades-slot-panel" />;
const UpgradesPanelSlot = ({ children }: { children: ReactNode }) => (
	<UpgradesSlotsProvider>
		<SlotPanel />
		{children}
	</UpgradesSlotsProvider>
);

vi.mock('@features/results/components/IterationsPicker', () => ({ IterationsPicker: () => <div data-testid="iterations-picker" /> }));
vi.mock('@features/results/components/SimulateAction', () => ({ SimulateAction: () => <button data-testid="dps-action" /> }));
vi.mock('@features/stat-weights/components/StatWeightsAction', () => ({ StatWeightsAction: () => <button data-testid="ep-weights-action" /> }));
vi.mock('./SidebarActions', () => ({ SidebarActions: () => <div data-testid="registry-actions" /> }));
vi.mock('./NoticeNativeSim', () => ({ NoticeNativeSim: () => null }));
vi.mock('./ReleaseInfo', () => ({ SpecReleaseInfo: () => <div data-testid="release-info" /> }));
const side = vi.hoisted(() => ({ current: 'single' as 'single' | 'upgrades' }));
vi.mock('./SimTabs', () => ({ SimTabsStrip: () => <div data-testid="sim-tabs" />, useSimSide: () => side.current }));
vi.mock('./SimPresetsDialog', () => ({ SimPresetsDialog: () => null }));
vi.mock('@features/character-stats', () => ({ CharacterStats: () => <div data-testid="character-stats" /> }));
vi.mock('@sim/hooks/useSimStatus', () => ({ useSimStatus: () => ({ status: 'ready', error: null }) }));
vi.mock('@features/results/components/SimResultsPanel', () => ({
	SimResultsPanel: () => <div data-testid="results-viewer" />,
	SimResultReadout: () => <span data-testid="sim-result-readout" />,
	SimWarnings: () => <div data-testid="warning-zone" />,
}));

const host = { disabled: false, simDisabled: false, warnings: {} } as SimHostObject<any>;

beforeEach(() => {
	side.current = 'single';
});

const isHidden = (testId: string) => screen.getByTestId(testId).classList.contains('hidden');

const dock = () => (
	<TabActivationContext value={vi.fn()}>
		<UpgradesPanelSlot>
			<SimDock host={host} sim={{} as never} spec={{} as never} simReady />
		</UpgradesPanelSlot>
	</TabActivationContext>
);

const renderDock = () => render(dock());

const press = (element: Element) => {
	fireEvent.pointerDown(element);
	fireEvent.mouseDown(element);
	fireEvent.pointerUp(element);
	fireEvent.mouseUp(element);
	fireEvent.click(element);
};

describe('SimDock', () => {
	it('keeps the run actions and the tabs on the bar, and the rest of the sidebar in the sheet', () => {
		renderDock();
		expect(screen.getByTestId('sim-dock').contains(screen.getByTestId('dps-action'))).toBe(true);
		expect(screen.getByTestId('sim-dock').contains(screen.getByTestId('sim-tabs'))).toBe(true);
		expect(screen.queryByTestId('sim-dock-sheet')).toBeNull();

		fireEvent.click(screen.getByTestId('sim-dock-toggle'));

		const sheet = screen.getByTestId('sim-dock-sheet');
		for (const part of ['iterations-picker', 'results-viewer', 'warning-zone', 'character-stats', 'release-info']) {
			expect(sheet.contains(screen.getByTestId(part))).toBe(true);
		}
		expect(screen.getAllByTestId('dps-action')).toHaveLength(1);
	});

	it('points the chevron down while the sheet is open', () => {
		renderDock();
		const toggle = screen.getByTestId('sim-dock-toggle');
		const chevron = () => toggle.querySelector('[data-icon]')?.getAttribute('data-icon');
		expect(chevron()).toBe('chevron-up');

		fireEvent.click(toggle);
		expect(chevron()).toBe('chevron-down');

		fireEvent.click(toggle);
		expect(chevron()).toBe('chevron-up');
	});

	it('keeps the sheet open through a press on the bar, and closes it on a press elsewhere', () => {
		renderDock();
		fireEvent.click(screen.getByTestId('sim-dock-toggle'));

		const toggle = screen.getByTestId('sim-dock-toggle');
		press(screen.getByTestId('dps-action'));
		expect(toggle.getAttribute('aria-expanded')).toBe('true');

		press(screen.getByTestId('sim-tabs'));
		expect(toggle.getAttribute('aria-expanded')).toBe('false');
	});

	it('shows the single sim bar on the single side and hides the Upgrades bar', () => {
		renderDock();
		expect(isHidden('sim-dock-single-bar')).toBe(false);
		expect(isHidden('sim-dock-upgrades-bar')).toBe(true);
	});

	it('shows the Upgrades bar on the Upgrades side and hides the single sim bar', () => {
		side.current = 'upgrades';
		renderDock();
		expect(isHidden('sim-dock-single-bar')).toBe(true);
		expect(isHidden('sim-dock-upgrades-bar')).toBe(false);
	});

	it('closes the single sheet when the Upgrades side opens', () => {
		const { rerender } = renderDock();
		fireEvent.click(screen.getByTestId('sim-dock-toggle'));
		expect(screen.getByTestId('sim-dock-toggle').getAttribute('aria-expanded')).toBe('true');

		side.current = 'upgrades';
		rerender(dock());

		expect(screen.getByTestId('sim-dock-toggle').getAttribute('aria-expanded')).toBe('false');
	});

	it('keeps the Upgrades sheet mounted while closed, for the panel the tab body portals into it', () => {
		side.current = 'upgrades';
		renderDock();
		expect(screen.getByTestId('sim-dock-upgrades-sheet')).toBeTruthy();
		expect(screen.getByTestId('sim-dock-upgrades-toggle').getAttribute('aria-expanded')).toBe('false');
	});
});
