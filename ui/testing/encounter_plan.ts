import { addEvent, defaultEvent, getEvent, type PlanEvent, type PlanEventKind, type PlanEventRef } from '@features/encounter-planner/model/lanes';
import type { EncounterPlan, PlanMove } from '@generated/proto/common';

const DEFAULT_WIDTH = 0.1;

const place = (event: PlanEvent, at: number) => {
	if ('healthPercent' in event) event.at = at;
	else if ('spawn' in event) event.spawn!.at = at;
	else if ('start' in event) {
		event.start!.at = at;
		event.end!.at = at + DEFAULT_WIDTH;
	} else (event as PlanMove).at!.at = at;
};

export const eventAt = <K extends PlanEventKind>(kind: K, at: number): PlanEvent<K> => {
	const event = defaultEvent(kind);
	place(event, at);
	return event;
};

export const addEventAt = (plan: EncounterPlan, kind: PlanEventKind, at: number): PlanEventRef => {
	const ref = addEvent(plan, kind);
	place(getEvent(plan, ref)!, at);
	return ref;
};
