import { Spec, Stat } from '@generated/proto/common';
import type { Database } from '@sim/proto/database';
import { render } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import {
	ENCHANT_NOTICES,
	ENCHANT_RATE_NOTICES,
	ITEM_INFO_NOTICES,
	ITEM_NOTICES,
	ITEM_RATE_NOTICES,
	registerAreaStatsNotices,
	registerSetBonusNotices,
} from './item_notices';

vi.mock('@i18n/localization', () => ({
	translateAreaType: (value: number) => `area-${value}`,
	translateStat: (value: number) => `stat-${value}`,
}));

// The real translator returns the key when no bundle is loaded, which would hide the rate; this one
// appends the interpolation values it was given.
vi.mock('@i18n/config', () => ({
	default: { t: (key: string, options?: { rate?: string }) => (options?.rate ? `${key} ${options.rate}` : key) },
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
		[1, { kind: 'ppm', rate: 1, origin: 'vanilla' }],
		[3, { kind: 'ppm', rate: 0.855, origin: 'vanilla' }],
		[4, { kind: 'ppm', rate: 1.5, origin: 'tbc' }],
		[5, { kind: 'chance', rate: 5, origin: 'vanilla' }],
		[6, { kind: 'chance', rate: 12.5, origin: 'tbc' }],
		[11, { kind: 'ppm', rate: 1.1, origin: 'observed' }],
		[12, { kind: 'chance', rate: 4, origin: 'observed' }],
		[13, { kind: 'ppm', origin: 'observed' }],
		[14, { kind: 'chance', origin: 'observed' }],
	]),
	UNVALIDATED_RATE_ENCHANTS: new Map([
		[2, { kind: 'ppm', rate: 2, origin: 'tbc' }],
		[9, { kind: 'chance', rate: 10, origin: 'vanilla' }],
	]),
	UNVALIDATED_RATE_SET_BONUSES: new Map([
		[
			9999,
			new Map([
				[2, { kind: 'chance', rate: 25, origin: 'tbc' }],
				[4, { kind: 'ppm', rate: 3, origin: 'vanilla' }],
			]),
		],
	]),
	PROC_RATE_ITEMS: new Map([[1, { kind: 'ppm', rate: 1.5 }]]),
	PROC_RATE_ENCHANTS: new Map([[2, { kind: 'chance', rate: 5 }]]),
	PROC_RATE_SET_BONUSES: new Map([
		[
			9998,
			new Map([
				[4, { kind: 'ppm', rate: 1 }],
				[2, { kind: 'chance', rate: 7.5 }],
			]),
		],
	]),
}));

const VANILLA_PPM_WARNING = 'gear_tab.item_notices.unvalidated_rate.ppm_vanilla_rate 1';
const TBC_PPM_WARNING = 'gear_tab.item_notices.unvalidated_rate.ppm_tbc_rate 2';
const VANILLA_CHANCE_WARNING = 'gear_tab.item_notices.unvalidated_rate.chance_vanilla_rate 10';
const TBC_CHANCE_WARNING = 'gear_tab.item_notices.unvalidated_rate.chance_tbc_rate 25';

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
		['a vanilla rate per minute, rounded to 2 decimals', 3, 'gear_tab.item_notices.unvalidated_rate.ppm_vanilla_rate 0.86'],
		['a TBC rate per minute', 4, 'gear_tab.item_notices.unvalidated_rate.ppm_tbc_rate 1.5'],
		['a vanilla chance', 5, 'gear_tab.item_notices.unvalidated_rate.chance_vanilla_rate 5'],
		['a TBC chance', 6, 'gear_tab.item_notices.unvalidated_rate.chance_tbc_rate 12.5'],
		['an observed rate per minute', 11, 'gear_tab.item_notices.unvalidated_rate.ppm_observed_rate 1.1'],
		['an observed chance', 12, 'gear_tab.item_notices.unvalidated_rate.chance_observed_rate 4'],
		['an observed rate per minute with no single value', 13, 'gear_tab.item_notices.unvalidated_rate.ppm_observed'],
		['an observed chance with no single value', 14, 'gear_tab.item_notices.unvalidated_rate.chance_observed'],
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

describe('the proc rate notices', () => {
	it('states a ppm at two decimals and a chance as a percent', () => {
		expect(paragraphs(ITEM_RATE_NOTICES.get(1))).toEqual(['gear_tab.item_notices.proc_rate 1.50']);
		expect(paragraphs(ENCHANT_RATE_NOTICES.get(2))).toEqual(['gear_tab.item_notices.proc_chance 5']);
	});
});

// A database holding only the given pieces of one set.
const setDatabase = (setID: number, itemIDs: number[]) => {
	const items = new Map(itemIDs.map(id => [id, { id, setId: setID }]));
	return { getAllItemIds: () => [...items.keys()], getItemById: (id: number) => items.get(id) } as unknown as Database;
};

describe('the proc rate notices of a set', () => {
	const SET_ID = 9998;
	const PIECES = [90020, 1];
	const ownRate = ITEM_RATE_NOTICES.get(1);
	const register = () => registerSetBonusNotices(setDatabase(SET_ID, PIECES));

	afterEach(() => {
		ITEM_RATE_NOTICES.delete(90020);
		ITEM_RATE_NOTICES.set(1, ownRate!);
	});

	it('adds each set bonus rate to every piece of the set', () => {
		register();

		expect(paragraphs(ITEM_RATE_NOTICES.get(90020))).toEqual([
			'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.proc_chance 7.5',
			'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.proc_rate 1.00',
		]);
		expect(paragraphs(ITEM_RATE_NOTICES.get(1))).toEqual([
			'gear_tab.item_notices.proc_rate 1.50',
			'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.proc_chance 7.5',
			'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.proc_rate 1.00',
		]);
	});

	it('gives the same notices when it runs twice', () => {
		register();
		const once = PIECES.map(id => paragraphs(ITEM_RATE_NOTICES.get(id)));
		register();

		expect(PIECES.map(id => paragraphs(ITEM_RATE_NOTICES.get(id)))).toEqual(once);
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
		registerSetBonusNotices(setDatabase(SET_ID, ITEM_IDS));

		for (const id of ITEM_IDS) {
			expect(setBonusPieces(ITEM_NOTICES.get(id)?.[Spec.SpecUnknown])).toEqual([
				{
					text: `gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.not_yet_implemented ${TBC_CHANCE_WARNING}`,
					tooltips: ['Grants a bonus.'],
				},
				{ text: `gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.unvalidated_rate.ppm_vanilla_rate 3`, tooltips: [] },
				{
					text: 'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.not_yet_implemented',
					tooltips: ['Grants another.', 'And a third.'],
				},
			]);
		}
	});

	it('keeps the missing item effect notice of a set piece and adds the set notice after it', () => {
		registerSetBonusNotices(setDatabase(SET_ID, [PIECE_WITH_EFFECT]));

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
		registerSetBonusNotices(setDatabase(GROVEKEEPER_ECLIPSE, ITEM_IDS));

		expect(setBonusPieces(ITEM_NOTICES.get(ITEM_IDS[0])?.[Spec.SpecUnknown])).toEqual([
			{
				text: 'gear_tab.item_notices.set_bonus_piece gear_tab.item_notices.not_fully_supported gear_tab.item_notices.partial_set_bonus.grovekeeper_eclipse_4',
				tooltips: [],
			},
		]);
	});

	it('writes no notice on the items of a set the list does not name', () => {
		registerSetBonusNotices(setDatabase(SET_ID + 1, ITEM_IDS));

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
			items: ['area-1: +2.9% stat-17, +2.07% stat-18'],
		});
		expect(ITEM_INFO_NOTICES.has(PLAIN)).toBe(false);
		expect(ITEM_NOTICES.has(RUNE)).toBe(false);
	});

	it('writes a rating that has a percent as that percent, and any other stat as its number', () => {
		const item = {
			id: RUNE,
			scalingOptions: { 0: { stats: {}, areaStats: [{ areaType: 1, stats: { [Stat.StatHitRating]: 4, [Stat.StatAgility]: 4 } }] } },
		};
		registerAreaStatsNotices({ getAllItems: () => [item] } as unknown as Database);

		expect(notice(ITEM_INFO_NOTICES.get(RUNE)).items).toEqual([`area-1: +4 stat-${Stat.StatAgility}, +0.4% stat-${Stat.StatHitRating}`]);
	});
});
