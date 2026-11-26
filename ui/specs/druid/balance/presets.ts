import * as PresetUtils from '@app/preset_utils';
import { APLRotation_Type } from '@generated/proto/apl';
import { Spec, Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';

import DefaultsJson from './defaults.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecBalanceDruid, DefaultsJson);

export const StandardRotation = PresetUtils.makePresetAPLRotation('Default', DefaultsJson.player.rotation);

export const StandardTalents = PresetUtils.makePresetTalents('Moonkin', Defaults.talents);

export const DefaultGear = PresetUtils.makePresetGear('Pre-BiS', DefaultsJson.player.equipment);

// Starts on Auto as the page does, which runs the same rotation.
export const DefaultBuild = PresetUtils.makePresetBuildFromSiml('Default', Spec.SpecBalanceDruid, DefaultsJson, {
	gear: DefaultGear,
	rotationType: APLRotation_Type.TypeAuto,
});

// Measured with the sim's stat weights on the default build.
export const DefaultEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.19,
		[Stat.StatSpirit]: 0.04,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.63,
		[Stat.StatNatureDamage]: 0.37,
		[Stat.StatHitRating]: 1.14,
		[Stat.StatCritRating]: 0.63,
		[Stat.StatHasteRating]: 0.96,
		[Stat.StatMP5]: 0.09,
	}),
);
