import * as PresetUtils from '@app/preset_utils';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultBuildJson from './builds/default.build.json';

// The spec's defaults, from the Default build.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecEnhancementShaman, DefaultBuildJson);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultBuildJson.player.rotation);

export const DefaultGear = PresetUtils.makePresetGear('Default', DefaultBuildJson.player.equipment);

export const DefaultBuild = PresetUtils.makePresetBuildFromJSON('Default', Spec.SpecEnhancementShaman, DefaultBuildJson, { gear: DefaultGear });

export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: DefaultBuildJson.player.talentsString,
	}),
};

// Measured on the earlier Testing gear with the Standard talents over a three minute fight. The gear has 8.2% hit, so hit
// is measured with 40 hit rating taken off, for its value below the 8% cap.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.32,
			[Stat.StatIntellect]: 1.99,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 5.2,
			[Stat.StatCritRating]: 2.42,
			[Stat.StatHasteRating]: 2.81,
			[Stat.StatExpertiseRating]: 3.5,
			[Stat.StatArmorPenetration]: 0.37,
			[Stat.StatSpellDamage]: 1.27,
			[Stat.StatNatureDamage]: 0.72,
			[Stat.StatSpellPiercing]: 2.52,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 14.18,
		},
	),
);
