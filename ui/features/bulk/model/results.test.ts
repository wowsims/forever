import { DistributionMetrics } from '@generated/proto/api';
import type { BulkTalentSetup, TopGearResult } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { Gear } from '@sim/proto/gear';
import { compareGearSet, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { describe, expect, it, vi } from 'vitest';

import { bulkResultKey, type BulkTalentRun, equipBulkResult, rankEncounterResults } from './results';

const ARMS_SETUP: BulkTalentSetup = { name: 'Arms', talentsString: '3030500130', isCurrent: false };

const current = compareGearSet(1);
const set2 = compareGearSet(2);
const set3 = compareGearSet(3);
const set4 = compareGearSet(4);
const set5 = compareGearSet(5);

const metrics = (avg: number) => DistributionMetrics.create({ avg, stdev: 10 });

// One talent setup's run: the current gear at `reference` DPS and the backend's top results, each with the extra fields given.
const run = (talents: BulkTalentSetup, reference: number, rows: Array<[Gear, number, Partial<TopGearResult>?]>, iterations = 100): BulkTalentRun => ({
	talents,
	result: {
		referenceDpsMetrics: metrics(reference),
		topGearResults: rows.map(([gear, avg, extra]) => ({ gear, dpsMetrics: metrics(avg), ...extra })),
		iterations,
		metrics: {},
	},
});

describe('rankEncounterResults', () => {
	it('ranks a lone current-talents run as before: the current gear spliced in at its reference DPS, paired errors kept', () => {
		const { chains, originalGearResults } = rankEncounterResults(
			[
				run(CURRENT_TALENT_SETUP, 1000, [
					[set2, 1100, { backendRank: 0, pairedErrorToBaseline: 3 }],
					[current, 1050],
					[set3, 1020],
				]),
			],
			current,
		);
		expect(chains.flat().map(r => r.dpsMetrics.avg)).toEqual([1100, 1020, 1000]);
		expect(chains.flat()[0].pairedErrorToBaseline).toBe(3);
		expect(originalGearResults.talents).toBe(CURRENT_TALENT_SETUP);
	});

	it('pools every talent setup and keeps the five best besides the reference', () => {
		const rows = rankEncounterResults(
			[
				run(CURRENT_TALENT_SETUP, 1000, [
					[set2, 1100],
					[set3, 1080],
					[set4, 1040],
					[set5, 1020],
				]),
				run(ARMS_SETUP, 1090, [
					[set2, 1110],
					[set3, 1070],
					[set4, 1030],
					[set5, 1010],
				]),
			],
			current,
		).chains.flat();
		expect(rows.map(r => [r.dpsMetrics.avg, r.talents.name])).toEqual([
			[1110, 'Arms'],
			[1100, ''],
			[1090, 'Arms'],
			[1080, ''],
			[1070, 'Arms'],
			[1000, ''],
		]);
	});

	it('keeps the current gear with another talent setup as a row of its own', () => {
		const rows = rankEncounterResults([run(CURRENT_TALENT_SETUP, 1000, []), run(ARMS_SETUP, 1080, [])], current).chains.flat();
		expect(rows.map(r => [r.dpsMetrics.avg, r.talents.name])).toEqual([
			[1080, 'Arms'],
			[1000, ''],
		]);
		expect(new Set(rows.map(bulkResultKey)).size).toBe(2);
	});

	it('drops the paired errors of rows from another talent setup, so their ties use the unpaired test', () => {
		const rows = rankEncounterResults(
			[
				run(CURRENT_TALENT_SETUP, 1000, []),
				run(ARMS_SETUP, 900, [[set2, 1001, { backendRank: 0, pairedErrorToBaseline: 0.01, pairedErrorToNextResult: 0.01 }]]),
			],
			current,
		).chains;
		const row = rows.flat().find(r => r.talents === ARMS_SETUP && r.gear === set2)!;
		expect([row.backendRank, row.pairedErrorToBaseline, row.pairedErrorToNextResult]).toEqual([undefined, undefined, undefined]);
		// Kept paired errors (0.01 against 1001) would split set2 from the reference, giving 3 chains. Dropped, 1001 against
		// 1000 at a spread of 10 over 100 iterations ties unpaired, and the Arms reference at 900 stands apart: 2 chains.
		expect(rows).toHaveLength(2);
	});

	it('takes the fewest iterations any of its talent runs reached', () => {
		expect(rankEncounterResults([run(CURRENT_TALENT_SETUP, 1000, [], 300), run(ARMS_SETUP, 1000, [], 200)], current).iterations).toBe(200);
	});

	it('counts the talent setups it ran', () => {
		expect(rankEncounterResults([run(CURRENT_TALENT_SETUP, 1000, []), run(ARMS_SETUP, 1000, [])], current).talentSetups).toBe(2);
	});
});

describe('equipBulkResult', () => {
	it("sets the row's gear and talents, and says whether the talents changed", () => {
		const player = { setGear: vi.fn(), setTalentsString: vi.fn(), getTalentsString: () => 'live' } as unknown as Player<any>;
		expect(equipBulkResult(player, { gear: set2, dpsMetrics: metrics(1100), talents: ARMS_SETUP })).toBe(true);
		expect(player.setGear).toHaveBeenCalledWith(set2);
		expect(player.setTalentsString).toHaveBeenCalledWith(ARMS_SETUP.talentsString);

		const armsPlayer = { setGear: vi.fn(), setTalentsString: vi.fn(), getTalentsString: () => ARMS_SETUP.talentsString } as unknown as Player<any>;
		expect(equipBulkResult(armsPlayer, { gear: set3, dpsMetrics: metrics(1100), talents: ARMS_SETUP })).toBe(false);
	});
});
