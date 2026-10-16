import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation } from '@generated/proto/apl';
import { EquipmentSpec, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';
import { subscribeAll, subscribePlayerField } from '@sim/state/subscriptions';

import * as WarlockInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecWarlock>({
	spec: Spec.SpecWarlock,

	className: 'warlock-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Warlock),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatShadowDamage,
		Stat.StatFireDamage,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatSpellPiercing,
		Stat.StatMP5,
	],
	// Reference stat against which to calculate EP. DPS classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellDamage,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpellDamage,
			Stat.StatShadowDamage,
			Stat.StatFireDamage,
			Stat.StatSpellPiercing,
			Stat.StatMP5,
		],
		[PseudoStat.PseudoStatSpellHitPercent, PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Wands],

	defaults: {
		// Default equipped gear.
		gear: EquipmentSpec.create(),

		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSpellHitPercent, 17);
		})(),

		// Default consumes settings.
		consumables: Presets.DefaultConsumables,

		// Default talents.
		talents: SavedTalents.create(),
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,

		// Default buffs and debuffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,

		other: Presets.OtherDefaults,
	},

	consumableStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatMP5,
		Stat.StatSpellDamage,
		Stat.StatCritRating,
		Stat.StatHitRating,
		Stat.StatHasteRating,
		Stat.StatShadowDamage,
		Stat.StatFireDamage,
	],
	playerInputs: {
		inputs: [WarlockInputs.SoulShardsInput()],
	},
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [WarlockInputs.PetInput(), WarlockInputs.ArmorInput(), WarlockInputs.DemonicSacrificeInput()],

	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatAttackPower],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.DistanceFromTarget],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.DEFAULT_APL],

		// Preset gear configurations that the user can quickly select.
		gear: [],
	},

	autoRotation: (_player: Player<Spec.SpecWarlock>): APLRotation => {
		return Presets.DEFAULT_APL.rotation.rotation!;
	},
	sections: [WarlockInputs.CursesSection],

	// Re-settles the demon out and the demon sacrificed when the talents or the options change, such as
	// dropping Demonic Pact or Demonic Sacrifice.
	derivedSettings: [
		{
			subscribe: player => subscribeAll([subscribePlayerField(player, 'talentsString'), subscribePlayerField(player, 'specOptions')]),
			apply: player => WarlockInputs.settleDemons(player),
		},
	],

	reforge: {},
});
