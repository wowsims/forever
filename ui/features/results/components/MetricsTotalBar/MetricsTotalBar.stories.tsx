import { SpellSchool } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { MetricsTotalBar } from './MetricsTotalBar';

const meta = {
	title: 'Features/Results/MetricsTotalBar',
	component: MetricsTotalBar,
	tags: ['autodocs'],
	args: { percentage: 62.5, max: 1_000_000, total: 5_855_400, value: 625_000, spellSchool: SpellSchool.SpellSchoolPhysical },
	decorators: [
		Story => (
			<div className="w-100">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof MetricsTotalBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Physical: Story = {};

export const Fire: Story = { args: { percentage: 30, value: 300_000, spellSchool: SpellSchool.SpellSchoolFire } };

export const WithShieldOverlay: Story = { args: { spellSchool: SpellSchool.SpellSchoolHoly, overlayValue: 200_000 } };

export const NoSchool: Story = { args: { spellSchool: null, percentage: 5, value: 50_000 } };
