import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const CannotShredTarget = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'cannotShredTarget',
	label: i18n.t('settings_tab.other.cannot_shred_target.label'),
	labelTooltip: i18n.t('settings_tab.other.cannot_shred_target.tooltip'),
});

export const PrepullClearcasting = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'prepullClearcasting',
	label: i18n.t('settings_tab.other.prepull_clearcasting.label'),
	labelTooltip: i18n.t('settings_tab.other.prepull_clearcasting.tooltip'),
});

export const AssumeBleedActive = InputHelpers.makeSpecOptionsBooleanInput<Spec.SpecFeralCatDruid>({
	fieldName: 'assumeBleedActive',
	label: i18n.t('settings_tab.other.assume_bleed_active.label'),
	labelTooltip: i18n.t('settings_tab.other.assume_bleed_active.tooltip'),
});
