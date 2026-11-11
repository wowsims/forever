import { AreaType, CreatureFamily, MobType } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { Sim } from '@sim/sim';
import { describe, expect, it } from 'vitest';

import { activeFilterCount } from './active_filters';

const defaults = (changes: Partial<DatabaseFilters> = {}) =>
	DatabaseFilters.create({ sources: Sim.ALL_SOURCES.slice(), raids: Sim.ALL_RAIDS.slice(), ...changes });

describe('activeFilterCount', () => {
	it('counts none while every filter stands on its default', () => {
		expect(activeFilterCount(defaults())).toBe(0);
	});

	it('counts a filter once, however many picks it holds', () => {
		expect(activeFilterCount(defaults({ areaTypes: [AreaType.AreaTypeDesert, AreaType.AreaTypeSnowy] }))).toBe(1);
		expect(activeFilterCount(defaults({ mobTypes: [MobType.MobTypeUndead], creatureFamilies: [CreatureFamily.CreatureFamilyPlant] }))).toBe(2);
	});

	it('counts the raids once with one or several unticked, and the sources once for an unticked source or the world drops', () => {
		expect(activeFilterCount(defaults({ raids: Sim.ALL_RAIDS.slice(1) }))).toBe(1);
		expect(activeFilterCount(defaults({ raids: Sim.ALL_RAIDS.slice(3) }))).toBe(1);
		expect(activeFilterCount(defaults({ sources: Sim.ALL_SOURCES.slice(1) }))).toBe(1);
		expect(activeFilterCount(defaults({ hideWorldDrops: true }))).toBe(1);
		expect(activeFilterCount(defaults({ sources: Sim.ALL_SOURCES.slice(2), hideWorldDrops: true }))).toBe(1);
	});

	it('counts the newly discovered flag and each level bound that is set', () => {
		expect(activeFilterCount(defaults({ newlyDiscoveredOnly: true, minIlvl: 40, maxIlvl: 80, maxRequiredLevel: 60 }))).toBe(4);
	});
});
