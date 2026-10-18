import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { DpsPriest_Options as Options, PriestOptions_Armor } from '@generated/proto/priest';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import SmiteTestingBuildJson from './builds/smite_testing.build.json';
import ShadowTestingBuildJson from './builds/testing.build.json';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', ShadowTestingBuildJson.player.rotation);
export const ROTATION_SMITE_DEFAULT = PresetUtils.makePresetAPLRotation('Smite', SmiteTestingBuildJson.player.rotation);

export const ShadowTestingGear = PresetUtils.makePresetGear('Shadow Testing', ShadowTestingBuildJson.player.equipment, { devOnly: true });
export const SmiteTestingGear = PresetUtils.makePresetGear('Smite Testing', SmiteTestingBuildJson.player.equipment, { devOnly: true });

export const StandardTalents = PresetUtils.makePresetTalents('Standard', SavedTalents.create({ talentsString: ShadowTestingBuildJson.player.talentsString }));
export const ShadowTestingBuild = PresetUtils.makePresetBuildFromJSON('Shadow Testing', Spec.SpecDpsPriest, ShadowTestingBuildJson, {
	gear: ShadowTestingGear,
});

export const SmiteTalents = PresetUtils.makePresetTalents('Smite', SavedTalents.create({ talentsString: SmiteTestingBuildJson.player.talentsString }));
export const SmiteTestingBuild = PresetUtils.makePresetBuildFromJSON('Smite Testing', Spec.SpecDpsPriest, SmiteTestingBuildJson, { gear: SmiteTestingGear });

export const DefaultOptions = Options.create({
	classOptions: {
		armor: PriestOptions_Armor.InnerFire,
		preShadowform: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.NightfinSoup.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
	resourcePotId: Consumables.MajorManaPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	elixirIds: [
		Consumables.ElixirOfShadowPower.id,
		Consumables.GreaterArcaneElixir.id,
		Consumables.ElixirOfCunning.id,
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfSages.id,
		Consumables.GreaterMagebloodElixir.id,
	],
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfSpirit: true,
	moonkinAura: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
});

export const DefaultDebuffs = Debuffs.create({
	judgementOfTheCrusader: true,
	judgementOfWisdom: true,
	curseOfElements: true,
});

export const OtherDefaults = {
	channelClipDelay: 100,
	distanceFromTarget: 20,
	profession1: Profession.Enchanting,
	profession2: Profession.Tailoring,
	race: Race.RaceUndead,
	reactionTime: 100,
};

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
