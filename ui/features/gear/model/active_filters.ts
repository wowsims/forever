import type { DatabaseFilters } from '@generated/proto/ui';

/**
 * How many filters narrow the item list beyond its defaults: each picked area, mob type and creature group, the newly
 * discovered flag and each level bound that is set. The source, raid and type lists start full and are not counted.
 */
export const activeFilterCount = (filters: DatabaseFilters): number =>
	filters.areaTypes.length +
	filters.mobTypes.length +
	filters.creatureFamilies.length +
	[filters.newlyDiscoveredOnly, filters.minIlvl, filters.maxIlvl, filters.maxRequiredLevel].filter(Boolean).length;
