import type { Meta, StoryObj } from '@storybook/react-vite';
import { designTargetNames, emptyDesign, overlappingDowntimesDesign } from '@testing/storybook/fixtures/encounter_design';
import { fn } from 'storybook/test';

import { DESIGN_ISSUE_KEYS, validateDesign } from '../../../model/design/validate';
import { DesignerIssues } from './DesignerIssues';

const overlapping = overlappingDowntimesDesign();

const meta = {
	title: 'Features/Encounter/EncounterDesigner/DesignerIssues',
	component: DesignerIssues,
	tags: ['autodocs'],
	args: { design: overlapping, issues: validateDesign(overlapping, designTargetNames.length), onSelect: fn() },
	argTypes: { design: { control: false }, issues: { control: false } },
} satisfies Meta<typeof DesignerIssues>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithIssues: Story = {};

export const OneTargetEveryLevel: Story = {
	args: {
		issues: [
			{ ref: { kind: 'downtime', index: 1 }, severity: 'error', key: DESIGN_ISSUE_KEYS.overlap },
			{ ref: { kind: 'downtime', index: 0 }, severity: 'warning', key: DESIGN_ISSUE_KEYS.outsideFight },
			{ ref: { kind: 'downtime', index: 2 }, severity: 'info', key: DESIGN_ISSUE_KEYS.allTargetsDown },
		],
	},
};

export const NoIssues: Story = { args: { design: emptyDesign(), issues: [] } };
