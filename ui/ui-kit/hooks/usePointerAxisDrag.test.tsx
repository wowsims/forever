import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { usePointerAxisDrag } from './usePointerAxisDrag';

const setup = () => {
	const calls = { start: vi.fn(), delta: vi.fn(), end: vi.fn(), click: vi.fn() };
	const Harness = () => {
		const handlers = usePointerAxisDrag({ pixelsPerSecond: () => 10, onStart: calls.start, onDelta: calls.delta, onEnd: calls.end });
		return (
			<div data-testid="bar" {...handlers} onClick={calls.click}>
				<span data-testid="left" data-handle="start" />
				<span data-testid="right" data-handle="end" />
				<span data-testid="body" />
				<button data-testid="button" />
			</div>
		);
	};
	render(<Harness />);
	const bar = screen.getByTestId('bar');
	bar.setPointerCapture = vi.fn();
	bar.hasPointerCapture = vi.fn(() => true);
	bar.releasePointerCapture = vi.fn();
	return { calls, bar };
};

const down = (target: Element, init: Record<string, unknown> = {}) => fireEvent.pointerDown(target, { pointerId: 1, button: 0, clientX: 100, ...init });

describe('usePointerAxisDrag', () => {
	it('waits for 4px, then reports seconds since the pointer went down', () => {
		const { calls, bar } = setup();
		down(bar);
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 103 });
		expect(calls.start).not.toHaveBeenCalled();
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 120 });
		expect(calls.start).toHaveBeenCalledWith('move');
		expect(calls.delta).toHaveBeenLastCalledWith(2, 'move');
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 90 });
		expect(calls.delta).toHaveBeenLastCalledWith(-1, 'move');
		expect(bar.setPointerCapture).toHaveBeenCalledWith(1);
		fireEvent.pointerUp(bar, { pointerId: 1 });
		expect(calls.end).toHaveBeenCalledWith(true);
	});

	it('reads the mode from the nearest data-handle', () => {
		const { calls, bar } = setup();
		down(screen.getByTestId('left'));
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		expect(calls.delta).toHaveBeenLastCalledWith(5, 'start');
		fireEvent.pointerUp(bar, { pointerId: 1 });
		down(screen.getByTestId('right'));
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		expect(calls.delta).toHaveBeenLastCalledWith(5, 'end');
		fireEvent.pointerUp(bar, { pointerId: 1 });
		down(screen.getByTestId('body'));
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		expect(calls.delta).toHaveBeenLastCalledWith(5, 'move');
	});

	it('ignores touch, other buttons and other pointers', () => {
		const { calls, bar } = setup();
		down(bar, { pointerType: 'touch' });
		down(bar, { button: 2 });
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 200 });
		expect(calls.delta).not.toHaveBeenCalled();
		down(bar);
		fireEvent.pointerMove(bar, { pointerId: 2, clientX: 200 });
		expect(calls.delta).not.toHaveBeenCalled();
	});

	it('cancels with onEnd(false) on pointercancel', () => {
		const { calls, bar } = setup();
		down(bar);
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		fireEvent.pointerCancel(bar, { pointerId: 1 });
		expect(calls.end).toHaveBeenCalledWith(false);
	});

	it('does not end a press that never became a drag, and lets its click through', () => {
		const { calls, bar } = setup();
		down(bar);
		fireEvent.pointerUp(bar, { pointerId: 1 });
		fireEvent.click(bar);
		expect(calls.end).not.toHaveBeenCalled();
		expect(calls.click).toHaveBeenCalledTimes(1);
	});

	it('suppresses the click after a drag, once', () => {
		const { calls, bar } = setup();
		down(bar);
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		fireEvent.pointerUp(bar, { pointerId: 1 });
		fireEvent.click(bar);
		expect(calls.click).not.toHaveBeenCalled();
		fireEvent.click(bar);
		expect(calls.click).toHaveBeenCalledTimes(1);
	});

	it('leaves presses on interactive children alone', () => {
		const { calls, bar } = setup();
		down(screen.getByTestId('button'));
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		expect(calls.delta).not.toHaveBeenCalled();
	});

	it('ends without committing when the pointer capture is lost', () => {
		const { calls, bar } = setup();
		down(bar);
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		fireEvent.lostPointerCapture(bar, { pointerId: 1 });
		expect(calls.end).toHaveBeenCalledWith(false);
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 160 });
		expect(calls.delta).toHaveBeenCalledTimes(1);
	});

	it('clears its state when setPointerCapture throws', () => {
		const { calls, bar } = setup();
		bar.setPointerCapture = vi.fn(() => {
			throw new Error('no such pointer');
		});
		down(bar);
		fireEvent.pointerMove(bar, { pointerId: 1, clientX: 150 });
		expect(calls.delta).not.toHaveBeenCalled();
	});
});
