import { BulkTopResults } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeBulkTopResults } from '@sim/settings/bulk_settings';
import type { BulkMode, BulkSlice } from '@sim/state/sim_store';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';

import { bulkRunsBuilds } from '../mode';

/** How many results a batch ranks per encounter; 0 is All. With power tools off Gear ranks 5, whatever is stored. */
export const bulkTopResults = (powerTools: boolean, slice: Pick<BulkSlice, 'mode' | 'gearTopResults' | 'buildsTopResults'>): number => {
	if (bulkRunsBuilds(powerTools, slice.mode)) return slice.buildsTopResults;
	return powerTools ? slice.gearTopResults : BULK_SIM_DEFAULT_TOP_RESULTS;
};

// 0 is All. No `settings` bump: the cap changes neither the combination count nor the stored BulkSettings.
export const setBulkTopResults = (player: Player<any>, mode: BulkMode, topResults: number) => {
	patchBulkState(player, mode === 'builds' ? { buildsTopResults: topResults } : { gearTopResults: topResults });
	const { gearTopResults, buildsTopResults } = bulkState(player);
	storeBulkTopResults(player, BulkTopResults.create({ gear: gearTopResults, builds: buildsTopResults }));
};
