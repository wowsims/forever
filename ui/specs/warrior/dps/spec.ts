import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLRotation, APLRotation_Type, SimpleRotation } from '@generated/proto/apl';
import { Cooldowns, HandType, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { DpsWarriorSpec, WarriorSunder } from '@generated/proto/warrior';
import * as Mechanics from '@sim/constants/mechanics';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { withSpellIds } from '@sim/proto/apl_utils';
import { SpecRotation } from '@sim/proto/spec_types';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as WarriorInputs from '../shared/inputs';
import * as WarriorPresets from '../shared/presets';
import * as DpsWarriorInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecDpsWarrior>({
	spec: Spec.SpecDpsWarrior,

	className: 'dps-warrior-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Warrior),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatHitRating,
		Stat.StatHasteRating,
		Stat.StatCritRating,
		Stat.StatExpertiseRating,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatStrength,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatStamina, Stat.StatStrength, Stat.StatAgility, Stat.StatAttackPower, Stat.StatArmorPenetration],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatOffHandHitPercent,
			PseudoStat.PseudoStatOffHandCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee],

	defaults: {
		// Default equipped gear.
		gear: Presets.PreBiSFuryGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.PreBiSFuryEpWeights.epWeights,
		statCaps: (() => {
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
			return expCap;
		})(),
		softCapBreakpoints: (() => {
			const meleeHitSoftCapConfig = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
				breakpoints: [8, 27],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [0.57 * Mechanics.HIT_RATING_PER_HIT_PERCENT, 0],
			});

			return [meleeHitSoftCapConfig];
		})(),
		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.SIMPLE_ROTATION,
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.FuryTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: WarriorPresets.DefaultRaidBuffs,
		partyBuffs: WarriorPresets.DefaultPartyBuffs,
		individualBuffs: WarriorPresets.DefaultIndividualBuffs,
		debuffs: WarriorPresets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	// Icon pickers are not part of the `InputConfig` union, so the two Battle Shout icon toggles
	// sit in the player icon row.
	playerIconInputs: [WarriorInputs.ShoutPicker(), WarriorInputs.StancePicker(), WarriorInputs.BattleShoutT2(), WarriorInputs.BattleShoutRank7()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	rotationInputs: DpsWarriorInputs.RotationInputs,
	// `simpleRotation` below never reads its Cooldowns argument.
	hideSimpleCooldowns: true,
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			WarriorInputs.StartingRage(),
			WarriorInputs.StanceSnapshot(),
			OtherInputs.DistanceFromTarget,
			WarriorInputs.QueueDelay(),
			OtherInputs.InputDelay,
			OtherInputs.TankAssignment,
			OtherInputs.InFrontOfTarget,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [Presets.PreBiSFuryEpWeights, Presets.PreBiSArmsEpWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.FuryTalents, Presets.ArmsTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.SIMPLE_DEFAULT_ROTATION, Presets.FURY_DEFAULT_ROTATION, Presets.ARMS_DEFAULT_ROTATION],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PreBiSFuryGear, Presets.PreBiSArmsGear],
		settings: [Presets.DefaultSettings],
		builds: [Presets.PreBiSFuryBuild, Presets.PreBiSArmsBuild],
	},

	autoRotation: (player: Player<Spec.SpecDpsWarrior>): APLRotation => {
		if (Presets.isArmsSpec(player) || Presets.isArmsKebabSpec(player)) {
			return Presets.ARMS_DEFAULT_ROTATION.rotation.rotation!;
		}

		return Presets.FURY_DEFAULT_ROTATION.rotation.rotation!;
	},

	simpleRotation: (player: Player<Spec.SpecDpsWarrior>, simple: SpecRotation<Spec.SpecDpsWarrior>, _: Cooldowns): APLRotation => {
		let {
			spec,
			sunderArmor = WarriorSunder.WarriorSunderHelp,
			useOverpower = true,
			useRend = true,
			useRecklessness = false,
			useStanceDance = false,
			useHamstring = true,
			useHeroicStrikeRank9 = false,
			useBattleShoutRank7 = false,
		} = simple;

		if (!spec) {
			if (Presets.isArmsSpec(player) || Presets.isArmsKebabSpec(player)) {
				spec = DpsWarriorSpec.DpsWarriorSpecArms;
			} else {
				spec = DpsWarriorSpec.DpsWarriorSpecFury;
			}
		}

		// The presets cast the ranks the Ruins of Ahn'Qiraj class books teach; each rank left unchosen moves to the
		// trainer's top rank.
		const trainerRanks = new Map<number, number>();
		if (!useHeroicStrikeRank9) trainerRanks.set(25286, 11567);
		if (!useBattleShoutRank7) trainerRanks.set(25289, 11551);
		const rotation = withSpellIds(
			spec == DpsWarriorSpec.DpsWarriorSpecFury ? Presets.FURY_DEFAULT_ROTATION.rotation.rotation! : Presets.ARMS_DEFAULT_ROTATION.rotation.rotation!,
			trainerRanks,
		);

		// With this group shown (for Arms, with Stance Dance as well), the racials group casts Elune's Light only while
		// Recklessness is down, more than 15 s remain, and either more than 195 s or at most 30 s remain.
		const recklessnessAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action.action.groupReference.groupName === 'Recklessness ON/OFF',
		);
		if (recklessnessAction) recklessnessAction.hide = !useRecklessness;

		const stanceDanceAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action.action.groupReference.groupName === 'Stance Dance ON/OFF',
		);
		if (stanceDanceAction) stanceDanceAction.hide = !useStanceDance;

		// Help stacks to 5 and leaves the rest to the raid, restacking once the debuff falls off even with no raid Sunder;
		// Maintain also refreshes; None casts none. Stacking and refreshing stay separate groups so the option can split them.
		const sunderArmorAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Sunder Armor',
		);
		if (sunderArmorAction) sunderArmorAction.hide = sunderArmor == WarriorSunder.WarriorSunderNone;

		const sunderRefreshAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Sunder Armor: Refresh',
		);
		if (sunderRefreshAction) sunderRefreshAction.hide = sunderArmor != WarriorSunder.WarriorSunderMaintain;

		const opWeaveAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Overpower Weaving',
		);
		if (opWeaveAction) opWeaveAction.hide = !useOverpower;

		const rendWeaveAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Rend Weave ON/OFF',
		);
		if (rendWeaveAction) rendWeaveAction.hide = !useRend;

		// Hamstring is woven for Windfury procs only while Windfury Totem is known: Fury below Overpower weaving, Arms above
		// Execute while the Windfury buff is down.
		const hamstringWeaveAction = rotation.priorityList.find(
			action => action.action?.action.oneofKind === 'groupReference' && action.action?.action.groupReference.groupName === 'Hamstring Weave ON/OFF',
		);
		if (hamstringWeaveAction) hamstringWeaveAction.hide = !useHamstring;

		return APLRotation.create({
			simple: SimpleRotation.create({}),
			...rotation,
		});
	},

	reforge: {
		updateSoftCaps: (softCaps, player, ctx) => {
			const gear = player.getGear();
			const mainHandType = gear.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType;
			const offHandType = gear.getEquippedItem(ItemSlot.ItemSlotOffHand)?.item.handType;
			const isFury =
				mainHandType &&
				[HandType.HandTypeOneHand, HandType.HandTypeMainHand].includes(mainHandType) &&
				offHandType &&
				[HandType.HandTypeOneHand, HandType.HandTypeOffHand].includes(offHandType);

			const softCapToModify = softCaps.find(sc => sc.unitStat.equalsPseudoStat(PseudoStat.PseudoStatMeleeHitPercent));
			if (softCapToModify) {
				if (isFury) {
					softCapToModify.breakpoints = ctx.defaults.softCapBreakpoints?.[0].breakpoints || [];
					softCapToModify.postCapEPs = ctx.defaults.softCapBreakpoints?.[0].postCapEPs || [];
				} else {
					softCapToModify.breakpoints = [8];
					softCapToModify.postCapEPs = [0];
				}
			}

			return softCaps;
		},
	},
});
