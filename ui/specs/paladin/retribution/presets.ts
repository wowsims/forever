import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, TristateEffect } from '@generated/proto/common';
import { RetributionPaladin_Options as RetributionPaladinOptions } from '@generated/proto/paladin';

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

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MajorFrenzyPotion.id,
	flaskId: Consumables.FlaskOfNaturalAggression.id,
	foodId: Consumables.SmokedDesertDumplings.id,
	conjuredId: Consumables.DemonicRune.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfAgilityIV.id, Consumables.ScrollOfStrengthIV.id],
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
	leaderOfThePack: true,
	battleShout: TristateEffect.TristateEffectRegular,
	strengthOfEarthTotem: true,
	windfuryTotem: true,
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
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: true,
	exposeArmor: true,
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceHuman,
};
