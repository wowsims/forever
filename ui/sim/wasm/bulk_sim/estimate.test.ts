import { describe, expect, it } from 'vitest';

import { BULK_SIM_DEFAULT_TOP_RESULTS } from './constants_auto_gen';
import { bulkSimHighStageFloor, estimateBulkSimIterations, shouldUseSingleStageBulkSim } from './estimate';

describe('estimateBulkSimIterations', () => {
	it('sims a run under 20 candidates in one stage at the high-stage floor per gear set and the baseline', () => {
		expect(estimateBulkSimIterations(0, BULK_SIM_DEFAULT_TOP_RESULTS)).toBe(bulkSimHighStageFloor());
		expect(estimateBulkSimIterations(19, BULK_SIM_DEFAULT_TOP_RESULTS)).toBe(bulkSimHighStageFloor() * 20);
	});

	it('pins the estimate at 20 candidates', () => {
		expect(estimateBulkSimIterations(20, BULK_SIM_DEFAULT_TOP_RESULTS)).toBe(210_000);
	});

	it('takes the staged estimate from 20 candidates on when it sims fewer', () => {
		expect(estimateBulkSimIterations(5000, BULK_SIM_DEFAULT_TOP_RESULTS)).toBeLessThan(bulkSimHighStageFloor() * 5001);
	});

	it('sims a run that shows every candidate in the high stage alone', () => {
		expect(estimateBulkSimIterations(28, 28)).toBe(bulkSimHighStageFloor() * 29);
		expect(estimateBulkSimIterations(5000, 5000)).toBe(bulkSimHighStageFloor() * 5001);
	});

	it('carries at least the top results through a culling stage', () => {
		// The medium stage at its 1,000 for the 100 and the baseline, then the high stage for the 30 kept and the baseline.
		expect(estimateBulkSimIterations(100, 30)).toBe(101 * 1000 + 31 * bulkSimHighStageFloor());
		expect(estimateBulkSimIterations(100, BULK_SIM_DEFAULT_TOP_RESULTS)).toBe(101 * 1000 + 26 * bulkSimHighStageFloor());
	});
});

describe('shouldUseSingleStageBulkSim', () => {
	it('skips the culling stages when every candidate is shown', () => {
		expect(shouldUseSingleStageBulkSim(28, 28)).toBe(true);
		expect(shouldUseSingleStageBulkSim(1000, 1000)).toBe(true);
	});

	it('culls a larger run that shows the default top results', () => {
		expect(shouldUseSingleStageBulkSim(28, BULK_SIM_DEFAULT_TOP_RESULTS)).toBe(false);
		expect(shouldUseSingleStageBulkSim(1000, BULK_SIM_DEFAULT_TOP_RESULTS)).toBe(false);
	});
});
