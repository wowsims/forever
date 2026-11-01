import { type DesignEventRef, eventsOf, occurrences } from '@features/encounter-designer/model/lanes';
import type { EncounterDesign } from '@generated/proto/common';

import type { ScheduledMarker } from '../../../model/timeline/chart/types';
import { SCHEDULED_ROWS } from './annotations';

// A repeat with a tiny interval would otherwise draw thousands of windows.
const MAX_OCCURRENCES = 200;

/**
 * The design's scheduled windows in seconds: its times are fractions of the fight, so they scale by the logged
 * iteration's duration. Jitter is ignored, and an occurrence is cut at the end of the fight.
 */
export function scheduledMarkers(design: EncounterDesign | undefined, duration: number): Array<ScheduledMarker> {
	if (!design) return [];
	const markers: Array<ScheduledMarker> = [];
	for (const { kind } of SCHEDULED_ROWS) {
		for (let index = 0; index < eventsOf(design, kind).length; index++) {
			const ref: DesignEventRef = { kind, index };
			for (const window of occurrences(design, ref, MAX_OCCURRENCES)) {
				const start = Math.max(window.start, 0) * duration;
				const end = Math.min(window.end, 1) * duration;
				if (start >= duration || end < start) continue;
				markers.push({ kind, start, end });
			}
		}
	}
	return markers;
}
