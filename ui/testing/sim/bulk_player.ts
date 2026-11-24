import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';

import { memoryEnv } from './memory_env';

// A warrior over a real store with the batch settings seeded, which is all the batch and enchant preset models read.
// Its storage is a map of its own, so what it stores loads back. `sim` parts are spread over the stub sim; every
// other part sits on the player.
export const makeBulkPlayer = ({ sim, ...parts }: { sim?: object } & Record<string, unknown> = {}): Player<any> => {
	const player = {
		storeKey: 0,
		sim: { store: createSimStore(), env: { storage: memoryEnv().storage }, ...sim },
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		getPlayerClass: () => PlayerSpecs.getPlayerClass(PlayerSpecs.DpsWarrior),
		...parts,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	return player;
};
