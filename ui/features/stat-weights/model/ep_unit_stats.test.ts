import { PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { EP_UNIT_STATS, isEpStat, visibleEpUnitStats } from './ep_unit_stats';

const names = (stats: UnitStat[]) => stats.map(stat => stat.getKey());

describe('EP_UNIT_STATS', () => {
	it('keeps every Stat', () => {
		const allStats = UnitStat.getAll().filter(stat => stat.isStat());
		expect(EP_UNIT_STATS.filter(stat => stat.isStat()).length).toBe(allStats.length);
		expect(allStats.length).toBeGreaterThan(0);
	});

	it('drops a pseudo-stat that is not on the list', () => {
		expect(EP_UNIT_STATS.some(stat => stat.isPseudoStat() && stat.getPseudoStat() === PseudoStat.PseudoStatMeleeSpeedMultiplier)).toBe(false);
	});
});

describe('isEpStat', () => {
	const statSet = { epStats: [Stat.StatStrength], epPseudoStats: [PseudoStat.PseudoStatMainHandDps] };

	it('answers from epStats for a Stat', () => {
		expect(isEpStat(UnitStat.fromStat(Stat.StatStrength), statSet)).toBe(true);
		expect(isEpStat(UnitStat.fromStat(Stat.StatAgility), statSet)).toBe(false);
	});

	it('answers from epPseudoStats for a PseudoStat', () => {
		expect(isEpStat(UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps), statSet)).toBe(true);
		expect(isEpStat(UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandDps), statSet)).toBe(false);
	});

	it('does not answer a PseudoStat out of epStats, though both are 0', () => {
		expect(isEpStat(UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps), { epStats: [Stat.StatStrength], epPseudoStats: [] })).toBe(false);
	});
});

describe('visibleEpUnitStats', () => {
	const statSet = { epStats: [Stat.StatStrength, Stat.StatAgility], epPseudoStats: [PseudoStat.PseudoStatSpellHitPercent] };

	it('shows only the spec stats when showAllStats is off', () => {
		expect(names(visibleEpUnitStats(statSet, false))).toEqual([
			UnitStat.fromStat(Stat.StatStrength).getKey(),
			UnitStat.fromStat(Stat.StatAgility).getKey(),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent).getKey(),
		]);
	});

	it('adds every other Stat when showAllStats is on', () => {
		const shown = visibleEpUnitStats(statSet, true);
		expect(shown.filter(stat => stat.isStat()).length).toBe(EP_UNIT_STATS.filter(stat => stat.isStat()).length);
		expect(names(shown)).toContain(UnitStat.fromStat(Stat.StatIntellect).getKey());
	});

	// Only the `Stat` clause is gated on `showAllStats`; an off-spec pseudo-stat is hidden
	// unconditionally, because its row carries an editable Current EP cell.
	it('keeps an off-spec pseudo-stat hidden when showAllStats is on', () => {
		expect(names(visibleEpUnitStats(statSet, true))).toEqual(
			expect.not.arrayContaining([
				UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps).getKey(),
				UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent).getKey(),
			]),
		);
	});

	it('keeps a pseudo-stat the spec does weight when showAllStats is on', () => {
		expect(names(visibleEpUnitStats(statSet, true))).toContain(UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent).getKey());
	});

	it('keeps EP_UNIT_STATS order', () => {
		const shown = names(visibleEpUnitStats(statSet, true));
		expect(shown).toEqual(names(EP_UNIT_STATS).filter(key => shown.includes(key)));
	});
});
