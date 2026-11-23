import type { Encounter } from '@generated/proto/common';

import { batchEncounterEntry } from '../model/batch_encounter';
import { useBulkEncounters } from './useBulkEncounters';
import { useBulkState } from './useBulkState';

// The picked encounter, or undefined while the Current encounter is picked or the pick is gone.
export const useBatchEncounter = (): Encounter | undefined => {
	const ref = useBulkState(slice => slice.batchEncounter);
	const { entries } = useBulkEncounters();
	return batchEncounterEntry(entries, ref)?.encounter;
};
