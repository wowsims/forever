import * as PresetUtils from '@app/preset_utils';
import { Spec, Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';

import DefaultsJson from './defaults.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecDpsPriest, DefaultsJson);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultsJson.player.rotation);

export const ShadowGear = PresetUtils.makePresetGear('Shadow', DefaultsJson.player.equipment);

export const StandardTalents = PresetUtils.makePresetTalents('Standard', Defaults.talents);
export const ShadowBuild = PresetUtils.makePresetBuildFromSiml('Shadow', Spec.SpecDpsPriest, DefaultsJson, { gear: ShadowGear });

// Measured with the sim's stat weights on the defaults.siml build. Haste rating loses damage on it.
export const ShadowEpWeights = PresetUtils.makePresetEpWeights(
	'Shadow',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.26,
		[Stat.StatSpellDamage]: 1.0,
		[Stat.StatShadowDamage]: 1.0,
		[Stat.StatSpellPiercing]: 2.49,
		[Stat.StatHitRating]: 1.2,
		[Stat.StatCritRating]: 0.99,
		[Stat.StatHasteRating]: -0.09,
		[Stat.StatMP5]: 0.01,
	}),
);

export const SmiteEpWeights = PresetUtils.makePresetEpWeights(
	'Smite',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.58,
		[Stat.StatSpellDamage]: 1.0,
		[Stat.StatHolyDamage]: 0.81,
		[Stat.StatShadowDamage]: 0.19,
		[Stat.StatSpirit]: 0.57,
		[Stat.StatHitRating]: 0.37,
		[Stat.StatCritRating]: 0.32,
		[Stat.StatHasteRating]: 0.43,
		[Stat.StatMP5]: 1.46,
	}),
);
