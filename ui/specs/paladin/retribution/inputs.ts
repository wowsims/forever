import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import { PaladinAura, PaladinJudgement, PaladinSeal } from '@generated/proto/paladin';
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
];

export const RotationInputs = {
	inputs: [
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
