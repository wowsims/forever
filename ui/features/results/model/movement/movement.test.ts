import { RaidSimResult } from '@generated/proto/api';
import { parseAll } from '@sim/proto/combat_log';
import { MAX_MELEE_RANGE, MIN_RANGED_RANGE } from '@sim/proto/sim_tables_auto_gen';
import { describe, expect, it } from 'vitest';

import { buildMovementRun, type MovementRun, type MovementTrack, pathOf, positionAt, rangeZone } from './movement';

// Lines as sim/core/movement.go prints them, for the player `p` and its pets.
const runOf = async (text: string, playerStartYards = 20, duration = 30): Promise<MovementRun | null> => {
	const logs = await parseAll(RaidSimResult.create({ logs: text.trim().replace(/^\s+/gm, '') }));
	const petNames = [...new Set(logs.filter(log => log.source?.isPet).map(log => log.source!.name))];
	const player = {
		name: 'p',
		logs: logs.filter(log => log.source && !log.source.isPet),
		pets: [...petNames, 'Idle'].map(name => ({ name, logs: logs.filter(log => log.source?.isPet && log.source.name === name) })),
	};
	return buildMovementRun({ player, playerStartYards, logs, duration });
};

const trackOf = async (text: string): Promise<MovementTrack> => (await runOf(text))!.tracks[0];

describe('buildMovementRun', () => {
	it('moves a unit along its walk and leaves it where the walk ended, the prepull included', async () => {
		const track = await trackOf(`
			[-19.00] [p (#1)] Walking from 12.00 to 25.00 yards over 2.000s
			[-17.00] [p (#1)] Stopped walking at 25.00 yards
		`);
		expect(positionAt(track, -19)).toBe(12);
		expect(positionAt(track, -18)).toBeCloseTo(18.5);
		expect(positionAt(track, -17)).toBe(25);
		expect(positionAt(track, 10)).toBe(25);
	});

	it('stands a unit where its first walk sets off until that walk starts', async () => {
		const track = await trackOf('[3.00] [p (#1)] Walking from 12.00 to 4.00 yards over 1.000s');
		expect(positionAt(track, -5)).toBe(12);
		expect(positionAt(track, 3)).toBe(12);
	});

	it('takes a replaced walk over from where the unit is when the new walk starts', async () => {
		// A speed change mid-walk: the second line walks on from the current position to the same spot, slower.
		const track = await trackOf(`
			[0.00] [p (#1)] Walking from 4.00 to 12.00 yards over 1.000s
			[0.50] [p (#1)] Walking from 8.00 to 12.00 yards over 0.800s
			[1.30] [p (#1)] Stopped walking at 12.00 yards
		`);
		expect(positionAt(track, 0.25)).toBeCloseTo(6);
		expect(positionAt(track, 1)).toBeCloseTo(10.5);
		expect(positionAt(track, 1.3)).toBe(12);
	});

	it('leaves a unit where an early stop catches it', async () => {
		const track = await trackOf(`
			[2.00] [p (#1)] Walking from 12.00 to 4.00 yards over 1.000s
			[2.25] [p (#1)] Stopped walking at 10.00 yards
		`);
		expect(positionAt(track, 2.1)).toBeCloseTo(11.2);
		expect(positionAt(track, 2.25)).toBe(10);
		expect(positionAt(track, 2.9)).toBe(10);
	});

	it('keeps a unit on the spot through a timed walk in place', async () => {
		const track = await trackOf(`
			[5.00] [p (#1)] Walking from 10.00 to 10.00 yards over 2.000s
			[7.00] [p (#1)] Stopped walking at 10.00 yards
		`);
		expect(positionAt(track, 6)).toBe(10);
	});

	it('starts the run at its first line and ends it with the fight', async () => {
		const run = (await runOf('[-19.00] [p (#1)] Walking from 12.00 to 25.00 yards over 1.857s', 12, 180))!;
		expect(run.start).toBe(-19);
		expect(run.end).toBe(180);
		expect((await runOf('[2.00] [p (#1)] Walking from 12.00 to 4.00 yards over 1.000s'))!.start).toBe(0);
	});

	it('tracks each pet that walked on its own, after the player', async () => {
		const run = (await runOf(`
			[-1.00] [p (#1)] Walking from 12.00 to 25.00 yards over 1.000s
			[0.00] [p (#1) - Cat] Walking from 12.00 to 4.00 yards over 0.555s
			[0.56] [p (#1) - Cat] Stopped walking at 4.00 yards
		`))!;
		expect(run.tracks.map(track => [track.name, track.isPet])).toEqual([
			['p', false],
			['Cat', true],
		]);
		const [player, cat] = run.tracks;
		expect(positionAt(player, 1)).toBe(25);
		expect(positionAt(cat, -1)).toBe(12);
		expect(positionAt(cat, 1)).toBe(4);
	});

	it('stands a player who never walks at its start distance while a pet walks', async () => {
		const run = (await runOf('[0.00] [p (#1) - Cat] Walking from 25.00 to 4.00 yards over 3.000s', 25))!;
		expect(run.tracks.map(track => track.name)).toEqual(['p', 'Cat']);
		expect(run.tracks[0].legs).toEqual([]);
		expect(positionAt(run.tracks[0], 10)).toBe(25);
	});

	it('has nothing to draw when no unit walked', async () => {
		expect(await runOf('[1.00] [p (#1)] Dynamic stat change: {"DodgeRating": 12.000,}')).toBeNull();
	});

	it('reaches the farthest any unit stands, and never short of the ranged breakpoint', async () => {
		expect((await runOf('[0.00] [p (#1)] Walking from 30.00 to 4.00 yards over 3.000s'))!.reach).toBe(30);
		expect((await runOf('[0.00] [p (#1)] Walking from 4.00 to 3.00 yards over 0.100s'))!.reach).toBe(MIN_RANGED_RANGE);
	});
});

describe('pathOf', () => {
	it('runs flat while a unit stands and along each walk while it moves', async () => {
		const track = await trackOf(`
			[2.00] [p (#1)] Walking from 12.00 to 4.00 yards over 1.000s
			[2.25] [p (#1)] Stopped walking at 10.00 yards
		`);
		expect(pathOf(track, -5, 30)).toEqual([
			{ time: -5, yards: 12 },
			{ time: 2, yards: 12 },
			{ time: 2.25, yards: 10 },
			{ time: 30, yards: 10 },
		]);
	});
});

describe('rangeZone', () => {
	it('puts a unit in melee range up to the melee breakpoint, at range from the ranged one, and in the dead zone between', () => {
		expect(rangeZone(MAX_MELEE_RANGE)).toBe('melee');
		expect(rangeZone(MAX_MELEE_RANGE + 0.01)).toBe('dead-zone');
		expect(rangeZone(MIN_RANGED_RANGE - 0.01)).toBe('dead-zone');
		expect(rangeZone(MIN_RANGED_RANGE)).toBe('ranged');
	});
});
