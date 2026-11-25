import type { Meta, StoryObj } from '@storybook/react-vite';
import { emptyPlan, overlappingDowntimesPlan, PLAN_BASE_SECONDS, planTargetNames } from '@testing/storybook/fixtures/encounter_plan';
import { fn } from 'storybook/test';

import { PLAN_ISSUE_KEYS, validatePlan } from '../../model/validate';
import { PlannerIssues } from './PlannerIssues';

const overlapping = overlappingDowntimesPlan();

const meta = {
	title: 'Features/EncounterPlanner/PlannerIssues',
	component: PlannerIssues,
	tags: ['autodocs'],
	args: { plan: overlapping, issues: validatePlan(overlapping, planTargetNames.length), baseSeconds: PLAN_BASE_SECONDS, onSelect: fn() },
	argTypes: { plan: { control: false }, issues: { control: false } },
} satisfies Meta<typeof PlannerIssues>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithIssues: Story = {};

export const OneTargetEveryLevel: Story = {
	args: {
		issues: [
			{ ref: { kind: 'downtime', index: 1 }, severity: 'error', key: PLAN_ISSUE_KEYS.overlap },
			{ ref: { kind: 'downtime', index: 0 }, severity: 'warning', key: PLAN_ISSUE_KEYS.outsideFight },
			{ ref: { kind: 'downtime', index: 2 }, severity: 'info', key: PLAN_ISSUE_KEYS.allTargetsDown, span: { start: 0.2, end: 0.25 } },
		],
	},
};

export const NoIssues: Story = { args: { plan: emptyPlan(), issues: [] } };
