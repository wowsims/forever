import { useSavedTalents } from '@features/talents/hooks/useSavedTalents';
import { usePlayer, useSpecPresets } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useUiStore } from '@sim/hooks/useUiStore';
import { useMemo } from 'react';

import { bulkTalentEntries } from '../model/talents';
import { useBulkState } from './useBulkState';

// The same sources the batch resolves when it starts: the spec's preset talents, the user's saved ones and the live talents.
export const useBulkTalents = () => {
	const playerSpec = usePlayer().getPlayerSpec();
	const presets = useSpecPresets().talents;
	const { entries: saved } = useSavedTalents();
	const checked = useBulkState(slice => slice.talents);
	const powerTools = useUiStore('showPowerTools');
	const current = usePlayerStore('talentsString');

	const entries = useMemo(() => bulkTalentEntries(presets, saved), [presets, saved]);

	return { playerSpec, current, entries, checked, powerTools };
};
