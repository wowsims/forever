import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { AreaType, ConsumesSpec, HealingModel, Profession, Race } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import { BearEp, HybridEp } from '../feral_ep';
import AdaptiveApl from './apls/adaptive.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
export const ROTATION_ADAPTIVE = PresetUtils.makePresetAPLRotation('Adaptive Bear', AdaptiveApl);

export const BearTalents = PresetUtils.makePresetTalents('Bear', SavedTalents.create({ talentsString: '050022-2322030212310212501-055' }));
export const HybridTalents = PresetUtils.makePresetTalents('Hybrid', SavedTalents.create({ talentsString: '050022-50230332123110121011-055' }));

export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	devOnly: true,
	tooltip: 'Leather at level 60 with no set bonus and no proc. The idol slot is empty: every idol has an effect.',
});

const encounter = Encounter.defaultEncounterProto();
encounter.areaTypes = [AreaType.AreaTypeForestGrassland, AreaType.AreaTypeMountainous];
encounter.zoneId = 616;
encounter.targets[0].minBaseDamage = 6000;
export const Level60Encounter = PresetUtils.makePresetEncounter('Mount Hyjal', encounter);

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
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.JujuMight.id, Consumables.ScrollOfStaminaIV.id, Consumables.LungJuiceCocktail.id],
	elixirIds: [
		Consumables.GiftOfArthas.id,
		Consumables.ElixirOfTheGrizzly.id,
		Consumables.ElixirOfCunning.id,
		Consumables.ElixirOfThePhalanx.id,
		Consumables.ElixirOfGreaterDefense.id,
	],
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
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
