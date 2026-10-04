import { petNameToSummonSpell } from '@sim/proto/action_id';
import { ActionMetrics, type UnitMetrics } from '@sim/proto/sim_result';
import { bucket } from '@sim/utils/collections';

/**
 * A player's action groups, then its pets'. A pet summoned by one of the player's spells goes under
 * that spell as one row; any other pet is a group of its own.
 */
export const groupActionsWithPets = (player: UnitMetrics, actionsOf: (unit: UnitMetrics) => Array<ActionMetrics>): Array<Array<ActionMetrics>> => {
	const groups = ActionMetrics.groupById(actionsOf(player));
	for (const pets of Object.values(bucket(player.pets, pet => pet.name))) {
		const actions = ActionMetrics.joinById(pets.flatMap(actionsOf), true);
		if (!actions.length) continue;
		const summonSpell = petNameToSummonSpell[pets[0].name];
		const summon = summonSpell ? groups.find(group => group[0].actionId.baseName === summonSpell) : undefined;
		if (summon) {
			summon.push(ActionMetrics.merge(actions, { actionIdOverride: pets[0].petActionId || undefined }));
		} else {
			groups.push(actions);
		}
	}
	return groups;
};
