import { useSavedEncounter } from '@features/encounter/hooks/useSavedEncounter';
import { Encounter } from '@generated/proto/common';
import { useSim, useSpecPresets } from '@sim/context/SimHostContext';
import { useUiStore } from '@sim/hooks/useUiStore';
import { useMemo } from 'react';
import { useStore } from 'zustand';

import { bulkEncounterEntries, checkedBulkEncounters } from '../model/encounters';
import { useBulkState } from './useBulkState';

// The same sources the batch resolves when it starts: the spec's preset encounters, the user's saved ones and the live encounter.
export const useBulkEncounters = () => {
	const sim = useSim();
	const presets = useSpecPresets().encounters;
	const { entries: saved } = useSavedEncounter();
	const checked = useBulkState(slice => slice.encounters);
	const live = useStore(sim.store, state => state.encounter);
	const powerTools = useUiStore('showPowerTools');

	const current = useMemo(() => Encounter.create(live), [live]);
	const entries = useMemo(() => bulkEncounterEntries(presets, saved), [presets, saved]);
	// As the run does, power tools off picks nothing beyond the current encounter.
	const picked = useMemo(() => (powerTools ? checkedBulkEncounters(entries, checked) : []), [powerTools, entries, checked]);

	return { current, entries, checked, picked };
};
