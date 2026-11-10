import { useSavedEncounter } from '@features/encounter/hooks/useSavedEncounter';
import { Encounter } from '@generated/proto/common';
import type { BulkEncounterRef } from '@generated/proto/ui';
import { useSimHost, useSpecPresets } from '@sim/context/SimHostContext';
import { useMemo } from 'react';
import { useStore } from 'zustand';

import { bulkEncounterEntries, type BulkEncounterEntry } from '../model/builds/encounters';
import { useBulkState } from './useBulkState';

// The same sources the batch resolves when it starts: the spec's preset encounters, the user's saved ones and the live
// encounter.
export const useBulkEncounters = (): { current: Encounter; entries: BulkEncounterEntry[]; checked: readonly BulkEncounterRef[] } => {
	const host = useSimHost();
	const presets = useSpecPresets();
	const { entries: saved } = useSavedEncounter();
	const checked = useBulkState(slice => slice.encounters);
	const liveEncounter = useStore(host.sim.store, state => state.encounter);

	const current = useMemo(() => Encounter.create(liveEncounter), [liveEncounter]);
	const entries = useMemo(() => bulkEncounterEntries(presets.encounters, saved), [presets, saved]);

	return { current, entries, checked };
};
