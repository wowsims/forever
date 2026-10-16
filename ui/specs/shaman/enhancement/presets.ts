import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, PseudoStat, Race, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanImbue } from '@generated/proto/shaman';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultApl from './apls/default.apl.json';
import TestingBuildJson from './builds/testing.build.json';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const DefaultGear = PresetUtils.makePresetGear('Testing', TestingBuildJson.player.equipment);

export const TestingBuild = PresetUtils.makePresetBuildFromJSON('Testing', Spec.SpecEnhancementShaman, TestingBuildJson);

export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '053503102-053030031005112251',
	}),
};

// Measured on the Pre-BiS gear with the Standard talents over a three minute fight. Hit is its value below the
// 8% cap.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.82,
			[Stat.StatIntellect]: 2.72,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 5.36,
			[Stat.StatCritRating]: 3.13,
			[Stat.StatHasteRating]: 2.51,
			[Stat.StatExpertiseRating]: 3.66,
			[Stat.StatArmorPenetration]: 0.5,
			[Stat.StatSpellDamage]: 1.15,
			[Stat.StatNatureDamage]: 0.63,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 14.2,
		},
	),
);

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: true,
});

// Windfury Weapon on the main hand outbids the party's Windfury Totem, so the default is Rockbiter Weapon.
export const DefaultOptions = EnhancementShamanOptions.create({
	classOptions: {
		shieldProcrate: 0,
		imbueMh: ShamanImbue.RockbiterWeapon,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
};

export const DefaultConsumables = ConsumesSpec.create({
	resourcePotId: Consumables.MajorManaPotion.id,
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.SmokedDesertDumplings.id,
	conjuredId: Consumables.DemonicRune.id,
	elixirIds: [
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfSages.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.ElixirOfNaturePower.id,
		Consumables.GreaterArcaneElixir.id,
		Consumables.WinterfallFirewater.id,
		Consumables.ElixirOfCunning.id,
		Consumables.ElixirOfFerocity.id,
	],
	explosiveId: Consumables.EZThroDarkBomb.id,
	mhImbueId: Consumables.ElementalSharpeningStone.id,
	extraIds: [Consumables.GoblinSapperCharge.id],
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfFortitude: true,
	giftOfTheWild: true,
	arcaneBrilliance: true,
	strengthOfEarthTotem: true,
	windfuryTotem: true,
	battleShout: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
});

export const DefaultDebuffs = Debuffs.create({
	huntersMark: true,
	judgementOfWisdom: true,
	giftOfArthas: true,
	exposeArmor: true,
	faerieFire: true,
	sunderArmor: true,
	curseOfElements: true,
	curseOfRecklessness: true,
});
