import { StatCapType } from '@generated/proto/api';
import { PseudoStat, Stat } from '@generated/proto/common';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import {
	applyBreakpointLimits,
	breakpointValueToDisplayPercentage,
	clearSoftCappedStats,
	toDefaultUnitStatValue,
	toRelativeSoftCaps,
	toVisualUnitStatPercentage,
} from './utils';

const CRIT_RATING = UnitStat.fromStat(Stat.StatCritRating);
const HIT_RATING = UnitStat.fromStat(Stat.StatHitRating);
const MELEE_HIT_PERCENT = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent);

const softCap = (unitStat: UnitStat, breakpoints: number[], postCapEPs: number[]) => new StatCap(unitStat, breakpoints, StatCapType.TypeSoftCap, postCapEPs);
const threshold = (unitStat: UnitStat, breakpoints: number[], postCapEPs: number[]) =>
	new StatCap(unitStat, breakpoints, StatCapType.TypeThreshold, postCapEPs);

describe('applyBreakpointLimits', () => {
	it('keeps every breakpoint up to and including the limit', () => {
		const [limited] = applyBreakpointLimits([softCap(HIT_RATING, [100, 200, 300], [3, 2, 1])], new Stats().withUnitStat(HIT_RATING, 200));

		expect(limited.breakpoints).toEqual([100, 200]);
		expect(limited.postCapEPs).toEqual([3, 2]);
	});

	it('trims a soft cap’s post-cap EPs with its breakpoints, but leaves a threshold’s alone', () => {
		const limits = new Stats().withUnitStat(HIT_RATING, 100);

		expect(applyBreakpointLimits([softCap(HIT_RATING, [100, 200], [2, 1])], limits)[0].postCapEPs).toEqual([2]);
		// A threshold's post-cap EPs are not per breakpoint, so trimming them would drop the wrong entries.
		expect(applyBreakpointLimits([threshold(HIT_RATING, [100, 200], [9, 8])], limits)[0].postCapEPs).toEqual([9, 8]);
	});

	it('ignores a limit no config declares as a breakpoint, and an unset limit', () => {
		const configs = [softCap(HIT_RATING, [100, 200], [2, 1])];

		expect(applyBreakpointLimits(configs, new Stats().withUnitStat(HIT_RATING, 123))[0].breakpoints).toEqual([100, 200]);
		expect(applyBreakpointLimits(configs, new Stats())[0].breakpoints).toEqual([100, 200]);
	});

	it('clones rather than trimming the configs it was given', () => {
		const original = softCap(HIT_RATING, [100, 200, 300], [3, 2, 1]);

		const [limited] = applyBreakpointLimits([original], new Stats().withUnitStat(HIT_RATING, 100));

		expect(limited).not.toBe(original);
		expect(original.breakpoints).toEqual([100, 200, 300]);
		expect(original.postCapEPs).toEqual([3, 2, 1]);
	});
});

describe('clearSoftCappedStats', () => {
	it('zeroes the hard cap of every soft-capped stat and leaves the rest', () => {
		const statCaps = new Stats().withUnitStat(HIT_RATING, 500).withUnitStat(CRIT_RATING, 300);

		const cleared = clearSoftCappedStats(statCaps, [softCap(HIT_RATING, [100], [1])]);

		expect(cleared.getUnitStat(HIT_RATING)).toBe(0);
		expect(cleared.getUnitStat(CRIT_RATING)).toBe(300);
	});

	it('returns the caps untouched when nothing is soft capped', () => {
		const statCaps = new Stats().withUnitStat(CRIT_RATING, 300);

		expect(clearSoftCappedStats(statCaps, []).getUnitStat(CRIT_RATING)).toBe(300);
	});
});

describe('toRelativeSoftCaps', () => {
	it('re-expresses a soft cap’s breakpoints as gaps to the current stats, in order', () => {
		const baseStats = new Stats().withUnitStat(HIT_RATING, 100);

		const [relative] = toRelativeSoftCaps([softCap(HIT_RATING, [200, 300], [2, 1])], baseStats);

		expect(relative.breakpoints).toEqual([100, 200]);
		expect(relative.postCapEPs).toEqual([2, 1]);
		expect(relative.capType).toBe(StatCapType.TypeSoftCap);
	});

	it('reverses a threshold’s breakpoints and gives every one the first post-cap EP', () => {
		const baseStats = new Stats().withUnitStat(HIT_RATING, 100);

		const [relative] = toRelativeSoftCaps([threshold(HIT_RATING, [200, 300], [7, 8])], baseStats);

		expect(relative.breakpoints).toEqual([200, 100]);
		expect(relative.postCapEPs).toEqual([7, 7]);
	});

	it('leaves the configs it was given alone', () => {
		const original = threshold(HIT_RATING, [200, 300], [7, 8]);

		toRelativeSoftCaps([original], new Stats());

		expect(original.breakpoints).toEqual([200, 300]);
		expect(original.postCapEPs).toEqual([7, 8]);
	});
});

describe('stat cap unit conversions', () => {
	it('converts a rating to its percentage and back', () => {
		const percent = toVisualUnitStatPercentage(1000, CRIT_RATING);

		expect(percent).toBeGreaterThan(0);
		expect(toDefaultUnitStatValue(percent, CRIT_RATING)).toBeCloseTo(1000, 6);
	});

	it('leaves a percentage pseudo-stat on its own scale', () => {
		expect(toVisualUnitStatPercentage(8, MELEE_HIT_PERCENT)).toBe(8);
		expect(toDefaultUnitStatValue(8, MELEE_HIT_PERCENT)).toBe(8);
	});

	it('shows a breakpoint to two decimal places', () => {
		expect(breakpointValueToDisplayPercentage(8, MELEE_HIT_PERCENT)).toBe('8.00');
	});
});
