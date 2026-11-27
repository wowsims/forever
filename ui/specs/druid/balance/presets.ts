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
		[Stat.StatIntellect]: 0.23,
		[Stat.StatSpirit]: 0.02,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.62,
		[Stat.StatNatureDamage]: 0.38,
		[Stat.StatHitRating]: 1.15,
		[Stat.StatCritRating]: 0.72,
		[Stat.StatHasteRating]: 1,
		[Stat.StatMP5]: 0.08,
	}),
);
