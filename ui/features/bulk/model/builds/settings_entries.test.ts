import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel, ItemSpec, ItemSwap, Profession, Race, Stat } from '@generated/proto/common';
import { Mage } from '@generated/proto/mage';
import { SavedSettings } from '@generated/proto/ui';
import { DpsWarrior } from '@generated/proto/warrior';
import { withLiveSpecOptions } from '@sim/proto/spec_functions';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { pickableBulkSettings, settingsContentKey } from './settings_entries';

// The stub host has none of the getters the live settings are read from, so an item swap preset reads them as the
// defaults.
vi.mock('@features/settings/model/saved_settings', async importOriginal => {
	const { SavedSettings: Settings } = await import('@generated/proto/ui');
	return { ...(await importOriginal<object>()), readSavedSettings: () => Settings.create() };
});

const SETTINGS_KEY = 'saved-settings';

const HEALER = SavedSettings.create({ race: Race.RaceDwarf, reactionTimeMs: 250 });
const LIVE = SavedSettings.create({ race: Race.RaceOrc, professions: [Profession.Alchemy, Profession.Herbalism] });
const SWAP = ItemSwap.create({ items: [ItemSpec.create({ id: 19019 })] });

// A preset settings entry "Tank", a preset item swap "Thunderfury", and the saved entries given.
const hostWith = (savedSettings: Record<string, unknown> = { Healer: SavedSettings.toJson(HEALER) }) => {
	const player = makeBulkPlayer();
	player.sim.env.storage.setItem(SETTINGS_KEY, JSON.stringify(savedSettings));
	return {
		player,
		sim: player.sim,
		individualConfig: {
			presets: {
				settings: [{ name: 'Tank', race: Race.RaceHuman, playerOptions: { inFrontOfTarget: true } }],
				itemSwaps: [{ name: 'Thunderfury', itemSwap: SWAP }],
			},
		},
		getSavedSettingsStorageKey: () => SETTINGS_KEY,
	} as unknown as IndividualSimHost<any>;
};

// HEALER as the Settings tab reads live settings: every message set and the item swap's slots all empty.
const HEALER_AS_READ = SavedSettings.create({
	...HEALER,
	raidBuffs: RaidBuffs.create(),
	partyBuffs: PartyBuffs.create(),
	playerBuffs: IndividualBuffs.create(),
	debuffs: Debuffs.create(),
	consumables: ConsumesSpec.create(),
	healingModel: HealingModel.create(),
	itemSwap: ItemSwap.create({ items: [ItemSpec.create(), ItemSpec.create()], prepullBonusStats: new Stats().toProto() }),
});

describe('settings entries', () => {
	it('lists the spec’s settings and item swap presets, then the saved entries', () => {
		const entries = pickableBulkSettings(hostWith());

		expect(entries.map(entry => [entry.name, entry.ref.name, entry.ref.preset])).toEqual([
			['Tank', 'Tank', true],
			['Thunderfury', 'Thunderfury', true],
			['Healer', 'Healer', false],
		]);
		expect(entries[0].settings).toEqual(expect.objectContaining({ race: Race.RaceHuman, inFrontOfTarget: true }));
		expect(entries[1].settings).toEqual(expect.objectContaining({ enableItemSwap: true, itemSwap: SWAP }));
		expect(entries[2].settings).toEqual(HEALER);
	});

	it('compares two entries by what the sim reads, not by their names', () => {
		expect(settingsContentKey(HEALER_AS_READ)).toBe(settingsContentKey(HEALER));
		expect(settingsContentKey(SavedSettings.create({ ...LIVE, darkIntentUptime: 50, challengeMode: true }))).toBe(settingsContentKey(LIVE));
		expect(settingsContentKey(SavedSettings.create({ ...LIVE, healingModel: HealingModel.create({ hps: 500 }) }))).not.toBe(settingsContentKey(LIVE));
		expect(settingsContentKey(HEALER)).not.toBe(settingsContentKey(LIVE));
	});

	it('tells two entries apart by their spec options, and reads one without them, or with another spec’s, with the live ones', () => {
		const withRage = (startingRage: number) =>
			SavedSettings.create({
				...LIVE,
				spec: { oneofKind: 'dpsWarrior', dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage } } }) },
			});

		expect(settingsContentKey(withRage(50))).not.toBe(settingsContentKey(withRage(0)));
		expect(settingsContentKey(LIVE)).not.toBe(settingsContentKey(withRage(50)));
		expect(settingsContentKey(withLiveSpecOptions(LIVE, withRage(50)))).toBe(settingsContentKey(withRage(50)));
		expect(settingsContentKey(withLiveSpecOptions(withRage(0), withRage(50)))).toBe(settingsContentKey(withRage(0)));
		const mageEntry = SavedSettings.create({ ...LIVE, spec: { oneofKind: 'mage', mage: Mage.create({ options: {} }) } });
		expect(settingsContentKey(withLiveSpecOptions(mageEntry, withRage(50)))).toBe(settingsContentKey(withRage(50)));
	});

	it('reads an item swap that is off as none, so live settings holding swap items match an entry without them', () => {
		const swapItems = ItemSwap.create({
			items: [ItemSpec.create({ id: 19019 })],
			prepullBonusStats: new Stats().withStat(Stat.StatStrength, 20).toProto(),
		});
		const live = (enableItemSwap: boolean) => SavedSettings.create({ ...LIVE, enableItemSwap, itemSwap: swapItems });
		const defaults = SavedSettings.create({ ...LIVE, darkIntentUptime: 50 });

		expect([false, true].map(on => settingsContentKey(live(on)) === settingsContentKey(defaults))).toEqual([true, false]);
	});

	it('reads a saved settings entry holding fields the protos do not list', () => {
		const settings = pickableBulkSettings(hostWith({ Healer: { race: 'RaceDwarf', reactionTimeMs: 250, retiredField: 3 } }));

		expect(settings.find(entry => entry.name === 'Healer')?.settings).toEqual(HEALER);
	});
});
