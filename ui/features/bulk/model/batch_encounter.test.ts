import { Encounter, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter } from '@generated/proto/ui';
import { bulkState, loadStoredBatchEncounter } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { batchEncounterEntry, batchEncounterSetup, setBatchEncounter } from './batch_encounter';
import { bulkEncounterEntries, CURRENT_ENCOUNTER_REF } from './builds/encounters';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string) => key },
}));

const encounter = (duration: number) => Encounter.create({ duration, durationVariation: 5, targets: [Target.create()] });

const PATCHWERK = encounter(180);
const COUNCIL = encounter(240);
const ENTRIES = bulkEncounterEntries(
	[{ name: 'Patchwerk', encounter: PATCHWERK }],
	[{ name: 'Patchwerk', data: SavedEncounter.create({ encounter: COUNCIL }) }],
);

const ref = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });

describe('batchEncounterEntry', () => {
	it('reads the empty ref as Current encounter', () => expect(batchEncounterEntry(ENTRIES, CURRENT_ENCOUNTER_REF)).toBeUndefined());
	it('tells a preset from a saved encounter of the same name', () => {
		expect(batchEncounterEntry(ENTRIES, ref('Patchwerk', true))?.encounter).toEqual(PATCHWERK);
		expect(batchEncounterEntry(ENTRIES, ref('Patchwerk'))?.encounter).toEqual(COUNCIL);
	});
	it('reads a ref no entry holds as Current encounter', () => expect(batchEncounterEntry(ENTRIES, ref('Gone'))).toBeUndefined());
});

describe('setBatchEncounter', () => {
	it('stores the pick per spec without bumping the settings counter', () => {
		const player = makeBulkPlayer();
		setBatchEncounter(player, ref('Council'));
		expect(bulkState(player).batchEncounter).toEqual(ref('Council'));
		expect(loadStoredBatchEncounter(player)).toEqual(ref('Council'));
		expect(bulkState(player).v.settings).toBe(0);
	});
});

describe('batchEncounterSetup', () => {
	const SAVED_KEY = 'saved-encounters';
	const current = encounter(60);
	const hostWith = (saved: Record<string, Encounter>) => {
		const player = makeBulkPlayer();
		const record = Object.fromEntries(
			Object.entries(saved).map(([name, proto]) => [name, SavedEncounter.toJson(SavedEncounter.create({ encounter: proto }))]),
		);
		player.sim.env.storage.setItem(SAVED_KEY, JSON.stringify(record));
		return {
			player,
			sim: Object.assign(player.sim, { encounter: { toProto: () => current } }),
			individualConfig: { presets: { encounters: [{ name: 'Patchwerk', encounter: PATCHWERK }] } },
			getSavedEncounterStorageKey: () => SAVED_KEY,
		} as unknown as IndividualSimHost<any>;
	};

	it('runs the picked encounter under its name', () => {
		const host = hostWith({});
		setBatchEncounter(host.player, ref('Patchwerk', true));
		expect(batchEncounterSetup(host)).toEqual({ name: 'Patchwerk', isCurrent: false, encounter: PATCHWERK });
	});

	it('runs the Encounter tab on Current encounter or when the pick is gone, and the pick returns once it is saved again', () => {
		const host = hostWith({});
		expect(batchEncounterSetup(host)).toEqual({ name: '', isCurrent: true, encounter: current });
		setBatchEncounter(host.player, ref('Council'));
		expect(batchEncounterSetup(host)).toEqual({ name: '', isCurrent: true, encounter: current });
		expect(bulkState(host.player).batchEncounter).toEqual(ref('Council'));

		const again = hostWith({ Council: COUNCIL });
		setBatchEncounter(again.player, ref('Council'));
		expect(batchEncounterSetup(again)).toEqual({ name: 'Council', isCurrent: false, encounter: COUNCIL });
	});
});
