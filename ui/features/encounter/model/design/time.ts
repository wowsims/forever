export type TimeWindow = { start: number; end: number };

export const clamp01 = (fraction: number): number => Math.min(1, Math.max(0, fraction));

export const fractionToSeconds = (fraction: number, baseSeconds: number): number => fraction * baseSeconds;

export const secondsToFraction = (seconds: number, baseSeconds: number): number => (baseSeconds > 0 ? seconds / baseSeconds : 0);

// A step of 0 or less leaves the value as it is.
export const snap = (seconds: number, step: number): number => (step > 0 ? Math.round(seconds / step) * step : seconds);

// Keeps the width and stops at the fight edges. An open end may run past the fight's end, so only the start stops there.
export const moveWindow = (start: number, end: number, deltaFraction: number, openEnd = false): TimeWindow => {
	const width = end - start;
	const nextStart = Math.min(openEnd ? 1 : 1 - width, Math.max(0, start + deltaFraction));
	return { start: nextStart, end: nextStart + width };
};

// The start follows the pointer inside the fight but never crosses the end minus the minimum width.
export const resizeStart = (start: number, end: number, deltaFraction: number, minWidth: number): TimeWindow => ({
	start: Math.max(0, Math.min(start + deltaFraction, end - minWidth, 1)),
	end,
});

// The end follows the pointer but never crosses the start plus the minimum width, nor the fight's end unless it is open.
export const resizeEnd = (start: number, end: number, deltaFraction: number, minWidth: number, openEnd = false): TimeWindow => ({
	start,
	end: Math.min(openEnd ? Infinity : 1, Math.max(end + deltaFraction, start + minWidth)),
});
