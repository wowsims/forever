// Configuration for spec-specific UI elements on the settings tab.
import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import { DpsWarriorSpec, WarriorSunder } from '@generated/proto/warrior';
import i18n from '@i18n/config';

import { isArmsKebabSpec, isArmsSpec, isFurySpec } from './presets';

// These don't need to be in a separate file but it keeps things cleaner.
export const RotationInputs = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecDpsWarrior, DpsWarriorSpec>({
			fieldName: 'spec',
			label: i18n.t('rotation_tab.options.warrior.dps.spec.label'),
			values: [
				{ name: i18n.t('rotation_tab.options.warrior.dps.spec.fury'), value: DpsWarriorSpec.DpsWarriorSpecFury },
				{ name: i18n.t('rotation_tab.options.warrior.dps.spec.arms'), value: DpsWarriorSpec.DpsWarriorSpecArms },
			],
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecDpsWarrior, WarriorSunder>({
			fieldName: 'sunderArmor',
			label: i18n.t('rotation_tab.options.warrior.sunder_armor.label'),
			values: [
				{ name: i18n.t('rotation_tab.options.warrior.sunder_armor.none'), value: WarriorSunder.WarriorSunderNone },
				{ name: i18n.t('rotation_tab.options.warrior.sunder_armor.help'), value: WarriorSunder.WarriorSunderHelp },
				{ name: i18n.t('rotation_tab.options.warrior.sunder_armor.maintain'), value: WarriorSunder.WarriorSunderMaintain },
			],
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useRecklessness',
			label: i18n.t('rotation_tab.options.warrior.dps.use_recklessness.label'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useStanceDance',
			label: i18n.t('rotation_tab.options.warrior.dps.use_stance_dance.label'),
			showWhen: player => {
				const { spec } = player.getSimpleRotation();
				return spec ? spec === DpsWarriorSpec.DpsWarriorSpecArms : isArmsSpec(player) || isArmsKebabSpec(player);
			},
			storeField: ['rotation', 'talentsString', 'gear'] as const,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useOverpower',
			label: i18n.t('rotation_tab.options.warrior.dps.use_overpower.label'),
			showWhen: player => isFurySpec(player) || isArmsKebabSpec(player),
			storeField: ['rotation', 'talentsString', 'gear'] as const,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useRend',
			label: i18n.t('rotation_tab.options.warrior.dps.use_rend.label'),
			showWhen: player => isFurySpec(player),
			storeField: ['rotation', 'talentsString', 'gear'] as const,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useHamstring',
			label: i18n.t('rotation_tab.options.warrior.dps.use_hamstring.label'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useHeroicStrikeRank9',
			label: i18n.t('rotation_tab.options.warrior.dps.use_heroic_strike_rank_9.label'),
			labelTooltip: i18n.t('rotation_tab.options.warrior.dps.use_heroic_strike_rank_9.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecDpsWarrior>({
			fieldName: 'useBattleShoutRank7',
			label: i18n.t('rotation_tab.options.warrior.dps.use_battle_shout_rank_7.label'),
			labelTooltip: i18n.t('rotation_tab.options.warrior.dps.use_battle_shout_rank_7.tooltip'),
		}),
	],
};
