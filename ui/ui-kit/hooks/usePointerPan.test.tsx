import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { usePointerPan } from './usePointerPan';

const setup = ({ page = false } = {}) => {
	const click = vi.fn();
	const outer = { current: null as HTMLDivElement | null };
	const Harness = () => {
		const { panning, handlers } = usePointerPan({ verticalScroller: () => (page ? null : outer.current) });
		return (
			<div ref={element => void (outer.current = element)} data-testid="outer">
				<div data-testid="scroller" data-panning={panning ? '' : undefined} {...handlers} onClick={click}>
					<span data-testid="empty" />
					<button data-testid="button" />
					<div role="slider" aria-valuenow={0} tabIndex={0} data-testid="bar">
						<span data-testid="handle" />
					</div>
				</div>
			</div>
		);
	};
	render(<Harness />);
	const scroller = screen.getByTestId('scroller');
	scroller.setPointerCapture = vi.fn();
	scroller.hasPointerCapture = vi.fn(() => true);
	scroller.releasePointerCapture = vi.fn();
	scroller.scrollLeft = 300;
	screen.getByTestId('outer').scrollTop = 200;
	return { click, scroller, outer: screen.getByTestId('outer') };
};

const down = (target: Element, init: Record<string, unknown> = {}) =>
	act(() => void fireEvent.pointerDown(target, { pointerId: 1, button: 0, pointerType: 'mouse', clientX: 100, clientY: 100, ...init }));
const move = (target: Element, clientX: number, clientY = 100, pointerId = 1) => act(() => void fireEvent.pointerMove(target, { pointerId, clientX, clientY }));
const up = (target: Element) => act(() => void fireEvent.pointerUp(target, { pointerId: 1 }));

afterEach(() => {
	vi.restoreAllMocks();
});

describe('usePointerPan', () => {
	it('waits for 4px, then scrolls both axes against the pointer and shows the grab while it does', () => {
		const { scroller, outer } = setup();
		down(screen.getByTestId('empty'));
		move(scroller, 103, 102);
		expect(scroller.scrollLeft).toBe(300);
		expect(scroller.hasAttribute('data-panning')).toBe(false);
		move(scroller, 60, 130);
		expect(scroller.scrollLeft).toBe(340);
		expect(outer.scrollTop).toBe(170);
		expect(scroller.hasAttribute('data-panning')).toBe(true);
		expect(scroller.setPointerCapture).toHaveBeenCalledWith(1);
		up(scroller);
		expect(scroller.hasAttribute('data-panning')).toBe(false);
		expect(scroller.releasePointerCapture).toHaveBeenCalledWith(1);
	});

	it('scrolls the page when there is no vertical scroller', () => {
		const scrollBy = vi.spyOn(window, 'scrollBy').mockImplementation(() => {});
		const { scroller } = setup({ page: true });
		down(scroller);
		move(scroller, 100, 140);
		expect(scrollBy).toHaveBeenLastCalledWith(0, -40 - window.scrollY);
	});

	it('ignores touch, other buttons and other pointers', () => {
		const { scroller } = setup();
		down(scroller, { pointerType: 'touch' });
		down(scroller, { button: 2 });
		move(scroller, 0);
		expect(scroller.scrollLeft).toBe(300);
		down(scroller);
		move(scroller, 0, 100, 2);
		expect(scroller.scrollLeft).toBe(300);
	});

	it('leaves presses on buttons, sliders and their parts to them', () => {
		const { scroller } = setup();
		for (const id of ['button', 'bar', 'handle']) {
			down(screen.getByTestId(id));
			move(scroller, 0);
		}
		expect(scroller.scrollLeft).toBe(300);
		expect(scroller.setPointerCapture).not.toHaveBeenCalled();
	});

	it('suppresses the click after a pan, once, and lets a plain press click', () => {
		const { click, scroller } = setup();
		down(scroller);
		up(scroller);
		fireEvent.click(scroller);
		expect(click).toHaveBeenCalledTimes(1);
		down(scroller);
		move(scroller, 50);
		up(scroller);
		fireEvent.click(scroller);
		expect(click).toHaveBeenCalledTimes(1);
		fireEvent.click(scroller);
		expect(click).toHaveBeenCalledTimes(2);
	});

	it('ends when the pointer capture is lost, so the next press pans again', () => {
		const { scroller } = setup();
		down(scroller);
		move(scroller, 50);
		act(() => void fireEvent.lostPointerCapture(scroller, { pointerId: 1 }));
		move(scroller, 0);
		expect(scroller.scrollLeft).toBe(350);
		down(scroller, { pointerId: 3 });
		move(scroller, 90, 100, 3);
		expect(scroller.scrollLeft).toBe(360);
	});

	it('clears its state when setPointerCapture throws', () => {
		const { scroller } = setup();
		scroller.setPointerCapture = vi.fn(() => {
			throw new Error('no such pointer');
		});
		down(scroller);
		move(scroller, 0);
		expect(scroller.scrollLeft).toBe(300);
	});
});
