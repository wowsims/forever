import { BulkEncounterRef } from '@generated/proto/ui';
import type { BulkEncounterSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeBatchEncounter } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';

import { type BulkEncounterEntry, pickableBulkEncounters } from './builds/encounters';
import { sameRef } from './builds/saved_refs';

// Undefined for the empty ref, which is the Current encounter, and for a ref no entry holds.
export const batchEncounterEntry = (entries: readonly BulkEncounterEntry[], ref: BulkEncounterRef): BulkEncounterEntry | undefined =>
	entries.find(entry => sameRef(entry.ref, ref));

// The picked encounter, or the Encounter tab's when none is picked or the pick is gone.
export const batchEncounterSetup = (host: IndividualSimHost<any>): BulkEncounterSetup => {
	const picked = batchEncounterEntry(pickableBulkEncounters(host), bulkState(host.player).batchEncounter);
	return picked
		? { name: picked.name, isCurrent: false, encounter: picked.encounter }
		: { name: '', isCurrent: true, encounter: host.sim.encounter.toProto() };
};

// No `settings` bump: the encounter changes neither the combination count nor the stored BulkSettings.
export const setBatchEncounter = (player: Player<any>, ref: BulkEncounterRef) => {
	patchBulkState(player, { batchEncounter: ref });
	storeBatchEncounter(player, ref);
};
