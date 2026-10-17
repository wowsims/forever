import { Encounter, ItemSlot, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter } from '@generated/proto/ui';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState, loadStoredBulkEncounters, patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, patchSlice, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { memoryEnv } from '@testing/sim/memory_env';
import { act, fireEvent, render, screen } from '@testing-library/react';
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

const SAVED_ENCOUNTERS_KEY = 'test-saved-encounters';
const CURRENT = { duration: 180, durationVariation: 5, targets: [Target.create()] };

interface MountOptions {
	reforger?: unknown;
	presets?: Array<{ name: string; encounter: Encounter }>;
	saved?: Record<string, Encounter>;
	checked?: BulkEncounterRef[];
	powerTools?: boolean;
}

const mount = ({ reforger = null, presets = [], saved = {}, checked = [], powerTools = true }: MountOptions = {}) => {
	const record = Object.entries(saved).map(([name, encounter]) => [name, SavedEncounter.toJson(SavedEncounter.create({ encounter }))]);
	localStorage.setItem(SAVED_ENCOUNTERS_KEY, JSON.stringify(Object.fromEntries(record)));
	const store = createSimStore();
	patchSlice(store, 'encounter', CURRENT);
	patchSlice(store, 'ui', { showPowerTools: powerTools });
	const gear = { getEquippedItem: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotFinger1 ? RING : null) };
	seedKeyed(store, 'players', STORE_KEY, { gear, v: zeroVersions(PLAYER_FIELDS) } as never);
	const player = {
		sim: { store, isNative: false, env: memoryEnv() },
		storeKey: STORE_KEY,
		getGear: () => gear,
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, { encounters: checked });
	const host = {
		player,
		sim: player.sim,
		reforger,
		reforgeOptions: null,
		individualConfig: { presets: { encounters: presets } },
		getStorageKey: (key: string) => `test-${key}`,
		getSavedEncounterStorageKey: () => SAVED_ENCOUNTERS_KEY,
	} as never;
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

	it('shows only the encounter, enchant and frozen slot groups, even for a spec with the gear optimizer', () => {
		const { container } = mount({ reforger: { settings: {} } });

		const triggers = [...container.querySelectorAll('[data-testid^="bulk-settings-group-"][data-testid$="-trigger"]')];

		expect(triggers.map(trigger => trigger.getAttribute('data-testid'))).toEqual([
			'bulk-settings-group-encounters-trigger',
			'bulk-settings-group-enchants-trigger',
			'bulk-settings-group-freezes-trigger',
		]);
	});
});

describe('BulkSettings encounters group', () => {
	const PATCHWERK = Encounter.create({ duration: 300, durationVariation: 30, targets: [Target.create()] });
	const COUNCIL = Encounter.create({ duration: 240, durationVariation: 10, targets: [Target.create(), Target.create(), Target.create()] });
	const SAME_AS_CURRENT = Encounter.create({ ...createSimStore().getState().encounter, ...CURRENT });
	const ref = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });

	const openEncounters = (options: MountOptions = {}) => {
		const mounted = mount(options);
		mounted.openGroup('encounters');
		return mounted;
	};
	const rows = () => screen.queryAllByTestId('bulk-encounter');
	const box = (name: string) => screen.getByRole('checkbox', { name });
	// The texts a checkbox is described by: the encounter summary, then the duplicate reason when there is one.
	const descriptions = (name: string) => (box(name).getAttribute('aria-describedby') ?? '').split(' ').map(id => document.getElementById(id)?.textContent);
	const hasLabelTooltip = (name: string) => screen.getByText(name, { selector: 'label' }).hasAttribute('data-tooltip-id');

	it('lists the presets, then the saved encounters, under their headings, each with its summary as a tooltip', () => {
		openEncounters({ presets: [{ name: 'Patchwerk', encounter: PATCHWERK }], saved: { Council: COUNCIL } });

		expect(rows()).toHaveLength(2);
		expect(rows().map(row => row.querySelector('label')?.textContent)).toEqual(['Patchwerk', 'Council']);
		expect(descriptions('Patchwerk')).toEqual(['5:00 ±30s · bulk_tab.settings.encounters.targets']);
		expect(descriptions('Council')).toEqual(['4:00 ±10s · bulk_tab.settings.encounters.targets']);
		expect(hasLabelTooltip('Patchwerk')).toBe(true);
		expect(hasLabelTooltip('Council')).toBe(true);
		expect(screen.getByText('bulk_tab.settings.encounters.presets')).toBeTruthy();
		expect(screen.getByText('bulk_tab.settings.encounters.saved')).toBeTruthy();
		expect(screen.queryByText('bulk_tab.settings.encounters.empty')).toBeNull();
	});

	it('omits the heading of an empty list', () => {
		openEncounters({ saved: { Council: COUNCIL } });

		expect(screen.queryByText('bulk_tab.settings.encounters.presets')).toBeNull();
		expect(screen.getByText('bulk_tab.settings.encounters.saved')).toBeTruthy();
	});

	it('shows the current encounter as a ticked, locked row that follows the live encounter', () => {
		const { player } = openEncounters();
		const name = 'bulk_tab.settings.encounters.current';
		const current = box(name);

		expect(current.getAttribute('aria-checked')).toBe('true');
		expect(current.getAttribute('aria-disabled')).toBe('true');
		expect(hasLabelTooltip(name)).toBe(true);
		expect(descriptions(name)).toEqual(['3:00 ±5s · bulk_tab.settings.encounters.targets']);

		act(() => patchSlice(player.sim.store, 'encounter', { duration: 200, useHealth: true }));

		expect(descriptions(name)).toEqual(['bulk_tab.settings.encounters.health_summary · bulk_tab.settings.encounters.targets']);
	});

	it('stores the checked refs in list order', () => {
		const { player } = openEncounters({ presets: [{ name: 'Patchwerk', encounter: PATCHWERK }], saved: { Council: COUNCIL } });

		fireEvent.click(box('Council'));
		fireEvent.click(box('Patchwerk'));

		const expected = [ref('Patchwerk', true), ref('Council')];
		expect(bulkState(player).encounters).toEqual(expected);
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual(expected);
		expect(box('Council').getAttribute('aria-checked')).toBe('true');
	});

	it('hides an encounter equal to the current one and keeps its stored state', () => {
		const { player } = openEncounters({ saved: { 'Same as current': SAME_AS_CURRENT, Council: COUNCIL }, checked: [ref('Same as current')] });

		expect(rows()).toHaveLength(1);
		expect(screen.queryByRole('checkbox', { name: 'Same as current' })).toBeNull();

		fireEvent.click(box('Council'));

		expect(bulkState(player).encounters).toEqual([ref('Same as current'), ref('Council')]);
	});

	it('marks the checkbox of a checked encounter equal to an earlier checked one', () => {
		openEncounters({
			presets: [{ name: 'Patchwerk', encounter: PATCHWERK }],
			saved: { 'Patchwerk again': PATCHWERK, Council: COUNCIL },
			checked: [ref('Patchwerk', true), ref('Patchwerk again')],
		});

		expect(['Patchwerk', 'Patchwerk again', 'Council'].map(name => box(name).getAttribute('aria-invalid'))).toEqual([null, 'true', null]);
		expect(descriptions('Patchwerk again')).toEqual(['5:00 ±30s · bulk_tab.settings.encounters.targets', 'bulk_tab.settings.encounters.duplicate']);
		expect(descriptions('Council')).toEqual(['4:00 ±10s · bulk_tab.settings.encounters.targets']);
		expect(descriptions('Patchwerk')).toEqual(['5:00 ±30s · bulk_tab.settings.encounters.targets']);
	});

	it('points to the saved encounters when there is nothing to pick', () => {
		openEncounters();

		expect(rows()).toHaveLength(0);
		expect(screen.getByText('bulk_tab.settings.encounters.empty')).toBeTruthy();
		expect(screen.getByTestId('bulk-encounter-current')).toBeTruthy();
	});

	it('counts every encounter the batch will run', () => {
		mount({ presets: [{ name: 'Patchwerk', encounter: PATCHWERK }], saved: { Council: COUNCIL }, checked: [ref('Patchwerk', true), ref('Council')] });

		expect(screen.getByTestId('bulk-encounters-count').textContent).toBe('× 3');
	});

	it('hides the group and counts the current encounter alone with power tools off, keeping the checked ones', () => {
		const checked = [ref('Patchwerk', true), ref('Council')];
		const presets = [{ name: 'Patchwerk', encounter: PATCHWERK }];
		const { player, container } = mount({ presets, saved: { Council: COUNCIL }, checked, powerTools: false });
		const trigger = () => container.querySelector('[data-testid="bulk-settings-group-encounters-trigger"]');

		expect(trigger()).toBeNull();
		expect(screen.queryByTestId('bulk-encounters-count')).toBeNull();
		expect(bulkState(player).encounters).toEqual(checked);

		act(() => patchSlice(player.sim.store, 'ui', { showPowerTools: true }));

		expect(trigger()).not.toBeNull();
		expect(screen.getByTestId('bulk-encounters-count').textContent).toBe('× 3');
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
			encounters={null}
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
