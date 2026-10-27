import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import {
	Hunter_Rotation_Playstyle as Playstyle,
	HunterOptions_Ammo,
	HunterOptions_PetAttackSpeed,
	HunterOptions_PetType,
	HunterOptions_QuiverBonus,
} from '@generated/proto/hunter';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { HunterSpecs } from '@sim/proto/spec_types';
import type { CustomSection } from '@sim/spec_config';
import { getEnumValues } from '@sim/utils/collections';

// The arrows the client's items describe, as the Ammo enum lists them: each value is the arrow's item
// id, so the icon and the hover tooltip come from the item. Every ranged weapon is taken to fire them.
export const AmmoInput = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, HunterOptions_Ammo>({
		fieldName: 'ammo',
		label: i18n.t('settings_tab.other.ammo.label'),
		labelTooltip: i18n.t('settings_tab.other.ammo.tooltip'),
		numColumns: 4,
		values: [
			{ value: HunterOptions_Ammo.AmmoNone, tooltip: i18n.t('settings_tab.other.ammo.no_ammo') },
			...getEnumValues<HunterOptions_Ammo>(HunterOptions_Ammo)
				.filter(value => value !== HunterOptions_Ammo.AmmoNone)
				.map(value => ({ value, actionId: ActionId.fromItemId(value) })),
		],
	});

// The quivers the client's items describe, one a ranged attack speed, listed from the QuiverBonus
// enum the way AmmoInput lists the arrows.
export const QuiverInput = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, HunterOptions_QuiverBonus>({
		fieldName: 'quiverBonus',
		label: i18n.t('settings_tab.other.quiver.label'),
		labelTooltip: i18n.t('settings_tab.other.quiver.tooltip'),
		numColumns: 4,
		values: [
			{ color: '82e89d', value: HunterOptions_QuiverBonus.QuiverNone, tooltip: i18n.t('settings_tab.other.quiver.no_quiver') },
			...getEnumValues<HunterOptions_QuiverBonus>(HunterOptions_QuiverBonus)
				.filter(value => value !== HunterOptions_QuiverBonus.QuiverNone)
				.map(value => ({ value, actionId: ActionId.fromItemId(value) })),
		],
	});

export const PetTypeInput = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, HunterOptions_PetType>({
		fieldName: 'petType',
		label: i18n.t('settings_tab.other.pet_type.label'),
		labelTooltip: i18n.t('settings_tab.other.pet_type.tooltip'),
		numColumns: 5,
		values: [
			{ value: HunterOptions_PetType.PetNone, actionId: ActionId.fromPetName(''), tooltip: i18n.t('settings_tab.other.pet_type.no_pet') },
			{ value: HunterOptions_PetType.Bat, actionId: ActionId.fromPetName('Bat'), tooltip: i18n.t('settings_tab.other.pet_type.bat') },
			{ value: HunterOptions_PetType.Bear, actionId: ActionId.fromPetName('Bear'), tooltip: i18n.t('settings_tab.other.pet_type.bear') },
			{
				value: HunterOptions_PetType.BirdOfPrey,
				actionId: ActionId.fromPetName('Bird of Prey'),
				tooltip: i18n.t('settings_tab.other.pet_type.bird_of_prey'),
			},
			{ value: HunterOptions_PetType.Boar, actionId: ActionId.fromPetName('Boar'), tooltip: i18n.t('settings_tab.other.pet_type.boar') },
			{
				value: HunterOptions_PetType.CarrionBird,
				actionId: ActionId.fromPetName('Carrion Bird'),
				tooltip: i18n.t('settings_tab.other.pet_type.carrion_bird'),
			},
			{ value: HunterOptions_PetType.Cat, actionId: ActionId.fromPetName('Cat'), tooltip: i18n.t('settings_tab.other.pet_type.cat') },
			{ value: HunterOptions_PetType.CoreHound, actionId: ActionId.fromPetName('Core Hound'), tooltip: i18n.t('settings_tab.other.pet_type.core_hound') },
			{ value: HunterOptions_PetType.Crab, actionId: ActionId.fromPetName('Crab'), tooltip: i18n.t('settings_tab.other.pet_type.crab') },
			{ value: HunterOptions_PetType.Crocilisk, actionId: ActionId.fromPetName('Crocilisk'), tooltip: i18n.t('settings_tab.other.pet_type.crocolisk') },
			{ value: HunterOptions_PetType.Fox, actionId: ActionId.fromPetName('Fox'), tooltip: i18n.t('settings_tab.other.pet_type.fox') },
			{ value: HunterOptions_PetType.Gorilla, actionId: ActionId.fromPetName('Gorilla'), tooltip: i18n.t('settings_tab.other.pet_type.gorilla') },
			{ value: HunterOptions_PetType.Hyena, actionId: ActionId.fromPetName('Hyena'), tooltip: i18n.t('settings_tab.other.pet_type.hyena') },
			{ value: HunterOptions_PetType.Raptor, actionId: ActionId.fromPetName('Raptor'), tooltip: i18n.t('settings_tab.other.pet_type.raptor') },
			{ value: HunterOptions_PetType.Scorpid, actionId: ActionId.fromPetName('Scorpid'), tooltip: i18n.t('settings_tab.other.pet_type.scorpid') },
			{ value: HunterOptions_PetType.Spider, actionId: ActionId.fromPetName('Spider'), tooltip: i18n.t('settings_tab.other.pet_type.spider') },
			{
				value: HunterOptions_PetType.Tallstrider,
				actionId: ActionId.fromPetName('Tallstrider'),
				tooltip: i18n.t('settings_tab.other.pet_type.tallstrider'),
			},
			{ value: HunterOptions_PetType.Turtle, actionId: ActionId.fromPetName('Turtle'), tooltip: i18n.t('settings_tab.other.pet_type.turtle') },
			{
				value: HunterOptions_PetType.WindSerpent,
				actionId: ActionId.fromPetName('Wind Serpent'),
				tooltip: i18n.t('settings_tab.other.pet_type.wind_serpent'),
			},
			{ value: HunterOptions_PetType.Wolf, actionId: ActionId.fromPetName('Wolf'), tooltip: i18n.t('settings_tab.other.pet_type.wolf') },
		],
	});

// The Faster Attack or Slower Attack passive the tamed creature carries on top of the 2 s swing.
// Every rank shares one icon, so the rank is written on it.
export const PetAttackSpeed = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, HunterOptions_PetAttackSpeed>({
		fieldName: 'petAttackSpeed',
		label: i18n.t('settings_tab.other.pet_attack_speed.label'),
		labelTooltip: i18n.t('settings_tab.other.pet_attack_speed.tooltip'),
		numColumns: 5,
		values: [
			{ value: HunterOptions_PetAttackSpeed.PetAttackSpeedNone, tooltip: i18n.t('settings_tab.other.pet_attack_speed.none') },
			{
				value: HunterOptions_PetAttackSpeed.SlowerAttackIII,
				actionId: ActionId.fromSpellId(1263114),
				text: 'III',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.slower_attack', { rank: 'III', percent: '-26%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.SlowerAttackII,
				actionId: ActionId.fromSpellId(1263113),
				text: 'II',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.slower_attack', { rank: 'II', percent: '-20%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackI,
				actionId: ActionId.fromSpellId(1263099),
				text: 'I',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'I', percent: '+17%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackII,
				actionId: ActionId.fromSpellId(1263100),
				text: 'II',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'II', percent: '+25%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackIII,
				actionId: ActionId.fromSpellId(1263102),
				text: 'III',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'III', percent: '+33%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackIV,
				actionId: ActionId.fromSpellId(1263104),
				text: 'IV',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'IV', percent: '+43%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackV,
				actionId: ActionId.fromSpellId(1263107),
				text: 'V',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'V', percent: '+54%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackVI,
				actionId: ActionId.fromSpellId(1263109),
				text: 'VI',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'VI', percent: '+67%' }),
			},
			{
				value: HunterOptions_PetAttackSpeed.FasterAttackVII,
				actionId: ActionId.fromSpellId(1263110),
				text: 'VII',
				tooltip: i18n.t('settings_tab.other.pet_attack_speed.faster_attack', { rank: 'VII', percent: '+100%' }),
			},
		],
	});

// Beast Training, the part of it that changes the pet's damage. Cobra Reflexes is a flag, but it is
// picked the way its neighbours are so the four read alike: label, then icon.
export const CobraReflexes = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, number>({
		fieldName: 'cobraReflexes',
		label: i18n.t('settings_tab.other.pet_training.cobra_reflexes.label'),
		labelTooltip: i18n.t('settings_tab.other.pet_training.cobra_reflexes.tooltip'),
		values: [
			{ value: 0, tooltip: i18n.t('settings_tab.other.pet_training.cobra_reflexes.untrained') },
			{ value: 1, actionId: ActionId.fromSpellId(25077), tooltip: i18n.t('settings_tab.other.pet_training.cobra_reflexes.trained') },
		],
		getValue: (player: Player<SpecType>) => (player.getClassOptions().cobraReflexes ? 1 : 0),
		setValue: (player: Player<SpecType>, newValue: number) => {
			const options = player.getClassOptions();
			options.cobraReflexes = newValue > 0;
			player.setClassOptions(options);
		},
	});

// One spell per rank, so the wowhead tooltip is the rank's own.
export const PetAggression = <SpecType extends HunterSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, number>({
		fieldName: 'petAggression',
		label: i18n.t('settings_tab.other.pet_training.pet_aggression.label'),
		labelTooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.tooltip'),
		numColumns: 3,
		values: [
			{ value: 0, tooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.untrained') },
			{ value: 1, actionId: ActionId.fromSpellId(6311), text: '1', tooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.rank', { rank: 1 }) },
			{ value: 2, actionId: ActionId.fromSpellId(6314), text: '2', tooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.rank', { rank: 2 }) },
			{ value: 3, actionId: ActionId.fromSpellId(6315), text: '3', tooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.rank', { rank: 3 }) },
			{ value: 4, actionId: ActionId.fromSpellId(6316), text: '4', tooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.rank', { rank: 4 }) },
			{ value: 5, actionId: ActionId.fromSpellId(6317), text: '5', tooltip: i18n.t('settings_tab.other.pet_training.pet_aggression.rank', { rank: 5 }) },
		],
	});

export const PetUptime = () =>
	InputHelpers.makeClassOptionsNumberInput<Spec.SpecHunter>({
		fieldName: 'petUptime',
		label: i18n.t('settings_tab.other.pet_uptime.label'),
		labelTooltip: i18n.t('settings_tab.other.pet_uptime.tooltip'),
		percent: true,
	});

// The pet's own settings block: the family and what the tamed creature carries, as a grid of
// labelled icon pickers, with the uptime under them.
export const PetSection: CustomSection<Spec.SpecHunter> = {
	id: 'hunter-pet-settings',
	title: i18n.t('settings_tab.other.pet_section.title'),
	iconGroupClassName: 'grid grid-cols-2 gap-x-6 gap-y-2',
	iconInputs: [PetTypeInput(), PetAttackSpeed(), CobraReflexes(), PetAggression()],
	inputs: [PetUptime()],
};

const isRanged = (player: Player<Spec.SpecHunter>) => player.getSimpleRotation().playstyle !== Playstyle.Melee;

// One simple rotation: the playstyle picks the priority list, the toggles the spells it casts.
export const RotationInputs = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecHunter, Playstyle>({
			fieldName: 'playstyle',
			label: i18n.t('rotation_tab.options.hunter.playstyle.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.playstyle.tooltip'),
			values: [
				{ name: i18n.t('rotation_tab.options.hunter.playstyle.values.turret'), value: Playstyle.Turret },
				{ name: i18n.t('rotation_tab.options.hunter.playstyle.values.weave'), value: Playstyle.Weave },
				{ name: i18n.t('rotation_tab.options.hunter.playstyle.values.melee'), value: Playstyle.Melee },
			],
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecHunter>({
			fieldName: 'pullRange',
			label: i18n.t('rotation_tab.options.hunter.pull_range.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.pull_range.tooltip'),
			positive: true,
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecHunter>({
			fieldName: 'timeToWeave',
			label: i18n.t('rotation_tab.options.hunter.time_to_weave.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.time_to_weave.tooltip'),
			positive: true,
			showWhen: (player: Player<Spec.SpecHunter>) => player.getSimpleRotation().playstyle === Playstyle.Weave,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecHunter>({
			fieldName: 'useAimed',
			label: i18n.t('rotation_tab.options.hunter.use_aimed.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.use_aimed.tooltip'),
			showWhen: isRanged,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecHunter>({
			fieldName: 'useMulti',
			label: i18n.t('rotation_tab.options.hunter.use_multi.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.use_multi.tooltip'),
			showWhen: isRanged,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecHunter>({
			fieldName: 'useArcane',
			label: i18n.t('rotation_tab.options.hunter.use_arcane.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.use_arcane.tooltip'),
			showWhen: isRanged,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecHunter>({
			fieldName: 'useSniper',
			label: i18n.t('rotation_tab.options.hunter.use_sniper.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.use_sniper.tooltip'),
			showWhen: isRanged,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecHunter>({
			fieldName: 'useHawk',
			label: i18n.t('rotation_tab.options.hunter.use_hawk.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.use_hawk.tooltip'),
		}),
	],
};
