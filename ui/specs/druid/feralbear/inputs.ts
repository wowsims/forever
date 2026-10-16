import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';

export const StartingRage = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecFeralBearDruid>({
	fieldName: 'startingRage',
	label: i18n.t('settings_tab.other.starting_rage.label'),
	labelTooltip: i18n.t('settings_tab.other.starting_rage.tooltip'),
});

export const AssumeBleedActive = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralBearDruid>({
	fieldName: 'assumeBleedActive',
	label: i18n.t('settings_tab.other.assume_bleed_active.label'),
	labelTooltip: i18n.t('settings_tab.other.assume_bleed_active.tooltip'),
});
