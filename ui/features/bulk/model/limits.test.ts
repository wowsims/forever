import { NATIVE_COMBINATIONS_LIMIT, NATIVE_ITERATIONS_LIMIT, WEB_COMBINATIONS_LIMIT, WEB_ITERATIONS_LIMIT } from '@sim/bulk/types';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// `isExternal` reads `window.location.hostname` once, at module load, so it cannot be steered by
// setting the location afterwards.
const isExternal = vi.hoisted(() => vi.fn(() => false));
vi.mock('@ui-kit/utils/dom', () => ({ isExternal }));

const { bulkCombinationsLimit, bulkIterationsLimit } = await import('./limits');

describe('the host a limit is for', () => {
	beforeEach(() => isExternal.mockReturnValue(false));

	it('trusts the resolved probe over the hostname, in both directions', () => {
		isExternal.mockReturnValue(true);
		expect(bulkIterationsLimit(true)).toBe(NATIVE_ITERATIONS_LIMIT);
		expect(bulkCombinationsLimit(true)).toBe(NATIVE_COMBINATIONS_LIMIT);
		isExternal.mockReturnValue(false);
		expect(bulkIterationsLimit(false)).toBe(WEB_ITERATIONS_LIMIT);
		expect(bulkCombinationsLimit(false)).toBe(WEB_COMBINATIONS_LIMIT);
	});

	it('falls back to the hostname only while the probe is unresolved', () => {
		isExternal.mockReturnValue(true);
		expect(bulkIterationsLimit(undefined)).toBe(WEB_ITERATIONS_LIMIT);
		expect(bulkCombinationsLimit(undefined)).toBe(WEB_COMBINATIONS_LIMIT);
		isExternal.mockReturnValue(false);
		expect(bulkIterationsLimit(undefined)).toBe(NATIVE_ITERATIONS_LIMIT);
		expect(bulkCombinationsLimit(undefined)).toBe(NATIVE_COMBINATIONS_LIMIT);
	});
});
