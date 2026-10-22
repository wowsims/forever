import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import { Player } from '@sim/player/player';
import { MAX_PARTY_SIZE } from '@sim/raid/party';

export type UNIT_SET = 'aura_sources' | 'aura_sources_targets_first' | 'self_and_pets' | 'targets' | 'players';

export const healingTargetIndices = (player: Player<any>): Array<number> => {
	if (!player.getPlayerSpec().isHealingSpec) return [];

	const raid = player.sim.raid;
	const targetCount = raid.getTargetDummies();
	if (targetCount <= 0) return [];

	const activeSlotCount = raid.getNumActiveParties() * MAX_PARTY_SIZE;
	return raid
		.getPlayers()
		.slice(0, activeSlotCount)
		.flatMap((raidPlayer, index) => (raidPlayer ? [] : [index]))
		.slice(0, targetCount);
};

export const healingTargetsAreGrouped = (player: Player<any>): boolean => {
	const settings = player.sim.raid.getHealingSettings();
	return settings.groupSize !== 5 && settings.assignment === HealingAssignment.RaidHealing;
};

export const healingRandomTargetEnabled = (player: Player<any>): boolean =>
	player.getPlayerSpec().isHealingSpec && player.sim.raid.getHealingSettings().assignment === HealingAssignment.RaidHealing;

const ownPetRefs = (player: Player<any>) =>
	player
		.getPetMetadatas()
		.asList()
		.map((_petMetadata, i) => UnitReference.create({ type: UnitType.Pet, index: i, owner: UnitReference.create({ type: UnitType.Self }) }));

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
				healingTargetIndices(player).map(index => UnitReference.create({ type: UnitType.Player, index })),
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
				healingTargetIndices(player).map(index => UnitReference.create({ type: UnitType.Player, index })),
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
				healingTargetIndices(player).map(index => UnitReference.create({ type: UnitType.Player, index })),
			].flat();
		},
	},
	players: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				player.sim.raid.getActivePlayers().map(player => UnitReference.create({ type: UnitType.Player, index: player.getRaidIndex() })),
				healingTargetIndices(player).map(index => UnitReference.create({ type: UnitType.Player, index })),
			].flat();
		},
	},
};
