import type { CombatLog } from '@sim/proto/combat_log';
import { isMovement } from '@sim/proto/combat_log';
import { MAX_MELEE_RANGE, MIN_RANGED_RANGE } from '@sim/proto/sim_tables_auto_gen';

/** One stretch of a walk, in seconds from the pull and in yards from the unit's target. */
export interface Leg {
	readonly start: number;
	readonly end: number;
	readonly from: number;
	readonly to: number;
}

export interface MovementTrack {
	readonly name: string;
	readonly isPet: boolean;
	/** Where the unit stands before its first leg. */
	readonly startYards: number;
	/** In time order; each ends at or before the next one starts. */
	readonly legs: ReadonlyArray<Leg>;
}

export interface MovementRun {
	/** The first line of the run, in seconds from the pull: the prepull's start, or 0 without one. */
	readonly start: number;
	/** The end of the fight. */
	readonly end: number;
	/** The player's track first, then each pet that walked. */
	readonly tracks: ReadonlyArray<MovementTrack>;
	/** The distance a view has to fit: the farthest any unit stands, and at least the ranged breakpoint. */
	readonly reach: number;
}

export interface MovementUnit {
	readonly name: string;
	readonly logs: ReadonlyArray<CombatLog>;
}

export interface MovementPlayer extends MovementUnit {
	readonly pets: ReadonlyArray<MovementUnit>;
}

export interface MovementRunParams {
	player: MovementPlayer;
	/** Where the player stands until its first walk, and all run when it never walks. */
	playerStartYards: number;
	/** Every line of the run, which says where the prepull starts. */
	logs: ReadonlyArray<CombatLog>;
	/** Seconds from the pull to the end of the fight. */
	duration: number;
}

type MutableLeg = { -readonly [K in keyof Leg]: Leg[K] };

function legsOf(logs: ReadonlyArray<CombatLog>): Array<Leg> {
	const legs: Array<MutableLeg> = [];
	for (const log of logs) {
		if (!isMovement(log)) continue;
		const previous = legs[legs.length - 1];
		// A line that lands while the walk before it is under way ends that walk where the line starts.
		if (previous && previous.end > log.timestamp) {
			previous.end = log.timestamp;
			previous.to = log.fromYards;
		}
		legs.push({ start: log.timestamp, end: log.timestamp + log.duration, from: log.fromYards, to: log.toYards });
	}
	return legs;
}

/** The run's movement, or null when no unit walked in it. */
export function buildMovementRun({ player, playerStartYards, logs, duration }: MovementRunParams): MovementRun | null {
	const playerLegs = legsOf(player.logs);
	const pets = player.pets.map(pet => ({ name: pet.name, legs: legsOf(pet.logs) })).filter(pet => pet.legs.length > 0);
	if (playerLegs.length === 0 && pets.length === 0) return null;

	const tracks: Array<MovementTrack> = [
		{ name: player.name, isPet: false, startYards: playerLegs[0]?.from ?? playerStartYards, legs: playerLegs },
		...pets.map(pet => ({ name: pet.name, isPet: true, startYards: pet.legs[0].from, legs: pet.legs })),
	];

	let start = 0;
	for (const log of logs) start = Math.min(start, log.timestamp);

	let reach = MIN_RANGED_RANGE;
	for (const track of tracks) {
		reach = Math.max(reach, track.startYards);
		for (const leg of track.legs) reach = Math.max(reach, leg.from, leg.to);
	}

	return { start, end: duration, tracks, reach };
}

/** Yards from the target at `time`: along a walk under way, else where the last walk ended. */
export function positionAt(track: MovementTrack, time: number): number {
	const { legs } = track;
	// The number of legs that have started by `time`.
	let lo = 0;
	let hi = legs.length;
	while (lo < hi) {
		const mid = (lo + hi) >> 1;
		if (legs[mid].start <= time) lo = mid + 1;
		else hi = mid;
	}
	if (lo === 0) return track.startYards;
	const leg = legs[lo - 1];
	if (time >= leg.end) return leg.to;
	return leg.from + ((leg.to - leg.from) * (time - leg.start)) / (leg.end - leg.start);
}

export interface PathPoint {
	readonly time: number;
	readonly yards: number;
}

/** The corners of a track's distance over `start` to `end`, for a line drawn through them in order. */
export function pathOf(track: MovementTrack, start: number, end: number): Array<PathPoint> {
	const points: Array<PathPoint> = [{ time: start, yards: track.startYards }];
	const add = (time: number, yards: number) => {
		const last = points[points.length - 1];
		if (last.time !== time || last.yards !== yards) points.push({ time, yards });
	};
	for (const leg of track.legs) {
		add(leg.start, points[points.length - 1].yards);
		add(leg.start, leg.from);
		add(leg.end, leg.to);
	}
	add(end, points[points.length - 1].yards);
	return points;
}

export type RangeZone = 'melee' | 'dead-zone' | 'ranged';

/** In melee range out to the melee breakpoint, at range from the ranged one out, and in the dead zone between them. */
export const rangeZone = (yards: number): RangeZone => (yards <= MAX_MELEE_RANGE ? 'melee' : yards < MIN_RANGED_RANGE ? 'dead-zone' : 'ranged');
