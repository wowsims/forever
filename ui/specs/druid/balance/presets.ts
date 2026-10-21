import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Stat, TristateEffect, UnitReference } from '@generated/proto/common';
import { BalanceDruid_Options as BalanceDruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import DefaultAPL from './apls/default.apl.json';
import P1GearJson from './gear_sets/p1.gear.json';
import PreBiSGearJson from './gear_sets/prebis.gear.json';

export const StandardRotation = PresetUtils.makePresetAPLRotation('Default', DefaultAPL);

export const StandardTalents = PresetUtils.makePresetTalents('Moonkin', SavedTalents.create({ talentsString: '5232220115501351--505003' }));

// Dungeon drops, world-drop greens with the "of Magic" suffix and crafted items whose reagents include no
// raid materials; no reputation or PvP rewards, and no Scourge Invasion drops. The neck is a Dire Maul
// quest reward. The cloak, the rings, the trinkets and the idol are sold by the Undermine Trader in
// Stranglethorn Vale. The staff, and the Pristine Scale and Fel Leather the leggings take, have no known source.
export const PreBiSGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSGearJson, { devOnly: true });

export const P1Gear = PresetUtils.makePresetGear('P1', P1GearJson, { devOnly: true });

// Measured on the P1 gear with the Moonkin talents and the default consumables and buffs.
export const P1EpWeights = PresetUtils.makePresetEpWeights(
	'P1',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.28,
		[Stat.StatSpirit]: 0,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.68,
		[Stat.StatNatureDamage]: 0.32,
		[Stat.StatHitRating]: 1.48,
		[Stat.StatCritRating]: 0.76,
		[Stat.StatHasteRating]: 0.99,
		[Stat.StatMP5]: 0.15,
	}),
);

// Measured on the Pre-BiS gear with the Moonkin talents and the default consumables and buffs.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS',
	Stats.fromMap({
		[Stat.StatIntellect]: 0.2,
		[Stat.StatSpirit]: 0,
		[Stat.StatSpellDamage]: 1,
		[Stat.StatArcaneDamage]: 0.68,
		[Stat.StatNatureDamage]: 0.32,
		[Stat.StatHitRating]: 1.45,
		[Stat.StatCritRating]: 0.76,
		[Stat.StatHasteRating]: 0.95,
		[Stat.StatMP5]: 0,
	}),
);

export const DefaultOptions = BalanceDruidOptions.create({
	classOptions: {
		innervateTarget: UnitReference.create(),
	},
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfSpirit: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfWisdom: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: Consumables.DemonicRune.id,
	elixirIds: [Consumables.GreaterArcaneElixir.id, Consumables.ElixirOfNaturePower.id, Consumables.ElixirOfTheOwl.id, Consumables.GreaterMagebloodElixir.id],
	explosiveId: Consumables.SAFTCleverDynamite.id,
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.LobsterStew.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
	resourcePotId: Consumables.MajorManaPotion.id,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Enchanting,
	profession2: Profession.Tailoring,
	race: Race.RaceNightElf,
};
