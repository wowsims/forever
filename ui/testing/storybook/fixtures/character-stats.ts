import type { MeleeCritCapInfo } from '@generated/proto/api';
import { PseudoStat, RangedWeaponType, Stat } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import { type StatAttribution, Stats, type UnitStat } from '@sim/proto/stats';

const stage = (strength: number, attackPower: number, meleeHit: number, offHandHit: number, rangedHaste: number) =>
	new Stats()
		.withStat(Stat.StatStrength, strength)
		.withStat(Stat.StatAttackPower, attackPower)
		.withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, meleeHit)
		.withPseudoStat(PseudoStat.PseudoStatOffHandHitPercent, offHandHit)
		.withPseudoStat(PseudoStat.PseudoStatRangedHastePercent, rangedHaste);

/** Each stage's own contribution, and their sum as `final`. */
export const makeStatAttribution = (): StatAttribution => {
	const base = stage(120, 160, 0, 0, 0);
	const gear = stage(180, 240, 3, 2, 0);
	const talents = stage(0, 0, 3, 3, 0);
	const buffs = stage(45, 190, 0, 0, 15);
	const consumes = stage(25, 0, 0, 0, 0);
	const debuffs = stage(0, 0, 2, 2, 0);
	return { base, gear, talents, buffs, consumes, debuffs, final: base.add(gear).add(talents).add(buffs).add(consumes).add(debuffs) };
};

/** A stand-in for the sheet's formatter: a pseudo stat as a percentage, a stat as a whole number. */
export const showStat = (deltaStats: Stats, unitStat: UnitStat): string => {
	const value = deltaStats.getUnitStat(unitStat);
	return unitStat.isPseudoStat() ? `${value.toFixed(2)}%` : value.toFixed(0);
};

export const makeRangedWeapon = (): UIItem => UIItem.create({ rangedWeaponType: RangedWeaponType.RangedWeaponTypeBow, weaponSpeed: 2.9 });

export const MISS_INFO = { base: 5, defense: 2.4, debuffs: 1, total: 8.4 };

export const AVOIDANCE_INFO = { miss: 8.4, dodge: 18.2, parry: 14.1, block: 42.3, total: 103, shear: 60.7 };

export const CRIT_IMMUNITY_INFO = { total: 5.6, delta: 0, defense: 4.6, talents: 1 };

export const CRIT_CAP_INFO: MeleeCritCapInfo = {
	glancing: 40,
	suppression: 4.8,
	remainingHit: 1.2,
	remainingExpertise: 2.5,
	block: 0,
	specOffset: 0,
	critCap: 61.1,
	delta: -8.4,
};
