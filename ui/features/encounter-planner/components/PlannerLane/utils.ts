import type { AddWave, EncounterPlan, HealthKeyframe, PlanBossAbility, PlanDamageAmp, PlanDowntime, PlanMove } from '@generated/proto/common';
import { PlanBossAbility_Kind, PlanTime } from '@generated/proto/common';
import type { AxisDragMode } from '@ui-kit/hooks/usePointerAxisDrag';

import type { HealthSegment, PlanEventKind, PlanEventRef, PlanLaneKind } from '../../model/lanes';
import {
	eventsOf,
	eventWindow,
	firesAtInstants,
	getEvent,
	healthSegments,
	LANE_EVENT_KINDS,
	moveSeconds,
	occurrences,
	PLAYER_RUN_SPEED,
	pulseTicks,
} from '../../model/lanes';
import type { TimeWindow } from '../../model/time';
import { clamp, secondsToFraction } from '../../model/time';
import { editEvent, plannerText } from '../EncounterPlanner/utils/configs';
import { formatPercent } from '../EncounterPlanner/utils/tips';

export const LANE_CLASSES = 'flex min-w-full border-b border-gray-700';
export const LANE_LABEL_CLASSES =
	'sticky left-0 z-3 box-border flex w-(--label-w) shrink-0 grow-0 basis-(--label-w) flex-col justify-center gap-1 border-r border-gray-700 bg-background px-2 py-1';
export const LANE_GRID_CLASSES = 'relative grid shrink-0 grow basis-planner-total auto-rows-(--planner-lane-height) grid-cols-1';

export const refKey = (ref: PlanEventRef): string => `${ref.kind}:${ref.index}`;

const MAX_OCCURRENCES = 500;

export interface LaneEvent {
	ref: PlanEventRef;
	window: TimeWindow;
	jitter: { start: number; end: number };
	/** Every occurrence after the first, which the lane draws as ghosts. A spell hit, a strike or an area hit has none: its occurrences are `ticks`. */
	repeats: ReadonlyArray<TimeWindow>;
	/** The start of every occurrence of a spell hit, a strike or an area hit, or every tick of a pulse's windows, which the lane draws as tick marks. */
	ticks: ReadonlyArray<number>;
	row: number;
	/** The window may run past the fight's end, as a wave's lifetime and a move's walk do. */
	openEnd: boolean;
	/** The start edge is the implied 0 s point of the boss health curve, which no drag moves. */
	startFixed: boolean;
}

const jitterOf = (plan: EncounterPlan, ref: PlanEventRef): { start: number; end: number } => {
	const event = getEvent(plan, ref);
	switch (ref.kind) {
		case 'bossHealth':
			return { start: 0, end: 0 };
		case 'move': {
			const jitter = (event as PlanMove).at?.jitter ?? 0;
			return { start: jitter, end: jitter };
		}
		case 'wave': {
			const jitter = (event as AddWave).spawn?.jitter ?? 0;
			return { start: jitter, end: jitter };
		}
		default: {
			const timed = event as PlanDowntime | PlanDamageAmp | PlanBossAbility;
			return { start: timed.start?.jitter ?? 0, end: timed.end?.jitter ?? 0 };
		}
	}
};

// Greedy interval packing over windows sorted by start: each one takes the first row whose last window ended before it.
// Points at the same instant land on separate rows.
const packRows = (extents: ReadonlyArray<TimeWindow>): Array<number> => {
	const order = extents.map((_, index) => index).sort((a, b) => extents[a].start - extents[b].start);
	const rowEnds: Array<number> = [];
	const rows = new Array<number>(extents.length);
	for (const index of order) {
		const { start, end } = extents[index];
		let row = rowEnds.findIndex(rowEnd => rowEnd < start);
		if (row < 0) row = rowEnds.length;
		rowEnds[row] = end;
		rows[index] = row;
	}
	return rows;
};

// The window an event covers on the axis. A boss health keyframe covers the segment that ends at it; a move covers
// its walk, which takes the seconds the sim gives it.
const windowOf = (plan: EncounterPlan, ref: PlanEventRef, baseSeconds: number): TimeWindow => {
	if (ref.kind === 'bossHealth') {
		const segment = healthSegments(plan).find(found => found.index === ref.index);
		return { start: segment?.start ?? 0, end: segment?.end ?? 0 };
	}
	const found = eventWindow(plan, ref);
	if (!found) return { start: 0, end: 0 };
	if (!('at' in found)) return found;
	if (ref.kind === 'move') return { start: found.at, end: found.at + secondsToFraction(moveSeconds(getEvent(plan, ref) as PlanMove), baseSeconds) };
	return { start: found.at, end: found.at };
};

// A pulse's ticks over all its windows, cut at the fight's end, at most MAX_OCCURRENCES of them.
const pulseLaneTicks = (windows: ReadonlyArray<TimeWindow>, intervalSeconds: number, baseSeconds: number): Array<number> => {
	const ticks: Array<number> = [];
	for (const { start, end } of windows) {
		ticks.push(...pulseTicks({ start, end: Math.min(end, 1) }, intervalSeconds, baseSeconds, MAX_OCCURRENCES - ticks.length));
		if (ticks.length >= MAX_OCCURRENCES) break;
	}
	return ticks;
};

// Whether a downtime sits on one target's row: its own, and on the first target every downtime that names no target of the
// `count` the encounter has, which validation flags.
const onTargetRow = (plan: EncounterPlan, index: number, target: { index: number; count: number }): boolean => {
	const named = plan.downtimes[index].targetIndex;
	return named === target.index || (target.index === 0 && (named < 0 || named >= target.count));
};

// The event kinds a row can add; the boss health shapes the first target alone.
export const rowKinds = (lane: PlanLaneKind, targetIndex?: number): ReadonlyArray<PlanEventKind> =>
	targetIndex ? LANE_EVENT_KINDS[lane].filter(kind => kind !== 'bossHealth') : LANE_EVENT_KINDS[lane];

// The lane's events in time order, each with the row that keeps it clear of the others, its repeats and jitter included.
// The boss health segments share the lane's first row, which no downtime takes. With a target, the 'boss' lane holds only
// that target's row; `target` means nothing in the other lanes.
export const laneEvents = (plan: EncounterPlan, lane: PlanLaneKind, baseSeconds: number, target?: { index: number; count: number }): Array<LaneEvent> => {
	const segments = healthSegments(plan);
	const events = rowKinds(lane, target?.index).flatMap(kind =>
		eventsOf(plan, kind).flatMap((_, index) => {
			if (target && kind === 'downtime' && !onTargetRow(plan, index, target)) return [];
			const ref: PlanEventRef = { kind, index };
			const window = windowOf(plan, ref, baseSeconds);
			const all = occurrences(plan, ref, MAX_OCCURRENCES);
			const ability = kind === 'ability' ? (getEvent(plan, ref) as PlanBossAbility) : undefined;
			const instants = !!ability && firesAtInstants(ability);
			const pulse = ability?.kind === PlanBossAbility_Kind.Pulse;
			return [
				{
					ref,
					window,
					jitter: jitterOf(plan, ref),
					repeats: instants ? [] : all.slice(1).map(repeat => ({ start: repeat.start, end: repeat.start + window.end - window.start })),
					ticks: instants ? all.map(occurrence => occurrence.start) : pulse ? pulseLaneTicks(all, ability.tickInterval, baseSeconds) : [],
					row: 0,
					openEnd: kind === 'wave' || kind === 'move',
					startFixed: kind === 'bossHealth' && segments.find(segment => segment.index === index)?.previous === undefined,
				},
			];
		}),
	);
	const packed = events.filter(event => event.ref.kind !== 'bossHealth');
	const rows = packRows(
		packed.map(event => ({
			start: event.window.start - event.jitter.start,
			end: Math.max(event.window.end, ...event.repeats.map(repeat => repeat.end)) + event.jitter.end,
		})),
	);
	const offset = events.length > packed.length ? 1 : 0;
	packed.forEach((event, index) => (event.row = rows[index] + offset));
	return events.sort((a, b) => a.window.start - b.window.start || a.row - b.row);
};

const EPSILON = 1e-9;

// Both edges of a boss health segment stay between its neighbours in time.
const boundedStart = (segments: ReadonlyArray<HealthSegment>, position: number, at: number) =>
	clamp(at, segments[position - 1]?.start ?? 0, segments[position].end);
const boundedEnd = (segments: ReadonlyArray<HealthSegment>, position: number, at: number) =>
	clamp(at, segments[position].start, segments[position + 1]?.end ?? 1);

// A drag on a boss health segment moves the keyframe at its end, and its start handle moves the keyframe before it, the
// boundary they share.
const withHealthWindow = (plan: EncounterPlan, ref: PlanEventRef, window: TimeWindow): EncounterPlan => {
	const segments = healthSegments(plan);
	const position = segments.findIndex(segment => segment.index === ref.index);
	const segment = segments[position];
	if (!segment) return editEvent(plan, ref, () => {});
	if (Math.abs(window.end - segment.end) > EPSILON) {
		const at = boundedEnd(segments, position, window.end);
		return editEvent(plan, ref, event => ((event as HealthKeyframe).at = at));
	}
	if (segment.previous === undefined || Math.abs(window.start - segment.start) <= EPSILON) return editEvent(plan, ref, () => {});
	const at = boundedStart(segments, position, window.start);
	return editEvent(plan, { kind: 'bossHealth', index: segment.previous }, event => ((event as HealthKeyframe).at = at));
};

// The window a boss health edit stores: the keyframe at the segment's end between its neighbours, or the one before the segment.
export const adjustHealth =
	(segments: ReadonlyArray<HealthSegment>, ref: PlanEventRef) =>
	(next: TimeWindow, mode: AxisDragMode, from: TimeWindow): TimeWindow => {
		const position = segments.findIndex(found => found.index === ref.index);
		const segment = segments[position];
		if (!segment) return from;
		if (mode === 'start') return segment.previous === undefined ? from : { start: boundedStart(segments, position, next.start), end: from.end };
		return { start: from.start, end: boundedEnd(segments, position, next.end) };
	};

export const healthLabel = (plan: EncounterPlan, segments: ReadonlyArray<HealthSegment>, ref: PlanEventRef): string => {
	const segment = segments.find(found => found.index === ref.index);
	const [from, to] = [segment?.fromPercent ?? 100, segment?.toPercent ?? 100];
	const curve =
		from === to
			? plannerText('health_hold', { percent: formatPercent(from) })
			: plannerText('health_label', { from: formatPercent(from), to: formatPercent(to) });
	const named = getEvent(plan, ref)?.name;
	return named ? `${named}: ${curve}` : curve;
};

// A segment dragged by its end, or its start handle, moves the edge it shares with its neighbour, which follows it.
export const followed = (segments: ReadonlyArray<HealthSegment>, dragging: { index: number; window: TimeWindow } | null, event: LaneEvent): TimeWindow => {
	if (event.ref.kind !== 'bossHealth' || !dragging) return event.window;
	const dragged = segments.find(found => found.index === dragging.index);
	const mine = segments.find(found => found.index === event.ref.index);
	if (!dragged || !mine || dragged === mine) return event.window;
	if (mine.previous === dragged.index) return { start: dragging.window.end, end: event.window.end };
	if (dragged.previous === mine.index) return { start: event.window.start, end: dragging.window.start };
	return event.window;
};

export const isHold = (segments: ReadonlyArray<HealthSegment>, ref: PlanEventRef): boolean => {
	const segment = ref.kind === 'bossHealth' ? segments.find(found => found.index === ref.index) : undefined;
	return !!segment && segment.fromPercent === segment.toPercent;
};

export const withWindow = (plan: EncounterPlan, ref: PlanEventRef, window: TimeWindow, baseSeconds: number): EncounterPlan => {
	if (ref.kind === 'bossHealth') return withHealthWindow(plan, ref, window);
	return editEvent(plan, ref, event => {
		switch (ref.kind) {
			case 'move': {
				const move = event as PlanMove;
				(move.at ??= PlanTime.create()).at = window.start;
				// A walk of yards stretches the yards and a walk of seconds the seconds, whichever the move holds.
				const seconds = (window.end - window.start) * baseSeconds;
				if (Math.abs(seconds - moveSeconds(move)) > EPSILON) {
					// Whole yards: a second of walking is 7 of them, too coarse a step for a drag.
					if (move.yards > 0) move.yards = Math.max(1, Math.round(seconds * PLAYER_RUN_SPEED));
					else move.seconds = seconds;
				}
				return;
			}
			case 'wave': {
				const wave = event as AddWave;
				(wave.spawn ??= PlanTime.create()).at = window.start;
				wave.lifetime = window.end - window.start;
				return;
			}
			default: {
				const timed = event as PlanDowntime | PlanDamageAmp | PlanBossAbility;
				(timed.start ??= PlanTime.create()).at = window.start;
				(timed.end ??= PlanTime.create()).at = window.end;
			}
		}
	});
};
