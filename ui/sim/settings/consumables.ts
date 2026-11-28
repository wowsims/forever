// Which consumables the settings offer a spec, and in which picker, from the fields db.json carries
// for each. `Player.getConsumes(true)` sends the potions and conjured items this finds, so it sits
// below the view layer.
import {
	Class,
	ConsumableType,
	HandType,
	ItemSlot,
	MobType,
	MobTypeStats,
	Profession,
	PseudoStat,
	Race,
	Stat,
	WeaponType,
	ZoneStats,
} from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ResourceType, SpellEffect } from '@generated/proto/spell';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import type { StoreField } from '../hooks/useStoreField';
import type { Player, SpecConfigData } from '../player/player';
import type { EquippedItem } from '../proto/equipped_item';
import { Stats, type StatSets, type UnitStat, unitStatsOf } from '../proto/stats';
import { memoLast } from '../utils/misc';

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

// The stats a spec weighs consumables by.
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

const enchantRelevanceCache = new WeakMap<EnchantRelevanceSpec, StatSets>();

// The consumables' rule on the stats the spec weighs. Enchants are weighed by the spec's EP stats alone, not by its
// consumableStats.
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

const statsWith = (consumable: Consumable, conditional: ReadonlyArray<MobTypeStats | ZoneStats>): Stats =>
	conditional.reduce((sum, { stats }) => sum.add(Stats.fromMap(stats)), new Stats(consumable.stats, consumable.pseudoStats));

// What relevance reads off a consumable and its spell effects, worked out once per consumable. A stat
// the consumable grants only against some mob types or in some zones counts as granted.
const factsCache = new WeakMap<Consumable, ConsumableFacts>();
const factsOf = (consumable: Consumable, db: SpellEffectSource): ConsumableFacts => {
	const cached = factsCache.get(consumable);
	if (cached?.db === db) return cached;
	const facts = {
		db,
		grantedStats: statsWith(consumable, [...consumable.mobTypeStats, ...consumable.zoneStats])
			.asProtoArray()
			.flatMap((value, stat) => (value > 0 ? [stat as Stat] : [])),
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

// The class, race, profession and skill the game files allow the consumable to.
export function consumableAllowed(consumable: Consumable, user: ConsumableUser): boolean {
	if (consumable.classAllowlist.length && !consumable.classAllowlist.includes(user.getClass())) return false;
	if (consumable.requiredProfession !== Profession.ProfessionUnknown && !user.hasProfession(consumable.requiredProfession, consumable.reqProfRank))
		return false;
	return !consumable.raceAllowlist.length || consumable.raceAllowlist.includes(user.getRace());
}

// Whether the consumable does something for these stats: a stat among them, a percent among them, a
// resource they spend, or, for a consumable that grants none of these, an effect of another kind.
function isConsumableRelevant(consumable: Consumable, db: SpellEffectSource, stats: Set<Stat>, pseudoStats: Set<PseudoStat>): boolean {
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

// The stat part of relevance for the stats `statsOf` gives, remembered for as long as it gives the
// same object. The class, race and profession are consumableAllowed's.
export function relevanceFor(
	consumables: ReadonlyArray<Consumable>,
	db: SpellEffectSource,
	statsOf: (user: ConsumableUser) => StatSets,
): (consumable: Consumable, user: ConsumableUser) => boolean {
	const cacheFor = memoLast<Map<Consumable, boolean>>();
	return (consumable, user) => {
		const current = statsOf(user);
		const relevant = cacheFor([current], () => new Map());
		let result = relevant.get(consumable);
		if (result === undefined) {
			result = isConsumableRelevant(consumable, db, current.stats, current.pseudoStats);
			relevant.set(consumable, result);
		}
		return result;
	};
}

// Every stat and percent some consumable grants, each once, in no set order.
export function consumableUniverse(consumables: ReadonlyArray<Consumable>, db: SpellEffectSource): UnitStat[] {
	const stats = new Set<Stat>();
	const pseudoStats = new Set<PseudoStat>();
	for (const consumable of consumables) {
		const { grantedStats, grantedPseudoStats } = factsOf(consumable, db);
		grantedStats.forEach(stat => stats.add(stat));
		grantedPseudoStats.forEach(pseudoStat => pseudoStats.add(pseudoStat));
	}
	return unitStatsOf({ stats, pseudoStats });
}

// What the pickers order consumables by: the EP weights, the class they value percents for, and the target's mob
// type, which decides the stats a consumable grants only against some targets.
export interface ConsumableOrder {
	weights: Stats;
	playerClass: Class;
	mobType: MobType | undefined;
}

export const consumableOrderOf = (player: Player<any>): ConsumableOrder => ({
	weights: player.getEpWeights(),
	playerClass: player.getClass(),
	mobType: player.sim.encounter.getTarget(0)?.mobType,
});

// The store fields consumableOrderOf reads.
export const CONSUMABLE_ORDER_FIELDS: ReadonlyArray<StoreField> = ['epWeights', 'encounter:targets'];

// The items by the EP their consumables' stats and percents add at the order's weights, highest
// first. A zone stat always counts: the sim takes every fight to be in the Natural flasks' zones.
export function orderedByEp<T>(items: ReadonlyArray<T>, consumableOf: (item: T) => Consumable, { weights, playerClass, mobType }: ConsumableOrder): Array<T> {
	return items
		.map(item => {
			const consumable = consumableOf(item);
			const conditional = [
				...consumable.mobTypeStats.filter(({ mobTypes }) => mobType !== undefined && mobTypes.includes(mobType)),
				...consumable.zoneStats,
			];
			const stats = statsWith(consumable, conditional);
			const grants = stats.asProtoArray().some(value => value > 0) || consumable.pseudoStats.some(value => value > 0);
			return { item, consumable, grants, ep: grants ? stats.computeEP(weights, playerClass) : 0 };
		})
		.sort(
			(a, b) =>
				Number(b.grants) - Number(a.grants) ||
				b.ep - a.ep ||
				b.consumable.lvl - a.consumable.lvl ||
				a.consumable.name.localeCompare(b.consumable.name, 'en'),
		)
		.map(({ item }) => item);
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
		consumables.filter(
			consumable =>
				consumableSlot(consumable) === slot && consumableAllowed(consumable, user) && isConsumableRelevant(consumable, db, stats, pseudoStats),
		),
	).map(consumable => consumable.id);
}
