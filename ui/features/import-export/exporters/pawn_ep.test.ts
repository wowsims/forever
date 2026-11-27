import { PseudoStat, Stat } from '@generated/proto/common';
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

	it.each([
		[Stat.StatArcaneDamage, 'ArcaneSpellDamage'],
		[Stat.StatFireDamage, 'FireSpellDamage'],
		[Stat.StatFrostDamage, 'FrostSpellDamage'],
		[Stat.StatHolyDamage, 'HolySpellDamage'],
		[Stat.StatNatureDamage, 'NatureSpellDamage'],
		[Stat.StatShadowDamage, 'ShadowSpellDamage'],
		[Stat.StatSpellPiercing, 'SpellPenetration'],
		[Stat.StatHealingPower, 'Healing'],
		[Stat.StatFeralAttackPower, 'FeralAp'],
		[Stat.StatArmorPenetration, 'ArmorPenetration'],
		[Stat.StatDefenseRating, 'DefenseRating'],
		[Stat.StatBlockValue, 'BlockValue'],
		[Stat.StatShadowResistance, 'ShadowResist'],
	])('writes stat %i as Pawn key %s', (stat, name) => {
		expect(exportFor(new Stats().withStat(stat, 0.58))).toContain(`${name}=0.580`);
	});

	it.each([
		[Stat.StatHitRating, 'HitRating=5.000'],
		[Stat.StatCritRating, 'CritRating=7.000'],
		[Stat.StatHasteRating, 'HasteRating=5.000'],
		[Stat.StatExpertiseRating, 'ExpertiseRating=5.000'],
		[Stat.StatDodgeRating, 'DodgeRating=6.000'],
		[Stat.StatParryRating, 'ParryRating=7.500'],
		[Stat.StatBlockRating, 'BlockRating=2.500'],
	])('writes stat %i per percentage point, as the tooltip shows it', (stat, entry) => {
		expect(exportFor(new Stats().withStat(stat, 0.5))).toContain(entry);
	});

	it('adds ranged attack power into Ap and writes it alone as Rap', () => {
		const tag = exportFor(new Stats().withStat(Stat.StatAttackPower, 1).withStat(Stat.StatRangedAttackPower, 0.5));
		expect(tag).toContain('Ap=1.500');
		expect(tag).toContain('Rap=0.500');
	});

	it('writes the off-hand weight as its difference from MeleeDps', () => {
		const tag = exportFor(new Stats().withPseudoStat(PseudoStat.PseudoStatMainHandDps, 2).withPseudoStat(PseudoStat.PseudoStatOffHandDps, 1.2));
		expect(tag).toContain('MeleeDps=2.000');
		expect(tag).toContain('OffHandDps=-0.800');
	});

	it('writes main-hand and ranged weapon DPS with no OffHandDps when the off hand has no weight', () => {
		const tag = exportFor(new Stats().withPseudoStat(PseudoStat.PseudoStatMainHandDps, 2).withPseudoStat(PseudoStat.PseudoStatRangedDps, 1.5));
		expect(tag).toContain('MeleeDps=2.000');
		expect(tag).toContain('RangedDps=1.500');
		expect(tag).not.toContain('OffHandDps');
	});

	it('leaves out stats Pawn has no key for', () => {
		const tag = exportFor(
			new Stats()
				.withStat(Stat.StatSpellDamage, 1)
				.withStat(Stat.StatBonusArmor, 1)
				.withStat(Stat.StatPhysicalDamage, 1)
				.withStat(Stat.StatHealth, 1)
				.withStat(Stat.StatMana, 1),
		);
		expect(tag).toBe('( Pawn: v1: "Warlock WoWSims Weights": Class=Warlock,SpellDamage=1.000 )');
	});
});
