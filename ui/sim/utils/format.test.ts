import { describe, expect, it } from 'vitest';

import { formatDeltaText } from './format';

describe('formatDeltaText', () => {
	it('gives the percentage of the reference either way', () => {
		expect(formatDeltaText(1000, 800, 0, undefined, false, true)).toEqual({ text: '-200 (20%)', tone: 'negative' });
		expect(formatDeltaText(1000, 1200, 0, undefined, false, true)).toEqual({ text: '+200 (20%)', tone: 'positive' });
		expect(formatDeltaText(156.45, 71.17, 2, undefined, true, true)).toEqual({ text: '-85.28 (54.51%)', tone: null });
	});

	it('colours a drop positive when lower is better, and leaves the percentage out unless asked', () => {
		expect(formatDeltaText(1000, 800, 0, true)).toEqual({ text: '-200', tone: 'positive' });
	});
});
