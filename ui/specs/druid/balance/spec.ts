import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as BalanceInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecBalanceDruid>({
	spec: Spec.SpecBalanceDruid,

	className: 'balance-druid-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Druid),
	// List any known bugs / issues here, and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatArcaneDamage,
		Stat.StatNatureDamage,
		Stat.StatSpellPiercing,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatMP5,
		Stat.StatSpirit,
	],
	epPseudoStats: [],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellDamage,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpellDamage,
			Stat.StatArcaneDamage,
			Stat.StatNatureDamage,
			Stat.StatSpellPiercing,
			Stat.StatMP5,
			Stat.StatSpirit,
		],
		[PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent, PseudoStat.PseudoStatSpellHitPercent],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.PreBiSGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.P1EpWeights.epWeights,
		// Default stat caps for stat weights tab. (also needed for reforging since we don't want to reforge above stat caps)
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSpellHitPercent, 17);
		})(),
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
		other: Presets.OtherDefaults,
		rotationType: APLRotation_Type.TypeAuto,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	playerInputs: {
		inputs: [BalanceInputs.InnervateAssignmentInput],
	},
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.DistanceFromTarget, OtherInputs.TankAssignment],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.P1EpWeights, Presets.PreBiSEpWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.StandardTalents],
		rotations: [Presets.StandardRotation],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PreBiSGear, Presets.P1Gear],
	},

	autoRotation: (_player: Player<Spec.SpecBalanceDruid>): APLRotation => {
		return Presets.StandardRotation.rotation.rotation!;
	},

	reforge: {},
});
