import { addEvent, defaultEvent, type DesignEvent, type DesignEventKind, type DesignEventRef, getEvent } from '@features/encounter-designer/model/lanes';
import type { DesignMove, EncounterDesign } from '@generated/proto/common';

const DEFAULT_WIDTH = 0.1;

const place = (event: DesignEvent, at: number) => {
	if ('healthPercent' in event) event.at = at;
	else if ('spawn' in event) event.spawn!.at = at;
	else if ('start' in event) {
		event.start!.at = at;
		event.end!.at = at + DEFAULT_WIDTH;
	} else (event as DesignMove).at!.at = at;
};

export const eventAt = <K extends DesignEventKind>(kind: K, at: number): DesignEvent<K> => {
	const event = defaultEvent(kind);
	place(event, at);
	return event;
};

export const addEventAt = (design: EncounterDesign, kind: DesignEventKind, at: number): DesignEventRef => {
	const ref = addEvent(design, kind);
	place(getEvent(design, ref)!, at);
	return ref;
};
