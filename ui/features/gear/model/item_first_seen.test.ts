import { describe, expect, it, vi } from 'vitest';

import { discoveredLabel, isNewlyDiscovered } from './item_first_seen';

// Local noon, so the calendar-day cases below sit well inside their days.
const NOW_MS = new Date(2026, 10, 8, 12).getTime();
const NOW_SECONDS = NOW_MS / 1000;
const DAY_SECONDS = 24 * 60 * 60;
const WINDOW_SECONDS = 7 * DAY_SECONDS;

describe('isNewlyDiscovered', () => {
	it.each([
		['an item first seen this second', NOW_SECONDS, true],
		['an item first seen exactly at the window’s start', NOW_SECONDS - WINDOW_SECONDS, true],
		['an item first seen one second before the window', NOW_SECONDS - WINDOW_SECONDS - 1, false],
		['an item first seen long ago', 1668842294, false],
	])('counts %s as %s', (_name, firstSeen, expected) => {
		vi.spyOn(Date, 'now').mockReturnValue(NOW_MS);
		expect(isNewlyDiscovered(firstSeen)).toBe(expected);
		vi.restoreAllMocks();
	});
});

describe('discoveredLabel', () => {
	it.each([
		['this morning', new Date(2026, 10, 8, 1).getTime() / 1000, 'gear_tab.item_notices.discovered_today'],
		['late last night', new Date(2026, 10, 7, 23, 30).getTime() / 1000, 'gear_tab.item_notices.discovered_yesterday'],
		['two days ago', NOW_SECONDS - 2 * DAY_SECONDS, 'gear_tab.item_notices.discovered_on'],
	])('names an item discovered %s', (_name, firstSeen, key) => {
		vi.spyOn(Date, 'now').mockReturnValue(NOW_MS);
		expect(discoveredLabel(firstSeen)).toBe(key);
		vi.restoreAllMocks();
	});
});
