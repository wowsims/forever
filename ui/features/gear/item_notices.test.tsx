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
	MISSING_ITEM_EFFECTS: new Map([
		[1, ['Does a thing.']],
		[90003, ['Procs a thing.']],
	]),
	MISSING_ENCHANT_EFFECTS: new Map([[2, ['Enchants a thing.']]]),
	MISSING_SET_BONUSES: new Map([
		[
			9999,
			new Map([
				[2, ['Grants a bonus.']],
				[6, ['Grants another.', 'And a third.']],
			]),
		],
	]),
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
	const PIECE_WITH_EFFECT = 90003;
	const effectNotice = ITEM_NOTICES.get(PIECE_WITH_EFFECT);

	afterEach(() => {
		ITEM_IDS.forEach(id => ITEM_NOTICES.delete(id));
		ITEM_NOTICES.set(PIECE_WITH_EFFECT, effectNotice!);
	});

	// Each piece count's line with the tooltip lists nested under it taken out, and those tooltips.
	const setBonusPieces = (node: ReactNode) => {
		const { container } = render(<>{node}</>);
		const heading = [...container.querySelectorAll('p')].find(p => p.textContent?.trim() === 'gear_tab.item_notices.set_bonus_warnings');
		return [...(heading?.nextElementSibling?.children ?? [])].map(piece => ({
			text: [...piece.childNodes]
				.filter(child => child.nodeName !== 'UL')
				.map(child => child.textContent)
				.join('')
				.trim(),
			tooltips: [...piece.querySelectorAll('li')].map(tooltip => tooltip.textContent),
		}));
	};

	it('names each piece count with its missing effects and rate warnings on every item of the set', () => {
		registerSetBonusNotices({ getItemIdsForSet: (setId: number) => (setId === SET_ID ? ITEM_IDS : []) } as unknown as Database);

		for (const id of ITEM_IDS) {
			expect(setBonusPieces(ITEM_NOTICES.get(id)?.[Spec.SpecUnknown])).toEqual([
				{
					text: `gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.not_yet_implemented ${TBC_CHANCE_WARNING}`,
					tooltips: ['Grants a bonus.'],
				},
				{ text: `gear_tab.item_notices.set_bonus_piece ${VANILLA_PPM_WARNING}`, tooltips: [] },
				{
					text: 'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.not_yet_implemented',
					tooltips: ['Grants another.', 'And a third.'],
				},
			]);
		}
	});

	it('keeps the missing item effect notice of a set piece and adds the set notice after it', () => {
		registerSetBonusNotices({ getItemIdsForSet: (setId: number) => (setId === SET_ID ? [PIECE_WITH_EFFECT] : []) } as unknown as Database);

		const node = ITEM_NOTICES.get(PIECE_WITH_EFFECT)?.[Spec.SpecUnknown];
		expect(paragraphs(node).map(text => text?.trim())).toEqual([
			'gear_tab.item_notices.missing_effect.item_listed',
			'gear_tab.item_notices.set_bonus_warnings',
		]);
		expect(notice(node).items[0]).toBe('Procs a thing.');
		expect(setBonusPieces(node)).toHaveLength(3);
	});

	it('names what a partly implemented bonus leaves out', () => {
		const GROVEKEEPER_ECLIPSE = 2114;
		registerSetBonusNotices({ getItemIdsForSet: (setId: number) => (setId === GROVEKEEPER_ECLIPSE ? ITEM_IDS : []) } as unknown as Database);

		expect(setBonusPieces(ITEM_NOTICES.get(ITEM_IDS[0])?.[Spec.SpecUnknown])).toEqual([
			{
				text: 'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.not_fully_supported gear_tab.item_notices.partial_set_bonus.grovekeeper_eclipse_4',
				tooltips: [],
			},
		]);
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
