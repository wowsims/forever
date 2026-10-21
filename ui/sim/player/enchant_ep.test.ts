import { ItemSlot, PseudoStat, Stat } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import * as Mechanics from '../constants/mechanics';
import { Stats } from '../proto/stats';
import { Player } from './player';

const weights = Stats.fromMap({ [Stat.StatStrength]: 2 }, { [PseudoStat.PseudoStatMainHandDps]: 10, [PseudoStat.PseudoStatOffHandDps]: 4 });

const computeEnchantEP = (enchant: Enchant, slot?: ItemSlot, weapon?: Item | null, epWeights: Stats = weights) => {
	const player = { enchantEPCache: new Map<number, number>(), computeStatsEP: (stats: Stats) => stats.computeEP(epWeights) } as unknown as Player<any>;
	const epPerWeaponDps = slot === undefined ? 0 : Player.prototype.computeWeaponDpsEP.call(player, slot);
	return Player.prototype.computeEnchantEP.call(player, enchant, weapon?.weaponSpeed ?? 0, epPerWeaponDps);
};

const enchantEP = (effectId: number, pseudoStats: Partial<Record<PseudoStat, number>>, epWeights: Stats) =>
	computeEnchantEP(Enchant.create({ effectId, pseudoStats: Stats.fromMap({}, pseudoStats).toProto().pseudoStats }), undefined, undefined, epWeights);

const striking = Enchant.create({ effectId: 1897, weaponDamage: 5 });
const sword = Item.create({ id: 1, weaponSpeed: 2.5 });

describe('Player.computeEnchantEP', () => {
	it('values flat weapon damage as the DPS it adds at the weapon speed of the slot it enchants', () => {
		expect(computeEnchantEP(striking, ItemSlot.ItemSlotMainHand, sword)).toBeCloseTo((5 / 2.5) * 10);
		expect(computeEnchantEP(striking, ItemSlot.ItemSlotOffHand, sword)).toBeCloseTo((5 / 2.5) * 4);
	});

	it('adds the weapon damage on top of the stats', () => {
		const enchant = Enchant.create({ effectId: 2, weaponDamage: 5, stats: new Stats().withStat(Stat.StatStrength, 3).asProtoArray() });
		expect(computeEnchantEP(enchant, ItemSlot.ItemSlotMainHand, sword)).toBeCloseTo(3 * 2 + (5 / 2.5) * 10);
	});

	it('gives weapon damage nothing without a weapon to read the speed off', () => {
		expect(computeEnchantEP(striking, ItemSlot.ItemSlotMainHand, null)).toBe(0);
		expect(computeEnchantEP(striking)).toBe(0);
	});

	it('values ArmorNoMods at its own weight, with no scale', () => {
		const armorWeights = Stats.fromMap({ [Stat.StatArmor]: 0.5, [Stat.StatArmorNoMods]: 0.125 });
		const armor = (stat: Stat) => Enchant.create({ effectId: 1504, stats: new Stats().withStat(stat, 125).asProtoArray() });
		expect(computeEnchantEP(armor(Stat.StatArmorNoMods), undefined, undefined, armorWeights)).toBeCloseTo(125 * 0.125);
		expect(computeEnchantEP(armor(Stat.StatArmor), undefined, undefined, armorWeights)).toBeCloseTo(125 * 0.5);
	});
});

describe('Player.computeEnchantEP on percent pseudo stats', () => {
	const ratingWeights = Stats.fromMap({ [Stat.StatDodgeRating]: 2, [Stat.StatCritRating]: 3, [Stat.StatHitRating]: 5 }, {});

	it('values a school hit percent at its per percent weight', () => {
		const schoolWeights = Stats.fromMap({}, { [PseudoStat.PseudoStatSchoolHitPercentHoly]: 23.2 });
		expect(enchantEP(1, { [PseudoStat.PseudoStatSchoolHitPercentHoly]: 1 }, schoolWeights)).toBeCloseTo(23.2);
	});

	it('values 2622 Enchant Cloak - Dodge at the dodge rating weight', () => {
		expect(enchantEP(2622, { [PseudoStat.PseudoStatDodgePercent]: 1 }, ratingWeights)).toBeCloseTo(Mechanics.DODGE_RATING_PER_DODGE_PERCENT * 2);
	});

	it("counts 2717 Might of the Scourge's melee and ranged crit once, at the crit rating weight", () => {
		expect(enchantEP(2717, { [PseudoStat.PseudoStatMeleeCritPercent]: 1, [PseudoStat.PseudoStatRangedCritPercent]: 1 }, ratingWeights)).toBeCloseTo(
			Mechanics.CRIT_RATING_PER_CRIT_PERCENT * 3,
		);
	});

	it('values the ranged-only 2523 Biznicks 247x128 Accurascope at the hit rating weight', () => {
		expect(enchantEP(2523, { [PseudoStat.PseudoStatRangedHitPercent]: 3 }, ratingWeights)).toBeCloseTo(3 * Mechanics.HIT_RATING_PER_HIT_PERCENT * 5);
	});

	it('values a melee-only hit percent at the hit rating weight', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHitPercent]: 1 }, ratingWeights)).toBeCloseTo(Mechanics.HIT_RATING_PER_HIT_PERCENT * 5);
	});

	it("uses a percent pseudo stat's own weight where it has one", () => {
		const ownWeight = Stats.fromMap({ [Stat.StatCritRating]: 3 }, { [PseudoStat.PseudoStatMeleeCritPercent]: 7 });
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeCritPercent]: 1 }, ownWeight)).toBeCloseTo(7);
	});

	it("values 2717 Might of the Scourge's ranged crit at what the crit rating is worth past the melee crit weight", () => {
		const ownWeight = Stats.fromMap({ [Stat.StatCritRating]: 3 }, { [PseudoStat.PseudoStatMeleeCritPercent]: 7 });
		expect(enchantEP(2717, { [PseudoStat.PseudoStatMeleeCritPercent]: 1, [PseudoStat.PseudoStatRangedCritPercent]: 1 }, ownWeight)).toBeCloseTo(
			7 + (Mechanics.CRIT_RATING_PER_CRIT_PERCENT * 3 - 7),
		);
	});

	it("values 2586 Falcon's Call's ranged hit at the ranged hit weight and its melee hit at what the hit rating is worth past it", () => {
		const hunterWeights = Stats.fromMap({ [Stat.StatHitRating]: 5 }, { [PseudoStat.PseudoStatRangedHitPercent]: 5 });
		expect(enchantEP(2586, { [PseudoStat.PseudoStatMeleeHitPercent]: 1, [PseudoStat.PseudoStatRangedHitPercent]: 1 }, hunterWeights)).toBeCloseTo(
			5 + (Mechanics.HIT_RATING_PER_HIT_PERCENT * 5 - 5),
		);
	});

	it('values melee and ranged hit with no spell hit at the melee and ranged hit weights', () => {
		const hitWeights = Stats.fromMap(
			{ [Stat.StatHitRating]: 5 },
			{ [PseudoStat.PseudoStatMeleeHitPercent]: 40, [PseudoStat.PseudoStatRangedHitPercent]: 30 },
		);
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHitPercent]: 1, [PseudoStat.PseudoStatRangedHitPercent]: 1 }, hitWeights)).toBeCloseTo(40 + 30);
	});

	it('values hit on every channel at the ranged hit weight for a spec that leaves hit rating unweighted', () => {
		const rangedHitWeights = Stats.fromMap({}, { [PseudoStat.PseudoStatRangedHitPercent]: 30 });
		expect(
			enchantEP(
				1,
				{
					[PseudoStat.PseudoStatMeleeHitPercent]: 1,
					[PseudoStat.PseudoStatRangedHitPercent]: 1,
					[PseudoStat.PseudoStatSpellHitPercent]: 1,
				},
				rangedHitWeights,
			),
		).toBeCloseTo(30);
	});

	it('values ranged hit alone at the ranged hit weight for a spec weighing it', () => {
		const hunterWeights = Stats.fromMap({ [Stat.StatHitRating]: 5 }, { [PseudoStat.PseudoStatRangedHitPercent]: 8 });
		expect(enchantEP(2523, { [PseudoStat.PseudoStatRangedHitPercent]: 3 }, hunterWeights)).toBeCloseTo(3 * 8);
	});

	it('values hit on every channel at the hit rating weight alone, though the spec weighs melee and ranged hit', () => {
		const hitWeights = Stats.fromMap(
			{ [Stat.StatHitRating]: 5 },
			{ [PseudoStat.PseudoStatMeleeHitPercent]: 40, [PseudoStat.PseudoStatRangedHitPercent]: 30 },
		);
		expect(
			enchantEP(
				1,
				{
					[PseudoStat.PseudoStatMeleeHitPercent]: 1,
					[PseudoStat.PseudoStatRangedHitPercent]: 1,
					[PseudoStat.PseudoStatSpellHitPercent]: 1,
				},
				hitWeights,
			),
		).toBeCloseTo(Mechanics.HIT_RATING_PER_HIT_PERCENT * 5);
	});

	it('values crit on every channel at the crit rating weight alone, though the spec weighs the crit percents', () => {
		const critWeights = Stats.fromMap(
			{ [Stat.StatCritRating]: 3 },
			{ [PseudoStat.PseudoStatMeleeCritPercent]: 7, [PseudoStat.PseudoStatRangedCritPercent]: 9, [PseudoStat.PseudoStatSpellCritPercent]: 11 },
		);
		expect(
			enchantEP(
				1220596,
				{
					[PseudoStat.PseudoStatMeleeCritPercent]: 1,
					[PseudoStat.PseudoStatRangedCritPercent]: 1,
					[PseudoStat.PseudoStatSpellCritPercent]: 1,
				},
				critWeights,
			),
		).toBeCloseTo(Mechanics.CRIT_RATING_PER_CRIT_PERCENT * 3);
	});

	it('values the spell crit beyond the part every channel holds at the spell crit weight', () => {
		const critWeights = Stats.fromMap({ [Stat.StatCritRating]: 3 }, { [PseudoStat.PseudoStatSpellCritPercent]: 11 });
		expect(
			enchantEP(
				1,
				{
					[PseudoStat.PseudoStatMeleeCritPercent]: 1,
					[PseudoStat.PseudoStatRangedCritPercent]: 1,
					[PseudoStat.PseudoStatSpellCritPercent]: 2,
				},
				critWeights,
			),
		).toBeCloseTo(Mechanics.CRIT_RATING_PER_CRIT_PERCENT * 3 + 11);
	});

	it('counts melee and spell hit once, as the largest, since one hit rating raises both', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHitPercent]: 1, [PseudoStat.PseudoStatSpellHitPercent]: 2 }, ratingWeights)).toBeCloseTo(
			2 * Mechanics.HIT_RATING_PER_HIT_PERCENT * 5,
		);
	});
});

describe('Player.computeEnchantEP on haste pseudo stats', () => {
	const hasteWeights = Stats.fromMap({ [Stat.StatHasteRating]: 2 }, {});

	it('values 34 Weapon Counterweight at the haste rating weight', () => {
		expect(enchantEP(34, { [PseudoStat.PseudoStatMeleeHastePercent]: 3 }, hasteWeights)).toBeCloseTo(3 * Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});

	it("counts 931 Enchant Gloves - Minor Haste's melee, ranged and cast speed once, as the one haste rating raises all three", () => {
		expect(
			enchantEP(
				931,
				{
					[PseudoStat.PseudoStatMeleeHastePercent]: 1,
					[PseudoStat.PseudoStatRangedHastePercent]: 1,
					[PseudoStat.PseudoStatSpellHastePercent]: 1,
				},
				hasteWeights,
			),
		).toBeCloseTo(Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});

	it('values ranged haste alone at the haste rating weight', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatRangedHastePercent]: 2 }, hasteWeights)).toBeCloseTo(2 * Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});
});

describe('Player.computeEnchantEP on haste pseudo stats for a spec weighing ranged haste', () => {
	const rangedHasteWeights = Stats.fromMap({ [Stat.StatHasteRating]: 2 }, { [PseudoStat.PseudoStatRangedHastePercent]: 7 });

	it('values 34 Weapon Counterweight at what the haste rating is worth past the ranged haste weight', () => {
		expect(enchantEP(34, { [PseudoStat.PseudoStatMeleeHastePercent]: 3, [PseudoStat.PseudoStatRangedHastePercent]: 0 }, rangedHasteWeights)).toBeCloseTo(
			3 * (Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2 - 7),
		);
	});

	it('values haste on every channel at the haste rating weight alone', () => {
		expect(
			enchantEP(
				931,
				{
					[PseudoStat.PseudoStatMeleeHastePercent]: 1,
					[PseudoStat.PseudoStatRangedHastePercent]: 1,
					[PseudoStat.PseudoStatSpellHastePercent]: 1,
				},
				rangedHasteWeights,
			),
		).toBeCloseTo(1 * Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});

	it('values melee and ranged haste with no cast speed, as Arcanum of Rapidity states them, at the haste rating weight: ranged haste at its own and melee haste at the rest', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHastePercent]: 1, [PseudoStat.PseudoStatRangedHastePercent]: 1 }, rangedHasteWeights)).toBeCloseTo(
			Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2 - 7 + 7,
		);
	});

	it('values ranged haste alone at the ranged haste weight', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatRangedHastePercent]: 2 }, rangedHasteWeights)).toBeCloseTo(2 * 7);
	});
});
