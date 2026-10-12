import { BulkDuration, BulkDurations } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeBulkDurations } from '@sim/settings/bulk_settings';

export interface DurationPair {
	duration: number;
	durationVariation: number;
}

export type DurationEncounter = DurationPair & { useHealth: boolean };

export interface BulkDurationRun extends DurationPair {
	isEncounter: boolean;
}

export const NEW_DURATION_STEP_SECONDS = 60;

const samePair = (a: DurationPair, b: DurationPair) => a.duration === b.duration && a.durationVariation === b.durationVariation;

// The encounter runs first; an extra equal to a pair already listed is skipped. A health fight has no duration to vary, so it runs once.
export const bulkDurationRuns = (encounter: DurationEncounter, extras: readonly DurationPair[]): BulkDurationRun[] => {
	const runs: BulkDurationRun[] = [{ duration: encounter.duration, durationVariation: encounter.durationVariation, isEncounter: true }];
	if (encounter.useHealth) return runs;
	for (const extra of extras) {
		if (runs.some(run => samePair(run, extra))) continue;
		runs.push({ duration: extra.duration, durationVariation: extra.durationVariation, isEncounter: false });
	}
	return runs;
};

export const isDuplicateDuration = (encounter: DurationPair, extras: readonly DurationPair[], index: number): boolean =>
	samePair(encounter, extras[index]) || extras.slice(0, index).some(earlier => samePair(earlier, extras[index]));

export const formatDurationLabel = ({ duration, durationVariation }: DurationPair): string => {
	const total = Math.round(duration);
	const seconds = String(total % 60).padStart(2, '0');
	return `${Math.floor(total / 60)}:${seconds} ±${durationVariation}s`;
};

// No `settings` bump: the durations change neither the combination count nor the stored BulkSettings.
const writeDurations = (player: Player<any>, durations: BulkDuration[]) => {
	patchBulkState(player, { durations });
	storeBulkDurations(player, BulkDurations.create({ durations }));
};

export const addBulkDuration = (player: Player<any>, encounter: DurationPair) => {
	const { durations } = bulkState(player);
	const last = durations.at(-1) ?? encounter;
	writeDurations(player, [
		...durations,
		BulkDuration.create({ duration: last.duration + NEW_DURATION_STEP_SECONDS, durationVariation: last.durationVariation }),
	]);
};

export const updateBulkDuration = (player: Player<any>, index: number, patch: Partial<DurationPair>) => {
	writeDurations(
		player,
		bulkState(player).durations.map((row, at) => (at === index ? BulkDuration.create({ ...row, ...patch }) : row)),
	);
};

export const removeBulkDuration = (player: Player<any>, index: number) => {
	writeDurations(
		player,
		bulkState(player).durations.filter((_, at) => at !== index),
	);
};
