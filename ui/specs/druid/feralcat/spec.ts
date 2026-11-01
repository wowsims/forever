import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ItemSlot, ItemSpec, ItemSwap, PseudoStat, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as FeralInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecFeralCatDruid>({
	spec: Spec.SpecFeralCatDruid,

	className: 'feral-druid-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Druid),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	// Extra stats shown in consumables picker (beyond epStats).
	consumableStats: [Stat.StatHasteRating, Stat.StatMana, Stat.StatSpirit, Stat.StatMP5, Stat.StatIntellect, Stat.StatStamina, Stat.StatArmor],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatFeralAttackPower,
		Stat.StatPhysicalDamage,
		Stat.StatHitRating,
		Stat.StatExpertiseRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatArmorPenetration,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatAttackPower,
			Stat.StatMana,
			Stat.StatArmorPenetration,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.PreRaidGear.gear,
		itemSwap: ItemSwap.create({
			items: Array.from({ length: 17 }, (_, slot) => ItemSpec.create(slot === ItemSlot.ItemSlotMainHand ? { id: 9449, enchant: 1896 } : {})),
		}),
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.DefaultEp,
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8).withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
		})(),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default rotation settings.
		rotationType: APLRotationType.TypeAPL,
		aplRotation: Presets.JVMES_APL.rotation.rotation!,
		encounter: Presets.DefaultEncounter,
		// Genesis, nature's majesty, nature's reach, ferocity, furor, feral swiftness, shredding attacks,
		// savage fury, feral charge, sharpened claws, predatory strikes, predatory discipline, blood frenzy, leader of
		// the pack, natural instinct, rend and tear, shifting power, heart of the wild, naturalist.
		// Shifting Power is absent until that talent is taken.
		talents: Presets.CatTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			arcaneIntellect: true,
			prayerOfSpirit: true,
			markOfTheWild: true,
			powerWordFortitude: true,
			strengthOfEarthTotem: 1,
			graceOfAirTotem: 1,
			windfuryTotem: true,
			battleShout: 1, // Rank 6
		}),
		partyBuffs: PartyBuffs.create({ manaSpringTotem: TristateEffect.TristateEffectRegular }),
		individualBuffs: IndividualBuffs.create({ greaterBlessingOfKings: true, greaterBlessingOfMight: 1 }),
		debuffs: Debuffs.create({ judgementOfWisdom: true, sunderArmor: true, giftOfArthas: true, vindication: true }),
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5, Stat.StatIntellect, Stat.StatStamina],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.InputDelay,
			OtherInputs.DistanceFromTarget,
			OtherInputs.TankAssignment,
			OtherInputs.InFrontOfTarget,
			FeralInputs.CannotShredTarget,
			FeralInputs.AssumeBleedActive,
			FeralInputs.PrepullClearcasting,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [Presets.DefaultEpWeights, Presets.HybridEpWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.CatTalents, Presets.HybridTalents],
		rotations: [Presets.JVMES_APL],
		builds: [],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.PreRaidGear, Presets.Level60Gear],
	},

	// Auto uses the adaptive rotation.
	autoRotation: (_player: Player<Spec.SpecFeralCatDruid>): APLRotation => {
		return Presets.JVMES_APL.rotation.rotation!;
	},

	hiddenMCDs: [],

	reforge: {},
});
