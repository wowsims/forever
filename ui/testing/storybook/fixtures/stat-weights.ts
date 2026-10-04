import type { StatWeightsResult } from '@generated/proto/api';
import { Stat, UnitStats } from '@generated/proto/common';
import type { DisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { UnitStat } from '@sim/proto/stats';

import type { EpColumn } from '../../../features/stat-weights/components/EpWeightsDialog/types';
import { buildEpColumns } from '../../../features/stat-weights/components/EpWeightsDialog/utils';
import { emptyStatWeightsResult } from '../../../features/stat-weights/model/ep_math';
import { statsTableColumns } from '../../../features/stat-weights/model/stats_table';

export const STAT_WEIGHT_STATS = [Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower, Stat.StatCritRating].map(stat => UnitStat.fromStat(stat));

/**
 * A finished run for a Strength-referenced melee spec: Strength 2.4 DPS per point, Agility 1.9,
 * Attack Power 1.0 and Crit Rating 3.1, each ±5%. EP is the DPS weight over Strength's; the healing
 * and threat metrics carry half the weights and the same EP.
 */
export const makeStatWeightsResult = (): StatWeightsResult => {
	const result = emptyStatWeightsResult();
	const weights: Array<[UnitStat, number]> = STAT_WEIGHT_STATS.map((stat, index) => [stat, [2.4, 1.9, 1, 3.1][index]]);
	const metrics = [result.dps!, result.hps!, result.tps!, result.dtps!, result.tmi!, result.pDeath!];
	metrics.forEach((values, metricIndex) => {
		const scale = metricIndex === 0 ? 1 : 0.5;
		for (const [stat, weight] of weights) {
			stat.setProtoValue(values.weights!, weight * scale);
			stat.setProtoValue(values.weightsStdev!, weight * scale * 0.05);
			stat.setProtoValue(values.epValues!, weight / 2.4);
			stat.setProtoValue(values.epValuesStdev!, (weight / 2.4) * 0.05);
		}
	});
	return result;
};

/** The dialog's column set for the metrics shown, reading its weights from `result`. */
export const makeEpColumns = (
	displayMetrics: DisplayMetrics = { damage: true, healing: false, threat: false },
	result: StatWeightsResult = makeStatWeightsResult(),
): EpColumn[] =>
	buildEpColumns(
		statsTableColumns({
			getPrevSimResult: () => result,
			getDefaultEpWeights: () => UnitStats.create(),
			getDpsEpRefStat: () => Stat.StatStrength,
			getHealEpRefStat: () => Stat.StatStrength,
			getTankEpRefStat: () => Stat.StatArmor,
		}),
		() => {},
	).filter(column => !column.metric || displayMetrics[column.metric]);
