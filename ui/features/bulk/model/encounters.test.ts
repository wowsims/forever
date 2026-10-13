import { Encounter, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter } from '@generated/proto/ui';
import { bulkState, loadStoredBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import {
	bulkEncounterRuns,
	checkedBulkEncounters,
	formatEncounterSummary,
	isDuplicateEncounter,
	pickableBulkEncounters,
	setBulkEncounterChecked,
} from './encounters';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string, options?: { count?: number }) => (options?.count === undefined ? key : `${key}:${options.count}`) },
}));

const SAVED_KEY = 'saved-encounters';

const encounter = (duration: number, durationVariation = 5, targets = 1) =>
	Encounter.create({ duration, durationVariation, targets: Array.from({ length: targets }, () => Target.create()) });

const named = (name: string, proto: Encounter) => ({ name, encounter: proto });

const CURRENT = encounter(180);

const hostWith = (presets: Array<{ name: string; encounter?: Encounter }>, saved: Record<string, Encounter>) => {
	const player = makeBulkPlayer();
	const record = Object.fromEntries(Object.entries(saved).map(([name, proto]) => [name, SavedEncounter.toJson(SavedEncounter.create({ encounter: proto }))]));
	player.sim.env.storage.setItem(SAVED_KEY, JSON.stringify(record));
	return {
		player,
		sim: player.sim,
		individualConfig: { presets: { encounters: presets } },
		getSavedEncounterStorageKey: () => SAVED_KEY,
	} as unknown as IndividualSimHost<any>;
};

describe('bulkEncounterRuns', () => {
	it('runs the current encounter first, then each picked one in order', () => {
		const runs = bulkEncounterRuns(CURRENT, [named('Council', encounter(90)), named('Patchwerk', encounter(300))]);

		expect(runs.map(run => [run.name, run.isCurrent, run.encounter.duration])).toEqual([
			['', true, 180],
			['Council', false, 90],
			['Patchwerk', false, 300],
		]);
	});

	it('skips a pick equal to the current encounter or to an earlier pick', () => {
		const runs = bulkEncounterRuns(CURRENT, [
			named('Same as current', encounter(180)),
			named('Council', encounter(90)),
			named('Council again', encounter(90)),
			named('Council, more variance', encounter(90, 20)),
		]);

		expect(runs.map(run => run.name)).toEqual(['', 'Council', 'Council, more variance']);
	});

	it('compares encounters without their api version', () => {
		const runs = bulkEncounterRuns({ ...CURRENT, apiVersion: 3 }, [named('Same as current', encounter(180))]);

		expect(runs).toHaveLength(1);
	});

	it('runs the current encounter alone when nothing is picked', () => {
		expect(bulkEncounterRuns(CURRENT, [])).toEqual([{ name: '', encounter: CURRENT, isCurrent: true }]);
	});
});

describe('isDuplicateEncounter', () => {
	it('flags a pick equal to the current encounter or to an earlier pick only', () => {
		const picked = [encounter(180), encounter(90), encounter(90), encounter(90, 5, 2)];

		expect(picked.map((_, index) => isDuplicateEncounter(CURRENT, picked, index))).toEqual([true, false, true, false]);
	});
});

describe('formatEncounterSummary', () => {
	it.each([
		[encounter(180), '3:00 ±5s · bulk_tab.settings.encounters.targets:1'],
		[encounter(90, 10, 3), '1:30 ±10s · bulk_tab.settings.encounters.targets:3'],
		[encounter(600, 30, 0), '10:00 ±30s · bulk_tab.settings.encounters.targets:1'],
		[{ ...encounter(180, 5, 2), useHealth: true }, 'bulk_tab.settings.encounters.health_summary · bulk_tab.settings.encounters.targets:2'],
	])('formats %#', (proto, summary) => {
		expect(formatEncounterSummary(proto)).toBe(summary);
	});
});

describe('pickableBulkEncounters', () => {
	it('lists the presets, then the saved encounters, and leaves out a preset with no encounter', () => {
		const host = hostWith(
			[
				{ name: 'Patchwerk', encounter: encounter(300) },
				{ name: 'Empty', encounter: undefined },
			],
			{ Council: encounter(90), Patchwerk: encounter(120) },
		);

		expect(pickableBulkEncounters(host).map(entry => [entry.name, entry.ref.preset, entry.encounter.duration])).toEqual([
			['Patchwerk', true, 300],
			['Council', false, 90],
			['Patchwerk', false, 120],
		]);
	});

	it('ignores a checked ref whose encounter was deleted', () => {
		const host = hostWith([], { Council: encounter(90) });
		const checked = [BulkEncounterRef.create({ name: 'Gone', preset: false }), BulkEncounterRef.create({ name: 'Council', preset: false })];

		expect(checkedBulkEncounters(pickableBulkEncounters(host), checked).map(entry => entry.name)).toEqual(['Council']);
	});

	it('tells a preset from a saved encounter of the same name', () => {
		const host = hostWith([{ name: 'Patchwerk', encounter: encounter(300) }], { Patchwerk: encounter(120) });
		const checked = [BulkEncounterRef.create({ name: 'Patchwerk', preset: false })];

		expect(checkedBulkEncounters(pickableBulkEncounters(host), checked).map(entry => entry.encounter.duration)).toEqual([120]);
	});
});

describe('setBulkEncounterChecked', () => {
	const PRESET = BulkEncounterRef.create({ name: 'Patchwerk', preset: true });
	const COUNCIL = BulkEncounterRef.create({ name: 'Council', preset: false });
	const GONE = BulkEncounterRef.create({ name: 'Gone', preset: false });

	it('stores the checked refs in list order so they load back', () => {
		const player = makeBulkPlayer();
		setBulkEncounterChecked(player, COUNCIL, true, [PRESET, COUNCIL]);
		setBulkEncounterChecked(player, PRESET, true, [PRESET, COUNCIL]);

		expect(bulkState(player).encounters).toEqual([PRESET, COUNCIL]);
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual([PRESET, COUNCIL]);
	});

	it('unchecks a ref and drops the refs no longer listed', () => {
		const player = makeBulkPlayer();
		setBulkEncounterChecked(player, GONE, true, [GONE, PRESET, COUNCIL]);
		setBulkEncounterChecked(player, COUNCIL, true, [GONE, PRESET, COUNCIL]);
		setBulkEncounterChecked(player, COUNCIL, false, [PRESET, COUNCIL]);

		expect(bulkState(player).encounters).toEqual([]);
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual([]);
	});
});
