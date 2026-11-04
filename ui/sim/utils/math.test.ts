import { describe, expect, it } from 'vitest';

import { roundSecondsRemaining, smoothSecondsRemaining } from './math';

describe('shown time left', () => {
	it('starts at the first estimate', () => {
		expect(smoothSecondsRemaining(undefined, 42, 1)).toBe(42);
	});

	it('counts down a second a second while the estimate does not move', () => {
		const shown = [40.4];
		for (let second = 0; second < 3; second++) shown.push(smoothSecondsRemaining(shown.at(-1), 40.4, 1));
		expect(shown.map(roundSecondsRemaining)).toEqual([40, 39, 38, 37]);
	});

	it('glides toward a lower estimate rather than jumping to it', () => {
		const first = smoothSecondsRemaining(60, 30, 1);
		const second = smoothSecondsRemaining(first, 29, 1);
		expect(first).toBeLessThan(59);
		expect(first).toBeGreaterThan(30);
		expect(second).toBeLessThan(first);
		expect(second).toBeGreaterThan(29);
	});

	it('rises only for an estimate more than 10 s or a quarter above it, and glides up then', () => {
		expect(smoothSecondsRemaining(25, 33, 1)).toBe(24);
		const raised = smoothSecondsRemaining(30, 45, 1);
		expect(raised).toBeGreaterThan(30);
		expect(raised).toBeLessThan(45);
	});

	it('counts down at most 3 s a second when the estimate collapses, so it takes several seconds to reach the floor', () => {
		const shown = [15];
		for (let second = 0; second < 3; second++) shown.push(smoothSecondsRemaining(shown.at(-1), 0, 1));
		expect(shown).toEqual([15, 12, 9, 6]);
	});

	it('shows whole seconds, and never less than 1 s while running', () => {
		expect([42.4, 42.6, 0.2, 0].map(roundSecondsRemaining)).toEqual([42, 43, 1, 1]);
	});
});
