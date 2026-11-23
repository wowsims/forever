import { usePlayer } from '@sim/context/SimHostContext';
import { useReadyStoreSubscribe } from '@sim/hooks/useReadyStoreSubscribe';
import { useSimReady } from '@sim/hooks/useSimReady';
import type { Player } from '@sim/player/player';
import { bulkState } from '@sim/settings/bulk_settings';
import { subscribeAll, subscribeBulkChange, subscribePlayerField, subscribeSimField } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import type { DropsItemList } from '../model/drops/items';
import { dropsSetupFor } from '../model/drops/settings';
import type { DropsNames, DropsSourceNode } from '../model/drops/sources';

/** Empty until the database is loaded. */
export interface DropsSetupState {
	tree: DropsSourceNode[];
	ticked: ReadonlySet<string>;
	list: DropsItemList;
	names: DropsNames;
}

const NOT_READY: DropsSetupState = {
	tree: [],
	ticked: new Set(),
	list: { items: [], leftOut: [] },
	names: { zone: () => null, npc: () => null, faction: () => null },
};

const readSetup = (player: Player<any>): DropsSetupState => ({ ...dropsSetupFor(player), ticked: new Set(bulkState(player).dropsSources) });

// Rebuilt when the database lands, the phase, the race (which sets the side `getFaction` reads), the gear or a Drops pick changes.
// The picks bump `settings` and a gear change bumps `items`, both of which `subscribeBulkChange` watches.
export const useDropsSetup = (): DropsSetupState => {
	const player = usePlayer();
	const ready = useSimReady();
	const subscribe = useMemo(
		() =>
			subscribeAll([
				subscribeSimField(player.sim, 'phase'),
				subscribePlayerField(player, 'race'),
				subscribePlayerField(player, 'gear'),
				subscribeBulkChange(player),
			]),
		[player],
	);
	return useReadyStoreSubscribe(subscribe, () => readSetup(player), ready) ?? NOT_READY;
};
