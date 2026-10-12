import { ItemSlot } from '@generated/proto/common';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { BulkSettingsViewProps } from './BulkSettings';

vi.mock('@sim/bulk/utils', async () => {
	const actual = await vi.importActual<typeof import('@sim/bulk/utils')>('@sim/bulk/utils');
	return { ...actual, getBulkPlayerCanDualWield: () => true, getBulkFreezeWeaponTypes: () => [] };
});
vi.mock('../../model/run', () => ({ runBulkBatch: vi.fn() }));

const { BulkSettings, BulkSettingsView } = await import('./BulkSettings');

const STORE_KEY = 5;
const RING = { id: 1, equals: () => false } as unknown as EquippedItem;

const mount = (reforger: unknown = null) => {
	const store = createSimStore();
	const gear = { getEquippedItem: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotFinger1 ? RING : null) };
	seedKeyed(store, 'players', STORE_KEY, { gear, v: zeroVersions(PLAYER_FIELDS) } as never);
	const player = { sim: { store, isNative: false }, storeKey: STORE_KEY, getGear: () => gear } as unknown as Player<any>;
	seedBulkSettings(player);
	const host = { player, sim: player.sim, reforger, reforgeOptions: null, getStorageKey: (key: string) => `test-${key}` } as never;
	const { container } = render(
		<SimHostProvider host={host}>
			<BulkSettings />
		</SimHostProvider>,
	);
	const pick = async (id: string, index: number) => {
		const trigger = container.querySelector<HTMLElement>(`#${id}`)!;
		await userEvent.click(trigger);
		const option = screen.getAllByRole('option')[index];
		const label = option.textContent;
		await userEvent.click(option);
		return [trigger, label] as const;
	};
	const openGroup = (group: string) => fireEvent.click(container.querySelector(`[data-testid="bulk-settings-group-${group}-trigger"]`)!);
	return { player, container, pick, openGroup };
};

beforeEach(() => {
	localStorage.clear();
});

describe('BulkSettings', () => {
	// Both selects list None, then the pair's first slot (Finger 1, Main Hand), then its second.
	it.each<[string, (player: Player<any>) => unknown, unknown]>([
		['freeze-ring', player => bulkState(player).frozenItems.get(BulkSimItemSlot.ItemSlotFinger), RING],
		['freeze-weapon', player => bulkState(player).frozenWeaponSlot, ItemSlot.ItemSlotMainHand],
	])('keeps the picked slot in the %s select', async (id, stored, expected) => {
		const { player, pick, openGroup } = mount();
		openGroup('freezes');

		const [select, picked] = await pick(id, 1);

		expect(stored(player)).toBe(expected);
		expect(select.textContent).toBe(picked);
	});

	it('opens no group before anything is stored', () => {
		const { container } = mount();

		expect(container.querySelector('#freeze-ring')).toBeNull();
	});

	it('remembers an opened group in local storage', () => {
		const { openGroup } = mount();

		openGroup('freezes');

		expect(JSON.parse(localStorage.getItem('test-__bulkSettingsGroups')!)).toEqual(['freezes']);
	});

	it('shows only the durations, enchant and frozen slot groups, even for a spec with the gear optimizer', () => {
		const { container } = mount({ settings: {} });

		const triggers = [...container.querySelectorAll('[data-testid^="bulk-settings-group-"][data-testid$="-trigger"]')];

		expect(triggers.map(trigger => trigger.getAttribute('data-testid'))).toEqual([
			'bulk-settings-group-durations-trigger',
			'bulk-settings-group-enchants-trigger',
			'bulk-settings-group-freezes-trigger',
		]);
	});
});

const RING_VIEW = {
	id: 'freeze-ring',
	label: 'Freeze ring',
	options: [
		{ value: -1, label: 'None' },
		{ value: ItemSlot.ItemSlotFinger1, label: 'Finger 1' },
	],
	value: -1,
	onChange: () => {},
};

const renderSettings = (props: Partial<BulkSettingsViewProps> = {}) => {
	const handlers = { onRun: vi.fn(), onOpenGroupsChange: vi.fn(), onFrozenWeaponSlotChange: vi.fn() };
	const { container } = render(
		<BulkSettingsView
			combinationsCount={<span data-testid="count" />}
			canRun
			openGroups={['freezes']}
			durations={null}
			enchants={null}
			frozenItems={[RING_VIEW]}
			weaponFreeze={null}
			{...handlers}
			{...props}
		/>,
	);
	return { container, ...handlers };
};

describe('BulkSettingsView', () => {
	it('runs the batch from its button, which is disabled while it cannot run', () => {
		const { onRun } = renderSettings();

		fireEvent.click(screen.getByTestId('bulk-settings-btn'));
		expect(onRun).toHaveBeenCalledOnce();
		expect(screen.getByTestId('count')).toBeTruthy();
	});

	it('disables the run button while the batch cannot run', () => {
		renderSettings({ canRun: false });

		expect((screen.getByTestId('bulk-settings-btn') as HTMLButtonElement).disabled).toBe(true);
	});

	it('renders one freeze select per paired slot', () => {
		const { container } = renderSettings();

		expect(container.querySelector('#freeze-ring')).not.toBeNull();
		expect(container.querySelector('#freeze-weapon')).toBeNull();
	});

	it('offers the weapon freeze and its weapon type filters to a dual wielder', async () => {
		const { container, onFrozenWeaponSlotChange } = renderSettings({
			weaponFreeze: { value: -1, weaponTypes: <span data-testid="weapon-types" /> },
		});

		await userEvent.click(container.querySelector<HTMLElement>('#freeze-weapon')!);
		await userEvent.click(screen.getAllByRole('option')[1]);

		expect(onFrozenWeaponSlotChange).toHaveBeenCalledWith(ItemSlot.ItemSlotMainHand);
		expect(screen.getByTestId('weapon-types')).toBeTruthy();
	});
});
