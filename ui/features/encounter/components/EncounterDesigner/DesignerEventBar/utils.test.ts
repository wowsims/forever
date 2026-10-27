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
		expect(barKeyAction(key('ArrowRight'), false)).toEqual({ type: 'edit', mode: 'move', steps: 1 });
		expect(barKeyAction(key('ArrowLeft', { shiftKey: true }), false)).toEqual({ type: 'edit', mode: 'move', steps: -5 });
	});

	it('resizes the end with Alt and the start with Alt+Shift', () => {
		expect(barKeyAction(key('ArrowRight', { altKey: true }), false)).toEqual({ type: 'edit', mode: 'end', steps: 1 });
		expect(barKeyAction(key('ArrowLeft', { altKey: true, shiftKey: true }), false)).toEqual({ type: 'edit', mode: 'start', steps: -1 });
	});

	it('does not resize a point', () => {
		expect(barKeyAction(key('ArrowRight', { altKey: true }), true)).toBeUndefined();
		expect(barKeyAction(key('ArrowRight'), true)).toEqual({ type: 'edit', mode: 'move', steps: 1 });
	});

	it('maps the other keys and leaves Ctrl and Meta chords alone', () => {
		expect(barKeyAction(key('ArrowUp'), false)).toEqual({ type: 'sibling', direction: -1 });
		expect(barKeyAction(key('ArrowDown'), false)).toEqual({ type: 'sibling', direction: 1 });
		expect(barKeyAction(key('Delete'), false)).toEqual({ type: 'remove' });
		expect(barKeyAction(key('Backspace'), false)).toEqual({ type: 'remove' });
		expect(barKeyAction(key('Enter'), false)).toEqual({ type: 'activate' });
		expect(barKeyAction(key('ArrowRight', { ctrlKey: true }), false)).toBeUndefined();
		expect(barKeyAction(key('a'), false)).toBeUndefined();
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
});
