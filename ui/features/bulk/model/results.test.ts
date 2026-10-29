import { applySavedSettings } from '@features/settings/model/saved_settings';
import { DistributionMetrics } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Race, Stat } from '@generated/proto/common';
import { SavedSettings } from '@generated/proto/ui';
import type { BulkGearSetup, BulkPlayerSettingsSetup, BulkRankedResult, BulkRotationSetup, BulkTalentSetup, TopGearResult } from '@sim/bulk/types';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { compareGearSet, CURRENT_ROTATION_SETUP, CURRENT_SETTINGS_SETUP, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { type BulkFinishedRequest, bulkResultKey, equipBulkResult, rankEncounterResults } from './results';

vi.mock('@features/settings/model/saved_settings', () => ({ applySavedSettings: vi.fn() }));

const ARMS_SETUP: BulkTalentSetup = { name: 'Arms', talentsString: '3030500130', isCurrent: false };

const HEALER: BulkPlayerSettingsSetup = { name: 'Healer', isCurrent: false, settings: SavedSettings.create({ race: Race.RaceDwarf }) };

const current = compareGearSet(1);
const set2 = compareGearSet(2);
const set3 = compareGearSet(3);
const set4 = compareGearSet(4);
const set5 = compareGearSet(5);

const STRENGTH = new Stats().withStat(Stat.StatStrength, 20);

const setup = (gearSet: string, gear: Gear, bonusStats = STRENGTH): BulkGearSetup => ({ gearSet, gear, bonusStats });

const RAID_2 = setup('Raid', set2);
const RAID_3 = setup('Raid', set3);
const RAID_4 = setup('Raid', set4);
const RAID_5 = setup('Raid', set5);
// A preset whose enchants the current gear already wears.
const RAID_CURRENT = setup('Raid', current);
// Another bundle's gear set, which the live player does not wear.
const ARMS_GEAR = setup('Arms', compareGearSet(6));

const rotationNamed = (name: string, rotation: APLRotation): BulkRotationSetup => ({ name, isCurrent: false, rotation });

const simple = (specRotationJson: string) => APLRotation.create({ type: APLRotationType.TypeSimple, simple: { specRotationJson } });

const metrics = (avg: number) => DistributionMetrics.create({ avg, stdev: 10 });

const ALL = 0;

const byGearKey = (...setups: BulkGearSetup[]) => new Map(setups.map(gearSetup => [getGearIdentityKey(gearSetup.gear.asSpec()), gearSetup]));

// One finished request: its baseline at `reference` DPS and the backend's top results, each with the extra fields given.
const request = (
	talents: BulkTalentSetup,
	reference: number,
	rows: Array<[Gear, number, Partial<TopGearResult>?]>,
	{
		iterations = 100,
		rotation = CURRENT_ROTATION_SETUP,
		settings = null,
		setups = null,
	}: { iterations?: number; rotation?: BulkRotationSetup; settings?: BulkPlayerSettingsSetup | null; setups?: BulkFinishedRequest['setups'] } = {},
): BulkFinishedRequest => ({
	reference: false,
	talents,
	rotation,
	settings,
	setups,
	result: {
		referenceDpsMetrics: metrics(reference),
		topGearResults: rows.map(([gear, avg, extra]) => ({ gear, dpsMetrics: metrics(avg), ...extra })),
		iterations,
		metrics: {},
	},
});

// A request that sims the live gear as its baseline, the first of a batch that sims the current setup.
const referenceRequest = (...args: Parameters<typeof request>): BulkFinishedRequest => ({ ...request(...args), reference: true });

const row = (gear: Gear, overrides: Partial<BulkRankedResult> = {}): BulkRankedResult => ({
	gear,
	dpsMetrics: metrics(1100),
	talents: CURRENT_TALENT_SETUP,
	rotation: CURRENT_ROTATION_SETUP,
	gearSetup: null,
	settings: null,
	...overrides,
});

describe('rankEncounterResults', () => {
	it('ranks a lone reference request with the current gear spliced in at its reference DPS, paired errors kept', () => {
		const { chains, originalGearResults, referenceIsCurrent, currentGear } = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [
					[set2, 1100, { backendRank: 0, pairedErrorToBaseline: 3 }],
					[current, 1050],
					[set3, 1020],
				]),
			],
			current,
			ALL,
		);
		expect(chains.flat().map(r => r.dpsMetrics.avg)).toEqual([1100, 1020, 1000]);
		expect(chains.flat()[0].pairedErrorToBaseline).toBe(3);
		expect(originalGearResults.talents).toBe(CURRENT_TALENT_SETUP);
		expect([referenceIsCurrent, currentGear]).toEqual([true, current]);
	});

	it('with no reference request ranks the best row as the reference, every other row below it, all unpaired', () => {
		const requests = [
			request(CURRENT_TALENT_SETUP, 1050, [[set3, 1020, { backendRank: 0, pairedErrorToBaseline: 0.01, pairedErrorToNextResult: 0.01 }]], {
				setups: { byGearKey: byGearKey(RAID_3), baseline: RAID_2 },
			}),
			request(ARMS_SETUP, 1100, [[set4, 1080]], { setups: { byGearKey: byGearKey(RAID_4), baseline: ARMS_GEAR } }),
		];

		const { chains, originalGearResults, referenceIsCurrent, currentGear } = rankEncounterResults(requests, current, ALL);

		expect(chains.flat().map(r => [r.dpsMetrics.avg, r.gearSetup])).toEqual([
			[1100, ARMS_GEAR],
			[1080, RAID_4],
			[1050, RAID_2],
			[1020, RAID_3],
		]);
		expect(originalGearResults).toBe(chains[0][0]);
		expect([referenceIsCurrent, currentGear]).toEqual([false, current]);
		const pairing = chains.flat().map(r => [r.backendRank, r.pairedErrorToBaseline, r.pairedErrorToNextResult]);
		expect(pairing).toEqual(Array(4).fill([undefined, undefined, undefined]));
		// The reference is one of the top results.
		expect(
			rankEncounterResults(requests, current, 2)
				.chains.flat()
				.map(r => r.dpsMetrics.avg),
		).toEqual([1100, 1080]);
	});

	// Seven rows at 1100 down to 1040, best first, and the reference at 1000.
	const SEVEN = [1100, 1090, 1080, 1070, 1060, 1050, 1040];

	it.each([
		[5, [...SEVEN.slice(0, 5), 1000]],
		[3, [...SEVEN.slice(0, 3), 1000]],
		[ALL, [...SEVEN, 1000]],
	])('keeps the best %i Gear-mode rows besides the reference, every row at 0', (topResults, expected) => {
		const rows = rankEncounterResults(
			[
				referenceRequest(
					CURRENT_TALENT_SETUP,
					1000,
					SEVEN.map((avg, index): [Gear, number] => [compareGearSet(index + 2), avg]),
				),
			],
			current,
			topResults,
		).chains.flat();
		expect(rows.map(r => r.dpsMetrics.avg)).toEqual(expected);
	});

	it.each([
		[ALL, [...SEVEN, 1000]],
		[3, [...SEVEN.slice(0, 3), 1000]],
	])('keeps the best %i rows of a Builds request with seven setups besides the reference, every row at 0', (topResults, expected) => {
		const setups = SEVEN.map((_, index) => setup('Raid', compareGearSet(index + 2)));
		const rows = rankEncounterResults(
			[
				referenceRequest(
					CURRENT_TALENT_SETUP,
					1000,
					setups.map((gearSetup, index): [Gear, number] => [gearSetup.gear, SEVEN[index]]),
					{ setups: { byGearKey: byGearKey(...setups), baseline: null } },
				),
			],
			current,
			topResults,
		).chains.flat();
		expect(rows.map(r => r.dpsMetrics.avg)).toEqual(expected);
	});

	it('pools every Builds request and keeps every row besides the reference', () => {
		const candidates = byGearKey(RAID_2, RAID_3, RAID_4, RAID_5);
		const rows = rankEncounterResults(
			[
				referenceRequest(
					CURRENT_TALENT_SETUP,
					1000,
					[
						[set2, 1100],
						[set3, 1080],
						[set4, 1040],
						[set5, 1020],
					],
					{ setups: { byGearKey: candidates, baseline: null } },
				),
				request(
					ARMS_SETUP,
					1090,
					[
						[set2, 1110],
						[set3, 1070],
						[set4, 1030],
						[set5, 1010],
					],
					{ setups: { byGearKey: candidates, baseline: RAID_CURRENT } },
				),
			],
			current,
			ALL,
		).chains.flat();
		expect(rows.map(r => [r.dpsMetrics.avg, r.talents.name])).toEqual([
			[1110, 'Arms'],
			[1100, ''],
			[1090, 'Arms'],
			[1080, ''],
			[1070, 'Arms'],
			[1040, ''],
			[1030, 'Arms'],
			[1020, ''],
			[1010, 'Arms'],
			[1000, ''],
		]);
	});

	it('keeps the current gear under another bundle’s talents as a row of its own', () => {
		const rows = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [], { setups: { byGearKey: new Map(), baseline: null } }),
				request(ARMS_SETUP, 1080, [], { setups: { byGearKey: new Map(), baseline: RAID_CURRENT } }),
			],
			current,
			ALL,
		).chains.flat();
		expect(rows.map(r => [r.dpsMetrics.avg, r.talents.name])).toEqual([
			[1080, 'Arms'],
			[1000, ''],
		]);
		expect(new Set(rows.map(bulkResultKey)).size).toBe(2);
	});

	it('drops the paired errors of rows from another request, so their ties use the unpaired test', () => {
		const rows = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [], { setups: { byGearKey: new Map(), baseline: null } }),
				request(ARMS_SETUP, 900, [[set2, 1001, { backendRank: 0, pairedErrorToBaseline: 0.01, pairedErrorToNextResult: 0.01 }]], {
					setups: { byGearKey: byGearKey(RAID_2), baseline: RAID_CURRENT },
				}),
			],
			current,
			ALL,
		).chains;
		const armsRow = rows.flat().find(r => r.talents === ARMS_SETUP && r.gear === set2)!;
		expect([armsRow.backendRank, armsRow.pairedErrorToBaseline, armsRow.pairedErrorToNextResult]).toEqual([undefined, undefined, undefined]);
		// Kept paired errors (0.01 against 1001) would split set2 from the reference, giving 3 chains. Dropped, 1001 against
		// 1000 at a spread of 10 over 100 iterations ties unpaired, and the Arms baseline at 900 stands apart: 2 chains.
		expect(rows).toHaveLength(2);
	});

	it('takes the fewest iterations any of its requests reached', () => {
		const requests = [referenceRequest(CURRENT_TALENT_SETUP, 1000, [], { iterations: 300 }), request(ARMS_SETUP, 1000, [], { iterations: 200 })];
		expect(rankEncounterResults(requests, current, ALL).iterations).toBe(200);
	});

	it('takes the reference from the first request’s baseline with no gear setup, even when a setup wears the current gear', () => {
		const { chains, originalGearResults } = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [[set2, 1100]], {
					settings: CURRENT_SETTINGS_SETUP,
					setups: { byGearKey: byGearKey(RAID_2), baseline: RAID_CURRENT },
				}),
			],
			current,
			ALL,
		);

		expect(originalGearResults).toEqual(
			expect.objectContaining({
				gear: current,
				dpsMetrics: metrics(1000),
				talents: CURRENT_TALENT_SETUP,
				rotation: CURRENT_ROTATION_SETUP,
				gearSetup: null,
				settings: CURRENT_SETTINGS_SETUP,
			}),
		);
		expect(chains.flat().map(r => [r.dpsMetrics.avg, r.gearSetup])).toEqual([
			[1100, RAID_2],
			[1000, null],
		]);
	});

	it('ranks a reference request with no candidates by its baseline alone', () => {
		const rows = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [], { setups: { byGearKey: new Map(), baseline: null } }),
				request(CURRENT_TALENT_SETUP, 1200, [[set2, 1080]], { setups: { byGearKey: byGearKey(RAID_2), baseline: null } }),
			],
			current,
			ALL,
		).chains.flat();

		expect(rows.map(r => [r.dpsMetrics.avg, r.gear, r.gearSetup])).toEqual([
			[1080, set2, RAID_2],
			[1000, current, null],
		]);
	});

	it('names each candidate row’s gear setup and drops a row no setup stands for', () => {
		const rows = rankEncounterResults(
			[
				referenceRequest(
					CURRENT_TALENT_SETUP,
					1000,
					[
						[set2, 1100, { backendRank: 0 }],
						[set4, 1090, { backendRank: 1 }],
						[set3, 1080, { backendRank: 2 }],
					],
					{ setups: { byGearKey: byGearKey(RAID_2, RAID_3), baseline: null } },
				),
				request(ARMS_SETUP, 900, [[set4, 1070]], { setups: { byGearKey: byGearKey(RAID_2), baseline: null } }),
			],
			current,
			ALL,
		).chains.flat();

		expect(rows.map(r => [r.dpsMetrics.avg, r.gearSetup, r.backendRank])).toEqual([
			[1100, RAID_2, 0],
			[1080, RAID_3, 2],
			[1000, null, undefined],
		]);
	});

	it('adds a later request’s baseline named by its setup, and none when no setup stands for its baseline', () => {
		const rows = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, []),
				request(CURRENT_TALENT_SETUP, 1090, [], { setups: { byGearKey: byGearKey(RAID_CURRENT), baseline: RAID_CURRENT } }),
				request(ARMS_SETUP, 1200, [], { setups: { byGearKey: new Map(), baseline: null } }),
			],
			current,
			ALL,
		).chains.flat();

		expect(rows.map(r => [r.dpsMetrics.avg, r.gear, r.talents, r.gearSetup])).toEqual([
			[1090, current, CURRENT_TALENT_SETUP, RAID_CURRENT],
			[1000, current, CURRENT_TALENT_SETUP, null],
		]);
	});

	it('adds a later request’s baseline as a row with its setup’s own gear, whatever gear is live', () => {
		const sentWithoutRings: BulkGearSetup = { ...setup('Fury', set3), ownGear: set5 };
		const rows = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [], { setups: { byGearKey: new Map(), baseline: null } }),
				request(ARMS_SETUP, 1050, [], { setups: { byGearKey: new Map(), baseline: ARMS_GEAR } }),
				request(ARMS_SETUP, 1020, [], { setups: { byGearKey: new Map(), baseline: sentWithoutRings } }),
			],
			current,
			ALL,
		).chains.flat();

		expect(rows.map(r => [r.dpsMetrics.avg, r.gear, r.gearSetup])).toEqual([
			[1050, ARMS_GEAR.gear, ARMS_GEAR],
			[1020, set5, sentWithoutRings],
			[1000, current, null],
		]);
	});

	it('keeps a later Builds request’s candidate that wears the live gear', () => {
		const rows = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [], { setups: { byGearKey: new Map(), baseline: null } }),
				request(ARMS_SETUP, 1040, [[current, 1080]], { setups: { byGearKey: byGearKey(RAID_CURRENT), baseline: ARMS_GEAR } }),
			],
			current,
			ALL,
		).chains.flat();

		expect(rows.map(r => [r.dpsMetrics.avg, r.gear, r.talents, r.gearSetup])).toEqual([
			[1080, current, ARMS_SETUP, RAID_CURRENT],
			[1040, ARMS_GEAR.gear, ARMS_SETUP, ARMS_GEAR],
			[1000, current, CURRENT_TALENT_SETUP, null],
		]);
	});

	it('gives every row its request’s settings and the reference row the first request’s', () => {
		const { chains, originalGearResults } = rankEncounterResults(
			[
				referenceRequest(CURRENT_TALENT_SETUP, 1000, [[set2, 1100]], {
					settings: CURRENT_SETTINGS_SETUP,
					setups: { byGearKey: byGearKey(RAID_2), baseline: null },
				}),
				request(ARMS_SETUP, 1020, [[set3, 1050]], { settings: HEALER, setups: { byGearKey: byGearKey(RAID_3), baseline: ARMS_GEAR } }),
			],
			current,
			ALL,
		);

		expect(originalGearResults.settings).toBe(CURRENT_SETTINGS_SETUP);
		expect(chains.flat().map(r => [r.dpsMetrics.avg, r.settings])).toEqual([
			[1100, CURRENT_SETTINGS_SETUP],
			[1050, HEALER],
			[1020, HEALER],
			[1000, CURRENT_SETTINGS_SETUP],
		]);
	});

	it('keys a row by gear, bonus stats, talents, rotation and settings', () => {
		const condition = (uuid: string) => APLRotation.create({ priorityList: [{ action: { condition: { uuid: { value: uuid } } } }] });
		const base = row(set2, { gearSetup: RAID_2, rotation: rotationNamed('Opener', condition('a')), settings: CURRENT_SETTINGS_SETUP });
		const keyOf = (overrides: Partial<BulkRankedResult>) => bulkResultKey({ ...base, ...overrides });

		// Names, DPS and uuids are not part of the key, and two current settings are the same settings.
		expect(
			keyOf({
				gearSetup: setup('Saved', set2),
				dpsMetrics: metrics(900),
				rotation: rotationNamed('Renamed', condition('b')),
				settings: { name: '', isCurrent: true, settings: null },
			}),
		).toBe(bulkResultKey(base));
		expect(keyOf({ settings: { ...HEALER, name: 'Renamed' } })).toBe(keyOf({ settings: HEALER }));
		expect(
			new Set([
				bulkResultKey(base),
				keyOf({ gear: set3 }),
				keyOf({ gearSetup: setup('Raid', set2, new Stats()) }),
				keyOf({ gearSetup: null }),
				keyOf({ talents: ARMS_SETUP }),
				keyOf({ rotation: CURRENT_ROTATION_SETUP }),
				keyOf({ rotation: rotationNamed('Opener', simple('{"useOpener":true}')) }),
				keyOf({ rotation: rotationNamed('Opener', simple('{"useOpener":false}')) }),
				keyOf({ settings: HEALER }),
			]).size,
		).toBe(9);
	});
});

describe('equipBulkResult', () => {
	const hostWith = () => {
		const player = { setGear: vi.fn(), setTalentsString: vi.fn(), setAplRotation: vi.fn(), setBonusStats: vi.fn() };
		return { player, host: { player } as unknown as IndividualSimHost<any> };
	};
	const opener = rotationNamed('Opener', simple('{"useOpener":true}'));

	beforeEach(() => {
		vi.mocked(applySavedSettings).mockClear();
	});

	it('sets gear, talents, rotation, bonus stats and the row’s settings for a Builds row and says it equipped a build', () => {
		const build = row(set2, { talents: ARMS_SETUP, rotation: opener, gearSetup: RAID_2, settings: HEALER });
		const { player, host } = hostWith();

		expect(equipBulkResult(host, build)).toBe('build');
		expect(player.setGear).toHaveBeenCalledWith(set2);
		expect(player.setTalentsString).toHaveBeenCalledWith(ARMS_SETUP.talentsString);
		expect(player.setAplRotation).toHaveBeenCalledWith(opener.rotation);
		expect(player.setBonusStats).toHaveBeenCalledWith(STRENGTH);
		expect(applySavedSettings).toHaveBeenCalledOnce();
		expect(applySavedSettings).toHaveBeenCalledWith(host, HEALER.settings);
	});

	it('leaves the settings alone for a row under the current settings', () => {
		const build = row(set2, { talents: ARMS_SETUP, rotation: opener, gearSetup: RAID_2, settings: CURRENT_SETTINGS_SETUP });
		const { player, host } = hostWith();

		// Still a build: its gear setup's talents, rotation and bonus stats apply.
		expect(equipBulkResult(host, build)).toBe('build');
		expect(player.setTalentsString).toHaveBeenCalledWith(ARMS_SETUP.talentsString);
		expect(applySavedSettings).not.toHaveBeenCalled();
	});

	it('sets only the gear for a Gear-mode row and says it equipped gear', () => {
		const { player, host } = hostWith();

		expect(equipBulkResult(host, row(set2, { talents: ARMS_SETUP }))).toBe('gear');
		expect(player.setGear).toHaveBeenCalledWith(set2);
		expect(player.setTalentsString).not.toHaveBeenCalled();
		expect(player.setAplRotation).not.toHaveBeenCalled();
		expect(player.setBonusStats).not.toHaveBeenCalled();
		expect(applySavedSettings).not.toHaveBeenCalled();
	});

	it('sets only the gear for the Builds reference row, the live setup no gear setup stands for, and says it equipped gear', () => {
		const { player, host } = hostWith();

		expect(equipBulkResult(host, row(current, { settings: CURRENT_SETTINGS_SETUP }))).toBe('gear');
		expect(player.setGear).toHaveBeenCalledWith(current);
		expect(player.setTalentsString).not.toHaveBeenCalled();
		expect(player.setAplRotation).not.toHaveBeenCalled();
		expect(player.setBonusStats).not.toHaveBeenCalled();
		expect(applySavedSettings).not.toHaveBeenCalled();
	});
});
