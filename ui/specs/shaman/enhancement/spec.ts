import * as OtherInputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import * as Mechanics from '@sim/constants/mechanics';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as ShamanInputs from '../shared/inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecEnhancementShaman>({
	spec: Spec.SpecEnhancementShaman,

	className: 'enhancement-shaman-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Shaman),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatIntellect,
		Stat.StatAttackPower,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatExpertiseRating,
		Stat.StatArmorPenetration,
		Stat.StatSpellDamage,
		Stat.StatNatureDamage,
		Stat.StatSpellPiercing,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAttackPower,
	consumableStats: [Stat.StatMana, Stat.StatMP5],
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatIntellect,
			Stat.StatMana,
			Stat.StatAttackPower,
			Stat.StatArmorPenetration,
			Stat.StatSpellDamage,
			Stat.StatNatureDamage,
			Stat.StatSpellPiercing,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee],

	defaults: {
		// Default equipped gear.
		gear: Presets.DefaultGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.PreBiSEpWeights.epWeights,
		statCaps: (() => {
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
			return expCap;
		})(),
		softCapBreakpoints: (() => {
			const meleeHitSoftCapConfig = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
				breakpoints: [8, 20, 27],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [1.9 * Mechanics.HIT_RATING_PER_HIT_PERCENT, 1.73 * Mechanics.HIT_RATING_PER_HIT_PERCENT, 0],
			});

			return [meleeHitSoftCapConfig];
		})(),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.StandardTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [ShamanInputs.ShamanImbueMH(), ShamanInputs.ShamanImbueMHSwap()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5, Stat.StatFireDamage],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [ShamanInputs.ShamanShieldProcrate(), OtherInputs.InputDelay, OtherInputs.TankAssignment, OtherInputs.InFrontOfTarget],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.PreBiSEpWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.StandardTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.ROTATION_PRESET_DEFAULT],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.DefaultGear],
		// Preset builds that the user can quickly select.
		builds: [Presets.TestingBuild],
	},

	autoRotation: (_: Player<Spec.SpecEnhancementShaman>): APLRotation => {
		return Presets.ROTATION_PRESET_DEFAULT.rotation.rotation!;
	},

	reforge: {},
});
