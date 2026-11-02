import { DesignDowntime, Encounter, EncounterDesign, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter } from '@generated/proto/ui';
import { bulkState, loadStoredBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { MY_ENCOUNTERS_STORAGE_KEY } from '@sim/state/storage_keys';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import {
	bulkEncounterList,
	bulkEncounterRuns,
	CURRENT_ENCOUNTER_REF,
	currentEncounterChecked,
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

const hostWith = (presets: Array<{ name: string; encounter?: Encounter }>, saved: Record<string, Encounter>, mine: Record<string, Encounter> = {}) => {
	const player = makeBulkPlayer();
	const record = Object.fromEntries(Object.entries(saved).map(([name, proto]) => [name, SavedEncounter.toJson(SavedEncounter.create({ encounter: proto }))]));
	player.sim.env.storage.setItem(SAVED_KEY, JSON.stringify(record));
	const myRecord = Object.fromEntries(Object.entries(mine).map(([name, proto]) => [name, Encounter.toJson(proto)]));
	player.sim.env.storage.setItem(`test-${MY_ENCOUNTERS_STORAGE_KEY}`, JSON.stringify(myRecord));
	return {
		player,
		sim: player.sim,
		individualConfig: { presets: { encounters: presets } },
		getStorageKey: (key: string) => `test-${key}`,
		getSavedEncounterStorageKey: () => SAVED_KEY,
	} as unknown as IndividualSimHost<any>;
};

describe('formatEncounterDuration', () => {
	it.each([
		[{ duration: 90, durationVariation: 10 }, '1:30 ±10s'],
		[{ duration: 300, durationVariation: null }, '5:00'],
	])('formats %#', (length, label) => {
		expect(formatEncounterDuration(length)).toBe(label);
	});
});

describe('formatEncounterSummary', () => {
	it.each([
		[encounter(180), '3:00 ±5s · bulk_tab.settings.encounters.targets:1'],
		[encounter(90, 10, 3), '1:30 ±10s · bulk_tab.settings.encounters.targets:3'],
		[encounter(600, 30, 0), '10:00 ±30s · bulk_tab.settings.encounters.targets:1'],
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

	it('lists My encounters between the presets and the saved encounters, leaving out one with no target', () => {
		const host = hostWith(
			[{ name: 'Patchwerk', encounter: encounter(300) }],
			{ Council: encounter(90) },
			{ Council: encounter(240), Empty: encounter(60, 5, 0) },
		);

		expect(pickableBulkEncounters(host).map(({ name, ref, encounter }) => [name, ref.preset, ref.myEncounter, encounter.duration])).toEqual([
			['Patchwerk', true, false, 300],
			['Council', false, true, 240],
			['Council', false, false, 90],
		]);
	});
});

const CURRENT = encounter(180);

const ref = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });

const entryOf = (name: string, proto: Encounter, preset = false) => ({ ref: ref(name, preset), name, encounter: proto });

describe('bulkEncounterList', () => {
	it('lists every preset and saved encounter, each checked by its stored ref, and the current encounter checked on fresh storage', () => {
		const host = hostWith([{ name: 'Patchwerk', encounter: encounter(300) }], { Council: encounter(90), Patchwerk: encounter(120) });
		const stored = bulkState(host.player).encounters;

		const rows = bulkEncounterList(CURRENT, pickableBulkEncounters(host), [...stored, ref('Patchwerk')]);

		expect(currentEncounterChecked(stored)).toBe(true);
		expect(rows.map(row => [row.name, row.ref.preset, row.checked, row.duplicateOf])).toEqual([
			['Patchwerk', true, false, null],
			['Council', false, false, null],
			['Patchwerk', false, true, null],
		]);
	});

	it('checks a saved encounter and an entry of My encounters of one name apart', () => {
		const host = hostWith([], { Council: encounter(90) }, { Council: encounter(240) });
		const mine = BulkEncounterRef.create({ name: 'Council', myEncounter: true });
		const rows = (checked: BulkEncounterRef[]) =>
			bulkEncounterList(CURRENT, pickableBulkEncounters(host), checked).map(row => [row.ref.myEncounter, row.checked, row.duplicateOf]);

		expect(rows([mine])).toEqual([
			[true, true, null],
			[false, false, null],
		]);
		expect(rows([ref('Council')])).toEqual([
			[true, false, null],
			[false, true, null],
		]);
	});

	it('names an earlier checked encounter for a later duplicate, keeping the stored state', () => {
		const entries = [
			entryOf('Raid', encounter(300), true),
			entryOf('Raid copy', encounter(300)),
			entryOf('Raid again', encounter(300)),
			entryOf('Council', encounter(90)),
			entryOf('Council copy', encounter(90)),
		];
		const checked = [CURRENT_ENCOUNTER_REF, ref('Raid', true), ref('Raid copy'), ref('Council copy')];

		expect(bulkEncounterList(CURRENT, entries, checked).map(row => [row.name, row.checked, row.duplicateOf])).toEqual([
			['Raid', true, null],
			['Raid copy', true, 'Raid'],
			['Raid again', false, 'Raid'],
			['Council', false, null],
			['Council copy', true, null],
		]);
		expect(bulkEncounterRuns(CURRENT, entries, checked).map(run => run.name)).toEqual(['', 'Raid', 'Council copy']);
	});

	it('disables an encounter that sims like the current one while that is ticked, and runs it under its own name while it is not', () => {
		const entries = [entryOf('Three minutes', encounter(180)), entryOf('Three minutes copy', encounter(180))];
		const ticked = [CURRENT_ENCOUNTER_REF, ref('Three minutes'), ref('Three minutes copy')];
		const unticked = [ref('Three minutes'), ref('Three minutes copy')];

		expect(bulkEncounterList(CURRENT, entries, ticked).map(row => [row.name, row.checked, row.duplicateOf])).toEqual([
			['Three minutes', true, 'bulk_tab.settings.encounters.current'],
			['Three minutes copy', true, 'bulk_tab.settings.encounters.current'],
		]);
		expect(bulkEncounterRuns(CURRENT, entries, ticked).map(run => run.name)).toEqual(['']);
		expect(bulkEncounterList(CURRENT, entries, unticked).map(row => [row.name, row.checked, row.duplicateOf])).toEqual([
			['Three minutes', true, null],
			['Three minutes copy', true, 'Three minutes'],
		]);
		expect(bulkEncounterRuns(CURRENT, entries, unticked)).toEqual([{ name: 'Three minutes', isCurrent: false, encounter: entries[0].encounter }]);
	});

	it('compares encounters without their proto version, name, My encounters mark, zone, target names or an empty design', () => {
		const entries = [
			entryOf('Raid', Encounter.create({ duration: 300, targets: [Target.create()] })),
			entryOf(
				'Raid elsewhere',
				Encounter.create({
					duration: 300,
					apiVersion: 7,
					name: 'Mine',
					myEncounter: true,
					zoneId: 5,
					design: EncounterDesign.create(),
					targets: [Target.create({ name: 'Boss' })],
				}),
			),
			entryOf('Longer raid', Encounter.create({ duration: 301, targets: [Target.create()] })),
			entryOf(
				'Raid with a downtime',
				Encounter.create({ duration: 300, design: EncounterDesign.create({ downtimes: [DesignDowntime.create()] }), targets: [Target.create()] }),
			),
		];

		expect(
			bulkEncounterList(
				CURRENT,
				entries,
				entries.map(entry => entry.ref),
			).map(row => row.duplicateOf),
		).toEqual([null, 'Raid', null, null]);
	});
});

describe('bulkEncounterRuns', () => {
	it('runs the current encounter first while ticked, then each checked encounter that duplicates none', () => {
		const entries = [entryOf('Raid', encounter(300), true), entryOf('Council', encounter(90)), entryOf('Raid copy', encounter(300))];

		expect(bulkEncounterRuns(CURRENT, entries, [CURRENT_ENCOUNTER_REF, ref('Raid copy'), ref('Raid', true)])).toEqual([
			{ name: '', isCurrent: true, encounter: CURRENT },
			{ name: 'Raid', isCurrent: false, encounter: entries[0].encounter },
		]);
		expect(bulkEncounterRuns(CURRENT, entries, [ref('Council')])).toEqual([{ name: 'Council', isCurrent: false, encounter: entries[1].encounter }]);
	});
});

describe('setBulkEncounterChecked', () => {
	it('stores the checked refs in list order so they load back, the current one included, dropping refs that are not pickable', () => {
		const player = makeBulkPlayer();
		const pickable = [ref('Raid', true), ref('Council')];
		const { settings } = bulkState(player).v;

		setBulkEncounterChecked(player, ref('Council'), true, pickable);
		setBulkEncounterChecked(player, ref('Raid', true), true, pickable);

		expect(loadStoredBulkEncounters(player)?.encounters).toEqual([CURRENT_ENCOUNTER_REF, ref('Raid', true), ref('Council')]);
		expect(bulkState(player).encounters).toEqual([CURRENT_ENCOUNTER_REF, ref('Raid', true), ref('Council')]);
		expect(bulkState(player).v.settings).toBe(settings);

		setBulkEncounterChecked(player, CURRENT_ENCOUNTER_REF, false, [ref('Raid', true), ref('Other')]);

		expect(bulkState(player).encounters).toEqual([ref('Raid', true)]);
		expect(currentEncounterChecked(loadStoredBulkEncounters(player)!.encounters)).toBe(false);
	});

	it('stores an entry of My encounters under its own ref, leaving a saved encounter of its name alone', () => {
		const player = makeBulkPlayer();
		const mine = BulkEncounterRef.create({ name: 'Council', myEncounter: true });

		setBulkEncounterChecked(player, mine, true, [mine, ref('Council')]);

		expect(loadStoredBulkEncounters(player)?.encounters).toEqual([CURRENT_ENCOUNTER_REF, mine]);
	});
});
