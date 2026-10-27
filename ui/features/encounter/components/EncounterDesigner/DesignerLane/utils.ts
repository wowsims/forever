import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, EncounterDesign, HealthKeyframe } from '@generated/proto/common';
import { DesignTime } from '@generated/proto/common';

import type { DesignEventKind, DesignEventRef, DesignLaneKind } from '../../../model/design/lanes';
import { eventsOf, eventWindow, getEvent, laneOf, occurrences } from '../../../model/design/lanes';
import type { TimeWindow } from '../../../model/design/time';
import type { DesignerEventBarShape } from '../DesignerEventBar';
import { editEvent } from '../utils/configs';

const EVENT_KINDS: ReadonlyArray<DesignEventKind> = ['downtime', 'bossHealth', 'ability', 'wave', 'move', 'amp'];

export const laneEventKinds = (lane: DesignLaneKind): ReadonlyArray<DesignEventKind> => EVENT_KINDS.filter(kind => laneOf(kind) === lane);

export const refKey = (ref: DesignEventRef): string => `${ref.kind}:${ref.index}`;

const MAX_OCCURRENCES = 500;

export interface LaneEvent {
	ref: DesignEventRef;
	window: TimeWindow;
	shape: DesignerEventBarShape;
	jitter: { start: number; end: number };
	/** Every occurrence after the first, which the lane draws as ghosts. */
	repeats: ReadonlyArray<TimeWindow>;
	row: number;
}

const SHAPES: Partial<Record<DesignEventKind, DesignerEventBarShape>> = { bossHealth: 'dot', move: 'diamond' };

const jitterOf = (design: EncounterDesign, ref: DesignEventRef): { start: number; end: number } => {
	const event = getEvent(design, ref);
	switch (ref.kind) {
		case 'bossHealth':
			return { start: 0, end: 0 };
		case 'move': {
			const jitter = (event as DesignMove).at?.jitter ?? 0;
			return { start: jitter, end: jitter };
		}
		case 'wave': {
			const jitter = (event as AddWave).spawn?.jitter ?? 0;
			return { start: jitter, end: jitter };
		}
		default: {
			const timed = event as DesignDowntime | DesignBossAbility | DesignDamageAmp;
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

// The lane's events in time order, each with the row that keeps it clear of the others, its repeats and jitter included.
export const laneEvents = (design: EncounterDesign, lane: DesignLaneKind): Array<LaneEvent> => {
	const events = laneEventKinds(lane).flatMap(kind =>
		eventsOf(design, kind).map((_, index) => {
			const ref: DesignEventRef = { kind, index };
			const found = eventWindow(design, ref);
			const window = !found ? { start: 0, end: 0 } : 'at' in found ? { start: found.at, end: found.at } : found;
			const all = occurrences(design, ref, MAX_OCCURRENCES);
			return { ref, window, shape: SHAPES[kind] ?? 'span', jitter: jitterOf(design, ref), repeats: all.slice(1), row: 0 };
		}),
	);
	const rows = packRows(
		events.map(event => ({
			start: event.window.start - event.jitter.start,
			end: Math.max(event.window.end, ...event.repeats.map(repeat => repeat.end)) + event.jitter.end,
		})),
	);
	events.forEach((event, index) => (event.row = rows[index]));
	return events.sort((a, b) => a.window.start - b.window.start || a.row - b.row);
};

export const withWindow = (design: EncounterDesign, ref: DesignEventRef, window: TimeWindow): EncounterDesign =>
	editEvent(design, ref, event => {
		switch (ref.kind) {
			case 'bossHealth':
				(event as HealthKeyframe).at = window.start;
				return;
			case 'move': {
				const move = event as DesignMove;
				(move.at ??= DesignTime.create()).at = window.start;
				return;
			}
			case 'wave': {
				const wave = event as AddWave;
				(wave.spawn ??= DesignTime.create()).at = window.start;
				wave.lifetime = window.end - window.start;
				return;
			}
			default: {
				const timed = event as DesignDowntime | DesignBossAbility | DesignDamageAmp;
				(timed.start ??= DesignTime.create()).at = window.start;
				(timed.end ??= DesignTime.create()).at = window.end;
			}
		}
	});
