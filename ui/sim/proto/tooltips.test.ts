import { DoneMultipliers } from '@generated/proto/api';
import { Class, SpellSchool } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import {
	renderTooltipHeader,
	renderTooltipText,
	type TooltipExpr,
	tooltipFromLink,
	type TooltipLineSide,
	type TooltipPlayer,
	tooltipPlayer,
	type TooltipPlayerState,
	type TooltipSegment,
} from './tooltips';

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

	it('reads a branch side written as null as empty', () => {
		const text: Array<TooltipSegment> = ['Wards ', { if: { t: 5 }, y: ['and absorbs '], n: null }, 'fire damage.'];
		expect(renderTooltipText(text, player({}))).toBe('Wards fire damage.');
		expect(renderTooltipText(text, player({}, { 5: 1 }))).toBe('Wards and absorbs fire damage.');
	});

	// Cryoblast's "$?$pl>=25[$s1][${42+($pl-16)*8} to ${57+($pl-16)*8}]": a range that grows with the level below 25.
	const belowLevel = (base: number) => ({ n: { o: '+', a: [base, { o: '*', a: [{ o: '-', a: [{ v: 'PL' }, 16] }, 8] }] }, p: 0 });
	const cryoblast: Array<TooltipSegment> = [
		'Inflicts ',
		{ if: { o: '>=', a: [{ v: 'PL' }, 25] }, y: ['115 to 129'], n: [belowLevel(42), ' to ', belowLevel(57)] },
		' Frost damage.',
	];

	it('picks the branch the player’s level meets', () => {
		expect(renderTooltipText(cryoblast, player({ PL: 20 }))).toBe('Inflicts 74 to 89 Frost damage.');
		expect(renderTooltipText(cryoblast, player({ PL: 60 }))).toBe('Inflicts 115 to 129 Frost damage.');
	});

	it('writes the size of a value the game writes unsigned, $abs', () => {
		const text: Array<TooltipSegment> = [{ n: { o: 'abs', a: [{ o: '-', a: [10, { v: 'AP' }] }] }, p: 0 }, '%'];
		expect(renderTooltipText(text, player({ AP: 25 }))).toBe('15%');
	});
});

describe('the player’s talents and done multipliers', () => {
	const schools = Object.keys(SpellSchool).filter(key => Number.isNaN(Number(key))).length;
	const everySchool = (factor: number) => DoneMultipliers.create({ damage: Array(schools).fill(factor), healing: 1 });
	const withPoints = (points: Record<number, number>, done?: DoneMultipliers) =>
		tooltipPlayer(undefined, undefined, new Map(Object.entries(points).map(([id, n]) => [Number(id), n])), { done });

	// Heart of Wyrmthalak's proc, Flame Lash 27656: a fire damage amount with no stat shares.
	const flameLash: Array<TooltipSegment> = [
		'Melee and Ranged attacks have a chance to deal ',
		{ abs: true, n: { fx: { b: 112, c: 4, k: 'dmg' } }, p: 0 },
		' to ',
		{ abs: true, n: { fx: { b: 168, c: 4, k: 'dmg' } }, p: 0 },
		' Fire damage.',
	];

	it('multiplies a damage amount by the sim’s done multiplier, Lone Wolf’s 20% with no pet out', () => {
		expect(renderTooltipText(flameLash, withPoints({}))).toBe('Melee and Ranged attacks have a chance to deal 112 to 168 Fire damage.');
		expect(renderTooltipText(flameLash, withPoints({}, everySchool(1)))).toBe('Melee and Ranged attacks have a chance to deal 112 to 168 Fire damage.');
		expect(renderTooltipText(flameLash, withPoints({}, everySchool(1.2)))).toBe('Melee and Ranged attacks have a chance to deal 134 to 202 Fire damage.');
	});

	it('takes the multiplier of the amount’s own school, and healing’s for a heal', () => {
		const fireOnly = everySchool(1);
		fireOnly.damage[SpellSchool.SpellSchoolFire] = 1.5;
		fireOnly.healing = 2;
		const frost: Array<TooltipSegment> = [{ n: { fx: { b: 100, c: 16, k: 'dmg' } }, p: 0 }];
		const heal: Array<TooltipSegment> = [{ n: { fx: { b: 100, k: 'hot' } }, p: 0 }];
		expect(renderTooltipText(flameLash, withPoints({}, fireOnly))).toBe('Melee and Ranged attacks have a chance to deal 168 to 252 Fire damage.');
		expect(renderTooltipText(frost, withPoints({}, fireOnly))).toBe('100');
		expect(renderTooltipText(heal, withPoints({}, fireOnly))).toBe('200');
	});

	it('adds flat modifiers to the base points before percentages, then the shares, then the whole amount’s modifiers', () => {
		const amount: Array<TooltipSegment> = [
			{
				n: {
					fx: {
						b: 100,
						s: [{ o: '*', a: [0.5, { v: 'SPFI' }] }],
						k: 'dmg',
						c: 4,
						m: [
							{ t: 1, v: [10], f: true },
							{ t: 2, v: [10] },
							{ t: 3, v: [10, 20], w: true },
						],
					},
				},
				p: 0,
			},
		];
		const player = (points: Record<number, number>): TooltipPlayer => ({ ...withPoints(points), value: name => (name === 'SPFI' ? 100 : 0) });
		expect(renderTooltipText(amount, player({}))).toBe('150');
		expect(renderTooltipText(amount, player({ 1: 1, 2: 1, 3: 2 }))).toBe('205');
	});

	// Volley 14295's damage with Barrage 19461, a Marksmanship talent stating SPELLMOD_DAMAGE.
	const volley: Array<TooltipSegment> = [
		'causing ',
		{ abs: true, n: { fx: { b: 112, c: 64, k: 'dmg', m: [{ t: 19461, v: [5, 10, 15], w: true }] } }, p: 0 },
		' Arcane damage',
	];
	// Cone of Cold 10161's damage with Improved Cone of Cold 11190.
	const coneOfCold: Array<TooltipSegment> = [
		{ n: { fx: { b: 328, c: 16, k: 'dmg', m: [{ t: 11190, v: [12, 23, 35], w: true }], s: [{ o: '*', a: [0.129, { v: 'SPFR' }] }] } }, p: 0 },
		' Frost damage',
	];

	it('works a damage talent’s points into its spells’ amounts', () => {
		expect(renderTooltipText(volley, withPoints({}))).toBe('causing 112 Arcane damage');
		expect(renderTooltipText(volley, withPoints({ 19461: 3 }))).toBe('causing 129 Arcane damage');
		expect(renderTooltipText(volley, withPoints({ 19461: 3 }, everySchool(1.2)))).toBe('causing 155 Arcane damage');
		expect(renderTooltipText(coneOfCold, withPoints({}))).toBe('328 Frost damage');
		expect(renderTooltipText(coneOfCold, withPoints({ 11190: 2 }))).toBe('403 Frost damage');
	});

	// Auto Shot 75's range with Hawk Eye 19498's flat yards, and Frostbolt 116's cast time with Improved Frostbolt
	// 11070's flat milliseconds, as the generator writes them.
	const autoShot: Array<[TooltipLineSide, TooltipLineSide]> = [
		['', ['8 - ', { n: { fx: { b: 35, m: [{ t: 19498, v: [2, 4, 6], f: true }] } }, p: 2, t: true }, ' yd range']],
		['Instant', ''],
		['Requires Ranged Weapon', ''],
	];
	const frostbolt: Array<[TooltipLineSide, TooltipLineSide]> = [
		[[{ n: { o: '/', a: [{ fx: { b: 1500, m: [{ t: 11070, v: [-100, -200, -300, -400, -500], f: true }] } }, 1000] }, p: 1, t: true }, ' sec cast'], ''],
	];

	it('works a talent’s points into the header’s range and cast time', () => {
		expect(renderTooltipHeader(autoShot, withPoints({}))).toEqual([
			['', '8 - 35 yd range'],
			['Instant', ''],
			['Requires Ranged Weapon', ''],
		]);
		expect(renderTooltipHeader(autoShot, withPoints({ 19498: 3 }))[0]).toEqual(['', '8 - 41 yd range']);
		expect(renderTooltipHeader(frostbolt, withPoints({}))).toEqual([['1.5 sec cast', '']]);
		expect(renderTooltipHeader(frostbolt, withPoints({ 11070: 5 }))).toEqual([['1 sec cast', '']]);
	});

	// Shank 1329283's 10 sec cooldown less 0.4 sec a point of 1329284, which crosses 10 and so picks its decimals on
	// the value, and Vanish 1856's 5 min less 45 sec a point of Elusiveness 13981, whose 4.25 min reads as the
	// generator's own text, as the generator writes them.
	const cooldownMs = (b: number, talent: number, values: Array<number>): TooltipExpr => ({
		o: 'round',
		a: [{ fx: { b, m: [{ t: talent, v: values, f: true }] } }, 0],
	});
	const shankCooldown = cooldownMs(10000, 1329284, [-400, -800, -1200, -1600, -2000]);
	const shankSeconds: TooltipExpr = { o: '/', a: [shankCooldown, 1000] };
	const shank: Array<[TooltipLineSide, TooltipLineSide]> = [
		['5 Energy', 'Melee Range'],
		['Instant', [{ n: { o: 'round', a: [shankSeconds, { o: 'cond', a: [{ o: '>=', a: [shankSeconds, 10] }, 0, 1] }] }, p: 1, t: true }, ' sec cooldown']],
	];
	const vanishCooldown = cooldownMs(300000, 13981, [-45000, -90000]);
	const vanish: Array<[TooltipLineSide, TooltipLineSide]> = [
		[
			'Instant',
			[
				{
					if: { o: '=', a: [vanishCooldown, 255000] },
					y: ['4.2 min cooldown'],
					n: [{ n: { o: '/', a: [vanishCooldown, 60000] }, p: 1, t: true }, ' min cooldown'],
				},
			],
		],
	];

	it('writes a cooldown the talents shorten in two significant digits', () => {
		expect(renderTooltipHeader(shank, withPoints({}))[1]).toEqual(['Instant', '10 sec cooldown']);
		expect(renderTooltipHeader(shank, withPoints({ 1329284: 1 }))[1]).toEqual(['Instant', '9.6 sec cooldown']);
		expect(renderTooltipHeader(shank, withPoints({ 1329284: 5 }))[1]).toEqual(['Instant', '8 sec cooldown']);
		expect(renderTooltipHeader(vanish, withPoints({}))).toEqual([['Instant', '5 min cooldown']]);
		expect(renderTooltipHeader(vanish, withPoints({ 13981: 1 }))).toEqual([['Instant', '4.2 min cooldown']]);
		expect(renderTooltipHeader(vanish, withPoints({ 13981: 2 }))).toEqual([['Instant', '3.5 min cooldown']]);
	});

	it('reads the player’s class and starting form for a condition on a known spell or an aura', () => {
		// Revenge's attack power share in Defensive Stance, form 18, and a text on a spell every priest knows.
		const revenge: Array<TooltipSegment> = [{ n: { o: '*', a: [{ o: 'cond', a: [{ o: '=', a: [{ v: 'FORM' }, 18] }, 0.225, 0.25] }, 1000] }, p: 0 }];
		const priestly: Array<TooltipSegment> = [{ if: { o: '=', a: [{ v: 'CLASS' }, Class.ClassPriest] }, y: ['Healed.'], n: ['Nothing.'] }];
		const as = (state: TooltipPlayerState) => tooltipPlayer(undefined, undefined, new Map(), state);
		expect(renderTooltipText(revenge, as({}))).toBe('250');
		expect(renderTooltipText(revenge, as({ playerClass: Class.ClassWarrior, form: 18 }))).toBe('225');
		expect(renderTooltipText(priestly, as({ playerClass: Class.ClassWarrior }))).toBe('Nothing.');
		expect(renderTooltipText(priestly, as({ playerClass: Class.ClassPriest }))).toBe('Healed.');
	});

	it('reads an empty branch as no text and drops the space it leaves at the end', () => {
		const text: Array<TooltipSegment> = ['Absorbs damage. ', { if: { t: 5 }, y: ['Reflects spells.'], n: null }];
		expect(renderTooltipText(text, withPoints({}))).toBe('Absorbs damage.');
		expect(renderTooltipText(text, withPoints({ 5: 1 }))).toBe('Absorbs damage. Reflects spells.');
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
