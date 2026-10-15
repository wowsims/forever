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
import { ItemEffect, ResourceType, SpellEffect } from '@generated/proto/spell';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import type { EquippedItem } from '../proto/equipped_item';
import { Stats, UnitStat } from '../proto/stats';
import {
	consumableAllowed,
	ConsumableSlot,
	consumableSlot,
	consumableUniverse,
	type ConsumableUser,
	explosiveUsedBeside,
	handTakesImbue,
	isEnchantRelevantFor,
	orderedByEp,
	relevanceFor,
	relevantConsumableIds,
} from './consumables';

const stats = (granted: Partial<Record<Stat, number>>) => {
	const out = new Array<number>(42).fill(0);
	for (const [stat, value] of Object.entries(granted)) out[Number(stat)] = value!;
	return out;
};

const pseudoStats = (granted: Partial<Record<PseudoStat, number>>) => {
	const out = new Array<number>(36).fill(0);
	for (const [pseudoStat, value] of Object.entries(granted)) out[Number(pseudoStat)] = value!;
	return out;
};

const onUse = (categoryId: number, cooldownMs = 0) =>
	ItemEffect.create({ effect: { oneofKind: 'onUse', onUse: { categoryId, cooldownMs, categoryCooldownMs: 0 } } });

const consumable = (fields: Partial<Consumable>) => Consumable.create({ id: 1, name: 'Test', ...fields });

const effects = new Map<number, SpellEffect>([
	[1, SpellEffect.create({ id: 1, miscValue0: { oneofKind: 'resourceType', resourceType: ResourceType.ResourceTypeMana } })],
	[2, SpellEffect.create({ id: 2, miscValue0: { oneofKind: 'resourceType', resourceType: ResourceType.ResourceTypeHealth } })],
	[3, SpellEffect.create({ id: 3, miscValue0: { oneofKind: 'resourceType', resourceType: ResourceType.ResourceTypeRage } })],
]);
const db = { getSpellEffect: (id: number) => effects.get(id) };

const user = (
	fields: Partial<{
		cls: Class;
		race: Race;
		professions: Profession[];
		profRank: number;
		epStats: Stat[];
		epPseudoStats: PseudoStat[];
		displayPseudoStats: PseudoStat[];
		mainHand: UIItem;
		offHand: UIItem;
	}> = {},
): ConsumableUser => ({
	getClass: () => fields.cls ?? Class.ClassWarrior,
	getRace: () => fields.race ?? Race.RaceOrc,
	hasProfession: (profession, rank = 0) => (fields.professions ?? []).includes(profession) && (!fields.profRank || fields.profRank >= rank),
	getEquippedItem: slot => {
		const item = slot === ItemSlot.ItemSlotMainHand ? fields.mainHand : slot === ItemSlot.ItemSlotOffHand ? fields.offHand : undefined;
		return item ? ({ item } as EquippedItem) : null;
	},
	getSpecConfig: () => ({
		epStats: fields.epStats ?? [Stat.StatStrength, Stat.StatAttackPower],
		epPseudoStats: fields.epPseudoStats,
		displayStats: (fields.displayPseudoStats ?? []).map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat)),
	}),
});

describe('consumableSlot', () => {
	it('puts each consumable in the picker of its type, and what no picker takes among the extras', () => {
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeConjured }))).toBe(ConsumableSlot.Conjured);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypePotion }))).toBe(ConsumableSlot.Potion);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeFlask }))).toBe(ConsumableSlot.Flask);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeElixir }))).toBe(ConsumableSlot.Elixir);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeFood }))).toBe(ConsumableSlot.Food);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeExplosive }))).toBe(ConsumableSlot.Explosive);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeImbue }))).toBe(ConsumableSlot.Imbue);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeScroll }))).toBe(ConsumableSlot.Extra);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeDevice }))).toBe(ConsumableSlot.Extra);
	});
});

describe('consumable relevance', () => {
	const offered = (candidate: Consumable, player: ConsumableUser) => relevantConsumableIds([candidate], consumableSlot(candidate), player, db).length === 1;

	it('keeps a consumable off a class, a profession or a race the game files do not allow it', () => {
		const mightyRage = consumable({ classAllowlist: [Class.ClassWarrior, Class.ClassDruid], stats: stats({ [Stat.StatStrength]: 60 }) });
		expect(offered(mightyRage, user())).toBe(true);
		expect(offered(mightyRage, user({ cls: Class.ClassRogue }))).toBe(false);

		const sapper = consumable({ type: ConsumableType.ConsumableTypeExplosive, requiredProfession: Profession.Engineering });
		expect(offered(sapper, user())).toBe(false);
		expect(offered(sapper, user({ professions: [Profession.Engineering] }))).toBe(true);

		// Dark Iron Bomb 16005: Engineering 285.
		const darkIronBomb = consumable({ type: ConsumableType.ConsumableTypeExplosive, requiredProfession: Profession.Engineering, reqProfRank: 285 });
		expect(offered(darkIronBomb, user({ professions: [Profession.Engineering] }))).toBe(true);
		expect(offered(darkIronBomb, user({ professions: [Profession.Engineering], profRank: 300 }))).toBe(true);
		expect(offered(darkIronBomb, user({ professions: [Profession.Engineering], profRank: 250 }))).toBe(false);

		const orcOnly = consumable({ type: ConsumableType.ConsumableTypeExplosive, raceAllowlist: [Race.RaceOrc] });
		expect(offered(orcOnly, user())).toBe(true);
		expect(offered(orcOnly, user({ race: Race.RaceHuman }))).toBe(false);
	});

	it('offers a stat consumable by a stat the spec weighs, Stamina bringing the resistances', () => {
		const giants = consumable({ type: ConsumableType.ConsumableTypeElixir, stats: stats({ [Stat.StatStrength]: 25 }) });
		const arthas = consumable({ type: ConsumableType.ConsumableTypeElixir, stats: stats({ [Stat.StatShadowResistance]: 10 }) });
		expect(offered(giants, user())).toBe(true);
		expect(offered(giants, user({ epStats: [Stat.StatIntellect] }))).toBe(false);
		expect(offered(arthas, user())).toBe(false);
		expect(offered(arthas, user({ epStats: [Stat.StatStamina] }))).toBe(true);
	});

	it('offers flat physical damage to a spec that weighs Attack Power', () => {
		const frenzy = consumable({ type: ConsumableType.ConsumableTypePotion, stats: stats({ [Stat.StatPhysicalDamage]: 40 }) });
		expect(offered(frenzy, user({ epStats: [Stat.StatAttackPower] }))).toBe(true);
		expect(offered(frenzy, user({ epStats: [Stat.StatIntellect] }))).toBe(false);
	});

	it('offers a one-channel percent to a spec that shows or weighs it, not by the rating the spec weighs', () => {
		// Dark Desire 22237: A_MOD_HIT_CHANCE 2, melee and ranged hit alone.
		const darkDesire = consumable({
			type: ConsumableType.ConsumableTypeOther,
			pseudoStats: pseudoStats({ [PseudoStat.PseudoStatMeleeHitPercent]: 2, [PseudoStat.PseudoStatRangedHitPercent]: 2 }),
		});
		const caster = { epStats: [Stat.StatSpellDamage, Stat.StatHitRating], displayPseudoStats: [PseudoStat.PseudoStatSpellHitPercent] };
		expect(offered(darkDesire, user(caster))).toBe(false);
		expect(offered(darkDesire, user({ ...caster, displayPseudoStats: [PseudoStat.PseudoStatMeleeHitPercent] }))).toBe(true);
		expect(offered(darkDesire, user({ ...caster, epPseudoStats: [PseudoStat.PseudoStatRangedHitPercent] }))).toBe(true);
	});

	it('offers an all-channel crit percent to a melee spec and to a caster that shows spell crit', () => {
		// Elixir of the Mongoose 13452: A_MOD_CRIT_PCT 2, melee, ranged and spell crit alike.
		const mongoose = consumable({
			type: ConsumableType.ConsumableTypeElixir,
			pseudoStats: pseudoStats({
				[PseudoStat.PseudoStatMeleeCritPercent]: 2,
				[PseudoStat.PseudoStatRangedCritPercent]: 2,
				[PseudoStat.PseudoStatSpellCritPercent]: 2,
			}),
		});
		const rogue = user({ cls: Class.ClassRogue, epStats: [Stat.StatAgility], displayPseudoStats: [PseudoStat.PseudoStatMeleeCritPercent] });
		const mage = user({ cls: Class.ClassMage, epStats: [Stat.StatIntellect], displayPseudoStats: [PseudoStat.PseudoStatSpellCritPercent] });
		expect(offered(mongoose, rogue)).toBe(true);
		expect(offered(mongoose, mage)).toBe(true);
	});

	it('offers a restore by the resource the spec spends', () => {
		const mana = consumable({ type: ConsumableType.ConsumableTypePotion, effectIds: [1] });
		const health = consumable({ type: ConsumableType.ConsumableTypePotion, effectIds: [2] });
		const rage = consumable({ type: ConsumableType.ConsumableTypePotion, effectIds: [3] });
		expect(offered(mana, user())).toBe(false);
		expect(offered(mana, user({ epStats: [Stat.StatIntellect] }))).toBe(true);
		expect(offered(health, user())).toBe(false);
		expect(offered(health, user({ epStats: [Stat.StatStamina] }))).toBe(true);
		expect(offered(rage, user())).toBe(true);
		expect(offered(rage, user({ epStats: [Stat.StatIntellect] }))).toBe(false);
	});

	it('offers a consumable of another kind to everyone, but not a stat consumable that grants nothing', () => {
		expect(offered(consumable({ type: ConsumableType.ConsumableTypeDevice }), user())).toBe(true);
		expect(offered(consumable({ type: ConsumableType.ConsumableTypeExplosive }), user())).toBe(true);
		expect(offered(consumable({ type: ConsumableType.ConsumableTypeOther }), user())).toBe(false);
		expect(offered(consumable({ type: ConsumableType.ConsumableTypeElixir }), user())).toBe(false);
	});
});

describe('handTakesImbue', () => {
	const item = (weaponType: WeaponType, handType = HandType.HandTypeOneHand) => UIItem.create({ weaponType, handType });
	// Dense Sharpening Stone 12404: SpellEquippedItems subclass mask 33219.
	const sharpeningStone = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 33219 });
	// Dense Weightstone 12643: 9264.
	const weightstone = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 9264 });
	// Brilliant Wizard Oil 20749: inventory type mask 6430720.
	const oil = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponInventoryTypeMask: 6430720 });
	// Conductive Shield Coating 228980: inventory type mask 16384.
	const coating = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponInventoryTypeMask: 16384 });

	it('takes a stone by the weapon subclass', () => {
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeSword))).toBe(true);
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeAxe, HandType.HandTypeTwoHand))).toBe(true);
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeMace))).toBe(false);
		expect(handTakesImbue(weightstone, item(WeaponType.WeaponTypeMace, HandType.HandTypeTwoHand))).toBe(true);
		expect(handTakesImbue(weightstone, item(WeaponType.WeaponTypeDagger))).toBe(false);
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeShield))).toBe(false);
	});

	it('takes an oil or a shield coating by the inventory type', () => {
		expect(handTakesImbue(oil, item(WeaponType.WeaponTypeDagger, HandType.HandTypeMainHand))).toBe(true);
		expect(handTakesImbue(oil, item(WeaponType.WeaponTypeStaff, HandType.HandTypeTwoHand))).toBe(true);
		expect(handTakesImbue(oil, item(WeaponType.WeaponTypeShield, HandType.HandTypeOffHand))).toBe(false);
		expect(handTakesImbue(coating, item(WeaponType.WeaponTypeShield, HandType.HandTypeOffHand))).toBe(true);
		expect(handTakesImbue(coating, item(WeaponType.WeaponTypeSword, HandType.HandTypeOffHand))).toBe(false);
	});

	it('takes nothing on an empty hand', () => {
		expect(handTakesImbue(oil, undefined)).toBe(false);
	});
});

describe('consumableAllowed', () => {
	it('allows a consumable by the class, race and profession rank the game files name, whatever it grants', () => {
		const sapper = consumable({ type: ConsumableType.ConsumableTypeExplosive, requiredProfession: Profession.Engineering, reqProfRank: 205 });
		const orcOnly = consumable({ id: 2, type: ConsumableType.ConsumableTypeExplosive, raceAllowlist: [Race.RaceOrc] });
		const rogueOnly = consumable({ id: 3, classAllowlist: [Class.ClassRogue], stats: stats({ [Stat.StatIntellect]: 5 }) });
		expect(consumableAllowed(sapper, user())).toBe(false);
		expect(consumableAllowed(sapper, user({ professions: [Profession.Engineering] }))).toBe(true);
		expect(consumableAllowed(sapper, user({ professions: [Profession.Engineering], profRank: 150 }))).toBe(false);
		expect(consumableAllowed(orcOnly, user())).toBe(true);
		expect(consumableAllowed(orcOnly, user({ race: Race.RaceHuman }))).toBe(false);
		expect(consumableAllowed(rogueOnly, user())).toBe(false);
		expect(consumableAllowed(rogueOnly, user({ cls: Class.ClassRogue }))).toBe(true);
	});
});

describe('relevanceFor', () => {
	const agility = consumable({ type: ConsumableType.ConsumableTypeElixir, stats: stats({ [Stat.StatAgility]: 25 }) });
	const manaPotion = consumable({ id: 2, type: ConsumableType.ConsumableTypePotion, effectIds: [1] });
	const deflector = consumable({ id: 3, type: ConsumableType.ConsumableTypeDevice });
	const rogueOnly = consumable({
		id: 4,
		type: ConsumableType.ConsumableTypeElixir,
		classAllowlist: [Class.ClassRogue],
		stats: stats({ [Stat.StatAgility]: 5 }),
	});
	const statsOf = (shown: Stat[]) => ({ stats: new Set(shown), pseudoStats: new Set<PseudoStat>() });

	it('reads the stats statsOf gives, and again once it gives another set', () => {
		let shown = statsOf([Stat.StatAgility]);
		const relevant = relevanceFor([agility, manaPotion], db, () => shown);
		expect(relevant(agility, user())).toBe(true);
		shown = statsOf([Stat.StatStrength]);
		expect(relevant(agility, user())).toBe(false);
	});

	it('restores mana while Mana, MP5 or Intellect is shown', () => {
		let shown = statsOf([Stat.StatIntellect]);
		const relevant = relevanceFor([manaPotion], db, () => shown);
		expect(relevant(manaPotion, user())).toBe(true);
		shown = statsOf([Stat.StatSpirit]);
		expect(relevant(manaPotion, user())).toBe(false);
	});

	it('keeps a consumable with no stat and an effect of another kind at any set', () => {
		const relevant = relevanceFor([deflector], db, () => statsOf([]));
		expect(relevant(deflector, user())).toBe(true);
	});

	it('leaves the class, race and profession to consumableAllowed', () => {
		const relevant = relevanceFor([rogueOnly], db, () => statsOf([Stat.StatAgility]));
		expect(relevant(rogueOnly, user())).toBe(true);
	});

	it('takes a stat granted only against some mob types or in some zones as granted', () => {
		const slaying = consumable({
			id: 5,
			type: ConsumableType.ConsumableTypePotion,
			mobTypeStats: [MobTypeStats.create({ mobTypes: [MobType.MobTypeDemon], stats: { [Stat.StatAttackPower]: 20 } })],
		});
		const natural = consumable({
			id: 6,
			type: ConsumableType.ConsumableTypeFlask,
			zoneStats: [ZoneStats.create({ zoneIds: [101], stats: { [Stat.StatHitRating]: 20 } })],
		});
		let shown = statsOf([Stat.StatAttackPower, Stat.StatHitRating]);
		const relevant = relevanceFor([slaying, natural], db, () => shown);
		expect(relevant(slaying, user())).toBe(true);
		expect(relevant(natural, user())).toBe(true);
		shown = statsOf([Stat.StatIntellect]);
		expect(relevant(slaying, user())).toBe(false);
		expect(relevant(natural, user())).toBe(false);
	});
});

describe('consumableUniverse', () => {
	it('lists every stat and percent a consumable grants, once', () => {
		const mongoose = consumable({
			id: 5,
			type: ConsumableType.ConsumableTypeElixir,
			stats: stats({ [Stat.StatAgility]: 25 }),
			pseudoStats: pseudoStats({ [PseudoStat.PseudoStatMeleeCritPercent]: 2 }),
		});
		const giants = consumable({ id: 6, type: ConsumableType.ConsumableTypeElixir, stats: stats({ [Stat.StatStrength]: 25 }) });
		const moreAgility = consumable({ id: 8, type: ConsumableType.ConsumableTypeFood, stats: stats({ [Stat.StatAgility]: 10 }) });
		const manaPotion = consumable({ id: 7, type: ConsumableType.ConsumableTypePotion, effectIds: [1] });
		const keys = consumableUniverse([mongoose, giants, moreAgility, manaPotion], db).map(stat => stat.getKey());
		expect(keys).toHaveLength(3);
		expect(new Set(keys)).toEqual(new Set(['StatStrength', 'StatAgility', 'PseudoStatMeleeCritPercent']));
	});

	it('lists a stat granted only against some mob types or in some zones', () => {
		const slaying = consumable({
			id: 9,
			type: ConsumableType.ConsumableTypePotion,
			mobTypeStats: [MobTypeStats.create({ mobTypes: [MobType.MobTypeDemon], stats: { [Stat.StatAttackPower]: 20 } })],
		});
		const natural = consumable({
			id: 10,
			type: ConsumableType.ConsumableTypeFlask,
			zoneStats: [ZoneStats.create({ zoneIds: [101], stats: { [Stat.StatHitRating]: 20 } })],
		});
		const keys = consumableUniverse([slaying, natural], db).map(stat => stat.getKey());
		expect(new Set(keys)).toEqual(new Set(['StatAttackPower', 'StatHitRating']));
	});
});

describe('isEnchantRelevantFor', () => {
	const melee = { epStats: [Stat.StatStrength, Stat.StatAttackPower, Stat.StatStamina], epPseudoStats: [PseudoStat.PseudoStatDodgePercent] };
	const caster = { epStats: [Stat.StatIntellect, Stat.StatSpellDamage], epPseudoStats: [] };
	const relevant = (enchant: UIEnchant, specConfig: { epStats: Stat[]; epPseudoStats: PseudoStat[] } = melee) =>
		isEnchantRelevantFor(enchant, { ...specConfig, displayStats: [] });
	const statEnchant = (effectId: number, stat: Stat) => UIEnchant.create({ effectId, stats: stats({ [stat]: 3 }) });
	const striking = UIEnchant.create({ effectId: 1897, weaponDamage: 5 });

	it('keeps an enchant granting a stat the spec weighs and drops one granting only stats it does not', () => {
		expect(relevant(statEnchant(3, Stat.StatStrength))).toBe(true);
		expect(relevant(statEnchant(4, Stat.StatIntellect))).toBe(false);
	});

	it('widens the weighed stats as the consumables do: stamina brings the resistances', () => {
		expect(relevant(statEnchant(5, Stat.StatFireResistance))).toBe(true);
		expect(relevant(statEnchant(5, Stat.StatFireResistance), caster)).toBe(false);
	});

	it('keeps a percent the spec weighs', () => {
		const dodge = UIEnchant.create({ effectId: 6, pseudoStats: pseudoStats({ [PseudoStat.PseudoStatDodgePercent]: 1 }) });
		expect(relevant(dodge)).toBe(true);
		expect(relevant(dodge, caster)).toBe(false);
	});

	it('counts flat weapon damage as physical damage, which weighing attack power covers', () => {
		expect(relevant(striking)).toBe(true);
		expect(relevant(striking, caster)).toBe(false);
	});

	it('keeps an enchant that grants no stat at all, such as a proc', () => {
		expect(relevant(UIEnchant.create({ effectId: 7 }), caster)).toBe(true);
	});
});

describe('explosiveUsedBeside', () => {
	it('names an explosive with a cooldown of its own', () => {
		expect(explosiveUsedBeside(consumable({ itemEffects: [onUse(24, 300000)] }))).toBe(true);
		expect(explosiveUsedBeside(consumable({ itemEffects: [onUse(24)] }))).toBe(false);
	});
});

describe('orderedByEp', () => {
	const strong = consumable({ id: 1, name: 'Strong', lvl: 10, stats: stats({ [Stat.StatStrength]: 10 }) });
	const critical = consumable({ id: 2, name: 'Critical', lvl: 20, pseudoStats: pseudoStats({ [PseudoStat.PseudoStatMeleeCritPercent]: 1 }) });
	const proc = consumable({ id: 3, name: 'Proc', lvl: 60 });
	const ids = (consumables: Array<Consumable>) => consumables.map(({ id }) => id);
	const orderedIds = (consumables: Array<Consumable>, weights: Stats, mobType: MobType | undefined, zoneId: number) =>
		ids(orderedByEp(consumables, consumable => consumable, { weights, mobType, zoneId }));
	const anywhere = (consumables: Array<Consumable>, weights: Stats) => orderedIds(consumables, weights, undefined, 0);

	it('orders by the EP the stats and percents add, highest first', () => {
		const weights = Stats.fromMap({ [Stat.StatStrength]: 1 }, { [PseudoStat.PseudoStatMeleeCritPercent]: 20 });
		expect(anywhere([strong, critical], weights)).toEqual([2, 1]);
	});

	it('orders again by other weights', () => {
		const weights = Stats.fromMap({ [Stat.StatStrength]: 3 }, { [PseudoStat.PseudoStatMeleeCritPercent]: 20 });
		expect(anywhere([critical, strong], weights)).toEqual([1, 2]);
	});

	it('puts one granting no stat after every one that does, whatever its level', () => {
		expect(anywhere([proc, strong, critical], new Stats())).toEqual([2, 1, 3]);
	});

	it('orders an equal EP by level, highest first, then by name', () => {
		const bravo = consumable({ id: 4, name: 'Bravo', lvl: 30, stats: stats({ [Stat.StatStrength]: 5 }) });
		const alpha = consumable({ id: 5, name: 'Alpha', lvl: 30, stats: stats({ [Stat.StatStrength]: 5 }) });
		const weights = Stats.fromMap({ [Stat.StatStrength]: 2 });
		expect(anywhere([bravo, strong, alpha], weights)).toEqual([1, 5, 4]);
	});

	it('counts a mob type stat against a target of one of its mob types only', () => {
		const slaying = consumable({
			id: 6,
			name: 'Slaying',
			lvl: 10,
			stats: stats({}),
			mobTypeStats: [MobTypeStats.create({ mobTypes: [MobType.MobTypeDemon, MobType.MobTypeUndead], stats: { [Stat.StatAttackPower]: 20 } })],
		});
		const weights = Stats.fromMap({ [Stat.StatStrength]: 1, [Stat.StatAttackPower]: 1 });
		expect(orderedIds([strong, proc, slaying], weights, MobType.MobTypeDemon, 0)).toEqual([6, 1, 3]);
		expect(orderedIds([strong, proc, slaying], weights, MobType.MobTypeUndead, 0)).toEqual([6, 1, 3]);
		expect(orderedIds([slaying, proc, strong], weights, MobType.MobTypeMechanical, 0)).toEqual([1, 3, 6]);
		expect(orderedIds([slaying, proc, strong], weights, undefined, 0)).toEqual([1, 3, 6]);
	});

	it('counts a zone stat in one of its zones only, and the plain stats everywhere', () => {
		const natural = consumable({
			id: 7,
			name: 'Natural',
			lvl: 10,
			stats: stats({ [Stat.StatStamina]: 10 }),
			zoneStats: [ZoneStats.create({ zoneIds: [101, 102], stats: { [Stat.StatHitRating]: 20 } })],
		});
		const sturdy = consumable({ id: 8, name: 'Sturdy', lvl: 10, stats: stats({ [Stat.StatStamina]: 15 }) });
		const weights = Stats.fromMap({ [Stat.StatStamina]: 1, [Stat.StatHitRating]: 1 });
		expect(orderedIds([sturdy, natural], weights, undefined, 101)).toEqual([7, 8]);
		expect(orderedIds([sturdy, natural], weights, undefined, 102)).toEqual([7, 8]);
		expect(orderedIds([natural, sturdy], weights, undefined, 103)).toEqual([8, 7]);
		expect(orderedIds([natural, sturdy], weights, undefined, 0)).toEqual([8, 7]);
	});
});

describe('relevantConsumableIds', () => {
	it('lists the relevant consumables of one slot, the highest rank of each', () => {
		const consumables = [
			consumable({ id: 13442, name: 'Mighty Rage Potion', type: ConsumableType.ConsumableTypePotion, categoryId: 4, effectIds: [3] }),
			consumable({ id: 13444, name: 'Major Mana Potion', type: ConsumableType.ConsumableTypePotion, categoryId: 4, effectIds: [1] }),
			consumable({ id: 9421, name: 'Major Healthstone', type: ConsumableType.ConsumableTypeConjured, categoryId: 1153, effectIds: [2] }),
		];
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, user(), db)).toEqual([13442]);
		expect(relevantConsumableIds(consumables, ConsumableSlot.Conjured, user({ epStats: [Stat.StatStamina] }), db)).toEqual([9421]);
	});

	it('picks the rank among the relevant ones, so an unusable top rank leaves the one below', () => {
		const consumables = [
			consumable({ id: 1, name: 'Rage Potion', type: ConsumableType.ConsumableTypePotion, categoryId: 4, effectIds: [3] }),
			consumable({
				id: 2,
				name: 'Rage Potion II',
				type: ConsumableType.ConsumableTypePotion,
				categoryId: 4,
				effectIds: [3],
				classAllowlist: [Class.ClassRogue],
			}),
		];
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, user(), db)).toEqual([1]);
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, user({ cls: Class.ClassRogue }), db)).toEqual([2]);
	});

	it('keeps the highest numeral of a name, and every consumable whose name has none', () => {
		const agility = stats({ [Stat.StatAgility]: 5 });
		const consumables = [
			consumable({ id: 3012, name: 'Scroll of Agility', type: ConsumableType.ConsumableTypeScroll, stats: agility }),
			consumable({ id: 1477, name: 'Scroll of Agility II', type: ConsumableType.ConsumableTypeScroll, stats: agility }),
			consumable({ id: 10309, name: 'Scroll of Agility IV', type: ConsumableType.ConsumableTypeScroll, stats: agility }),
			consumable({ id: 4425, name: 'Scroll of Agility III', type: ConsumableType.ConsumableTypeScroll, stats: agility }),
			consumable({ id: 5633, name: 'Great Rage Potion', type: ConsumableType.ConsumableTypePotion, effectIds: [3] }),
			consumable({ id: 13442, name: 'Mighty Rage Potion', type: ConsumableType.ConsumableTypePotion, effectIds: [3] }),
		];
		const warrior = user({ epStats: [Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower] });
		expect(relevantConsumableIds(consumables, ConsumableSlot.Extra, warrior, db)).toEqual([10309]);
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, warrior, db)).toEqual([5633, 13442]);
	});
});
