import { PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Stats, UnitStat } from '@sim/proto/stats';
import { TONE_TEXT } from '@ui-kit/utils/colors';

const PER_ATTACK_RATING = new Map<PseudoStat, PseudoStat>([
	[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMainHandHitRating],
	[PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMainHandCritRating],
	[PseudoStat.PseudoStatOffHandHitPercent, PseudoStat.PseudoStatOffHandHitRating],
	[PseudoStat.PseudoStatOffHandCritPercent, PseudoStat.PseudoStatOffHandCritRating],
	[PseudoStat.PseudoStatRangedHitPercent, PseudoStat.PseudoStatRangedHitRating],
	[PseudoStat.PseudoStatRangedCritPercent, PseudoStat.PseudoStatRangedCritRating],
]);

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
	const perAttackRating = unitStat.isPseudoStat() ? PER_ATTACK_RATING.get(unitStat.getPseudoStat()) : undefined;
	let rootRatingValue = rootStat !== null ? deltaStats.getStat(rootStat) : null;
	// Bonus stats carry the rating alone, with no per-attack value to read.
	if (perAttackRating !== undefined && deltaStats.getPseudoStat(perAttackRating) !== 0) {
		rootRatingValue = deltaStats.getPseudoStat(perAttackRating);
	}
	let percentDecimals = 2;
	let derivedPercentOrPointsValue = unitStat.convertDefaultUnitsToPercent(deltaStats.getUnitStat(unitStat));
	let displayPrefix = '';
	let displaySuffix = i18n.t('sidebar.character_stats.percent_suffix');

	if (unitStat.equalsStat(Stat.StatDefenseRating)) {
		// Defense is a skill level, not a percentage: no suffix, no decimals, and the rating is
		// always shown even at zero.
		displaySuffix = '';
		percentDecimals = 0;
		if (includeBase) {
			derivedPercentOrPointsValue! += player.getBaseDefense();
		}
	} else if (rootStat == Stat.StatBlockValue) {
		if (rootRatingValue !== null) {
			if (rootRatingValue > 0) {
				rootRatingValue *= deltaStats.getPseudoStat(PseudoStat.PseudoStatBlockValueMultiplier) || 1;
			}
			rootRatingValue += Math.max(0, deltaStats.getStat(Stat.StatStrength) * deltaStats.getPseudoStat(PseudoStat.PseudoStatBlockValuePerStrength) - 1);
		}
	} else if (rootStat && SCHOOL_DAMAGE_STATS.includes(rootStat)) {
		// School damage reads as "total (+school)": the rating column carries the school's own
		// value plus generic spell damage, the parenthesised value is the school-only bonus.
		if (rootRatingValue !== null) {
			displayPrefix = '+';
			displaySuffix = '';
			percentDecimals = 0;
			derivedPercentOrPointsValue = rootRatingValue;
			rootRatingValue += deltaStats.getStat(Stat.StatSpellDamage);
		}
	} else if (unitStat.isPseudoStat() && SPEED_MULTIPLIERS.includes(unitStat.getPseudoStat())) {
		// A speed multiplier reads as a signed percentage over 1, like the haste rows; a stage delta
		// without the base stage is already the change.
		derivedPercentOrPointsValue = (deltaStats.getPseudoStat(unitStat.getPseudoStat()) - (includeBase ? 1 : 0)) * 100;
		if (derivedPercentOrPointsValue >= 0) displayPrefix = '+';
	}

	const hideRootRating =
		(rootRatingValue === null || (rootRatingValue === 0 && derivedPercentOrPointsValue !== null)) && !unitStat.equalsStat(Stat.StatDefenseRating);
	const rootRatingString = hideRootRating ? '' : String(Math.round(rootRatingValue!));
	const percentOrPointsString =
		derivedPercentOrPointsValue === null ? '' : displayPrefix + `${derivedPercentOrPointsValue.toFixed(percentDecimals)}` + displaySuffix;
	const wrappedPercentOrPointsString = hideRootRating || derivedPercentOrPointsValue === null ? percentOrPointsString : ` (${percentOrPointsString})`;
	return rootRatingString + wrappedPercentOrPointsString;
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
