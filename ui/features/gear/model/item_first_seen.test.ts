import i18n from '@i18n/config';
import { describe, expect, it, vi } from 'vitest';

import { discoveredLabel, isNewlyDiscovered } from './item_first_seen';

// Local noon, so the calendar-day cases below sit well inside their days.
const NOW_MS = new Date(2026, 10, 8, 12).getTime();
const NOW_SECONDS = NOW_MS / 1000;
const HOUR_SECONDS = 60 * 60;
const DAY_SECONDS = 24 * HOUR_SECONDS;
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
		['59 minutes ago', NOW_SECONDS - 59 * 60, 'gear_tab.item_notices.discovered_within_hour'],
		['an hour ago', NOW_SECONDS - HOUR_SECONDS, 'gear_tab.item_notices.discovered_ago'],
		['late last night, under a day ago', new Date(2026, 10, 7, 23, 30).getTime() / 1000, 'gear_tab.item_notices.discovered_ago'],
		['a day and an hour ago', NOW_SECONDS - DAY_SECONDS - HOUR_SECONDS, 'gear_tab.item_notices.discovered_yesterday'],
		['two days ago', NOW_SECONDS - 2 * DAY_SECONDS, 'gear_tab.item_notices.discovered_on'],
	])('names an item discovered %s', (_name, firstSeen, key) => {
		vi.spyOn(Date, 'now').mockReturnValue(NOW_MS);
		expect(discoveredLabel(firstSeen)).toBe(key);
		vi.restoreAllMocks();
	});

	it.each([
		[HOUR_SECONDS, '1 hour ago'],
		[5 * HOUR_SECONDS + 59 * 60, '5 hours ago'],
		[23 * HOUR_SECONDS, '23 hours ago'],
	])('counts whole hours for an item discovered %i seconds ago', (secondsAgo, when) => {
		vi.spyOn(Date, 'now').mockReturnValue(NOW_MS);
		const t = vi.spyOn(i18n, 't');
		discoveredLabel(NOW_SECONDS - secondsAgo);
		expect(t).toHaveBeenCalledWith('gear_tab.item_notices.discovered_ago', { when });
		vi.restoreAllMocks();
	});
});
