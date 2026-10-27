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
import PreBiSArmsGearJson from './gear_sets/prebis_arms.gear.json';
import PreBiSFuryGearJson from './gear_sets/prebis_fury.gear.json';

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

// Dungeon and world drops, the warrior's dungeon tier sets and crafted items whose reagents include no
// raid materials; no reputation or PvP rewards. The shoulders and bracers are world-drop greens with the
// "of the Tiger" suffix, and the neck is a Darkmoon Faire ticket reward. The cloak, the first ring and the
// Adaptive Combat Assistant are sold by the Undermine Trader in Stranglethorn Vale; the Dawn Armor pattern and
// the second ring have no known source.
export const PreBiSFuryGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSFuryGearJson, { ...FURY_PRESET_OPTIONS, devOnly: true });

// The best build on the Pre-BiS Fury gear over fights of 60 to 240 s. Enrage takes no points: with no tank
// assigned the boss attacks no one, so the sim gives it no value.
export const FuryTalents = PresetUtils.makePresetTalents('Fury', SavedTalents.create({ talentsString: '30305213002-15053105102210501' }));

// Measured on the Pre-BiS Fury gear with the Fury talents.
export const PreBiSFuryEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 2.25,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.41,
			[Stat.StatHitRating]: 2.29,
			[Stat.StatHasteRating]: 2.29,
			[Stat.StatCritRating]: 2.92,
			[Stat.StatExpertiseRating]: 1.9,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 9.89,
			[PseudoStat.PseudoStatOffHandDps]: 3.15,
		},
	),
);

// The same sources as the Pre-BiS Fury gear; the cloak and the bow are world-drop greens with the "of the
// Tiger" suffix too.
export const PreBiSArmsGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSArmsGearJson, { ...ARMS_PRESET_OPTIONS, devOnly: true });

// The Battle Stance build: Rend keeps Bloodthrill's Overpower procs coming, and Spearing Strike fills the
// global cooldowns Mortal Strike and Slam leave.
export const ArmsTalents = PresetUtils.makePresetTalents('Arms', SavedTalents.create({ talentsString: '30325213132515201-0505-2' }));

// Measured on the Pre-BiS Arms gear with the Arms talents. Hit is its value below the 8% cap.
export const PreBiSArmsEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.82,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.45,
			[Stat.StatHitRating]: 4.59,
			[Stat.StatHasteRating]: 2,
			[Stat.StatCritRating]: 2.36,
			[Stat.StatExpertiseRating]: 2.23,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 17.69,
			[PseudoStat.PseudoStatOffHandDps]: 0,
		},
	),
);

export const FURY_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Fury', DefaultFuryApl);
export const ARMS_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Arms', DefaultArmsApl);

export const SIMPLE_ROTATION = DpsWarrior_Rotation.create({
	spec: DpsWarriorSpec.DpsWarriorSpecFury,
	sunderArmor: WarriorSunder.WarriorSunderHelp,
	useOverpower: true,
	useRend: true,
	useRecklessness: false,
	useStanceDance: false,
	useHamstring: true,
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
		hasBsT2: true,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	resourcePotId: Consumables.MightyRagePotion.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	ohImbueId: Consumables.DenseSharpeningStone.id,
	elixirIds: [Consumables.ElixirOfFerocity.id, Consumables.ElixirOfTheMongoose.id, Consumables.WinterfallFirewater.id],
});

// The Pre-BiS Arms two-hander is an axe, so the main hand takes the sharpening stone and there is no off hand.
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

export const PreBiSArmsBuild = PresetUtils.makePresetBuild('Pre-BiS', {
	group: 'Arms',
	gear: PreBiSArmsGear,
	talents: ArmsTalents,
	rotation: SIMPLE_ARMS_DEFAULT_ROTATION,
	epWeights: PreBiSArmsEpWeights,
	settings: { name: 'Pre-BiS - Arms', consumables: ArmsConsumables },
});
