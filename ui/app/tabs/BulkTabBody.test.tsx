import type { BulkEncounterResults, BulkRankedResult, TopGearResult } from '@sim/bulk/types';
import { REPO_RELEASES_URL } from '@sim/constants/other';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { CURRENT_ROTATION_SETUP, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BulkTabActionsView, type BulkTabActionsViewProps, BulkTabBody } from './BulkTabBody';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => false }));
vi.mock('@features/bulk/components/BulkItemSearch', () => ({ BulkItemSearch: () => null }));
vi.mock('@features/bulk/components/BulkPickerGroups', () => ({ BulkPickerGroups: () => null }));
vi.mock('@features/bulk/components/BulkProgress', () => ({ BulkProgressDialog: () => null }));
vi.mock('@features/bulk/components/BulkSettings', () => ({ BulkSettings: () => <div data-testid="bulk-settings-stub" /> }));
vi.mock('@features/bulk/components/BulkResults/BulkResultRow', () => ({
	BulkResultRow: ({ result }: { result: TopGearResult }) => <div data-testid="row" data-avg={result.dpsMetrics.avg} />,
}));
vi.mock('@features/gear/components/SelectorModal', () => ({ SelectorModal: () => null }));
vi.mock('@features/import-export', () => ({ BulkGearImporterDialog: () => null }));
const downloadString = vi.hoisted(() => vi.fn());
vi.mock('@ui-kit/utils/dom', async importOriginal => ({ ...(await importOriginal<typeof import('@ui-kit/utils/dom')>()), downloadString }));

const props = (overrides: Partial<BulkTabActionsViewProps> = {}): BulkTabActionsViewProps => ({
	showNativeDownload: false,
	onImportBags: vi.fn(),
	onImportFavorites: vi.fn(),
	onClear: vi.fn(),
	...overrides,
});

describe('BulkTabActionsView', () => {
	it('links the native download only when asked to', () => {
		const { rerender } = render(<BulkTabActionsView {...props()} />);
		expect(screen.queryByRole('link')).toBeNull();

		rerender(<BulkTabActionsView {...props({ showNativeDownload: true })} />);
		expect(screen.getByRole('link', { name: 'bulk_tab.download_native' }).getAttribute('href')).toBe(REPO_RELEASES_URL);
	});

	it('reports each action', () => {
		const handlers = props();
		render(<BulkTabActionsView {...handlers} />);

		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.actions.import_bags' }));
		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.actions.import_favorites' }));
		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.actions.clear_items' }));

		expect(handlers.onImportBags).toHaveBeenCalledTimes(1);
		expect(handlers.onImportFavorites).toHaveBeenCalledTimes(1);
		expect(handlers.onClear).toHaveBeenCalledTimes(1);
	});
});

describe('BulkTabBody', () => {
	const STORE_KEY = 3;
	const BASELINE_DPS = 1000;

	// A gear set with no slots, equal only to itself: enough for the CSV export.
	const gear = () => {
		const set = { getItemSlots: () => [], getEquippedItem: () => null, equals: (other: unknown) => other === set };
		return set;
	};
	const result = (avg: number) =>
		({
			gear: gear(),
			dpsMetrics: { avg, stdev: 1 },
			talents: CURRENT_TALENT_SETUP,
			rotation: CURRENT_ROTATION_SETUP,
			gearSetup: null,
			settings: null,
		}) as unknown as BulkRankedResult;

	// Each encounter's best set beats the current gear by `best - BASELINE_DPS`; a null length is a health-based encounter.
	const entry = (name: string, length: [duration: number, variation: number] | null, best: number, isCurrent = false): BulkEncounterResults => {
		const baseline = result(BASELINE_DPS);
		return {
			name,
			isCurrent,
			duration: length?.[0] ?? null,
			durationVariation: length?.[1] ?? null,
			results: {
				chains: [[result(best)], [baseline]],
				originalGearResults: baseline,
				referenceIsCurrent: true,
				currentGear: baseline.gear,
				iterations: 100,
			},
		};
	};

	const mount = (entries: BulkEncounterResults[] | null) => {
		const store = createSimStore();
		const player = { sim: { store, isNative: false }, storeKey: STORE_KEY } as unknown as Player<any>;
		seedBulkSettings(player);
		patchBulkState(player, { started: entries !== null, results: entries });
		const host = { player, sim: player.sim } as never;
		render(
			<SimHostProvider host={host}>
				<BulkTabBody />
			</SimHostProvider>,
		);
		return player;
	};

	const sidebar = () => screen.queryByTestId('bulk-encounter-results');
	const sidebarRows = () =>
		within(sidebar()!)
			.getAllByRole('button')
			.filter(button => button.hasAttribute('aria-pressed'));
	const exportButton = () => screen.getByTestId('bulk-results-export-csv');
	const shownBest = () => screen.getAllByTestId('row')[0].getAttribute('data-avg');
	const settings = () => screen.getByTestId('bulk-settings-stub');
	// Both panes stay mounted. A pane that stops being active turns inert at once and gains `hidden` only when its fade-out
	// ends, which never happens here, so `inert` marks the pane that is going away.
	const paneOf = (element: Element) => element.closest('[data-testid="tab-pane"]')!;
	const isShown = (element: Element) => !element.closest('[hidden], [inert]');
	const THREE = [entry('', [180, 5], 1100, true), entry('Patchwerk', [300, 30], 1036.5), entry('Council', null, 1250)];

	it('shows the batch settings inside the Setup pane and no encounter list', () => {
		mount(null);

		expect(paneOf(settings()).id).toBe('bulkSetupTab');
		expect(isShown(settings())).toBe(true);
		expect(sidebar()).toBeNull();
	});

	it('lists one row per finished encounter inside the Results pane, current first, with the best DPS and its gain', () => {
		mount(THREE);

		expect(isShown(settings())).toBe(false);
		expect(paneOf(sidebar()!).id).toBe('bulkResultsTab');
		expect(isShown(sidebar()!)).toBe(true);
		expect(sidebarRows().map(row => row.textContent)).toEqual([
			'bulk_tab.results.current 3:00 ±5s1,100.00 +100.00 (10.00%)',
			'Patchwerk 5:00 ±30s1,036.50 +36.50 (3.65%)',
			'Council bulk_tab.settings.encounters.health_summary1,250.00 +250.00 (25.00%)',
		]);
		expect(sidebarRows().map(row => row.getAttribute('aria-pressed'))).toEqual(['true', 'false', 'false']);
		expect(shownBest()).toBe('1100');
	});

	it('shows only the DPS for an encounter where the current gear stays best', () => {
		mount([entry('', [180, 5], 1100, true), entry('Patchwerk', [300, 30], BASELINE_DPS)]);

		expect(sidebarRows()[1].textContent).toBe('Patchwerk 5:00 ±30s1,000.00');
	});

	it('shows each encounter’s best DPS with no gain, and says so, when the batch did not sim the current setup', () => {
		const againstBest = (shown: BulkEncounterResults): BulkEncounterResults => ({
			...shown,
			results: { ...shown.results, originalGearResults: shown.results.chains[0][0], referenceIsCurrent: false },
		});
		mount([againstBest(entry('', [180, 5], 1100, true)), againstBest(entry('Patchwerk', [300, 30], 1036.5))]);

		expect(sidebarRows().map(row => row.textContent)).toEqual(['bulk_tab.results.current 3:00 ±5s1,100.00', 'Patchwerk 5:00 ±30s1,036.50']);
		expect(within(sidebar()!).getByText('bulk_tab.results.encounters_caption_best')).toBeTruthy();
	});

	it('shows the ranking of the encounter picked in the list', () => {
		mount(THREE);

		fireEvent.click(sidebarRows()[2]);

		expect(sidebarRows().map(row => row.getAttribute('aria-pressed'))).toEqual(['false', 'false', 'true']);
		expect(shownBest()).toBe('1250');
	});

	it('starts a new run on its first encounter again', () => {
		const player = mount(THREE);
		fireEvent.click(sidebarRows()[1]);

		act(() => patchBulkState(player, { results: THREE.slice() }));

		expect(sidebarRows().map(row => row.getAttribute('aria-pressed'))).toEqual(['true', 'false', 'false']);
		expect(shownBest()).toBe('1100');
	});

	it('lists a run with one finished result too, with the export and without Compare encounters', () => {
		const player = mount([entry('', [180, 5], 1100, true)]);
		act(() => patchSlice(player.sim.store, 'ui', { showPowerTools: true }));

		expect(sidebarRows()).toHaveLength(1);
		expect(within(sidebar()!).getByTestId('bulk-results-export-csv')).toBeTruthy();
		expect(within(sidebar()!).queryByText('bulk_tab.compare.button')).toBeNull();
		expect(isShown(settings())).toBe(false);
		expect(shownBest()).toBe('1100');
	});

	it('shows the settings again, and hides the encounter list, when Setup is picked', () => {
		mount(THREE);

		fireEvent.click(screen.getByRole('tab', { name: 'bulk_tab.tabs.setup' }));

		expect(isShown(settings())).toBe(true);
		expect(isShown(sidebar()!)).toBe(false);
	});

	it('exports every encounter as CSV from the bottom of the encounter list', () => {
		mount(THREE);

		expect(sidebar()!.contains(exportButton())).toBe(true);
		fireEvent.click(exportButton());

		expect(downloadString).toHaveBeenCalledOnce();
		const [csv, fileName, mimeType] = downloadString.mock.calls[0];
		expect([fileName, mimeType]).toEqual(['wowsims-batch-results.csv', 'text/csv']);
		expect(csv.trimEnd().split('\n')).toHaveLength(1 + 3 * 2);
	});

	it('puts the CSV export above the ranking when there is no encounter list', () => {
		mount([entry('', [180, 5], 1100, true)]);

		expect(paneOf(exportButton()).id).toBe('bulkResultsTab');
		fireEvent.click(exportButton());

		expect(downloadString).toHaveBeenCalledWith(expect.any(String), 'wowsims-batch-results.csv', 'text/csv');
	});

	it('offers Compare encounters under the encounter list only with power tools on', () => {
		const player = mount(THREE);
		const compareButton = () => within(sidebar()!).queryByText('bulk_tab.compare.button');

		expect(compareButton()).toBeNull();

		act(() => patchSlice(player.sim.store, 'ui', { showPowerTools: true }));

		expect(compareButton()).not.toBeNull();
	});
});
