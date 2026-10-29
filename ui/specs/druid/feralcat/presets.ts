import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { AreaType, ConsumesSpec, Profession, Race } from '@generated/proto/common';
import { FeralCatDruid_Options as FeralDruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Encounter } from '@sim/raid/encounter';

import { CatEp, HybridEp } from '../feral_ep';
import AdaptiveApl from './apls/adaptive.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
import PreRaidGearJson from './gear_sets/pre_raid.gear.json';

// Another raid member's bleed keeps Rend and Tear up.
export const DefaultOptions = FeralDruidOptions.create({ assumeBleedActive: true });

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfNaturalAggression.id,
	foodId: Consumables.SwiftstrikeSteak.id,
	conjuredId: Consumables.DemonicRune.id,
	explosiveId: Consumables.EZThroDarkBomb.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	swapMhImbueId: Consumables.DenseWeightstone.id,
	extraIds: [Consumables.JujuMight.id, Consumables.GroundScorpokAssay.id, Consumables.GoblinSapperCharge.id],
	elixirIds: [
		Consumables.ElixirOfCunning.id,
		Consumables.ElixirOfSages.id,
		Consumables.ElixirOfThePhalanx.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.ElixirOfTheGrizzly.id,
	],
	resourcePotId: Consumables.MajorManaPotion.id,
	offensivePotId: Consumables.MightyRagePotion.id,
});

const encounter = Encounter.defaultEncounterProto();
encounter.areaTypes = [AreaType.AreaTypeForestGrassland, AreaType.AreaTypeMountainous];
encounter.zoneId = 616;
export const DefaultEncounter = PresetUtils.makePresetEncounter('Mount Hyjal', encounter);

export const OtherDefaults = {
	distanceFromTarget: 25,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceTauren,
	reactionTime: 250,
};

// Leather at 60 with no set bonus and no proc. The stand-in the rotation was tuned on.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	devOnly: true,
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreRaidGearJson, {
	devOnly: true,
	tooltip: 'Pre-raid leather with The Unstoppable Force, Wolfshead Helm and full enchants.',
});

export const DefaultEp = CatEp;
export const DefaultEpWeights = PresetUtils.makePresetEpWeights('Cat', DefaultEp);
export const HybridEpWeights = PresetUtils.makePresetEpWeights('Hybrid', HybridEp);

// Uses Thistle Tea and optional Sapper, with racial actions gated by availability.
// Uses Dash for the approach; Sapper can be used without leaving Cat Form.
export const JVMES_APL = {
	...PresetUtils.makePresetAPLRotation('Adaptive Cat', AdaptiveApl),
	tooltip:
		'Maintains Faerie Fire and bleeds, pools for later Berserks, and uses optional consumables and racials. Supports an MCP opener from the weapon swap slot.',
};
export const CatTalents = PresetUtils.makePresetTalents('Cat', SavedTalents.create({ talentsString: '050022-35200032121130212051-053' }));
export const HybridTalents = PresetUtils.makePresetTalents('Hybrid', SavedTalents.create({ talentsString: '050022-35230332121132212051-' }));
