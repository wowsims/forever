// Which consumables the settings offer a spec, and in which picker, from the fields db.json carries
// for each. `Player.getConsumes(true)` sends the potions and conjured items this finds, so it sits
// below the view layer; the pickers in the settings feature read the same functions.
import { Class, ConsumableType, HandType, ItemSlot, Profession, PseudoStat, Race, Stat, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ResourceType, SpellEffect } from '@generated/proto/spell';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import type { SpecConfigData } from '../player/player';
import type { EquippedItem } from '../proto/equipped_item';

export enum ConsumableSlot {
	Potion,
	Conjured,
	Flask,
	Elixir,
	Food,
	Explosive,
	// A sharpening stone, weightstone, oil or shield coating: a temporary enchant, one per hand.
	Imbue,
	// A poison, firestone, spellstone or Scroll of Imbue: one per hand, beside the hand's temporary enchant.
	WeaponImbue,
	// Everything else, picked any number at a time into extra_ids.
	Extra,
}

export interface ConsumableUser {
	getClass(): Class;
	getRace(): Race;
	hasProfession(profession: Profession, rank?: number): boolean;
	getEquippedItem(slot: ItemSlot): EquippedItem | null;
	getSpecConfig(): Pick<SpecConfigData<any>, 'epStats' | 'consumableStats' | 'epPseudoStats' | 'displayStats'>;
}

export interface SpellEffectSource {
	getSpellEffect(effectId: number): SpellEffect | undefined;
}

// The generator types a consumable by the category its on-use shares where that says potion, conjured
// item, food or drink, and by its item subclass otherwise.
export function consumableSlot(consumable: Consumable): ConsumableSlot {
	switch (consumable.type) {
		case ConsumableType.ConsumableTypeConjured:
			return ConsumableSlot.Conjured;
		case ConsumableType.ConsumableTypePotion:
			return ConsumableSlot.Potion;
		case ConsumableType.ConsumableTypeFlask:
			return ConsumableSlot.Flask;
		case ConsumableType.ConsumableTypeElixir:
			return ConsumableSlot.Elixir;
		case ConsumableType.ConsumableTypeFood:
			return ConsumableSlot.Food;
		case ConsumableType.ConsumableTypeExplosive:
			return ConsumableSlot.Explosive;
		case ConsumableType.ConsumableTypeImbue:
			return ConsumableSlot.Imbue;
		case ConsumableType.ConsumableTypeWeaponImbue:
			return ConsumableSlot.WeaponImbue;
		default:
			return ConsumableSlot.Extra;
	}
}

// An explosive with a cooldown of its own besides the one explosives share is used beside the
// spec's explosive, so it is picked into extra_ids.
export const explosiveUsedBeside = (consumable: Consumable): boolean =>
	consumable.itemEffects.some(effect => effect.effect.oneofKind === 'onUse' && effect.effect.onUse.cooldownMs > 0);

// ItemSubclassWeapon of an equipped weapon: the bit a temporary enchant's weapon subclass mask sets
// for it. None for a shield or an off-hand frill.
const clientWeaponSubclass = (item: Item): number | undefined => {
	const twoHand = item.handType === HandType.HandTypeTwoHand;
	switch (item.weaponType) {
		case WeaponType.WeaponTypeAxe:
			return twoHand ? 1 : 0;
		case WeaponType.WeaponTypeMace:
			return twoHand ? 5 : 4;
		case WeaponType.WeaponTypeSword:
			return twoHand ? 8 : 7;
		case WeaponType.WeaponTypePolearm:
			return 6;
		case WeaponType.WeaponTypeStaff:
			return 10;
		case WeaponType.WeaponTypeFist:
			return 13;
		case WeaponType.WeaponTypeDagger:
			return 15;
		default:
			return undefined;
	}
};

// InventoryType of an item in a hand: the bit a temporary enchant's inventory type mask sets for it.
const clientInventoryType = (item: Item): number | undefined => {
	if (item.weaponType === WeaponType.WeaponTypeShield) return 14;
	if (item.weaponType === WeaponType.WeaponTypeOffHand) return 23;
	switch (item.handType) {
		case HandType.HandTypeOneHand:
			return 13;
		case HandType.HandTypeTwoHand:
			return 17;
		case HandType.HandTypeMainHand:
			return 21;
		case HandType.HandTypeOffHand:
			return 22;
		default:
			return undefined;
	}
};

const maskAllows = (mask: number, bit: number | undefined): boolean => mask === 0 || (bit !== undefined && (mask & (1 << bit)) !== 0);

export function handTakesImbue(consumable: Consumable, item: Item | undefined): boolean {
	if (!item) return false;
	return maskAllows(consumable.weaponSubclassMask, clientWeaponSubclass(item)) && maskAllows(consumable.weaponInventoryTypeMask, clientInventoryType(item));
}

// The stats a spec weighs consumables by. Stamina brings health and the resistances with it, and
// Attack Power the flat physical damage a Frenzy Potion or Bogling Root adds.
export function consumableStatsOf(specConfig: Pick<SpecConfigData<any>, 'epStats' | 'consumableStats'>): Set<Stat> {
	const stats = new Set([...(specConfig.consumableStats ?? []), ...specConfig.epStats]);
	if (stats.has(Stat.StatAttackPower)) stats.add(Stat.StatPhysicalDamage);
	if (stats.has(Stat.StatStamina)) {
		for (const stat of [
			Stat.StatHealth,
			Stat.StatArcaneResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatNatureResistance,
			Stat.StatShadowResistance,
		]) {
			stats.add(stat);
		}
	}
	return stats;
}

// The percents a spec shows or weighs. A consumable lists a hit or crit that reaches one channel
// alone by that channel's percents, so a spec that shows spell hit and not melee hit is not offered
// melee hit, though both come from the one hit rating the spec weighs.
export function consumablePseudoStatsOf(specConfig: Pick<SpecConfigData<any>, 'epPseudoStats' | 'displayStats'>): Set<PseudoStat> {
	const shown = (specConfig.displayStats ?? []).filter(unitStat => unitStat.isPseudoStat()).map(unitStat => unitStat.getPseudoStat());
	return new Set([...shown, ...(specConfig.epPseudoStats ?? [])]);
}

type EnchantRelevanceSpec = Pick<SpecConfigData<any>, 'epStats' | 'epPseudoStats' | 'displayStats'>;

const enchantRelevanceCache = new WeakMap<EnchantRelevanceSpec, { stats: Set<Stat>; pseudoStats: Set<PseudoStat> }>();

// The consumables' rule on the stats the spec weighs: an enchant counts when it grants one of them or a percent the
// spec shows or weighs, flat weapon damage counting as physical damage. One that grants none of these (a proc) counts.
// Enchants are weighed by the spec's EP stats alone, not by its consumableStats.
export function isEnchantRelevantFor(enchant: Enchant, specConfig: EnchantRelevanceSpec): boolean {
	const grantedStats: Stat[] = enchant.stats.flatMap((value, stat) => (value ? [stat] : []));
	if (enchant.weaponDamage) grantedStats.push(Stat.StatPhysicalDamage);
	const grantedPseudoStats: PseudoStat[] = enchant.pseudoStats.flatMap((value, pseudoStat) => (value ? [pseudoStat] : []));
	if (!grantedStats.length && !grantedPseudoStats.length) return true;

	let relevant = enchantRelevanceCache.get(specConfig);
	if (!relevant) {
		relevant = { stats: consumableStatsOf({ epStats: specConfig.epStats }), pseudoStats: consumablePseudoStatsOf(specConfig) };
		enchantRelevanceCache.set(specConfig, relevant);
	}
	const { stats, pseudoStats } = relevant;
	return grantedStats.some(stat => stats.has(stat)) || grantedPseudoStats.some(pseudoStat => pseudoStats.has(pseudoStat));
}

interface ConsumableFacts {
	db: SpellEffectSource;
	grantedStats: Array<Stat>;
	grantedPseudoStats: Array<PseudoStat>;
	restores: Array<ResourceType>;
}

// What relevance reads off a consumable and its spell effects, worked out once per consumable.
const factsCache = new WeakMap<Consumable, ConsumableFacts>();
const factsOf = (consumable: Consumable, db: SpellEffectSource): ConsumableFacts => {
	const cached = factsCache.get(consumable);
	if (cached?.db === db) return cached;
	const facts = {
		db,
		grantedStats: consumable.stats.flatMap((value, stat) => (value > 0 ? [stat as Stat] : [])),
		grantedPseudoStats: consumable.pseudoStats.flatMap((value, pseudoStat) => (value > 0 ? [pseudoStat as PseudoStat] : [])),
		restores: consumable.effectIds.flatMap(effectId => {
			const miscValue = db.getSpellEffect(effectId)?.miscValue0;
			return miscValue?.oneofKind === 'resourceType' ? [miscValue.resourceType] : [];
		}),
	};
	factsCache.set(consumable, facts);
	return facts;
};

const restoresForSpec = (resource: ResourceType, stats: Set<Stat>): boolean => {
	switch (resource) {
		case ResourceType.ResourceTypeMana:
			return stats.has(Stat.StatMana) || stats.has(Stat.StatMP5) || stats.has(Stat.StatIntellect);
		case ResourceType.ResourceTypeHealth:
			return stats.has(Stat.StatHealth);
		default:
			return stats.has(Stat.StatAttackPower);
	}
};

// Types whose consumables exist to grant a stat: one granting none the spec weighs does nothing for it.
const STAT_TYPES = [
	ConsumableType.ConsumableTypeUnknown,
	ConsumableType.ConsumableTypeConjured,
	ConsumableType.ConsumableTypeFlask,
	ConsumableType.ConsumableTypeFood,
	ConsumableType.ConsumableTypeScroll,
	ConsumableType.ConsumableTypeOther,
	ConsumableType.ConsumableTypeElixir,
];

// Whether the settings offer the consumable to this player: the class, race, profession and skill the
// game files allow it to, and either a stat the spec weighs, a percent the spec shows or weighs, a
// resource the spec spends, or, for a consumable that grants none of these, an effect of another kind
// (damage, an absorb, a weapon enchant). A temporary enchant is checked against a hand with
// handTakesImbue.
function isConsumableRelevant(consumable: Consumable, user: ConsumableUser, db: SpellEffectSource, stats: Set<Stat>, pseudoStats: Set<PseudoStat>): boolean {
	if (consumable.classAllowlist.length && !consumable.classAllowlist.includes(user.getClass())) return false;
	if (consumable.requiredProfession !== Profession.ProfessionUnknown && !user.hasProfession(consumable.requiredProfession, consumable.reqProfRank))
		return false;
	if (consumable.raceAllowlist.length && !consumable.raceAllowlist.includes(user.getRace())) return false;

	const { grantedStats, grantedPseudoStats, restores } = factsOf(consumable, db);
	if (
		consumable.buffsMainStat ||
		grantedStats.some(stat => stats.has(stat)) ||
		grantedPseudoStats.some(pseudoStat => pseudoStats.has(pseudoStat)) ||
		restores.some(resource => restoresForSpec(resource, stats))
	)
		return true;
	return !grantedStats.length && !grantedPseudoStats.length && !restores.length && !STAT_TYPES.includes(consumable.type);
}

// Roman-numeral ranks of one consumable share a name up to the numeral; a sim request lists the
// highest of those the player could use, so an unusable top rank leaves a usable lower one listed.
const RANK_SUFFIX = / (I|II|III|IV|V|VI|VII|VIII|IX|X)$/;
const ROMAN = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X'];

const rankOf = (consumable: Consumable) => {
	const match = consumable.name.match(RANK_SUFFIX);
	return { base: `${consumable.type}:${match ? consumable.name.slice(0, match.index) : consumable.name}`, rank: match ? ROMAN.indexOf(match[1]) + 1 : 1 };
};

// The ranks of each name, in the order given.
const rankLadders = (consumables: ReadonlyArray<Consumable>): Array<Array<{ consumable: Consumable; rank: number }>> => {
	const ladders = new Map<string, Array<{ consumable: Consumable; rank: number }>>();
	for (const consumable of consumables) {
		const { base, rank } = rankOf(consumable);
		const ladder = ladders.get(base);
		if (ladder) ladder.push({ consumable, rank });
		else ladders.set(base, [{ consumable, rank }]);
	}
	return [...ladders.values()];
};

function highestRanks(consumables: ReadonlyArray<Consumable>): Array<Consumable> {
	const kept = new Set(rankLadders(consumables).map(ladder => ladder.reduce((best, entry) => (entry.rank > best.rank ? entry : best)).consumable));
	return consumables.filter(consumable => kept.has(consumable));
}

// isConsumableRelevant for one spec's stats and percents, remembered for as long as the class, race
// and the profession requirements the player meets stay the same.
export function relevanceFor(
	consumables: ReadonlyArray<Consumable>,
	db: SpellEffectSource,
	stats: Set<Stat>,
	pseudoStats: Set<PseudoStat>,
): (consumable: Consumable, user: ConsumableUser) => boolean {
	const requirements = [
		...new Map(
			consumables
				.filter(consumable => consumable.requiredProfession !== Profession.ProfessionUnknown)
				.map(consumable => [`${consumable.requiredProfession}:${consumable.reqProfRank}`, consumable] as const),
		),
	];
	let key = '';
	let relevant = new Map<Consumable, boolean>();
	return (consumable, user) => {
		const met = requirements.filter(([, requirement]) => user.hasProfession(requirement.requiredProfession, requirement.reqProfRank));
		const userKey = `${user.getClass()}:${user.getRace()}:${met.map(([requirementKey]) => requirementKey)}`;
		if (userKey !== key) {
			key = userKey;
			relevant = new Map();
		}
		let result = relevant.get(consumable);
		if (result === undefined) {
			result = isConsumableRelevant(consumable, user, db, stats, pseudoStats);
			relevant.set(consumable, result);
		}
		return result;
	};
}

// The ids of the potions and conjured items the spec could use, which a sim request lists for the
// rotation to reach.
export function relevantConsumableIds(
	consumables: ReadonlyArray<Consumable>,
	slot: ConsumableSlot,
	user: ConsumableUser,
	db: SpellEffectSource,
): Array<number> {
	const stats = consumableStatsOf(user.getSpecConfig());
	const pseudoStats = consumablePseudoStatsOf(user.getSpecConfig());
	return highestRanks(
		consumables.filter(consumable => consumableSlot(consumable) === slot && isConsumableRelevant(consumable, user, db, stats, pseudoStats)),
	).map(consumable => consumable.id);
}
