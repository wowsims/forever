import * as PresetUtils from '@app/preset_utils';
import { Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultApl from './apls/default.apl.siml';
import SmiteTestingBuildJson from './builds/smite_testing.build.json';
import DefaultsJson from './defaults.siml';
import DefaultGear from './gear_sets/default.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecDpsPriest, DefaultsJson);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const ROTATION_SMITE_DEFAULT = PresetUtils.makePresetAPLRotation('Smite', SmiteTestingBuildJson.player.rotation);

export const ShadowGear = PresetUtils.makePresetGear('Shadow', DefaultGear, { devOnly: true });
export const SmiteTestingGear = PresetUtils.makePresetGear('Smite Testing', SmiteTestingBuildJson.player.equipment, { devOnly: true });

export const StandardTalents = PresetUtils.makePresetTalents('Standard', Defaults.talents);
export const ShadowBuild = PresetUtils.makePresetBuildFromSiml('Shadow', Spec.SpecDpsPriest, DefaultsJson, { gear: ShadowGear });

export const SmiteTalents = PresetUtils.makePresetTalents('Smite', SavedTalents.create({ talentsString: SmiteTestingBuildJson.player.talentsString }));
export const SmiteTestingBuild = PresetUtils.makePresetBuildFromJSON('Smite Testing', Spec.SpecDpsPriest, SmiteTestingBuildJson, { gear: SmiteTestingGear });

export const ShadowEpWeights = PresetUtils.makePresetEpWeights(
	'Shadow',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.24,
		[Stat.StatSpellDamage]: 1.0,
		[Stat.StatShadowDamage]: 1.0,
		[Stat.StatSpirit]: 0.01,
		[Stat.StatHitRating]: 1.14,
		[Stat.StatCritRating]: 0.84,
		[Stat.StatHasteRating]: 0.73,
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
