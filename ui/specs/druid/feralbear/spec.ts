import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLAction, APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Cooldowns, EquipmentSpec, ItemSlot, PseudoStat, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { FeralBearDruid_Rotation as DruidRotation } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as FeralBearInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecFeralBearDruid>({
	spec: Spec.SpecFeralBearDruid,
	enableHealing: true,

	className: 'feral-bear-druid-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Druid),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	epRatios: [0, 0, 0.6, 0, 1.0, 0],
	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatAttackPower,
		Stat.StatFeralAttackPower,
		Stat.StatArmor,
		Stat.StatBonusArmor,
		Stat.StatDodgeRating,
		Stat.StatDefenseRating,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatExpertiseRating,
		Stat.StatPhysicalDamage,
		Stat.StatArmorPenetration,
	],
	epPseudoStats: [],
	epReferenceStat: Stat.StatAgility,
	tankRefStat: Stat.StatStamina,
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatAgility,
			Stat.StatStrength,
			Stat.StatAttackPower,
			Stat.StatArmor,
			Stat.StatBonusArmor,
			Stat.StatDodgeRating,
			Stat.StatDefenseRating,
			Stat.StatNatureResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatArcaneResistance,
			Stat.StatShadowResistance,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatDodgePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),

	defaults: {
		gear: EquipmentSpec.create(),
		epWeights: new Stats(),
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8);
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
			const critImmunityCap = new Stats().withPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent, 5.6);
			return hitCap.add(expCap).add(critImmunityCap);
		})(),
		other: Presets.OtherDefaults,
		consumables: Presets.DefaultConsumables,
		rotationType: APLRotationType.TypeAPL,
		aplRotation: Presets.ROTATION_DEFAULT.rotation.rotation!,
		talents: SavedTalents.create(),
		specOptions: Presets.DefaultOptions,
		raidBuffs: RaidBuffs.create({
			arcaneBrilliance: true,
			giftOfTheWild: true,
			prayerOfFortitude: true,
			prayerOfShadowProtection: true,
			prayerOfSpirit: true,
		}),
		partyBuffs: PartyBuffs.create({
			battleShout: TristateEffect.TristateEffectRegular,
			windfuryTotem: true,
			manaSpringTotem: TristateEffect.TristateEffectRegular,
			strengthOfEarthTotem: true,
		}),
		individualBuffs: IndividualBuffs.create({
			greaterBlessingOfKings: true,
			greaterBlessingOfMight: true,
			thorns: true,
		}),
		debuffs: Debuffs.create({
			exposeArmor: true,
			faerieFire: true,
			giftOfArthas: false,
			curseOfRecklessness: true,
			insectSwarm: true,
			judgementOfWisdom: true,
			sunderArmor: true,
		}),
	},

	playerIconInputs: [],
	rotationInputs: FeralBearInputs.FeralBearRotationConfig,
	includeBuffDebuffInputs: [Stat.StatStamina, Stat.StatArmor],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	otherInputs: {
		inputs: [
			FeralBearInputs.StartingRage,
			OtherInputs.InputDelay,
			OtherInputs.TankAssignment,
			OtherInputs.InspirationUptime,
			OtherInputs.IncomingHps,
			OtherInputs.HealingCadence,
			OtherInputs.HealingCadenceVariation,
			OtherInputs.AbsorbFrac,
			OtherInputs.BurstWindow,
			OtherInputs.HpPercentForDefensives,
			OtherInputs.InFrontOfTarget,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotRanged],

	encounterPicker: {
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		talents: [],
		// ROTATION_SIMPLE is kept in presets.ts for reference but omitted here —
		// the APL rotation is more user-friendly and handles CDs, re-shifting, and
		// on-use items more easily.
		rotations: [Presets.ROTATION_DEFAULT],
		gear: [],
	},

	autoRotation: (_player: Player<Spec.SpecFeralBearDruid>): APLRotation => {
		return Presets.ROTATION_DEFAULT.rotation.rotation!;
	},

	simpleRotation: (_player: Player<Spec.SpecFeralBearDruid>, simple: DruidRotation, _cooldowns: Cooldowns): APLRotation => {
		const doRotation = APLAction.fromJsonString(
			`{"bearOptimalRotationAction":{"maintainFaerieFire":${simple.maintainFaerieFire},"maintainDemoralizingRoar":${simple.maintainDemoralizingRoar},"maulRageThreshold":${simple.maulRageThreshold},"swipeUsage":${simple.swipeUsage},"swipeApThreshold":${simple.swipeApThreshold}}}`,
		);
		return APLRotation.create({
			priorityList: [APLListItem.create({ action: doRotation })],
		});
	},

	reforge: {},
});
