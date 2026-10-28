import { healingTargetDummies } from '@features/encounter/model/target_dummies';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Cooldowns, Profession, Spec } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import type { IndividualSimUIConfig } from '@sim/spec_config';
import { batch } from '@sim/state/batch';

import { applyBuild } from './apply_build';

export type DefaultsHost<SpecType extends Spec> = IndividualSimHost<SpecType> & {
	readonly statWeightActionSettings: StatWeightActionSettings;
	readonly statSweepSettings: StatSweepSettings;
};

export function applyDefaultRotation<SpecType extends Spec>(player: Player<SpecType>, config: IndividualSimUIConfig<SpecType>) {
	batch(() => {
		const defaultRotationType = config.defaults.rotationType || APLRotationType.TypeAuto;

		// A TypeAPL spec ships its own priority list. Setting only the type would load
		// it with an empty APL, and nothing backfills it afterwards.
		if (defaultRotationType === APLRotationType.TypeAPL && config.defaults.aplRotation) {
			player.setAplRotation(APLRotation.create(config.defaults.aplRotation));
			return;
		}

		player.setAplRotation(
			APLRotation.create({
				type: defaultRotationType,
			}),
		);

		if (!config.defaults.simpleRotation) {
			return;
		}

		const defaultSimpleRotation = config.defaults.simpleRotation || player.specTypeFunctions.rotationCreate();
		player.setSimpleRotation(defaultSimpleRotation);
		player.setSimpleCooldowns(
			Cooldowns.create({
				hpPercentForDefensives: player.playerSpec.isTankSpec ? 0.4 : 0,
			}),
		);
	});
}

export function applyEmptyAplRotation<SpecType extends Spec>(player: Player<SpecType>) {
	batch(() => {
		player.setAplRotation(
			APLRotation.create({
				type: APLRotationType.TypeAPL,
			}),
		);
	});
}

// `host` rather than its parts: this runs from the persistence load queued on the first
// `waitForInit()`, by which time the constructor has set `reforger` — capturing it up front
// would silently skip the reforge defaults.
export function applyIndividualDefaults<SpecType extends Spec>(host: DefaultsHost<SpecType>) {
	const { player, sim, individualConfig: config } = host;
	batch(() => {
		const tankSpec = player.getPlayerSpec().isTankSpec;
		const healingSpec = player.getPlayerSpec().isHealingSpec;

		player.applySharedDefaults();
		player.setRace(config.defaults.other?.race || player.getPlayerClass().races[0]);
		player.setGear(sim.db.lookupEquipmentSpec(config.defaults.gear));
		if (config.defaults.itemSwap) {
			player.itemSwapSettings.setItemSwapSettings(true, sim.db.lookupItemSwap(config.defaults.itemSwap));
		}
		player.setConsumes(config.defaults.consumables);
		applyDefaultRotation(player, config);
		player.setTalentsString(config.defaults.talents.talentsString);
		player.setSpecOptions(config.defaults.specOptions);
		player.setBuffs(config.defaults.individualBuffs);
		player.getParty()!.setBuffs(config.defaults.partyBuffs);
		player.getRaid()!.setBuffs(config.defaults.raidBuffs);
		player.setEpWeights(config.defaults.epWeights);
		// Tank specs can name their own EP reference stat; without this the dialog
		// falls back to StatArmor for every one of them.
		player.setRefStat('tankRefStat', config.tankRefStat);

		const defaultRatios = player.getDefaultEpRatios(tankSpec, healingSpec);
		player.setEpRatios(defaultRatios);
		player.setProfession1(config.defaults.other?.profession1 || Profession.Engineering);

		if (config.defaults.other?.profession2 === undefined) {
			player.setProfession2(Profession.ProfessionUnknown);
		} else {
			player.setProfession2(config.defaults.other.profession2);
		}

		player.setDistanceFromTarget(config.defaults.other?.distanceFromTarget || 0);
		player.setChannelClipDelay(config.defaults.other?.channelClipDelay || 0);
		player.setReactionTime(config.defaults.other?.reactionTime || 100);

		host.reforger?.applyDefaults();

		// Healers follow their healing settings. Balance keeps one ally so Innervate can leave the druid.
		if (healingSpec) {
			sim.raid.setTargetDummies(healingTargetDummies(sim.raid.getHealingSettings()));
		} else if (player.getSpec() === Spec.SpecBalanceDruid) {
			sim.raid.setTargetDummies(1);
		} else {
			sim.raid.setTargetDummies(0);
		}
		if (config.defaults.encounter?.encounter) {
			sim.encounter.fromProto(config.defaults.encounter.encounter);
		} else {
			sim.encounter.applyDefaults();
		}
		if (config.defaults.other?.healingModel) {
			player.setHealingModel(config.defaults.other.healingModel);
		}
		sim.raid.setDebuffs(config.defaults.debuffs);
		sim.applyDefaults(tankSpec, healingSpec);

		if (config.defaults.other?.iterationCount) {
			sim.setIterations(config.defaults.other!.iterationCount!);
		}

		if (tankSpec) {
			sim.raid.setTanks([player.makeUnitReference()]);
		} else {
			sim.raid.setTanks([]);
		}

		host.statWeightActionSettings.applyDefaults();
		host.statSweepSettings.applyDefaults();

		if (config.defaultBuild) {
			applyBuild(config.defaultBuild, host);
		}
	});
}
