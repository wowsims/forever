import { describe, expect, it } from 'vitest';

import { dimPath, ringAround } from './placement';

const viewport = { width: 1600, height: 900 };

describe('ringAround', () => {
	it('pads the box and keeps the ring inside the viewport', () => {
		expect(ringAround({ left: 100, top: 100, width: 50, height: 20 }, viewport)).toEqual({ left: 94, top: 94, width: 62, height: 32 });
		expect(ringAround({ left: 0, top: 0, width: 50, height: 20 }, viewport)).toEqual({ left: 3, top: 3, width: 53, height: 23 });
		expect(ringAround({ left: 1560, top: 880, width: 40, height: 20 }, viewport)).toEqual({ left: 1554, top: 874, width: 43, height: 23 });
	});
});

describe('dimPath', () => {
	it('is the whole viewport with nothing ringed, and cuts a square hole for a ring', () => {
		expect(dimPath(null, viewport)).toBe('M0 0H1600V900H0Z');
		expect(dimPath({ left: 10, top: 20, width: 100, height: 40 }, viewport)).toBe('M0 0H1600V900H0ZM10 20H110V60H10Z');
	});
});
