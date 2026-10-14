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
	UNVALIDATED_RATE_ITEMS: new Map([
		[1, { kind: 'ppm', origin: 'vanilla' }],
		[3, { kind: 'ppm', origin: 'vanilla' }],
		[4, { kind: 'ppm', origin: 'tbc' }],
		[5, { kind: 'chance', origin: 'vanilla' }],
		[6, { kind: 'chance', origin: 'tbc' }],
		[7, { kind: 'ppm', origin: 'unknown' }],
		[8, { kind: 'chance', origin: 'unknown' }],
	]),
	UNVALIDATED_RATE_ENCHANTS: new Map([
		[2, { kind: 'ppm', origin: 'tbc' }],
		[9, { kind: 'chance', origin: 'vanilla' }],
	]),
	UNVALIDATED_RATE_SET_BONUSES: new Map([
		[
			9999,
			new Map([
				[2, { kind: 'chance', origin: 'tbc' }],
				[4, { kind: 'ppm', origin: 'vanilla' }],
			]),
		],
	]),
}));

const VANILLA_PPM_WARNING = 'gear_tab.item_notices.unvalidated_rate.ppm_vanilla';
const TBC_PPM_WARNING = 'gear_tab.item_notices.unvalidated_rate.ppm_tbc';
const VANILLA_CHANCE_WARNING = 'gear_tab.item_notices.unvalidated_rate.chance_vanilla';
const TBC_CHANCE_WARNING = 'gear_tab.item_notices.unvalidated_rate.chance_tbc';

const notice = (node: ReactNode) => {
	const { container } = render(<>{node}</>);
	return { heading: container.querySelector('p')?.textContent, items: [...container.querySelectorAll('li')].map(item => item.textContent) };
};

const paragraphs = (node: ReactNode) => {
	const { container } = render(<>{node}</>);
	return [...container.querySelectorAll('p')].map(paragraph => paragraph.textContent);
};

describe('the item notice table', () => {
	it.each([
		['item', ITEM_NOTICES, 1, 'Does a thing.'],
		['enchant', ENCHANT_NOTICES, 2, 'Enchants a thing.'],
	] as const)('lists the tooltips a missing %s effect carries', (kind, table, id, tooltip) => {
		expect(notice(table.get(id)?.[Spec.SpecUnknown])).toEqual({
			heading: `gear_tab.item_notices.missing_effect.${kind}_listed`,
			items: [tooltip],
		});
	});

	it.each([
		['a vanilla rate per minute', 3, VANILLA_PPM_WARNING],
		['a TBC rate per minute', 4, TBC_PPM_WARNING],
		['a vanilla chance', 5, VANILLA_CHANCE_WARNING],
		['a TBC chance', 6, TBC_CHANCE_WARNING],
		['a rate per minute of no known origin', 7, 'gear_tab.item_notices.unvalidated_rate.ppm_unknown'],
		['a chance of no known origin', 8, 'gear_tab.item_notices.unvalidated_rate.chance_unknown'],
	])('warns that %s needs validating on an item with no other notice', (_, id, warning) => {
		expect(paragraphs(ITEM_NOTICES.get(id)?.[Spec.SpecUnknown])).toEqual([warning]);
	});

	it.each([
		['item', ITEM_NOTICES, 1, VANILLA_PPM_WARNING],
		['enchant', ENCHANT_NOTICES, 2, TBC_PPM_WARNING],
	] as const)('keeps the missing effect notice of an %s with an unvalidated rate and adds the rate warning after it', (kind, table, id, warning) => {
		expect(paragraphs(table.get(id)?.[Spec.SpecUnknown])).toEqual([`gear_tab.item_notices.missing_effect.${kind}_listed`, warning]);
	});

	it('warns that an enchant rate needs validating on an enchant with no other notice', () => {
		expect(paragraphs(ENCHANT_NOTICES.get(9)?.[Spec.SpecUnknown])).toEqual([VANILLA_CHANCE_WARNING]);
	});
});

describe('registerSetBonusNotices', () => {
	const SET_ID = 9999;
	const ITEM_IDS = [90001, 90002];

	afterEach(() => {
		ITEM_IDS.forEach(id => ITEM_NOTICES.delete(id));
	});

	it('names the piece count and the rate warning of each unvalidated bonus on every item of the set', () => {
		registerSetBonusNotices({ getItemIdsForSet: (setId: number) => (setId === SET_ID ? ITEM_IDS : []) } as unknown as Database);

		for (const id of ITEM_IDS) {
			const { heading, items } = notice(ITEM_NOTICES.get(id)?.[Spec.SpecUnknown]);
			expect(heading?.trim()).toBe('gear_tab.item_notices.set_bonus_warnings');
			expect(items).toEqual([
				`gear_tab.item_notices.set_bonus_piece ${TBC_CHANCE_WARNING}`,
				`gear_tab.item_notices.set_bonus_piece ${VANILLA_PPM_WARNING}`,
			]);
		}
	});

	it('writes no notice on the items of a set the list does not name', () => {
		registerSetBonusNotices({ getItemIdsForSet: (setId: number) => (setId === SET_ID + 1 ? ITEM_IDS : []) } as unknown as Database);

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
			heading: 'gear_tab.item_notices.area_stats',
			items: ['area-1: +29 stat-17, +29 stat-18'],
		});
		expect(ITEM_INFO_NOTICES.has(PLAIN)).toBe(false);
		expect(ITEM_NOTICES.has(RUNE)).toBe(false);
	});
});
