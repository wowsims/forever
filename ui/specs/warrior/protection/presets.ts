import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorStance } from '@generated/proto/warrior';
import { Stats } from '@sim/proto/stats';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import * as WarriorPresets from '../shared/presets';
import GenericApl from './apls/default.apl.json';
import PreBiSBalancedGearJson from './gear_sets/prebis_balanced.gear.json';
import PreBiSOffensiveGearJson from './gear_sets/prebis_offensive.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

// The rotation hides Autocast Other Cooldowns, so an on-use trinket fires only when the "CD: Offensive Trinkets" group
// names its item id.
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

// Dungeon and world drops, quest rewards, the warrior's dungeon tier set, crafted items whose reagents include no raid
// materials and vendor items; no reputation, PvP or Silithus rewards, except one exalted PvP faction, Alterac Valley.
// The most threat that survives the default target on the default healing. Crest of Retribution drops in Stratholme,
// Blackstone Ring in Maraudon, Ironfoe in Blackrock Depths, Battleborn Armbraces and Handcrafted Mastersmith Leggings
// in Blackrock Spire, Satyr's Bow and Counterattack Lodestone in Dire Maul; Mark of Fordring is the reward for In
// Dreams. Don Julio's Band is sold by the Alterac Valley quartermasters, Howler's Furs and Adaptive Combat Assistant by
// the Undermine Trader in Stranglethorn Vale, and the Breastplate and Gauntlets of Heroism by the Battlegear of Heroism
// vendors in Orgrimmar and Ironforge. Spaulders of Containment, Girdle of Extracted Answers and Furbolg Berserker's
// Boots have no source in the item database. Counterattack Lodestone's on-use is not modelled; it counts for its attack
// power alone, 30 and another 30 against a Mechanical target such as the default one.
export const PreBiSOffensiveGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSOffensiveGearJson, { group: 'Offensive', devOnly: true });

// The same sources, built to take less damage at a small cost in threat; its defense stays far below the 140 that
// makes a level 63 boss unable to crit. Ironfoe drops in Blackrock Depths and Handcrafted Mastersmith Leggings in
// Blackrock Spire; Royal Seal of Eldre'Thalas is the reward for Codex of Defense in Dire Maul. The Immovable Object and
// Don Julio's Band are sold by the Alterac Valley quartermasters, Stalwart Cloak, Valiant's Waistguard, Cutthroat's
// Signet and Adaptive Combat Assistant by the Undermine Trader, and the Heroism pieces by the Battlegear of Heroism
// vendors in Orgrimmar and Ironforge. Horn of Ve'han and Furbolg Berserker's Boots have no source in the item database.
export const PreBiSBalancedGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSBalancedGearJson, { group: 'Balanced', devOnly: true });

export const OffensiveTalents = PresetUtils.makePresetTalents('Offensive', SavedTalents.create({ talentsString: '3-0505000501-050032120331021351' }));

export const BalancedTalents = PresetUtils.makePresetTalents('Balanced', SavedTalents.create({ talentsString: '05-05050005-150531120300001351' }));

// Measured on each Pre-BiS gear set with its talents and the preset rotation, against the default target on the
// default healing: 0.6 x the threat weights (Strength = 1) plus 1.15 x the TMI weights (Stamina = 1). Both sets sit at
// or past the 8% hit cap, and hit is its value past it. Block rating measures within noise of zero on both.
// ArmorNoMods weighs as Armor; the healing model's Inspiration scales only Armor, and the two measure less than 0.01
// apart.
export const PreBiSOffensiveEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Offensive',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.62,
			[Stat.StatAgility]: 0.58,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 0.05,
			[Stat.StatCritRating]: 0.44,
			[Stat.StatHasteRating]: 0.27,
			[Stat.StatExpertiseRating]: 3.02,
			[Stat.StatArmorPenetration]: 0.12,
			[Stat.StatDefenseRating]: 0.43,
			[Stat.StatBlockRating]: 0,
			[Stat.StatBlockValue]: 0.6,
			[Stat.StatDodgeRating]: 0.21,
			[Stat.StatParryRating]: 0.21,
			[Stat.StatArmor]: 0.05,
			[Stat.StatBonusArmor]: 0.05,
			[Stat.StatArmorNoMods]: 0.05,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.45,
		},
	),
);

export const PreBiSBalancedEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Balanced',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.62,
			[Stat.StatAgility]: 0.66,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 0.04,
			[Stat.StatCritRating]: 0.43,
			[Stat.StatHasteRating]: 0.17,
			[Stat.StatExpertiseRating]: 2.35,
			[Stat.StatArmorPenetration]: 0.11,
			[Stat.StatDefenseRating]: 0.53,
			[Stat.StatBlockRating]: 0,
			[Stat.StatBlockValue]: 0.76,
			[Stat.StatDodgeRating]: 0.29,
			[Stat.StatParryRating]: 0.26,
			[Stat.StatArmor]: 0.06,
			[Stat.StatBonusArmor]: 0.06,
			[Stat.StatArmorNoMods]: 0.06,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.42,
		},
	),
);

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 0,
		useBattleShout: true,
		defaultStance: WarriorStance.WarriorStanceDefensive,
		hasBsT2: false,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	offensivePotId: Consumables.MajorFrenzyPotion.id,
	elixirIds: [
		Consumables.ElixirOfFerocity.id,
		Consumables.ElixirOfTheMongoose.id,
		Consumables.WinterfallFirewater.id,
		Consumables.ElixirOfGreaterDefense.id,
		Consumables.ElixirOfThePhalanx.id,
	],
	mhImbueId: Consumables.DenseWeightstone.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.RumseyRum.id],
});

// Classic Era Molten Core logs: the healing that lands on the main tank, overhealing included, comes to 2.4 x
// its damage taken, in heals about 0.3 s apart, 1.5% of it as absorbs. The HPS is 2.4 x the 318 damage taken per
// second of the Balanced Pre-BiS gear against the default target, rounded down to 760; the Inspiration uptime is not
// from the logs.
export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 760,
		cadenceSeconds: 0.3,
		cadenceVariation: 0.64,
		absorbFrac: 0.015,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};

export const PreBiSOffensiveBuild = PresetUtils.makePresetBuild('Pre-BiS', {
	group: 'Offensive',
	gear: PreBiSOffensiveGear,
	talents: OffensiveTalents,
	rotation: ROTATION_DEFAULT,
	epWeights: PreBiSOffensiveEpWeights,
	settings: { name: 'Pre-BiS - Offensive', consumables: DefaultConsumables },
});

export const PreBiSBalancedBuild = PresetUtils.makePresetBuild('Pre-BiS', {
	group: 'Balanced',
	gear: PreBiSBalancedGear,
	talents: BalancedTalents,
	rotation: ROTATION_DEFAULT,
	epWeights: PreBiSBalancedEpWeights,
	settings: { name: 'Pre-BiS - Balanced', consumables: DefaultConsumables },
});
