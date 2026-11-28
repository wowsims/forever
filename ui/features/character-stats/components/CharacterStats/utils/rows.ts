import { OFF_HAND_PERCENTS, STAT_GROUPS, type StatGroupKey } from '@features/stat-filters/model/stat_groups';
import { PseudoStat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';

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

export interface RowGroup {
	key: StatGroupKey;
	rows: Row[];
}

const displayStatOf = (stat: UnitStat): DisplayStat => {
	const offHandStat = OFF_HAND_PERCENTS.get(stat.getKey());
	return offHandStat ? { stat, offHandStat } : { stat };
};

const statGroups = new Map<StatGroupKey, Array<DisplayStat>>(STAT_GROUPS.map(({ key, stats }) => [key, stats.map(displayStatOf)]));

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

		// The tank readouts hang off the defensive block.
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
