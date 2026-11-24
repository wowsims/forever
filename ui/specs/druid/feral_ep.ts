import { PseudoStat, Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';

// Four races, 180s ±5s; each role uses its default gear.
// Hit uses 7–8% and expertise uses 5.5–6.5%; rating weights are per point.
// Cat measures DPS; Bear includes calibrated TMI reduction value.
export const CatEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.682284654993632,
		[Stat.StatAgility]: 0.80021045768010879,
		[Stat.StatIntellect]: 0.002942893913774882,
		[Stat.StatSpirit]: 0.0013883160002777773,
		[Stat.StatAttackPower]: 0.27984748897699391,
		[Stat.StatHitRating]: 0.63634488206942008,
		[Stat.StatCritRating]: 0.6478483333198749,
		[Stat.StatHasteRating]: 0.23789134061447764,
		[Stat.StatArmorPenetration]: 0.10570076484754429,
		[Stat.StatExpertiseRating]: 0.6100170601767172,
		[Stat.StatMP5]: 0.0032799379035417116,
		[Stat.StatPhysicalDamage]: 1.9904491868544347,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.453697898696126,
		[PseudoStat.PseudoStatFeralCombatSkill]: 5.2397937314911784,
	},
);
export const BearEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.4312040340426293,
		[Stat.StatAgility]: 0.40063693988113031,
		[Stat.StatStamina]: 0.69384805000851679,
		[Stat.StatIntellect]: 0.0016385788832416568,
		[Stat.StatAttackPower]: 0.19610089665063155,
		[Stat.StatHitRating]: 0.80926081266515482,
		[Stat.StatCritRating]: 0.3442018159426658,
		[Stat.StatHasteRating]: 0.13392530158969285,
		[Stat.StatArmorPenetration]: 0.077610224316905826,
		[Stat.StatExpertiseRating]: 1.8850939467561696,
		[Stat.StatDefenseRating]: 0.73234657773690359,
		[Stat.StatDodgeRating]: 0.12542113807352423,
		[Stat.StatArmor]: 0.097712083396817001,
		[Stat.StatBonusArmor]: 0.097712083396817265,
		[Stat.StatHealth]: 0.053677070228118295,
		[Stat.StatPhysicalDamage]: 1.2030266536479386,
		[Stat.StatArmorNoMods]: 0.021330339335240509,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.7454125531088405,
		[PseudoStat.PseudoStatFeralCombatSkill]: 1.8450885594460389,
	},
);

// Fitted separately from the Hybrid Cat and Bear gear, with balanced primary stats.
export const HybridEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.49217856168012131,
		[Stat.StatAgility]: 0.49217856168012131,
		[Stat.StatStamina]: 0.49217856168012131,
		[Stat.StatIntellect]: 0.01580608050429377,
		[Stat.StatSpirit]: 0.0064393065859538016,
		[Stat.StatAttackPower]: 0.21504709359060581,
		[Stat.StatHitRating]: 0.77831741648066799,
		[Stat.StatCritRating]: 0.41319671802570568,
		[Stat.StatHasteRating]: 0.18024681397604059,
		[Stat.StatArmorPenetration]: 0.081257446380030379,
		[Stat.StatExpertiseRating]: 1.5101657570047333,
		[Stat.StatDefenseRating]: 0.49877530016712757,
		[Stat.StatDodgeRating]: 0.065089577077526786,
		[Stat.StatArmor]: 0.080693940546502516,
		[Stat.StatBonusArmor]: 0.080693940546501614,
		[Stat.StatHealth]: 0.037474365264994307,
		[Stat.StatMP5]: 0.020473044756836145,
		[Stat.StatPhysicalDamage]: 1.4457866464211619,
		[Stat.StatArmorNoMods]: 0.017722496655995984,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.5924584652247376,
		[PseudoStat.PseudoStatFeralCombatSkill]: 3.0625745419758879,
	},
);
