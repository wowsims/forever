import { Spec } from '@generated/proto/common';
import type { Database } from '@sim/proto/database';
import { render } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ENCHANT_NOTICES, ITEM_INFO_NOTICES, ITEM_NOTICES, registerAreaStatsNotices, registerSetBonusNotices } from './item_notices';

vi.mock('@i18n/localization', () => ({
	translateAreaType: (value: number) => `area-${value}`,
	translateStat: (value: number) => `stat-${value}`,
}));

vi.mock('@sim/constants/missing_effects_auto_gen', () => ({
	MISSING_ITEM_EFFECTS: new Map([[1, ['Does a thing.']]]),
	MISSING_ENCHANT_EFFECTS: new Map([[2, ['Enchants a thing.']]]),
}));

const notice = (node: ReactNode) => {
	const { container } = render(<>{node}</>);
	return { heading: container.querySelector('p')?.textContent, items: [...container.querySelectorAll('li')].map(item => item.textContent) };
};

describe('the item notice table', () => {
	it.each([
		['item', ITEM_NOTICES, 1, 'Does a thing.'],
		['enchant', ENCHANT_NOTICES, 2, 'Enchants a thing.'],
	] as const)('lists the tooltips a missing %s effect carries', (kind, table, id, tooltip) => {
		expect(notice(table.get(id)?.[Spec.SpecUnknown])).toEqual({
			heading: `The following ${kind} effect (on-use or proc) is not implemented!`,
			items: [tooltip],
		});
	});
});

describe('registerSetBonusNotices', () => {
	const SET_ID = 9999;
	const ITEM_IDS = [90001, 90002];

	afterEach(() => {
		ITEM_IDS.forEach(id => ITEM_NOTICES.delete(id));
	});

	it('writes no notice while no set is listed', () => {
		registerSetBonusNotices({ getItemIdsForSet: (setId: number) => (setId === SET_ID ? ITEM_IDS : []) } as unknown as Database);

		for (const id of ITEM_IDS) {
			expect(ITEM_NOTICES.has(id)).toBe(false);
		}
	});
});

describe('registerAreaStatsNotices', () => {
	const RUNE = 90010;
	const PLAIN = 90011;
	const db = {
		getAllItems: () => [
			{ id: RUNE, scalingOptions: { 0: { stats: { 17: 15 }, areaStats: [{ areaType: 1, stats: { 17: 29, 18: 29 } }] } } },
			{ id: PLAIN, scalingOptions: { 0: { stats: { 17: 15 }, areaStats: [] } } },
		],
	} as unknown as Database;

	afterEach(() => {
		ITEM_INFO_NOTICES.delete(RUNE);
		ITEM_INFO_NOTICES.delete(PLAIN);
	});

	it('writes an info notice naming each area and the stats it adds, and none for an item without any', () => {
		registerAreaStatsNotices(db);

		expect(notice(ITEM_INFO_NOTICES.get(RUNE))).toEqual({
			heading: 'Only while the encounter is in one of these areas:',
			items: ['area-1: +29 stat-17, +29 stat-18'],
		});
		expect(ITEM_INFO_NOTICES.has(PLAIN)).toBe(false);
		expect(ITEM_NOTICES.has(RUNE)).toBe(false);
	});
});
