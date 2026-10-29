import type { AxisDragMode } from '@ui-kit/hooks/usePointerAxisDrag';

import type { TimeWindow } from '../../model/time';
import { moveWindow, resizeEnd, resizeStart } from '../../model/time';

export type BarKeyAction =
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
