import { PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';

import type { ExporterDefinition } from './types';

const STAT_NAMES: Record<Stat, string> = {
	[Stat.StatStrength]: 'strength',
	[Stat.StatAgility]: 'agility',
	[Stat.StatStamina]: 'stamina',
	[Stat.StatIntellect]: 'intellect',
	[Stat.StatSpirit]: 'spirit',
	[Stat.StatMP5]: 'mp5',
	[Stat.StatAttackPower]: 'attackPower',
	[Stat.StatExpertiseRating]: 'expertise',
	[Stat.StatMana]: 'mana',
	[Stat.StatArmor]: 'armor',
	[Stat.StatRangedAttackPower]: 'attackPower',
	[Stat.StatDodgeRating]: 'dodge',
	[Stat.StatParryRating]: 'parry',
	[Stat.StatHealth]: 'health',
	[Stat.StatBonusArmor]: 'armorBonus',
	[Stat.StatArmorNoMods]: '',
	[Stat.StatHP5]: '',
	[Stat.StatHealingPower]: 'healing',
	[Stat.StatSpellDamage]: 'spellDamage',
	[Stat.StatArcaneDamage]: 'arcaneDamage',
	[Stat.StatFireDamage]: 'fireDamage',
	[Stat.StatFrostDamage]: 'frostDamage',
	[Stat.StatHolyDamage]: 'holyDamage',
	[Stat.StatNatureDamage]: 'natureDamage',
	[Stat.StatShadowDamage]: 'shadowDamage',
	[Stat.StatPhysicalDamage]: 'physicalDamage',
	[Stat.StatHitRating]: 'hit',
	[Stat.StatCritRating]: 'crit',
	[Stat.StatHasteRating]: 'haste',
	[Stat.StatSpellPiercing]: 'spellPen',
	[Stat.StatFeralAttackPower]: '',
	[Stat.StatArmorPenetration]: '',
	[Stat.StatDefenseRating]: 'defense',
	[Stat.StatBlockRating]: 'block',
	[Stat.StatBlockValue]: 'blockValueBonus',
	[Stat.StatArcaneResistance]: 'arcaneResist',
	[Stat.StatFireResistance]: 'fireResist',
	[Stat.StatFrostResistance]: 'frostResist',
	[Stat.StatNatureResistance]: 'natureResist',
	[Stat.StatShadowResistance]: 'shadowResist',
};

const PSEUDO_STAT_NAMES: Partial<Record<PseudoStat, string>> = {
	[PseudoStat.PseudoStatMainHandDps]: 'dps',
	[PseudoStat.PseudoStatRangedDps]: 'rangedDps',
};

const getName = (stat: UnitStat): string => (stat.isStat() ? STAT_NAMES[stat.getStat()] : (PSEUDO_STAT_NAMES[stat.getPseudoStat()] ?? ''));

// Sixty Upgrades takes the ratings as a percent chance, so a per-rating weight is scaled to per-percent.
// Its Forever mode has one hit and one crit that cover melee, ranged and spells, as the rating does.
const getScale = (stat: UnitStat): number => (stat.isStat() ? (stat.convertPercentToRating(1) ?? 1) : 1);

export const SIXTY_UPGRADES_EP_EXPORTER: ExporterDefinition = {
	title: i18n.t('export.sixty_upgrades_ep.title'),
	allowDownload: true,
	getData: host => {
		const player = host.player;
		const epValues = player.getEpWeights();
		const allUnitStats = UnitStat.getAll();

		const namesToWeights: Record<string, number> = {};
		allUnitStats.forEach(stat => {
			const statName = getName(stat);
			const weight = epValues.getUnitStat(stat) * getScale(stat);
			if (weight == 0 || statName == '') {
				return;
			}

			if (namesToWeights[statName]) {
				namesToWeights[statName] += weight;
			} else {
				namesToWeights[statName] = weight;
			}
		});

		return (
			`https://sixtyupgrades.com/forever/ep/import?name=${encodeURIComponent(`${player.getPlayerSpec().friendlyName} WoWSims Weights`)}` +
			Object.keys(namesToWeights)
				.map(statName => `&${statName}=${namesToWeights[statName].toFixed(3)}`)
				.join('')
		);
	},
};
