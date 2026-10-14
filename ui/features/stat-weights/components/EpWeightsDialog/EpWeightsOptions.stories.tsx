import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/EpWeightsOptions',
	component: EpWeightsOptions,
	tags: ['autodocs'],
	args: {
		statsType: StatsType.Ep,
		onStatsTypeChange: fn(),
	},
	argTypes: { statsType: { control: 'select', options: [StatsType.Ep, StatsType.Weight] } },
} satisfies Meta<typeof EpWeightsOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Weights: Story = { args: { statsType: StatsType.Weight } };
