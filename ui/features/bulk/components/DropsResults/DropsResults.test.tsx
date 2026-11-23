import { Faction, ItemSlot } from '@generated/proto/common';
import { ContentPhase, type UIItem, UIItem_FactionRestriction } from '@generated/proto/ui';
import type { BulkRankedResult } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import type { Gear } from '@sim/proto/gear';
import { loadStoredBulkDropsView, patchBulkState } from '@sim/settings/bulk_settings';
import { patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import {
	BLACKROCK_SPIRE,
	DRAKKISATH,
	dropsItemsFixture,
	dropsNamesFixture,
	dropsResultsFixture,
	REND,
	RIVENDARE,
	WORN_ITEM_ID,
} from '@testing/storybook/fixtures/bulk-drops';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DropsResults } from './DropsResults';

vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string, options?: { z?: string; count?: number }) =>
			options?.z ? `${key} ${options.z}` : options?.count === undefined ? key : `${key}#${options.count}`,
	},
}));
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: '', ready: true }) }));
const toastAdd = vi.hoisted(() => vi.fn());
vi.mock('@ui-kit/Toast', () => ({ toastManager: { add: toastAdd } }));

const RESULTS = 'bulk_tab.drops.results';
const FELSTRIKER = 2001;
const MARK = 2006;
const REIVER = 2005;
const DRACORIAN = 2004;
const BLACKBLADE = 2002;
const names = dropsNamesFixture();
const onActivateTab = vi.fn();

let player: Player<any>;
let gear: Gear;

interface MountOptions {
	started?: boolean;
	phase?: ContentPhase;
	view?: 'bosses' | 'items';
	runSources?: (all: string[]) => string[];
	chains?: (chains: BulkRankedResult[][]) => BulkRankedResult[][];
	/** The faction the run started with; the player's live one is `liveFaction`. */
	runFaction?: Faction;
	liveFaction?: Faction;
	itemsOf?: (items: UIItem[]) => UIItem[];
}

const mount = ({
	started = true,
	phase,
	view,
	runSources,
	chains,
	runFaction = Faction.Alliance,
	liveFaction = Faction.Alliance,
	itemsOf,
}: MountOptions = {}) => {
	const items = itemsOf ? itemsOf(dropsItemsFixture()) : dropsItemsFixture();
	const { dropsResults, dropsRunSources } = dropsResultsFixture();
	gear = dropsResults[0].results.currentGear;
	if (chains) dropsResults[0].results.chains = chains(dropsResults[0].results.chains);
	player = makeBulkPlayer({
		getGear: () => gear,
		setGear: (next: Gear) => {
			gear = next;
		},
		getFaction: () => liveFaction,
		getSpec: () => PlayerSpecs.DpsWarrior.specID,
		sim: {
			getPhase(this: { store: { getState: () => any } }) {
				return this.store.getState().sim.phase;
			},
			db: {
				getItemById: (id: number) => items.find(item => item.id === id),
				getZone: (id: number) => ({ name: names.zone(id) }),
				getNpc: (id: number) => ({ name: names.npc(id) }),
				getRepFaction: () => undefined,
			},
		},
	});
	if (phase !== undefined) patchSlice(player.sim.store, 'sim', { phase });
	patchBulkState(player, {
		dropsStarted: started,
		dropsRunFaction: runFaction,
		dropsResults: started ? dropsResults : null,
		dropsRunSources: runSources ? runSources(dropsRunSources) : dropsRunSources,
		...(view ? { dropsView: view } : {}),
	});
	return render(
		<SimHostProvider host={{ player, sim: player.sim } as never}>
			<DropsResults onActivateTab={onActivateTab} />
		</SimHostProvider>,
	);
};

const cards = () => screen.queryAllByTestId('drops-boss-card');
const cardLabels = () => cards().map(card => within(card).getByTestId('drops-boss-label').textContent);
const resultRows = () => screen.queryAllByTestId('drops-result-row');
const rowIds = () => resultRows().map(row => (row.hasAttribute('data-current') ? 'current' : Number(row.getAttribute('data-item-id'))));
const rowOf = (id: number) => resultRows().find(row => row.getAttribute('data-item-id') === String(id))!;
const view = (name: string) => screen.getByRole('button', { name: `${RESULTS}.view.${name}` });

beforeEach(() => {
	onActivateTab.mockClear();
	toastAdd.mockClear();
});

describe('DropsResults', () => {
	it('opens on the stored view, Bosses by default, and stores a switch to Items', () => {
		mount();
		expect(cards().length).toBeGreaterThan(0);
		expect(view('bosses').getAttribute('aria-pressed')).toBe('true');

		fireEvent.click(view('items'));

		expect(loadStoredBulkDropsView(player)).toBe('items');
		expect(cards()).toHaveLength(0);
		expect(resultRows().length).toBeGreaterThan(0);
	});

	it('shows the summary tiles: the current gear, the best drop and the best boss', () => {
		mount();
		expect(screen.getByTestId('drops-summary-current').textContent).toContain('750');
		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('Felstriker');
		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('4.00%');
		expect(screen.getByTestId('drops-summary-best-boss').textContent).toContain('Warchief Rend Blackhand');
	});

	it('ranks the cards by best gain and dims the ones with no upgrade, last', () => {
		mount();
		expect(cardLabels()).toEqual(['Warchief Rend Blackhand', 'Baron Rivendare', 'Overlord Wyrmthalak', 'General Drakkisath']);
		expect(cards().map(card => card.getAttribute('data-upgrade'))).toEqual([null, null, null, 'none']);
	});

	it('shows N / M upgrade, pluralised on M, with the worn item in the strip', () => {
		mount();
		const rend = cards()[0];
		expect(within(rend).getByTestId('drops-boss-upgrades').textContent).toContain(`1 / 2 ${RESULTS}.upgrades#2`);
		const strip = within(rend).getAllByTestId('drops-boss-strip-item');
		expect(strip.map(item => item.getAttribute('data-item-id'))).toEqual([String(FELSTRIKER), String(WORN_ITEM_ID)]);
		expect(strip[0].getAttribute('data-worn')).toBeNull();
		expect(strip[1].hasAttribute('data-worn')).toBe(true);
		expect(within(cards()[3]).getByTestId('drops-boss-upgrades').textContent).toContain(`0 / 1 ${RESULTS}.upgrades#1`);
	});

	it('filters the item list to a card’s source on a click, and clears it', () => {
		mount();
		fireEvent.click(cards()[0]);

		expect(loadStoredBulkDropsView(player)).toBe('items');
		expect(rowIds()).toEqual([FELSTRIKER, 'current']);
		expect(screen.getByTestId('drops-source-filter').textContent).toContain('Warchief Rend Blackhand');

		fireEvent.click(screen.getByRole('button', { name: `${RESULTS}.clear_filter` }));

		expect(screen.queryByTestId('drops-source-filter')).toBeNull();
		expect(resultRows().length).toBeGreaterThan(2);
	});

	it('shows a new run’s results unfiltered, their variations folded', () => {
		mount();
		fireEvent.click(cards()[0]);
		fireEvent.click(screen.getByRole('checkbox', { name: `${RESULTS}.show_all_variations` }));

		act(() => patchBulkState(player, { dropsResults: dropsResultsFixture().dropsResults }));

		expect(screen.queryByTestId('drops-source-filter')).toBeNull();
		expect(rowIds()).toEqual([FELSTRIKER, MARK, 'current', REIVER, BLACKBLADE, DRACORIAN]);
	});

	it('toggles Show all variations when its label text is clicked', () => {
		mount();
		fireEvent.click(view('items'));
		const box = screen.getByRole('checkbox', { name: `${RESULTS}.show_all_variations` });
		expect(box.getAttribute('aria-checked')).toBe('false');

		fireEvent.click(screen.getByText(`${RESULTS}.show_all_variations`));

		expect(box.getAttribute('aria-checked')).toBe('true');
		expect(screen.queryByTestId('drops-result-hidden')).toBeNull();
	});

	it('lists one row per item with its variations folded, and every row with Show all variations', () => {
		mount();
		fireEvent.click(view('items'));

		expect(screen.getByTestId('drops-result-header').textContent).toContain(`${RESULTS}.columns.item`);
		expect(rowIds()).toEqual([FELSTRIKER, MARK, 'current', REIVER, BLACKBLADE, DRACORIAN]);
		expect(within(rowOf(MARK)).getByTestId('drops-result-hidden').textContent).toContain(`${RESULTS}.variations_hidden`);

		fireEvent.click(within(rowOf(MARK)).getByRole('button', { name: 'bulk_tab.drops.setup.show' }));

		expect(rowIds()).toEqual([FELSTRIKER, MARK, MARK, 'current', REIVER, BLACKBLADE, DRACORIAN]);
		expect(screen.queryByTestId('drops-result-hidden')).toBeNull();
	});

	it('shows the current gear row at its rank with no Equip', () => {
		mount();
		fireEvent.click(view('items'));

		const current = resultRows()[2];
		expect(current.hasAttribute('data-current')).toBe(true);
		expect(within(current).queryByRole('button', { name: 'bulk_tab.results.equip_button' })).toBeNull();
		expect(within(current).getByText(`${RESULTS}.current_gear`)).toBeTruthy();
	});

	it('sizes a row’s delta from the paired difference, not from the averages', () => {
		mount();
		fireEvent.click(view('items'));

		// Reiver Claws averages 749 against 750 but pairs +3 against the baseline.
		expect(within(rowOf(REIVER)).getByTestId('results-reference-diff').textContent).toBe('+3.00 (0.40%)');
	});

	it('reads a row with a paired error but no paired difference by the averages', () => {
		mount({
			view: 'items',
			chains: chains => {
				chains.flat()[0].pairedDpsDiffToBaseline = 0;
				return chains;
			},
		});

		// Felstriker's 780 against 750 at stdev 100 and 10,000 iterations each: z = 30 / sqrt(2) ≈ 21.213.
		const diff = within(rowOf(FELSTRIKER)).getByTestId('results-reference-diff');
		expect(diff.textContent).toBe('+30.00 (4.00%)');
		expect(diff.getAttribute('data-tooltip-content')).toBe('common.significance.significant 21.213');
	});

	it('shows a culled row without a ± and marked lower precision', () => {
		mount();
		fireEvent.click(view('items'));

		expect(rowOf(REIVER).getAttribute('data-precision')).toBe('low');
		expect(within(rowOf(REIVER)).queryByTestId('topline-result-stdev')).toBeNull();
		expect(within(rowOf(REIVER)).getByText(`${RESULTS}.low_precision`)).toBeTruthy();
		expect(rowOf(FELSTRIKER).getAttribute('data-precision')).toBeNull();
		expect(within(rowOf(FELSTRIKER)).getByTestId('topline-result-stdev')).toBeTruthy();
	});

	it('equips a row’s gear and opens the gear tab', () => {
		mount();
		fireEvent.click(view('items'));

		fireEvent.click(within(rowOf(FELSTRIKER)).getByRole('button', { name: 'bulk_tab.results.equip_button' }));

		expect(player.getGear().getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.id).toBe(FELSTRIKER);
		expect(onActivateTab).toHaveBeenCalledWith('gear-tab');
		expect(toastAdd).toHaveBeenCalledTimes(1);
	});

	it('keeps the run’s cards under a phase that lists none of its items', () => {
		mount({ phase: ContentPhase.ContentPhaseNone });

		expect(cardLabels()).toEqual(['Warchief Rend Blackhand', 'Baron Rivendare', 'Overlord Wyrmthalak', 'General Drakkisath']);
	});

	it('opens on the Items view when it is the stored one', () => {
		mount({ view: 'items' });

		expect(cards()).toHaveLength(0);
		expect(view('items').getAttribute('aria-pressed')).toBe('true');
		expect(resultRows().length).toBeGreaterThan(0);
	});

	it('names no card for the best drop when no ticked source of the run reaches it', () => {
		mount({ runSources: all => all.filter(key => !key.endsWith(`:${REND}`) && !key.endsWith(`:${RIVENDARE}`)) });

		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('Felstriker');
		expect(screen.getByTestId('drops-summary-best-boss').textContent).toBe(`${RESULTS}.best_boss—`);
	});

	it('draws a bracket around tied rows and keeps every row keyed once', () => {
		const error = vi.spyOn(console, 'error').mockImplementation(() => {});
		mount({ view: 'items', chains: chains => [[...chains[0], ...chains[1]], chains[2], [...chains[3], ...chains[4]], ...chains.slice(5)] });
		fireEvent.click(screen.getByRole('checkbox', { name: `${RESULTS}.show_all_variations` }));

		expect(screen.getAllByTestId('bulk-results-tie-group')).toHaveLength(2);
		expect(resultRows()).toHaveLength(7);
		expect(error).not.toHaveBeenCalled();
		error.mockRestore();
	});

	it('groups by the sources the run started with, not the sources ticked now', () => {
		mount();
		act(() => patchBulkState(player, { dropsSources: [`drop:${BLACKROCK_SPIRE}:${DRAKKISATH}`] }));

		expect(cardLabels()).toEqual(['Warchief Rend Blackhand', 'Baron Rivendare', 'Overlord Wyrmthalak', 'General Drakkisath']);
	});

	it('invites a run before any Drops run and shows nothing once one started', () => {
		mount({ started: false });
		expect(screen.getByText(`${RESULTS}.run_simulation`)).toBeTruthy();

		act(() => patchBulkState(player, { dropsStarted: true }));

		expect(screen.queryByText(`${RESULTS}.run_simulation`)).toBeNull();
		expect(screen.queryByTestId('drops-results')).toBeNull();
	});

	it('reads a non-finalist row’s margin at its own iteration count', () => {
		mount({
			view: 'items',
			chains: chains => {
				const felstriker = chains.flat()[0];
				felstriker.dpsMetrics = { ...felstriker.dpsMetrics, aggregatorData: { n: 2500, sumSq: 0 } };
				return chains;
			},
		});

		// 1.96 * 100 / sqrt(2500) = 3.92 against 1.96 at the 10,000 iterations of the other rows.
		expect(within(rowOf(FELSTRIKER)).getByTestId('topline-result-stdev').textContent).toBe('(4)');
		expect(within(rowOf(MARK)).getByTestId('topline-result-stdev').textContent).toBe('(2)');
	});

	it('keeps the run’s cards when the race changes after the run', () => {
		mount({
			runFaction: Faction.Alliance,
			liveFaction: Faction.Horde,
			itemsOf: items => items.map(item => (item.id === FELSTRIKER ? { ...item, factionRestriction: UIItem_FactionRestriction.ALLIANCE_ONLY } : item)),
		});

		expect(within(cards()[0]).getAllByTestId('drops-boss-strip-item')[0].getAttribute('data-item-id')).toBe(String(FELSTRIKER));
		expect(cardLabels()).toEqual(['Warchief Rend Blackhand', 'Baron Rivendare', 'Overlord Wyrmthalak', 'General Drakkisath']);
	});

	it('numbers the rows without counting the current gear’s', () => {
		mount({ view: 'items' });

		expect(resultRows().map(row => row.firstElementChild!.textContent)).toEqual(['1', '2', '', '3', '4', '5']);
	});

	it('shows the best DPS as the current gear plus the gain, not the raw mean', () => {
		mount({
			chains: chains => {
				const felstriker = chains.flat()[0];
				felstriker.pairedDpsDiffToBaseline = 20;
				return chains;
			},
		});

		expect(within(cards()[0]).getByText('770.00')).toBeTruthy();
		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('770.00');
	});
});
