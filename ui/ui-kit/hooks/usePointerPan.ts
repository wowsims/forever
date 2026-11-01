import type { MouseEvent as ReactMouseEvent, PointerEvent as ReactPointerEvent } from 'react';
import { useRef, useState } from 'react';

type UsePointerPanOptions = {
	/** The element that scrolls vertically, read when the pointer goes down; `null` is the page. */
	verticalScroller: () => HTMLElement | null;
};

const PAN_THRESHOLD_PX = 4;
const INTERACTIVE = 'button, a, input, select, textarea, [role="slider"]';

const scrollTopOf = (element: HTMLElement | null) => (element ? element.scrollTop : window.scrollY);

// Grab-to-pan: the element the handlers sit on scrolls horizontally, and `verticalScroller` vertically. Touch already
// pans both axes natively, and anything interactive keeps the pointer for itself.
export const usePointerPan = ({ verticalScroller }: UsePointerPanOptions) => {
	const pan = useRef<{ pointerId: number; x: number; y: number; scrollLeft: number; scrollTop: number; vertical: HTMLElement | null } | null>(null);
	const panned = useRef(false);
	const [panning, setPanning] = useState(false);

	const onPointerDown = (event: ReactPointerEvent<HTMLElement>) => {
		if (pan.current || event.button !== 0 || event.pointerType === 'touch') return;
		if ((event.target as Element).closest(INTERACTIVE)) return;
		const vertical = verticalScroller();
		pan.current = {
			pointerId: event.pointerId,
			x: event.clientX,
			y: event.clientY,
			scrollLeft: event.currentTarget.scrollLeft,
			scrollTop: scrollTopOf(vertical),
			vertical,
		};
		panned.current = false;
		try {
			event.currentTarget.setPointerCapture(event.pointerId);
		} catch {
			pan.current = null;
		}
	};

	const onPointerMove = (event: ReactPointerEvent<HTMLElement>) => {
		const start = pan.current;
		if (event.pointerId !== start?.pointerId) return;
		const dx = event.clientX - start.x;
		const dy = event.clientY - start.y;
		if (!panned.current) {
			if (Math.abs(dx) < PAN_THRESHOLD_PX && Math.abs(dy) < PAN_THRESHOLD_PX) return;
			panned.current = true;
			setPanning(true);
		}
		event.currentTarget.scrollLeft = start.scrollLeft - dx;
		if (start.vertical) start.vertical.scrollTop = start.scrollTop - dy;
		else window.scrollBy(0, start.scrollTop - dy - window.scrollY);
	};

	const finish = (event: ReactPointerEvent<HTMLElement>) => {
		if (event.pointerId !== pan.current?.pointerId) return;
		if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
		pan.current = null;
		if (event.type === 'pointercancel' || event.type === 'lostpointercapture') panned.current = false;
		setPanning(false);
	};

	// Suppresses the click a pan would otherwise deliver to whatever it ended on.
	const onClickCapture = (event: ReactMouseEvent<HTMLElement>) => {
		if (!panned.current) return;
		panned.current = false;
		event.preventDefault();
		event.stopPropagation();
	};

	return {
		panning,
		handlers: { onPointerDown, onPointerMove, onPointerUp: finish, onPointerCancel: finish, onLostPointerCapture: finish, onClickCapture },
	};
};
