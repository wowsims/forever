import { PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Stats, UnitStat } from '@sim/proto/stats';
import { TONE_TEXT } from '@ui-kit/utils/colors';

const SCHOOL_DAMAGE_STATS = [
	Stat.StatArcaneDamage,
	Stat.StatFireDamage,
	Stat.StatFrostDamage,
	Stat.StatHolyDamage,
	Stat.StatNatureDamage,
	Stat.StatShadowDamage,
];

const SPEED_MULTIPLIERS = [PseudoStat.PseudoStatMeleeSpeedMultiplier, PseudoStat.PseudoStatRangedSpeedMultiplier];

/** `includeBase` says the delta covers the base stage, which is where the base defense skill and the speed multipliers' 1 belong. */
export const statDisplayString = (player: Player<any>, deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean): string => {
	const rootStat = unitStat.hasRootStat() ? unitStat.getRootStat() : null;
	let points = rootStat !== null ? deltaStats.getStat(rootStat) : null;
	const percent = unitStat.convertDefaultUnitsToPercent(deltaStats.getUnitStat(unitStat));
	const percentSuffix = i18n.t('sidebar.character_stats.percent_suffix');

	if (unitStat.equalsStat(Stat.StatDefenseRating)) {
		// Defense is a skill level, not a percentage: no suffix and no decimals.
		return (percent! + (includeBase ? player.getBaseDefense() : 0)).toFixed(0);
	}
	if (rootStat === Stat.StatBlockValue) {
		if (points! > 0) {
			points! *= deltaStats.getPseudoStat(PseudoStat.PseudoStatBlockValueMultiplier) || 1;
		}
		points! += Math.max(0, deltaStats.getStat(Stat.StatStrength) * deltaStats.getPseudoStat(PseudoStat.PseudoStatBlockValuePerStrength) - 1);
	} else if (rootStat && SCHOOL_DAMAGE_STATS.includes(rootStat)) {
		// School damage reads as "total (+school)": the total is the school's own value plus
		// generic spell damage, the parenthesised value is the school-only bonus.
		const total = points! + deltaStats.getStat(Stat.StatSpellDamage);
		const school = `+${points!.toFixed(0)}`;
		return total === 0 ? school : `${Math.round(total)} (${school})`;
	} else if (unitStat.isPseudoStat() && SPEED_MULTIPLIERS.includes(unitStat.getPseudoStat())) {
		// A speed multiplier reads as a signed percentage over 1, like the haste rows; a stage delta
		// without the base stage is already the change.
		const change = (deltaStats.getPseudoStat(unitStat.getPseudoStat()) - (includeBase ? 1 : 0)) * 100;
		return (change >= 0 ? '+' : '') + change.toFixed(2) + percentSuffix;
	}

	if (percent !== null) return percent.toFixed(2) + percentSuffix;
	return points === null ? '' : String(Math.round(points));
};

export const shouldShowMeleeCritCap = (player: Player<any>): boolean => player.getPlayerSpec().isMeleeDpsSpec;

export const shouldShowCritImmunity = (player: Player<any>): boolean => player.getPlayerSpec().isTankSpec;

export const meleeCritCapDisplayString = (playerCritCapDelta: number): string => {
	if (playerCritCapDelta === 0.0) {
		return i18n.t('sidebar.character_stats.crit_cap.exact');
	}

	const prefix = playerCritCapDelta > 0 ? i18n.t('sidebar.character_stats.crit_cap.over_by') : i18n.t('sidebar.character_stats.crit_cap.under_by');
	return `${prefix} ${Math.abs(playerCritCapDelta).toFixed(2)}%`;
};

/** Reads the opposite way round to the melee crit cap: the delta is cap minus total, so positive means short of it. */
export const critImmunityCapDisplayString = (player: Player<any>): string => {
	const critImmuneDelta = player.getCritImmunity();

	if (critImmuneDelta.toFixed(2) === '0.00') {
		return i18n.t('sidebar.character_stats.crit_cap.exact');
	}

	const prefix = critImmuneDelta > 0 ? i18n.t('sidebar.character_stats.crit_cap.under_by') : i18n.t('sidebar.character_stats.crit_cap.over_by');
	return `${prefix} ${Math.abs(critImmuneDelta).toFixed(2)}%`;
};

export const bonusTone = (value: number): 'neutral' | 'positive' | 'negative' => (value === 0 ? 'neutral' : value > 0 ? 'positive' : 'negative');

export const bonusStatClass = (bonusStatValue: number): string => {
	const tone = bonusTone(bonusStatValue);
	return tone === 'neutral' ? 'text-white' : TONE_TEXT[tone];
};

/** The melee crit cap reads the other way round: over the cap is bad. */
export const critCapClass = (capDelta: number): string => (capDelta === 0 ? 'text-white' : TONE_TEXT[capDelta > 0 ? 'negative' : 'positive']);

/** Crit immunity is short-of-cap, so a positive delta is the bad one; zero is tested to two decimals. */
export const critImmunityClass = (capDelta: number): string =>
	capDelta.toFixed(2) === '0.00' ? 'text-white' : TONE_TEXT[capDelta > 0 ? 'negative' : 'positive'];
