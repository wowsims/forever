import type { ActionId } from '@sim/proto/action_id';
import { classTalentsConfig } from '@sim/talents/factory';

/** A talent tree as the talent calculator shows it: its name and its place among the class's trees. */
export interface TalentTree {
	name: string;
	index: number;
}

const treesBySpellId = new Map(
	Object.values(classTalentsConfig).flatMap(config =>
		config.flatMap((tree, index) => tree.talents.map(talent => [talent.spellId, { name: tree.name, index }] as const)),
	),
);

/**
 * The tree of the held talent an entry is named after. `held` is the unit's talents with points in
 * them, so a unit with none, such as a pet, has no trees.
 */
export const talentTrees = (held: Array<ActionId>) => {
	const byName = new Map(
		held.flatMap(talent => {
			const tree = treesBySpellId.get(talent.spellId);
			return tree ? [[talent.baseName, tree] as const] : [];
		}),
	);
	return (name: string): TalentTree | undefined => byName.get(name);
};
