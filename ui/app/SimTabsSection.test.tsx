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
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimHostObject } from './individual_sim_ui';
import { SimTabsPanels, SimTabsStrip } from './SimTabs';
import { SimTabsSection } from './SimTabsSection';

vi.mock('../tracking/utils', () => ({ trackPageView: () => {} }));
vi.mock('@sim/bulk/utils', async () => {
	const actual = await vi.importActual<typeof import('@sim/bulk/utils')>('@sim/bulk/utils');
	return { ...actual, getBulkPlayerCanDualWield: () => false };
});
vi.mock('@features/bulk/model/run/run', () => ({ runBulkBatch: vi.fn() }));
// The stub host has none of the raid and player getters the live settings are read from.
vi.mock('@features/settings/hooks/useCurrentSettings', () => ({ useCurrentSettings: () => null }));
// The Drops pick reads the database and the live gear, which the stub host has none of.
vi.mock('@features/bulk/hooks/useDropsSetup', () => ({ useDropsSetup: () => ({ list: { items: [] } }) }));
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

const header = () => within(document.querySelector('header')!);
const click = (element: HTMLElement) => act(() => void fireEvent.click(element));
const sideButton = (side: 'single' | 'upgrades') =>
	within(screen.getByTestId('sim-side-switch')).getByText(`sim_sides.${side}`).closest('button') as HTMLButtonElement;
const modeTabs = () => [...within(screen.getByTestId('bulk-mode-tabs')).getAllByRole('tab')].map(tab => tab.getAttribute('data-testid'));
const selectedMode = () =>
	[...within(screen.getByTestId('bulk-mode-tabs')).getAllByRole('tab')]
		.filter(tab => tab.getAttribute('aria-selected') === 'true')
		.map(tab => tab.getAttribute('data-testid'));
const openPanelId = () =>
	[...document.querySelectorAll<HTMLElement>('[role="tabpanel"]')].find(panel => !panel.hasAttribute('hidden'))?.firstElementChild?.id ?? null;
const groups = () =>
	[...document.querySelectorAll('[data-testid^="bulk-settings-group-"][data-testid$="-trigger"]')].map(trigger => trigger.getAttribute('data-testid'));

beforeEach(() => {
	localStorage.clear();
});

describe('SimTabsSection side switch', () => {
	it('shows the single sim tabs and no batch tab until Upgrades is picked', () => {
		mount({ powerTools: false });

		expect(screen.queryByTestId('bulk-mode-tabs')).toBeNull();
		expect(header().queryByTestId(BULK_TAB_ID)).toBeNull();
		expect(header().getByTestId('sim-tabs')).toBeTruthy();
		expect(sideButton('single').getAttribute('aria-pressed')).toBe('true');
		expect(sideButton('upgrades').getAttribute('aria-pressed')).toBe('false');
	});

	it('opens the Upgrades side with the mode tabs in place of the single sim tabs', async () => {
		mount({ powerTools: false });

		await click(sideButton('upgrades'));

		expect(sideButton('upgrades').getAttribute('aria-pressed')).toBe('true');
		expect(openPanelId()).toBe(BULK_TAB_ID);
		expect(screen.queryByTestId('sim-tabs')).toBeNull();
		expect(modeTabs()).toEqual(['bulk-mode-gear', 'bulk-mode-drops']);
		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger', 'bulk-settings-group-freezes-trigger']);
	});

	it('returns to the single sim tab that was open before Upgrades', async () => {
		mount({ powerTools: false });
		await click(header().getByTestId('rotation-tab'));
		expect(header().getByTestId('rotation-tab').getAttribute('aria-selected')).toBe('true');

		await click(sideButton('upgrades'));
		await click(sideButton('single'));

		expect(header().getByTestId('rotation-tab').getAttribute('aria-selected')).toBe('true');
		expect(screen.queryByTestId('bulk-mode-tabs')).toBeNull();
	});

	it('returns to the first single tab when no other was opened', async () => {
		mount({ powerTools: false });

		await click(sideButton('upgrades'));
		await click(sideButton('single'));

		expect(header().getByTestId('gear-tab').getAttribute('aria-selected')).toBe('true');
	});
});

describe('SimTabsSection Upgrades modes', () => {
	it('marks Gear as the mode in a stored Builds with power tools off', async () => {
		const player = mount({ powerTools: false });
		act(() => patchBulkState(player, { mode: 'builds' }));
		await click(sideButton('upgrades'));

		expect(selectedMode()).toEqual(['bulk-mode-gear']);
	});

	it('switches to Drops from its tab without power tools', async () => {
		const player = mount({ powerTools: false });
		await click(sideButton('upgrades'));

		await click(screen.getByTestId('bulk-mode-drops'));

		expect(bulkState(player).mode).toBe('drops');
		expect(loadStoredBulkMode(player)).toBe('drops');
		expect(selectedMode()).toEqual(['bulk-mode-drops']);
		expect(screen.getByTestId('bulk-run-card')).toBeTruthy();
		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger']);
	});

	it('offers Builds with power tools on and opens the run card alone from it', async () => {
		const player = mount({ powerTools: true });
		await click(sideButton('upgrades'));
		expect(modeTabs()).toEqual(['bulk-mode-gear', 'bulk-mode-drops', 'bulk-mode-builds']);

		await click(screen.getByTestId('bulk-mode-builds'));

		expect(bulkState(player).mode).toBe('builds');
		expect(loadStoredBulkMode(player)).toBe('builds');
		expect(selectedMode()).toEqual(['bulk-mode-builds']);
		expect(groups()).toEqual([]);
		expect(screen.getByTestId('bulk-run-card')).toBeTruthy();
	});

	it('keeps the stored mode with power tools off, and runs Gear until they are back on', async () => {
		const player = mount({ powerTools: false });
		act(() => patchBulkState(player, { mode: 'builds' }));
		await click(sideButton('upgrades'));

		expect(modeTabs()).toEqual(['bulk-mode-gear', 'bulk-mode-drops']);
		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger', 'bulk-settings-group-freezes-trigger']);

		act(() => patchSlice(player.sim.store, 'ui', { showPowerTools: true }));

		expect(selectedMode()).toEqual(['bulk-mode-builds']);
		expect(groups()).toEqual([]);
	});
});
