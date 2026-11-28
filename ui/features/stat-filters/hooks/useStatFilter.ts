import { useSimHost } from '@sim/context/SimHostContext';
import type { UnitStat } from '@sim/proto/stats';
import { useMemo } from 'react';
import { useStore } from 'zustand';

import { isCustomised, shownStatsReader, type StatFilterScope, type StatFilterSlot, withSelection } from '../model/stat_filters';

/**
 * One panel's stat filter, shaped as the stat filter popover's props. The host re-renders only when this slot's shown
 * stats or its customised flag change, and `shown` keeps its identity while it holds the same stats, so a write to
 * another filter rebuilds nothing. The writers read the filters at write time, so a write to another field since the
 * last render survives.
 */
export const useStatFilter = (slot: StatFilterSlot, scope: StatFilterScope) => {
	const host = useSimHost();
	const { sim, player } = host;
	const readShown = useMemo(() => shownStatsReader(slot, scope), [slot, scope]);
	const shown = useStore(sim.store, state => readShown(state.sim.filters));
	const customised = useStore(sim.store, state => isCustomised(state.sim.filters, slot));
	return {
		universe: scope.universe,
		shown,
		customised,
		groupsStorageKey: host.getStorageKey(`__${slot}StatFilterGroups`),
		playerClass: player.getClass(),
		onChange: (stats: UnitStat[]) => sim.setFilters(withSelection(sim.getFilters(), slot, stats)),
		onReset: () => sim.setFilters(withSelection(sim.getFilters(), slot, undefined)),
	};
};
