import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation } from '@generated/proto/apl';
import { ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import type { SpecRotation } from '@sim/proto/spec_types';
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
		Stat.StatSpirit,
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
		// Default equipped gear, that of the default simple rotation's build.
		gear: Presets.DemoFireBrandGear.gear,

		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSpellHitPercent, 17);
		})(),

		// Default consumes settings.
		consumables: Presets.DefaultConsumables,

		simpleRotation: Presets.DefaultSimpleRotation,

		// Default talents.
		talents: Presets.DemoFireBrandTalents.data,
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
	rotationInputs: WarlockInputs.RotationInputs,
	// `simpleRotation` below never reads its Cooldowns argument.
	hideSimpleCooldowns: true,
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [WarlockInputs.PetInput(), WarlockInputs.ArmorInput(), WarlockInputs.DemonicSacrificeInput()],

	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.ChannelClipDelay, OtherInputs.DistanceFromTarget],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [Presets.DemoFireBrandTalents, Presets.DemoShadowBrandTalents, Presets.DestroTalents, Presets.SmRuinTalents, Presets.AffWrackTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.DEMO_FIRE_BRAND_APL, Presets.DEMO_SHADOW_BRAND_APL, Presets.DESTRO_APL, Presets.SM_RUIN_APL, Presets.AFF_WRACK_APL],

		// Preset gear configurations that the user can quickly select.
		gear: [Presets.DemoFireBrandGear, Presets.DemoShadowBrandGear, Presets.DestroGear, Presets.SmRuinGear, Presets.AffWrackGear],
		// Talents, demons and rotation together, one per build.
		builds: Presets.Builds,
	},

	autoRotation: (_player: Player<Spec.SpecWarlock>): APLRotation => {
		return Presets.DEMO_FIRE_BRAND_APL.rotation.rotation!;
	},

	simpleRotation: (_player: Player<Spec.SpecWarlock>, simple: SpecRotation<Spec.SpecWarlock>): APLRotation => {
		return Presets.SimpleRotationApls[simple.rotation].rotation.rotation!;
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
