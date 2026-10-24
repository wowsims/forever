import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import { InnervateAssignment } from '@generated/proto/druid';
import i18n from '@i18n/config';

const assignment = (key: string) => i18n.t(`settings_tab.other.innervate_assignment.${key}`);

export const InnervateAssignmentInput = InputHelpers.makeSpecOptionsEnumInput<Spec.SpecBalanceDruid, InnervateAssignment>({
	fieldName: 'innervateAssignment',
	label: i18n.t('settings_tab.other.innervate_assignment.label'),
	labelTooltip: i18n.t('settings_tab.other.innervate_assignment.tooltip'),
	values: [
		{ name: assignment('self'), value: InnervateAssignment.InnervateAssignmentSelf },
		{ name: assignment('other'), value: InnervateAssignment.InnervateAssignmentOther },
	],
});
