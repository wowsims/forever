import { DistributionMetrics } from '@generated/proto/api';
import type { APLRotation } from '@generated/proto/apl';
import type { Encounter, UnitStats } from '@generated/proto/common';
import type { SavedSettings } from '@generated/proto/ui';

import type { EquippedItem } from '../proto/equipped_item';
import { Gear } from '../proto/gear';
import type { Stats } from '../proto/stats';

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

/** A gear set and the talents, rotation and settings it ran with. Only rows from the reference request carry the backend rank and paired errors. */
export interface BulkRankedResult extends TopGearResult {
	talents: BulkTalentSetup;
	rotation: BulkRotationSetup;
	/** Null on Gear-mode rows and on the current gear's reference row. */
	gearSetup: BulkGearSetup | null;
	/** Null on Gear-mode rows. */
	settings: BulkPlayerSettingsSetup | null;
}

export interface BulkPickerEntry {
	/** Index into the batch's item array, or -1/-2 for the two equipped slots this group covers. */
	index: number;
	item: EquippedItem;
}

export interface BulkResults {
	chains: BulkRankedResult[][];
	/** The row every gain is measured against: the current gear's, or the best row when the batch did not sim the current setup. */
	originalGearResults: BulkRankedResult;
	/** False when the reference is the best row. */
	referenceIsCurrent: boolean;
	/** The gear worn when the batch ran, which the rows list their item changes against. */
	currentGear: Gear;
	/** Frozen at run end: the tie brackets were built with it, so the per-row margins have to agree. */
	iterations: number;
}

export interface BulkTalentSetup {
	/** Empty for the current talents, which the views label themselves. */
	name: string;
	talentsString: string;
	isCurrent: boolean;
}

export interface BulkGearSetup {
	/** Empty for the current gear, which the views label themselves. */
	gearSet: string;
	gear: Gear;
	/** Set on a setup a request sends without its ring enchants: the gear it holds, which its result rows carry. */
	ownGear?: Gear;
	/** The gear set's own gear, before the enchant presets' enchants. */
	gearSetGear?: Gear;
	bonusStats: Stats;
}

export interface BulkRotationSetup {
	/** Empty for the current rotation, which the views label themselves. */
	name: string;
	isCurrent: boolean;
	/** As picked, before a Simple or Auto rotation is resolved for a gear set and talents. */
	rotation: APLRotation;
}

export interface BulkEncounterSetup {
	/** Empty for the current encounter, which the views label themselves. */
	name: string;
	isCurrent: boolean;
	encounter: Encounter;
}

export interface BulkPlayerSettingsSetup {
	/** Empty for the current settings, which the views label themselves. */
	name: string;
	isCurrent: boolean;
	/** Null for the current settings, which send none. */
	settings: SavedSettings | null;
}

export interface BulkBuildOverride {
	talentsString: string;
	rotation: APLRotation;
	bonusStats: UnitStats;
}

export interface BulkSimOverrides {
	encounter?: Encounter;
	/** The Settings tab's saved settings the request runs under; absent for the current settings. */
	settings?: SavedSettings;
	build?: BulkBuildOverride;
	/** The gear the request sims as its baseline in place of the live gear; absent in Gear mode and on a Builds batch's reference request. */
	baselineGear?: Gear;
	/** Sent in place of the generated ones, their ring enchants stripped without Enchanting; a candidate that repeats the baseline is dropped. */
	candidates?: readonly Gear[];
	/** How many results the sim ranks and returns; 5 when absent, every candidate at 0. A cap that covers every candidate skips culling. */
	topResults?: number;
}

export interface BulkEncounterResults {
	/** Empty for the current encounter, which the views label themselves. */
	name: string;
	isCurrent: boolean;
	/** Seconds. */
	duration: number;
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
