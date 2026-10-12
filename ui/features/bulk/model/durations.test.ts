import { BulkDuration } from '@generated/proto/ui';
import { bulkState, loadStoredBulkDurations } from '@sim/settings/bulk_settings';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it } from 'vitest';

import { addBulkDuration, bulkDurationRuns, formatDurationLabel, isDuplicateDuration, removeBulkDuration, updateBulkDuration } from './durations';

const ENCOUNTER = { duration: 180, durationVariation: 5, useHealth: false };
const pair = (duration: number, durationVariation: number) => ({ duration, durationVariation });

describe('bulkDurationRuns', () => {
	it('runs the encounter first, then the extras in order', () => {
		expect(bulkDurationRuns(ENCOUNTER, [pair(90, 10), pair(300, 30)])).toEqual([
			{ duration: 180, durationVariation: 5, isEncounter: true },
			{ duration: 90, durationVariation: 10, isEncounter: false },
			{ duration: 300, durationVariation: 30, isEncounter: false },
		]);
	});

	it('skips an extra equal to the encounter or an earlier extra', () => {
		const runs = bulkDurationRuns(ENCOUNTER, [pair(180, 5), pair(90, 10), pair(90, 10), pair(90, 20)]);
		expect(runs.map(run => [run.duration, run.durationVariation])).toEqual([
			[180, 5],
			[90, 10],
			[90, 20],
		]);
	});

	it('runs a health fight once', () => {
		expect(bulkDurationRuns({ ...ENCOUNTER, useHealth: true }, [pair(90, 10)])).toEqual([{ duration: 180, durationVariation: 5, isEncounter: true }]);
	});
});

describe('isDuplicateDuration', () => {
	it('flags a row equal to the encounter or to an earlier row only', () => {
		const extras = [pair(180, 5), pair(90, 10), pair(90, 10)];
		expect(extras.map((_, index) => isDuplicateDuration(ENCOUNTER, extras, index))).toEqual([true, false, true]);
	});
});

describe('formatDurationLabel', () => {
	it.each([
		[180, 5, '3:00 ±5s'],
		[90, 10, '1:30 ±10s'],
		[45, 0, '0:45 ±0s'],
		[600, 30, '10:00 ±30s'],
	])('formats %ss ±%ss as %s', (duration, durationVariation, label) => {
		expect(formatDurationLabel(pair(duration, durationVariation))).toBe(label);
	});
});

describe('extra duration rows', () => {
	it('starts a new row a minute after the last row', () => {
		const player = makeBulkPlayer();
		addBulkDuration(player, ENCOUNTER);
		addBulkDuration(player, ENCOUNTER);

		expect(bulkState(player).durations).toEqual([BulkDuration.create(pair(240, 5)), BulkDuration.create(pair(300, 5))]);
	});

	it('stores every change so it loads back', () => {
		const player = makeBulkPlayer();
		addBulkDuration(player, ENCOUNTER);
		addBulkDuration(player, ENCOUNTER);
		updateBulkDuration(player, 0, { durationVariation: 20 });
		removeBulkDuration(player, 1);

		expect(bulkState(player).durations).toEqual([BulkDuration.create(pair(240, 20))]);
		expect(loadStoredBulkDurations(player)?.durations).toEqual([BulkDuration.create(pair(240, 20))]);
	});
});
