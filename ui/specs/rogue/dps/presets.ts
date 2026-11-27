import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { Rogue_Options as RogueOptions, Rogue_Rotation as RogueRotation } from '@generated/proto/rogue';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultAPL from './apls/default.apl.json';
import BackstabGear from './gear_sets/backstab.gear.json';
import HemoGear from './gear_sets/hemo.gear.json';
import MutiGear from './gear_sets/muti.gear.json';
import SinisterGear from './gear_sets/sinister.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const TALENTS_MUTILATE_ASN = {
	name: 'Asn Muti',
	data: SavedTalents.create({
		talentsString: '00530311231251031-32030302020103-001',
	}),
};
export const TALENTS_COMBAT_SS = {
	name: 'Combat SS',
	data: SavedTalents.create({
		talentsString: '00530320201-3200132202012321551',
	}),
};
export const TALENTS_COMBAT_BS = {
	name: 'Combat BS',
	data: SavedTalents.create({
		talentsString: '005303002-3203030202012321351-005',
	}),
};
export const TALENTS_SUB_HEMO = {
	name: 'Subtlety Hemo',
	data: SavedTalents.create({
		talentsString: '00530301-320003-005233031321001151',
	}),
};
export const TALENTS_COMBAT_VENOM = {
	name: 'Combat Venom',
	data: SavedTalents.create({
		talentsString: '00530320241001-320013220201232115',
	}),
};
export const TALENTS_HEMO_SF = {
	name: 'Hemo Seal Fate',
	data: SavedTalents.create({
		talentsString: '0053031121135102-02-0052030313210001',
	}),
};

export const DEFAULT_APL = PresetUtils.makePresetAPLRotation('Direct APL', DefaultAPL);

export const DEFAULT_SIMPLE_ROTATION = RogueRotation.create({ useBackstab: false, useExposeArmor: false, startInStealth: false });
export const SIMPLE_ROTATION = PresetUtils.makePresetSimpleRotation('Configurable', Spec.SpecRogue, DEFAULT_SIMPLE_ROTATION);
const SIMPLE_BACKSTAB_BUILD_ROTATION = PresetUtils.makePresetSimpleRotation(
	'Configurable',
	Spec.SpecRogue,
	RogueRotation.create({ useBackstab: true, useExposeArmor: false }),
);

export const BACKSTAB_GEAR = PresetUtils.makePresetGear('Backstab Combat', BackstabGear);
export const MUTI_GEAR = PresetUtils.makePresetGear('Mutilate', MutiGear);
export const SINISTER_GEAR = PresetUtils.makePresetGear('Slow Main Hand', SinisterGear);
export const HEMO_GEAR = PresetUtils.makePresetGear('Backstab Subtlety', HemoGear);

// Preset options for EP weights
export const EP_PRESET_COMBAT_SINISTER = PresetUtils.makePresetEpWeights(
	'Combat Sinister',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 0.01,
			[Stat.StatAgility]: 1.96,
			[Stat.StatStrength]: 1.1,
			[Stat.StatCritRating]: 1.7,
			[Stat.StatHasteRating]: 1.72,
			[Stat.StatHitRating]: 2.42,
			[Stat.StatExpertiseRating]: 2.55,
			[Stat.StatArmorPenetration]: 0.41,
			[Stat.StatSpellPiercing]: 0.19,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatPhysicalDamage]: 7.97,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 10.57,
			[PseudoStat.PseudoStatOffHandDps]: 2.93,
		},
	),
);

export const EP_PRESET_COMBAT_DAGS = PresetUtils.makePresetEpWeights(
	'Combat Daggers',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 0.01,
			[Stat.StatAgility]: 1.95,
			[Stat.StatStrength]: 1.1,
			[Stat.StatCritRating]: 1.68,
			[Stat.StatHasteRating]: 1.7,
			[Stat.StatHitRating]: 2.35,
			[Stat.StatExpertiseRating]: 2.21,
			[Stat.StatArmorPenetration]: 0.43,
			[Stat.StatSpellPiercing]: 0.21,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatPhysicalDamage]: 9.9,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 10.18,
			[PseudoStat.PseudoStatOffHandDps]: 2.97,
		},
	),
);

export const EP_PRESET_SUBTLETY = PresetUtils.makePresetEpWeights(
	'Subtlety',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 0.01,
			[Stat.StatAgility]: 1.93,
			[Stat.StatStrength]: 1.1,
			[Stat.StatCritRating]: 1.63,
			[Stat.StatHasteRating]: 1.43,
			[Stat.StatHitRating]: 2.12,
			[Stat.StatExpertiseRating]: 2.42,
			[Stat.StatArmorPenetration]: 0.33,
			[Stat.StatSpellPiercing]: 0.16,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatPhysicalDamage]: 6.83,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 10.47,
			[PseudoStat.PseudoStatOffHandDps]: 2.27,
		},
	),
);

export const EP_PRESET_ASSASSINATION = PresetUtils.makePresetEpWeights(
	'Assassination',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 0.01,
			[Stat.StatAgility]: 2.06,
			[Stat.StatStrength]: 1.1,
			[Stat.StatCritRating]: 2.02,
			[Stat.StatHasteRating]: 1.93,
			[Stat.StatHitRating]: 2.8,
			[Stat.StatExpertiseRating]: 2.77,
			[Stat.StatArmorPenetration]: 0.4,
			[Stat.StatSpellPiercing]: 0.65,
			[Stat.StatAttackPower]: 1.0,
			[Stat.StatPhysicalDamage]: 10.55,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 8.44,
			[PseudoStat.PseudoStatOffHandDps]: 4.16,
		},
	),
);

export const DefaultOptions = RogueOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.ElixirOfGreaterFortitude.id, Consumables.ElixirOfGreaterStrength.id],
	extraIds: [Consumables.JujuMight.id, Consumables.GroundScorpokAssay.id, Consumables.DragonbreathChili.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	offensivePotId: Consumables.MajorFrenzyPotion.id,
	conjuredId: Consumables.ThistleTea.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	ohImbueId: Consumables.DenseSharpeningStone.id,
	mhWeaponImbueId: Consumables.InstantPoisonVI.id,
	ohWeaponImbueId: Consumables.DeadlyPoisonIV.id,
});

const FlippedPoisons = ConsumesSpec.create({
	...DefaultConsumables,
	mhWeaponImbueId: Consumables.DeadlyPoisonIV.id,
	ohWeaponImbueId: Consumables.InstantPoisonVI.id,
});

const IP_MH_Settings = {
	name: 'Default',
	consumables: DefaultConsumables,
};
const _DP_MH_Settings = {
	name: 'Flipped Poisons',
	consumables: FlippedPoisons,
};

export const PRESET_BUILD_MUTILATE_ASN = PresetUtils.makePresetBuild('Asn Muti', {
	talents: TALENTS_MUTILATE_ASN,
	rotation: SIMPLE_ROTATION,
	gear: MUTI_GEAR,
	settings: IP_MH_Settings,
	epWeights: EP_PRESET_ASSASSINATION,
});

export const PRESET_BUILD_COMBAT_SS = PresetUtils.makePresetBuild('Combat SS', {
	talents: TALENTS_COMBAT_SS,
	rotation: SIMPLE_ROTATION,
	gear: SINISTER_GEAR,
	settings: IP_MH_Settings,
	epWeights: EP_PRESET_COMBAT_SINISTER,
});

export const PRESET_BUILD_COMBAT_BS = PresetUtils.makePresetBuild('Combat BS', {
	talents: TALENTS_COMBAT_BS,
	rotation: SIMPLE_BACKSTAB_BUILD_ROTATION,
	gear: BACKSTAB_GEAR,
	settings: IP_MH_Settings,
	epWeights: EP_PRESET_COMBAT_DAGS,
});

export const PRESET_BUILD_SUB_HEMO = PresetUtils.makePresetBuild('Subtlety Hemo', {
	talents: TALENTS_SUB_HEMO,
	rotation: SIMPLE_ROTATION,
	gear: HEMO_GEAR,
	settings: IP_MH_Settings,
	epWeights: EP_PRESET_SUBTLETY,
});

export const PRESET_BUILD_COM_VEN = PresetUtils.makePresetBuild('Combat Venom', {
	talents: TALENTS_COMBAT_VENOM,
	rotation: SIMPLE_ROTATION,
	gear: SINISTER_GEAR,
	settings: IP_MH_Settings,
	epWeights: EP_PRESET_COMBAT_SINISTER,
});

export const PRESET_BUILD_HEMO_SF = PresetUtils.makePresetBuild('Hemo Seal Fate', {
	talents: TALENTS_HEMO_SF,
	rotation: SIMPLE_ROTATION,
	gear: HEMO_GEAR,
	settings: IP_MH_Settings,
	epWeights: EP_PRESET_SUBTLETY,
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
