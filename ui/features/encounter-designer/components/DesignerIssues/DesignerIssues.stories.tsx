import type { Meta, StoryObj } from '@storybook/react-vite';
import { DESIGN_BASE_SECONDS, designTargetNames, emptyDesign, overlappingDowntimesDesign } from '@testing/storybook/fixtures/encounter_design';
import { fn } from 'storybook/test';

import { DESIGN_ISSUE_KEYS, validateDesign } from '../../model/validate';
import { DesignerIssues } from './DesignerIssues';

const overlapping = overlappingDowntimesDesign();

const meta = {
	title: 'Features/EncounterDesigner/DesignerIssues',
	component: DesignerIssues,
	tags: ['autodocs'],
	args: { design: overlapping, issues: validateDesign(overlapping, designTargetNames.length), baseSeconds: DESIGN_BASE_SECONDS, onSelect: fn() },
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
			{ ref: { kind: 'downtime', index: 2 }, severity: 'info', key: DESIGN_ISSUE_KEYS.allTargetsDown, span: { start: 0.2, end: 0.25 } },
		],
	},
};

export const NoIssues: Story = { args: { design: emptyDesign(), issues: [] } };
