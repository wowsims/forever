import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsItemListFixture, dropsNamesFixture, dropsTreeFixture } from '@testing/storybook/fixtures/bulk-drops';
import { fn } from 'storybook/test';

import { DropsItemsView } from './DropsItems';

const meta = {
	title: 'Features/Bulk/DropsItems',
	component: DropsItemsView,
	tags: ['autodocs'],
	args: {
		list: dropsItemListFixture(),
		grouping: 'slot',
		tree: dropsTreeFixture(),
		names: dropsNamesFixture(),
		onGroupingChange: fn(),
		onIncludedChange: fn(),
	},
	argTypes: { list: { control: false }, tree: { control: false }, names: { control: false } },
	decorators: [
		Story => (
			<div className="w-full max-w-170">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsItemsView>;

export default meta;
type Story = StoryObj<typeof meta>;

// Worn rows grey out, and a row's unticked box stays unticked.
export const Default: Story = {};

// A row names only its other sources, never the boss it sits under.
export const BySource: Story = { args: { grouping: 'source' } };

// No source ticked: the grouping switch and the bulk actions wait for a first item.
export const Empty: Story = { args: { list: { items: [] } } };
