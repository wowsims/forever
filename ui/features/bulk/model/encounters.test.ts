import { Encounter, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter } from '@generated/proto/ui';
import { bulkState, loadStoredBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import {
	bulkEncounterList,
	bulkEncounterRuns,
	formatEncounterDuration,
	formatEncounterSummary,
	pickableBulkEncounters,
	setBulkEncounterChecked,
} from './encounters';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string, options?: { count?: number }) => (options?.count === undefined ? key : `${key}:${options.count}`) },
}));

const SAVED_KEY = 'saved-encounters';

const encounter = (duration: number, durationVariation = 5, targets = 1) =>
	Encounter.create({ duration, durationVariation, targets: Array.from({ length: targets }, () => Target.create()) });

const ref = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });

const entry = (name: string, proto: Encounter, preset = false) => ({ ref: ref(name, preset), name, encounter: proto });

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
	const checkedAll = (entries: ReturnType<typeof entry>[]) => entries.map(item => item.ref);

	it('runs the current encounter first, then each checked one in list order', () => {
		const entries = [entry('Council', encounter(90)), entry('Unchecked', encounter(120)), entry('Patchwerk', encounter(300))];
		const runs = bulkEncounterRuns(CURRENT, entries, [ref('Patchwerk'), ref('Council')], true);

		expect(runs.map(run => [run.name, run.isCurrent, run.encounter.duration])).toEqual([
			['', true, 180],
			['Council', false, 90],
			['Patchwerk', false, 300],
		]);
	});

	it('skips a checked one equal to the current encounter or to an earlier checked one', () => {
		const entries = [
			entry('Same as current', encounter(180)),
			entry('Council', encounter(90)),
			entry('Council again', encounter(90)),
			entry('Council, more variance', encounter(90, 20)),
		];
		const runs = bulkEncounterRuns(CURRENT, entries, checkedAll(entries), true);

		expect(runs.map(run => run.name)).toEqual(['', 'Council', 'Council, more variance']);
	});

	it('compares encounters without their api version', () => {
		const entries = [entry('Same as current', encounter(180))];

		expect(bulkEncounterRuns({ ...CURRENT, apiVersion: 3 }, entries, checkedAll(entries), true)).toHaveLength(1);
	});

	it('runs the current encounter alone when nothing is checked', () => {
		expect(bulkEncounterRuns(CURRENT, [entry('Council', encounter(90))], [], true)).toEqual([{ name: '', encounter: CURRENT, isCurrent: true }]);
	});

	it('runs the current encounter alone with power tools off, whatever is checked', () => {
		const entries = [entry('Council', encounter(90))];

		expect(bulkEncounterRuns(CURRENT, entries, checkedAll(entries), false)).toEqual([{ name: '', encounter: CURRENT, isCurrent: true }]);
	});
});

describe('formatEncounterDuration', () => {
	it.each([
		[{ duration: 90, durationVariation: 10 }, '1:30 ±10s'],
		[{ duration: 300, durationVariation: null }, '5:00'],
		[{ duration: null, durationVariation: null }, 'bulk_tab.settings.encounters.health_summary'],
	])('formats %#', (length, label) => {
		expect(formatEncounterDuration(length)).toBe(label);
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
	it('leaves out a preset or a saved encounter with no target', () => {
		const host = hostWith(
			[
				{ name: 'No preset target', encounter: encounter(300, 5, 0) },
				{ name: 'Patchwerk', encounter: encounter(300) },
			],
			{ 'No saved target': encounter(90, 5, 0), Council: encounter(90) },
		);

		expect(pickableBulkEncounters(host).map(entry => entry.name)).toEqual(['Patchwerk', 'Council']);
	});

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

		expect(bulkEncounterRuns(CURRENT, pickableBulkEncounters(host), [ref('Gone'), ref('Council')], true).map(run => run.name)).toEqual(['', 'Council']);
	});

	it('tells a preset from a saved encounter of the same name', () => {
		const host = hostWith([{ name: 'Patchwerk', encounter: encounter(300) }], { Patchwerk: encounter(120) });

		expect(bulkEncounterRuns(CURRENT, pickableBulkEncounters(host), [ref('Patchwerk')], true).map(run => run.encounter.duration)).toEqual([180, 120]);
	});
});

describe('bulkEncounterList', () => {
	const listed = (entries: ReturnType<typeof entry>[], checked: BulkEncounterRef[]) =>
		bulkEncounterList(CURRENT, entries, checked).map(item => [item.name, item.checked, item.duplicate]);

	it('leaves out an entry equal to the current encounter, checked or not', () => {
		const entries = [
			entry('Same as current', encounter(180)),
			entry('Council', encounter(90)),
			entry('Also current', { ...encounter(180), apiVersion: 2 }),
		];

		expect(listed(entries, [ref('Same as current')])).toEqual([['Council', false, false]]);
	});

	it('lists a checked entry equal to an earlier checked one and flags it', () => {
		const entries = [entry('Patchwerk', encounter(300), true), entry('Patchwerk again', encounter(300)), entry('Council', encounter(90))];

		expect(listed(entries, [ref('Patchwerk', true), ref('Patchwerk again'), ref('Council')])).toEqual([
			['Patchwerk', true, false],
			['Patchwerk again', true, true],
			['Council', true, false],
		]);
	});

	it('flags only checked entries, against earlier checked ones', () => {
		const entries = [entry('Patchwerk', encounter(300)), entry('Patchwerk again', encounter(300))];

		expect(listed(entries, [ref('Patchwerk again')])).toEqual([
			['Patchwerk', false, false],
			['Patchwerk again', true, false],
		]);
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
