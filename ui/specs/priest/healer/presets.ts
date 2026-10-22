import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, Stat } from '@generated/proto/common';
import { HealerPriest_Options as HealerPriestOptions, PriestOptions_Armor } from '@generated/proto/priest';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DiscTestingBuildJson from './builds/disc_testing.build.json';
import HolyTestingBuildJson from './builds/holy_testing.build.json';
import PreraidGearJson from './gear_sets/preraid.gear.json';

export const DefaultOptions = HealerPriestOptions.create({
	prayerOfMendingJumps: 5,
	prayerOfMendingDelay: 2.5,
	holyNovaHealTargets: 5,
	classOptions: {
		armor: PriestOptions_Armor.InnerFire,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	foodId: Consumables.SunnyTea.id,
	mhImbueId: Consumables.BrilliantManaOil.id,
	resourcePotId: Consumables.MajorManaPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	elixirIds: [
		Consumables.ElixirOfTheGrizzly.id,
		Consumables.ElixirOfCunning.id,
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfSages.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.GreaterClericsElixir.id,
	],
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Tailoring,
	profession2: Profession.Enchanting,
};

export const ROTATION_DISC_DEFAULT = PresetUtils.makePresetAPLRotation('Disc Testing', DiscTestingBuildJson.player.rotation);
export const ROTATION_HOLY_DEFAULT = PresetUtils.makePresetAPLRotation('Holy Testing', HolyTestingBuildJson.player.rotation);

export const DiscTalents = PresetUtils.makePresetTalents('Disc', SavedTalents.create({ talentsString: '005005031330011531-335020030301' }));
export const HolyTalents = PresetUtils.makePresetTalents('Holy', SavedTalents.create({ talentsString: '00500502103-33505003030121531' }));

export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreraidGearJson, { devOnly: true });

export const DiscSingleTargetEpWeights = PresetUtils.makePresetEpWeights(
	'Disc Single-Target',
	Stats.fromMap({
		[Stat.StatHealingPower]: 1.0,
		[Stat.StatCritRating]: 0.7,
		[Stat.StatHasteRating]: 0.15,
		[Stat.StatIntellect]: 0.2,
		[Stat.StatSpirit]: 0.15,
		[Stat.StatMP5]: 0.4,
	}),
);

export const DiscRaidHealingEpWeights = PresetUtils.makePresetEpWeights(
	'Disc Raid Healing',
	Stats.fromMap({
		[Stat.StatHealingPower]: 1.0,
		[Stat.StatCritRating]: 0.48,
		[Stat.StatHasteRating]: 0.1,
		[Stat.StatIntellect]: 0.2,
		[Stat.StatSpirit]: 0.1,
		[Stat.StatMP5]: 0.19,
	}),
);

export const HolySingleTargetEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Single-Target',
	Stats.fromMap({
		[Stat.StatHealingPower]: 1.0,
		[Stat.StatCritRating]: 0.3,
		[Stat.StatHasteRating]: 0.5,
		[Stat.StatIntellect]: 0.35,
		[Stat.StatSpirit]: 0.3,
		[Stat.StatMP5]: 0.6,
	}),
);

export const HolyRaidHealingEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Raid Healing',
	Stats.fromMap({
		[Stat.StatHealingPower]: 1.0,
		[Stat.StatCritRating]: 0.23,
		[Stat.StatHasteRating]: 0.48,
		[Stat.StatIntellect]: 0.3,
		[Stat.StatSpirit]: 0.25,
		[Stat.StatMP5]: 0.6,
	}),
);
