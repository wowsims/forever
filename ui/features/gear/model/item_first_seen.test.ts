import { afterEach, describe, expect, it, vi } from 'vitest';

import { isNewlyDiscovered } from './item_first_seen';

const NOW_MS = Date.UTC(2026, 10, 8, 12);
const NOW_SECONDS = NOW_MS / 1000;
const WINDOW_SECONDS = 14 * 24 * 60 * 60;

describe('isNewlyDiscovered', () => {
	it.each([
		['an item first seen this second', NOW_SECONDS, true],
		['an item first seen exactly at the window’s start', NOW_SECONDS - WINDOW_SECONDS, true],
		['an item first seen one second before the window', NOW_SECONDS - WINDOW_SECONDS - 1, false],
		['an item first seen long ago', 1668842294, false],
	])('counts %s as %s', (_name, firstSeen, expected) => {
		expect(isNewlyDiscovered(firstSeen, NOW_MS)).toBe(expected);
	});
});

describe('loadItemFirstSeen', () => {
	// The loader keeps its promise for the page, so each case loads the module afresh.
	const freshLoader = async () => {
		vi.resetModules();
		return (await import('./item_first_seen')).loadItemFirstSeen;
	};

	afterEach(() => {
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
	});

	it('maps each item id to its first-seen second, fetching once for every caller', async () => {
		const fetch = vi.fn(async () =>
			Response.json([
				{ id: 720, firstSeen: 1668842294 },
				{ id: 744, firstSeen: NOW_SECONDS },
			]),
		);
		vi.stubGlobal('fetch', fetch);
		const load = await freshLoader();

		const [first, second] = await Promise.all([load(), load()]);

		expect(first).toBe(second);
		expect([...first]).toEqual([
			[720, 1668842294],
			[744, NOW_SECONDS],
		]);
		expect(fetch).toHaveBeenCalledOnce();
		expect(fetch).toHaveBeenCalledWith('/forever/assets/database/item_first_seen.json');
	});

	it.each([
		['a missing file', async () => new Response('Not found', { status: 404 })],
		['a failed request', async () => Promise.reject(new TypeError('Failed to fetch'))],
	])('gives an empty map and one warning for %s, and does not fetch again', async (_name, respond) => {
		const fetch = vi.fn(respond);
		vi.stubGlobal('fetch', fetch);
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const load = await freshLoader();

		expect((await load()).size).toBe(0);
		expect((await load()).size).toBe(0);
		expect(warn).toHaveBeenCalledOnce();
		expect(fetch).toHaveBeenCalledOnce();
	});
});
