import { BULK_TAB_ID } from '@features/hooks/useActivateTab';
import { APLRotation } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { bulkState, loadStoredBulkMode, patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, patchSlice, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { memoryEnv } from '@testing/sim/memory_env';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimHostObject } from './individual_sim_ui';
import { SimTabsPanels, SimTabsStrip } from './SimTabs';
import { SimTabsSection } from './SimTabsSection';

vi.mock('../tracking/utils', () => ({ trackPageView: () => {} }));
vi.mock('@sim/bulk/utils', async () => {
	const actual = await vi.importActual<typeof import('@sim/bulk/utils')>('@sim/bulk/utils');
	return { ...actual, getBulkPlayerCanDualWield: () => false };
});
vi.mock('@features/bulk/model/run', () => ({ runBulkBatch: vi.fn() }));
// The stub host has none of the raid and player getters the live settings are read from.
vi.mock('@features/settings/hooks/useCurrentSettings', () => ({ useCurrentSettings: () => null }));
vi.mock('@features/results/components/DetailedResults', () => ({ DetailedResults: () => null }));
vi.mock('./tabs/GearTabBody', () => ({ GearTabBody: () => null }));
vi.mock('./tabs/SettingsTabBody', () => ({ SettingsTabBody: () => null }));
vi.mock('./tabs/TalentsTabBody', () => ({ TalentsTabBody: () => null }));
vi.mock('./tabs/RotationTabPane', () => ({ RotationTabPane: () => null }));
// The batch pane is its settings sidebar alone, where the mode shows, under the Builds plan the pane provides.
vi.mock('./tabs/BulkTabBody', async () => {
	const { BulkSettings } = await import('@features/bulk/components/BulkSettings');
	const { BulkBuildsProvider } = await import('@features/bulk/context/BulkBuildsContext');
	return {
		BulkTabBody: () => (
			<BulkBuildsProvider>
				<BulkSettings />
			</BulkBuildsProvider>
		),
	};
});

const STORE_KEY = 3;

const mount = ({ powerTools }: { powerTools: boolean }) => {
	const store = createSimStore();
	patchSlice(store, 'ui', { showPowerTools: powerTools });
	const gear = new Gear({});
	seedKeyed(store, 'players', STORE_KEY, { gear, bonusStats: new Stats(), talentsString: '', v: zeroVersions(PLAYER_FIELDS) } as never);
	const player = {
		sim: { store, isNative: false, env: memoryEnv(), waitForInit: () => Promise.resolve() },
		storeKey: STORE_KEY,
		getGear: () => gear,
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		aplRotation: APLRotation.create(),
		getResolvedAplRotationForBuild: ({ rotation }: { rotation: APLRotation }) => rotation,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	const host = {
		player,
		sim: player.sim,
		simDisabled: false,
		reforger: null,
		reforgeOptions: null,
		individualConfig: { presets: { gear: [], encounters: [], talents: [], rotations: [] } },
		getStorageKey: (key: string) => `test-${key}`,
		getSavedGearStorageKey: () => 'test-saved-gear',
		getSavedEncounterStorageKey: () => 'test-saved-encounters',
		getSavedSettingsStorageKey: () => 'test-saved-settings',
		getSavedTalentsStorageKey: () => 'test-saved-talents',
		getSavedRotationStorageKey: () => 'test-saved-rotations',
	} as unknown as SimHostObject<any>;
	render(
		<SimHostProvider host={host}>
			<SimTabsSection host={host}>
				<header>
					<SimTabsStrip />
				</header>
				<main>
					<SimTabsPanels />
				</main>
			</SimTabsSection>
		</SimHostProvider>,
	);
	return player;
};

const batchTab = () => within(document.querySelector('header')!).getByTestId(BULK_TAB_ID);
const click = (element: HTMLElement) => act(() => void fireEvent.click(element));
const nextFrame = () => new Promise(resolve => requestAnimationFrame(resolve));
// A whole press, one frame long, since Base UI opens a pressed menu on the next frame.
const press = (element: HTMLElement, pointerType: 'mouse' | 'touch') =>
	act(async () => {
		fireEvent.pointerDown(element, { pointerType, button: 0 });
		fireEvent.mouseDown(element, { button: 0 });
		fireEvent.pointerUp(element, { pointerType, button: 0 });
		fireEvent.mouseUp(element, { button: 0 });
		fireEvent.click(element, { button: 0 });
		await nextFrame();
	});
const hover = (element: HTMLElement) =>
	act(async () => {
		fireEvent.pointerEnter(element, { pointerType: 'mouse' });
		fireEvent.mouseEnter(element);
		fireEvent.mouseMove(element);
		await nextFrame();
	});
const leave = (element: HTMLElement) =>
	act(async () => {
		fireEvent.pointerLeave(element, { pointerType: 'mouse' });
		fireEvent.mouseLeave(element);
		await nextFrame();
	});
const modeItems = () => screen.queryAllByRole('menuitem').map(item => item.getAttribute('data-testid'));
const groups = () =>
	[...document.querySelectorAll('[data-testid^="bulk-settings-group-"][data-testid$="-trigger"]')].map(trigger => trigger.getAttribute('data-testid'));

beforeEach(() => {
	localStorage.clear();
});

describe('SimTabsSection batch tab', () => {
	it('is a plain Batch tab with no mode menu without power tools', async () => {
		mount({ powerTools: false });

		expect(batchTab().textContent).toBe('bulk_tab.title');
		expect(batchTab().hasAttribute('aria-haspopup')).toBe(false);

		await press(batchTab(), 'mouse');

		expect(modeItems()).toEqual([]);
		expect(batchTab().getAttribute('aria-selected')).toBe('true');
		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger', 'bulk-settings-group-freezes-trigger']);
	});

	it('opens the batch tab in its mode on a mouse click with power tools on, leaving no menu open', async () => {
		const player = mount({ powerTools: true });
		act(() => patchBulkState(player, { mode: 'builds' }));

		await hover(batchTab());
		await press(batchTab(), 'mouse');
		expect(modeItems()).toEqual(['bulk-mode-gear', 'bulk-mode-builds']);
		await leave(batchTab());

		expect(modeItems()).toEqual([]);
		expect(bulkState(player).mode).toBe('builds');
		expect(batchTab().getAttribute('aria-selected')).toBe('true');
		expect(groups()).toEqual([]);
	});

	it('opens the batch tab and the menu on a tap with power tools on', async () => {
		mount({ powerTools: true });

		await press(batchTab(), 'touch');

		expect(modeItems()).toEqual(['bulk-mode-gear', 'bulk-mode-builds']);
		expect(batchTab().getAttribute('aria-selected')).toBe('true');
	});

	it.each([
		['Enter', '{Enter}'],
		['Space', ' '],
		['ArrowDown', '{ArrowDown}'],
	])('opens the menu from the keyboard with %s', async (_name, key) => {
		mount({ powerTools: true });
		act(() => batchTab().focus());

		await userEvent.keyboard(key);

		expect(modeItems()).toEqual(['bulk-mode-gear', 'bulk-mode-builds']);
		expect(batchTab().getAttribute('aria-selected')).toBe('true');
	});

	it('moves along the tab strip with ArrowLeft and ArrowRight onto and off the batch tab', async () => {
		mount({ powerTools: true });
		const gear = within(document.querySelector('header')!).getByTestId('gear-tab');

		// The strip loops: left of the first tab is the batch tab, the last one.
		fireEvent.keyDown(gear, { key: 'ArrowLeft' });
		await waitFor(() => expect(document.activeElement).toBe(batchTab()));
		expect(batchTab().getAttribute('aria-selected')).toBe('true');

		fireEvent.keyDown(batchTab(), { key: 'ArrowRight' });
		await waitFor(() => expect(document.activeElement).toBe(gear));
		expect(gear.getAttribute('aria-selected')).toBe('true');
	});

	it('reads Batch on a tab that carries the menu with power tools on', () => {
		mount({ powerTools: true });

		expect(batchTab().textContent).toBe('bulk_tab.title');
		expect(batchTab().getAttribute('role')).toBe('tab');
		expect(batchTab().getAttribute('aria-haspopup')).toBe('menu');
	});

	it('switches to Builds from the hover menu, opening the batch tab with the run card alone', async () => {
		const player = mount({ powerTools: true });
		await hover(batchTab());
		expect(modeItems()).toEqual(['bulk-mode-gear', 'bulk-mode-builds']);
		expect(batchTab().getAttribute('aria-selected')).toBe('false');

		await click(screen.getByTestId('bulk-mode-builds'));

		expect(bulkState(player).mode).toBe('builds');
		expect(loadStoredBulkMode(player)).toBe('builds');
		expect(batchTab().textContent).toBe('bulk_tab.title');
		expect(batchTab().getAttribute('aria-selected')).toBe('true');
		expect(groups()).toEqual([]);
		expect(screen.getByTestId('bulk-run-card')).toBeTruthy();
	});

	it('opens the batch tab in the picked mode while another tab holds focus, and focuses the batch tab', async () => {
		const player = mount({ powerTools: true });
		const rotation = within(document.querySelector('header')!).getByTestId('rotation-tab');
		await userEvent.click(rotation);
		expect(document.activeElement).toBe(rotation);
		expect(rotation.getAttribute('aria-selected')).toBe('true');

		await hover(batchTab());
		await act(async () => {
			fireEvent.click(screen.getByTestId('bulk-mode-builds'));
			await nextFrame();
		});

		expect(modeItems()).toEqual([]);
		expect(bulkState(player).mode).toBe('builds');
		expect(document.activeElement).toBe(batchTab());
		expect(batchTab().getAttribute('aria-selected')).toBe('true');
		expect(rotation.getAttribute('aria-selected')).toBe('false');
		expect(screen.getByTestId('bulk-run-card')).toBeTruthy();
	});

	it('leaves the focused tab open when the menu closes on Escape without a pick', async () => {
		const player = mount({ powerTools: true });
		const rotation = within(document.querySelector('header')!).getByTestId('rotation-tab');
		await userEvent.click(rotation);

		await hover(batchTab());
		expect(modeItems()).toEqual(['bulk-mode-gear', 'bulk-mode-builds']);
		await userEvent.keyboard('{Escape}');
		await act(nextFrame);

		expect(modeItems()).toEqual([]);
		expect(bulkState(player).mode).toBe('gear');
		expect(rotation.getAttribute('aria-selected')).toBe('true');
		expect(batchTab().getAttribute('aria-selected')).toBe('false');
	});

	it('keeps the stored mode with power tools off, and runs Gear until they are back on', () => {
		const player = mount({ powerTools: false });
		act(() => patchBulkState(player, { mode: 'builds' }));

		expect(batchTab().textContent).toBe('bulk_tab.title');
		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger', 'bulk-settings-group-freezes-trigger']);

		act(() => patchSlice(player.sim.store, 'ui', { showPowerTools: true }));

		expect(batchTab().textContent).toBe('bulk_tab.title');
		expect(groups()).toEqual([]);
	});
});
