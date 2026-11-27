import { describe, expect, it } from 'vitest';

import { renderTooltipText, tooltipFromLink, type TooltipPlayer, type TooltipSegment } from './tooltips';

const player = (values: Record<string, number>, talents: Record<number, number> = {}): TooltipPlayer => ({
	value: name => values[name] ?? 0,
	talentPoints: spellId => talents[spellId] ?? 0,
});

describe('renderTooltipText', () => {
	// Shadow Bolt's "$s1 Shadow damage": the store's range with 0.857 of shadow spell power on each end.
	const shadowBolt: Array<TooltipSegment> = [
		'causing ',
		{ abs: true, n: { o: '+', a: [253, { o: '*', a: [0.857, { v: 'SPS' }] }] }, p: 0 },
		' to ',
		{ abs: true, n: { o: '+', a: [283, { o: '*', a: [0.857, { v: 'SPS' }] }] }, p: 0 },
		' Shadow damage.',
	];

	it('works the player’s stats into the numbers', () => {
		expect(renderTooltipText(shadowBolt, player({ SPS: 0 }))).toBe('causing 253 to 283 Shadow damage.');
		expect(renderTooltipText(shadowBolt, player({ SPS: 100 }))).toBe('causing 339 to 369 Shadow damage.');
	});

	// Revenge's talent multiplier: 1.6 with Improved Revenge (12797) at its first rank, 1 without.
	const revenge: Array<TooltipSegment> = [{ n: { o: '*', a: [{ o: 'cond', a: [{ t: 12797 }, 1.6, 1] }, 10] }, p: 0 }, ' damage'];

	it('reads a talent condition off the player’s points', () => {
		expect(renderTooltipText(revenge, player({}))).toBe('10 damage');
		expect(renderTooltipText(revenge, player({}, { 12797: 1 }))).toBe('16 damage');
	});

	it('takes the branch a talent rank picks, and closes up the space an empty branch leaves', () => {
		const text: Array<TooltipSegment> = ['A ', { if: { t: 5, r: 2 }, y: ['big'], n: [] }, ' hit.'];
		expect(renderTooltipText(text, player({}, { 5: 1 }))).toBe('A hit.');
		expect(renderTooltipText(text, player({}, { 5: 2 }))).toBe('A big hit.');
	});
});

describe('tooltipFromLink', () => {
	it('reads the spell or item a link to its page names, and the item’s suffix', () => {
		expect(tooltipFromLink('https://wowhead.com/forever/spell=1317257?rank=1')).toEqual({ spellId: 1317257, itemId: undefined, randomSuffix: undefined });
		expect(tooltipFromLink('https://wowhead.com/forever/de/item=15063?level=60&bonus=1201')).toEqual({
			spellId: undefined,
			itemId: 15063,
			randomSuffix: 1201,
		});
		expect(tooltipFromLink('https://www.wowhead.com/forever/gear-planner')).toEqual({ spellId: undefined, itemId: undefined, randomSuffix: undefined });
	});
});
