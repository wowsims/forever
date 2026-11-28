import type { RefObject } from 'react';
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';

const VIEWPORT_PADDING = 8;
const CURSOR_OFFSET = 12;

/**
 * Places a hover tooltip near a point in client coordinates, flipping and clamping so it stays whole
 * inside the viewport. Written as custom properties rather than `left`/`top` so the stylesheet keeps
 * the units, and anchored to the viewport, which keeps a tooltip out of the scroller it is drawn over.
 */
const placeTooltip = (host: HTMLElement | null, anchorX: number, anchorY: number) => {
	if (!host) return;
	const rect = host.getBoundingClientRect();

	let left = anchorX + CURSOR_OFFSET;
	if (left + rect.width > window.innerWidth - VIEWPORT_PADDING) left = anchorX - CURSOR_OFFSET - rect.width;
	left = Math.max(VIEWPORT_PADDING, Math.min(left, window.innerWidth - VIEWPORT_PADDING - rect.width));

	let top = anchorY + CURSOR_OFFSET;
	if (top + rect.height > window.innerHeight - VIEWPORT_PADDING) top = window.innerHeight - VIEWPORT_PADDING - rect.height;
	top = Math.max(VIEWPORT_PADDING, top);

	host.style.setProperty('--tooltip-x', String(Math.round(left)));
	host.style.setProperty('--tooltip-y', String(Math.round(top)));
};

export interface HoverTooltip<T> {
	/** Put it on the tooltip's own element; the hook positions that element and nothing else. */
	ref: RefObject<HTMLDivElement | null>;
	content: T | null;
	/** Opens the tooltip on `content`, anchored at a point in client coordinates. */
	show: (content: T, x: number, y: number) => void;
	/** Moves an open tooltip without re-rendering it. `y` defaults to where it already is, which is a horizontal cursor follow. */
	moveTo: (x: number, y?: number) => void;
	hide: () => void;
}

/**
 * A tooltip anchored to a point rather than to an element: what the timeline's chart and its rotation
 * both need, because neither anchor is a stable element. `owner` is the element whose items open it.
 */
export const useHoverTooltip = <T>(owner: RefObject<HTMLElement | null>): HoverTooltip<T> => {
	const ref = useRef<HTMLDivElement>(null);
	const anchor = useRef({ x: 0, y: 0 });
	const [content, setContent] = useState<T | null>(null);

	// The content is measured only once React has rendered it, so the anchor is applied again here:
	// the placement during the event that opened it ran against the previous size.
	useLayoutEffect(() => {
		placeTooltip(ref.current, anchor.current.x, anchor.current.y);
	}, [content]);

	const moveTo = useCallback((x: number, y = anchor.current.y) => {
		anchor.current = { x, y };
		placeTooltip(ref.current, x, y);
	}, []);

	const show = useCallback((next: T, x: number, y: number) => {
		anchor.current = { x, y };
		setContent(next);
	}, []);

	const hide = useCallback(() => setContent(null), []);

	// The point it was opened at moves with a scroll, and a tap opens it with no `mouseleave` to follow, so a scroll or a
	// press outside the owner closes it. A press inside the owner is left to the item it lands on.
	const open = content !== null;
	useEffect(() => {
		if (!open) return;
		const onPointerDown = (event: PointerEvent) => {
			if (!(event.target instanceof Node && owner.current?.contains(event.target))) hide();
		};
		document.addEventListener('scroll', hide, { capture: true, passive: true });
		document.addEventListener('pointerdown', onPointerDown, { capture: true, passive: true });
		return () => {
			document.removeEventListener('scroll', hide, { capture: true });
			document.removeEventListener('pointerdown', onPointerDown, { capture: true });
		};
	}, [open, hide, owner]);

	return { ref, content, show, moveTo, hide };
};
