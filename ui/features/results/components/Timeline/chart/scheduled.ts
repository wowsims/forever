import { eventsOf, occurrences, type PlanEventRef } from '@features/encounter-planner/model/lanes';
import { type EncounterPlan, PlanBossAbility_Kind } from '@generated/proto/common';

import type { ScheduledKind, ScheduledMarker } from '../../../model/timeline/chart/types';
import { SCHEDULED_ROWS } from './annotations';

// A repeat with a tiny interval would otherwise draw thousands of windows.
const MAX_OCCURRENCES = 200;

// The plan events a row draws: a pulse's row draws the boss abilities that are pulses, every other row its own kind.
const rowRefs = (plan: EncounterPlan, kind: ScheduledKind): Array<PlanEventRef> => {
	if (kind !== 'pulse') return eventsOf(plan, kind).map((_, index) => ({ kind, index }));
	return plan.bossAbilities.flatMap((ability, index) => (ability.kind === PlanBossAbility_Kind.Pulse ? [{ kind: 'ability' as const, index }] : []));
};

/**
 * The plan's scheduled windows in seconds: its times are fractions of the fight, so they scale by the logged
 * iteration's duration. Jitter is ignored, and an occurrence is cut at the end of the fight.
 */
export function scheduledMarkers(plan: EncounterPlan | undefined, duration: number): Array<ScheduledMarker> {
	if (!plan) return [];
	const markers: Array<ScheduledMarker> = [];
	for (const { kind } of SCHEDULED_ROWS) {
		for (const ref of rowRefs(plan, kind)) {
			for (const window of occurrences(plan, ref, MAX_OCCURRENCES)) {
				const start = Math.max(window.start, 0) * duration;
				const end = Math.min(window.end, 1) * duration;
				if (start >= duration || end < start) continue;
				markers.push({ kind, start, end });
			}
		}
	}
	return markers;
}
