import { StatCapType } from '@generated/proto/api';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';

/**
 * Trims every breakpoint above its stat's limit. Only the first config declaring the limit as one
 * of its breakpoints is trimmed, matching the vanilla `softCapsConfigWithLimits` getter.
 */
export const applyBreakpointLimits = (softCaps: StatCap[], breakpointLimits: Stats): StatCap[] => {
	const limited = StatCap.cloneSoftCaps(softCaps);
	for (const [unitStat, limit] of breakpointLimits.asUnitStatArray()) {
		if (!limit) continue;
		const config = limited.find(config => config.unitStat.equals(unitStat));
		if (!config || !config.breakpoints.some(breakpoint => breakpoint == limit)) continue;
		config.breakpoints = config.breakpoints.filter(breakpoint => breakpoint <= limit);
		if (config.capType === StatCapType.TypeSoftCap) {
			config.postCapEPs = config.postCapEPs.slice(0, config.breakpoints.length);
		}
	}
	return limited;
};

/** Drops the hard cap of every soft-capped stat, so the breakpoints are the only cap the solve sees. */
export const clearSoftCappedStats = (statCaps: Stats, softCaps: StatCap[]): Stats =>
	softCaps.reduce((caps, { unitStat }) => caps.withUnitStat(unitStat, 0), statCaps);

/**
 * Re-expresses each breakpoint as the gap between the current stats and that breakpoint, which is
 * what the solver's soft caps are measured in. Thresholds are evaluated largest-first and share a
 * single post-cap EP -- the residual value just past the discontinuity -- so the largest reachable
 * threshold is the one targeted.
 */
export const toRelativeSoftCaps = (softCaps: StatCap[], baseStats: Stats): StatCap[] =>
	softCaps.map(config => {
		const relativeBreakpoints = config.breakpoints.map(breakpoint => baseStats.computeGapToCap(config.unitStat, breakpoint));
		let weights = config.postCapEPs.slice();
		if (config.capType == StatCapType.TypeThreshold) {
			relativeBreakpoints.reverse();
			weights = Array(relativeBreakpoints.length).fill(weights[0]);
		}
		return new StatCap(config.unitStat, relativeBreakpoints, config.capType, weights);
	});

/** Rating to the percentage the cap inputs display. */
export const toVisualUnitStatPercentage = (statValue: number, unitStat: UnitStat): number => unitStat.convertDefaultUnitsToPercent(statValue)!;

/** The inverse of `toVisualUnitStatPercentage`, for what the user types back in. */
export const toDefaultUnitStatValue = (value: number, unitStat: UnitStat): number => unitStat.convertPercentToDefaultUnits(value)!;

/** A configured breakpoint as the tooltips and limit selects show it. */
export const breakpointValueToDisplayPercentage = (value: number, unitStat: UnitStat): string => unitStat.convertDefaultUnitsToPercent(value)!.toFixed(2);
