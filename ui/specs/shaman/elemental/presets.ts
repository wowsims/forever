import * as PresetUtils from '@app/preset_utils';
import { Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultBuildJson from './builds/default.build.json';

// The spec's defaults, from the Default build.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecElementalShaman, DefaultBuildJson);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultBuildJson.player.rotation);

export const DefaultGear = PresetUtils.makePresetGear('Default', DefaultBuildJson.player.equipment, { devOnly: true });

export const DefaultBuild = PresetUtils.makePresetBuildFromJSON('Default', Spec.SpecElementalShaman, DefaultBuildJson, { gear: DefaultGear });

// Measured on the earlier Testing gear with the Standard talents over a three minute fight. Hit is its value below the
// 17% cap.
export const DefaultEpWeights = PresetUtils.makePresetEpWeights(
	'Default',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.19,
		[Stat.StatSpirit]: 0.02,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatFireDamage]: 0.31,
		[Stat.StatNatureDamage]: 0.69,
		[Stat.StatSpellPiercing]: 1.81,
		[Stat.StatHitRating]: 1.18,
		[Stat.StatCritRating]: 0.61,
		[Stat.StatHasteRating]: 1.23,
		[Stat.StatMP5]: 0.04,
	}),
);

export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: DefaultBuildJson.player.talentsString,
	}),
};

export const ManaTideTalents = PresetUtils.makePresetTalents('Mana Tide', SavedTalents.create({ talentsString: '5500331013001351-04-053150001001' }));
