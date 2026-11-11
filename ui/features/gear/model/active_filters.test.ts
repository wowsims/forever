import { AreaType, CreatureFamily, MobType } from '@generated/proto/common';
import { DatabaseFilters, SourceFilterOption } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import { activeFilterCount } from './active_filters';

describe('activeFilterCount', () => {
	it('counts none on filters that narrow nothing, whatever the source list holds', () => {
		expect(activeFilterCount(DatabaseFilters.create({ sources: [SourceFilterOption.SourceCrafting] }))).toBe(0);
	});

	it('counts each bonus pick, the newly discovered flag and each level bound that is set', () => {
		const filters = DatabaseFilters.create({
			areaTypes: [AreaType.AreaTypeDesert, AreaType.AreaTypeSnowy],
			mobTypes: [MobType.MobTypeUndead],
			creatureFamilies: [CreatureFamily.CreatureFamilyPlant],
			newlyDiscoveredOnly: true,
			minIlvl: 40,
			maxRequiredLevel: 60,
		});
		expect(activeFilterCount(filters)).toBe(7);
	});
});
