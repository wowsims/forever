import type { ActionMetrics } from '@sim/proto/sim_result';
import { useMemo } from 'react';

import { useGroupRanks } from '../context/GroupRanksContext';
import type { MetricRow } from '../model/grouping';
import { buildActionRows } from '../model/rank_groups';
import type { SimResultData } from '../model/result_data';
import { useSimResult } from './useSimResult';

const NO_ROWS: Array<MetricRow<ActionMetrics>> = [];

/**
 * The last result and the rows an action table builds from it, ranks grouped as the toolbar has it.
 * `groupsOf` and `shouldCollapse` must be stable, such as module-level functions, or the rows rebuild on
 * every render. Without `shouldCollapse`, a pet's single action still sits under the pet.
 */
export const useActionRows = (
	groupsOf: (resultData: SimResultData) => Array<Array<ActionMetrics>>,
	shouldCollapse?: (metric: ActionMetrics) => boolean,
): { resultData: SimResultData | null; rows: Array<MetricRow<ActionMetrics>> } => {
	const resultData = useSimResult();
	const ranksGrouped = useGroupRanks();
	const rows = useMemo(
		() => (resultData ? buildActionRows(groupsOf(resultData), ranksGrouped, shouldCollapse) : NO_ROWS),
		[resultData, groupsOf, shouldCollapse, ranksGrouped],
	);
	return { resultData, rows };
};
