import { describe, expect, it } from 'vitest';

import { barKeyAction, editWindow } from './utils';

const key = (name: string, modifiers: { shiftKey?: boolean; altKey?: boolean; ctrlKey?: boolean; metaKey?: boolean } = {}) => ({
	key: name,
	shiftKey: false,
	altKey: false,
	ctrlKey: false,
	metaKey: false,
	...modifiers,
});

describe('barKeyAction', () => {
	it('moves by one step, five with Shift', () => {
		expect(barKeyAction(key('ArrowRight'))).toEqual({ type: 'edit', mode: 'move', steps: 1 });
		expect(barKeyAction(key('ArrowLeft', { shiftKey: true }))).toEqual({ type: 'edit', mode: 'move', steps: -5 });
	});

	it('resizes the end with Alt and the start with Alt+Shift', () => {
		expect(barKeyAction(key('ArrowRight', { altKey: true }))).toEqual({ type: 'edit', mode: 'end', steps: 1 });
		expect(barKeyAction(key('ArrowLeft', { altKey: true, shiftKey: true }))).toEqual({ type: 'edit', mode: 'start', steps: -1 });
	});

	it('duplicates on Ctrl+D or Cmd+D and leaves the other chords alone', () => {
		expect(barKeyAction(key('d', { ctrlKey: true }))).toEqual({ type: 'duplicate' });
		expect(barKeyAction(key('D', { metaKey: true }))).toEqual({ type: 'duplicate' });
		expect(barKeyAction(key('d'))).toBeUndefined();
		expect(barKeyAction(key('d', { ctrlKey: true, shiftKey: true }))).toBeUndefined();
	});

	it('maps the other keys and leaves Ctrl and Meta chords alone', () => {
		expect(barKeyAction(key('ArrowUp'))).toEqual({ type: 'sibling', direction: -1 });
		expect(barKeyAction(key('ArrowDown'))).toEqual({ type: 'sibling', direction: 1 });
		expect(barKeyAction(key('Delete'))).toEqual({ type: 'remove' });
		expect(barKeyAction(key('Backspace'))).toEqual({ type: 'remove' });
		expect(barKeyAction(key('Enter'))).toEqual({ type: 'activate' });
		expect(barKeyAction(key('ArrowRight', { ctrlKey: true }))).toBeUndefined();
		expect(barKeyAction(key('a'))).toBeUndefined();
	});
});

describe('editWindow', () => {
	const window = { start: 0.2, end: 0.4 };

	it('moves the whole window and stops at the fight edges', () => {
		expect(editWindow(window, 'move', 0.1, 0.01).start).toBeCloseTo(0.3);
		expect(editWindow(window, 'move', 0.1, 0.01).end).toBeCloseTo(0.5);
		expect(editWindow(window, 'move', -1, 0.01)).toEqual({ start: 0, end: expect.closeTo(0.2) });
	});

	it('resizes one edge and keeps the minimum width', () => {
		expect(editWindow(window, 'start', 0.1, 0.01)).toEqual({ start: expect.closeTo(0.3), end: 0.4 });
		expect(editWindow(window, 'start', 0.5, 0.05)).toEqual({ start: expect.closeTo(0.35), end: 0.4 });
		expect(editWindow(window, 'end', -0.5, 0.05)).toEqual({ start: 0.2, end: expect.closeTo(0.25) });
	});

	it('keeps an open end past the fight when the window moves or its end grows', () => {
		const pastTheEnd = { start: 0.9, end: 1.1 };
		expect(editWindow(pastTheEnd, 'move', 0.02, 0.01, true)).toEqual({ start: expect.closeTo(0.92), end: expect.closeTo(1.12) });
		expect(editWindow(pastTheEnd, 'end', 0.1, 0.01, true).end).toBeCloseTo(1.2);
		expect(editWindow(window, 'move', 0.9, 0.01, true)).toEqual({ start: 1, end: expect.closeTo(1.2) });
		expect(editWindow(window, 'move', 0.9, 0.01).end).toBe(1);
	});
});
