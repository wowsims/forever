import * as PresetUtils from '@app/preset_utils';
import { Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultsJson from './defaults.siml';
import BudgetGearJson from './gear_sets/budget.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecElementalShaman, DefaultsJson);

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultsJson.player.rotation);

export const DefaultGear = PresetUtils.makePresetGear('Default', DefaultsJson.player.equipment);

export const BudgetGear = PresetUtils.makePresetGear('Budget', BudgetGearJson);

export const DefaultBuild = PresetUtils.makePresetBuildFromSiml('Default', Spec.SpecElementalShaman, DefaultsJson, { gear: DefaultGear });

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

export const StandardTalents = PresetUtils.makePresetTalents('Standard', Defaults.talents);

export const ManaTideTalents = PresetUtils.makePresetTalents('Mana Tide', SavedTalents.create({ talentsString: '5500331013001351-04-053150001001' }));
