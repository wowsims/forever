import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { type BulkMode, createSimStore, patchSlice } from '@sim/state/sim_store';
import { memoryEnv } from '@testing/sim/memory_env';

// A host over a real store with power tools on and the batch in `mode` at `topResults` (0 is All), for the Top results
// control. Its picks are stored in memory.
export const makeBulkTopResultsHost = (mode: BulkMode, topResults: number): IndividualSimHost<any> => {
	const player = {
		sim: { store: createSimStore(), env: memoryEnv(), isNative: false },
		storeKey: 0,
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
	} as unknown as Player<any>;
	patchSlice(player.sim.store, 'ui', { showPowerTools: true });
	seedBulkSettings(player);
	patchBulkState(player, mode === 'builds' ? { mode, buildsTopResults: topResults } : { mode, gearTopResults: topResults });
	return { player, sim: player.sim } as unknown as IndividualSimHost<any>;
};
