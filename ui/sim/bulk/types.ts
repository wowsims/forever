import { DistributionMetrics } from '@generated/proto/api';
import type { APLRotation } from '@generated/proto/apl';

import type { EquippedItem } from '../proto/equipped_item';
import { Gear } from '../proto/gear';

export const WEB_ITERATIONS_LIMIT = 1_000_000;
export const NATIVE_ITERATIONS_LIMIT = 10_000_000;

export const WEB_COMBINATIONS_LIMIT = 5_000;
export const NATIVE_COMBINATIONS_LIMIT = 100_000;

export type OptimisationStage = 'low' | 'medium' | 'high';

export interface TopGearResult {
	gear: Gear;
	dpsMetrics: DistributionMetrics;
	backendRank?: number;
	pairedErrorToNextResult?: number;
	pairedErrorToBaseline?: number;
}

/** A gear set and the talent setup it ran with. Only rows from the current talents' run carry the backend rank and paired errors. */
export interface BulkRankedResult extends TopGearResult {
	talents: BulkTalentSetup;
}

export interface BulkPickerEntry {
	/** Index into the batch's item array, or -1/-2 for the two equipped slots this group covers. */
	index: number;
	item: EquippedItem;
}

export interface BulkResults {
	chains: BulkRankedResult[][];
	originalGearResults: BulkRankedResult;
	/** Frozen at run end: the tie brackets were built with it, so the per-row margins have to agree. */
	iterations: number;
	/** How many talent setups the encounter ran; rows name theirs when it is more than one. */
	talentSetups: number;
}

export interface BulkTalentSetup {
	/** Empty for the current talents, which the views label themselves. */
	name: string;
	talentsString: string;
	isCurrent: boolean;
}

export interface BulkTalentsOverride {
	talentsString: string;
	rotation: APLRotation;
}

export interface BulkEncounterResults {
	/** Empty for the current encounter, which the views label themselves. */
	name: string;
	isCurrent: boolean;
	/** Seconds; null for a health-based encounter, which has no set length. */
	duration: number | null;
	durationVariation: number | null;
	results: BulkResults;
}

export interface BulkSimRoundConfig {
	currentRound: number;
	totalRounds: number;
	title?: string;
	stageCurrentRound?: number;
	stageRounds?: number;
}

export interface BulkSimProgressConfig extends BulkSimRoundConfig {
	aggregateCompletedIterations?: number;
	aggregateTotalIterations?: number;
	aggregateStartedAt?: number;
	useSimCountProgress?: boolean;
}
