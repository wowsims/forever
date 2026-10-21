import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession } from '@generated/proto/common';
import { HolyPaladin_Options as HolyPaladinOptions } from '@generated/proto/paladin';

import DefaultGear from './gear_sets/default.gear.json';

// Level 60 blues from dungeons, quests, reputations and crafting: a placeholder until Phase 1
// gear is known.
export const GEAR_DEFAULT = PresetUtils.makePresetGear('Default', DefaultGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

export const DefaultOptions = HolyPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	foodId: Consumables.SoaringPamplona.id,
	resourcePotId: Consumables.MajorManaPotion.id,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Enchanting,
};
