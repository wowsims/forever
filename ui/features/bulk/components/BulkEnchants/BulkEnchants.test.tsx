import { EnchantType, ItemSlot } from '@generated/proto/common';
import { DatabaseFilters, UIEnchant } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: 'https://www.wowhead.com/classic/spell=1', ready: true }),
}));
vi.mock('@sim/hooks/useActionIdWowheadDataset', () => ({ useActionIdWowheadDataset: () => ({}) }));
vi.mock('@sim/proto/enchants', () => ({ getEnchantDescription: async (enchant: UIEnchant) => enchant.name }));

const { BulkEnchants } = await import('./BulkEnchants');
const { saveBulkEnchantPreset } = await import('../../model/enchants');

const AGILITY = UIEnchant.create({ effectId: 849, spellId: 13882, name: 'Lesser Agility', lvl: 45 });
const GREATER_RESISTANCE = UIEnchant.create({ effectId: 1888, itemId: 6, name: 'Greater Resistance', lvl: 45 });
const LESSER_IMPACT = UIEnchant.create({ effectId: 1897, itemId: 2, name: 'Lesser Impact', lvl: 30 });
const SUPERIOR_STRIKING = UIEnchant.create({ effectId: 1897, itemId: 3, name: 'Superior Striking', lvl: 60 });

const SHIELD_SPIKE = UIEnchant.create({ effectId: 1704, itemId: 4, name: 'Thorium Shield Spike', enchantType: EnchantType.EnchantTypeShield });
const HEALING = UIEnchant.create({ effectId: 2566, itemId: 5, name: 'Healing Power' });

// The game database lists every weapon enchant, shield ones included, under both hands.
const ENCHANTS = new Map<ItemSlot, UIEnchant[]>([
	[ItemSlot.ItemSlotBack, [AGILITY, HEALING, GREATER_RESISTANCE]],
	[ItemSlot.ItemSlotMainHand, [LESSER_IMPACT, SUPERIOR_STRIKING, SHIELD_SPIKE]],
	[ItemSlot.ItemSlotOffHand, [LESSER_IMPACT, SUPERIOR_STRIKING, SHIELD_SPIKE]],
]);

// `loaded: false` stands for the page before the game database arrives, when `getEnchants` cannot run.
const mount = ({ loaded = true } = {}) => {
	const storage = new Map<string, string>();
	const store = createSimStore();
	const player = {
		storeKey: 1,
		sim: {
			store,
			waitForInit: () => (loaded ? Promise.resolve() : new Promise(() => {})),
			getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
			setFilters: (filters: DatabaseFilters) => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
			env: {
				storage: {
					getItem: (key: string) => storage.get(key) ?? null,
					setItem: (key: string, value: string) => storage.set(key, value),
					removeItem: (key: string) => storage.delete(key),
				},
			},
		},
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		getEnchants: (slot: ItemSlot) => {
			if (!loaded) throw new TypeError("Cannot read properties of null (reading 'getEnchants')");
			return ENCHANTS.get(slot) ?? [];
		},
		isEnchantRelevant: (enchant: UIEnchant) => enchant !== HEALING,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	render(
		<SimHostProvider host={{ player, sim: player.sim } as never}>
			<BulkEnchants />
		</SimHostProvider>,
	);
	return player;
};

const nameInput = () => screen.getByLabelText('bulk_tab.settings.enchant_presets.name') as HTMLInputElement;
const saveButton = () => screen.getByTestId('bulk-enchant-preset-save') as HTMLButtonElement;
const openNewPreset = () => userEvent.click(screen.getByRole('button', { name: 'bulk_tab.settings.enchant_presets.new' }));
const expandSlots = async () => {
	for (const trigger of screen.getAllByTestId('bulk-enchant-slot-trigger')) await userEvent.click(trigger);
};
const searchField = () => screen.getByTestId('bulk-enchant-search');
const showAllToggle = () => screen.getByRole('checkbox', { name: 'bulk_tab.settings.enchant_presets.show_all' });
const tileNames = (slot: HTMLElement) =>
	within(slot)
		.getAllByRole('checkbox')
		.map(tile => tile.textContent);

describe('BulkEnchants', () => {
	it('creates a preset through the dialog, and a tick puts it in the batch', async () => {
		const player = mount();

		await openNewPreset();
		await userEvent.type(nameInput(), 'Beasts');
		await expandSlots();
		await userEvent.click(screen.getByRole('checkbox', { name: /Lesser Agility/ }));
		await userEvent.click(saveButton());

		expect(bulkState(player).enchantPresets.map(preset => [preset.name, preset.checked])).toEqual([['Beasts', false]]);
		expect(bulkState(player).enchantPresets[0].slots).toEqual([{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY.effectId] }]);

		fireEvent.click(screen.getByRole('checkbox', { name: 'Beasts' }));

		expect(bulkState(player).enchantPresets[0].checked).toBe(true);
	});

	it('shows a tile per row of an effect id, and ticks them together', async () => {
		const player = mount();

		await openNewPreset();
		await userEvent.type(nameInput(), 'Weapons');
		await expandSlots();
		await userEvent.click(screen.getByRole('checkbox', { name: /Superior Striking/ }));

		expect(screen.getByRole('checkbox', { name: /Lesser Impact/ }).getAttribute('aria-checked')).toBe('true');

		await userEvent.click(saveButton());

		expect(bulkState(player).enchantPresets[0].slots).toEqual([{ slot: ItemSlot.ItemSlotMainHand, effectIds: [1897] }]);
	});

	it('opens the slots with ticked enchants when editing, each with its tick count', async () => {
		const player = mount();
		saveBulkEnchantPreset(player, { name: 'Undead', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY.effectId] }] });

		await userEvent.click(await screen.findByRole('button', { name: /enchant_presets\.edit/ }));

		expect(screen.getByRole('checkbox', { name: /Lesser Agility/ }).getAttribute('aria-checked')).toBe('true');
		expect(screen.queryByRole('checkbox', { name: /Superior Striking/ })).toBeNull();
		expect(screen.getAllByTestId('bulk-enchant-slot-count').map(pill => pill.textContent)).toEqual(['1']);
	});

	it('narrows the slots to the enchants a search matches, opened, and restores the slots opened by hand when cleared', async () => {
		mount();

		await openNewPreset();
		await userEvent.click(screen.getAllByTestId('bulk-enchant-slot-trigger')[0]);
		await userEvent.type(searchField(), 'strik');

		expect(screen.getByRole('checkbox', { name: /Superior Striking/ })).toBeTruthy();
		expect(screen.queryByRole('checkbox', { name: /Lesser Impact/ })).toBeNull();
		expect(screen.getAllByTestId('bulk-enchant-slot')).toHaveLength(1);

		await userEvent.clear(searchField());

		expect(screen.getAllByTestId('bulk-enchant-slot')).toHaveLength(3);
		expect(screen.queryByRole('checkbox', { name: /Superior Striking/ })).toBeNull();
		expect(screen.getByRole('checkbox', { name: /Lesser Agility/ })).toBeTruthy();
	});

	it('matches every search term on its own, in any order', async () => {
		mount();

		await openNewPreset();
		await userEvent.type(searchField(), 'striking superior');

		expect(tileNames(screen.getByTestId('bulk-enchant-slot'))).toEqual(['Superior Striking']);
	});

	it('keeps a slot closed by hand during a search closed until the search changes', async () => {
		mount();

		await openNewPreset();
		await userEvent.type(searchField(), 'superior');
		await userEvent.click(screen.getByTestId('bulk-enchant-slot-trigger'));

		expect(screen.queryByRole('checkbox', { name: /Superior Striking/ })).toBeNull();

		await userEvent.type(searchField(), ' strik');

		expect(screen.getByRole('checkbox', { name: /Superior Striking/ })).toBeTruthy();
	});

	it('opens the slots Show all brings into a search', async () => {
		mount();

		await openNewPreset();
		await userEvent.type(searchField(), 'healing');

		expect(screen.queryAllByTestId('bulk-enchant-slot')).toHaveLength(0);

		await userEvent.click(showAllToggle());

		expect(screen.getByRole('checkbox', { name: /Healing Power/ })).toBeTruthy();
	});

	it('says so when a search matches no enchant', async () => {
		mount();

		await openNewPreset();
		await userEvent.type(searchField(), 'mongoose');

		expect(screen.queryAllByTestId('bulk-enchant-slot')).toHaveLength(0);
		expect(screen.getByText('bulk_tab.search.no_results')).toBeTruthy();
	});

	it('groups weapon enchants under Weapons and shield enchants under the non-weapon off hand only', async () => {
		mount();

		await openNewPreset();
		await expandSlots();

		const [, weapons, offHand] = screen.getAllByTestId('bulk-enchant-slot');
		expect(within(weapons).getByText('bulk_tab.settings.enchant_presets.slot_weapons')).toBeTruthy();
		expect(within(weapons).queryByRole('checkbox', { name: /Shield Spike/ })).toBeNull();
		expect(within(weapons).getByRole('checkbox', { name: /Superior Striking/ })).toBeTruthy();
		expect(within(offHand).getByText('bulk_tab.settings.enchant_presets.slot_off_hand')).toBeTruthy();
		expect(tileNames(offHand)).toEqual(['Thorium Shield Spike']);
	});

	it('lists each slot highest lvl first, the higher effect id first on a tie', async () => {
		mount();

		await openNewPreset();
		await expandSlots();

		const [back, weapons] = screen.getAllByTestId('bulk-enchant-slot');
		expect(tileNames(back)).toEqual(['Greater Resistance', 'Lesser Agility']);
		expect(tileNames(weapons)).toEqual(['Superior Striking', 'Lesser Impact']);
	});

	it('hides enchants worth nothing to the spec until Show all is on, which it saves to the gear picker’s filter', async () => {
		const player = mount();

		await openNewPreset();
		await expandSlots();

		expect(screen.queryByRole('checkbox', { name: /Healing Power/ })).toBeNull();
		expect(screen.getByRole('checkbox', { name: /Lesser Agility/ })).toBeTruthy();

		await userEvent.click(showAllToggle());

		expect(screen.getByRole('checkbox', { name: /Healing Power/ })).toBeTruthy();
		expect(player.sim.getFilters().showAllEnchants).toBe(true);
	});

	it('lists every enchant when the gear picker’s Show all filter is saved on', async () => {
		const player = mount();
		act(() => player.sim.setFilters(DatabaseFilters.create({ showAllEnchants: true })));

		await openNewPreset();
		await expandSlots();

		expect(showAllToggle().getAttribute('aria-checked')).toBe('true');
		expect(screen.getByRole('checkbox', { name: /Healing Power/ })).toBeTruthy();
	});

	it('keeps a ticked enchant visible even when it is worth nothing to the spec', async () => {
		const player = mount();
		saveBulkEnchantPreset(player, { name: 'Healer', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [HEALING.effectId] }] });

		await userEvent.click(await screen.findByRole('button', { name: /enchant_presets\.edit/ }));

		expect(screen.getByRole('checkbox', { name: /Healing Power/ }).getAttribute('aria-checked')).toBe('true');
	});

	it('renders the presets before the game database has loaded', () => {
		mount({ loaded: false });

		expect(screen.getByRole('button', { name: 'bulk_tab.settings.enchant_presets.new' })).toBeTruthy();
	});

	it('disables Save while the name is taken by another preset', async () => {
		const player = mount();
		saveBulkEnchantPreset(player, { name: 'Undead', slots: [] });

		await openNewPreset();
		await userEvent.type(nameInput(), 'Undead');

		expect(saveButton().disabled).toBe(true);
		expect(nameInput().getAttribute('aria-invalid')).toBe('true');
		expect(screen.getByText('bulk_tab.settings.enchant_presets.name_duplicate')).toBeTruthy();

		await userEvent.type(nameInput(), ' 2');

		expect(saveButton().disabled).toBe(false);
	});
});
