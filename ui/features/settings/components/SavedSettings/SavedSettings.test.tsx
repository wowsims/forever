import { ItemSpec, ItemSwap, Race, Spec } from '@generated/proto/common';
import { SavedSettings as SavedSettingsProto } from '@generated/proto/ui';
import { DpsWarrior, DpsWarrior_Options } from '@generated/proto/warrior';
import { specTypeFunctions } from '@sim/proto/spec_functions';
import { createSimStore } from '@sim/state/sim_store';
import { chipNamed, chips, nameInput, readStored, renderReady, saveButton, source } from '@testing/saved_panel';
import { act, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SavedSettings } from './SavedSettings';

vi.mock('@sim/state/subscriptions', async () =>
	(await import('@testing/sim/subscriptions')).mockSubscriptions((await import('@testing/saved_panel')).source.subscribe),
);

const savedSettingsModel = vi.hoisted(() => ({
	readSavedSettings: vi.fn(),
	applySavedSettings: vi.fn(),
}));
vi.mock('@features/settings/model/saved_settings', () => savedSettingsModel);

const STRINGS = vi.hoisted(
	() =>
		({
			'settings_tab.saved_settings.settings': 'Settings',
			'settings_tab.saved_settings.title': 'Saved Settings',
			'settings_tab.saved_settings.settings_name': 'Settings Name',
			'settings_tab.saved_settings.save_settings': 'Save Settings',
			'common.name': 'Name',
		}) as Record<string, string>,
);
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => STRINGS[key] ?? key } }));

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const STORAGE_KEY = 'mop-warrior-savedSettings';

const settingsWith = (race: Race) => SavedSettingsProto.create({ race });
const storedJson = (race: Race) => SavedSettingsProto.toJson(settingsWith(race));
const rageSpec = (startingRage: number): SavedSettingsProto['spec'] => ({
	oneofKind: 'dpsWarrior',
	dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage } } }),
});

let host: any;
let presets: Array<any>;
let itemSwapPresets: Array<any>;

const setup = () => {
	source.listeners.clear();
	trackEvent.mockClear();
	window.localStorage.clear();
	presets = [];
	itemSwapPresets = [];

	let current = settingsWith(Race.RaceHuman);
	savedSettingsModel.readSavedSettings.mockImplementation(() => current);
	savedSettingsModel.applySavedSettings.mockImplementation((_h: unknown, settings: SavedSettingsProto) => {
		current = settings;
		source.notify();
	});

	let config: { presets: { settings: Array<any>; itemSwaps: Array<any> } } | undefined;
	host = {
		player: { getParty: () => ({}), getSpec: () => Spec.SpecDpsWarrior, specTypeFunctions: specTypeFunctions[Spec.SpecDpsWarrior] },
		sim: { store: createSimStore(), waitForInit: () => Promise.resolve(), raid: {} },
		get individualConfig() {
			config ??= { presets: { settings: presets, itemSwaps: itemSwapPresets } };
			return config;
		},
		getSavedSettingsStorageKey: () => STORAGE_KEY,
	};
};

const renderPanel = () => renderReady(<SavedSettings />, host);
const stored = () => readStored(STORAGE_KEY);

beforeEach(setup);

describe('SavedSettings', () => {
	describe('saving', () => {
		it('writes the current settings under the typed name and shows it as a chip', async () => {
			savedSettingsModel.readSavedSettings.mockReturnValue(settingsWith(Race.RaceOrc));
			await renderPanel();

			fireEvent.change(nameInput(), { target: { value: 'Raiding' } });
			fireEvent.click(saveButton());

			expect(chips('custom').map(chip => chip.textContent)).toEqual(['Raiding']);
			expect(stored()).toEqual({ Raiding: storedJson(Race.RaceOrc) });
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'save', label: 'Settings' });
		});
	});

	describe('loading', () => {
		it('applies a stored set through applySavedSettings and reports it', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Raiding: storedJson(Race.RaceOrc) }));
			await renderPanel();

			fireEvent.click(chipNamed('Raiding').querySelector('[data-testid="saved-data-set-name"]')!);

			expect(savedSettingsModel.applySavedSettings).toHaveBeenCalledTimes(1);
			expect(savedSettingsModel.applySavedSettings.mock.calls[0]![1].race).toBe(Race.RaceOrc);
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'load', label: 'Settings' });
		});

		it("shows a settings preset's tooltip", async () => {
			presets = [{ name: 'Council', race: Race.RaceOrc, tooltip: 'For the council fight' }];
			await renderPanel();

			expect(chipNamed('Council').querySelector('[data-testid="saved-data-set-name"]')!.getAttribute('data-tooltip-content')).toBe(
				'For the council fight',
			);
		});

		it("loads a settings preset's item swap set with it", async () => {
			const itemSwap = ItemSwap.create({ items: [ItemSpec.create({ id: 15257 })] });
			presets = [{ name: 'Swap', playerOptions: { enableItemSwap: true, itemSwap } }];
			await renderPanel();

			fireEvent.click(chipNamed('Swap').querySelector('[data-testid="saved-data-set-name"]')!);

			const loaded = savedSettingsModel.applySavedSettings.mock.lastCall![1];
			expect(loaded.enableItemSwap).toBe(true);
			expect(loaded.itemSwap).toEqual(itemSwap);
		});

		it("loads a settings preset's spec options with it, and none for a preset without them", async () => {
			presets = [
				{ name: 'Rage', specOptions: DpsWarrior_Options.create({ classOptions: { startingRage: 50 } }) },
				{ name: 'Plain', race: Race.RaceOrc },
			];
			await renderPanel();

			fireEvent.click(chipNamed('Rage').querySelector('[data-testid="saved-data-set-name"]')!);
			expect(savedSettingsModel.applySavedSettings.mock.lastCall![1].spec).toEqual({
				oneofKind: 'dpsWarrior',
				dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage: 50 } } }),
			});

			fireEvent.click(chipNamed('Plain').querySelector('[data-testid="saved-data-set-name"]')!);
			expect(savedSettingsModel.applySavedSettings.mock.lastCall![1].spec.oneofKind).toBeUndefined();
		});

		// The item-swap preset holds the settings read when the presets are built, without spec options. A later race
		// change drops it from active, a spec option toggle does not, and loading it hands over no spec options.
		it("bakes the item-swap preset's settings at build time without spec options, which loading it leaves alone", async () => {
			const swapSettings = (race: Race, startingRage: number) =>
				SavedSettingsProto.create({ race, enableItemSwap: true, itemSwap: ItemSwap.create(), spec: rageSpec(startingRage) });
			let live = swapSettings(Race.RaceHuman, 0);
			savedSettingsModel.readSavedSettings.mockImplementation(() => live);
			savedSettingsModel.applySavedSettings.mockImplementation((_h: unknown, next: SavedSettingsProto) => {
				live = next;
				source.notify();
			});
			itemSwapPresets = [{ name: 'Council Swap', itemSwap: ItemSwap.create() }];
			await renderPanel();

			expect(chipNamed('Council Swap').hasAttribute('data-active')).toBe(true);

			await act(async () => {
				savedSettingsModel.applySavedSettings(host, swapSettings(Race.RaceHuman, 50));
			});
			expect(chipNamed('Council Swap').hasAttribute('data-active')).toBe(true);

			fireEvent.click(chipNamed('Council Swap').querySelector('[data-testid="saved-data-set-name"]')!);
			expect(savedSettingsModel.applySavedSettings.mock.lastCall![1].spec.oneofKind).toBeUndefined();

			await act(async () => {
				savedSettingsModel.applySavedSettings(host, swapSettings(Race.RaceOrc, 50));
			});
			expect(chipNamed('Council Swap').hasAttribute('data-active')).toBe(false);
		});
	});

	describe('active entry', () => {
		const storedWithRage = (startingRage: number) =>
			SavedSettingsProto.toJson(SavedSettingsProto.create({ race: Race.RaceHuman, spec: rageSpec(startingRage) }));

		beforeEach(() => {
			savedSettingsModel.readSavedSettings.mockImplementation(() => SavedSettingsProto.create({ race: Race.RaceHuman, spec: rageSpec(50) }));
		});

		it('shows an entry saved without spec options as active when the rest matches, and fills the name input with it', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Old: storedJson(Race.RaceHuman), 'No Rage': storedWithRage(0) }));
			await renderPanel();

			expect(chipNamed('Old').hasAttribute('data-active')).toBe(true);
			expect(chipNamed('No Rage').hasAttribute('data-active')).toBe(false);
			expect(nameInput().value).toBe('Old');
		});

		it('shows an entry with the live spec options as active, and a settings preset without spec options', async () => {
			presets = [{ name: 'Plain', race: Race.RaceHuman }];
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Rage: storedWithRage(50) }));
			await renderPanel();

			expect(chipNamed('Rage').hasAttribute('data-active')).toBe(true);
			expect(chipNamed('Plain').hasAttribute('data-active')).toBe(true);
		});
	});
});
