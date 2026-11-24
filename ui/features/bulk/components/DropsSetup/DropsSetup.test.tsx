import { ArmorType, ItemSlot, Race } from '@generated/proto/common';
import { ContentPhase, DatabaseFilters } from '@generated/proto/ui';
import { translateBulkSlotName } from '@i18n/localization';
import { BulkSimItemSlot } from '@sim/bulk/utils';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { raceToFaction } from '@sim/proto/utils';
import { bulkState, loadStoredBulkDropsExcluded, loadStoredBulkDropsFilters, loadStoredBulkDropsSources, patchBulkState } from '@sim/settings/bulk_settings';
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
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DropsSetup } from './DropsSetup';

// Appends the values a string takes to its key, so a test reads the numbers in the text.
vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string, values: Record<string, unknown> = {}) => {
			const shown = Object.entries(values).filter(([name]) => name !== 'ns');
			return shown.length ? `${key} ${shown.map(([name, value]) => `${name}=${value}`).join(' ')}` : key;
		},
	},
}));
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: '', ready: true }) }));
// Every row, so the assertions are about what the dialog shows rather than what it windowed.
vi.mock('@ui-kit/VirtualList', () => ({
	VirtualList: ({ count, renderRow }: { count: number; renderRow: (index: number) => ReactNode }) => (
		<div data-testid="virtual-list">
			{Array.from({ length: count }, (_unused, index) => (
				<div key={index}>{renderRow(index)}</div>
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
const FELSTRIKER = 2001;
const BLACKBLADE = 2002;
const DRACORIAN = 2004;
const MARK = 2006;
const LIONHEART = ALLIANCE_ITEM_ID;
const DEATH_GRIPS = 2007;
const NON_WORN = [FELSTRIKER, BLACKBLADE, DRACORIAN, WEARABLE_ITEM_ID, MARK, DEATH_GRIPS, LIONHEART].sort((a, b) => a - b);
const HANDS = [DRACORIAN, WEARABLE_ITEM_ID, DEATH_GRIPS, LIONHEART].sort((a, b) => a - b);
const slotName = (slot: BulkSimItemSlot) => translateBulkSlotName(slot);

const openBlackrockSpire = () => {
	disclose(category);
	disclose('Blackrock Spire');
};
const tickEndgame = () => fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.endgame_dungeons` }));

const chooseButton = () => screen.getByRole('button', { name: `${SETUP}.choose_items` });
const openDialog = () => fireEvent.click(chooseButton());
const dialog = () => screen.getByTestId('drops-items-dialog');
const rows = () => screen.queryAllByTestId('drops-item-row');
const rowIds = () => rows().map(row => Number(row.getAttribute('data-item-id')));
const row = (id: number) => rows().find(shown => Number(shown.getAttribute('data-item-id')) === id);
// `hidden`, since the filters menu over the dialog hides the dialog from the accessibility tree.
const railEntries = () => within(screen.getByTestId('drops-items-rail')).getAllByRole('button', { hidden: true });
const rail = () => railEntries().map(entry => entry.textContent);
const picked = () => railEntries().find(entry => entry.getAttribute('aria-current') === 'true')?.textContent;
const pick = (title: string) => fireEvent.click(railEntries().find(entry => entry.textContent!.startsWith(title))!);
const press = (name: string) => fireEvent.click(within(dialog()).getByRole('button', { name: `${SETUP}.${name}` }));
const search = (value: string) => fireEvent.input(within(dialog()).getByTestId('drops-items-search'), { target: { value } });
const group = (grouping: 'by_slot' | 'by_source') => fireEvent.click(within(dialog()).getByRole('button', { name: `${SETUP}.${grouping}` }));
const text = (testId: string) => screen.queryByTestId(testId)?.textContent;
const sorted = (ids: number[]) => [...ids].sort((a, b) => a - b);

beforeEach(() => {
	mount();
});

describe('DropsSetup', () => {
	it('ticks nothing on fresh storage and lists no item', () => {
		expect(box(category).getAttribute('aria-checked')).toBe('false');
		expect(screen.getByText(`${SETUP}.empty`)).toBeTruthy();
		expect(chooseButton().hasAttribute('disabled') || chooseButton().getAttribute('aria-disabled') === 'true').toBe(true);
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

	it('opens a node from a press anywhere on its row but the box, and leaves it unticked', () => {
		fireEvent.click(screen.getAllByTestId('drops-source-count')[0]);

		expect(screen.getByRole('button', { name: category }).getAttribute('aria-expanded')).toBe('true');
		expect(box(category).getAttribute('aria-checked')).toBe('false');
	});

	it('opens a node on its row button, with its children mounted while open', () => {
		expect(screen.queryByRole('checkbox', { name: 'Blackrock Spire' })).toBeNull();
		disclose(category);
		expect(screen.getByRole('button', { name: category }).getAttribute('aria-expanded')).toBe('true');
		expect(screen.queryByRole('checkbox', { name: 'Overlord Wyrmthalak' })).toBeNull();
		disclose('Blackrock Spire');
		expect(box('Overlord Wyrmthalak')).toBeTruthy();
		disclose('Blackrock Spire');
		expect(screen.queryByRole('checkbox', { name: 'Overlord Wyrmthalak' })).toBeNull();
	});

	it('ticks the endgame dungeons from the quick pick and clears them', () => {
		tickEndgame();

		expect(loadStoredBulkDropsSources(player)).toEqual([...BRS_LEAVES, key(STRATHOLME, RIVENDARE), key(STRATHOLME, BARTHILAS)].sort());
		expect(screen.queryByRole('button', { name: /molten|onyxia/i })).toBeNull();

		fireEvent.click(screen.getByRole('button', { name: `${SETUP}.quick_picks.clear` }));

		expect(loadStoredBulkDropsSources(player)).toEqual([]);
		expect(screen.getByText(`${SETUP}.empty`)).toBeTruthy();
	});

	it('counts the ticked items on the card, the worn one left out, with a chip for each slot that holds a ticked item', () => {
		tickEndgame();

		expect(text('drops-items-ticked')).toBe(`${SETUP}.items_ticked ticked=7 total=7`);
		expect(screen.queryByTestId('drops-items-hidden')).toBeNull();
		expect(screen.getAllByTestId('drops-items-slot').map(chip => chip.textContent)).toEqual([
			`${slotName(BulkSimItemSlot.ItemSlotHands)}4`,
			`${slotName(BulkSimItemSlot.ItemSlotFinger)}1`,
			`${slotName(BulkSimItemSlot.ItemSlotHandWeapon)}2`,
		]);

		act(() => patchBulkState(player, { dropsExcluded: [MARK, DEATH_GRIPS] }, ['settings']));

		expect(text('drops-items-ticked')).toBe(`${SETUP}.items_ticked ticked=5 total=7`);
		expect(screen.getAllByTestId('drops-items-slot').map(chip => chip.textContent)).toEqual([
			`${slotName(BulkSimItemSlot.ItemSlotHands)}3`,
			`${slotName(BulkSimItemSlot.ItemSlotHandWeapon)}2`,
		]);
	});

	it('opens the dialog from the card on All, with every listed item once and each group’s ticked and listed counts', () => {
		tickEndgame();
		expect(screen.queryByTestId('drops-items-dialog')).toBeNull();

		openDialog();

		expect(picked()).toBe(`${SETUP}.all7/7`);
		expect(sorted(rowIds())).toEqual(sorted([...NON_WORN, WORN_ITEM_ID]));
		expect(rail()).toEqual([
			`${SETUP}.all7/7`,
			`${slotName(BulkSimItemSlot.ItemSlotHands)}4/4`,
			`${slotName(BulkSimItemSlot.ItemSlotFinger)}1/1`,
			`${slotName(BulkSimItemSlot.ItemSlotHandWeapon)}2/2`,
		]);

		fireEvent.click(box('Reiver Claws'));

		expect(rail().slice(0, 2)).toEqual([`${SETUP}.all6/7`, `${slotName(BulkSimItemSlot.ItemSlotHands)}3/4`]);
	});

	it('lists each item once under All by source, and under a boss only that boss’s rows', () => {
		tickEndgame();
		openDialog();
		group('by_source');

		expect(rail()).toEqual([
			`${SETUP}.all7/7`,
			'General DrakkisathBlackrock Spire1/1',
			'Overlord WyrmthalakBlackrock Spire2/2',
			'Warchief Rend BlackhandBlackrock Spire2/2',
			'Baron RivendareStratholme2/2',
			'Magistrate BarthilasStratholme1/1',
		]);
		expect(rowIds().filter(id => id === FELSTRIKER)).toHaveLength(1);
		expect(sorted(rowIds())).toEqual(sorted([...NON_WORN, WORN_ITEM_ID]));

		pick('Baron Rivendare');

		expect(sorted(rowIds())).toEqual([FELSTRIKER, DRACORIAN]);
	});

	it('names an item’s first source with a count of the rest, and under a boss only the other sources', () => {
		tickEndgame();
		openDialog();

		expect(within(row(FELSTRIKER)!).getByText('Warchief Rend Blackhand')).toBeTruthy();
		expect(within(row(FELSTRIKER)!).getByTestId('drops-item-more-sources').textContent).toBe('+1');
		expect(
			within(row(FELSTRIKER)!)
				.getAllByTestId('drops-item-tag')
				.map(tag => tag.textContent),
		).toEqual(['bulk_tab.drops.items.tags.mh', 'bulk_tab.drops.items.tags.oh']);

		group('by_source');
		pick('Warchief Rend Blackhand');
		expect(within(row(FELSTRIKER)!).queryByText('Warchief Rend Blackhand')).toBeNull();
		expect(within(row(FELSTRIKER)!).getByText('Baron Rivendare')).toBeTruthy();

		pick('Baron Rivendare');
		expect(within(row(FELSTRIKER)!).getByText('Warchief Rend Blackhand')).toBeTruthy();
		expect(within(row(FELSTRIKER)!).queryByTestId('drops-item-more-sources')).toBeNull();
	});

	it('goes back to All when the grouping changes', () => {
		tickEndgame();
		openDialog();
		pick(slotName(BulkSimItemSlot.ItemSlotHands));
		expect(sorted(rowIds())).toEqual(HANDS);

		group('by_source');

		expect(picked()).toBe(`${SETUP}.all7/7`);
	});

	it('moves the pick down and up the rail with the arrow keys, wrapping at the ends', () => {
		tickEndgame();
		openDialog();

		fireEvent.keyDown(dialog(), { key: 'ArrowDown' });
		expect(picked()).toBe(`${slotName(BulkSimItemSlot.ItemSlotHands)}4/4`);
		expect(sorted(rowIds())).toEqual(HANDS);

		fireEvent.keyDown(dialog(), { key: 'ArrowUp' });
		fireEvent.keyDown(dialog(), { key: 'ArrowUp' });
		expect(picked()).toBe(`${slotName(BulkSimItemSlot.ItemSlotHandWeapon)}2/2`);
		expect(sorted(rowIds())).toEqual([FELSTRIKER, BLACKBLADE, WORN_ITEM_ID]);
	});

	it('narrows the rows to the search', () => {
		tickEndgame();
		openDialog();

		search('claws');
		expect(rowIds()).toEqual([WEARABLE_ITEM_ID]);

		search('rivendare gauntlets');
		expect(rowIds()).toEqual([DRACORIAN]);

		search('');
		expect(rows()).toHaveLength(8);
	});

	it('narrows the rows to the filters, stores them per spec, and keeps the hidden ticks in what runs', () => {
		tickEndgame();
		openDialog();
		const settingsBefore = bulkState(player).v.settings;
		const filtersButton = () => within(dialog()).getByTestId('drops-items-filters-button');
		expect(filtersButton().textContent).toBe('gear_tab.gear_picker.filters_button');

		fireEvent.click(filtersButton());
		fireEvent.click(document.getElementById(`filters-armor-type-${ArmorType.ArmorTypePlate}`)!);

		expect(rowIds()).not.toEqual(expect.arrayContaining([DRACORIAN]));
		expect(sorted(rowIds())).toEqual([FELSTRIKER, BLACKBLADE, WORN_ITEM_ID, MARK]);
		expect(loadStoredBulkDropsFilters(player)!.armorTypes).not.toContain(ArmorType.ArmorTypePlate);
		expect(filtersButton().textContent).toBe('gear_tab.gear_picker.filters_button_active active=1');
		expect(text('drops-items-hidden')).toBe(`${SETUP}.hidden_ticked count=4`);
		expect(text('drops-items-ticked')).toBe(`${SETUP}.items_ticked ticked=7 total=7`);
		expect(rail()[0]).toBe(`${SETUP}.all7/7`);
		expect(bulkState(player).dropsExcluded).toEqual([]);
		expect(bulkState(player).v.settings).toBe(settingsBefore);

		fireEvent.keyDown(document.activeElement!, { key: 'Escape' });
		press('exclude_shown');

		expect(loadStoredBulkDropsExcluded(player)).toEqual([FELSTRIKER, BLACKBLADE, MARK]);
	});

	it('includes and excludes only the shown rows that are not worn', () => {
		tickEndgame();
		openDialog();

		pick(slotName(BulkSimItemSlot.ItemSlotHands));
		press('exclude_shown');
		expect(loadStoredBulkDropsExcluded(player)).toEqual(HANDS);

		search('claws');
		press('include_shown');
		expect(loadStoredBulkDropsExcluded(player)).toEqual(HANDS.filter(id => id !== WEARABLE_ITEM_ID));

		search('');
		pick(slotName(BulkSimItemSlot.ItemSlotHandWeapon));
		press('exclude_shown');
		expect(loadStoredBulkDropsExcluded(player)).toEqual(sorted([FELSTRIKER, BLACKBLADE, DRACORIAN, DEATH_GRIPS, LIONHEART]));
	});

	it('lists a worn item greyed out and unticked items unchecked', () => {
		tickEndgame();
		act(() => patchBulkState(player, { dropsExcluded: [MARK] }, ['settings']));
		openDialog();

		const worn = row(WORN_ITEM_ID)!;
		expect(worn.hasAttribute('data-worn')).toBe(true);
		expect(within(worn).getByRole('checkbox').getAttribute('aria-disabled')).toBe('true');
		expect(within(worn).getByText('bulk_tab.drops.items.worn')).toBeTruthy();
		expect(box('Mark of the Dragon Lord').getAttribute('aria-checked')).toBe('false');
		expect(box('Felstriker').getAttribute('aria-checked')).toBe('true');

		fireEvent.click(worn);
		expect(bulkState(player).dropsExcluded).toEqual([MARK]);
	});

	it('ticks a row from a press on its sources', () => {
		tickEndgame();
		openDialog();

		fireEvent.click(within(row(FELSTRIKER)!).getByText('Warchief Rend Blackhand'));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([FELSTRIKER]);
	});

	it('lists only what the class can use, with no word of the class-only and random-suffix items', () => {
		act(() => patchBulkState(player, { dropsSources: [key(BLACKROCK_SPIRE, DRAKKISATH), key(STRATHOLME, RIVENDARE)] }, ['settings']));
		openDialog();

		expect(sorted(rowIds())).toEqual([2001, 2002, 2004]);
		expect(screen.queryByText('Robe of the Archmage')).toBeNull();
		expect(screen.queryByText('Cloak of the Eagle')).toBeNull();
	});

	it('rebuilds the list when the phase, the race or the gear changes', () => {
		tickEndgame();
		openDialog();
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

		// Every fixture item is above the gear picker's Max ilvl of 1 and still listed.
		const listed = rowIds();
		act(() => patchSlice(player.sim.store, 'sim', { filters: DatabaseFilters.create({ maxIlvl: 1 }) }));
		act(() => patchKeyed(player.sim.store, 'players', player.storeKey, {}, ['gear']));
		expect(rowIds()).toEqual(listed);

		// Every fixture item is of phase 1.
		act(() => patchSlice(player.sim.store, 'sim', { phase: ContentPhase.ContentPhaseNone }));
		expect(rows()).toHaveLength(0);
	});

	it('stores one unticked row and ticks it again', () => {
		tickEndgame();
		openDialog();

		fireEvent.click(box('Reiver Claws'));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([WEARABLE_ITEM_ID]);
		expect(box('Reiver Claws').getAttribute('aria-checked')).toBe('false');

		fireEvent.click(box('Reiver Claws'));

		expect(loadStoredBulkDropsExcluded(player)).toEqual([]);
	});
});
