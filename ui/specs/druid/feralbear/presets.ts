import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import DefaultApl from './apls/default.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL', DefaultApl);

// Nature's Reach (melee hit) and Nature's Majesty (melee crit) are worth more to the bear than Furor and
// Naturalist. Feral Instinct raises only Swipe, which the single-target APL does not cast.
export const BearTalents = PresetUtils.makePresetTalents(
	'Bear',
	SavedTalents.create({ talentsString: '0500220000000000-55200322120132012551-0000000000000000' }),
);
// Tanks and plays cat on one set of talents. As a bear, about the same threat as the Bear talents and a
// TMI 2 to 3 higher, for no Natural Reaction.
export const HybridTalents = PresetUtils.makePresetTalents(
	'Hybrid 9/42/0',
	SavedTalents.create({ talentsString: '0500220000000000-35230332121132212051-0000000000000000' }),
);

// Leather at 60 with no set bonus and no proc, about the item level of the cat preset.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. The idol slot is empty: every idol has an effect.',
});

// The raid target swings for 6000 before armor.
const level60Encounter = Encounter.defaultEncounterProto();
level60Encounter.targets[0].minBaseDamage = 6000;
export const Level60Encounter = {
	...PresetUtils.makePresetEncounter('Level 60 boss', level60Encounter),
	tooltip: 'The raid target, hitting for 6000 before armor.',
};

// EP estimates with agility as 1. A bear's agility is crit, armor and dodge but no attack
// power, so strength is worth about two agility. Dodge costs damage: fewer hits taken is less rage.
// Defense nets to 0, its Thick Hide armor against the hits it avoids. Stamina keeps the bear alive longer.
// Feral attack power uses attack power's weight through the form-only dependency. In each preset
// ArmorNoMods weighs the armor weight over 4.6: Dire Bear Form makes a point of armor from items 4.6
// points of armor and leaves ArmorNoMods as it is.
export const Level60DpsEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 2.0714,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1103,
		[Stat.StatAttackPower]: 0.9415,
		[Stat.StatFeralAttackPower]: 0.9415,
		[Stat.StatArmor]: 0.026,
		[Stat.StatBonusArmor]: 0.026,
		[Stat.StatArmorNoMods]: 0.0057,
		[Stat.StatDodgeRating]: -0.1059,
		[Stat.StatDefenseRating]: 0,
		[Stat.StatHitRating]: 3.7805,
		[Stat.StatCritRating]: 1.3776,
		[Stat.StatHasteRating]: 0.8425,
		[Stat.StatExpertiseRating]: 7.1745,
		[Stat.StatPhysicalDamage]: 6.0513,
		[Stat.StatArmorPenetration]: 0.3601,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 13.1813,
	},
);

// Threat with agility as 1.
export const Level60ThreatEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 2.005,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1084,
		[Stat.StatAttackPower]: 0.9114,
		[Stat.StatFeralAttackPower]: 0.9114,
		[Stat.StatArmor]: 0.0255,
		[Stat.StatBonusArmor]: 0.0255,
		[Stat.StatArmorNoMods]: 0.0055,
		[Stat.StatDodgeRating]: -0.1069,
		[Stat.StatDefenseRating]: 0,
		[Stat.StatHitRating]: 3.7022,
		[Stat.StatCritRating]: 1.3788,
		[Stat.StatHasteRating]: 0.821,
		[Stat.StatExpertiseRating]: 7.0002,
		[Stat.StatPhysicalDamage]: 5.8673,
		[Stat.StatArmorPenetration]: 0.3485,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 12.7593,
	},
);

// Survival is TMI with stamina as 1, as the EP dialog scales it for the bear. Hit, crit, haste and
// physical damage moved TMI by less than 0.04, so they are 0. The bear forms scale item armor and bonus
// armor alike, so the two weigh the same. Expertise is kept: fewer parries mean fewer parry-hasted boss
// swings.
export const Level60SurvivalEp = Stats.fromMap({
	[Stat.StatStamina]: 1,
	[Stat.StatAgility]: 0.1964,
	[Stat.StatArmor]: 0.1617,
	[Stat.StatBonusArmor]: 0.1617,
	[Stat.StatArmorNoMods]: 0.0352,
	[Stat.StatDodgeRating]: 0.1777,
	[Stat.StatDefenseRating]: 0.8496,
	[Stat.StatExpertiseRating]: 0.3212,
});

export const Level60DpsEpWeights = PresetUtils.makePresetEpWeights('Level 60 DPS', Level60DpsEp);
export const Level60ThreatEpWeights = PresetUtils.makePresetEpWeights('Level 60 Threat', Level60ThreatEp);
export const Level60SurvivalEpWeights = PresetUtils.makePresetEpWeights('Level 60 Survival', Level60SurvivalEp);

// Another raid member's bleed keeps Rend and Tear up.
export const DefaultOptions = DruidOptions.create({
	startingRage: 0,
	assumeBleedActive: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.GiftOfArthas.id],
	foodId: Consumables.DirgesKickinChimaerokChops.id,
	defensivePotId: Consumables.GreaterStoneshieldPotion.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id, Consumables.ScrollOfProtectionIV.id],
});

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	distanceFromTarget: 0,
	reactionTime: 250,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
