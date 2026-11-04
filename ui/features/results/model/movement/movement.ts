import { OtherAction } from '@generated/proto/common';
import type { ActionId } from '@sim/proto/action_id';
import type { CastLog, CombatLog } from '@sim/proto/combat_log';
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

/** One of the player's casts: when it began, when it went off or was cancelled, and the spell it shows. */
export interface MovementCast {
	readonly actionId: ActionId;
	readonly start: number;
	readonly end: number;
	/** Seconds the cast takes to go off, a hardcast's or Auto Shot's windup; 0 for an instant. */
	readonly castTime: number;
	/** It went off at `end`; a cancelled cast stopped there instead. */
	readonly completed: boolean;
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
	/** The player's ranged weapon fires only from MIN_RANGED_RANGE out, so the yards between the breakpoints are a dead zone. */
	readonly deadZone: boolean;
	/** The player's casts but its walks, in the order they began. */
	readonly casts: ReadonlyArray<MovementCast>;
}

export interface MovementUnit {
	readonly name: string;
	readonly logs: ReadonlyArray<CombatLog>;
}

export interface MovementPlayer extends MovementUnit {
	readonly castLogs: ReadonlyArray<CastLog>;
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
	/** The player's ranged weapon fires only from MIN_RANGED_RANGE out. */
	deadZone: boolean;
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
export function buildMovementRun({ player, playerStartYards, logs, duration, deadZone }: MovementRunParams): MovementRun | null {
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

	const casts = player.castLogs.flatMap((log): Array<MovementCast> => {
		if (!log.actionId || log.actionId.otherId === OtherAction.OtherActionMove) return [];
		const completed = !log.castCancelledLog;
		const end = log.timestamp + (completed ? log.castTime : log.cancelTime);
		return [{ actionId: log.actionId, start: log.timestamp, end, castTime: log.castTime, completed }];
	});

	return { start, end: duration, tracks, reach, deadZone, casts };
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

export type RangeZone = 'melee' | 'dead_zone' | 'ranged';

/** In melee range out to the melee breakpoint, and at range past it, but for a weapon with a dead zone only from the ranged breakpoint out. */
export const rangeZone = (yards: number, deadZone: boolean): RangeZone =>
	yards <= MAX_MELEE_RANGE ? 'melee' : deadZone && yards < MIN_RANGED_RANGE ? 'dead_zone' : 'ranged';

/** The casts under way at `time`, hardcasts and Auto Shot's windup alike, in the order they began. */
export const castsUnderWay = (run: MovementRun, time: number): Array<MovementCast> =>
	run.casts.filter(cast => cast.castTime > 0 && cast.start <= time && time < cast.end);

/** The casts that went off in the `span` seconds up to `time`, the latest first. */
export const castsGoneOff = (run: MovementRun, time: number, span: number): Array<MovementCast> =>
	run.casts.filter(cast => cast.completed && cast.end <= time && time - cast.end < span).sort((a, b) => b.end - a.end);
