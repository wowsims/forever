import type { Meta, StoryObj } from '@storybook/react-vite';
import { loadTopline, type ToplineFixture } from '@testing/storybook/fixtures/results-views';
import { expect, within } from 'storybook/test';

import { SimResultReadoutView } from './SimResultReadout';

const meta = {
	title: 'Features/Results/SimResultsPanel/Components/SimResultReadout',
	component: SimResultReadoutView,
	tags: ['autodocs'],
	decorators: [
		Story => (
			<div className="flex w-48 bg-background p-2">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SimResultReadoutView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NoRun: Story = {};

export const Running: Story = {
	args: { run: { metric: 'dps', value: 1061.06, fraction: 0.52 } },
	play: async ({ canvasElement }) => {
		await expect(within(canvasElement).getByTestId('sim-result-readout-run')).toHaveTextContent('1061.06');
	},
};

export const Presims: Story = { args: { run: { metric: 'dps' } } };

export const Result: Story = {
	args: { freshness: 'fresh' },
	loaders: [loadTopline],
	render: (args, { loaded }) => <SimResultReadoutView {...args} metric={(loaded as ToplineFixture).metrics[0]} />,
};

export const Outdated: Story = {
	...Result,
	args: { freshness: 'stale' },
};
