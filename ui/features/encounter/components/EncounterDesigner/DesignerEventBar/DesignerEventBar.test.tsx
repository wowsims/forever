import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { DesignerEventBar, type DesignerEventBarProps } from './DesignerEventBar';

const BASE_SECONDS = 180;
const PPS = 5;

const props = (): DesignerEventBarProps => ({
	name: 'Wave 1',
	window: { start: 20 / BASE_SECONDS, end: 60 / BASE_SECONDS },
	row: 0,
	baseSeconds: BASE_SECONDS,
	snapSeconds: 1,
	pixelsPerSecond: () => PPS,
	valueText: () => 'Wave 1',
	tabIndex: 0,
	onChange: vi.fn(),
	onSelect: vi.fn(),
	onActivate: vi.fn(),
	onRemove: vi.fn(),
	onDuplicate: vi.fn(),
	onRenameStart: vi.fn(),
	onRename: vi.fn(),
	onSibling: vi.fn(),
});

// happy-dom has no pointer capture, which the drag hook needs to start.
beforeEach(() => {
	vi.spyOn(Element.prototype, 'setPointerCapture').mockImplementation(() => {});
	vi.spyOn(Element.prototype, 'hasPointerCapture').mockReturnValue(false);
});
afterEach(() => vi.restoreAllMocks());

const drag = (bar: HTMLElement, dx: number, altKey = false, during?: () => void) => {
	const pointer = { pointerId: 1, pointerType: 'mouse', button: 0, altKey };
	act(() => void fireEvent.pointerDown(bar, { ...pointer, clientX: 100 }));
	act(() => void fireEvent.pointerMove(bar, { ...pointer, clientX: 100 + dx }));
	during?.();
	act(() => void fireEvent.pointerUp(bar, { ...pointer, clientX: 100 + dx }));
};

describe('DesignerEventBar drag', () => {
	it('moves the event with a drag: one change, snapped to the step', () => {
		const p = props();
		render(<DesignerEventBar {...p} />);
		drag(screen.getByRole('slider'), 50);
		expect(p.onChange).toHaveBeenCalledTimes(1);
		const next = vi.mocked(p.onChange).mock.calls[0][0];
		expect(next.start).toBeCloseTo(30 / BASE_SECONDS);
		expect(next.end).toBeCloseTo(70 / BASE_SECONDS);
		expect(p.onDuplicate).not.toHaveBeenCalled();
	});

	it('duplicates instead on an Alt drag: the original stays, one copy at the dropped place', () => {
		const p = props();
		render(<DesignerEventBar {...p} />);
		drag(screen.getByRole('slider'), 50, true, () => {
			expect(document.querySelector('[data-copy]')).not.toBeNull();
			expect(screen.getByRole('slider').getAttribute('aria-valuenow')).toBe('20');
		});
		expect(p.onChange).not.toHaveBeenCalled();
		expect(p.onSelect).not.toHaveBeenCalled();
		expect(p.onDuplicate).toHaveBeenCalledTimes(1);
		expect(vi.mocked(p.onDuplicate).mock.calls[0][0]).toBeCloseTo(10 / BASE_SECONDS);
		expect(document.querySelector('[data-copy]')).toBeNull();
	});

	it('discards the copy of an Alt drag, and the move of a plain one, on Escape', () => {
		for (const altKey of [true, false]) {
			const p = props();
			const { unmount } = render(<DesignerEventBar {...p} />);
			drag(screen.getByRole('slider'), 50, altKey, () => act(() => void fireEvent.keyDown(screen.getByRole('slider'), { key: 'Escape' })));
			expect(p.onChange).not.toHaveBeenCalled();
			expect(p.onDuplicate).not.toHaveBeenCalled();
			unmount();
		}
	});

	it('duplicates in place offset by one step on Ctrl+D', () => {
		const p = props();
		render(<DesignerEventBar {...p} snapSeconds={2} />);
		act(() => void fireEvent.keyDown(screen.getByRole('slider'), { key: 'd', ctrlKey: true }));
		expect(p.onDuplicate).toHaveBeenCalledWith(2 / BASE_SECONDS);
	});
});

describe('DesignerEventBar rename', () => {
	it('starts on a double click of the label and on F2', () => {
		const p = props();
		render(<DesignerEventBar {...p} />);
		act(() => void fireEvent.doubleClick(screen.getByText('Wave 1')));
		act(() => void fireEvent.keyDown(screen.getByRole('slider'), { key: 'F2' }));
		expect(p.onRenameStart).toHaveBeenCalledTimes(2);
	});

	it('commits the text on Enter, once', () => {
		const p = props();
		render(<DesignerEventBar {...p} renaming />);
		const input = screen.getByRole('textbox') as HTMLInputElement;
		expect(input.value).toBe('Wave 1');
		act(() => void fireEvent.change(input, { target: { value: 'Adds' } }));
		act(() => void fireEvent.keyDown(input, { key: 'Enter' }));
		act(() => void fireEvent.blur(input));
		expect(p.onRename).toHaveBeenCalledExactlyOnceWith('Adds');
		expect(p.onRemove).not.toHaveBeenCalled();
	});

	it('drops the text on Escape', () => {
		const p = props();
		render(<DesignerEventBar {...p} renaming />);
		act(() => void fireEvent.keyDown(screen.getByRole('textbox'), { key: 'Escape' }));
		act(() => void fireEvent.blur(screen.getByRole('textbox')));
		expect(p.onRename).toHaveBeenCalledExactlyOnceWith(undefined);
	});

	it('keeps the keys of the text field off the bar', () => {
		const p = props();
		render(<DesignerEventBar {...p} renaming />);
		act(() => void fireEvent.keyDown(screen.getByRole('textbox'), { key: 'Delete' }));
		expect(p.onRemove).not.toHaveBeenCalled();
	});
});
