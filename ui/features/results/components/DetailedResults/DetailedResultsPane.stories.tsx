import { Tabs } from '@base-ui/react/tabs';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { loadSimResult, type ResultFixture, resultsHostParts } from '@testing/storybook/fixtures/results-views';

import { DetailedResultsPane } from './DetailedResultsPane';

const meta = {
	title: 'Features/Results/DetailedResults/Components/DetailedResultsPane',
	component: DetailedResultsPane,
	tags: ['autodocs'],
	args: {
		id: 'damageTab',
		children: <p className="p-4">The pane's table.</p>,
	},
	argTypes: { children: { control: false } },
	loaders: [loadSimResult],
	decorators: [
		(Story, context) => withSimHost(resultsHostParts((context.loaded as ResultFixture).result))(Story, context),
		(Story, { parameters }) => (
			<Tabs.Root value="damageTab" className="group/dr flex flex-col" data-no-results={parameters.noResults ? '' : undefined}>
				<Story />
			</Tabs.Root>
		),
	],
} satisfies Meta<typeof DetailedResultsPane>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ContentOnly: Story = {};

export const WithTopline: Story = { args: { topline: true } };

export const WithHistogram: Story = {
	args: { topline: true, histogram: true },
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
};

export const NoResults: Story = { args: { topline: true }, parameters: { noResults: true } };
