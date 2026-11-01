import { describe, expect, it } from 'vitest';

import { clamp, fractionToSeconds, moveWindow, resizeEnd, resizeStart, secondsToFraction, snap } from './time';

describe('time', () => {
	it('converts between fractions and seconds', () => {
		expect(fractionToSeconds(0.25, 240)).toBe(60);
		expect(secondsToFraction(60, 240)).toBe(0.25);
		expect(secondsToFraction(60, 0)).toBe(0);
	});

	it('snaps to the step and ignores a step of 0', () => {
		expect(snap(7.4, 0.5)).toBe(7.5);
		expect(snap(7.4, 0)).toBe(7.4);
	});

	it('clamps to 0..1', () => {
		expect([clamp(-0.2, 0, 1), clamp(0.4, 0, 1), clamp(1.3, 0, 1)]).toEqual([0, 0.4, 1]);
	});

	it('moves a window keeping its width inside the fight', () => {
		expect(moveWindow(0.2, 0.4, 0.1)).toEqual({ start: 0.30000000000000004, end: 0.5 });
		const left = moveWindow(0.2, 0.4, -0.5);
		expect(left.start).toBe(0);
		expect(left.end).toBeCloseTo(0.2);
		const right = moveWindow(0.2, 0.4, 0.9);
		expect(right.end).toBe(1);
		expect(right.start).toBeCloseTo(0.8);
	});

	it('resizes the start without crossing the end minus the minimum width', () => {
		expect(resizeStart(0.2, 0.4, 0.1, 0.05)).toEqual({ start: 0.30000000000000004, end: 0.4 });
		expect(resizeStart(0.2, 0.4, 0.5, 0.05).start).toBeCloseTo(0.35);
		expect(resizeStart(0.2, 0.4, -0.5, 0.05)).toEqual({ start: 0, end: 0.4 });
	});

	it('resizes the end without crossing the start plus the minimum width', () => {
		expect(resizeEnd(0.2, 0.4, -0.5, 0.05)).toEqual({ start: 0.2, end: 0.25 });
		expect(resizeEnd(0.2, 0.4, 0.9, 0.05)).toEqual({ start: 0.2, end: 1 });
	});

	it('lets an open end run past the fight and stops only the start at its end', () => {
		const past = moveWindow(0.9, 1.1, 0.05, true);
		expect(past.start).toBeCloseTo(0.95);
		expect(past.end).toBeCloseTo(1.15);
		expect(moveWindow(0.9, 1.1, 0.5, true).start).toBe(1);
		expect(resizeEnd(0.9, 1.1, 0.2, 0.05, true).end).toBeCloseTo(1.3);
		expect(resizeStart(0.9, 1.3, 0.3, 0.05).start).toBe(1);
	});
});
