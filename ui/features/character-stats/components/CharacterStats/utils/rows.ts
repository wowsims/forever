import { PseudoStat, Stat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';

import { shouldShowCritImmunity, shouldShowMeleeCritCap } from './stat_display';

/**
 * A row's stat plus its per-row display flags. `notEditable` suppresses the bonus-stat picker, and
 * `offHandStat` pairs an off-hand pseudo stat with a main-hand one so a single row can show both
 * while dual wielding.
 */
export type DisplayStat = {
	stat: UnitStat;
	notEditable?: boolean;
	offHandStat?: UnitStat;
};

export type Row =
	| { kind: 'stat'; id: string; displayStat: DisplayStat }
	| { kind: 'miss'; id: string }
	| { kind: 'avoidance'; id: string }
	| { kind: 'crit-immunity'; id: string }
	| { kind: 'melee-crit-cap'; id: string };

/** `primary` (health and mana) shows without a header; the rest are collapsible. */
export type StatGroupKey = 'primary' | 'attributes' | 'physical' | 'spell' | 'defense' | 'resistance';

export interface RowGroup {
	key: StatGroupKey;
	rows: Row[];
}

const statGroups = new Map<StatGroupKey, Array<DisplayStat>>([
	['primary', [{ stat: UnitStat.fromStat(Stat.StatHealth) }, { stat: UnitStat.fromStat(Stat.StatMana) }]],
	[
		'attributes',
		[
			{ stat: UnitStat.fromStat(Stat.StatStrength) },
			{ stat: UnitStat.fromStat(Stat.StatAgility) },
			{ stat: UnitStat.fromStat(Stat.StatStamina) },
			{ stat: UnitStat.fromStat(Stat.StatIntellect) },
			{ stat: UnitStat.fromStat(Stat.StatSpirit) },
		],
	],
	[
		'physical',
		[
			{ stat: UnitStat.fromStat(Stat.StatAttackPower) },
			{ stat: UnitStat.fromStat(Stat.StatFeralAttackPower) },
			{ stat: UnitStat.fromStat(Stat.StatRangedAttackPower) },
			{
				stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
				offHandStat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent),
			},
			{
				stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
				offHandStat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent),
			},
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHastePercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedSpeedMultiplier) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent) },
			{ stat: UnitStat.fromStat(Stat.StatArmorPenetration) },
		],
	],
	[
		'spell',
		[
			{ stat: UnitStat.fromStat(Stat.StatSpellDamage) },
			{ stat: UnitStat.fromStat(Stat.StatHealingPower) },
			{ stat: UnitStat.fromStat(Stat.StatArcaneDamage) },
			{ stat: UnitStat.fromStat(Stat.StatFireDamage) },
			{ stat: UnitStat.fromStat(Stat.StatFrostDamage) },
			{ stat: UnitStat.fromStat(Stat.StatHolyDamage) },
			{ stat: UnitStat.fromStat(Stat.StatNatureDamage) },
			{ stat: UnitStat.fromStat(Stat.StatShadowDamage) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFrost) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentHoly) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentNature) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellCritPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHastePercent) },
			{ stat: UnitStat.fromStat(Stat.StatSpellPiercing) },
			{ stat: UnitStat.fromStat(Stat.StatMP5) },
		],
	],
	[
		'defense',
		[
			{ stat: UnitStat.fromStat(Stat.StatArmor) },
			{ stat: UnitStat.fromStat(Stat.StatBonusArmor) },
			{ stat: UnitStat.fromStat(Stat.StatDefenseRating) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatParryPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent) },
			{ stat: UnitStat.fromStat(Stat.StatBlockValue) },
		],
	],
	[
		'resistance',
		[
			{ stat: UnitStat.fromStat(Stat.StatArcaneResistance) },
			{ stat: UnitStat.fromStat(Stat.StatFireResistance) },
			{ stat: UnitStat.fromStat(Stat.StatFrostResistance) },
			{ stat: UnitStat.fromStat(Stat.StatNatureResistance) },
			{ stat: UnitStat.fromStat(Stat.StatShadowResistance) },
		],
	],
]);

const statsInGroup = (groupedStats: ReadonlyArray<DisplayStat>, stats: ReadonlyArray<UnitStat>) =>
	groupedStats.filter(displayStat => stats.some(stat => stat.equals(displayStat.stat)));

const keyStatsIn = (group: StatGroupKey, keyStats: ReadonlyArray<UnitStat>) => statsInGroup(statGroups.get(group)!, keyStats).length;

/** Attributes and the spec's main group: Defense for a tank, otherwise whichever of Physical and Spell holds more of its key stats. */
export const defaultOpenStatGroups = (isTank: boolean, keyStats: ReadonlyArray<UnitStat>): StatGroupKey[] => {
	if (isTank) return ['attributes', 'defense'];
	return ['attributes', keyStatsIn('spell', keyStats) > keyStatsIn('physical', keyStats) ? 'spell' : 'physical'];
};

export interface RowLayout {
	groups: RowGroup[];
	/** The stats actually on screen: the avoidance tooltip only lists parry and block if their rows are. */
	shownStats: UnitStat[];
}

export const buildRows = (player: Player<any>, statList: Array<UnitStat>): RowLayout => {
	const showCritImmunity = shouldShowCritImmunity(player);
	const showMeleeCritCap = shouldShowMeleeCritCap(player);
	const groups: RowGroup[] = [];
	const shownStats: UnitStat[] = [];

	statGroups.forEach((groupedStats, key) => {
		const filtered = statsInGroup(groupedStats, statList);
		if (!filtered.length) return;

		const rows: Row[] = [];
		filtered.forEach(displayStat => {
			shownStats.push(displayStat.stat);
			rows.push({ kind: 'stat', id: `${key}-${displayStat.stat.getKey()}`, displayStat });
		});

		// The melee crit cap sits directly under the Melee Crit row, in the Physical block.
		if (key === 'physical' && showMeleeCritCap) {
			const meleeCritRow = filtered.findIndex(displayStat => displayStat.stat.equalsPseudoStat(PseudoStat.PseudoStatMeleeCritPercent));
			if (meleeCritRow !== -1) {
				rows.splice(meleeCritRow + 1, 0, { kind: 'melee-crit-cap', id: `${key}-melee-crit-cap` });
			}
		}

		// The tank readouts hang off the defensive block, in the order the sheet has always shown them.
		if (key === 'defense' && showCritImmunity) {
			rows.push(
				{ kind: 'miss', id: `${key}-miss` },
				{ kind: 'avoidance', id: `${key}-avoidance` },
				{ kind: 'crit-immunity', id: `${key}-crit-immunity` },
			);
		}

		groups.push({ key, rows });
	});

	return { groups, shownStats };
};
