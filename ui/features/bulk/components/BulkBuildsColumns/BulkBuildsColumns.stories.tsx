import type { Meta, StoryObj } from '@storybook/react-vite';
import { buildBundlesProps, bundleListProps, encounterListProps } from '@testing/storybook/fixtures/bulk-builds';

import { BuildBundlesView } from '../BuildBundles';
import { EncounterListView } from '../EncounterList';
import { BulkBuildsColumnsView } from './BulkBuildsColumns';

const meta = {
	title: 'Features/Bulk/BulkBuildsColumns',
	component: BulkBuildsColumnsView,
	tags: ['autodocs'],
	args: {
		builds: <BuildBundlesView {...buildBundlesProps()} />,
		encounters: <EncounterListView {...encounterListProps()} />,
	},
	argTypes: {
		builds: { control: false },
		encounters: { control: false },
	},
} satisfies Meta<typeof BulkBuildsColumnsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

// Two columns beside the sidebar and the run card, as at 1440 wide.
export const Narrow: Story = {
	decorators: [
		Story => (
			<div className="w-3xl">
				<Story />
			</div>
		),
	],
};

export const Phone: Story = {
	decorators: [
		Story => (
			<div className="w-96">
				<Story />
			</div>
		),
	],
};

export const EmptyLists: Story = {
	args: {
		builds: <BuildBundlesView {...buildBundlesProps({ list: bundleListProps({ rows: [] }), enchantPicks: [] })} />,
		encounters: <EncounterListView {...encounterListProps({ presets: [], saved: [] })} />,
	},
};
