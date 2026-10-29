import type { Player } from '@sim/player/player';
import { patchBulkState, storeBulkMode } from '@sim/settings/bulk_settings';
import type { BulkMode } from '@sim/state/sim_store';

// The one gate on Builds mode: with power tools off the batch runs in Gear mode whatever is stored.
export const bulkRunsBuilds = (powerTools: boolean, mode: BulkMode): boolean => powerTools && mode === 'builds';

// No `settings` bump: the mode changes neither the combination count nor the stored BulkSettings.
export const setBulkMode = (player: Player<any>, mode: BulkMode) => {
	patchBulkState(player, { mode });
	storeBulkMode(player, mode);
};
