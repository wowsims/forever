import { PseudoStat, Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';

// Four races, 180s ±5s; both roles use their Hybrid gear and talents.
// Hit uses 7–8% and expertise uses 5.5–6.5%; rating weights are per point.
// Bear adds 0.6809523132780807 times TMI reduction and 0.021943136856483227 times physical EHP in DPS-equivalent units.
export const CatEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.59774328235395635,
		[Stat.StatAgility]: 0.67739837324518826,
		[Stat.StatIntellect]: 0.051207546726651565,
		[Stat.StatSpirit]: 0.021392051970248963,
		[Stat.StatAttackPower]: 0.24484098279105812,
		[Stat.StatHitRating]: 0.57506783459257915,
		[Stat.StatCritRating]: 0.53029331437688609,
		[Stat.StatHasteRating]: 0.20705187119677856,
		[Stat.StatArmorPenetration]: 0.082569483793804868,
		[Stat.StatExpertiseRating]: 0.54983836580173662,
		[Stat.StatMP5]: 0.07172401996399011,
		[Stat.StatPhysicalDamage]: 1.8839876385731529,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.1684497515228793,
		[PseudoStat.PseudoStatFeralCombatSkill]: 4.8381787077522524,
	},
);
export const BearEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.42180737046904571,
		[Stat.StatAgility]: 0.3421522795778138,
		[Stat.StatStamina]: 1.0195506528230021,
		[Stat.StatIntellect]: 0.00078995084079949685,
		[Stat.StatAttackPower]: 0.19174686753624698,
		[Stat.StatHitRating]: 0.76595503128327558,
		[Stat.StatCritRating]: 0.32686593029331917,
		[Stat.StatHasteRating]: 0.28299290102469654,
		[Stat.StatArmorPenetration]: 0.076942239137796131,
		[Stat.StatExpertiseRating]: 1.4908756993816326,
		[Stat.StatDefenseRating]: 0.44509948787822168,
		[Stat.StatDodgeRating]: -0.0061422686303552639,
		[Stat.StatArmor]: 0.20223136852452989,
		[Stat.StatBonusArmor]: 0.20223136852452961,
		[Stat.StatHealth]: 0.078520181776663769,
		[Stat.StatPhysicalDamage]: 1.1856849521257626,
		[Stat.StatArmorNoMods]: 0.043895084366266171,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.6844561455074536,
		[PseudoStat.PseudoStatFeralCombatSkill]: 2.1799893988195431,
	},
);

// Hybrid averages the two role components on their shared scale.
export const HybridEp = CatEp.scale(0.5).add(BearEp.scale(0.5));
