import type { MouseEvent as ReactMouseEvent, PointerEvent as ReactPointerEvent } from 'react';
import { useRef } from 'react';

export type AxisDragMode = 'move' | 'start' | 'end';

export type UsePointerAxisDragOptions = {
	pixelsPerSecond: () => number;
	onStart?: (mode: AxisDragMode) => void;
	onDelta: (seconds: number, mode: AxisDragMode) => void;
	onEnd: (commit: boolean) => void;
};

const DRAG_THRESHOLD_PX = 4;

const modeOf = (target: EventTarget): AxisDragMode => {
	const handle = (target as Element).closest('[data-handle]')?.getAttribute('data-handle');
	return handle === 'start' || handle === 'end' ? handle : 'move';
};

// Horizontal drag of one element on a seconds axis. The delta is the total distance since the pointer went down, in
// seconds, so the caller applies it to the value it held when onStart fired.
export const usePointerAxisDrag = ({ pixelsPerSecond, onStart, onDelta, onEnd }: UsePointerAxisDragOptions) => {
	const drag = useRef<{ pointerId: number; x: number; mode: AxisDragMode; started: boolean } | null>(null);
	const dragged = useRef(false);

	const onPointerDown = (event: ReactPointerEvent<HTMLElement>) => {
		if (drag.current || event.button !== 0 || event.pointerType === 'touch') return;
		if ((event.target as Element).closest('button, a, input, select, textarea')) return;
		drag.current = { pointerId: event.pointerId, x: event.clientX, mode: modeOf(event.target), started: false };
		dragged.current = false;
		try {
			event.currentTarget.setPointerCapture(event.pointerId);
		} catch {
			drag.current = null;
		}
	};

	const onPointerMove = (event: ReactPointerEvent<HTMLElement>) => {
		const current = drag.current;
		if (!current || event.pointerId !== current.pointerId) return;
		const dx = event.clientX - current.x;
		if (!current.started) {
			if (Math.abs(dx) < DRAG_THRESHOLD_PX) return;
			current.started = true;
			dragged.current = true;
			onStart?.(current.mode);
		}
		const pps = pixelsPerSecond();
		if (pps > 0) onDelta(dx / pps, current.mode);
	};

	const finish = (event: ReactPointerEvent<HTMLElement>) => {
		const current = drag.current;
		if (!current || event.pointerId !== current.pointerId) return;
		if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
		drag.current = null;
		const cancelled = event.type === 'pointercancel' || event.type === 'lostpointercapture';
		if (cancelled) dragged.current = false;
		if (current.started) onEnd(!cancelled);
	};

	// Suppresses the click a drag would otherwise deliver to whatever it ended on.
	const onClickCapture = (event: ReactMouseEvent<HTMLElement>) => {
		if (!dragged.current) return;
		dragged.current = false;
		event.preventDefault();
		event.stopPropagation();
	};

	return { onPointerDown, onPointerMove, onPointerUp: finish, onPointerCancel: finish, onLostPointerCapture: finish, onClickCapture };
};
