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
	it('builds displayStatOrder at module scope over the 39-stat shape', () => {
		expect(UnitStat.createDisplayStatArray([Stat.StatHealth], []).length).toBe(1);
		expect(new Stats().asProtoArray().length).toBe(39);
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

	// The inverse of the parentStat case below: convertRatingToPercent has no branch for
	// ReducedCritTakenPercent either, so it returns null. convertEpToRatingScale declares
	// `number` and asserted that null away with `!`, handing callers a null that only blew up
	// once something did arithmetic on it - the Prot Paladin soft-cap tooltip called .toFixed()
	// on it and took the sidebar down on hover.
	it('falls back to the raw EP value when a stat has no percent conversion', () => {
		const unitStat = UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent);
		expect(unitStat.convertRatingToPercent(1)).toBeNull();
		expect(unitStat.convertEpToRatingScale(12.5)).toBe(12.5);
		expect(() => unitStat.convertEpToRatingScale(0).toFixed(2)).not.toThrow();
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
