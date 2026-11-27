import { PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';

import type { ExporterDefinition } from './types';

// Pawn's scale keys. An empty name is a stat Pawn has no key for: Pawn reads armor only as Armor and drops Health and Mana on import.
const STAT_NAMES: Record<Stat, string> = {
	[Stat.StatStrength]: 'Strength',
	[Stat.StatAgility]: 'Agility',
	[Stat.StatStamina]: 'Stamina',
	[Stat.StatIntellect]: 'Intellect',
	[Stat.StatSpirit]: 'Spirit',
	[Stat.StatMP5]: 'Mp5',
	[Stat.StatAttackPower]: 'Ap',
	[Stat.StatExpertiseRating]: 'ExpertiseRating',
	[Stat.StatMana]: '',
	[Stat.StatArmor]: 'Armor',
	[Stat.StatRangedAttackPower]: 'Rap',
	[Stat.StatDodgeRating]: 'DodgeRating',
	[Stat.StatParryRating]: 'ParryRating',
	[Stat.StatHealth]: '',
	[Stat.StatBonusArmor]: '',
	[Stat.StatArmorNoMods]: '',
	[Stat.StatHealingPower]: 'Healing',
	[Stat.StatSpellDamage]: 'SpellDamage',
	[Stat.StatArcaneDamage]: 'ArcaneSpellDamage',
	[Stat.StatFireDamage]: 'FireSpellDamage',
	[Stat.StatFrostDamage]: 'FrostSpellDamage',
	[Stat.StatHolyDamage]: 'HolySpellDamage',
	[Stat.StatNatureDamage]: 'NatureSpellDamage',
	[Stat.StatShadowDamage]: 'ShadowSpellDamage',
	[Stat.StatPhysicalDamage]: '',
	[Stat.StatHitRating]: 'HitRating',
	[Stat.StatCritRating]: 'CritRating',
	[Stat.StatHasteRating]: 'HasteRating',
	[Stat.StatSpellPiercing]: 'SpellPenetration',
	[Stat.StatFeralAttackPower]: 'FeralAp',
	[Stat.StatArmorPenetration]: 'ArmorPenetration',
	[Stat.StatDefenseRating]: 'DefenseRating',
	[Stat.StatBlockRating]: 'BlockRating',
	[Stat.StatBlockValue]: 'BlockValue',
	[Stat.StatArcaneResistance]: 'ArcaneResist',
	[Stat.StatFireResistance]: 'FireResist',
	[Stat.StatFrostResistance]: 'FrostResist',
	[Stat.StatNatureResistance]: 'NatureResist',
	[Stat.StatShadowResistance]: 'ShadowResist',
};

// Pawn has no key for weapon skill, Feral Combat skill, school hit or percent ranged hit, so those weights stay out.
const PSEUDO_STAT_NAMES: Partial<Record<PseudoStat, string>> = {
	[PseudoStat.PseudoStatMainHandDps]: 'MeleeDps',
	[PseudoStat.PseudoStatOffHandDps]: 'OffHandDps',
	[PseudoStat.PseudoStatRangedDps]: 'RangedDps',
};

const getName = (stat: UnitStat): string => (stat.isStat() ? STAT_NAMES[stat.getStat()] : (PSEUDO_STAT_NAMES[stat.getPseudoStat()] ?? ''));

export const PAWN_EP_EXPORTER: ExporterDefinition = {
	title: i18n.t('export.pawn_ep.title'),
	allowDownload: true,
	getData: host => {
		const player = host.player;
		const epValues = player.getEpWeights();
		const allUnitStats = UnitStat.getAll();

		const namesToWeights: Record<string, number> = {};
		allUnitStats.forEach(stat => {
			const statName = getName(stat);
			if (statName == '') {
				return;
			}
			// Pawn stores the number a tooltip line shows, and the game shows hit, crit, haste, expertise, dodge, parry and block as a percent ("Improves your chance to hit by 1.8%"), so those weights go out per percentage point.
			const weight = stat.storedEpToShown(epValues.getUnitStat(stat));
			if (weight != 0) {
				namesToWeights[statName] = weight;
			}
		});
		// Pawn scores an item's "+X Attack Power" line as Ap alone, and that item grants as much ranged attack power, so the ranged weight goes into Ap as well as Rap.
		if (namesToWeights.Rap) {
			namesToWeights.Ap = (namesToWeights.Ap ?? 0) + namesToWeights.Rap;
		}
		// Pawn scores every melee weapon at MeleeDps and an Off Hand weapon at OffHandDps on top, so OffHandDps holds what the off-hand weight differs by.
		if (namesToWeights.OffHandDps) {
			namesToWeights.OffHandDps -= namesToWeights.MeleeDps ?? 0;
		}

		return (
			`( Pawn: v1: "${player.getPlayerSpec().friendlyName} WoWSims Weights": Class=${player.getPlayerClass().friendlyName},` +
			Object.keys(namesToWeights)
				.map(statName => `${statName}=${namesToWeights[statName].toFixed(3)}`)
				.join(',') +
			' )'
		);
	},
};
