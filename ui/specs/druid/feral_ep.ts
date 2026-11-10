import { PseudoStat, Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';

// Hybrid 9/42/0, four races, 180s ±120s, and Adaptive rotations.
// Hit uses 7–8% and expertise uses 5.5–6.5%; rating weights are per point.
// Cat is DPS per stat; Bear adds calibrated DTPS and TMI reduction value.
export const CatEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.49630486278728336,
		[Stat.StatAgility]: 0.5159261908819341,
		[Stat.StatAttackPower]: 0.20467203716254861,
		[Stat.StatFeralAttackPower]: 0.20466522318514968,
		[Stat.StatHitRating]: 0.4078999148291679,
		[Stat.StatCritRating]: 0.38018320500522035,
		[Stat.StatHasteRating]: 0.294614735295168,
		[Stat.StatArmorPenetration]: 0.06660080481481513,
		[Stat.StatExpertiseRating]: 0.3856003085648115,
		[Stat.StatPhysicalDamage]: 1.9125857214075173,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.0463121920819978,
	},
);

export const BearEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.46935464459097603,
		[Stat.StatAgility]: 0.4497333164963253,
		[Stat.StatStamina]: 0.9656595073782596,
		[Stat.StatAttackPower]: 0.21308564484691025,
		[Stat.StatFeralAttackPower]: 0.21308564484691025,
		[Stat.StatHitRating]: 0.7657291219047078,
		[Stat.StatCritRating]: 0.340533171530225,
		[Stat.StatHasteRating]: 0.24298121005700268,
		[Stat.StatArmorPenetration]: 0.0802239756770389,
		[Stat.StatExpertiseRating]: 1.9612700593371994,
		[Stat.StatDefenseRating]: 0.9529964781955986,
		[Stat.StatDodgeRating]: 0.13653626021319415,
		[Stat.StatArmor]: 0.20349776153972676,
		[Stat.StatBonusArmor]: 0.20391902537209813,
		[Stat.StatHealth]: 0.07228807547682363,
		[Stat.StatPhysicalDamage]: 1.3497636434965008,
		[Stat.StatArmorNoMods]: 0.04116475035877257,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.988082747976817,
	},
);

export const HybridEp = CatEp.add(BearEp).scale(0.5);
