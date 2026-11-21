import { makeBooleanDebuffInput, makeBooleanPartyBuffInput, makeMultistateRaidBuffInput } from '@features/inputs/icon_inputs';
import { Class, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { DEBUFF_ITEM_IDS } from '@sim/proto/buff_items';
import { Party } from '@sim/raid/party';

import * as Generated from './buffs_debuffs_auto_gen';
import { IconPickerStatOption, inDisplayOrder, statOptionGroup } from './stat_options';

// Every buff the game files resolve to a spell has its input generated from the manifest;
// the rows below are the ones it cannot produce, and the registries at the end of this file
// interleave the two. Buffs whose whole effect bids in one exclusive category in the sim, so that
// only the strongest applies, share a group.
export * from './buffs_debuffs_auto_gen';

// The generated const takes its name from the proto field; the specs name the buff.
export const Innervate = Generated.Innervates;
export const PowerInfusion = Generated.PowerInfusions;
export const ManaTideTotem = Generated.ManaTideTotems;

// Raid Buffs
// The raid's Battle Shout steps through the trainer's Rank 6, Rank 7 from Manual of Battle Shout VII (the first
// improved icon) and Rank 7 from a warrior wearing three pieces of Battlegear of Wrath (both improved icons).
export const BattleShout = {
	...makeMultistateRaidBuffInput({
		actionId: ActionId.fromSpellId(11551),
		numStates: 4,
		fieldName: 'battleShout',
		label: 'Battle Shout',
	}),
	improvedId: ActionId.fromItemId(21298),
	improvedId2: ActionId.fromSpellId(23563),
};
const BattleShoutOption: IconPickerStatOption = { config: BattleShout, stats: [Stat.StatAttackPower], ownerClass: Class.ClassWarrior };

// Debuffs
// The raid's Puncture Armor field holds the weapon that procs it. Each weapon is an icon of its own that writes its
// state, so picking one clears the other two.
const punctureArmorWeapon = (value: number, itemId: number, weapon: string) =>
	makeBooleanDebuffInput({
		actionId: ActionId.fromItemId(itemId),
		fieldName: 'punctureArmor',
		value,
		label: `Puncture Armor (${weapon})`,
	});
const PunctureArmorWeapons = ['Bashguuder', 'Rivenspike', 'Vibroblade'].map((weapon, index) =>
	punctureArmorWeapon(index + 1, DEBUFF_ITEM_IDS.punctureArmor![index], weapon),
);
const PunctureArmorOptions: IconPickerStatOption[] = PunctureArmorWeapons.map(config => ({
	config,
	stats: [Stat.StatAttackPower],
}));

// Party Buffs
export const RetributionAura = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(10301),
	fieldName: 'retributionAura',
	label: 'Retribution Aura',
	// A damage shield only matters on the unit being hit, so only tanks get to pick it. The spell
	// power it scales with is the RetributionAuraSpellPower other-input, shown under the same rule.
	showWhen: (party: Party) => !!party.getPlayer(0)?.getPlayerSpec().isTankSpec,
});

// Darkspear 12802 is not in the game, so the input shows only while a saved setting still picks it, which lets that pick be cleared.
export const Darkspear = makeBooleanPartyBuffInput({
	actionId: ActionId.fromSpellId(16939),
	fieldName: 'darkspear',
	label: 'Darkspear',
	showWhen: (party: Party) => party.getBuffs().darkspear,
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
	Generated.DarkInspiration,
	Generated.AtieshMage,
	Generated.AtieshWarlock,
	Generated.AtieshDruid,
	Generated.AtieshPriest,
	{ config: Darkspear, stats: [Stat.StatCritRating] },
	Generated.AdmiralsHat,
	Generated.HeadmastersCharge,
	Generated.StoneskinTotem,
]);

export const BUFFS_CONFIG = inDisplayOrder(
	[...Generated.GENERATED_RAID_BUFFS_CONFIG, BattleShoutOption, ...Generated.GENERATED_INDIVIDUAL_BUFFS_CONFIG],
	[
		Generated.ArcaneIntellect,
		Generated.GreaterBlessingOfKings,
		Generated.PrayerOfSpirit,
		Generated.MarkOfTheWild,
		Generated.Thorns,
		Generated.PowerWordFortitude,
		Generated.GreaterBlessingOfMight,
		Generated.GreaterBlessingOfWisdom,
		statOptionGroup(i18n.t('settings_tab.raid_buffs.shadow_resistance'), [Generated.ShadowProtection, Generated.ShadowResistanceAura]),
		statOptionGroup(i18n.t('settings_tab.raid_buffs.fire_resistance'), [Generated.FireResistanceAura, Generated.FireResistanceTotem]),
		statOptionGroup(i18n.t('settings_tab.raid_buffs.frost_resistance'), [Generated.FrostResistanceAura, Generated.FrostResistanceTotem]),
		statOptionGroup(i18n.t('settings_tab.raid_buffs.nature_resistance'), [Generated.NatureResistanceTotem, Generated.AspectOfTheWild]),
		Generated.StrengthOfEarthTotem,
		Generated.GraceOfAirTotem,
		Generated.WindfuryTotem,
		Generated.FlametongueTotem,
		statOptionGroup(
			i18n.t('settings_tab.raid_buffs.attack_power'),
			[BattleShout, Generated.FuriousHowl],
			i18n.t('settings_tab.raid_buffs.attack_power_tooltip'),
		),
		Generated.TrueshotAura,
		Generated.LeaderOfThePack,
		Generated.MoonkinAura,
		Generated.Innervates,
		Generated.PowerInfusions,
	],
);

export const DEBUFFS_CONFIG = inDisplayOrder(
	[...Generated.GENERATED_DEBUFFS_CONFIG, ...PunctureArmorOptions],
	[
		Generated.HuntersMark,
		Generated.JudgementOfTheCrusader,
		Generated.JudgementOfLight,
		Generated.JudgementOfWisdom,
		Generated.CurseOfElements,
		Generated.CurseOfWeakness,
		statOptionGroup(i18n.t('settings_tab.debuffs.major_armor_reduction'), [Generated.SunderArmor, Generated.ExposeArmor]),
		statOptionGroup(i18n.t('settings_tab.debuffs.minor_armor_reduction'), [
			Generated.FaerieFire,
			Generated.CurseOfRecklessness,
			Generated.DustCloud,
			Generated.ArmorShatter,
			...PunctureArmorWeapons,
		]),
		Generated.GiftOfArthas,
		statOptionGroup(i18n.t('settings_tab.debuffs.attack_power_reduction'), [
			Generated.DemoralizingShout,
			Generated.DemoralizingRoar,
			Generated.DemoralizingScreech,
			Generated.HexOfWeakness,
			Generated.Vindication,
		]),
		// Thunderfury's Cyclone slows attacks in Thunder Clap's category; its nature resistance cut applies on its own.
		statOptionGroup(i18n.t('settings_tab.debuffs.attack_speed_reduction'), [Generated.ThunderClap, Generated.Thunderfury]),
		statOptionGroup(i18n.t('settings_tab.debuffs.hit_chance_reduction'), [Generated.InsectSwarm, Generated.ScorpidSting]),
	],
);

export const DEBUFFS_MISC_CONFIG = [] as IconPickerStatOption[];
