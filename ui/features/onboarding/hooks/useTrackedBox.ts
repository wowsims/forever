import { useLayoutEffect, useState } from 'react';

import type { Box } from '../model/placement';

/** When a found element is measured again: a part still sliding or loading into place settles without a scroll or resize to say so. */
const SETTLE_CHECKS = [120, 400, 800];

const toBox = (rect: DOMRect): Box => ({ left: rect.left, top: rect.top, width: rect.width, height: rect.height });

/** The element's viewport box, followed as the page scrolls, resizes or reflows; null with no element. */
export const useTrackedBox = (element: Element | null): Box | null => {
	const [box, setBox] = useState<Box | null>(null);
	useLayoutEffect(() => {
		if (!element) {
			setBox(null);
			return;
		}
		let frame = 0;
		const measure = () => {
			cancelAnimationFrame(frame);
			frame = requestAnimationFrame(() => setBox(toBox(element.getBoundingClientRect())));
		};
		setBox(toBox(element.getBoundingClientRect()));
		const settle = SETTLE_CHECKS.map(delay => window.setTimeout(measure, delay));
		const observer = typeof ResizeObserver === 'undefined' ? undefined : new ResizeObserver(measure);
		observer?.observe(element);
		window.addEventListener('resize', measure);
		window.addEventListener('scroll', measure, true);
		return () => {
			cancelAnimationFrame(frame);
			settle.forEach(timer => clearTimeout(timer));
			observer?.disconnect();
			window.removeEventListener('resize', measure);
			window.removeEventListener('scroll', measure, true);
		};
	}, [element]);
	return box;
};
