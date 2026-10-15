import type { Meta, StoryObj } from '@storybook/react-vite';

import { ResultMetricValue } from './ResultMetricValue';

const meta = {
	title: 'Features/Results/ResultMetricValue',
	component: ResultMetricValue,
	tags: ['autodocs'],
	args: { average: '788.23', unit: 'DPS', margin: '3' },
	decorators: [
		Story => (
			<div className="font-bold">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ResultMetricValue>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithMargin: Story = {};

export const NoMargin: Story = { args: { margin: undefined } };

export const Small: Story = { args: { size: 'sm', average: '12.40', unit: 'TPS' } };
