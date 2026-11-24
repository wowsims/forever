import { ItemSlot, Race } from '@generated/proto/common';
import { ContentPhase, DatabaseFilters } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { raceToFaction } from '@sim/proto/utils';
import { loadStoredBulkDropsExcluded, loadStoredBulkDropsSources, patchBulkState } from '@sim/settings/bulk_settings';
import { patchKeyed, patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import {
	ALLIANCE_ITEM_ID,
	BARTHILAS,
	BLACKROCK_SPIRE,
	DRAKKISATH,
	dropsItemsFixture,
	dropsNamesFixture,
	REND,
	RIVENDARE,
	STRATHOLME,
	WEARABLE_ITEM_ID,
	WORN_ITEM_ID,
	WYRMTHALAK,
} from '@testing/storybook/fixtures/bulk-drops';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DropsSetup } from './DropsSetup';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: '', ready: true }) }));
// Every row, since jsdom gives the virtualiser no height.
vi.mock('@ui-kit/VirtualList', () => ({
	VirtualList: ({ count, renderRow }: any) => (
		<div>
			{Array.from({ length: count }, (_unused, index) => (
				<div key={index} data-testid="virtual-list-row">
					{renderRow(index)}
				</div>
			))}
		</div>
	),
}));

const SETUP = 'bulk_tab.drops.setup';
const key = (zoneId: number, npcId: number) => `drop:${zoneId}:${npcId}`;
const names = dropsNamesFixture();
const BRS_LEAVES = [REND, DRAKKISATH, WYRMTHALAK].map(npc => key(BLACKROCK_SPIRE, npc)).sort();

const WORN_SLOTS = [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotHands];
let player: Player<any>;
let worn: number[];

const mount = () => {
	worn = [WORN_ITEM_ID];
	const items = dropsItemsFixture();
	player = makeBulkPlayer({
		getGear: () =>
			new Gear(Object.fromEntries(worn.map((id, index) => [WORN_SLOTS[index], makeEquippedItem({ item: items.find(item => item.id === id)! })]))),
		getFaction: () => raceToFaction[player.sim.store.getState().players[player.storeKey].race as Race],
		getSpec: () => PlayerSpecs.DpsWarrior.specID,
		getClass: () => PlayerSpecs.DpsWarrior.classID,
		sim: {
			getFilters(this: { store: SimStore }) {
				return DatabaseFilters.clone(this.store.getState().sim.filters);
			},
			getPhase(this: { store: SimStore }) {
				return this.store.getState().sim.phase;
			},
			db: {
				getAllItems: () => items,
				getZone: (id: number) => ({ name: names.zone(id) }),
				getNpc: (id: number) => ({ name: names.npc(id) }),
				getRepFaction: () => undefined,
			},
		},
	});
	patchSlice(player.sim.store, 'sim', { phase: ContentPhase.ContentPhase1 });
	seedKeyed(player.sim.store, 'players', player.storeKey, { race: Race.RaceHuman, v: zeroVersions(PLAYER_FIELDS) } as never);
	render(
		<SimHostProvider host={{ player, sim: player.sim } as never}>
			<DropsSetup />
		</SimHostProvider>,
	);
};

const box = (name: string) => screen.getByRole('checkbox', { name });
const disclose = (name: string) => fireEvent.click(screen.getByRole('button', { name }));
const category = 'bulk_tab.drops.sources.categories.drop_dungeon';
const rows = () => screen.queryAllByTestId('drops-item-row');
const rowIds = () => rows().map(row => Number(row.getAttribute('data-item-id')));
const headers = () => screen.queryAllByTestId('drops-group-header').map(header => header.textContent);
const FELSTRIKER = 2001;
const openBlackrockSpire = () => {
	disclose(category);
	disclose('Blackrock Spire');
};

beforeEach(() => {
	mount();
});

describe('DropsSetup', () => {
	it('ticks nothing on fresh storage and lists no item', () => {
		expect(box(category).getAttribute('aria-checked')).toBe('false');
		expect(rows()).toHaveLength(0);
		expect(screen.getByText(`${SETUP}.empty`)).toBeTruthy();
	});

	it('ticks a zone with all its bosses, shows it mixed after unticking one, and stores the leaves', () => {
		openBlackrockSpire();
		fireEvent.click(box('Blackrock Spire'));

		expect(box('Blackrock Spire').getAttribute('aria-checked')).toBe('true');
		expect(loadStoredBulkDropsSources(player)).toEqual(BRS_LEAVES);

		fireEvent.click(box('Overlord Wyrmthalak'));

		expect(box('Blackrock Spire').getAttribute('aria-checked')).toBe('mixed');
		expect(box(category).getAttribute('aria-checked')).toBe('mixed');
		expect(loadStoredBulkDropsSources(player)).toEqual(BRS_LEAVES.filter(leaf => leaf !== key(BLACKROCK_SPIRE, WYRMTHALAK)));
	});

	it('opens a node only on its disclosure button, with its children mounted while open', () => {
		expect(screen.queryByRole('checkbox', { name: 'Blackrock Spire' })).toBeNull();
		disclose(category);
		expect(screen.getByRole('button', { name: category }).getAttribute('aria-expanded')).toBe('true');
		expect(screen.queryByRole('checkbox', { name: 'Overlord Wyrmthalak' })).toBeNull();
		disclose('Blackrock Spire');
		expect(box('Overlord Wyrmthalak')).toBeTruthy();
		disclose('Blackrock Spire');
		expect(screen.queryByRole('checkbox', { name: 'Overlord Wyrmthalak' })).toBeNull();
	});

	it('ticks the five L60 dungeons from the quick pick and clears them', () => {
		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.l60_dungeons` }));

		expect(loadStoredBulkDropsSources(player)).toEqual([...BRS_LEAVES, key(STRATHOLME, RIVENDARE), key(STRATHOLME, BARTHILAS)].sort());
		expect(screen.queryByRole('button', { name: /molten|onyxia/i })).toBeNull();

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.clear` }));

		expect(loadStoredBulkDropsSources(player)).toEqual([]);
		expect(rows()).toHaveLength(0);
	});

	it('lists a worn item greyed out and unticked items unchecked, and Exclude all stores every listed id', () => {
		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.l60_dungeons` }));

		const worn = rows().find(row => Number(row.getAttribute('data-item-id')) === WORN_ITEM_ID)!;
		expect(worn.hasAttribute('data-worn')).toBe(true);
		expect(within(worn).getByRole('checkbox').getAttribute('aria-disabled')).toBe('true');
		expect(within(worn).getByText('bulk_tab.drops.items.worn')).toBeTruthy();
		const listed = rows().filter(row => !row.hasAttribute('data-worn'));
		expect(listed.length).toBeGreaterThan(0);
		for (const row of listed) expect(within(row).getByRole('checkbox').getAttribute('aria-checked')).toBe('true');

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.exclude_all` }));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([2001, 2002, 2004, 2005, 2006, 2007, ALLIANCE_ITEM_ID]);
		for (const row of listed) expect(within(row).getByRole('checkbox').getAttribute('aria-checked')).toBe('false');

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.include_all` }));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([]);
	});

	it('lists only what the class can use, with no word of the class-only and random-suffix items', () => {
		act(() => patchBulkState(player, { dropsSources: [key(BLACKROCK_SPIRE, DRAKKISATH), key(STRATHOLME, RIVENDARE)] }, ['settings']));

		expect(rowIds().sort()).toEqual([2001, 2002, 2004]);
		expect(screen.queryByText('Robe of the Archmage')).toBeNull();
		expect(screen.queryByText('Cloak of the Eagle')).toBeNull();
	});

	it('groups the same items by boss on the switch, an item with two ticked sources under both', () => {
		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.l60_dungeons` }));
		expect(rowIds().filter(id => id === FELSTRIKER)).toHaveLength(1);

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.by_source` }));

		expect(rowIds().filter(id => id === FELSTRIKER)).toHaveLength(2);
		expect(headers()).toEqual([
			'General DrakkisathBlackrock Spire',
			'Overlord WyrmthalakBlackrock Spire',
			'Warchief Rend BlackhandBlackrock Spire',
			'Baron RivendareStratholme',
			'Magistrate BarthilasStratholme',
		]);

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.by_slot` }));

		expect(rowIds().filter(id => id === FELSTRIKER)).toHaveLength(1);
	});

	it('names an item’s first source with a count of the rest, and under a boss only the other sources', () => {
		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.l60_dungeons` }));
		const felstriker = () => rows().filter(row => Number(row.getAttribute('data-item-id')) === FELSTRIKER);

		const [bySlot] = felstriker();
		expect(within(bySlot).getByText('Warchief Rend Blackhand')).toBeTruthy();
		expect(within(bySlot).getByTestId('drops-item-more-sources').textContent).toBe('+1');

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.by_source` }));

		const [underRend, underRivendare] = felstriker();
		expect(within(underRend).queryByText('Warchief Rend Blackhand')).toBeNull();
		expect(within(underRend).getByText('Baron Rivendare')).toBeTruthy();
		expect(within(underRivendare).getByText('Warchief Rend Blackhand')).toBeTruthy();
		expect(within(underRivendare).queryByTestId('drops-item-more-sources')).toBeNull();
	});

	it('rebuilds the list when the phase, the race or the gear changes', () => {
		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.l60_dungeons` }));
		const row = (id: number) => rows().find(shown => Number(shown.getAttribute('data-item-id')) === id);
		expect(row(ALLIANCE_ITEM_ID)).toBeTruthy();
		expect(row(WEARABLE_ITEM_ID)!.hasAttribute('data-worn')).toBe(false);

		// Orc is Horde: the Alliance-only item leaves.
		act(() => patchKeyed(player.sim.store, 'players', player.storeKey, { race: Race.RaceOrc }, ['race']));
		expect(row(ALLIANCE_ITEM_ID)).toBeUndefined();
		expect(row(WEARABLE_ITEM_ID)).toBeTruthy();

		// Wearing an item greys its row out.
		worn.push(WEARABLE_ITEM_ID);
		act(() => patchKeyed(player.sim.store, 'players', player.storeKey, {}, ['gear']));
		expect(row(WEARABLE_ITEM_ID)!.hasAttribute('data-worn')).toBe(true);

		// A Max ilvl of 1 leaves out every fixture item.
		act(() => patchSlice(player.sim.store, 'sim', { filters: DatabaseFilters.create({ maxIlvl: 1 }) }));
		expect(rows()).toHaveLength(0);
		act(() => patchSlice(player.sim.store, 'sim', { filters: DatabaseFilters.create() }));
		expect(rows().length).toBeGreaterThan(0);

		// Every fixture item is of phase 1.
		act(() => patchSlice(player.sim.store, 'sim', { phase: ContentPhase.ContentPhaseNone }));
		expect(rows()).toHaveLength(0);
	});

	it('stores one unticked row and ticks it again', () => {
		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.l60_dungeons` }));

		fireEvent.click(box('Reiver Claws'));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([WEARABLE_ITEM_ID]);
		expect(box('Reiver Claws').getAttribute('aria-checked')).toBe('false');

		fireEvent.click(box('Reiver Claws'));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([]);
	});
});
