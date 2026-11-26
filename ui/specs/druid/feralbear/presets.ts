import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, Race } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import { BearEp, HybridEp } from '../feral_ep';
import AdaptiveApl from './apls/adaptive.apl.siml';
import Level60GearSet from './gear_sets/default.gear.siml';
export const ROTATION_ADAPTIVE = PresetUtils.makePresetAPLRotation('Adaptive Bear', AdaptiveApl);

export const BearTalents = PresetUtils.makePresetTalents('Bear', SavedTalents.create({ talentsString: '050022-5023030212310112501-055' }));
export const HybridTalents = PresetUtils.makePresetTalents('Hybrid', SavedTalents.create({ talentsString: '050012-50230302123110125011-055' }));

export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearSet, {
	devOnly: true,
	tooltip: 'Leather at level 60 with no set bonus and no proc. The idol slot is empty: every idol has an effect.',
});

const encounter = Encounter.defaultEncounterProto();
encounter.duration = 180;
encounter.durationVariation = 5;
encounter.targets[0].minBaseDamage = 6500;
export const DefaultEncounter = PresetUtils.makePresetEncounter('Default', encounter);

export const DefaultEp = BearEp;
export const DefaultEpWeights = PresetUtils.makePresetEpWeights('Bear', DefaultEp);
export const HybridEpWeights = PresetUtils.makePresetEpWeights('Hybrid', HybridEp);

// Another raid member's bleed keeps Rend and Tear up.
export const DefaultOptions = DruidOptions.create({
	startingRage: 0,
	assumeBleedActive: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfNaturalPrecision.id,
	foodId: Consumables.DirgesKickinChimaerokChops.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	mhImbueId: Consumables.ElementalSharpeningStone.id,
	swapMhImbueId: Consumables.DenseWeightstone.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.JujuMight.id, Consumables.LungJuiceCocktail.id],
	elixirIds: [Consumables.GiftOfArthas.id, Consumables.ElixirOfTheGrizzly.id, Consumables.ElixirOfCunning.id, Consumables.ElixirOfThePhalanx.id],
	defensivePotId: Consumables.GreaterStoneshieldPotion.id,
	offensivePotId: Consumables.MightyRagePotion.id,
});

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceTauren,
	distanceFromTarget: 25,
	reactionTime: 250,
	healingModel: HealingModel.create({
		hps: 1300,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
