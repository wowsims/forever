import { describe, expect, it } from 'vitest';

import { ZoomController } from './zoom';

// A scroller laid out at a fractional width, which clientWidth rounds to the nearest pixel.
const scrollerOf = (width: number) => {
	const scroller = document.createElement('div');
	Object.defineProperty(scroller, 'clientWidth', { value: Math.round(width) });
	scroller.getBoundingClientRect = () => DOMRect.fromRect({ width, height: 100 });
	return scroller;
};

describe('ZoomController.fitToWidth', () => {
	it('fits the axis inside a fractional scroller, never past it', () => {
		const scroller = scrollerOf(1341.6);
		const zoom = new ZoomController({ scroller, styleHosts: [], labelWidth: () => 192, inset: () => 24, scrollVerticalBy: () => {}, onChange: () => {} });
		zoom.setDuration(180);
		zoom.fitToWidth();
		expect(zoom.pps * 180 + 192 + 2 * 24).toBeCloseTo(1341, 9);
	});

	it('keeps the whole width of a scroller clientWidth rounds down', () => {
		const scroller = scrollerOf(1341.4);
		const zoom = new ZoomController({ scroller, styleHosts: [], labelWidth: () => 192, scrollVerticalBy: () => {}, onChange: () => {} });
		zoom.setDuration(100);
		zoom.fitToWidth();
		expect(zoom.pps * 100 + 192).toBeCloseTo(1341, 9);
	});
});
