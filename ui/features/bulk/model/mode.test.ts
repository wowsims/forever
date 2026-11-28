import type { BulkEncounterResults } from '@sim/bulk/types';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';
import { describe, expect, it } from 'vitest';

import { bulkMenuModes, bulkModeResults, bulkModeStarted, bulkRunMode, bulkRunsBuilds, bulkRunsDrops } from './mode';
import { bulkTopResults } from './run/top_results';

const GEAR = { name: 'gear' } as unknown as BulkEncounterResults;
const DROPS = { name: 'drops' } as unknown as BulkEncounterResults;

describe('bulk mode gates', () => {
	it('runs a stored Builds as Gear without power tools and keeps Drops', () => {
		expect(bulkRunMode(false, 'builds')).toBe('gear');
		expect(bulkRunMode(false, 'drops')).toBe('drops');
		expect(bulkRunMode(true, 'builds')).toBe('builds');
		expect(bulkRunsDrops(false, 'drops')).toBe(true);
		expect(bulkRunsBuilds(false, 'builds')).toBe(false);
	});

	it('lists Builds in the menu only with power tools on', () => {
		expect(bulkMenuModes(false)).toEqual(['gear', 'drops']);
		expect(bulkMenuModes(true)).toEqual(['gear', 'drops', 'builds']);
	});

	it('reads each mode’s own results and started flag', () => {
		const slice = {
			gearResults: [GEAR],
			dropsResults: [DROPS],
			buildsResults: null,
			gearStarted: true,
			dropsStarted: false,
			buildsStarted: true,
		};
		expect(bulkModeResults('drops', slice)).toEqual([DROPS]);
		expect(bulkModeStarted('drops', slice)).toBe(false);
		expect(bulkModeResults('builds', slice)).toBeNull();
		expect(bulkModeResults('gear', slice)).toEqual([GEAR]);
	});

	it('ranks the default top results in Drops whatever is stored', () => {
		expect(bulkTopResults(true, { mode: 'drops', gearTopResults: 0, buildsTopResults: 0 })).toBe(BULK_SIM_DEFAULT_TOP_RESULTS);
	});
});
