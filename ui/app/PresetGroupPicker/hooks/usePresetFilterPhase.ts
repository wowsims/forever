import { CURRENT_PHASE, type Phase } from '@sim/constants/other';
import { useSimHost } from '@sim/context/SimHostContext';
import { useTypedLocalStorage } from '@ui-kit/hooks/useTypedLocalStorage';
import { useCallback, useMemo } from 'react';

import { PRESET_FILTER_STORAGE_KEY } from '../../storage_keys';

/** Anything the picker filters: only the phase is read, so builds and gear presets both fit. */
type Phased = { phase?: Phase };

interface StoredFilter {
	phase: number;
}

const parseFilter = (value: unknown): StoredFilter | undefined =>
	value && typeof value === 'object' && typeof (value as StoredFilter).phase === 'number' ? { phase: (value as StoredFilter).phase } : undefined;

export interface PresetFilterPhase {
	phases: Array<Phase>;
	phase: Phase;
	/** No-op for an item with no phase, so clicking an unphased chip leaves the bar where it is. */
	selectPhase: (phase?: Phase) => void;
}

export const phasesOf = (items: ReadonlyArray<Phased>): Array<Phase> =>
	[...new Set(items.map(item => item.phase).filter((phase): phase is Phase => phase !== undefined))].sort((a, b) => a - b);

/**
 * The phase tab the grouped pickers are filtered by. Stored under the spec's `__presetFilters__`
 * key and written only when a phase is actually chosen, so a default page load leaves that key absent.
 */
export const usePresetFilterPhase = (items: ReadonlyArray<Phased>): PresetFilterPhase => {
	const host = useSimHost();
	const [stored, setStored] = useTypedLocalStorage<StoredFilter>(host.getStorageKey(PRESET_FILTER_STORAGE_KEY), parseFilter);

	const phases = useMemo(() => phasesOf(items), [items]);

	const phase = useMemo(() => {
		const candidate = stored?.phase as Phase | undefined;
		if (candidate !== undefined && phases.includes(candidate)) return candidate;
		if (phases.includes(CURRENT_PHASE)) return CURRENT_PHASE;
		return phases[0] ?? CURRENT_PHASE;
	}, [stored, phases]);

	const selectPhase = useCallback(
		(next?: Phase) => {
			if (next === undefined || !phases.includes(next)) return;
			setStored({ phase: next });
		},
		[phases, setStored],
	);

	return { phases, phase, selectPhase };
};
