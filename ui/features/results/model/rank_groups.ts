import type { ActionId } from '@sim/proto/action_id';
import { ActionMetrics } from '@sim/proto/sim_result';
import { bucket } from '@sim/utils/collections';

import { buildMetricRows, type MetricGrouping, type MetricRow } from './grouping';

type ActionGroup = Array<ActionMetrics>;

// A group holds the tags of one spell, unless a summoned pet rides along in it, which leaves it unkeyed;
// so does a spell with no rank, which keeps its own row next to the ranks of a spell of the same name.
// The lowest tag is part of the key, since two spells of one name can differ only by tag (Penance and its
// heal), while a rank whose extra tag never landed (a DoT that never ticked) still joins the other ranks.
const rankKey = (group: ActionGroup): string | undefined => {
	const { actionId, unit } = group[0];
	if (!actionId.hasRank || unit?.isPet || group.some(action => action.actionId.spellId !== actionId.spellId)) return undefined;
	return `${unit?.unitIndex}|${actionId.baseName}|${Math.min(...group.map(action => action.actionId.tag))}`;
};

const spansRanks = (ids: Array<ActionId>) => new Set(ids.map(id => id.rank)).size > 1;

const notPet = (metric: ActionMetrics) => !metric.unit?.isPet;

/**
 * The rows of an action table, one per group as `buildMetricRows` builds them, a pet's actions under the
 * pet. With `ranksGrouped`, the groups of one unit's ranked spells that share a base name and their lowest
 * tag and hold at least two ranks become one row named after the highest rank without its rank, with a child row
 * per rank, and a rank with several tags keeps a row per tag under it.
 */
export const buildActionRows = (
	groups: Array<ActionGroup>,
	ranksGrouped: boolean,
	shouldCollapse: (metric: ActionMetrics) => boolean = notPet,
): Array<MetricRow<ActionMetrics>> => {
	const grouping: MetricGrouping<ActionMetrics> = {
		merge: metrics => ActionMetrics.merge(metrics, { removeTag: true, actionIdOverride: metrics[0].unit?.petActionId || undefined }),
		shouldCollapse,
	};
	if (!ranksGrouped) return buildMetricRows(groups, grouping);

	// One rank of a spell: its one member as it is, tag label included, or the rank merged over its tags with
	// a row per tag.
	const rankRow = (group: ActionGroup): MetricRow<ActionMetrics> => (group.length === 1 ? { metric: group[0] } : buildMetricRows([group], grouping)[0]);
	// One tag other than 0 across the whole family names the spell (Penance (Heal)), so the joined row keeps
	// it; with tag 0 or several tags it takes the plain name.
	const familyRow = (family: Array<ActionGroup>): MetricRow<ActionMetrics> => {
		const highest = family.reduce((top, group) => (group[0].actionId.rank > top[0].actionId.rank ? group : top));
		const tags = new Set(family.flat().map(action => action.actionId.tag));
		const keepsTag = tags.size === 1 && !tags.has(0);
		const ranks = family.map(rankRow);
		const metric = ActionMetrics.merge(
			ranks.map(rank => rank.metric),
			{ removeTag: !keepsTag, actionIdOverride: highest[0].actionId.withoutRank() },
		);
		return { metric, subRows: ranks };
	};

	const keyed = groups.flatMap(group => {
		const key = rankKey(group);
		return key ? [{ group, key }] : [];
	});
	const families = Object.values(bucket(keyed, entry => entry.key))
		.map(family => family.map(entry => entry.group))
		.filter(family => spansRanks(family.map(group => group[0].actionId)));
	const familyAt = new Map(families.map(family => [family[0], family]));
	const joined = new Set(families.flat());

	return groups.flatMap(group => {
		const family = familyAt.get(group);
		if (family) return [familyRow(family)];
		return joined.has(group) ? [] : buildMetricRows([group], grouping);
	});
};
