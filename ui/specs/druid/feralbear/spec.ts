import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ItemSlot, ItemSpec, ItemSwap, PseudoStat, Spec, Stat, TristateEffect } from '@generated/proto/common';
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
		Stat.StatArmorNoMods,
		Stat.StatDodgeRating,
		Stat.StatDefenseRating,
		Stat.StatHitRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatExpertiseRating,
		Stat.StatPhysicalDamage,
		Stat.StatArmorPenetration,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps],
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
			Stat.StatArmorNoMods,
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
		gear: Presets.Level60Gear.gear,
		itemSwap: ItemSwap.create({
			items: Array.from({ length: 17 }, (_, slot) => ItemSpec.create(slot === ItemSlot.ItemSlotMainHand ? { id: 9449, enchant: 1900 } : {})),
		}),
		encounter: Presets.Level60Encounter,
		epWeights: Presets.DefaultEp,
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8);
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
			const critImmunityCap = new Stats().withPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent, 5.6);
			return hitCap.add(expCap).add(critImmunityCap);
		})(),
		other: Presets.OtherDefaults,
		consumables: Presets.DefaultConsumables,
		rotationType: APLRotationType.TypeAPL,
		aplRotation: Presets.ROTATION_ADAPTIVE.rotation.rotation!,
		talents: Presets.BearTalents.data,
		specOptions: Presets.DefaultOptions,
		raidBuffs: RaidBuffs.create({
			arcaneIntellect: true,
			prayerOfSpirit: true,
			markOfTheWild: true,
			powerWordFortitude: true,
			shadowProtection: true,
			strengthOfEarthTotem: 1,
			graceOfAirTotem: 1,
			windfuryTotem: true,
			battleShout: 1, // Rank 6
		}),
		partyBuffs: PartyBuffs.create({ devotionAura: true, manaSpringTotem: TristateEffect.TristateEffectRegular, retributionAura: true }),
		individualBuffs: IndividualBuffs.create({ greaterBlessingOfKings: true, greaterBlessingOfMight: 1, thorns: true }),
		debuffs: Debuffs.create({
			judgementOfLight: true,
			judgementOfWisdom: true,

			faerieFire: false,
			exposeArmor: true,
			sunderArmor: true,
			vindication: true,
		}),
	},

	playerIconInputs: [],
	includeBuffDebuffInputs: [Stat.StatStamina, Stat.StatArmor],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	otherInputs: {
		inputs: [
			FeralBearInputs.StartingRage,
			FeralBearInputs.AssumeBleedActive,
			OtherInputs.InputDelay,
			OtherInputs.DistanceFromTarget,
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
		epWeights: [Presets.DefaultEpWeights, Presets.HybridEpWeights],
		talents: [Presets.BearTalents, Presets.HybridTalents],
		rotations: [Presets.ROTATION_ADAPTIVE],
		gear: [Presets.Level60Gear],
		encounters: [Presets.Level60Encounter],
	},

	// Auto uses the adaptive rotation.
	autoRotation: (_player: Player<Spec.SpecFeralBearDruid>): APLRotation => {
		return Presets.ROTATION_ADAPTIVE.rotation.rotation!;
	},

	reforge: {},
});
