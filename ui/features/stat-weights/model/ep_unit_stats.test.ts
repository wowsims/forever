import { PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { EP_UNIT_STATS, epFilterScope, isEpStat } from './ep_unit_stats';

const names = (stats: ReadonlyArray<UnitStat>) => stats.map(stat => stat.getKey());

describe('EP_UNIT_STATS', () => {
	it('keeps every Stat except Feral Attack Power', () => {
		const allStats = UnitStat.getAll().filter(stat => stat.isStat() && stat.getStat() !== Stat.StatFeralAttackPower);
		expect(EP_UNIT_STATS.filter(stat => stat.isStat()).length).toBe(allStats.length);
		expect(allStats.length).toBeGreaterThan(0);
		expect(EP_UNIT_STATS.some(stat => stat.equalsStat(Stat.StatFeralAttackPower))).toBe(false);
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

describe('epFilterScope', () => {
	const statSet = { epStats: [Stat.StatAgility, Stat.StatStrength], epPseudoStats: [PseudoStat.PseudoStatSpellHitPercent] };

	it('defaults to the spec stats, in EP_UNIT_STATS order', () => {
		expect(names(epFilterScope(statSet).defaults)).toEqual([
			UnitStat.fromStat(Stat.StatStrength).getKey(),
			UnitStat.fromStat(Stat.StatAgility).getKey(),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent).getKey(),
		]);
	});

	it('offers every supported Stat', () => {
		const { universe } = epFilterScope(statSet);
		expect(universe.filter(stat => stat.isStat()).length).toBe(EP_UNIT_STATS.filter(stat => stat.isStat()).length);
		expect(names(universe)).toContain(UnitStat.fromStat(Stat.StatIntellect).getKey());
	});

	it('offers only the pseudo-stats the spec weights', () => {
		const pseudoStats = epFilterScope(statSet).universe.filter(stat => stat.isPseudoStat());
		expect(names(pseudoStats)).toEqual([UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent).getKey()]);
	});

	it('keeps EP_UNIT_STATS order', () => {
		const universe = names(epFilterScope(statSet).universe);
		expect(universe).toEqual(names(EP_UNIT_STATS).filter(key => universe.includes(key)));
	});
});

it('exposes Feral Combat Skill only for specs that weigh it', () => {
	const skill = UnitStat.fromPseudoStat(PseudoStat.PseudoStatFeralCombatSkill);
	expect(EP_UNIT_STATS.some(stat => stat.equals(skill))).toBe(true);
	expect(epFilterScope({ epStats: [], epPseudoStats: [PseudoStat.PseudoStatFeralCombatSkill] }).defaults.some(stat => stat.equals(skill))).toBe(true);
	expect(epFilterScope({ epStats: [], epPseudoStats: [] }).universe.some(stat => stat.equals(skill))).toBe(false);
});
