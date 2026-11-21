import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import { PaladinAura, PaladinJudgement, PaladinSeal, PaladinSealTwist } from '@generated/proto/paladin';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { getEnumValues } from '@sim/utils/collections';

type RankedSpell = { spellId: number; rank?: number };

// The highest rank of each aura and assigned judgement; Concentration Aura has one rank.
export const AURA_SPELLS: Partial<Record<PaladinAura, RankedSpell>> = {
	[PaladinAura.RetributionAura]: { spellId: 10301, rank: 5 },
	[PaladinAura.DevotionAura]: { spellId: 10293, rank: 7 },
	[PaladinAura.ConcentrationAura]: { spellId: 19746 },
	[PaladinAura.FireResistanceAura]: { spellId: 19900, rank: 3 },
	[PaladinAura.FrostResistanceAura]: { spellId: 19898, rank: 3 },
	[PaladinAura.ShadowResistanceAura]: { spellId: 19896, rank: 3 },
};
export const JUDGE_SPELLS: Partial<Record<PaladinJudgement, RankedSpell>> = {
	[PaladinJudgement.JudgementOfTheCrusader]: { spellId: 20308, rank: 6 },
	[PaladinJudgement.JudgementOfLight]: { spellId: 20349, rank: 4 },
	[PaladinJudgement.JudgementOfWisdom]: { spellId: 20357, rank: 3 },
};
export const SEAL_SPELLS: Partial<Record<PaladinSeal, RankedSpell>> = {
	[PaladinSeal.SealOfCommand]: { spellId: 20920, rank: 5 },
	[PaladinSeal.SealOfRighteousness]: { spellId: 20293, rank: 8 },
};

const spellIcon = ({ spellId, rank }: RankedSpell) => ActionId.fromSpellId(spellId, rank);

// One yard outside Judgement's 10 yd range. With an assigned judgement, the default rotation's walk in
// waits for Judgement to come into range, and a wait whose condition already holds never ends, so a start
// inside that range never reaches melee.
export const MIN_PULL_RANGE = 11;

const MAX_ECHO_STACKS = 3;

type Twist = PaladinSealTwist.SealTwistCommand | PaladinSealTwist.SealTwistFury;

const TWIST_SPELLS: Record<Twist, RankedSpell> = {
	[PaladinSealTwist.SealTwistCommand]: SEAL_SPELLS[PaladinSeal.SealOfCommand]!,
	[PaladinSealTwist.SealTwistFury]: { spellId: 20423, rank: 7 },
};

// The Echo stacks a twist pools before each swap: of its own seal, and of Seal of Righteousness.
export const TWIST_STACKS: Record<Twist, { own: number; righteousness: number }> = {
	[PaladinSealTwist.SealTwistCommand]: { own: 3, righteousness: 1 },
	[PaladinSealTwist.SealTwistFury]: { own: 1, righteousness: 3 },
};

// The twist the rotation runs under Unsealed Wrath: Seal of Fury when picked, otherwise Seal of Command when
// it is talented, and none without it.
export const twistFor = (twist: PaladinSealTwist, sealOfCommand: boolean): PaladinSealTwist =>
	twist === PaladinSealTwist.SealTwistFury
		? PaladinSealTwist.SealTwistFury
		: sealOfCommand
			? PaladinSealTwist.SealTwistCommand
			: PaladinSealTwist.SealTwistUnknown;

const playerTwist = (player: Player<Spec.SpecRetributionPaladin>): PaladinSealTwist =>
	twistFor(player.getSimpleRotation().uwTwist, player.getTalents().sealOfCommand);

// An unset stack count is the twist's own default.
export const echoStacks = (stacks: number, fallback: number): number => Math.min(MAX_ECHO_STACKS, stacks || fallback);

// The Echo of Righteousness stacks pooled before the swap to the twisted seal. With no twist the count is unused and
// reads as Seal of Command's.
export const righteousnessStacks = (stacks: number, twist: PaladinSealTwist): number =>
	echoStacks(stacks, TWIST_STACKS[twist === PaladinSealTwist.SealTwistFury ? twist : PaladinSealTwist.SealTwistCommand].righteousness);

const ECHO_STACK_VALUES = Array.from({ length: MAX_ECHO_STACKS }, (_, index) => ({ name: String(index + 1), value: index + 1 }));

// Each option shows the icon and tooltip of the spell it casts.
export const RotationIconInputs = [
	InputHelpers.makeRotationEnumIconInput<Spec.SpecRetributionPaladin, PaladinAura>({
		fieldName: 'aura',
		label: i18n.t('rotation_tab.options.paladin.retribution.aura.label'),
		labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.aura.tooltip'),
		numColumns: 3,
		getValue: player => player.getSimpleRotation().aura,
		values: getEnumValues<PaladinAura>(PaladinAura)
			.filter(value => value !== PaladinAura.AuraUnknown)
			.map(value => ({ value, actionId: spellIcon(AURA_SPELLS[value]!) })),
	}),
	InputHelpers.makeRotationEnumIconInput<Spec.SpecRetributionPaladin, PaladinJudgement>({
		fieldName: 'assignedJudgement',
		label: i18n.t('rotation_tab.options.paladin.retribution.assigned_judge.label'),
		labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.assigned_judge.tooltip'),
		numColumns: 2,
		getValue: player => player.getSimpleRotation().assignedJudgement,
		values: [
			{ value: PaladinJudgement.JudgementNone, tooltip: i18n.t('rotation_tab.options.paladin.retribution.assigned_judge.none') },
			...getEnumValues<PaladinJudgement>(PaladinJudgement)
				.filter(value => value !== PaladinJudgement.JudgementNone)
				.map(value => ({ value, actionId: spellIcon(JUDGE_SPELLS[value]!) })),
		],
	}),
	InputHelpers.makeRotationEnumIconInput<Spec.SpecRetributionPaladin, PaladinSeal>({
		fieldName: 'seal',
		label: i18n.t('rotation_tab.options.paladin.retribution.seal.label'),
		labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.seal.tooltip'),
		numColumns: 2,
		getValue: player => {
			const seal = player.getSimpleRotation().seal;
			if (seal === PaladinSeal.SealUnknown) {
				const talents = player.getTalents();
				return talents.holyShock
					? PaladinSeal.SealOfRighteousness
					: talents.sealOfCommand
						? PaladinSeal.SealOfCommand
						: PaladinSeal.SealOfRighteousness;
			}

			return seal;
		},
		showWhen: player => {
			const { sealOfCommand, unsealedWrath } = player.getTalents();
			return !unsealedWrath && sealOfCommand;
		},
		values: getEnumValues<PaladinSeal>(PaladinSeal)
			.filter(value => value !== PaladinSeal.SealUnknown)
			.map(value => ({ value, actionId: spellIcon(SEAL_SPELLS[value]!) })),
	}),
	InputHelpers.makeRotationEnumIconInput<Spec.SpecRetributionPaladin, PaladinSealTwist>({
		fieldName: 'uwTwist',
		label: i18n.t('rotation_tab.options.paladin.retribution.uw_twist.label'),
		labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.uw_twist.tooltip'),
		numColumns: 2,
		getValue: playerTwist,
		// Picking another twist starts it from its own Echo stacks. Without Unsealed Wrath the picker is hidden, and a
		// hidden icon picker writes its zero value on the next rotation change, so the twist is kept for when the
		// talent comes back.
		setValue: (player: Player<Spec.SpecRetributionPaladin>, newValue: PaladinSealTwist) => {
			if (!player.getTalents().unsealedWrath || newValue === playerTwist(player)) return;
			const rotation = player.getSimpleRotation();
			rotation.uwTwist = newValue;
			if (newValue === PaladinSealTwist.SealTwistCommand) {
				rotation.uwSocStacks = TWIST_STACKS[newValue].own;
				rotation.uwSorStacks = TWIST_STACKS[newValue].righteousness;
			} else if (newValue === PaladinSealTwist.SealTwistFury) {
				rotation.uwSofStacks = TWIST_STACKS[newValue].own;
				rotation.uwSorStacks = TWIST_STACKS[newValue].righteousness;
			}
			player.setSimpleRotation(rotation);
		},
		showWhen: player => player.getTalents().unsealedWrath,
		values: [
			{
				value: PaladinSealTwist.SealTwistUnknown,
				tooltip: i18n.t('rotation_tab.options.paladin.retribution.uw_twist.none'),
				showWhen: player => !player.getTalents().sealOfCommand,
			},
			{
				value: PaladinSealTwist.SealTwistCommand,
				actionId: spellIcon(TWIST_SPELLS[PaladinSealTwist.SealTwistCommand]),
				showWhen: player => player.getTalents().sealOfCommand,
			},
			{ value: PaladinSealTwist.SealTwistFury, actionId: spellIcon(TWIST_SPELLS[PaladinSealTwist.SealTwistFury]) },
		],
	}),
];

export const RotationInputs = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecRetributionPaladin, number>({
			fieldName: 'uwSocStacks',
			label: i18n.t('rotation_tab.options.paladin.retribution.uw_soc_stacks.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.uw_soc_stacks.tooltip'),
			values: ECHO_STACK_VALUES,
			showWhen: player => player.getTalents().unsealedWrath && playerTwist(player) === PaladinSealTwist.SealTwistCommand,
			getValue: player => echoStacks(player.getSimpleRotation().uwSocStacks, TWIST_STACKS[PaladinSealTwist.SealTwistCommand].own),
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecRetributionPaladin, number>({
			fieldName: 'uwSofStacks',
			label: i18n.t('rotation_tab.options.paladin.retribution.uw_sof_stacks.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.uw_sof_stacks.tooltip'),
			values: ECHO_STACK_VALUES,
			showWhen: player => player.getTalents().unsealedWrath && playerTwist(player) === PaladinSealTwist.SealTwistFury,
			getValue: player => echoStacks(player.getSimpleRotation().uwSofStacks, TWIST_STACKS[PaladinSealTwist.SealTwistFury].own),
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecRetributionPaladin, number>({
			fieldName: 'uwSorStacks',
			label: i18n.t('rotation_tab.options.paladin.retribution.uw_sor_stacks.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.uw_sor_stacks.tooltip'),
			values: ECHO_STACK_VALUES,
			showWhen: player => player.getTalents().unsealedWrath && playerTwist(player) !== PaladinSealTwist.SealTwistUnknown,
			getValue: player => righteousnessStacks(player.getSimpleRotation().uwSorStacks, playerTwist(player)),
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecRetributionPaladin>({
			fieldName: 'pullRange',
			label: i18n.t('rotation_tab.options.paladin.retribution.pull_range.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.pull_range.tooltip', { min: MIN_PULL_RANGE }),
			positive: true,
			getValue: player => player.getSimpleRotation().pullRange,
			setValue: (player: Player<Spec.SpecRetributionPaladin>, newValue: number) => {
				const rotation = player.getSimpleRotation();
				rotation.pullRange = Math.max(MIN_PULL_RANGE, newValue);
				player.setSimpleRotation(rotation);
			},
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecRetributionPaladin>({
			fieldName: 'delayMajorCds',
			label: i18n.t('rotation_tab.options.paladin.retribution.delay_major_cds.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.delay_major_cds.tooltip'),
			positive: true,
			getValue: player => player.getSimpleRotation().delayMajorCds,
		}),
	],
};
