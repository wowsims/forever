import { makeBooleanIndividualBuffInput, makeBooleanPartyBuffInput } from '@features/inputs/icon_inputs';
import { Class, Stat } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import { Party } from '@sim/raid/party';

import * as Generated from './buffs_debuffs_auto_gen';
import { IconPickerStatOption, inDisplayOrder } from './stat_options';

// Every buff the game files resolve to a spell has its input generated from the manifest;
// the rows below are the ones it cannot produce, and the registries at the end of this file
// interleave the two.
export * from './buffs_debuffs_auto_gen';

// The generated const takes its name from the proto field; the specs name the buff.
export const Innervate = Generated.Innervates;
export const PowerInfusion = Generated.PowerInfusions;
export const ManaTideTotem = Generated.ManaTideTotems;

// Individual Buffs
export const GreaterBlessingOfSalvation = makeBooleanIndividualBuffInput({
	actionId: ActionId.fromSpellId(25895),
	fieldName: 'greaterBlessingOfSalvation',
	label: 'Mass Salvation',
	showWhen: player => !player.getPlayerSpec().isTankSpec && !player.getPlayerSpec().isHealingSpec,
});
export const GreaterBlessingOfLight = makeBooleanIndividualBuffInput({
	actionId: ActionId.fromSpellId(25890),
	fieldName: 'greaterBlessingOfLight',
	label: 'Greater Blessing of Light',
	showWhen: player => player.getPlayerSpec().isHealingSpec || player.getPlayerSpec().isTankSpec,
});

// Party Buffs
export const RetributionAura = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(10301),
	fieldName: 'retributionAura',
	label: 'Retribution Aura',
	// A damage shield only matters on the unit being hit, so only tanks get to pick it. The spell
	// power it scales with is the RetributionAuraSpellPower other-input, shown under the same rule.
	showWhen: (party: Party) => !!party.getPlayer(0)?.getPlayerSpec().isTankSpec,
});

export const PARTY_BUFFS_CONFIG = inDisplayOrder(Generated.GENERATED_PARTY_BUFFS_CONFIG, [
	Generated.BloodPact,
	Generated.DevotionAura,
	Generated.ManaSpringTotem,
	Generated.ManaTideTotems,
	{
		config: RetributionAura,
		stats: [Stat.StatDefenseRating],
		ownerClass: Class.ClassPaladin,
	},
	Generated.ConcentrationAura,
	Generated.AtieshMage,
	Generated.AtieshWarlock,
	Generated.AtieshDruid,
	Generated.AtieshPriest,
	Generated.Darkspear,
	Generated.AdmiralsHat,
	Generated.StoneskinTotem,
]);

export const BUFFS_CONFIG = inDisplayOrder(
	[...Generated.GENERATED_RAID_BUFFS_CONFIG, ...Generated.GENERATED_INDIVIDUAL_BUFFS_CONFIG],
	[
		Generated.ArcaneBrilliance,
		Generated.GreaterBlessingOfKings,
		Generated.PrayerOfSpirit,
		Generated.GiftOfTheWild,
		Generated.Thorns,
		Generated.PrayerOfFortitude,
		Generated.GreaterBlessingOfMight,
		Generated.GreaterBlessingOfWisdom,
		{ config: GreaterBlessingOfSalvation, stats: [] },
		{ config: GreaterBlessingOfLight, stats: [Stat.StatHealingPower], ownerClass: Class.ClassPaladin },
		Generated.PrayerOfShadowProtection,
		Generated.FireResistanceAura,
		Generated.FrostResistanceAura,
		Generated.ShadowResistanceAura,
		Generated.FireResistanceTotem,
		Generated.FrostResistanceTotem,
		Generated.NatureResistanceTotem,
		Generated.StrengthOfEarthTotem,
		Generated.GraceOfAirTotem,
		Generated.WindfuryTotem,
		Generated.FlametongueTotem,
		Generated.BattleShout,
		Generated.TrueshotAura,
		Generated.FuriousHowl,
		Generated.AspectOfTheWild,
		Generated.LeaderOfThePack,
		Generated.MoonkinAura,
		Generated.Innervates,
		Generated.PowerInfusions,
	],
);

export const DEBUFFS_CONFIG = inDisplayOrder(Generated.GENERATED_DEBUFFS_CONFIG, [
	Generated.HuntersMark,
	Generated.JudgementOfTheCrusader,
	Generated.JudgementOfLight,
	Generated.JudgementOfWisdom,
	Generated.CurseOfElements,
	Generated.CurseOfRecklessness,
	Generated.CurseOfWeakness,
	Generated.FaerieFire,
	Generated.DustCloud,
	Generated.ExposeArmor,
	Generated.SunderArmor,
	Generated.ArmorShatter,
	Generated.GiftOfArthas,
	Generated.DemoralizingRoar,
	Generated.DemoralizingShout,
	Generated.HexOfWeakness,
	Generated.DemoralizingScreech,
	Generated.Vindication,
	Generated.ThunderClap,
	Generated.Thunderfury,
	Generated.InsectSwarm,
	Generated.ScorpidSting,
]);

export const DEBUFFS_MISC_CONFIG = [] as IconPickerStatOption[];
