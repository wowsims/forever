import { Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { PAWN_EP_EXPORTER } from './pawn_ep';

const exportFor = (weights: Stats) =>
	PAWN_EP_EXPORTER.getData(
		{
			player: {
				getEpWeights: () => weights,
				getPlayerSpec: () => ({ friendlyName: 'Warlock' }),
				getPlayerClass: () => ({ friendlyName: 'Warlock' }),
			},
		} as any,
		{} as any,
	);

describe('PAWN_EP_EXPORTER', () => {
	it('writes a Pawn scale tag with the class and each weight', () => {
		expect(exportFor(new Stats().withStat(Stat.StatSpellDamage, 1).withStat(Stat.StatIntellect, 0.158))).toBe(
			'( Pawn: v1: "Warlock WoWSims Weights": Class=Warlock,Intellect=0.158,SpellDamage=1.000 )',
		);
	});

	it('writes school spell damage and spell piercing under Pawn names', () => {
		const tag = exportFor(
			new Stats()
				.withStat(Stat.StatFireDamage, 0.18)
				.withStat(Stat.StatShadowDamage, 0.82)
				.withStat(Stat.StatArcaneDamage, 0.1)
				.withStat(Stat.StatFrostDamage, 0.2)
				.withStat(Stat.StatHolyDamage, 0.3)
				.withStat(Stat.StatNatureDamage, 0.4)
				.withStat(Stat.StatSpellPiercing, 0.58),
		);
		expect(tag).toContain('FireSpellDamage=0.180');
		expect(tag).toContain('ShadowSpellDamage=0.820');
		expect(tag).toContain('ArcaneSpellDamage=0.100');
		expect(tag).toContain('FrostSpellDamage=0.200');
		expect(tag).toContain('HolySpellDamage=0.300');
		expect(tag).toContain('NatureSpellDamage=0.400');
		expect(tag).toContain('SpellPenetration=0.580');
	});

	it('keeps ranged attack power apart from attack power', () => {
		const tag = exportFor(new Stats().withStat(Stat.StatAttackPower, 1).withStat(Stat.StatRangedAttackPower, 0.5));
		expect(tag).toContain('Ap=1.000');
		expect(tag).toContain('Rap=0.500');
	});

	it('writes the healing, feral, tank and resistance stats under Pawn names', () => {
		const tag = exportFor(
			new Stats()
				.withStat(Stat.StatHealingPower, 1)
				.withStat(Stat.StatFeralAttackPower, 1)
				.withStat(Stat.StatArmorPenetration, 1)
				.withStat(Stat.StatDefenseRating, 1)
				.withStat(Stat.StatBlockRating, 1)
				.withStat(Stat.StatBlockValue, 1)
				.withStat(Stat.StatShadowResistance, 1),
		);
		for (const name of ['Healing', 'FeralAp', 'ArmorPenetration', 'DefenseRating', 'BlockRating', 'BlockValue', 'ShadowResist']) {
			expect(tag).toContain(`${name}=1.000`);
		}
	});

	it('leaves out stats Pawn has no key for', () => {
		const tag = exportFor(new Stats().withStat(Stat.StatSpellDamage, 1).withStat(Stat.StatBonusArmor, 1).withStat(Stat.StatPhysicalDamage, 1));
		expect(tag).toBe('( Pawn: v1: "Warlock WoWSims Weights": Class=Warlock,SpellDamage=1.000 )');
	});
});
