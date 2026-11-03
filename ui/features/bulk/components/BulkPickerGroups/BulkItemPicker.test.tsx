import { OpenSelectorModalContext, type SelectorModalState, useSelectorModalState } from '@features/gear/hooks/useSelectorModal';
import { SelectorModalTabs } from '@features/gear/types';
import { ItemSlot, ItemSpec } from '@generated/proto/common';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState, patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, patchKeyed, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { BulkItemPickerState } from './BulkItemPicker';

vi.mock('@features/gear/components/ItemCell', () => ({
	ItemDetailCell: ({
		className,
		rootDataAttributes,
		action,
		testId,
		onOpen,
	}: {
		className?: unknown;
		rootDataAttributes?: Record<string, string>;
		action?: unknown;
		testId?: string;
		onOpen?: (tab: number) => void;
	}) => (
		<div data-testid={testId} className={String(className)} {...rootDataAttributes} onClick={() => onOpen?.(0)}>
			{action as never}
		</div>
	),
}));

vi.mock('@sim/proto/items', async () => {
	const actual = await vi.importActual<typeof import('@sim/proto/items')>('@sim/proto/items');
	return { ...actual, getEligibleItemSlots: () => [ItemSlot.ItemSlotFinger1], canEquipItem: () => true };
});

vi.mock('@sim/bulk/utils', async () => {
	const actual = await vi.importActual<typeof import('@sim/bulk/utils')>('@sim/bulk/utils');
	return { ...actual, getBulkPlayerCanDualWield: () => true };
});

const { BulkPickerGroups } = await import('./BulkPickerGroups');
const { BulkItemPickerView } = await import('./BulkItemPicker');
const { createBulkGearData } = await import('../../model/gear/gear_data');
const { addBulkItems, hasBulkItem, loadEquippedBulkItems, seedBulkPickerGroups } = await import('../../model/gear/items');

const STORE_KEY = 7;

const fakeItem = (id: number, name: string, enchant = 0): EquippedItem => {
	const item = {
		id,
		item: { id, name, unique: false },
		_item: { id, name, unique: false },
		asSpec: () => ItemSpec.create({ id, enchant }),
		equals: (other: EquippedItem | null) => !!other && other.id === id,
		withDynamicStats: () => item as unknown as EquippedItem,
	};
	return item as unknown as EquippedItem;
};

const BATCHED = fakeItem(1, 'Batched Ring');
const WORN = fakeItem(2, 'Worn Ring');
const PICKED = fakeItem(3, 'Picked Ring');
const ENCHANTED = fakeItem(4, 'Enchanted Ring', 2928);

const mount = () => {
	const store = createSimStore();
	const gear = { getEquippedItem: () => null };
	seedKeyed(store, 'players', STORE_KEY, { gear, v: zeroVersions(PLAYER_FIELDS) } as never);

	const db = { lookupItemSpec: (spec: ItemSpec) => [BATCHED, WORN, PICKED, ENCHANTED].find(item => item.id === spec.id) ?? null };
	const player = {
		sim: { store, db, isNative: false, encounter: { getAreaTypes: () => [] } },
		storeKey: STORE_KEY,
		getGear: () => gear,
		getEquippedItems: () => [WORN],
		getSpec: () => 0,
		getPlayerSpec: () => ({ friendlyName: 'test' }),
	} as unknown as Player<any>;

	seedBulkSettings(player);
	seedBulkPickerGroups(player);
	addBulkItems(player, [BATCHED.asSpec()], true);

	let modal: SelectorModalState | null = null;
	const Harness = () => {
		modal = useSelectorModalState();
		return (
			<OpenSelectorModalContext value={modal.openTab}>
				<BulkPickerGroups />
			</OpenSelectorModalContext>
		);
	};

	const host = { player, sim: player.sim } as never;
	const { container } = render(
		<SimHostProvider host={host}>
			<Harness />
		</SimHostProvider>,
	);

	const cell = () => container.querySelector<HTMLElement>('[data-testid="bulk-item-picker"]')!;
	return {
		player,
		cell,
		modal: () => modal!,
		clickAdd: (group: string) =>
			fireEvent.click(
				within(container.querySelector<HTMLElement>(`[data-gear-group="${group}"]`)!).getByRole('button', { name: 'bulk_tab.picker.add_items_to' }),
			),
		view: () => ({
			equipped: cell().hasAttribute('data-equipped'),
			removable: !!cell().querySelector('[data-testid="item-picker-actions-btn"]'),
		}),
	};
};

const WEAPON_BEFORE = fakeItem(11, 'Old Axe');
const WEAPON_AFTER = fakeItem(12, 'New Axe');

const mountFrozenWeapon = () => {
	const store = createSimStore();
	let worn = WEAPON_BEFORE;
	const gearOf = (weapon: EquippedItem) => ({ getEquippedItem: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotMainHand ? weapon : null) });
	let gear = gearOf(worn);
	seedKeyed(store, 'players', STORE_KEY, { gear, v: zeroVersions(PLAYER_FIELDS) } as never);

	const player = {
		sim: { store, db: { lookupItemSpec: () => null }, isNative: false, encounter: { getAreaTypes: () => [] } },
		storeKey: STORE_KEY,
		getGear: () => gear,
		getEquippedItems: () => {
			const slots: Array<EquippedItem> = [];
			slots[ItemSlot.ItemSlotMainHand] = worn;
			return slots;
		},
		getSpec: () => 0,
		getPlayerSpec: () => ({ friendlyName: 'test' }),
	} as unknown as Player<any>;

	seedBulkSettings(player);
	seedBulkPickerGroups(player);
	patchBulkState(player, { frozenWeaponSlot: ItemSlot.ItemSlotMainHand });
	loadEquippedBulkItems(player);

	const host = { player, sim: player.sim } as never;
	const { container } = render(
		<SimHostProvider host={host}>
			<BulkPickerGroups />
		</SimHostProvider>,
	);

	const cell = () => container.querySelector<HTMLElement>('[data-testid="bulk-item-picker"]')!;
	return {
		player,
		cell,
		swapWeapon: () =>
			act(() => {
				worn = WEAPON_AFTER;
				gear = gearOf(worn);
				patchKeyed(store, 'players', STORE_KEY, { gear } as never, ['gear']);
				loadEquippedBulkItems(player);
			}),
		frozen: () => cell().hasAttribute('data-frozen'),
	};
};

describe('BulkItemPicker when the batch entry is swapped for an item the player is wearing', () => {
	it('seeds one editable cell for the batched ring', () => {
		const h = mount();
		expect(bulkState(h.player).pickerGroups.get(BulkSimItemSlot.ItemSlotFinger)).toEqual([{ index: 0, item: BATCHED }]);
		expect(h.view()).toEqual({ equipped: false, removable: true });
	});

	it('replaces the entry item in the store while keeping the same mounted instance', () => {
		const h = mount();
		const before = h.cell();

		act(() => createBulkGearData(h.player, BulkSimItemSlot.ItemSlotFinger, 0).equipItem(WORN));

		expect(bulkState(h.player).pickerGroups.get(BulkSimItemSlot.ItemSlotFinger)).toEqual([{ index: 0, item: WORN }]);
		expect(h.cell()).toBe(before);
	});

	it('marks the cell as equipped and non-removable in the commit that brings the worn item, and keeps it on the next unrelated bulk notification', () => {
		const h = mount();

		act(() => createBulkGearData(h.player, BulkSimItemSlot.ItemSlotFinger, 0).equipItem(WORN));
		expect(h.view()).toEqual({ equipped: true, removable: false });

		act(() => patchBulkState(h.player, {}, ['items']));
		expect(h.view()).toEqual({ equipped: true, removable: false });
	});
});

describe('BulkItemPickerGroup add button', () => {
	it.each([
		['rings', ItemSlot.ItemSlotFinger1],
		['trinkets', ItemSlot.ItemSlotTrinket1],
		['weapons', ItemSlot.ItemSlotMainHand],
		['head', ItemSlot.ItemSlotHead],
	])('opens the item list for the %s group on its slot', (group, itemSlot) => {
		const h = mount();

		h.clickAdd(group);

		expect(h.modal().open).toBe(true);
		expect(h.modal().request).toMatchObject({ slot: itemSlot, tab: SelectorModalTabs.Items });
		expect(h.modal().request!.gearData.getEquippedItem()).toBeNull();
	});

	it('adds each picked item to the group and leaves the modal open', () => {
		const h = mount();
		h.clickAdd('rings');
		const { gearData } = h.modal().request!;

		act(() => gearData.equipItem(PICKED));
		act(() => gearData.equipItem(ENCHANTED));

		const rings = bulkState(h.player).pickerGroups.get(BulkSimItemSlot.ItemSlotFinger)!;
		expect(rings.map(entry => entry.item)).toEqual([BATCHED, PICKED, ENCHANTED]);
		expect(screen.getAllByTestId('bulk-item-picker')).toHaveLength(3);
		expect(h.modal().open).toBe(true);
		expect(gearData.getEquippedItem()).toBeNull();
	});

	it("adds an enchanted pick in the shape the list's in-batch mark looks for", () => {
		const h = mount();
		h.clickAdd('rings');

		act(() => h.modal().request!.gearData.equipItem(ENCHANTED));

		expect(hasBulkItem(h.player, ItemSpec.create({ id: ENCHANTED.id }))).toBe(true);
	});
});

describe('BulkItemPicker on the frozen main-hand cell when the player swaps that weapon', () => {
	it('shows the frozen weapon cell as frozen before the swap', () => {
		const h = mountFrozenWeapon();
		expect(bulkState(h.player).pickerGroups.get(BulkSimItemSlot.ItemSlotHandWeapon)).toEqual([{ index: -1, item: WEAPON_BEFORE }]);
		expect(h.frozen()).toBe(true);
	});

	it('keeps the same mounted instance while the equipped entry item is replaced', () => {
		const h = mountFrozenWeapon();
		const before = h.cell();

		h.swapWeapon();

		expect(bulkState(h.player).pickerGroups.get(BulkSimItemSlot.ItemSlotHandWeapon)).toEqual([{ index: -1, item: WEAPON_AFTER }]);
		expect(h.cell()).toBe(before);
	});

	it('keeps the frozen slot marked frozen in the commit that brings the new weapon, and on the next unrelated bulk notification', () => {
		const h = mountFrozenWeapon();

		h.swapWeapon();
		expect(h.frozen()).toBe(true);

		act(() => patchBulkState(h.player, {}, ['items']));
		expect(h.frozen()).toBe(true);
	});
});

const renderPicker = (state: BulkItemPickerState, onRemove?: () => void) => {
	const onOpen = vi.fn();
	render(<BulkItemPickerView slot={ItemSlot.ItemSlotFinger1} item={BATCHED} state={state} onOpen={onOpen} onRemove={onRemove} />);
	return { onOpen, cell: screen.getByTestId('bulk-item-picker') };
};

describe('BulkItemPickerView', () => {
	it.each<[BulkItemPickerState, boolean, boolean, boolean]>([
		['frozen', true, false, true],
		['equipped', false, true, false],
		['editable', false, false, false],
	])('marks a %s item frozen %s and equipped %s, with a lock %s', (state, frozen, equipped, lock) => {
		const { cell } = renderPicker(state);

		expect(cell.hasAttribute('data-frozen')).toBe(frozen);
		expect(cell.hasAttribute('data-equipped')).toBe(equipped);
		expect(!!screen.queryByTestId('item-picker-frozen')).toBe(lock);
	});

	it('removes the item from its remove button', () => {
		const onRemove = vi.fn();
		renderPicker('editable', onRemove);

		fireEvent.click(screen.getByTestId('item-picker-actions-btn'));

		expect(onRemove).toHaveBeenCalledOnce();
		expect(screen.getByTestId('item-picker-actions-btn').querySelector('.fa-xmark')).not.toBeNull();
	});

	it('offers no remove button for an item that cannot be removed', () => {
		renderPicker('equipped');

		expect(screen.queryByTestId('item-picker-actions-btn')).toBeNull();
	});

	it('opens the selector from the cell', () => {
		const { onOpen, cell } = renderPicker('editable');

		fireEvent.click(cell);

		expect(onOpen).toHaveBeenCalledWith(0);
	});
});
