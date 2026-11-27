import { Faction, ItemSlot, ItemSpec } from '@generated/proto/common';
import { ContentPhase, DungeonDifficulty, type UIEnchant, type UIItem, UIItem_FactionRestriction, UIItemSource } from '@generated/proto/ui';
import { translateSlotName } from '@i18n/localization';
import type { BulkRankedResult } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { EquippedItem } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import { bulkState, loadStoredBulkDropsOrder, loadStoredBulkDropsView, patchBulkState } from '@sim/settings/bulk_settings';
import { type BulkSlice, patchSlice, PLAYER_FIELDS, type PlayerSlice, seedKeyed, zeroVersions } from '@sim/state/sim_store';
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
	WYRMTHALAK,
} from '@testing/storybook/fixtures/bulk-drops';
import { makeUIItem } from '@testing/storybook/fixtures/protos';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { setDropsOrder } from '../../model/drops/settings';
import { runBulkBatch } from '../../model/run/run';
import { DropsResults } from './DropsResults';

vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string, options?: { z?: string; count?: number }) =>
			options?.z ? `${key} ${options.z}` : options?.count === undefined ? key : `${key}#${options.count}`,
	},
}));
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: '', ready: true }) }));
vi.mock('@sim/proto/enchants', () => ({ getEnchantDescription: async (enchant: UIEnchant) => enchant.name }));
const toastAdd = vi.hoisted(() => vi.fn(() => 'toast-1'));
const toastClose = vi.hoisted(() => vi.fn());
// The standard area's manager only: the results' own inline note renders through the real kit.
vi.mock('@ui-kit/Toast', async importOriginal => ({ ...(await importOriginal<object>()), toastManager: { add: toastAdd, close: toastClose } }));
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('../../model/run/run', () => ({ runBulkBatch: vi.fn(async () => {}) }));

const RESULTS = 'bulk_tab.drops.results';
const FELSTRIKER = 2001;
const MARK = 2006;
const REIVER = 2005;
const DRACORIAN = 2004;
const BLACKBLADE = 2002;
const names = dropsNamesFixture();

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
	combined?: BulkSlice['dropsCombined'];
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
	combined = null,
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
		getClass: () => PlayerSpecs.DpsWarrior.classID,
		sim: {
			getPhase(this: { store: { getState: () => any } }) {
				return this.store.getState().sim.phase;
			},
			db: {
				getItemById: (id: number) => items.find(item => item.id === id),
				lookupItemSpec: (spec: ItemSpec) => {
					const item = items.find(candidate => candidate.id === spec.id);
					return item ? new EquippedItem({ item }) : null;
				},
				getAllItems: () => items,
				getZone: (id: number) => ({ name: names.zone(id) }),
				getNpc: (id: number) => ({ name: names.npc(id) }),
				getRepFaction: () => undefined,
			},
		},
	});
	// The item cells' tooltips read the worn gear from the player's slice.
	seedKeyed(player.sim.store, 'players', 0, { gear, v: zeroVersions(PLAYER_FIELDS) } as unknown as PlayerSlice);
	if (phase !== undefined) patchSlice(player.sim.store, 'sim', { phase });
	patchBulkState(player, {
		dropsStarted: started,
		dropsRunFaction: runFaction,
		dropsResults: started ? dropsResults : null,
		dropsRunSources: runSources ? runSources(dropsRunSources) : dropsRunSources,
		dropsCombined: combined,
		...(view ? { dropsView: view } : {}),
	});
	return render(
		<SimHostProvider host={{ player, sim: player.sim } as never}>
			<DropsResults />
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
	toastAdd.mockClear();
	toastClose.mockClear();
	vi.mocked(runBulkBatch).mockClear();
	localStorage.clear();
});

const orderField = () => within(screen.getByTestId('drops-order')).getByRole('combobox');
const orderOptions = async () => {
	await userEvent.click(orderField());
	return screen.getAllByRole('option');
};
const ranks = () => cards().map(card => within(card).queryByTestId('drops-boss-rank')?.textContent ?? null);

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

	it('opens on a warning in flow that cannot be closed, naming the results next upgrades only', async () => {
		mount();
		const note = await within(screen.getByTestId('drops-next-upgrade-note')).findByTestId('sim-toast');
		expect(note.getAttribute('data-variant')).toBe('warning');
		expect(within(note).getByTestId('sim-toast-title').textContent).toBe(`${RESULTS}.next_upgrade_title`);
		expect(within(note).getByTestId('sim-toast-body').textContent).toBe(`${RESULTS}.next_upgrade_note`);
		expect(within(note).queryByTestId('sim-toast-close')).toBeNull();
		expect(screen.getByTestId('drops-next-upgrade-note').getAttribute('data-inline')).toBe('true');
	});

	it('shows the summary tiles: the current gear, the best drop and the card to farm first with its upgrades', () => {
		mount();
		expect(screen.getByTestId('drops-summary-current').textContent).toContain('750');
		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('Felstriker');
		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('4.00%');
		const farmFirst = screen.getByTestId('drops-summary-farm-first').textContent;
		expect(farmFirst).toContain('Warchief Rend Blackhand');
		expect(farmFirst).toContain(`${RESULTS}.farm_first_upgrades#1`);
		expect(farmFirst).toContain('Blackrock Spire');
	});

	it('ranks the cards by best gain and dims the ones with no upgrade, last', () => {
		mount();
		expect(cardLabels()).toEqual(['Warchief Rend Blackhand', 'Baron Rivendare', 'Overlord Wyrmthalak', 'General Drakkisath']);
		expect(cards().map(card => card.getAttribute('data-upgrade'))).toEqual([null, null, null, 'none']);
	});

	it('shows N / M upgrade, pluralised on M, with the worn item in the strip and each item’s level', () => {
		mount();
		const rend = cards()[0];
		expect(within(rend).getByTestId('drops-boss-upgrades').textContent).toContain(`1 / 2 ${RESULTS}.upgrades#2`);
		const strip = within(rend).getAllByTestId('drops-boss-strip-item');
		expect(strip.map(item => item.getAttribute('data-item-id'))).toEqual([String(FELSTRIKER), String(WORN_ITEM_ID)]);
		expect(strip.map(item => within(item).getByTestId('item-picker-ilvl').textContent)).toEqual(['63', '63']);
		expect(strip[0].getAttribute('data-worn')).toBeNull();
		expect(strip[1].hasAttribute('data-worn')).toBe(true);
		expect(within(cards()[3]).getByTestId('drops-boss-upgrades').textContent).toContain(`0 / 1 ${RESULTS}.upgrades#1`);
	});

	it('shows a card’s combined gain over the combined request’s own baseline, and marks a set that left an upgrade out', () => {
		mount();
		expect(screen.queryByTestId('drops-boss-combined')).toBeNull();

		const wyrmthalak = `drop:${BLACKROCK_SPIRE}:${WYRMTHALAK}`;
		act(() =>
			patchBulkState(player, {
				dropsCombined: new Map([[wyrmthalak, { itemIds: [MARK, REIVER], leftOutIds: [DRACORIAN], gain: 19, margin: 1.9, baselineDps: 760 }]]),
			}),
		);

		const [combined] = screen.getAllByTestId('drops-boss-combined');
		expect(within(cards()[2]).getByTestId('drops-boss-combined')).toBe(combined);
		expect(combined.textContent).toContain('+2.50%');
		expect(within(combined).getByTestId('drops-boss-combined-margin').textContent).toBe('±0.25');
		expect(combined.textContent).toContain('769.00');
		expect(within(combined).getByTestId('drops-boss-combined-partial')).toBeTruthy();
	});

	it('filters the item list to a card’s source on a click, and clears it', () => {
		mount();
		fireEvent.click(within(cards()[0]).getByTestId('drops-boss-select'));

		expect(loadStoredBulkDropsView(player)).toBe('items');
		expect(rowIds()).toEqual([FELSTRIKER, 'current']);
		expect(screen.getByTestId('drops-source-filter').textContent).toContain('Warchief Rend Blackhand');

		fireEvent.click(screen.getByRole('button', { name: `${RESULTS}.clear_filter` }));

		expect(screen.queryByTestId('drops-source-filter')).toBeNull();
		expect(resultRows().length).toBeGreaterThan(2);
	});

	it('shows a new run’s results unfiltered', () => {
		mount();
		fireEvent.click(within(cards()[0]).getByTestId('drops-boss-select'));

		act(() => patchBulkState(player, { dropsResults: dropsResultsFixture().dropsResults }));

		expect(screen.queryByTestId('drops-source-filter')).toBeNull();
		expect(rowIds()).toEqual([FELSTRIKER, MARK, MARK, 'current', REIVER, BLACKBLADE, DRACORIAN]);
	});

	it('lists every row of an item that upgrades in more than one slot, and one row for an item that upgrades in none', () => {
		mount();
		fireEvent.click(view('items'));

		// Mark of the Dragon Lord upgrades in both ring slots; Blackblade of Shahram has one losing row.
		expect(rowIds()).toEqual([FELSTRIKER, MARK, MARK, 'current', REIVER, BLACKBLADE, DRACORIAN]);
	});

	it('shows the current gear row at its rank, labelled, with no gear cells and its Equip hidden', () => {
		mount();
		fireEvent.click(view('items'));

		const current = resultRows()[3];
		expect(current.hasAttribute('data-current')).toBe(true);
		expect(within(current).getByTestId('bulk-equip-btn').hasAttribute('data-hidden')).toBe(true);
		expect(within(current).getByText('bulk_tab.results.current_gear')).toBeTruthy();
		expect(within(current).queryAllByTestId('bulk-result-item')).toHaveLength(0);
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

	it('marks an item that binds on equip and shows a row’s enchant on its own line, with the CSV export beside the view switch', async () => {
		mount();
		fireEvent.click(view('items'));

		expect(within(rowOf(REIVER)).getByTestId('drops-item-boe').textContent).toBe(`${RESULTS}.boe`);
		expect(within(rowOf(FELSTRIKER)).queryByTestId('drops-item-boe')).toBeNull();
		expect(await within(rowOf(DRACORIAN)).findByText('Enchant Gloves - Greater Agility')).toBeTruthy();
		expect(within(rowOf(DRACORIAN)).getByTestId('drops-item-label').textContent).not.toContain('(');
		expect(within(rowOf(FELSTRIKER)).queryByTestId('drops-item-enchant')).toBeNull();
		act(() => patchBulkState(player, { mode: 'drops' }));
		expect(within(screen.getByTestId('drops-results')).getByTestId('bulk-results-export-csv')).toBeTruthy();
	});

	it('equips a row’s gear in place and announces it', () => {
		mount();
		fireEvent.click(view('items'));

		fireEvent.click(within(rowOf(FELSTRIKER)).getByRole('button', { name: 'bulk_tab.results.equip_button' }));

		expect(player.getGear().getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.id).toBe(FELSTRIKER);
		expect(toastAdd).toHaveBeenCalledTimes(1);
	});

	it('equips an item’s best row from the [+] on its card icon, and leaves a worn item without one', () => {
		mount();
		const felstriker = cards()
			.flatMap(card => within(card).getAllByTestId('drops-boss-strip-item'))
			.find(entry => entry.getAttribute('data-item-id') === String(FELSTRIKER))!;

		fireEvent.click(within(felstriker).getByTestId('drops-boss-equip'));

		expect(player.getGear().getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.id).toBe(FELSTRIKER);
		expect(toastAdd).toHaveBeenCalledTimes(1);
		for (const worn of screen.queryAllByTestId('drops-boss-strip-item').filter(entry => entry.hasAttribute('data-worn'))) {
			expect(within(worn).queryByTestId('drops-boss-equip')).toBeNull();
		}
	});

	it('orders the cards by the stored pick, with the rank under Recommended alone, and leaves Farm first on the Recommended first card', async () => {
		mount();
		expect(orderField().textContent).toBe(`${RESULTS}.order.recommended`);
		expect(ranks()).toEqual(['1', '2', '3', null]);

		const options = await orderOptions();
		expect(options.map(option => option.textContent)).toEqual(['recommended', 'gain', 'upgrades', 'zone'].map(order => `${RESULTS}.order.${order}`));
		await userEvent.click(options[2]);

		expect([bulkState(player).dropsOrder, loadStoredBulkDropsOrder(player)]).toEqual(['upgrades', 'upgrades']);
		expect(cardLabels()).toEqual(['Overlord Wyrmthalak', 'Warchief Rend Blackhand', 'Baron Rivendare', 'General Drakkisath']);
		expect(ranks()).toEqual([null, null, null, null]);
		expect(cards().map(card => card.getAttribute('data-upgrade'))).toEqual([null, null, null, 'none']);
		expect(screen.getByTestId('drops-summary-farm-first').textContent).toContain('Warchief Rend Blackhand');
	});

	it('offers Combined gain on a run with combined results, the cards with one first', async () => {
		const wyrmthalak = `drop:${BLACKROCK_SPIRE}:${WYRMTHALAK}`;
		mount({ combined: new Map([[wyrmthalak, { itemIds: [MARK, REIVER], leftOutIds: [], gain: 19, margin: 1.9, baselineDps: 760 }]]) });

		const options = await orderOptions();
		expect(options.at(-1)!.textContent).toBe(`${RESULTS}.order.combined`);
		await userEvent.click(options.at(-1)!);

		expect(loadStoredBulkDropsOrder(player)).toBe('combined');
		expect(cardLabels()).toEqual(['Overlord Wyrmthalak', 'Warchief Rend Blackhand', 'Baron Rivendare', 'General Drakkisath']);
	});

	it('reads a stored Combined gain order as Recommended on a run without combined results, and keeps it stored', async () => {
		mount();
		act(() => setDropsOrder(player, 'combined'));

		expect(orderField().textContent).toBe(`${RESULTS}.order.recommended`);
		expect(ranks()).toEqual(['1', '2', '3', null]);
		expect((await orderOptions()).map(option => option.textContent)).not.toContain(`${RESULTS}.order.combined`);
		expect(loadStoredBulkDropsOrder(player)).toBe('combined');
	});

	it('offers Run Drops again on the equip toast once the recount is in', () => {
		mount();
		act(() => patchBulkState(player, { mode: 'drops', dropsSources: bulkState(player).dropsRunSources, combinationsPending: true, combinations: 3 }));
		fireEvent.click(view('items'));
		fireEvent.click(within(rowOf(FELSTRIKER)).getByRole('button', { name: 'bulk_tab.results.equip_button' }));

		const [[toast]] = toastAdd.mock.calls as unknown as [[{ variant: string; body: ReactNode }]];
		render(<SimHostProvider host={{ player, sim: player.sim } as never}>{toast.body}</SimHostProvider>);
		const again = screen.getByTestId('drops-run-again') as HTMLButtonElement;
		expect(again.textContent).toBe(`${RESULTS}.run_again`);
		expect(again.disabled).toBe(true);

		act(() => patchBulkState(player, { combinationsPending: false }));
		expect(again.disabled).toBe(false);
		fireEvent.click(again);

		expect(toastClose).toHaveBeenCalledWith('toast-1');
		expect(runBulkBatch).toHaveBeenCalledTimes(1);
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

	it('names no card to farm first when no ticked source of the run holds an upgrade', () => {
		mount({ runSources: all => all.filter(key => ![REND, RIVENDARE, WYRMTHALAK].some(npc => key.endsWith(`:${npc}`))) });

		expect(screen.getByTestId('drops-summary-best-drop').textContent).toContain('Felstriker');
		expect(screen.getByTestId('drops-summary-farm-first').textContent).toBe(`${RESULTS}.farm_first—`);
	});

	// Felstriker needs level 40 here, so it leaves Deadmines below the cutoff while Blackrock Spire's other loot keeps it.
	it('makes no card for a ticked dungeon the level cutoff hides, though a run’s item drops there', () => {
		const DEADMINES = 1581;
		const VANCLEEF = 639;
		const deadminesDrop = UIItemSource.create({
			source: {
				oneofKind: 'drop',
				drop: { difficulty: DungeonDifficulty.DifficultyDungeon, npcId: VANCLEEF, zoneId: DEADMINES, otherName: '', category: '' },
			},
		});
		mount({
			itemsOf: items => [
				...items.map(item => (item.id === FELSTRIKER ? { ...item, reqLvl: 40, sources: [...item.sources, deadminesDrop] } : item)),
				makeUIItem({ id: 5191, name: 'Cruel Barb', reqLvl: 18, sources: [deadminesDrop] }),
			],
			runSources: all => [...all, `drop:${DEADMINES}:${VANCLEEF}`],
		});

		expect(cardLabels()).not.toContain(`#${VANCLEEF}`);
		expect(cardLabels()).toContain('Warchief Rend Blackhand');
	});

	it('draws a bracket around tied rows and keeps every row keyed once', () => {
		const error = vi.spyOn(console, 'error').mockImplementation(() => {});
		// Mark of the Dragon Lord's two ring rows tie by the paired error to the next result; the culled Reiver Claws ties with the current gear.
		mount({
			view: 'items',
			chains: chains => {
				const rowAt = (avg: number) => chains.flat().find(row => row.dpsMetrics.avg === avg)!;
				rowAt(765).pairedErrorToNextResult = 10;
				rowAt(749).pairedErrorToBaseline = 10;
				return chains;
			},
		});

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

	it('names a row’s new item beside its slot, with its slot and first source under it and every source in its tooltip', () => {
		mount({ view: 'items' });

		const row = rowOf(FELSTRIKER);
		const label = within(row).getByTestId('drops-item-label');
		expect(label.firstElementChild!.textContent).toBe('Felstriker');
		const detail = label.lastElementChild!;
		expect(detail.textContent).toBe(`${translateSlotName(ItemSlot.ItemSlotMainHand)} · ${names.npc(REND)} +1`);
		expect(detail.getAttribute('data-tooltip-content')).toBe(`${names.npc(REND)}\n${names.npc(RIVENDARE)}`);
	});

	it('draws only the slots a row changes', () => {
		mount({ view: 'items' });

		expect(within(rowOf(FELSTRIKER)).getAllByTestId('bulk-result-item')).toHaveLength(1);
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
