import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as Presets from './presets';

export default defineSpec<Spec.SpecRetributionPaladin>({
	spec: Spec.SpecRetributionPaladin,

	className: 'retribution-paladin-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Paladin),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	consumableStats: [Stat.StatMana, Stat.StatHolyDamage],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatSpellPiercing,
		Stat.StatAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatHitRating,
		Stat.StatHasteRating,
		Stat.StatCritRating,
		Stat.StatExpertiseRating,
		Stat.StatMP5,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatSchoolHitPercentHoly],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatIntellect,
			Stat.StatAttackPower,
			Stat.StatSpellDamage,
			Stat.StatMP5,
			Stat.StatMana,
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatHolyDamage,
			Stat.StatSpellPiercing,
			Stat.StatArcaneResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatNatureResistance,
			Stat.StatShadowResistance,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatMeleeSpeedMultiplier,
			PseudoStat.PseudoStatSpellHastePercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSchoolHitPercentHoly,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee],

	defaults: {
		// Default equipped gear.
		gear: Presets.TwistOfLightGearPreset.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.TwistOfLightEpWeights.epWeights,
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8);
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);

			return hitCap.add(expCap);
		})(),
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.TwistOfLightTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,

		rotationType: APLRotation_Type.TypeAuto,
		aplRotation: Presets.TwistOfLightAplPreset.rotation.rotation!,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatDefenseRating, Stat.StatMP5, Stat.StatHolyDamage],
	excludeBuffDebuffInputs: [
		BuffDebuffInputs.AtieshDruid,
		BuffDebuffInputs.AtieshMage,
		BuffDebuffInputs.AtieshPriest,
		BuffDebuffInputs.AtieshWarlock,
		BuffDebuffInputs.ConcentrationAura,
		BuffDebuffInputs.CurseOfWeakness,
		BuffDebuffInputs.DemoralizingRoar,
		BuffDebuffInputs.DemoralizingScreech,
		BuffDebuffInputs.DemoralizingShout,
		BuffDebuffInputs.HexOfWeakness,
		BuffDebuffInputs.InsectSwarm,
		BuffDebuffInputs.ScorpidSting,
		BuffDebuffInputs.StoneskinTotem,
		BuffDebuffInputs.Thorns,
		BuffDebuffInputs.ThunderClap,
		BuffDebuffInputs.Thunderfury,
		BuffDebuffInputs.Vindication,
	],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.TankAssignment, OtherInputs.InFrontOfTarget],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.TwistOfLightEpWeights, Presets.HolyShockEpWeights],
		rotations: [Presets.TwistOfLightAplPreset, Presets.HolyShockAplPreset],
		// Preset talents that the user can quickly select.
		talents: [Presets.TwistOfLightTalents, Presets.HolyShockTalents],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.TwistOfLightGearPreset, Presets.HolySHockGearPreset],
	},

	autoRotation: (player: Player<Spec.SpecRetributionPaladin>): APLRotation => {
		return player.getTalents().holyShock ? Presets.HolyShockAplPreset.rotation.rotation! : Presets.TwistOfLightAplPreset.rotation.rotation!;
	},

	reforge: {
		getEPDefaults: (player: Player<Spec.SpecRetributionPaladin>): Stats => {
			return player.getTalents().holyShock ? Presets.HolyShockEpWeights.epWeights : Presets.TwistOfLightEpWeights.epWeights;
		},
	},
});
