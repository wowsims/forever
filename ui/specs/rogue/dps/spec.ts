import * as InputHelpers from '@features/inputs/input_helpers';
import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type, APLValueVariable } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Cooldowns, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import * as Mechanics from '@sim/constants/mechanics';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import * as AplUtils from '@sim/proto/apl_utils';
import type { SpecRotation } from '@sim/proto/spec_types';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as Presets from './presets';

export default defineSpec<Spec.SpecRogue>({
	spec: Spec.SpecRogue,

	className: 'rogue-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Rogue),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [
		'The APL is in constant flux due to bug fixes and new findings; if your DPS drops dramatically, reset it back to "Auto" in the Rotation tab!',
		'Mutilate does not have a default APL currently. It will not be automatically used when talented.',
	],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatHitRating,
		Stat.StatExpertiseRating,
		Stat.StatArmorPenetration,
		Stat.StatSpellPiercing,
		Stat.StatAttackPower,
		Stat.StatPhysicalDamage,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStamina, Stat.StatAgility, Stat.StatStrength, Stat.StatAttackPower, Stat.StatArmorPenetration, Stat.StatSpellPiercing],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatOffHandHitPercent,
			PseudoStat.PseudoStatOffHandCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSpellCritPercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee],

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_SLOW_MH_GEAR.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.EP_PRESET_COMBAT_SINISTER.epWeights,
		statCaps: (() => {
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
			return expCap;
		})(),
		softCapBreakpoints: (() => {
			const meleeHitSoftCapConfig = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
				breakpoints: [8, 27],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [3.06 * Mechanics.HIT_RATING_PER_HIT_PERCENT, 0],
			});

			return [meleeHitSoftCapConfig];
		})(),
		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.DEFAULT_SIMPLE_ROTATION,
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.TALENTS_COMBAT_SS.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			giftOfTheWild: true,
			strengthOfEarthTotem: true,
			windfuryTotem: true,
			battleShout: 1, // Rank 6
			leaderOfThePack: true,
			graceOfAirTotem: true,
			moonkinAura: true,
			prayerOfFortitude: true,
		}),
		partyBuffs: PartyBuffs.create({}),
		individualBuffs: IndividualBuffs.create({
			greaterBlessingOfKings: true,
			greaterBlessingOfMight: true,
		}),
		debuffs: Debuffs.create({
			curseOfRecklessness: true,
			faerieFire: true,
			giftOfArthas: true,
			sunderArmor: true,
			curseOfElements: true,
			armorShatter: true,
		}),
	},

	playerInputs: {
		inputs: [],
	},
	rotationInputs: {
		inputs: [
			InputHelpers.makeRotationBooleanInput<Spec.SpecRogue>({
				fieldName: 'useBackstab',
				label: i18n.t('rotation_tab.options.rogue.use_backstab.label'),
			}),
			InputHelpers.makeRotationBooleanInput<Spec.SpecRogue>({
				fieldName: 'useExposeArmor',
				label: i18n.t('rotation_tab.options.rogue.use_expose_armor.label'),
			}),
		],
	},
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatHitRating, Stat.StatSpellDamage, Stat.StatSpellPiercing],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InFrontOfTarget, OtherInputs.InputDelay],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [Presets.EP_PRESET_ASSASSINATION, Presets.EP_PRESET_COMBAT_DAGS, Presets.EP_PRESET_COMBAT_SINISTER, Presets.EP_PRESET_SUBTLETY],
		// Preset talents that the user can quickly select.
		talents: [
			Presets.TALENTS_MUTILATE_ASN,
			Presets.TALENTS_MUTILATE_COM,
			Presets.TALENTS_COMBAT_SS,
			Presets.TALENTS_COMBAT_BS,
			Presets.TALENTS_SUB_HEMO,
			Presets.TALENTS_SUB_BS,
		],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.SIMPLE_ROTATION, Presets.DEFAULT_APL],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.P1_BS_COMBAT_GEAR, Presets.P1_MUTI_GEAR, Presets.P1_SLOW_MH_GEAR, Presets.P1_BS_SUB_GEAR],
		builds: [
			Presets.PRESET_BUILD_MUTILATE_ASN,
			Presets.PRESET_BUILD_MUTILATE_COM,
			Presets.PRESET_BUILD_COMBAT_SS,
			Presets.PRESET_BUILD_COMBAT_BS,
			Presets.PRESET_BUILD_SUB_HEMO,
			Presets.PRESET_BUILD_SUB_BS,
		],
	},

	autoRotation: (_player: Player<Spec.SpecRogue>): APLRotation => {
		return Presets.DEFAULT_APL.rotation.rotation!;
	},

	simpleRotation: (_player: Player<Spec.SpecRogue>, simple: SpecRotation<Spec.SpecRogue>, cooldowns: Cooldowns): APLRotation => {
		const actions = AplUtils.simpleCooldownActions(cooldowns);
		const rotation = APLRotation.clone(Presets.DEFAULT_APL.rotation.rotation!);
		const values: Record<string, string> = {
			'Use Backstab': String(simple.useBackstab),
			'Use Expose Armor': String(simple.useExposeArmor),
		};
		rotation.valueVariables = rotation.valueVariables.map(variable =>
			variable.name in values ? APLValueVariable.fromJson({ name: variable.name, value: { const: { val: values[variable.name] } } }) : variable,
		);
		rotation.priorityList = [...actions.map(action => APLListItem.create({ action })), ...rotation.priorityList];

		return rotation;
	},

	// reforge: {},
});
