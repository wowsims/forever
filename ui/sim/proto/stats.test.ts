import { PseudoStat, Stat } from '@generated/proto/common';
import { describe, expect, it, vi } from 'vitest';

// `stats.ts` only needs localization for display names, and `@i18n/localization` drags the whole
// i18next bootstrap in with it. Stubbing it keeps this file about the stat model.
vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import * as Mechanics from '../constants/mechanics';
import { Stats, UnitStat } from './stats';

describe('UnitStat', () => {
	// `displayStatOrder` is built at module scope, so a throw inside getRootStat there takes every
	// transitive importer down with it.
	it('builds displayStatOrder at module scope over the 40-stat shape', () => {
		expect(UnitStat.createDisplayStatArray([Stat.StatHealth], []).length).toBe(1);
		expect(new Stats().asProtoArray().length).toBe(40);
	});

	it('shows HP5 directly after Health', () => {
		const order = UnitStat.createDisplayStatArray([Stat.StatHP5, Stat.StatHealth], []);
		expect(order.map(stat => stat.getStat())).toEqual([Stat.StatHealth, Stat.StatHP5]);
	});

	it('leaves the six school hit pseudo-stats without a root', () => {
		const schools = [
			PseudoStat.PseudoStatSchoolHitPercentArcane,
			PseudoStat.PseudoStatSchoolHitPercentFire,
			PseudoStat.PseudoStatSchoolHitPercentFrost,
			PseudoStat.PseudoStatSchoolHitPercentHoly,
			PseudoStat.PseudoStatSchoolHitPercentNature,
			PseudoStat.PseudoStatSchoolHitPercentShadow,
		];
		for (const school of schools) {
			expect(UnitStat.fromPseudoStat(school).hasRootStat()).toBe(false);
		}
	});

	it('roots ExpertisePercent at ExpertiseRating and converts at the gametable ratio, with no quarter-point floor', () => {
		const percent = UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent);
		expect(percent.getRootStat()).toBe(Stat.StatExpertiseRating);
		expect(UnitStat.fromStat(Stat.StatExpertiseRating).convertRatingToPercent(12)).toBeCloseTo(1.2);
		expect(percent.convertPercentToRating(6.5)).toBeCloseTo(6.5 * Mechanics.EXPERTISE_RATING_PER_EXPERTISE_PERCENT);
	});

	it('roots every melee, off-hand, ranged and spell percent at the one rating of its kind', () => {
		const roots: Array<[PseudoStat, Stat]> = [
			[PseudoStat.PseudoStatMeleeHitPercent, Stat.StatHitRating],
			[PseudoStat.PseudoStatOffHandHitPercent, Stat.StatHitRating],
			[PseudoStat.PseudoStatRangedHitPercent, Stat.StatHitRating],
			[PseudoStat.PseudoStatSpellHitPercent, Stat.StatHitRating],
			[PseudoStat.PseudoStatMeleeCritPercent, Stat.StatCritRating],
			[PseudoStat.PseudoStatOffHandCritPercent, Stat.StatCritRating],
			[PseudoStat.PseudoStatRangedCritPercent, Stat.StatCritRating],
			[PseudoStat.PseudoStatSpellCritPercent, Stat.StatCritRating],
			[PseudoStat.PseudoStatMeleeHastePercent, Stat.StatHasteRating],
			[PseudoStat.PseudoStatRangedHastePercent, Stat.StatHasteRating],
			[PseudoStat.PseudoStatSpellHastePercent, Stat.StatHasteRating],
		];
		for (const [pseudoStat, rating] of roots) {
			expect(UnitStat.fromPseudoStat(pseudoStat).getRootStat()).toBe(rating);
		}
	});

	it('leaves the per-attack rating pseudo-stats without a root', () => {
		const ratings = [
			PseudoStat.PseudoStatMainHandHitRating,
			PseudoStat.PseudoStatMainHandCritRating,
			PseudoStat.PseudoStatOffHandHitRating,
			PseudoStat.PseudoStatOffHandCritRating,
			PseudoStat.PseudoStatRangedHitRating,
			PseudoStat.PseudoStatRangedCritRating,
		];
		for (const rating of ratings) {
			expect(UnitStat.fromPseudoStat(rating).hasRootStat()).toBe(false);
			expect(UnitStat.fromPseudoStat(rating).convertPercentToRating(10)).toBeNull();
		}
	});

	it('converts each rating and its percents through the one constant', () => {
		expect(UnitStat.fromStat(Stat.StatHitRating).convertRatingToPercent(Mechanics.HIT_RATING_PER_HIT_PERCENT)).toBeCloseTo(1);
		expect(UnitStat.fromStat(Stat.StatCritRating).convertRatingToPercent(Mechanics.CRIT_RATING_PER_CRIT_PERCENT)).toBeCloseTo(1);
		expect(UnitStat.fromStat(Stat.StatHasteRating).convertRatingToPercent(Mechanics.HASTE_RATING_PER_HASTE_PERCENT)).toBeCloseTo(1);
		expect(UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent).convertPercentToRating(1)).toBeCloseTo(Mechanics.HIT_RATING_PER_HIT_PERCENT);
		expect(UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire).convertPercentToRating(1)).toBeCloseTo(Mechanics.HIT_RATING_PER_HIT_PERCENT);
		expect(UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent).convertPercentToRating(1)).toBeCloseTo(Mechanics.CRIT_RATING_PER_CRIT_PERCENT);
	});

	it('counts a rating, or a percent rooted at one, as a percent rating, but not defense', () => {
		expect(UnitStat.fromStat(Stat.StatDodgeRating).isPercentRating()).toBe(true);
		expect(UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent).isPercentRating()).toBe(true);
		expect(UnitStat.fromStat(Stat.StatDefenseRating).isPercentRating()).toBe(false);
		expect(UnitStat.fromStat(Stat.StatDefenseRating).convertPercentToRating(1)).toBeCloseTo(Mechanics.DEFENSE_RATING_PER_DEFENSE_LEVEL);
		expect(UnitStat.fromStat(Stat.StatStrength).isPercentRating()).toBe(false);
	});

	// The soft-cap tooltip calls .toFixed() on the result, so a stat with no rating still gets a number.
	it('falls back to the raw EP value when a stat has no percent conversion', () => {
		const unitStat = UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent);
		expect(unitStat.convertEpToRatingScale(12.5)).toBe(12.5);
		expect(() => unitStat.convertEpToRatingScale(0).toFixed(2)).not.toThrow();
	});

	it.each(
		(
			[
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent), 10],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent), 14],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHastePercent), 10],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent), 10],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent), 12],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatParryPercent), 15],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent), 5],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire), 10],
				[UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps), 1],
				[UnitStat.fromStat(Stat.StatCritRating), 1],
			] as const
		).map(([unitStat, scale]) => [unitStat.getKey(), unitStat, scale] as const),
	)('epRatingScale of %s is %i', (_key, unitStat, scale) => {
		expect(unitStat.epRatingScale()).toBe(scale);
	});

	it('needs a parentStat to turn ReducedCritTakenPercent back into a rating', () => {
		const unitStat = UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent);
		expect(unitStat.convertPercentToRating(1)).toBeNull();
		expect(unitStat.convertPercentToRating(1, Stat.StatDefenseRating)).toBeCloseTo(
			Mechanics.DEFENSE_RATING_PER_DEFENSE_LEVEL / Mechanics.MISS_DODGE_PARRY_BLOCK_CRIT_CHANCE_PER_DEFENSE,
		);
	});
});

describe('Stats.computeGapToCap', () => {
	it('divides a haste gap by the matching speed multiplier', () => {
		const stats = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHastePercent, 10).withPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier, 2);
		expect(stats.computeGapToCap(UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHastePercent), 20)).toBeCloseTo(5);
	});

	it('never reports an exactly-zero gap', () => {
		const stats = new Stats().withStat(Stat.StatHitRating, 100);
		expect(stats.computeGapToCap(UnitStat.fromStat(Stat.StatHitRating), 100)).toBeGreaterThan(0);
	});
});
