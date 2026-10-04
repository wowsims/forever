import type { Meta, StoryObj } from '@storybook/react-vite';

import { StatWeightValue } from './StatWeightValue';

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/StatWeightValue',
	component: StatWeightValue,
	tags: ['autodocs'],
	args: { value: 1.92, stdev: 0.4, iterations: 1000 },
	argTypes: { sign: { control: 'select', options: [undefined, 'positive', 'negative'] } },
} satisfies Meta<typeof StatWeightValue>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Positive: Story = { args: { sign: 'positive', className: 'text-success' } };
export const Negative: Story = { args: { sign: 'negative', className: 'text-danger' } };
