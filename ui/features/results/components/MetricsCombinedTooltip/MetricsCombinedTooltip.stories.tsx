import { SpellSchool } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withTooltipSurface } from '@testing/storybook/decorators';

import { MetricsCombinedTooltip } from './MetricsCombinedTooltip';

const meta = {
	title: 'Features/Results/MetricsCombinedTooltip',
	component: MetricsCombinedTooltip,
	tags: ['autodocs'],
	args: {
		groups: [
			{
				name: 'Hits',
				totalPercentage: 100,
				spellSchool: SpellSchool.SpellSchoolPhysical,
				data: [
					{ name: 'Hit', value: 3800, percentage: 54 },
					{ name: 'Critical Hit', value: 1700, percentage: 24 },
					{ name: 'Glancing Blow', value: 500, percentage: 7 },
				],
			},
			{
				name: 'Ticks',
				totalPercentage: 100,
				spellSchool: SpellSchool.SpellSchoolPhysical,
				data: [{ name: 'Tick', value: 1040, percentage: 15 }],
			},
		],
	},
	argTypes: { groups: { control: false } },
	decorators: [
		Story => (
			<div className="min-w-96 text-xs">
				<Story />
			</div>
		),
		withTooltipSurface,
		Story => (
			<div className="w-max">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof MetricsCombinedTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Grouped: Story = {};

export const WithAverages: Story = {
	args: {
		groups: [
			{
				totalPercentage: 100,
				spellSchool: SpellSchool.SpellSchoolFire,
				data: [
					{ name: 'Hit', value: 64, percentage: 80, average: 1210.4 },
					{ name: 'Critical Hit', value: 16, percentage: 20, average: 2420.8 },
				],
			},
		],
	},
};

export const WithRanges: Story = {
	args: {
		headerValues: [undefined, 'Amount'],
		groups: [
			{
				totalPercentage: 100,
				spellSchool: SpellSchool.SpellSchoolHoly,
				data: [
					{ name: 'Hit', value: 3800, percentage: 70, average: 773, min: 650, max: 900 },
					{ name: 'Critical Hit', value: 1700, percentage: 30, average: 1546, min: 1300, max: 1800 },
				],
			},
			{
				name: 'Spec metrics',
				totalPercentage: 100,
				spellSchool: SpellSchool.SpellSchoolHoly,
				data: [{ name: 'Divine Aegis shield', value: 4920, percentage: 89, average: 410, min: 300, max: 520 }],
			},
		],
	},
};

export const CustomHeaders: Story = { args: { headerValues: ['Outcome', 'Casts'] } };
