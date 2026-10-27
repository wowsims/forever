export type TimeWindow = { start: number; end: number };

export const clamp01 = (fraction: number): number => Math.min(1, Math.max(0, fraction));

export const fractionToSeconds = (fraction: number, baseSeconds: number): number => fraction * baseSeconds;

export const secondsToFraction = (seconds: number, baseSeconds: number): number => (baseSeconds > 0 ? seconds / baseSeconds : 0);

// A step of 0 or less leaves the value as it is.
export const snap = (seconds: number, step: number): number => (step > 0 ? Math.round(seconds / step) * step : seconds);

// Keeps the width and stops at the fight edges.
export const moveWindow = (start: number, end: number, deltaFraction: number): TimeWindow => {
	const width = end - start;
	const nextStart = Math.min(1 - width, Math.max(0, start + deltaFraction));
	return { start: nextStart, end: nextStart + width };
};

// The start follows the pointer but never crosses the end minus the minimum width.
export const resizeStart = (start: number, end: number, deltaFraction: number, minWidth: number): TimeWindow => ({
	start: Math.max(0, Math.min(start + deltaFraction, end - minWidth)),
	end,
});

// The end follows the pointer but never crosses the start plus the minimum width.
export const resizeEnd = (start: number, end: number, deltaFraction: number, minWidth: number): TimeWindow => ({
	start,
	end: Math.min(1, Math.max(end + deltaFraction, start + minWidth)),
});
