import { describe, expect, it } from 'vitest';

import { hasFieldFrame } from './utils';

describe('hasFieldFrame', () => {
	it('is false for a bare control', () => {
		expect(hasFieldFrame({})).toBe(false);
		expect(hasFieldFrame({ label: '', description: '' })).toBe(false);
	});

	it.each([
		['a label', { label: 'Race' }],
		['a description', { description: 'Picks the race.' }],
		['a layout alone', { layout: 'inline' as const }],
	])('is true for %s', (_, frameProps) => {
		expect(hasFieldFrame(frameProps)).toBe(true);
	});

	it('is true for a clear slot alone', () => {
		expect(hasFieldFrame({}, { label: 'Clear' })).toBe(true);
	});
});
