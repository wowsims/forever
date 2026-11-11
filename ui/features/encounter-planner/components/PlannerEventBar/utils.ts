import type { AxisDragMode } from '@ui-kit/hooks/usePointerAxisDrag';
import { cssVars } from '@ui-kit/utils/css';

import type { TimeWindow } from '../../model/time';
import { moveWindow, resizeEnd, resizeStart } from '../../model/time';

export const BAR_CLASSES =
	'top-1 bottom-1 flex w-timeline-segment min-w-2 items-center rounded-xs border border-transparent bg-(--bar) bg-(image:--planner-ticks) bg-position-(--planner-ticks-at) bg-size-(--planner-ticks-size) bg-no-repeat data-selected:border-white';

// An absolutely placed grid item with an auto end line reaches the grid's edge, so the span is spelled out.
export const barVars = (row: number, start: number, duration: number) =>
	cssVars({ '--row': `${row + 1} / span 1`, '--t': String(start), '--dur': String(duration) });

// A pulse's ticks as a faded line every interval across the bar's background, under its text. The first tick, at the
// window's start, is the bar's own edge, so the pattern draws from the second. A boss pulse's first tick lands at its
// cast's end, so its pattern starts one interval before that and leaves the cast clear.
export const tickVars = (seconds: number, castSeconds = 0) => {
	if (seconds <= 0) return undefined;
	const ticks = `repeating-linear-gradient(to right, transparent 0 calc(var(--pps) * ${seconds} - 1px), var(--planner-pulse-tick) 0 calc(var(--pps) * ${seconds}))`;
	if (castSeconds <= 0) return cssVars({ '--planner-ticks': ticks });
	const from = `var(--pps) * ${castSeconds - seconds}`;
	return cssVars({ '--planner-ticks': ticks, '--planner-ticks-at': `calc(${from}) 0`, '--planner-ticks-size': `calc(100% - ${from}) 100%` });
};

type BarKeyAction =
	| { type: 'edit'; mode: AxisDragMode; steps: number }
	| { type: 'remove' }
	| { type: 'activate' }
	| { type: 'duplicate' }
	| { type: 'sibling'; direction: -1 | 1 };

type BarKey = Pick<KeyboardEvent, 'key' | 'shiftKey' | 'altKey' | 'ctrlKey' | 'metaKey'>;

const SHIFT_STEPS = 5;

// Arrows move by one snap step, five with Shift. Alt resizes the end, Alt+Shift the start. Ctrl or Cmd with D duplicates.
export const barKeyAction = (event: BarKey): BarKeyAction | undefined => {
	if (event.ctrlKey || event.metaKey) return event.key.toLowerCase() === 'd' && !event.altKey && !event.shiftKey ? { type: 'duplicate' } : undefined;
	switch (event.key) {
		case 'ArrowLeft':
		case 'ArrowRight': {
			const direction = event.key === 'ArrowLeft' ? -1 : 1;
			if (!event.altKey) return { type: 'edit', mode: 'move', steps: direction * (event.shiftKey ? SHIFT_STEPS : 1) };
			return { type: 'edit', mode: event.shiftKey ? 'start' : 'end', steps: direction };
		}
		case 'ArrowUp':
			return { type: 'sibling', direction: -1 };
		case 'ArrowDown':
			return { type: 'sibling', direction: 1 };
		case 'Delete':
		case 'Backspace':
			return { type: 'remove' };
		case 'Enter':
			return { type: 'activate' };
		default:
			return undefined;
	}
};

export const editWindow = (window: TimeWindow, mode: AxisDragMode, deltaFraction: number, minWidth: number, openEnd = false): TimeWindow => {
	switch (mode) {
		case 'move':
			return moveWindow(window.start, window.end, deltaFraction, openEnd);
		case 'start':
			return resizeStart(window.start, window.end, deltaFraction, minWidth);
		case 'end':
			return resizeEnd(window.start, window.end, deltaFraction, minWidth, openEnd);
	}
};
