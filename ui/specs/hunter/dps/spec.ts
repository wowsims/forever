import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as other_inputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLListItem, APLRotation, APLValueVariable } from '@generated/proto/apl';
import { Cooldowns, HandType, ItemSlot, MobType, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { Hunter_Rotation_Playstyle as Playstyle } from '@generated/proto/hunter';
import i18n from '@i18n/config';
import { translateMobType } from '@i18n/localization';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import * as AplUtils from '@sim/proto/apl_utils';
import { SpecRotation } from '@sim/proto/spec_types';
import { StatCap, UnitStat } from '@sim/proto/stats';
import { targetMobType } from '@sim/proto/target';
import { defineSpec } from '@sim/spec_config';
import { subscribeAll, subscribeEncounterField, subscribePlayerField } from '@sim/state/subscriptions';

import * as HunterInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecHunter>({
	spec: Spec.SpecHunter,

	className: 'hunter-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Hunter),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [
		// No Track spell names a mechanical target or one of no type, so Improved Tracking adds nothing against them.
		simUI => ({
			updateOn: subscribeAll([subscribePlayerField(simUI.player, 'talentsString'), subscribeEncounterField(simUI.sim.encounter, 'targets')]),
			getContent: () => {
				const mobType = targetMobType(simUI.sim.encounter.primaryTarget);
				const untracked = mobType === MobType.MobTypeUnknown || mobType === MobType.MobTypeMechanical;
				if (!simUI.player.getTalents().improvedTracking || !untracked) return '';
				return i18n.t('sidebar.warnings.hunter_improved_tracking', { mobType: translateMobType(mobType) });
			},
			action: {
				label: i18n.t('sidebar.warnings.set_target_mob_type', { mobType: translateMobType(MobType.MobTypeDragonkin) }),
				run: () => simUI.sim.encounter.modifyTarget(0, target => (target.mobType = MobType.MobTypeDragonkin)),
			},
		}),
	],
	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatIntellect,
		Stat.StatMP5,
		Stat.StatAttackPower,
		Stat.StatRangedAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatHitRating,
		Stat.StatHasteRating,
		Stat.StatCritRating,
		Stat.StatExpertiseRating,
		Stat.StatPhysicalDamage,
	],
	epPseudoStats: [
		PseudoStat.PseudoStatRangedHitPercent,
		PseudoStat.PseudoStatRangedCritPercent,
		PseudoStat.PseudoStatMainHandDps,
		PseudoStat.PseudoStatOffHandDps,
		PseudoStat.PseudoStatRangedDps,
	],
	consumableStats: [Stat.StatStamina, Stat.StatHealth, Stat.StatMana],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatIntellect,
			Stat.StatMP5,
			Stat.StatAttackPower,
			Stat.StatRangedAttackPower,
			Stat.StatArmorPenetration,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatRangedHitPercent,
			PseudoStat.PseudoStatRangedCritPercent,
			PseudoStat.PseudoStatRangedHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee, WeaponSkillGroup.Ranged],
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotRanged, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	defaults: {
		...Presets.Defaults,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.WeaveEpWeights.epWeights,
		softCapBreakpoints: [
			StatCap.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent, {
				breakpoints: [8],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [0],
			}),
		],
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [HunterInputs.QuiverInput(), HunterInputs.AmmoInput()],
	sections: [HunterInputs.PetSection],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatSpirit, Stat.StatCritRating, Stat.StatSpellDamage],
	// Rows that never move hunter DPS: stamina, threat, caster and party-only buffs and the debuffs only a tank wants.
	excludeBuffDebuffInputs: [
		BuffDebuffInputs.AdmiralsHat,
		BuffDebuffInputs.AtieshMage,
		BuffDebuffInputs.AtieshWarlock,
		BuffDebuffInputs.BloodPact,
		BuffDebuffInputs.ConcentrationAura,
		BuffDebuffInputs.DemoralizingRoar,
		BuffDebuffInputs.DemoralizingScreech,
		BuffDebuffInputs.HexOfWeakness,
		BuffDebuffInputs.JudgementOfTheCrusader,
		BuffDebuffInputs.MoonkinAura,
		BuffDebuffInputs.PowerInfusion,
		BuffDebuffInputs.PowerWordFortitude,
		BuffDebuffInputs.ScorpidSting,
		BuffDebuffInputs.ThunderClap,
		BuffDebuffInputs.Thunderfury,
	],
	rotationInputs: HunterInputs.RotationInputs,
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [other_inputs.InputDelay, other_inputs.ChannelClipDelay, other_inputs.DistanceFromTarget, other_inputs.TankAssignment],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.WeaveEpWeights, Presets.TurretEpWeights, Presets.MeleeEpWeights],
		// Preset talents that the user can quickly select.
		talents: Presets.BuildTalents,
		// Preset rotations that the user can quickly select.
		rotations: [
			Presets.SimpleRotation,
			Presets.WeaveAPLRotation,
			Presets.NormalWeaveAPLRotation,
			Presets.CleanWeaveAPLRotation,
			Presets.TurretAPLRotation,
			Presets.TurretMultiAPLRotation,
			Presets.TurretHawkAPLRotation,
			Presets.MeleeAPLRotation,
		],
		// Preset gear configurations that the user can quickly select.
		gear: [
			Presets.PreBisWeaveGear,
			Presets.PreBisWeaveCheapGear,
			Presets.PreBisTurretGear,
			Presets.PreBisTurretCheapGear,
			Presets.PreBisMeleeGear,
			Presets.PreBisMeleeCheapGear,
		],
		builds: Presets.Builds,
	},

	// With a two-hander in the main hand the hunter weaves it in; with anything else it stays at range.
	autoRotation: (player: Player<Spec.SpecHunter>): APLRotation => {
		const twoHander = player.getGear().getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeTwoHand;
		return APLRotation.fromJson(Presets.PlaystyleAPLs[twoHander ? Playstyle.Weave : Playstyle.Turret]);
	},

	// A rank toggle turned on casts the book rank in place of the trainer's top rank.
	simpleRotation: (_player: Player<Spec.SpecHunter>, simple: SpecRotation<Spec.SpecHunter>, cooldowns: Cooldowns): APLRotation => {
		const actions = AplUtils.simpleCooldownActions(cooldowns);
		const hawksAtRange = simple.playstyle === Playstyle.Turret && simple.useHawk;
		const bookRanks = new Map<number, number>();
		if (simple.useSerpentStingRank9) bookRanks.set(13555, 25295);
		if (simple.useAspectOfTheHawkRank7) bookRanks.set(14322, 25296);
		const rotation = AplUtils.withSpellIds(
			APLRotation.fromJson(hawksAtRange ? Presets.TurretHawkPlaystyleAPL : Presets.PlaystyleAPLs[simple.playstyle]),
			bookRanks,
		);

		const values: Record<string, string> = {
			'Pull range': String(simple.pullRange || 25),
			'Use Aimed Shot': String(simple.useAimed),
			'Use Multi-Shot': String(simple.useMulti),
			'Use Arcane Shot': String(simple.useArcane),
			'Use Sniper Shot': String(simple.useSniper),
			'Use Summon Hawk': String(simple.useHawk),
		};
		rotation.valueVariables = rotation.valueVariables.map(variable =>
			variable.name in values ? APLValueVariable.fromJson({ name: variable.name, value: { const: { val: values[variable.name] } } }) : variable,
		);

		return APLRotation.create({
			prepullActions: rotation.prepullActions,
			priorityList: [...actions.map(action => APLListItem.create({ action })), ...rotation.priorityList],
			groups: rotation.groups,
			valueVariables: rotation.valueVariables,
		});
	},

	reforge: {},
});
