import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel, Profession, TristateEffect } from '@generated/proto/common';
import { ProtectionPaladin_Options as ProtectionPaladinOptions } from '@generated/proto/paladin';

import DefaultApl from './apls/default.apl.json';
import DefaultGear from './gear_sets/default.gear.json';

// Seal of Righteousness, Judgement and Holy Strike on cooldown. A rotation worth the name waits
// for the Forever numbers above level 20.
export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

// Level 60 blues from dungeons, quests, reputations and crafting: a placeholder until Phase 1
// gear is known.
export const GEAR_DEFAULT = PresetUtils.makePresetGear('Default', DefaultGear, {
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

export const DefaultOptions = ProtectionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.NightfinSoup.id,
	potId: Consumables.GreaterStoneshieldPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id, Consumables.ScrollOfAgilityIV.id, Consumables.ScrollOfProtectionIV.id],
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfShadowProtection: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	graceOfAirTotem: false,
	strengthOfEarthTotem: true,
	windfuryTotem: false,
	battleShout: TristateEffect.TristateEffectMissing,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	thorns: true,
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
	greaterBlessingOfMight: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfWisdom: true,
	judgementOfLight: true,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: true,
	exposeArmor: true,
	insectSwarm: true,
});

export const OtherDefaults = {
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
