import type { Meta, StoryObj } from '@storybook/react-vite';
import { loadTopline, type ToplineFixture } from '@testing/storybook/fixtures/results-views';

import { ResultMetricList } from './ResultMetricList';

const meta = {
	title: 'Features/Results/ResultMetricList/ResultMetricList',
	component: ResultMetricList,
	tags: ['autodocs'],
	args: { metrics: [], layout: 'row' },
	argTypes: { metrics: { control: false }, referenceDiffs: { control: false } },
	loaders: [loadTopline],
	render: (args, { loaded }) => <ResultMetricList {...args} metrics={(loaded as ToplineFixture).metrics} />,
} satisfies Meta<typeof ResultMetricList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Row: Story = {};

export const List: Story = {
	args: { layout: 'list' },
	decorators: [
		Story => (
			<div className="w-72">
				<Story />
			</div>
		),
	],
};

export const ListAgainstAReference: Story = {
	...List,
	render: (args, { loaded }) => {
		const { metrics, referenceDiffs } = loaded as ToplineFixture;
		return <ResultMetricList {...args} metrics={metrics} referenceDiffs={referenceDiffs} />;
	},
};
