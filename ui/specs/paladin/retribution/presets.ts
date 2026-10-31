import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, PseudoStat, Race, Spec, Stat, TristateEffect } from '@generated/proto/common';
import {
	PaladinAura,
	PaladinJudgement,
	PaladinSeal,
	RetributionPaladin_Options as RetributionPaladinOptions,
	RetributionPaladin_Rotation as RetributionPaladinRotation,
} from '@generated/proto/paladin';
import { SavedTalents } from '@generated/proto/ui';
import * as Mechanics from '@sim/constants/mechanics';
import type { PresetSettings } from '@sim/presets/types';
import { Stats } from '@sim/proto/stats';

import DefaultApl from './apls/default.apl.json';
import HolyShockApl from './apls/holy_shock.apl.json';
import DefaultGear from './gear_sets/default.gear.json';
import HolyShockGear from './gear_sets/holy_shock.gear.json';
import HolyShockItemSwapGear from './gear_sets/holy_shock_item_swap.gear.json';

export const HolyShockAplPreset = PresetUtils.makePresetAPLRotation('Holy Shock', HolyShockApl);
export const DefaultAplPreset = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const DefaultSimpleRotation = RetributionPaladinRotation.create({
	aura: PaladinAura.RetributionAura,
	assignedJudgement: PaladinJudgement.JudgementOfTheCrusader,
	nonTwistSeal: PaladinSeal.SealOfCommand,
	pullRange: 25,
});
export const SimpleRotation = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecRetributionPaladin, DefaultSimpleRotation);

export const HolyShockGearPreset = PresetUtils.makePresetGear('Holy Shock', HolyShockGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});
export const DefaultGearPreset = PresetUtils.makePresetGear('Default', DefaultGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

// Shin Blade, which the Holy Shock APL swaps in while the Manual Crowd Pummeler is on cooldown.
export const HolyShockItemSwap = PresetUtils.makePresetItemSwapGear('Holy Shock', HolyShockItemSwapGear);

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {},
});

export const HolyShockTalents = PresetUtils.makePresetTalents('Holy Shock', SavedTalents.create({ talentsString: '550030032101012--052252300012303' }));
export const LightsVigilTalents = PresetUtils.makePresetTalents("Light's Vigil", SavedTalents.create({ talentsString: '55003003201121251--05225030201' }));
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
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 2.32 * Mechanics.HIT_RATING_PER_HIT_PERCENT,
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
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 0.6 * Mechanics.HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

export const DefaultConsumables = ConsumesSpec.create({
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.LobsterStew.id,
	conjuredId: Consumables.DemonicRune.id,
	extraIds: [Consumables.ROIDS.id, Consumables.GoblinSapperCharge.id, Consumables.DragonbreathChili.id, Consumables.GordokGreenGrog.id],
	elixirIds: [
		Consumables.ElixirOfHolyPower.id,
		Consumables.ElixirOfSages.id,
		Consumables.ElixirOfTheMongoose.id,
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfTheGrizzly.id,
		Consumables.ElixirOfThePhalanx.id,
		Consumables.GreaterArcaneElixir.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.JujuMight.id,
	],
	mhImbueId: Consumables.BrilliantWizardOil.id,
	swapMhImbueId: Consumables.BrilliantWizardOil.id,
});

export const HolyShockConsumables = ConsumesSpec.create({
	...DefaultConsumables,
	extraIds: [Consumables.CerebralCortexCompound.id, Consumables.GoblinSapperCharge.id, Consumables.DragonbreathChili.id, Consumables.GordokGreenGrog.id],
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	strengthOfEarthTotem: true,
	windfuryTotem: true,
	battleShout: 1, // Rank 6
	leaderOfThePack: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
	greaterBlessingOfMight: true,
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
	distanceFromTarget: 25,
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

export const DefaultSettings = settings('Default', DefaultConsumables);
export const HolyShockSettings = settings('Holy Shock', HolyShockConsumables, HolyShockItemSwap);

// The Simple rotation picks the Default or the Holy Shock APL from the talents.
export const DefaultBuild = PresetUtils.makePresetBuild('Twist of Light', {
	gear: DefaultGearPreset,
	talents: TwistOfLightTalents,
	rotationType: APLRotationType.TypeSimple,
	settings: DefaultSettings,
});

export const HolyShockBuild = PresetUtils.makePresetBuild('Holy Shock', {
	gear: HolyShockGearPreset,
	talents: HolyShockTalents,
	rotationType: APLRotationType.TypeSimple,
	settings: HolyShockSettings,
});
