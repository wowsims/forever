import { Stat } from '@generated/proto/common';
import { SavedEPWeights } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { createSimStore, patchKeyed, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { chipNamed, chips, nameInput, readStored, renderReady, saveButton, source } from '@testing/saved_panel';
import { fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SavedEpWeights } from './SavedEpWeights';

vi.mock('@sim/state/subscriptions', async () =>
	(await import('@testing/sim/subscriptions')).mockSubscriptions((await import('@testing/saved_panel')).source.subscribe),
);

const STRINGS = vi.hoisted(
	() =>
		({
			'sidebar.buttons.stat_weights.modal.ep': 'EP',
			'sidebar.buttons.stat_weights.title': 'Stat Weights',
			'sidebar.buttons.stat_weights.saved_ep_weights.title': 'Saved EP weights',
			'common.name': 'Name',
		}) as Record<string, string>,
);
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => STRINGS[key] ?? key } }));

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const STORAGE_KEY = 'tbc-warrior-savedEPWeights';

const weights = (agility: number) => new Stats().withStat(Stat.StatAgility, agility);
const storedJson = (agility: number) => SavedEPWeights.toJson(SavedEPWeights.create({ epWeights: weights(agility).toProto() }));
const MIXED = new Stats().withStat(Stat.StatAgility, 4).withStat(Stat.StatCritRating, 1.75).withStat(Stat.StatStrength, 2);
const MIXED_JSON = SavedEPWeights.toJson(SavedEPWeights.create({ epWeights: MIXED.toProto() }));

class FakePlayer {
	storeKey = 0;
	sim = { store: createSimStore() };

	constructor() {
		seedKeyed(this.sim.store, 'players', this.storeKey, { epWeights: new Stats(), v: zeroVersions(PLAYER_FIELDS) } as never);
	}

	get epWeights(): Stats {
		return this.sim.store.getState().players[this.storeKey].epWeights;
	}
	set epWeights(next: Stats) {
		patchKeyed(this.sim.store, 'players', this.storeKey, { epWeights: next }, ['epWeights']);
	}

	getEpWeights() {
		return this.epWeights;
	}
	setEpWeights = vi.fn((next: Stats) => {
		this.epWeights = next;
	});
}

let player: FakePlayer;
let host: any;
let presets: Array<any>;

const setup = () => {
	source.listeners.clear();
	trackEvent.mockClear();
	window.localStorage.clear();
	player = new FakePlayer();
	presets = [];
	host = {
		player,
		sim: { waitForInit: () => Promise.resolve() },
		get individualConfig() {
			return { presets: { epWeights: presets } };
		},
		getSavedEPWeightsStorageKey: () => STORAGE_KEY,
	};
};

const renderManager = () => renderReady(<SavedEpWeights />, host);
const stored = () => readStored(STORAGE_KEY);

beforeEach(setup);

describe('SavedEpWeights', () => {
	describe('saving', () => {
		it('writes the current weights under the typed name and shows them as a chip', async () => {
			player.epWeights = weights(5);
			await renderManager();

			fireEvent.change(nameInput(), { target: { value: 'Raiding' } });
			fireEvent.click(saveButton());

			expect(chips('custom').map(chip => chip.textContent)).toEqual(['Raiding']);
			expect(stored()).toEqual({ Raiding: storedJson(5) });
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'save', label: 'EP' });
		});

		it('marks the freshly saved set active, because it matches the current weights', async () => {
			player.epWeights = weights(5);
			await renderManager();

			fireEvent.change(nameInput(), { target: { value: 'Raiding' } });
			fireEvent.click(saveButton());

			expect(chipNamed('Raiding').hasAttribute('data-active')).toBe(true);
		});
	});

	describe('loading', () => {
		it('restores a stored set and reports it', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Raiding: storedJson(4) }));
			await renderManager();

			fireEvent.click(chipNamed('Raiding').querySelector('[data-testid="saved-data-set-name"]')!);

			expect(player.setEpWeights).toHaveBeenCalledTimes(1);
			expect(player.getEpWeights().equals(weights(4))).toBe(true);
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'load', label: 'EP' });
		});

		it("runs a preset's onLoad", async () => {
			const onLoad = vi.fn();
			presets = [{ name: 'Default', epWeights: weights(3), onLoad }];
			await renderManager();

			fireEvent.click(chipNamed('Default').querySelector('[data-testid="saved-data-set-name"]')!);

			expect(onLoad).toHaveBeenCalledWith(player);
		});

		it('leaves the clicked set in the name input even when a preset holds the same weights', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Mine: storedJson(3) }));
			presets = [{ name: 'Default', epWeights: weights(3) }];
			await renderManager();

			fireEvent.click(chipNamed('Mine').querySelector('[data-testid="saved-data-set-name"]')!);

			expect(nameInput().value).toBe('Mine');
		});

		it('marks every set holding the current weights active, not just the one that named the input', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Mine: storedJson(3) }));
			presets = [{ name: 'Default', epWeights: weights(3) }];
			player.epWeights = weights(3);
			await renderManager();

			expect(chipNamed('Mine').hasAttribute('data-active')).toBe(true);
			expect(chipNamed('Default').hasAttribute('data-active')).toBe(true);
		});

		it('stamps the name input with the set the current weights match', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Raiding: storedJson(4) }));
			player.epWeights = weights(4);
			await renderManager();

			expect(nameInput().value).toBe('Raiding');
			expect(chipNamed('Raiding').hasAttribute('data-active')).toBe(true);
		});

		// The only part of this that can lose data, and no gate covers it: an entry saved under the old
		// storage key has to keep loading. Every stat, not just the one the rest of these tests carry.
		it('reads a set stored under the key the vanilla manager wrote', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Raiding: MIXED_JSON }));
			await renderManager();

			expect(chips('custom').map(chip => chip.textContent)).toEqual(['Raiding']);
			fireEvent.click(chipNamed('Raiding').querySelector('[data-testid="saved-data-set-name"]')!);
			expect(player.getEpWeights().equals(MIXED)).toBe(true);
		});

		it('writes a set back in that same shape', async () => {
			player.epWeights = MIXED;
			await renderManager();
			fireEvent.change(nameInput(), { target: { value: 'Raiding' } });
			fireEvent.click(saveButton());

			expect(stored()).toEqual({ Raiding: MIXED_JSON });
		});
	});

	describe('presets', () => {
		it('hides a preset whose enableWhen is false', async () => {
			presets = [
				{ name: 'On', epWeights: weights(1), enableWhen: () => true },
				{ name: 'Off', epWeights: weights(2), enableWhen: () => false },
			];
			await renderManager();

			expect(chipNamed('On').hasAttribute('data-disabled')).toBe(false);
			expect(chipNamed('Off').hasAttribute('data-disabled')).toBe(true);
		});

		it('renders each section only once it has a chip', async () => {
			presets = [{ name: 'Default', epWeights: weights(3) }];
			await renderManager();

			expect(document.querySelector('[data-testid="saved-data-presets"]')).toBeTruthy();
			expect(document.querySelector('[data-testid="saved-data-custom"]')).toBeNull();
		});
	});
});
