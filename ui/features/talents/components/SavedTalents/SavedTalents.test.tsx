import { SavedTalents as SavedTalentsProto } from '@generated/proto/ui';
import { createSimStore, patchKeyed, PLAYER_FIELDS, type PlayerSlice, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SavedTalents } from './SavedTalents';

const KEY = 6;

const STRINGS = vi.hoisted(
	() =>
		({
			'talents_tab.saved_talents.label': 'Talents',
			'talents_tab.saved_talents.title': 'Saved Talents',
			'talents_tab.saved_talents.name_label': 'Name',
			'talents_tab.saved_talents.save_button': 'Save Talents',
			'talents_tab.saved_talents.delete.tooltip': 'Delete saved talents',
			'talents_tab.saved_talents.delete.confirm': "Delete saved talents '{{name}}'?",
			'talents_tab.saved_talents.alerts.choose_name': 'Choose a label for your saved talents!',
			'talents_tab.saved_talents.alerts.name_exists': 'Talents with name {{name}} already exists.',
			'common.name': 'Name',
		}) as Record<string, string>,
);
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => STRINGS[key] ?? key } }));

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const STORAGE_KEY = 'tbc-warrior-savedTalents';

const talentsWith = (talentsString: string) => SavedTalentsProto.create({ talentsString });
const storedJson = (talentsString: string) => SavedTalentsProto.toJson(talentsWith(talentsString));

class FakePlayer {
	readonly storeKey = KEY;
	readonly sim: { store: SimStore };

	constructor() {
		const store = createSimStore();
		seedKeyed(store, 'players', KEY, { talentsString: '', v: zeroVersions(PLAYER_FIELDS) } as unknown as PlayerSlice);
		this.sim = { store };
	}

	private slice() {
		return this.sim.store.getState().players[KEY];
	}

	get talentsString() {
		return this.slice().talentsString;
	}
	set talentsString(next: string) {
		patchKeyed(this.sim.store, 'players', KEY, { talentsString: next }, ['talentsString']);
	}
	getTalentsString() {
		return this.talentsString;
	}
	setTalentsString = vi.fn((next: string) => {
		this.talentsString = next;
	});
}

let player: FakePlayer;
let host: any;
let presets: Array<any>;
let waitForInit: () => Promise<void>;

const setup = () => {
	trackEvent.mockClear();
	window.localStorage.clear();
	player = new FakePlayer();
	presets = [];
	waitForInit = () => Promise.resolve();
	host = {
		player,
		sim: { store: player.sim.store, waitForInit: () => waitForInit() },
		get individualConfig() {
			return { presets: { talents: presets } };
		},
		getSavedTalentsStorageKey: () => STORAGE_KEY,
	};
};

const renderPanel = async () => {
	const result = renderWithHost(<SavedTalents />, host);
	await act(async () => {});
	return result;
};

const nameInput = () => document.querySelector<HTMLInputElement>('[data-testid="saved-data-save-input"]')!;
const saveButton = () => document.querySelector<HTMLButtonElement>('[data-testid="saved-data-save-button"]')!;
const chips = (section: 'presets' | 'custom') => [
	...document.querySelectorAll(`:is(.saved-data-${section}, [data-testid="saved-data-${section}"]) [data-testid="saved-data-set-chip"]`),
];
const chipNamed = (name: string) => [...document.querySelectorAll('[data-testid="saved-data-set-chip"]')].find(chip => chip.textContent?.startsWith(name))!;
const stored = () => JSON.parse(window.localStorage.getItem(STORAGE_KEY) ?? 'null');
const popover = () => document.querySelector('[data-testid="sim-confirm-popover"]');
const popoverText = () => popover()?.querySelector('[data-testid="sim-confirm-popover-message"]')?.textContent;
const popoverButtons = () => [...(popover()?.querySelectorAll<HTMLButtonElement>('[data-testid="sim-confirm-popover-actions"] button') ?? [])];

beforeEach(setup);

describe('SavedTalents', () => {
	describe('saving', () => {
		it('writes the current talents under the typed name and shows it as a chip', async () => {
			player.talentsString = '12321';
			await renderPanel();

			fireEvent.change(nameInput(), { target: { value: 'Raiding' } });
			fireEvent.click(saveButton());

			expect(chips('custom').map(chip => chip.textContent)).toEqual(['Raiding']);
			expect(stored()).toEqual({ Raiding: storedJson('12321') });
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'save', label: 'Talents' });
		});
	});

	describe('loading', () => {
		it('restores a stored set and reports it', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Raiding: storedJson('111') }));
			await renderPanel();

			fireEvent.click(chipNamed('Raiding').querySelector('[data-testid="saved-data-set-name"]')!);

			expect(player.setTalentsString).toHaveBeenCalledWith('111');
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'load', label: 'Talents' });
		});

		it("runs a preset's onLoad", async () => {
			const onLoad = vi.fn();
			presets = [{ name: 'Default', data: talentsWith('222'), onLoad }];
			await renderPanel();

			fireEvent.click(chipNamed('Default').querySelector('[data-testid="saved-data-set-name"]')!);
			expect(onLoad).toHaveBeenCalledWith(player);
		});
	});

	describe('deleting', () => {
		it('removes the set and rewrites storage once confirmed', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Raiding: storedJson('111'), Other: storedJson('222') }));
			await renderPanel();

			fireEvent.click(chipNamed('Raiding').querySelector('[data-testid="saved-data-set-delete"]')!);
			expect(popoverText()).toBe("Delete saved talents 'Raiding'?");
			const [, confirm] = popoverButtons();
			fireEvent.click(confirm);

			expect(chips('custom').map(chip => chip.textContent)).toEqual(['Other']);
			expect(stored()).toEqual({ Other: storedJson('222') });
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'delete', label: 'Talents' });
		});
	});
});
