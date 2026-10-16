import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, PseudoStat, Race, Stat, TristateEffect } from '@generated/proto/common';
import { RetributionPaladin_Options as RetributionPaladinOptions } from '@generated/proto/paladin';
import { SavedTalents } from '@generated/proto/ui';
import type { PresetSettings } from '@sim/presets/types';
import { Stats } from '@sim/proto/stats';

import HolyShockApl from './apls/holy_shock.apl.json';
import TwistOfLightApl from './apls/twist_of_light.apl.json';
import HolyShockGear from './gear_sets/holy_shock.gear.json';
import HolyShockItemSwapGear from './gear_sets/holy_shock_item_swap.gear.json';
import TwistOfLightGear from './gear_sets/twist_of_light.gear.json';

export const HolyShockAplPreset = PresetUtils.makePresetAPLRotation('Holy Shock', HolyShockApl);
export const TwistOfLightAplPreset = PresetUtils.makePresetAPLRotation('Twist of Light', TwistOfLightApl);

export const HolyShockGearPreset = PresetUtils.makePresetGear('Holy Shock', HolyShockGear, {
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});
export const TwistOfLightGearPreset = PresetUtils.makePresetGear('Twist of Light', TwistOfLightGear, {
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

// Shin Blade, which the Holy Shock APL swaps in while the Manual Crowd Pummeler is on cooldown.
export const HolyShockItemSwap = PresetUtils.makePresetItemSwapGear('Holy Shock', HolyShockItemSwapGear);

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {},
});

export const HolyShockTalents = PresetUtils.makePresetTalents('Holy Shock', SavedTalents.create({ talentsString: '550030032101012--052252300012303' }));
export const TwistOfLightTalents = PresetUtils.makePresetTalents('Twist of Light', SavedTalents.create({ talentsString: '55003003--05225331101330321' }));

export const HolyShockEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Shock',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.5,
			[Stat.StatAgility]: 3.11,
			[Stat.StatIntellect]: 4.53,
			[Stat.StatSpellDamage]: 5.78,
			[Stat.StatSpellPiercing]: 0.01,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 10.25,
			[Stat.StatCritRating]: 4.66,
			[Stat.StatHasteRating]: 4.06,
			[Stat.StatArmorPenetration]: 0.26,
			[Stat.StatExpertiseRating]: 6.2,
			[Stat.StatMP5]: 0.01,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.04,
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 2.32,
		},
	),
);

export const TwistOfLightEpWeights = PresetUtils.makePresetEpWeights(
	'Twist of Light',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.4,
			[Stat.StatAgility]: 2.23,
			[Stat.StatIntellect]: 2.06,
			[Stat.StatSpellDamage]: 2.36,
			[Stat.StatSpellPiercing]: 0.01,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 6.83,
			[Stat.StatCritRating]: 3.08,
			[Stat.StatHasteRating]: 22.87,
			[Stat.StatArmorPenetration]: 0.24,
			[Stat.StatExpertiseRating]: 5.1,
			[Stat.StatMP5]: 0.23,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.85,
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 0.6,
		},
	),
);

export const TwistOfLightConsumables = ConsumesSpec.create({
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.LobsterStew.id,
	conjuredId: Consumables.DemonicRune.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ROIDS.id],
	elixirIds: [
		Consumables.ElixirOfHolyPower.id,
		Consumables.ElixirOfSages.id,
		Consumables.ElixirOfTheMongoose.id,
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfTheGrizzly.id,
		Consumables.GreaterArcaneElixir.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.JujuMight.id,
	],
	mhImbueId: Consumables.BrilliantWizardOil.id,
});

export const HolyShockConsumables = ConsumesSpec.create({
	...TwistOfLightConsumables,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.CerebralCortexCompound.id],
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	strengthOfEarthTotem: true,
	windfuryTotem: true,
	battleShout: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
	greaterBlessingOfMight: true,
	greaterBlessingOfSalvation: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfLight: true,
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
	reactionTime: 100,
	iterationCount: 25000,
	race: Race.RaceHuman,
};

const settings = (name: string, consumables: ConsumesSpec, itemSwap?: PresetUtils.PresetItemSwap): PresetSettings => ({
	name,
	race: OtherDefaults.race,
	consumables,
	raidBuffs: DefaultRaidBuffs,
	partyBuffs: DefaultPartyBuffs,
	buffs: DefaultIndividualBuffs,
	debuffs: DefaultDebuffs,
	playerOptions: {
		profession1: OtherDefaults.profession1,
		profession2: OtherDefaults.profession2,
		distanceFromTarget: OtherDefaults.distanceFromTarget,
		reactionTimeMs: OtherDefaults.reactionTime,
		enableItemSwap: !!itemSwap,
		itemSwap: itemSwap?.itemSwap,
	},
});

export const TwistOfLightSettings = settings('Twist of Light', TwistOfLightConsumables);
export const HolyShockSettings = settings('Holy Shock', HolyShockConsumables, HolyShockItemSwap);

// The Auto rotation picks the Twist of Light or the Holy Shock APL from the talents.
export const TwistOfLightBuild = PresetUtils.makePresetBuild('Twist of Light', {
	gear: TwistOfLightGearPreset,
	talents: TwistOfLightTalents,
	rotationType: APLRotationType.TypeAuto,
	settings: TwistOfLightSettings,
});

export const HolyShockBuild = PresetUtils.makePresetBuild('Holy Shock', {
	gear: HolyShockGearPreset,
	talents: HolyShockTalents,
	rotationType: APLRotationType.TypeAuto,
	settings: HolyShockSettings,
});
