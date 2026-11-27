import { ItemSlot, Stat } from '@generated/proto/common';

import { CHARACTER_LEVEL } from '../constants/mechanics';
import type { Gear } from './gear';
import type { Stats } from './stats';

// The spell tooltips gen_db writes from the game files (tools/database/tooltips.go), with the numbers that grow
// with the player left as expressions over the player's stats and talents.

/** A number: a constant, a player value named as the game's tooltip tokens name it, a talent rank or an operation. */
export type TooltipExpr = number | { v: string } | { t: number; r?: number } | { o: string; a: Array<TooltipExpr> };

export type TooltipSegment = string | { n: TooltipExpr; p: number; abs?: boolean } | { if: TooltipExpr; y: Array<TooltipSegment>; n: Array<TooltipSegment> };

export interface SpellTooltipData {
	/** Name. */
	n: string;
	/** The rank or kind under the name, such as "Rank 5". */
	r?: string;
	/** Icon name. */
	i?: string;
	/** The cost, range, cast time and cooldown lines as [left, right]. */
	h?: Array<[string, string]>;
	/** Description. */
	d?: Array<TooltipSegment>;
	/** The buff's text. */
	a?: Array<TooltipSegment>;
	/** A talent's points. */
	m?: number;
	/** A talent's description at each rank from rank 1, where the ranks read differently. */
	x?: Array<Array<TooltipSegment>>;
}

/** What an item's tooltip takes from the game files beyond the item database. */
export interface ItemTooltipData {
	/** Binding: 1 when picked up, 2 when equipped, 3 when used. */
	b?: number;
	/** Flavour text. */
	f?: string;
	/** Spells as [trigger, spell id, cooldown in ms]: trigger 0 is a use, 1 an equip effect, 2 a chance on hit. */
	e?: Array<[number, number, number]>;
	/** For an item the item database does not hold, a consumable or an enchant's item: name, quality, icon, item level and required level. */
	n?: string;
	q?: number;
	i?: string;
	l?: number;
	r?: number;
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
}

const TOOLTIPS_URL = '/forever/assets/database/tooltips.json';
const EMPTY: TooltipFile = { spells: {}, items: {}, sets: {} };

let loaded: TooltipFile = EMPTY;
let loading: Promise<TooltipFile> | null = null;

/** Loads the tooltips once; a failed load is tried again on the next call and reads as no tooltips meanwhile. */
export const loadTooltips = (): Promise<TooltipFile> =>
	(loading ??= fetch(TOOLTIPS_URL)
		.then(response => (response.ok ? response.json() : Promise.reject(new Error(`HTTP ${response.status}`))))
		// A file missing a section, such as one written before the section existed, reads that section as empty.
		.then((data: Partial<TooltipFile>) => (loaded = { spells: data.spells ?? {}, items: data.items ?? {}, sets: data.sets ?? {} }))
		.catch(error => {
			console.warn(`Tooltips failed to load from ${TOOLTIPS_URL}:`, error);
			loading = null;
			return EMPTY;
		}));

/** Tooltips already loaded, for a render that cannot wait. */
export const loadedSpellTooltip = (spellId: number): SpellTooltipData | undefined => loaded.spells[spellId];
export const loadedItemTooltip = (itemId: number): ItemTooltipData | undefined => loaded.items[itemId];
export const loadedSetTooltip = (setId: number): SetTooltipData | undefined => loaded.sets[setId];

export const hasLoadedSpellTooltip = (spellId: number): boolean => !!loaded.spells[spellId];

export const hasSpellTooltip = async (spellId: number): Promise<boolean> => !!(await loadTooltips()).spells[spellId];

export const hasItemTooltip = async (itemId: number): Promise<boolean> => !!(await loadTooltips()).items[itemId];

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
	const file = await loadTooltips();
	if (kind === 'item') {
		const item = file.items[id];
		return item?.n && item.i ? { name: item.n, icon: item.i, rank: 0, hasBuff: false } : undefined;
	}
	const spell = file.spells[id];
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
	return OPERATIONS[expr.o]?.(expr.a.map(arg => evaluateTooltipExpr(arg, player))) ?? 0;
};

/** The text of a tooltip for the player. A branch left empty leaves a doubled space, which is closed up. */
export const renderTooltipText = (segments: Array<TooltipSegment>, player: TooltipPlayer): string =>
	segments
		.map(segment => {
			if (typeof segment === 'string') return segment;
			if ('if' in segment) return renderTooltipText(evaluateTooltipExpr(segment.if, player) ? segment.y : segment.n, player);
			const value = evaluateTooltipExpr(segment.n, player);
			return (segment.abs ? Math.abs(value) : value).toFixed(segment.p);
		})
		.join('')
		.replace(/ {2,}/g, ' ');

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

/** The player as a tooltip reads it: the sim's final stats, a school's spell power on top of spell damage, the level and the weapons. */
export const tooltipPlayer = (stats: Stats | undefined, gear: Gear | undefined, talentPoints: ReadonlyMap<number, number>): TooltipPlayer => {
	const weapon = (slot: ItemSlot) => gear?.getEquippedItem(slot)?.item;
	return {
		value: name => {
			if (name === 'PL') return CHARACTER_LEVEL;
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
	};
};
