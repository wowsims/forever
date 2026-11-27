import type { DoneMultipliers } from '@generated/proto/api';
import { type Class, ItemSlot, SpellSchool, Stat } from '@generated/proto/common';

import { CHARACTER_LEVEL } from '../constants/mechanics';
import type { Gear } from './gear';
import type { Stats } from './stats';

// The spell tooltips gen_db writes from the game files (tools/database/tooltips.go), with the numbers that grow
// with the player left as expressions over the player's stats and talents.

/** A number: a constant, a player value named as the game's tooltip tokens name it, a talent rank, an operation or a value the player's modifiers change. */
export type TooltipExpr = number | { v: string } | { t: number; r?: number } | { o: string; a: Array<TooltipExpr> } | { fx: TooltipModifiedValue };

/** A damage or healing amount's kind, direct or periodic. */
export type TooltipAmountKind = 'dmg' | 'dot' | 'heal' | 'hot';

/**
 * A value the player's talents can change, and the done multipliers too where it is damage or healing: the base
 * points, the attack power and spell power shares added after them, the kind, the school mask of damage and the
 * talents' modifiers.
 */
export interface TooltipModifiedValue {
	b: TooltipExpr;
	s?: Array<TooltipExpr>;
	k?: TooltipAmountKind;
	c?: number;
	m?: Array<TooltipModifier>;
}

/**
 * A talent's spell modifier: its value at each rank from rank 1, added when flat (f) and a percentage otherwise. A
 * whole one (w) applies to the damage or healing amount with its shares, any other to the base points.
 */
export interface TooltipModifier {
	t: number;
	v: Array<number>;
	f?: boolean;
	w?: boolean;
}

/** A number with p decimal places, or at most p with trailing zeros dropped (t); a branch, whose empty side is null. */
export type TooltipSegment =
	| string
	| { n: TooltipExpr; p: number; abs?: boolean; t?: boolean }
	| { if: TooltipExpr; y: Array<TooltipSegment> | null; n: Array<TooltipSegment> | null };

/** One side of a header line: its text, or the segments of one whose number the player's talents can change. */
export type TooltipLineSide = string | Array<TooltipSegment>;

export interface SpellTooltipData {
	/** Name. */
	n: string;
	/** The rank or kind under the name, such as "Rank 5". */
	r?: string;
	/** Icon name. */
	i?: string;
	/** The lines above the description as [left, right]: cost and range, cast time and cooldown, then requirements, tools and reagents. */
	h?: Array<[TooltipLineSide, TooltipLineSide]>;
	/** Description. */
	d?: Array<TooltipSegment>;
	/** The buff's text. */
	a?: Array<TooltipSegment>;
	/** A talent's points. */
	m?: number;
	/** A talent's description at each rank from rank 1, where the ranks read differently. */
	x?: Array<Array<TooltipSegment>>;
}

/** A requirement an item states: a profession and the skill in it, or a faction and the standing (0 Hated to 7 Exalted). */
export interface ItemRequirementData {
	n: string;
	r?: number;
}

/** What an item's tooltip takes from the game files beyond the item database. */
export interface ItemTooltipData {
	/** Binding: 1 when picked up, 2 when equipped, 3 when used. */
	b?: number;
	/** Flavour text. */
	f?: string;
	/**
	 * Spells as [trigger, spell id, cooldown in ms, charges]: trigger 0 is a use, 1 an equip effect, 2 a chance on hit.
	 * The cooldown is 0 and the charges absent where the game writes none.
	 */
	e?: Array<[number, number, number, number?]>;
	/** Sell price in copper. */
	s?: number;
	/** The most a player may carry ("Unique" at 1, "Unique (N)" above), or -1 for an item only one of which may be worn. */
	u?: number;
	/** The profession skill and the reputation the item asks for. */
	k?: ItemRequirementData;
	g?: ItemRequirementData;
	/** For an item the item database does not hold, a consumable or an enchant's item: name, quality, icon, item level and required level. */
	n?: string;
	q?: number;
	i?: string;
	l?: number;
	r?: number;
	/** The inventory type, such as 1 for the head or 24 for ammo. */
	o?: number;
	/** The subclass the game names beside the slot or a bag's slots, such as Arrow or Quiver; a bag's or quiver's slots; the damage per second an arrow adds. */
	t?: string;
	c?: number;
	d?: number;
	/** The classes that may use it; none for every class. */
	a?: Array<Class>;
}

export interface SetTooltipData {
	n: string;
	/** The pieces' item ids. */
	p: Array<number>;
	/** Bonuses as [pieces worn, spell id]. */
	b: Array<[number, number]>;
}

interface TooltipFile {
	spells: Record<string, SpellTooltipData>;
	items: Record<string, ItemTooltipData>;
	sets: Record<string, SetTooltipData>;
	/** The name the game writes on an enchanted item, by the enchant's effect id. */
	enchants: Record<string, string>;
}

const TOOLTIPS_URL = '/forever/assets/database/tooltips.json';
// Every other item and named spell in the game files, which the site loads the first time it shows one.
const EXTRA_URL = '/forever/assets/database/tooltips_extra.json';
const EMPTY: TooltipFile = { spells: {}, items: {}, sets: {}, enchants: {} };

const fetchFile = (url: string): Promise<TooltipFile> =>
	fetch(url)
		.then(response => (response.ok ? response.json() : Promise.reject(new Error(`HTTP ${response.status}`))))
		// A file missing a section, such as one written before the section existed, reads that section as empty.
		.then((data: Partial<TooltipFile>) => ({ spells: data.spells ?? {}, items: data.items ?? {}, sets: data.sets ?? {}, enchants: data.enchants ?? {} }));

let loaded: TooltipFile = EMPTY;
let extra: TooltipFile = EMPTY;
let loading: Promise<TooltipFile> | null = null;
let loadingExtra: Promise<TooltipFile> | null = null;
let version = 0;
const listeners = new Set<() => void>();

const settled = () => {
	version++;
	listeners.forEach(listener => listener());
};

/** Loads the tooltips once; a failed load is tried again on the next call and reads as no tooltips meanwhile. */
export const loadTooltips = (): Promise<TooltipFile> =>
	(loading ??= fetchFile(TOOLTIPS_URL)
		.then(data => {
			loaded = data;
			settled();
			return data;
		})
		.catch(error => {
			console.warn(`Tooltips failed to load from ${TOOLTIPS_URL}:`, error);
			loading = null;
			return EMPTY;
		}));

const loadExtraTooltips = (): Promise<TooltipFile> =>
	(loadingExtra ??= fetchFile(EXTRA_URL)
		.then(data => {
			extra = data;
			settled();
			return data;
		})
		.catch(error => {
			console.warn(`Tooltips failed to load from ${EXTRA_URL}:`, error);
			loadingExtra = null;
			return EMPTY;
		}));

/** For a render that reads the loaded tooltips: a change each time a file arrives. */
export const subscribeTooltips = (listener: () => void) => {
	listeners.add(listener);
	return () => listeners.delete(listener);
};
export const tooltipsVersion = () => version;

/** Tooltips already loaded, for a render that cannot wait. An id neither file holds yet starts loading the second file. */
const lookUp = <T>(section: (file: TooltipFile) => Record<string, T>, id: number): T | undefined => {
	const own = section(loaded)[id] ?? section(extra)[id];
	if (!own && id && loading) void loadExtraTooltips();
	return own;
};
export const loadedSpellTooltip = (spellId: number): SpellTooltipData | undefined => lookUp(file => file.spells, spellId);
export const loadedItemTooltip = (itemId: number): ItemTooltipData | undefined => lookUp(file => file.items, itemId);
export const loadedSetTooltip = (setId: number): SetTooltipData | undefined => loaded.sets[setId];
export const loadedEnchantName = (effectId: number): string | undefined => loaded.enchants[effectId];

export const hasLoadedSpellTooltip = (spellId: number): boolean => !!loaded.spells[spellId] || !!extra.spells[spellId];

// The first file answers what the site's pages show; an id it lacks waits for the second.
const holds = async <T>(section: (file: TooltipFile) => Record<string, T>, id: number): Promise<T | undefined> =>
	section(await loadTooltips())[id] ?? section(await loadExtraTooltips())[id];

export const hasSpellTooltip = async (spellId: number): Promise<boolean> => !!(await holds(file => file.spells, spellId));

export const hasItemTooltip = async (itemId: number): Promise<boolean> => !!(await holds(file => file.items, itemId));

/** The tooltip host every spell and item anchor opts into (ui/features/tooltips). */
export const SPELL_TOOLTIP_ID = 'spell-tooltip';

const WOWHEAD_LINK = 'a[href*="wowhead.com/forever/"]';

/**
 * What the host opens on, as wowhead's script did: an anchor marked for it, and any link to a spell or item page that no
 * other tooltip claims. Every icon links to its page, so a view needs nothing more to show the tooltip.
 */
export const SPELL_TOOLTIP_SELECTOR = [
	`[data-tooltip-id="${SPELL_TOOLTIP_ID}"]`,
	`${WOWHEAD_LINK}[href*="spell="]:not([data-tooltip-id])`,
	`${WOWHEAD_LINK}[href*="item="]:not([data-tooltip-id])`,
].join(', ');

/** The spell or item a link to its page names, with the random suffix an item link carries. */
export const tooltipFromLink = (href: string): { spellId?: number; itemId?: number; randomSuffix?: number } => {
	const id = (key: string) => Number(new RegExp(`(?:^|[/?&])${key}=(\\d+)`).exec(href)?.[1]) || undefined;
	return { spellId: id('spell'), itemId: id('item'), randomSuffix: id('bonus') };
};

/** Props for an anchor whose tooltip is the spell's own, which also keep wowhead's script off it. A talent anchor shows the talent's ranks; the spell a talent teaches shows as a spell elsewhere. */
export const spellTooltipAnchor = (spellId: number, options: { buff?: boolean; talent?: boolean } = {}): Record<string, string> => ({
	'data-tooltip-id': SPELL_TOOLTIP_ID,
	'data-spell-tooltip': String(spellId),
	...(options.buff ? { 'data-spell-buff': 'true' } : {}),
	...(options.talent ? { 'data-spell-talent': 'true' } : {}),
	'data-disable-wowhead-tooltip': 'true',
});

/** Props for an anchor whose tooltip is an enchant's: the enchant by its effect id, and the spell or item that tells rows sharing an effect id apart. */
export const enchantTooltipAnchor = (enchant: { effectId: number; spellId: number; itemId: number }): Record<string, string> => ({
	'data-tooltip-id': SPELL_TOOLTIP_ID,
	'data-enchant-tooltip': String(enchant.effectId),
	...(enchant.spellId ? { 'data-enchant-spell': String(enchant.spellId) } : {}),
	...(enchant.itemId ? { 'data-enchant-item': String(enchant.itemId) } : {}),
	'data-disable-wowhead-tooltip': 'true',
});

/** A spell's or an item's name, icon and rank from the tooltips, the same fields wowhead's tooltip lookup answered. */
export const tooltipIconData = async (
	kind: 'spell' | 'item',
	id: number,
): Promise<{ name: string; icon: string; rank: number; hasBuff: boolean } | undefined> => {
	if (kind === 'item') {
		const item = await holds(file => file.items, id);
		return item?.n && item.i ? { name: item.n, icon: item.i, rank: 0, hasBuff: false } : undefined;
	}
	const spell = await holds(file => file.spells, id);
	if (!spell?.i) return undefined;
	return { name: spell.n, icon: spell.i, rank: Number(/(\d+)/.exec(spell.r ?? '')?.[1] ?? 0), hasBuff: !!spell.a?.length };
};

/** Props for an anchor whose tooltip is the item's own: the item, and the random suffix and enchant it carries. */
export const itemTooltipAnchor = (itemId: number, options: { randomSuffix?: number; enchant?: number } = {}): Record<string, string> => ({
	'data-tooltip-id': SPELL_TOOLTIP_ID,
	'data-item-tooltip': String(itemId),
	...(options.randomSuffix ? { 'data-item-suffix': String(options.randomSuffix) } : {}),
	...(options.enchant ? { 'data-item-enchant': String(options.enchant) } : {}),
	'data-disable-wowhead-tooltip': 'true',
});

/** What a tooltip's expressions read off the player. */
export interface TooltipPlayer {
	value: (name: string) => number;
	/** The points spent in the talent whose rank 1 spell this is. */
	talentPoints: (spellId: number) => number;
	/** The factor on the player's damage of a school mask, or on its healing; 1 where absent. */
	doneMultiplier?: (kind: TooltipAmountKind, school: number) => number;
}

const truth = (value: boolean) => (value ? 1 : 0);

const OPERATIONS: Record<string, (args: Array<number>) => number> = {
	'+': ([a, b]) => a + b,
	'-': ([a, b]) => a - b,
	'*': ([a, b]) => a * b,
	'/': ([a, b]) => a / b,
	'^': ([a, b]) => a ** b,
	neg: ([a]) => -a,
	max: ([a, b]) => Math.max(a, b),
	min: ([a, b]) => Math.min(a, b),
	abs: ([a]) => Math.abs(a),
	floor: ([a]) => Math.floor(a),
	clamp: ([a, low, high]) => Math.min(Math.max(a, low), high),
	round: ([a, places]) => Math.round(a * 10 ** places) / 10 ** places,
	cond: ([test, a, b]) => (test > 0 ? a : b),
	'>': ([a, b]) => truth(a > b),
	'>=': ([a, b]) => truth(a >= b),
	'<': ([a, b]) => truth(a < b),
	'=': ([a, b]) => truth(a === b),
	'!=': ([a, b]) => truth(a !== b),
	and: ([a, b]) => truth(a !== 0 && b !== 0),
	or: ([a, b]) => truth(a !== 0 || b !== 0),
	not: ([a]) => truth(a === 0),
};

export const evaluateTooltipExpr = (expr: TooltipExpr, player: TooltipPlayer): number => {
	if (typeof expr === 'number') return expr;
	if ('v' in expr) return player.value(expr.v);
	if ('t' in expr) return truth(player.talentPoints(expr.t) >= (expr.r ?? 1));
	if ('fx' in expr) return evaluateModifiedValue(expr.fx, player);
	return OPERATIONS[expr.o]?.(expr.a.map(arg => evaluateTooltipExpr(arg, player))) ?? 0;
};

/**
 * The base points with the flat modifiers added and the percentages multiplied in, then the shares, then the whole
 * amount's modifiers and the done multiplier: with no point in any of the talents and multipliers of 1, the plain sum.
 */
const evaluateModifiedValue = (value: TooltipModifiedValue, player: TooltipPlayer): number => {
	let flat = 0;
	let percent = 1;
	let wholeFlat = 0;
	let wholePercent = 1;
	for (const mod of value.m ?? []) {
		const points = player.talentPoints(mod.t);
		if (points <= 0 || !mod.v.length) continue;
		const amount = mod.v[Math.min(points, mod.v.length) - 1];
		if (mod.w) {
			if (mod.f) wholeFlat += amount;
			else wholePercent *= 1 + amount / 100;
		} else if (mod.f) flat += amount;
		else percent *= 1 + amount / 100;
	}
	let total = (evaluateTooltipExpr(value.b, player) + flat) * percent;
	for (const share of value.s ?? []) total += evaluateTooltipExpr(share, player);
	total = (total + wholeFlat) * wholePercent;
	return value.k ? total * (player.doneMultiplier?.(value.k, value.c ?? 0) ?? 1) : total;
};

const renderSegments = (segments: Array<TooltipSegment>, player: TooltipPlayer): string =>
	segments
		.map(segment => {
			if (typeof segment === 'string') return segment;
			if ('if' in segment) return renderSegments((evaluateTooltipExpr(segment.if, player) ? segment.y : segment.n) ?? [], player);
			const value = evaluateTooltipExpr(segment.n, player);
			const fixed = (segment.abs ? Math.abs(value) : value).toFixed(segment.p);
			return segment.t ? String(Number(fixed)) : fixed;
		})
		.join('');

/**
 * The text of a tooltip for the player. An empty branch can leave a doubled space, which is closed up, or a space or
 * line break at either end, which is trimmed as the generator trims its own.
 */
export const renderTooltipText = (segments: Array<TooltipSegment>, player: TooltipPlayer): string =>
	renderSegments(segments, player)
		.replace(/ {2,}/g, ' ')
		.replace(/^[ \n]+|[ \n]+$/g, '');

/** The header lines for the player, leaving out a line with nothing on either side. */
export const renderTooltipHeader = (lines: Array<[TooltipLineSide, TooltipLineSide]>, player: TooltipPlayer): Array<[string, string]> =>
	lines
		.map(sides => sides.map(side => (typeof side === 'string' ? side : renderTooltipText(side, player))) as [string, string])
		.filter(([left, right]) => left || right);

const SCHOOL_POWER: Record<string, Stat> = {
	SPA: Stat.StatArcaneDamage,
	SPFI: Stat.StatFireDamage,
	SPFR: Stat.StatFrostDamage,
	SPH: Stat.StatHolyDamage,
	SPN: Stat.StatNatureDamage,
	SPS: Stat.StatShadowDamage,
};

const PLAIN_STATS: Record<string, Stat> = {
	AP: Stat.StatAttackPower,
	RAP: Stat.StatRangedAttackPower,
	SP: Stat.StatSpellDamage,
	HEAL: Stat.StatHealingPower,
	SPI: Stat.StatSpirit,
	STR: Stat.StatStrength,
	AGI: Stat.StatAgility,
	INT: Stat.StatIntellect,
	STA: Stat.StatStamina,
};

// The school whose multiplier a spell of several schools takes, as the sim picks it: the first of these its mask names.
const MULTIPLIER_SCHOOLS: Array<[number, SpellSchool]> = [
	[1, SpellSchool.SpellSchoolPhysical],
	[64, SpellSchool.SpellSchoolArcane],
	[4, SpellSchool.SpellSchoolFire],
	[16, SpellSchool.SpellSchoolFrost],
	[2, SpellSchool.SpellSchoolHoly],
	[8, SpellSchool.SpellSchoolNature],
	[32, SpellSchool.SpellSchoolShadow],
];

const multiplierSchool = (mask: number): SpellSchool => MULTIPLIER_SCHOOLS.find(([bit]) => mask & bit)?.[1] ?? SpellSchool.SpellSchoolPhysical;

/** What the sim and the player's settings say beside the stats: the done multipliers, the class and the starting form. */
export interface TooltipPlayerState {
	done?: DoneMultipliers;
	playerClass?: Class;
	/** The form or stance the player starts the fight in, by its id in the game files; 0 for none. */
	form?: number;
}

/**
 * The player as a tooltip reads it: the sim's final stats, a school's spell power on top of spell damage, the level,
 * the weapons, the talents, the class and the form (CLASS and FORM), and the sim's done multipliers.
 */
export const tooltipPlayer = (
	stats: Stats | undefined,
	gear: Gear | undefined,
	talentPoints: ReadonlyMap<number, number>,
	{ done, playerClass, form }: TooltipPlayerState = {},
): TooltipPlayer => {
	const weapon = (slot: ItemSlot) => gear?.getEquippedItem(slot)?.item;
	return {
		value: name => {
			if (name === 'PL') return CHARACTER_LEVEL;
			if (name === 'CLASS') return playerClass ?? 0;
			if (name === 'FORM') return form ?? 0;
			if (name in PLAIN_STATS) return stats?.getStat(PLAIN_STATS[name]) ?? 0;
			if (name in SCHOOL_POWER) return (stats?.getStat(Stat.StatSpellDamage) ?? 0) + (stats?.getStat(SCHOOL_POWER[name]) ?? 0);
			const mainHand = weapon(ItemSlot.ItemSlotMainHand);
			const offHand = weapon(ItemSlot.ItemSlotOffHand);
			switch (name) {
				case 'mwb':
					return mainHand?.weaponDamageMin ?? 0;
				case 'MWB':
					return mainHand?.weaponDamageMax ?? 0;
				case 'mws':
					return mainHand?.weaponSpeed ?? 0;
				case 'owb':
					return offHand?.weaponDamageMin ?? 0;
				case 'OWB':
					return offHand?.weaponDamageMax ?? 0;
				case 'ows':
					return offHand?.weaponSpeed ?? 0;
			}
			return 0;
		},
		talentPoints: spellId => talentPoints.get(spellId) ?? 0,
		doneMultiplier: (kind, school) => (kind === 'heal' || kind === 'hot' ? done?.healing : done?.damage[multiplierSchool(school)]) || 1,
	};
};
