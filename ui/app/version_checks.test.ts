import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { checkWhileOpen, VERSION_CHECK_EVERY_MS } from './version_checks';

const setVisibility = (state: DocumentVisibilityState) => {
	Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => state });
	document.dispatchEvent(new Event('visibilitychange'));
};

describe('checkWhileOpen', () => {
	let stop = () => {};

	beforeEach(() => {
		vi.useFakeTimers();
		setVisibility('visible');
	});

	afterEach(() => {
		stop();
		vi.useRealTimers();
	});

	it('checks on the interval while the page is visible, and not before', () => {
		const check = vi.fn();
		stop = checkWhileOpen(check).stop;

		vi.advanceTimersByTime(VERSION_CHECK_EVERY_MS - 1);
		expect(check).not.toHaveBeenCalled();
		vi.advanceTimersByTime(1);
		expect(check).toHaveBeenCalledTimes(1);
	});

	it('waits while hidden and checks once the page is visible again', () => {
		const check = vi.fn();
		stop = checkWhileOpen(check).stop;

		setVisibility('hidden');
		vi.advanceTimersByTime(2 * VERSION_CHECK_EVERY_MS);
		expect(check).not.toHaveBeenCalled();

		setVisibility('visible');
		expect(check).toHaveBeenCalledTimes(1);
	});

	it('checks when Back or Forward restores the page from memory', () => {
		const check = vi.fn();
		stop = checkWhileOpen(check).stop;

		window.dispatchEvent(Object.assign(new Event('pageshow'), { persisted: false }));
		window.dispatchEvent(Object.assign(new Event('pageshow'), { persisted: true }));

		expect(check).toHaveBeenCalledTimes(1);
	});

	it('stops checking once stopped', () => {
		const check = vi.fn();
		checkWhileOpen(check).stop();

		vi.advanceTimersByTime(2 * VERSION_CHECK_EVERY_MS);
		window.dispatchEvent(Object.assign(new Event('pageshow'), { persisted: true }));

		expect(check).not.toHaveBeenCalled();
	});
});
