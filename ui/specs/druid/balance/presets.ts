import * as PresetUtils from '@app/preset_utils';
import { Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultAPL from './apls/default.apl.siml';
import DefaultBuildJson from './builds/default.build.json';
import DefaultsJson from './defaults.siml';
import P1GearJson from './gear_sets/p1.gear.siml';
import PreBiSGearJson from './gear_sets/prebis.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecBalanceDruid, DefaultsJson);

export const StandardRotation = PresetUtils.makePresetAPLRotation('Default', DefaultAPL);

export const StandardTalents = PresetUtils.makePresetTalents('Moonkin', Defaults.talents);

// Six Feralheart pieces sold in Orgrimmar and Ironforge, Orb of the Darkmoon for Darkmoon Faire tickets, Tearfall
// Bracers of Magic from Stratholme, Rune Band of Wizardry and Eye of the Beast from Blackrock Spire, Therazane's Touch
// from the Alterac Valley vendors and the crafted Ironfeather Leggings. The cloak, a ring, a trinket and the idol are
// sold by the Undermine Trader in Stranglethorn Vale. Hand of Edward the Odd, and the Pristine Scale and Fel Leather the
// leggings take, have no known source.
export const PreBiSGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSGearJson, { devOnly: true });

export const P1Gear = PresetUtils.makePresetGear('P1', P1GearJson, { devOnly: true });

export const DefaultBuildGear = PresetUtils.makePresetGear('Default', DefaultBuildJson.player.equipment, { devOnly: true });
export const DefaultBuildTalents = PresetUtils.makePresetTalents('Default', SavedTalents.create({ talentsString: DefaultBuildJson.player.talentsString }));
export const DefaultBuildRotation = PresetUtils.makePresetAPLRotation('Default Build', DefaultBuildJson.player.rotation);
export const DefaultBuild = PresetUtils.makePresetBuildFromJSON('Default', Spec.SpecBalanceDruid, DefaultBuildJson, { gear: DefaultBuildGear });

// Measured on the P1 gear with the Moonkin talents and the default consumables and buffs.
export const P1EpWeights = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.25,
		[Stat.StatSpirit]: 0.05,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.63,
		[Stat.StatNatureDamage]: 0.37,
		[Stat.StatHitRating]: 1.2,
		[Stat.StatCritRating]: 0.77,
		[Stat.StatHasteRating]: 1.03,
		[Stat.StatMP5]: 0.13,
	}),
);

// Measured on the Pre-BiS gear with the Moonkin talents and the default consumables and buffs.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.2,
		[Stat.StatSpirit]: 0.04,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.63,
		[Stat.StatNatureDamage]: 0.37,
		[Stat.StatHitRating]: 1.17,
		[Stat.StatCritRating]: 0.63,
		[Stat.StatHasteRating]: 0.97,
		[Stat.StatMP5]: 0.13,
	}),
);
