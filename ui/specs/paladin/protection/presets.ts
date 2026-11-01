import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel, Profession, Race, TristateEffect } from '@generated/proto/common';
import { ProtectionPaladin_Options as ProtectionPaladinOptions } from '@generated/proto/paladin';
import { SavedTalents } from '@generated/proto/ui';

import DefaultApl from './apls/default.apl.json';
import DefaultGear from './gear_sets/default.gear.json';

// Retribution Aura and Seal of the Crusader before the pull, then Judgement, Swift Judgement, Seal
// of Fury, Consecration, Holy Shock while Consecrated Ground is on the target, Holy Shield, Holy
// Strike and Exorcism.
export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Level 60 blues from dungeons, quests, reputations and crafting: a placeholder until Phase 1
// gear is known.
export const GEAR_DEFAULT = PresetUtils.makePresetGear('Default', DefaultGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

export const DefaultTalents = PresetUtils.makePresetTalents('Default', SavedTalents.create({ talentsString: '050050002-05302102213301531-502' }));

export const DefaultOptions = ProtectionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.BakedSalmon.id,
	defensivePotId: Consumables.GreaterStoneshieldPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.JujuMight.id, Consumables.ScrollOfStaminaIV.id],
	elixirIds: [
		Consumables.ElixirOfHolyPower.id,
		Consumables.GreaterArcaneElixir.id,
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfSages.id,
		Consumables.ElixirOfGreaterFortitude.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.ElixirOfGreaterDefense.id,
		Consumables.ElixirOfFerocity.id,
		Consumables.ElixirOfTheMongoose.id,
	],
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: 1,
	giftOfTheWild: 1,
	prayerOfFortitude: 1,
	prayerOfShadowProtection: 1,
	windfuryTotem: true,
	graceOfAirTotem: 1,
	battleShout: 1, // Rank 6
	leaderOfThePack: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	thorns: true,
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: 1,
	greaterBlessingOfMight: 1,
});

export const DefaultDebuffs = Debuffs.create({
	judgementOfTheCrusader: true,
	curseOfElements: true,
	judgementOfWisdom: true,
	judgementOfLight: true,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: true,
	exposeArmor: true,
	giftOfArthas: true,
	demoralizingShout: true,
	insectSwarm: true,
});

export const OtherDefaults = {
	race: Race.RaceHuman,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	distanceFromTarget: 5,
	iterationCount: 25000,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
