import type { Player } from '@sim/player/player';
import { patchBulkState, storeBulkMode } from '@sim/settings/bulk_settings';
import type { BulkMode, BulkSlice } from '@sim/state/sim_store';

// The mode a batch runs in. With power tools off a stored Builds runs as Gear; Drops is no power tool and stays.
export const bulkRunMode = (powerTools: boolean, mode: BulkMode): BulkMode => (mode === 'builds' && !powerTools ? 'gear' : mode);

export const bulkRunsBuilds = (powerTools: boolean, mode: BulkMode): boolean => bulkRunMode(powerTools, mode) === 'builds';

export const bulkRunsDrops = (powerTools: boolean, mode: BulkMode): boolean => bulkRunMode(powerTools, mode) === 'drops';

// Builds is offered only with power tools on.
export const bulkMenuModes = (powerTools: boolean): readonly BulkMode[] => (powerTools ? ['gear', 'drops', 'builds'] : ['gear', 'drops']);

export const bulkModeResults = (runMode: BulkMode, slice: Pick<BulkSlice, 'gearResults' | 'dropsResults' | 'buildsResults'>) => {
	if (runMode === 'drops') return slice.dropsResults;
	return runMode === 'builds' ? slice.buildsResults : slice.gearResults;
};

export const bulkModeStarted = (runMode: BulkMode, slice: Pick<BulkSlice, 'gearStarted' | 'dropsStarted' | 'buildsStarted'>) => {
	if (runMode === 'drops') return slice.dropsStarted;
	return runMode === 'builds' ? slice.buildsStarted : slice.gearStarted;
};

// Bumps `settings`: Drops counts from its own sources, exclusions and enchant preset pool, so a mode switch recounts.
export const setBulkMode = (player: Player<any>, mode: BulkMode) => {
	patchBulkState(player, { mode }, ['settings']);
	storeBulkMode(player, mode);
};
