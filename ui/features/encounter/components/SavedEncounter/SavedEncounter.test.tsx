import { Encounter as EncounterProto } from '@generated/proto/common';
import { SavedEncounter as SavedEncounterProto } from '@generated/proto/ui';
import { createSimStore } from '@sim/state/sim_store';
import { chipNamed, chips, nameInput, readStored, renderReady, saveButton, source } from '@testing/saved_panel';
import { fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SavedEncounter } from './SavedEncounter';

vi.mock('@sim/state/subscriptions', async () =>
	(await import('@testing/sim/subscriptions')).mockSubscriptions((await import('@testing/saved_panel')).source.subscribe),
);

const STRINGS = vi.hoisted(
	() =>
		({
			'settings_tab.saved_encounters.encounter': 'Encounter',
			'settings_tab.saved_encounters.title': 'Saved Encounters',
			'settings_tab.saved_encounters.encounter_name': 'Encounter Name',
			'settings_tab.saved_encounters.save_encounter': 'Save Encounter',
			'common.name': 'Name',
		}) as Record<string, string>,
);
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => STRINGS[key] ?? key } }));

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const STORAGE_KEY = 'tbc-encounter-saved';

const encounterWith = (duration: number) => EncounterProto.create({ duration });
const storedJson = (duration: number) => SavedEncounterProto.toJson(SavedEncounterProto.create({ encounter: encounterWith(duration) }));

class FakeEncounter {
	proto = encounterWith(180);
	toProto() {
		return this.proto;
	}
	fromProto = vi.fn((proto: EncounterProto) => {
		this.proto = proto;
		source.notify();
	});
}

let encounter: FakeEncounter;
let host: any;
let presets: Array<any>;

const setup = () => {
	source.listeners.clear();
	trackEvent.mockClear();
	window.localStorage.clear();
	encounter = new FakeEncounter();
	presets = [];
	host = {
		sim: { store: createSimStore(), waitForInit: () => Promise.resolve(), encounter },
		get individualConfig() {
			return { presets: { encounters: presets } };
		},
		getSavedEncounterStorageKey: () => STORAGE_KEY,
	};
};

const renderPanel = () => renderReady(<SavedEncounter />, host);
const stored = () => readStored(STORAGE_KEY);

beforeEach(setup);

describe('SavedEncounter', () => {
	describe('saving', () => {
		it('writes the current encounter under the typed name and shows it as a chip', async () => {
			encounter.proto = encounterWith(500);
			await renderPanel();

			fireEvent.change(nameInput(), { target: { value: 'Council' } });
			fireEvent.click(saveButton());

			expect(chips('custom').map(chip => chip.textContent)).toEqual(['Council']);
			expect(stored()).toEqual({ Council: storedJson(500) });
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'save', label: 'Encounter' });
		});
	});

	describe('loading', () => {
		it('restores a stored set and reports it', async () => {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Council: storedJson(400) }));
			await renderPanel();

			fireEvent.click(chipNamed('Council').querySelector('[data-testid="saved-data-set-name"]')!);

			expect(encounter.fromProto).toHaveBeenCalledTimes(1);
			expect(encounter.fromProto.mock.calls[0]![0].duration).toBe(400);
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'load', label: 'Encounter' });
		});

		it("shows a preset's tooltip on its anchor", async () => {
			presets = [{ name: 'Council', encounter: encounterWith(400), tooltip: 'Tank swaps at 3s' }];
			await renderPanel();

			expect(chipNamed('Council').querySelector('[data-testid="saved-data-set-name"]')!.getAttribute('data-tooltip-content')).toBe('Tank swaps at 3s');
		});

		// Encounter presets carry only `tooltip`, never `enableWhen`/`onLoad`. Adding either back here
		// would silently diverge from that.
		it('never disables an encounter preset', async () => {
			presets = [{ name: 'Council', encounter: encounterWith(400) }];
			await renderPanel();

			expect(chipNamed('Council').hasAttribute('data-disabled')).toBe(false);
		});
	});
});
