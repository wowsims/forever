import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation } from '@generated/proto/apl';
import { ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as PriestSharedInputs from '../shared/inputs';
import * as PriestInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecDpsPriest>({
	spec: Spec.SpecDpsPriest,

	className: 'priest-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Priest),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [
		'Some items may display and use stats a litle higher than their original value.',
		'Procs from Weapons, Trinkets and other Items are not yet supported',
	],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatShadowDamage,
		Stat.StatHolyDamage,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatMana,
		Stat.StatMP5,
		Stat.StatSpirit,
	],
	epPseudoStats: [PseudoStat.PseudoStatSchoolHitPercentShadow, PseudoStat.PseudoStatSchoolHitPercentHoly],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellDamage,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatMP5,
			Stat.StatSpirit,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatSpellDamage,
			Stat.StatShadowDamage,
			Stat.StatHolyDamage,
		],
		[
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHastePercent,
			PseudoStat.PseudoStatSchoolHitPercentShadow,
			PseudoStat.PseudoStatSchoolHitPercentHoly,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Wands],

	defaults: {
		// Default equipped gear.
		gear: Presets.PreraidGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow, 17).withPseudoStat(PseudoStat.PseudoStatSchoolHitPercentHoly, 17);
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
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [PriestSharedInputs.ArmorInput(), PriestInputs.ShadowformInput()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshDruid, BuffDebuffInputs.AtieshMage, BuffDebuffInputs.AtieshWarlock],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.ChannelClipDelay, OtherInputs.TankAssignment, OtherInputs.DistanceFromTarget],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.ShadowEpWeights, Presets.SmiteEpWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.StandardTalents, Presets.SmiteTalents],
		rotations: [Presets.ROTATION_PRESET_DEFAULT, Presets.ROTATION_SMITE_DEFAULT],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PreraidGear, Presets.PreraidSmite],
	},

	autoRotation: (priest: Player<Spec.SpecDpsPriest>): APLRotation => {
		const [disc, , shadow] = priest.getTalentTreePoints();

		if (disc > shadow) {
			return Presets.ROTATION_SMITE_DEFAULT.rotation.rotation!;
		}
		return Presets.ROTATION_PRESET_DEFAULT.rotation.rotation!;
	},

	reforge: {},
});
