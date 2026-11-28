import type { DatabaseFilters } from '@generated/proto/ui';
import { Sim } from '@sim/sim';

/**
 * How many filters stand off their default and so narrow the item list. The armor and weapon type lists, whose
 * defaults follow the class, are not counted.
 */
export const activeFilterCount = (filters: DatabaseFilters): number =>
	[
		filters.areaTypes.length,
		filters.mobTypes.length,
		filters.creatureFamilies.length,
		filters.itemStats.length,
		filters.newlyDiscoveredOnly,
		filters.minIlvl,
		filters.maxIlvl,
		filters.maxRequiredLevel,
		filters.sources.length < Sim.ALL_SOURCES.length || filters.hideWorldDrops,
		filters.raids.length < Sim.ALL_RAIDS.length,
	].filter(Boolean).length;
