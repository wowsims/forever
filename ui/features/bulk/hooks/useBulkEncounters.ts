import { useSavedEncounter } from '@features/encounter/hooks/useSavedEncounter';
import { Encounter } from '@generated/proto/common';
import { useSim, useSpecPresets } from '@sim/context/SimHostContext';
import { useUiStore } from '@sim/hooks/useUiStore';
import { useMemo } from 'react';
import { useStore } from 'zustand';

import { bulkEncounterEntries } from '../model/encounters';
import { useBulkState } from './useBulkState';

// The same sources the batch resolves when it starts: the spec's preset encounters, the user's saved ones and the live encounter.
export const useBulkEncounters = () => {
	const sim = useSim();
	const presets = useSpecPresets().encounters;
	const { entries: saved } = useSavedEncounter();
	const checked = useBulkState(slice => slice.encounters);
	const powerTools = useUiStore('showPowerTools');
	// With power tools off the batch runs the current encounter alone, whatever it is, and nothing lists the encounters,
	// so the live one is not read and `current` stays a default encounter.
	const live = useStore(sim.store, state => (powerTools ? state.encounter : undefined));

	const current = useMemo(() => Encounter.create(live), [live]);
	const entries = useMemo(() => bulkEncounterEntries(presets, saved), [presets, saved]);

	return { current, entries, checked, powerTools };
};
