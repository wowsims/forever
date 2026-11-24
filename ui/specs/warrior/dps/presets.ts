import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { makeSpecChangeWarningToast } from '@features/settings/utils/spec_change_warning_toast';
import { ConsumesSpec, HandType, ItemSlot, Profession, PseudoStat, Race, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { DpsWarrior_Options as WarriorOptions, DpsWarrior_Rotation, DpsWarriorSpec, WarriorStance, WarriorSunder } from '@generated/proto/warrior';
import { Player } from '@sim/player/player';
import type { PresetSettings } from '@sim/presets/types';
import { Stats } from '@sim/proto/stats';

import * as WarriorPresets from '../shared/presets';
import DefaultArmsApl from './apls/arms.apl.json';
import DefaultFuryApl from './apls/fury.apl.json';
import DefaultFury2HApl from './apls/fury_2h.apl.json';
import PreBiSArmsGearJson from './gear_sets/prebis_arms.gear.json';
import PreBiSFuryGearJson from './gear_sets/prebis_fury.gear.json';
import PreBiSFury2HGearJson from './gear_sets/prebis_fury_2h.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const isArmsSpec = (player: Player<Spec.SpecDpsWarrior>) =>
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeTwoHand;

export const isArmsKebabSpec = (player: Player<Spec.SpecDpsWarrior>) => player.getTalents().mortalStrike && isFurySpec(player);

export const isFurySpec = (player: Player<Spec.SpecDpsWarrior>) =>
	player.getTalents().bloodthirst ||
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeMainHand ||
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeOneHand;

const FURY_PRESET_OPTIONS = {
	group: 'Fury',
	onLoad: (player: Player<Spec.SpecDpsWarrior>) => {
		makeSpecChangeWarningToast(
			[
				{
					condition: isArmsSpec,
					message: 'Check your gear: You have a two-handed weapon equipped, but the selected option is for dual wield.',
				},
				{
					condition: (player: Player<Spec.SpecDpsWarrior>) => !player.getTalents().dualWieldSpecialization,
					message: "Check your talents: You have selected a dual-wield spec but don't have [Dual Wield Specialization] talented.",
				},
			],
			player,
		);
	},
};

const ARMS_PRESET_OPTIONS = {
	group: 'Arms',
	onLoad: (player: Player<Spec.SpecDpsWarrior>) => {
		makeSpecChangeWarningToast(
			[
				{
					condition: isFurySpec,
					message: 'Check your gear: You have a one-handed weapon equipped, but the selected option is for two-handed weapons.',
				},
			],
			player,
		);
	},
};

// Dungeon and world drops, quest rewards, vendor items and crafted items whose reagents include no raid
// materials; no raid, raid quest, reputation or PvP rewards, except Alterac Valley's: Don Julio's Band, which its
// quartermasters sell, and Bloodseeker, a quest reward. Furbolg Berserker's Boots have no known source. The cloak, the
// belt and the first ring are sold by the Undermine Trader in Stranglethorn Vale. The Devilsaur Gauntlets and Leggings make the two-piece set. Persuader and Flurry Axe beat Ironfoe in
// either hand up to two targets; Flurry Axe, the off hand, is a world-drop epic.
export const PreBiSFuryGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSFuryGearJson, FURY_PRESET_OPTIONS);

// The Fury build for fights of 60 to 300 s.
export const FuryTalents = PresetUtils.makePresetTalents('Fury', SavedTalents.create({ talentsString: '30303213002-050530045122100231' }));

// The Fury build against a Giant or Dragonkin boss: an Enrage point goes to Spearing Strike, which deals extra
// damage to both, and the Fury rotation swaps to Battle Stance for it. On other targets it does less than Fury.
export const FuryDragonkinGiantTalents = PresetUtils.makePresetTalents(
	'Fury Dragonkin/Giant',
	SavedTalents.create({ talentsString: '30303213102-050530035122100231' }),
);

// Measured on Pre-BiS Fury gear with Gauntlets of Heroism, Titanic Leggings and Hand of Justice, and with Fury
// talents that take Booming Voice and no Enrage.
export const PreBiSFuryEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.7,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.39,
			[Stat.StatHitRating]: 1.69,
			[Stat.StatHasteRating]: 1.76,
			[Stat.StatCritRating]: 2.26,
			[Stat.StatExpertiseRating]: 2.35,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 12.59,
			[PseudoStat.PseudoStatOffHandDps]: 2.66,
		},
	),
);

// The Fury 2H set with Blackhand's Breadth in place of Hand of Justice and the Strength necklace enchant in place of
// Agility. Destiny, the weapon, has no known source, and its Strength proc rate comes from logs (EV-0132). It beats
// Arcanite Champion, the Blacksmithing craft, by 4 to 5 %.
export const PreBiSArmsGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSArmsGearJson, ARMS_PRESET_OPTIONS);

// The Battle Stance build: Rend keeps Bloodthrill's Overpower procs coming, and Spearing Strike fills the
// global cooldowns Mortal Strike and Slam leave.
export const ArmsTalents = PresetUtils.makePresetTalents('Arms', SavedTalents.create({ talentsString: '30303213132215201-55050000002-2' }));

// Measured on the Pre-BiS Arms gear and talents with Stance Dance on. The gear sits at the 8% hit cap; Hit is its value
// below it.
export const PreBiSArmsEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.76,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.46,
			[Stat.StatHitRating]: 5.03,
			[Stat.StatHasteRating]: 1.29,
			[Stat.StatCritRating]: 2.32,
			[Stat.StatExpertiseRating]: 1.57,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 16.71,
			[PseudoStat.PseudoStatOffHandDps]: 0,
		},
	),
);

const FURY_2H_PRESET_OPTIONS = {
	group: 'Fury 2H',
	onLoad: (player: Player<Spec.SpecDpsWarrior>) => {
		makeSpecChangeWarningToast(
			[
				{
					condition: (player: Player<Spec.SpecDpsWarrior>) => !player.getTalents().twoHandedWeaponSpecialization,
					message: "Check your talents: You have selected a two-hander spec but don't have [Two-Handed Weapon Specialization] talented.",
				},
			],
			player,
		);
	},
};

// The Pre-BiS Fury sources with Destiny in place of the dual-wield pair. Don Julio's Band and Bloodseeker are Alterac
// Valley rewards. Destiny has no known source; its Strength proc rate comes from logs (EV-0132). Hand of Justice beats
// Blackhand's Breadth on fights of 180 s and longer; Blackhand's Breadth wins up to 120 s.
export const PreBiSFury2HGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSFury2HGearJson, FURY_2H_PRESET_OPTIONS);

// The Fury build for a two-hander: Two-Handed Weapon Specialization, Blood Craze, a Gore Drinker point and a fifth
// Enrage point in place of Dual Wield Specialization and Furious Precision.
export const Fury2HTalents = PresetUtils.makePresetTalents('Fury 2H', SavedTalents.create({ talentsString: '30303213032-050500350122110231' }));

// Measured on the Pre-BiS Fury 2H gear and talents. The gear sits at the 8% hit cap; Hit is its value below it.
export const PreBiSFury2HEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Fury 2H',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.94,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.46,
			[Stat.StatHitRating]: 5.17,
			[Stat.StatHasteRating]: 2.06,
			[Stat.StatCritRating]: 2.55,
			[Stat.StatExpertiseRating]: 2.26,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.89,
			[PseudoStat.PseudoStatOffHandDps]: 0,
		},
	),
);

// The rotations have no Autocast Other Cooldowns, so an on-use trinket fires only from the "CD: Trinkets" group.
// Blackhand's Breadth has no self buff for Multiple CD Usages, so its two lines there time the end of the fight with
// Remaining Time and, with Recklessness on, keep the mark from running under it.
export const FURY_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Fury', DefaultFuryApl);
// The Fury rotation with a two-hander's thresholds: Heroic Strike from 70 rage and Cleave from 30, Hamstring without
// the Delay Check and on up to two targets, Overpower swaps up to 70 rage and Rend swaps up to 30 on any number of
// targets.
export const FURY_2H_ROTATION = PresetUtils.makePresetAPLRotation('Fury 2H', DefaultFury2HApl);
export const ARMS_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Arms', DefaultArmsApl);

export const SIMPLE_ROTATION = DpsWarrior_Rotation.create({
	spec: DpsWarriorSpec.DpsWarriorSpecFury,
	sunderArmor: WarriorSunder.WarriorSunderHelp,
	useOverpower: true,
	useRend: true,
	useRecklessness: false,
	useStanceDance: true,
	useHamstring: true,
	useHeroicStrikeRank9: false,
	useBattleShoutRank7: false,
});
export const SIMPLE_DEFAULT_ROTATION = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecDpsWarrior, SIMPLE_ROTATION);
export const SIMPLE_ARMS_DEFAULT_ROTATION = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecDpsWarrior, {
	...SIMPLE_ROTATION,
	spec: DpsWarriorSpec.DpsWarriorSpecArms,
});

export const DefaultOptions = WarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 50,
		useBattleShout: true,
		defaultStance: WarriorStance.WarriorStanceBerserker,
		hasBsT2: false,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	foodId: Consumables.SmokedDesertDumplings.id,
	// A dynamite states no stun, so a level 63 boss does not shrug it off the way it does the EZ-Thro Dark Bomb.
	explosiveId: Consumables.SAFTCleverDynamite.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	ohImbueId: Consumables.DenseSharpeningStone.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.JujuPower.id, Consumables.JujuMight.id, Consumables.BoglingRoot.id, Consumables.ROIDS.id],
	elixirIds: [Consumables.ElixirOfTheMongoose.id],
	resourcePotId: Consumables.MightyRagePotion.id,
});

// Destiny, the two-hander of the Arms and Fury 2H Pre-BiS sets, is a sword, so the main hand takes the sharpening
// stone and there is no off hand.
export const ArmsConsumables = ConsumesSpec.create({
	...DefaultConsumables,
	mhImbueId: Consumables.DenseSharpeningStone.id,
	ohImbueId: 0,
});

export const OtherDefaults = {
	race: Race.RaceOrc,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 25,
	reactionTime: 100,
};

export const DefaultSettings: PresetSettings = {
	name: 'Defaults',
	race: OtherDefaults.race,
	specOptions: DefaultOptions,
	consumables: DefaultConsumables,
	buffs: WarriorPresets.DefaultIndividualBuffs,
	partyBuffs: WarriorPresets.DefaultPartyBuffs,
	raidBuffs: WarriorPresets.DefaultRaidBuffs,
	debuffs: WarriorPresets.DefaultDebuffs,
	playerOptions: {
		profession1: OtherDefaults.profession1,
		profession2: OtherDefaults.profession2,
		distanceFromTarget: OtherDefaults.distanceFromTarget,
		reactionTimeMs: OtherDefaults.reactionTime,
	},
};

export const PreBiSFuryBuild = PresetUtils.makePresetBuild('Pre-BiS', {
	group: 'Fury',
	gear: PreBiSFuryGear,
	talents: FuryTalents,
	rotation: SIMPLE_DEFAULT_ROTATION,
	epWeights: PreBiSFuryEpWeights,
	settings: { name: 'Pre-BiS - Fury', consumables: DefaultConsumables },
});

export const PreBiSFury2HBuild = PresetUtils.makePresetBuild('Pre-BiS', {
	group: 'Fury 2H',
	gear: PreBiSFury2HGear,
	talents: Fury2HTalents,
	rotation: SIMPLE_DEFAULT_ROTATION,
	epWeights: PreBiSFury2HEpWeights,
	settings: { name: 'Pre-BiS - Fury 2H', consumables: ArmsConsumables },
});

export const PreBiSArmsBuild = PresetUtils.makePresetBuild('Pre-BiS', {
	group: 'Arms',
	gear: PreBiSArmsGear,
	talents: ArmsTalents,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	epWeights: PreBiSArmsEpWeights,
	settings: { name: 'Pre-BiS - Arms', consumables: ArmsConsumables },
});
