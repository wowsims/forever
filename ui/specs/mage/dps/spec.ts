import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { Cooldowns, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { SpecRotation } from '@sim/proto/spec_types';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as MageInputs from './inputs';
import * as Presets from './presets';
import { DefaultConsumables, DefaultDebuffs, DefaultIndividualBuffs, DefaultPartyBuffs, DefaultRaidBuffs } from './presets';

export default defineSpec<Spec.SpecMage>({
	spec: Spec.SpecMage,

	requiredTalentRows: [],
	className: 'mage-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Mage),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellDamage,
		Stat.StatArcaneDamage,
		Stat.StatFrostDamage,
		Stat.StatFireDamage,
		Stat.StatSpellPiercing,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatMana,
		Stat.StatMP5,
	],
	epPseudoStats: [PseudoStat.PseudoStatSchoolHitPercentArcane, PseudoStat.PseudoStatSchoolHitPercentFire, PseudoStat.PseudoStatSchoolHitPercentFrost],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellDamage,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatSpellDamage,
			Stat.StatFrostDamage,
			Stat.StatFireDamage,
			Stat.StatArcaneDamage,
			Stat.StatSpellPiercing,
		],
		[
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSchoolHitPercentArcane,
			PseudoStat.PseudoStatSchoolHitPercentFire,
			PseudoStat.PseudoStatSchoolHitPercentFrost,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHastePercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Wands],

	consumableStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatMP5,
		Stat.StatMana,
		Stat.StatSpellDamage,
		Stat.StatFrostDamage,
		Stat.StatFireDamage,
		Stat.StatArcaneDamage,
		Stat.StatCritRating,
		Stat.StatHitRating,
		Stat.StatHasteRating,
	],

	defaults: {
		// Default equipped gear.
		gear: Presets.P1ArcaneGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane, 17);
		})(),
		// Default consumes settings.
		consumables: DefaultConsumables,
		// Default talents.
		talents: Presets.ArcaneTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: DefaultRaidBuffs,

		partyBuffs: DefaultPartyBuffs,
		individualBuffs: DefaultIndividualBuffs,

		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.ArcaneMageSimpleRotation,
		debuffs: DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [MageInputs.MageArmorInputs()],
	rotationInputs: MageInputs.ArcaneMageRotationConfig,
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.DistanceFromTarget, OtherInputs.TankAssignment],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.DEFAULT_APL, Presets.APL_ARCANE_SIMPLE],
		// Preset talents that the user can quickly select.
		talents: [Presets.ArcaneTalents],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.P1ArcaneGear],
	},

	autoRotation: (_player: Player<Spec.SpecMage>): APLRotation => {
		// const numTargets = player.sim.encounter.targets.length;
		// if (numTargets >= 2) {
		// 	return Presets.ROTATION_PRESET_CLEAVE.rotation.rotation!;
		// } else {
		return Presets.DEFAULT_APL.rotation.rotation!;
		// }
	},

	// TODO: The simple settings are not applied yet. The default APL (apls/default.apl.json) has no
	// value variables for them to set, so the simple rotation is that APL as it stands.
	simpleRotation: (_player: Player<Spec.SpecMage>, _simple: SpecRotation<Spec.SpecMage>, _cooldowns: Cooldowns): APLRotation => {
		// const actions = AplUtils.simpleCooldownActions(cooldowns);
		// const rotation = APLRotation.clone(Presets.ROTATION_PRESET_ARCANE.rotation.rotation!);
		//
		// const { conserveStart = 20, conserveEnd = 30, delayMajorCDs = 10 } = simple;
		//
		// const conserveStartString = APLValueVariable.fromJson({
		// 	name: 'Conserve Start',
		// 	value: { const: { val: String(conserveStart) + '%' } },
		// });
		//
		// const conserveEndString = APLValueVariable.fromJson({
		// 	name: 'Conserve End',
		// 	value: { const: { val: String(conserveEnd) + '%' } },
		// });
		//
		// const delayMajorCDsString = APLValueVariable.fromJson({
		// 	name: 'Delay Major CDs',
		// 	value: { const: { val: String(delayMajorCDs) + 's' } },
		// });
		//
		// rotation.valueVariables[0] = conserveStartString;
		// rotation.valueVariables[1] = conserveEndString;
		// rotation.valueVariables[2] = delayMajorCDsString;
		//
		// // The user's timings replace the preset's own logic for those spells (e.g. Icy Veins lust/drums alignment).
		// AplUtils.hideScheduledCooldownCasts(rotation, cooldowns);
		//
		// return APLRotation.create({
		// 	prepullActions: rotation.prepullActions,
		// 	priorityList: [
		// 		...actions.map(action =>
		// 			APLListItem.create({
		// 				action: action,
		// 			}),
		// 		),
		// 		...rotation.priorityList,
		// 	],
		// 	groups: rotation.groups,
		// 	valueVariables: rotation.valueVariables,
		// });
		return APLRotation.clone(Presets.DEFAULT_APL.rotation.rotation!);
	},

	reforge: {},
});
