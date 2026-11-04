import { type RefObject, useEffect, useRef, useState } from 'react';

/** Paints in CSS pixels; `width` and `height` are the canvas's laid-out size. */
export type CanvasPaint = (ctx: CanvasRenderingContext2D, width: number, height: number) => void;

/**
 * Keeps a canvas's backing store at its laid-out size times the device pixel ratio, so lines stay sharp on a
 * HiDPI screen, and repaints it with `paint` whenever the canvas resizes or `paint` changes. A canvas with no
 * size, as in a hidden tab, is left alone until it has one.
 */
export function useCanvasPainter(paint: CanvasPaint): RefObject<HTMLCanvasElement | null> {
	const ref = useRef<HTMLCanvasElement | null>(null);
	const [size, setSize] = useState({ width: 0, height: 0 });

	useEffect(() => {
		const canvas = ref.current;
		if (!canvas) return;
		const observer = new ResizeObserver(() => setSize({ width: canvas.clientWidth, height: canvas.clientHeight }));
		observer.observe(canvas);
		return () => observer.disconnect();
	}, []);

	useEffect(() => {
		const canvas = ref.current;
		const ctx = canvas?.getContext('2d');
		if (!canvas || !ctx || size.width === 0 || size.height === 0) return;
		const ratio = window.devicePixelRatio || 1;
		const width = Math.round(size.width * ratio);
		const height = Math.round(size.height * ratio);
		if (canvas.width !== width) canvas.width = width;
		if (canvas.height !== height) canvas.height = height;
		ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
		ctx.clearRect(0, 0, size.width, size.height);
		paint(ctx, size.width, size.height);
	}, [paint, size]);

	return ref;
}
