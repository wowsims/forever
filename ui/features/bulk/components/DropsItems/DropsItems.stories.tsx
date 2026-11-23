import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsItemListFixture, dropsNamesFixture } from '@testing/storybook/fixtures/bulk-drops';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DropsItemsView } from './DropsItems';

const meta = {
	title: 'Features/Bulk/DropsItems',
	component: DropsItemsView,
	tags: ['autodocs'],
	args: {
		list: dropsItemListFixture(),
		grouping: 'slot',
		names: dropsNamesFixture(),
		onGroupingChange: fn(),
		onIncludedChange: fn(),
	},
	argTypes: { list: { control: false }, names: { control: false } },
	decorators: [
		Story => (
			<div className="w-170">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsItemsView>;

export default meta;
type Story = StoryObj<typeof meta>;

// Worn rows grey out, and a row's unticked box stays unticked.
export const Default: Story = {};

export const BySource: Story = { args: { grouping: 'source' } };

export const Empty: Story = { args: { list: { items: [], leftOut: [] } } };

export const LeftOutShown: Story = {
	play: async ({ canvasElement }) => {
		const footer = within(canvasElement).getByTestId('drops-left-out');
		await userEvent.click(within(footer).getByRole('button', { name: i18n.t('bulk_tab.drops.setup.show') }));
		await expect(within(footer).getByText('Robe of the Archmage')).toBeTruthy();
	},
};
