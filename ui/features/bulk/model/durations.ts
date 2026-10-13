import { BulkDuration, BulkDurations } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeBulkDurations } from '@sim/settings/bulk_settings';
import { formatDurationSeconds } from '@sim/utils/format';

export interface DurationPair {
	duration: number;
	durationVariation: number;
}

type DurationEncounter = DurationPair & { useHealth: boolean };

interface BulkDurationRun extends DurationPair {
	isEncounter: boolean;
}

const NEW_DURATION_STEP_SECONDS = 60;

const samePair = (a: DurationPair, b: DurationPair) => a.duration === b.duration && a.durationVariation === b.durationVariation;

export const isDuplicateDuration = (encounter: DurationPair, extras: readonly DurationPair[], index: number): boolean =>
	samePair(encounter, extras[index]) || extras.slice(0, index).some(earlier => samePair(earlier, extras[index]));

// The encounter runs first; an extra equal to a pair already listed is skipped. A health fight has no duration to vary, so it runs once.
export const bulkDurationRuns = (encounter: DurationEncounter, extras: readonly DurationPair[]): BulkDurationRun[] => {
	const runs: BulkDurationRun[] = [{ duration: encounter.duration, durationVariation: encounter.durationVariation, isEncounter: true }];
	if (encounter.useHealth) return runs;
	const kept = extras.filter((_, index) => !isDuplicateDuration(encounter, extras, index));
	for (const extra of kept) runs.push({ duration: extra.duration, durationVariation: extra.durationVariation, isEncounter: false });
	return runs;
};

export const formatDurationLabel = ({ duration, durationVariation }: DurationPair): string =>
	`${formatDurationSeconds(duration, { separatorStyle: 'colon', minimumUnit: 'minutes' })} ±${durationVariation}s`;

// No `settings` bump: the durations change neither the combination count nor the stored BulkSettings.
const writeDurations = (player: Player<any>, durations: BulkDuration[]) => {
	patchBulkState(player, { durations });
	storeBulkDurations(player, BulkDurations.create({ durations }));
};

export const addBulkDuration = (player: Player<any>, encounter: DurationPair) => {
	const { durations } = bulkState(player);
	const last = durations.at(-1) ?? encounter;
	let duration = last.duration + NEW_DURATION_STEP_SECONDS;
	const taken = (candidate: DurationPair) => samePair(candidate, encounter) || durations.some(row => samePair(row, candidate));
	while (taken({ duration, durationVariation: last.durationVariation })) duration += NEW_DURATION_STEP_SECONDS;
	writeDurations(player, [...durations, BulkDuration.create({ duration, durationVariation: last.durationVariation })]);
};

// An edit that would equal the encounter or another row is dropped and reported as `false`.
export const updateBulkDuration = (player: Player<any>, encounter: DurationPair, index: number, patch: Partial<DurationPair>): boolean => {
	const { durations } = bulkState(player);
	const next = BulkDuration.create({ ...durations[index], ...patch });
	if (samePair(next, encounter) || durations.some((row, at) => at !== index && samePair(row, next))) return false;
	writeDurations(
		player,
		durations.map((row, at) => (at === index ? next : row)),
	);
	return true;
};

export const removeBulkDuration = (player: Player<any>, index: number) => {
	writeDurations(
		player,
		bulkState(player).durations.filter((_, at) => at !== index),
	);
};
