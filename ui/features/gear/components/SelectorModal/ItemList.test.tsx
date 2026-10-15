import { Class, ItemQuality, ItemSlot, ItemType, ScalingItemProperties, Stat } from '@generated/proto/common';
import { DatabaseFilters, UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore, patchSlice, type SimStore } from '@sim/state/sim_store';
import { storeSource as store } from '@testing/gear';
import { fakeHost } from '@testing/sim/fake_host';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { type ItemData, type ItemListType, SelectorModalTabs } from '../../types';
import type { ItemListToggle, ItemListToolbarViewProps } from './ItemList';
import type { SelectorTab } from './utils';

vi.mock('@sim/state/subscriptions', async () =>
	(await import('@testing/sim/subscriptions')).mockSubscriptions((await import('@testing/gear')).storeSource.subscribe),
);

// Every row, so the assertions are about what the list selected rather than what it windowed.
vi.mock('@ui-kit/VirtualList', () => ({
	VirtualList: ({ count, rowClassName, renderRow }: any) => (
		<div className="virtual-list">
			{Array.from({ length: count }, (_unused, index) => (
				<div key={index} data-testid="virtual-list-row" className={rowClassName?.(index)}>
					{renderRow(index)}
				</div>
			))}
		</div>
	),
}));

vi.mock('./ItemListRow', () => ({
	ItemListRow: ({ itemData, active, ep, favourited, onToggleFavourite }: any) => (
		<span
			data-row={itemData.name}
			data-active={active ? '' : undefined}
			data-ep={ep?.ep}
			data-equipped-ep={String(ep?.equippedEP)}
			data-favourited={String(favourited)}>
			<button type="button" data-toggle={itemData.name} onClick={onToggleFavourite} />
		</span>
	),
}));

const filtersMenu = vi.hoisted(() => ({ opens: [] as boolean[], setOpen: null as ((open: boolean) => void) | null }));
vi.mock('../FiltersMenu', () => ({
	FiltersMenu: ({ open, onOpenChange, slot }: any) => {
		filtersMenu.opens.push(open);
		filtersMenu.setOpen = onOpenChange;
		return <span data-filters-menu={String(open)} data-filters-slot={slot} />;
	},
}));

const { ItemList, ItemListToolbarView } = await import('./ItemList');
const { Player } = await import('@sim/player/player');
const { Stats } = await import('@sim/proto/stats');
const { isEnchantRelevantFor } = await import('@sim/settings/consumables');

// `sortItemIdxs` reads the base scaling entry, not `ilvl`, whenever it sorts by item level.
const item = (id: number, name: string, ilvl: number) =>
	Item.create({
		id,
		name,
		ilvl,
		quality: ItemQuality.ItemQualityEpic,
		scalingOptions: { 0: ScalingItemProperties.create({ ilvl }) },
	});

const row = (id: number, name: string, ilvl: number, phase = 1): ItemData<ItemListType, string> => ({
	item: item(id, name, ilvl) as unknown as ItemListType,
	name,
	searchText: name,
	id,
	actionId: { itemId: id } as any,
	quality: ItemQuality.ItemQualityEpic,
	phase,
	ilvl,
	ignoreEPFilter: false,
	nameDescription: '',
	onEquip: () => undefined,
});

const ROWS = [row(1, 'Alpha', 500), row(2, 'Beta', 520), row(3, 'Gamma', 510)];

const enchantRow = (enchant: Enchant): ItemData<ItemListType, string> => ({
	item: enchant as unknown as ItemListType,
	name: enchant.name,
	searchText: enchant.name,
	id: enchant.effectId,
	actionId: { spellId: enchant.effectId } as any,
	quality: ItemQuality.ItemQualityEpic,
	phase: 1,
	ilvl: 0,
	ignoreEPFilter: false,
	nameDescription: '',
	onEquip: () => undefined,
});

describe('ItemList', () => {
	let filters: DatabaseFilters;
	let showEPValues: boolean;
	let host: IndividualSimHost<any>;
	let setFilters: ReturnType<typeof vi.fn>;
	let realStore: SimStore;

	const tab = (label: SelectorModalTabs, over: Partial<SelectorTab> = {}): SelectorTab =>
		({
			label,
			itemData: ROWS,
			computeEP: (candidate: any) => candidate?.ilvl ?? 0,
			equippedToItem: (equipped: EquippedItem | null) => equipped?.item ?? null,
			onRemove: () => undefined,
			...over,
		}) as SelectorTab;

	const setup = ({
		label = SelectorModalTabs.Items,
		slot = ItemSlot.ItemSlotHead,
		equipped = null as EquippedItem | null,
		over = {} as Partial<SelectorTab>,
	} = {}) =>
		render(
			<SimHostProvider host={host}>
				<ItemList tab={tab(label, over)} slot={slot} equippedItem={equipped} />
			</SimHostProvider>,
		);

	const names = (container: HTMLElement) => Array.from(container.querySelectorAll('[data-row]')).map(node => (node as HTMLElement).dataset.row);

	beforeEach(() => {
		filtersMenu.opens.length = 0;
		filters = DatabaseFilters.create({});
		showEPValues = true;
		setFilters = vi.fn((next: DatabaseFilters) => {
			filters = next;
		});
		realStore = createSimStore();
		patchSlice(realStore, 'sim', { phase: 5 });
		patchSlice(realStore, 'ui', { showEPValues });
		host = fakeHost({
			sim: { store: realStore },
			player: {
				sim: {
					store: realStore,
					db: { getNpc: () => undefined },
					getFilters: () => DatabaseFilters.clone(filters),
					setFilters,
					getPhase: () => 5,
					getShowEPValues: () => showEPValues,
				},
				getEquippedItem: () => null,
				getPlayerClass: () => ({ weaponTypes: [{ weaponType: 1 }] }),
				getClass: () => Class.ClassWarrior,
				filterItemData: (idxs: number[]) => idxs,
				filterEnchantData: (idxs: number[]) => idxs,
			},
		});
	});

	it('offers the filters button and its dialog only on the items tab', () => {
		const { container, unmount } = setup();
		expect(container.querySelector('[data-testid="selector-modal-filters-button"]')).not.toBeNull();
		expect(container.querySelector('[data-filters-menu]')?.getAttribute('data-filters-slot')).toBe(String(ItemSlot.ItemSlotHead));
		unmount();

		const enchants = setup({ label: SelectorModalTabs.Enchants });
		expect(enchants.container.querySelector('[data-testid="selector-modal-filters-button"]')).toBeNull();
		expect(enchants.container.querySelector('[data-filters-menu]')).toBeNull();
	});

	it('opens the filters dialog from the filters button', () => {
		const { container } = setup();
		expect(container.querySelector('[data-filters-menu]')?.getAttribute('data-filters-menu')).toBe('false');

		act(() => container.querySelector<HTMLButtonElement>('[data-testid="selector-modal-filters-button"]')!.click());
		expect(container.querySelector('[data-filters-menu]')?.getAttribute('data-filters-menu')).toBe('true');

		act(() => filtersMenu.setOpen!(false));
		expect(container.querySelector('[data-filters-menu]')?.getAttribute('data-filters-menu')).toBe('false');
	});

	it('withholds the EP option and the EP column from a trinket slot, which it computes no EP for', () => {
		const { container, unmount } = setup({ slot: ItemSlot.ItemSlotTrinket1 });
		expect(container.querySelector('[data-testid="selector-modal-show-ep-values"]')).toBeNull();
		expect(container.querySelector('#show-ep-values-selector')).toBeNull();
		unmount();

		const head = setup();
		expect(head.container.querySelector('[data-testid="selector-modal-show-ep-values"]')).not.toBeNull();
		expect(head.container.querySelector('#show-ep-values-selector')).not.toBeNull();
		expect(head.container.querySelector('[data-testid="ep-label"]')).not.toBeNull();
	});

	it('lists enchants by lvl, highest first with ties in the order given, and offers no EP on the enchants tab', () => {
		const enchants = [
			Enchant.create({ effectId: 1, name: 'Low', lvl: 20 }),
			Enchant.create({ effectId: 2, name: 'HighFirst', lvl: 60 }),
			Enchant.create({ effectId: 3, name: 'Mid', lvl: 40 }),
			Enchant.create({ effectId: 4, name: 'HighSecond', lvl: 60 }),
		];
		const computeEP = vi.fn((candidate: any) => (candidate.effectId === 1 ? 100 : 0));
		const { container } = setup({
			label: SelectorModalTabs.Enchants,
			slot: ItemSlot.ItemSlotMainHand,
			over: { itemData: enchants.map(enchantRow), computeEP },
		});
		expect(names(container)).toEqual(['HighFirst', 'HighSecond', 'Mid', 'Low']);
		expect(computeEP).not.toHaveBeenCalled();
		expect(container.querySelector('[data-testid="ep-label"]')).toBeNull();
		expect(container.querySelector('[data-testid="selector-modal-show-ep-values"]')).toBeNull();
		expect(container.querySelector('[data-testid="selector-modal-show-all-enchants"]')).not.toBeNull();

		act(() => container.querySelector<HTMLElement>('[data-testid="ilvl-label"] button')!.click());
		expect(names(container)).toEqual(['Low', 'Mid', 'HighFirst', 'HighSecond']);
	});

	it('offers the weapon options in a main hand and, for a warrior, an off hand — never in another slot', () => {
		const shown = (slot: ItemSlot, label = SelectorModalTabs.Items) => {
			const { container, unmount } = setup({ slot, label });
			const box = !!container.querySelector('[data-testid="selector-modal-show-1h-weapons"]');
			expect(!!container.querySelector('#show-1h-weapons-selector')).toBe(box);
			unmount();
			return box;
		};

		expect(shown(ItemSlot.ItemSlotMainHand)).toBe(true);
		expect(shown(ItemSlot.ItemSlotOffHand)).toBe(true);
		expect(shown(ItemSlot.ItemSlotHead)).toBe(false);
		expect(shown(ItemSlot.ItemSlotMainHand, SelectorModalTabs.Enchants)).toBe(false);
	});

	it('withholds the off hand weapon options from a class that is not a warrior', () => {
		(host.player as any).getClass = () => Class.ClassRogue;
		const { container } = setup({ slot: ItemSlot.ItemSlotOffHand });
		expect(container.querySelector('[data-testid="selector-modal-show-2h-weapons"]')).toBeNull();
	});

	it('names the remove button after the tab it is on', () => {
		const labelFor = (label: SelectorModalTabs) => {
			const { container, unmount } = setup({ label });
			const text = container.querySelector('[data-testid="selector-modal-remove-button"]')!.textContent;
			unmount();
			return text;
		};

		expect(labelFor(SelectorModalTabs.Items)).toContain('unequip_item');
		expect(labelFor(SelectorModalTabs.Enchants)).toContain('remove_enchant');
		expect(labelFor(SelectorModalTabs.RandomSuffixes)).toContain('remove_random_suffix');
	});

	it('narrows the rows to the search, without losing the sort the user chose', () => {
		const { container } = setup();
		expect(names(container)).toEqual(['Beta', 'Gamma', 'Alpha']);

		act(() => container.querySelector<HTMLElement>('[data-testid="ilvl-label"] button')!.click());
		expect(names(container)).toEqual(['Beta', 'Gamma', 'Alpha']);
		act(() => container.querySelector<HTMLElement>('[data-testid="ilvl-label"] button')!.click());
		expect(names(container)).toEqual(['Alpha', 'Gamma', 'Beta']);

		fireEvent.input(container.querySelector('[data-testid="selector-modal-search"]')!, { target: { value: 'a' } });
		expect(names(container)).toEqual(['Alpha', 'Gamma', 'Beta']);

		fireEvent.input(container.querySelector('[data-testid="selector-modal-search"]')!, { target: { value: 'bet' } });
		expect(names(container)).toEqual(['Beta']);
	});

	it('drops a row from a later phase than the one the sim is on', () => {
		(host.player.sim as any).getPhase = () => 2;
		patchSlice(realStore, 'sim', { phase: 2 });
		const { container } = setup({ over: { itemData: [row(1, 'Alpha', 500, 1), row(2, 'Beta', 520, 5)] } });
		expect(names(container)).toEqual(['Alpha']);
	});

	it('marks the equipped row active and hands the row its EP against the equipped one', () => {
		const equipped = { item: item(2, 'Beta', 520) } as unknown as EquippedItem;
		const { container } = setup({ equipped });

		const rows = Array.from(container.querySelectorAll('[data-testid="virtual-list-row"]'));
		expect(rows.map(node => node.querySelector('[data-row]')!.hasAttribute('data-active'))).toEqual([true, false, false]);
		expect(Array.from(container.querySelectorAll('[data-row]')).map(node => (node as HTMLElement).dataset.equippedEp)).toEqual(['520', '520', '520']);
	});

	it('floats a favourited row to the top and writes a toggle back through the sim', () => {
		filters = DatabaseFilters.create({ favoriteItems: [1] });
		const { container } = setup();
		expect(names(container)).toEqual(['Alpha', 'Beta', 'Gamma']);
		expect(Array.from(container.querySelectorAll('[data-row]')).map(node => (node as HTMLElement).dataset.favourited)).toEqual(['true', 'false', 'false']);

		act(() => container.querySelector<HTMLButtonElement>('[data-toggle=Beta]')!.click());
		expect(setFilters.mock.calls[0][0].favoriteItems).toEqual([1, 2]);

		act(() => container.querySelector<HTMLButtonElement>('[data-toggle=Alpha]')!.click());
		expect(setFilters.mock.calls[1][0].favoriteItems).toEqual([2]);
	});

	it('hides the EP column through the list class when EP values are off', () => {
		showEPValues = false;
		patchSlice(realStore, 'ui', { showEPValues: false });
		const { container } = setup();
		expect(container.querySelector('[data-testid="selector-modal-list"]')!.hasAttribute('data-hide-ep')).toBe(true);
	});

	it('lists the enchants worth something to the spec, the procs and the equipped one, and every enchant once Show All is ticked', () => {
		const head = (effectId: number, name: string, over: Partial<Enchant> = {}) => Enchant.create({ effectId, name, type: ItemType.ItemTypeHead, ...over });
		const granting = (stat: Stat) => ({ stats: new Stats().withStat(stat, 3).asProtoArray() });
		const relevant = head(1, 'Relevant', granting(Stat.StatStrength));
		const worthless = head(2, 'Worthless', granting(Stat.StatIntellect));
		const proc = head(3, 'Proc');
		const equipped = head(4, 'Equipped', granting(Stat.StatIntellect));
		Object.assign(host.player, {
			getEquippedItem: () => ({ item: Item.create({ type: ItemType.ItemTypeHead }), enchant: equipped }),
			filterEnchantData: Player.prototype.filterEnchantData,
			isEnchantRelevant: (enchant: Enchant) => isEnchantRelevantFor(enchant, { epStats: [Stat.StatStrength], epPseudoStats: [], displayStats: [] }),
		});

		const { container } = setup({ label: SelectorModalTabs.Enchants, over: { itemData: [relevant, worthless, proc, equipped].map(enchantRow) } });
		expect(names(container).sort()).toEqual(['Equipped', 'Proc', 'Relevant']);

		act(() => container.querySelector<HTMLElement>('#show-all-enchants-selector')!.click());
		expect(setFilters.mock.calls[0][0].showAllEnchants).toBe(true);
		act(() => store.notify());
		expect(names(container).sort()).toEqual(['Equipped', 'Proc', 'Relevant', 'Worthless']);
	});

	it('re-reads the rows when the sim announces a filter change', () => {
		const { container } = setup();
		expect(names(container)).toEqual(['Beta', 'Gamma', 'Alpha']);

		act(() => {
			filters = DatabaseFilters.create({ favoriteItems: [3] });
			store.notify();
		});
		expect(names(container)).toEqual(['Gamma', 'Beta', 'Alpha']);
	});
});

const toggle = (name: string, value = false): ItemListToggle => ({
	id: `show-${name}-selector`,
	testId: `selector-modal-show-${name}`,
	label: `Show ${name}`,
	value,
	onChange: vi.fn(),
});

const renderView = (props: Partial<ItemListToolbarViewProps> = {}) => {
	const all: ItemListToolbarViewProps = {
		search: '',
		onSearchChange: vi.fn(),
		phase: 2,
		onPhaseChange: vi.fn(),
		toggles: [],
		remove: { label: 'Unequip Item', onClick: vi.fn() },
		...props,
	};
	const utils = render(<ItemListToolbarView {...all} />);
	return { ...utils, props: all };
};

describe('ItemListToolbarView', () => {
	it('writes the search back as it is typed', () => {
		const { props } = renderView({ search: 'thunder' });
		const input = screen.getByTestId('selector-modal-search') as HTMLInputElement;
		expect(input.value).toBe('thunder');

		fireEvent.input(input, { target: { value: 'thunderfury' } });
		expect(props.onSearchChange).toHaveBeenCalledWith('thunderfury');
	});

	it('offers the filters button beside a filters dialog and opens it', () => {
		const onOpenFilters = vi.fn();
		renderView({ filtersMenu: <span data-testid="filters-dialog" />, onOpenFilters });
		expect(screen.getByTestId('filters-dialog')).toBeTruthy();

		fireEvent.click(screen.getByTestId('selector-modal-filters-button'));
		expect(onOpenFilters).toHaveBeenCalledOnce();
	});

	it('leaves the filters button out when there is no filters dialog', () => {
		renderView();
		expect(screen.queryByTestId('selector-modal-filters-button')).toBeNull();
	});

	it('shows the phase and writes a pick back', async () => {
		const { props } = renderView({ phase: 3 });
		await userEvent.click(within(screen.getByTestId('selector-modal-phase-selector')).getByRole('combobox'));
		const options = screen.getAllByRole('option');
		expect(options.map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'false', 'true', 'false', 'false']);

		await userEvent.click(options[4]);
		expect(props.onPhaseChange).toHaveBeenCalledWith(5);
	});

	it('renders one checkbox per toggle, in order, under its own id and test id', () => {
		const { container } = renderView({ toggles: [toggle('1h-weapons', true), toggle('ep-values')] });
		const wrappers = Array.from(container.querySelectorAll('[data-testid^="selector-modal-show-"]'));
		expect(wrappers.map(node => node.getAttribute('data-testid'))).toEqual(['selector-modal-show-1h-weapons', 'selector-modal-show-ep-values']);
		expect(container.querySelector<HTMLInputElement>('#show-1h-weapons-selector')!.checked).toBe(true);
		expect(container.querySelector<HTMLInputElement>('#show-ep-values-selector')!.checked).toBe(false);
		expect(screen.getByText('Show 1h-weapons')).toBeTruthy();
	});

	it('writes a toggle back through its own callback', () => {
		const weapons = toggle('1h-weapons');
		const ep = toggle('ep-values', true);
		const { container } = renderView({ toggles: [weapons, ep] });

		fireEvent.click(container.querySelector('#show-ep-values-selector')!);
		expect(ep.onChange).toHaveBeenCalledWith(false);
		expect(weapons.onChange).not.toHaveBeenCalled();
	});

	it('names the remove button and calls it', () => {
		const onClick = vi.fn();
		renderView({ remove: { label: 'Remove Enchant', onClick } });
		const button = screen.getByTestId('selector-modal-remove-button');
		expect(button.textContent).toBe('Remove Enchant');

		fireEvent.click(button);
		expect(onClick).toHaveBeenCalledOnce();
	});

	it('leaves the remove button out without a remove action', () => {
		renderView({ remove: undefined });
		expect(screen.queryByTestId('selector-modal-remove-button')).toBeNull();
	});
});
