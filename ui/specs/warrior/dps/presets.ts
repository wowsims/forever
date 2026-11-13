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

// Dungeon and world drops, quest rewards, vendor items and crafted items whose reagents include no raid
// materials; no raid, raid quest, reputation or PvP rewards. The cloak, the belt and the first ring are sold by the
// Undermine Trader in Stranglethorn Vale. The Dawn Armor pattern and its Pristine Scale and Fel Leather have no
// known source, and neither do the Gauntlets and Battleboots of Heroism.
export const PreBiSFuryGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSFuryGearJson, { ...FURY_PRESET_OPTIONS, devOnly: true });

// The Fury build for fights of 60 to 240 s. Enrage takes no points: with no tank assigned the boss attacks no
// one, so the sim gives it no value.
export const FuryTalents = PresetUtils.makePresetTalents('Fury', SavedTalents.create({ talentsString: '30303213002-350530005122100231-1' }));

// Measured on the Pre-BiS Fury gear with the Fury talents.
export const PreBiSFuryEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.91,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.4,
			[Stat.StatHitRating]: 2,
			[Stat.StatHasteRating]: 1.97,
			[Stat.StatCritRating]: 2.49,
			[Stat.StatExpertiseRating]: 2.14,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 11.48,
			[PseudoStat.PseudoStatOffHandDps]: 3.26,
		},
	),
);

// The same sources as the Pre-BiS Fury gear, with one exception: the weapon, the Premier High Warlord's
// Greatsword, a PvP reward. The shoulders are a world-drop green with the "of the Tiger" suffix. An epic
// two-hander puts the set on even terms with the Fury set's epic Persuader.
export const PreBiSArmsGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSArmsGearJson, { ...ARMS_PRESET_OPTIONS, devOnly: true });

// The Battle Stance build: Rend keeps Bloodthrill's Overpower procs coming, and Spearing Strike fills the
// global cooldowns Mortal Strike and Slam leave.
export const ArmsTalents = PresetUtils.makePresetTalents('Arms', SavedTalents.create({ talentsString: '30303213132215201-55050000002-2' }));

// Measured on the Pre-BiS Arms gear with the Arms talents. Hit is its value below the 8% cap, where the gear sits.
export const PreBiSArmsEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS - Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.72,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.44,
			[Stat.StatHitRating]: 4.09,
			[Stat.StatHasteRating]: 1.36,
			[Stat.StatCritRating]: 2.24,
			[Stat.StatExpertiseRating]: 1.67,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 17.63,
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
	...WarriorPresets.DefaultConsumables,
	resourcePotId: Consumables.MightyRagePotion.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	ohImbueId: Consumables.DenseSharpeningStone.id,
	elixirIds: [Consumables.ElixirOfFerocity.id, Consumables.ElixirOfTheMongoose.id, Consumables.WinterfallFirewater.id],
});

// The Pre-BiS Arms two-hander is a sword, so the main hand takes the sharpening stone and there is no off hand.
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
