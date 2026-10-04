import type { Meta, StoryObj } from '@storybook/react-vite';
import { loadTopline, type ToplineFixture } from '@testing/storybook/fixtures/results-views';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SimResultSummaryView } from './SimResultSummary';

const meta = {
	title: 'Features/Results/SimResultsPanel/Components/SimResultSummary',
	component: SimResultSummaryView,
	tags: ['autodocs'],
	args: {
		metrics: [],
		onSetReference: fn(),
		onSwapReference: fn(),
		onClearReference: fn(),
		onOpenBreakdown: fn(),
		freshness: 'fresh',
	},
	loaders: [loadTopline],
	render: (args, { loaded }) => <SimResultSummaryView {...args} metrics={(loaded as ToplineFixture).metrics} />,
	decorators: [
		Story => (
			<div className="w-72">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SimResultSummaryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NoReference: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('results-sim-set-reference'));
		await expect(args.onSetReference).toHaveBeenCalledOnce();
	},
};

export const WithReference: Story = {
	render: (args, { loaded }) => {
		const { metrics, referenceDiffs } = loaded as ToplineFixture;
		return <SimResultSummaryView {...args} metrics={metrics} referenceDiffs={referenceDiffs} />;
	},
};
