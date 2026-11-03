import { type DesignEventRef, eventsOf, occurrences } from '@features/encounter-designer/model/lanes';
import { DesignBossAbility_Kind, type EncounterDesign } from '@generated/proto/common';

import type { ScheduledKind, ScheduledMarker } from '../../../model/timeline/chart/types';
import { SCHEDULED_ROWS } from './annotations';

// A repeat with a tiny interval would otherwise draw thousands of windows.
const MAX_OCCURRENCES = 200;

// The design events a row draws: a pulse's row draws the boss abilities that are pulses, every other row its own kind.
const rowRefs = (design: EncounterDesign, kind: ScheduledKind): Array<DesignEventRef> => {
	if (kind !== 'pulse') return eventsOf(design, kind).map((_, index) => ({ kind, index }));
	return design.bossAbilities.flatMap((ability, index) => (ability.kind === DesignBossAbility_Kind.Pulse ? [{ kind: 'ability' as const, index }] : []));
};

/**
 * The design's scheduled windows in seconds: its times are fractions of the fight, so they scale by the logged
 * iteration's duration. Jitter is ignored, and an occurrence is cut at the end of the fight.
 */
export function scheduledMarkers(design: EncounterDesign | undefined, duration: number): Array<ScheduledMarker> {
	if (!design) return [];
	const markers: Array<ScheduledMarker> = [];
	for (const { kind } of SCHEDULED_ROWS) {
		for (const ref of rowRefs(design, kind)) {
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
