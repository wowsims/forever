import { healingGroupSize } from '@features/encounter/model/target_dummies';
import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import { Player } from '@sim/player/player';

export type UNIT_SET = 'aura_sources' | 'aura_sources_targets_first' | 'self_and_pets' | 'targets' | 'players';

/** The target dummy slots a healer's rotation offers as heal targets; other specs get none. */
export const healTargetSlots = (player: Player<any>): Array<number> => (player.getPlayerSpec().isHealingSpec ? player.sim.raid.getTargetDummySlots() : []);

/** Raid healing files heal targets by group and offers a random friendly target. A dungeon group has no assignment to pick, so it never does. */
export const isRaidHealing = (player: Player<any>): boolean => {
	const settings = player.sim.raid.getHealingSettings();
	return healingGroupSize(settings) !== 5 && settings.assignment === HealingAssignment.RaidHealing;
};

const ownPetRefs = (player: Player<any>) =>
	player
		.getPetMetadatas()
		.asList()
		.map((_petMetadata, i) => UnitReference.create({ type: UnitType.Pet, index: i, owner: UnitReference.create({ type: UnitType.Self }) }));

const healingTargetRefs = (player: Player<any>) => healTargetSlots(player).map(index => UnitReference.create({ type: UnitType.Player, index }));

export const unitSets: Record<
	UNIT_SET,
	{
		// Uses target icon by default instead of person icon. This should be set to true for inputs that default to CurrentTarget.
		targetUI?: boolean;
		getUnits: (player: Player<any>) => Array<UnitReference | undefined>;
	}
> = {
	aura_sources: {
		getUnits: player => {
			return [
				undefined,
				ownPetRefs(player),
				UnitReference.create({ type: UnitType.CurrentTarget }),
				UnitReference.create({ type: UnitType.PreviousTarget }),
				UnitReference.create({ type: UnitType.NextTarget }),
				player.sim.raid
					.getActivePlayers()
					.filter(filter => filter != player)
					.map(mapPlayer => UnitReference.create({ type: UnitType.Player, index: mapPlayer.getRaidIndex() })),
				healingTargetRefs(player),
				player.sim.encounter.targetsMetadata.asList().map((targetMetadata, i) => UnitReference.create({ type: UnitType.Target, index: i })),
			].flat();
		},
	},
	aura_sources_targets_first: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				player.sim.encounter.targetsMetadata.asList().map((targetMetadata, i) => UnitReference.create({ type: UnitType.Target, index: i })),
				UnitReference.create({ type: UnitType.Self }),
				ownPetRefs(player),
				healingTargetRefs(player),
			].flat();
		},
	},
	// Lists every own pet. The sim drops a value read from a pet without that resource, so a bare action condition counts as always true and an And or Or loses that branch.
	self_and_pets: {
		getUnits: player => {
			return [undefined, ownPetRefs(player)].flat();
		},
	},
	targets: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				player.sim.encounter.targetsMetadata.asList().map((_targetMetadata, i) => UnitReference.create({ type: UnitType.Target, index: i })),
				UnitReference.create({ type: UnitType.PreviousTarget }),
				UnitReference.create({ type: UnitType.NextTarget }),
				healingTargetRefs(player),
			].flat();
		},
	},
	players: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				player.sim.raid.getActivePlayers().map(player => UnitReference.create({ type: UnitType.Player, index: player.getRaidIndex() })),
				healingTargetRefs(player),
			].flat();
		},
	},
};
