import { PseudoStat, Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';

// Hybrid reference build, four races, 180s ±120s, and Adaptive rotations.
// Hit uses 7–8% and expertise uses 5.5–6.5%; rating weights are per point.
// Cat is DPS per stat; Bear includes calibrated TMI reduction value.
// Feral Combat Skill uses the next point from base skill at 7% hit.
export const CatEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.5942180657086884,
		[Stat.StatAgility]: 0.6546615973907315,
		[Stat.StatAttackPower]: 0.24378941350988748,
		[Stat.StatFeralAttackPower]: 0.24313926679561954,
		[Stat.StatHitRating]: 0.5211576370723761,
		[Stat.StatCritRating]: 0.5057159311903373,
		[Stat.StatHasteRating]: 0.39306887257711326,
		[Stat.StatArmorPenetration]: 0.07421506952196227,
		[Stat.StatExpertiseRating]: 0.49040516114181354,
		[Stat.StatPhysicalDamage]: 1.8630615922745424,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.142010195278739,
		[PseudoStat.PseudoStatFeralCombatSkill]: 5.326180320036634,
	},
);

export const BearEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.4860504938346992,
		[Stat.StatAgility]: 0.4334793978220709,
		[Stat.StatStamina]: 1.0028749432917183,
		[Stat.StatAttackPower]: 0.22093229824112756,
		[Stat.StatFeralAttackPower]: 0.22093229824112756,
		[Stat.StatHitRating]: 0.6879449708049289,
		[Stat.StatCritRating]: 0.36326222397559405,
		[Stat.StatHasteRating]: 0.3581301339921136,
		[Stat.StatArmorPenetration]: 0.08262889794633035,
		[Stat.StatExpertiseRating]: 1.775973949024296,
		[Stat.StatDefenseRating]: 0.8758307060275772,
		[Stat.StatDodgeRating]: 0.10739498989242388,
		[Stat.StatArmor]: 0.1743755390508764,
		[Stat.StatBonusArmor]: 0.17437553905087635,
		[Stat.StatHealth]: 0.07717112351322317,
		[Stat.StatPhysicalDamage]: 1.3666857276700761,
		[Stat.StatArmorNoMods]: 0.03691185978147275,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 3.0930521753757847,
		[PseudoStat.PseudoStatFeralCombatSkill]: 2.5478378284939915,
	},
);

// This role split balances Hybrid Strength, Agility and Stamina in the reference setup.
const hybridCatFraction = 0.4651707224542499;
export const HybridEp = CatEp.scale(hybridCatFraction).add(BearEp.scale(1 - hybridCatFraction));
