import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, Stat } from '@generated/proto/common';
import { ElementalShaman_Options as ElementalShamanOptions } from '@generated/proto/shaman';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import TestingBuildJson from './builds/testing.build.json';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', TestingBuildJson.player.rotation);

export const DefaultGear = PresetUtils.makePresetGear('Testing', TestingBuildJson.player.equipment);

export const TestingBuild = PresetUtils.makePresetBuildFromJSON('Testing', Spec.SpecElementalShaman, TestingBuildJson);

// Measured on the Testing gear with the Standard talents over a three minute fight. Hit is its value below the
// 17% cap.
export const DefaultEpWeights = PresetUtils.makePresetEpWeights(
	'Testing',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.17,
		[Stat.StatSpirit]: 0.05,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatFireDamage]: 0.32,
		[Stat.StatNatureDamage]: 0.68,
		[Stat.StatHitRating]: 1.16,
		[Stat.StatCritRating]: 0.59,
		[Stat.StatHasteRating]: 1.33,
		[Stat.StatMP5]: 0.04,
	}),
);

export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: TestingBuildJson.player.talentsString,
	}),
};

export const DefaultOptions = ElementalShamanOptions.create({
	classOptions: {
		shieldProcrate: 0,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Leatherworking,
	profession2: Profession.Enchanting,
	race: Race.RaceTroll,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfSpirit: true,
	moonkinAura: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	exposeArmor: true,
	faerieFire: true,
	giftOfArthas: true,
	judgementOfWisdom: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: Consumables.DemonicRune.id,
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.NightfinSoup.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
});
