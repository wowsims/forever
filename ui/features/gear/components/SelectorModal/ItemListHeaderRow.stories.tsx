import { SortDirection } from '@sim/constants/other';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemListSortBy } from '../../model/item_sort';
import { SelectorModalTabs } from '../../types';
import { ItemListHeaderRow } from './ItemListHeaderRow';

const meta = {
	title: 'Features/Gear/SelectorModal/Components/ItemListHeaderRow',
	component: ItemListHeaderRow,
	tags: ['autodocs'],
	args: {
		label: SelectorModalTabs.Items,
		ep: { showValues: true, tooltipId: 'ep-tooltip' },
		sortBy: ItemListSortBy.EP,
		sortDirection: SortDirection.DESC,
		onSort: fn(),
	},
	decorators: [
		Story => (
			<div className="w-4xl">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ItemListHeaderRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ItemsTab: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(within(canvasElement).getByTestId('ilvl-label')).getByRole('button'));
		await expect(args.onSort).toHaveBeenCalledWith(ItemListSortBy.ILVL);
	},
};

export const EnchantsTab: Story = { args: { label: SelectorModalTabs.Enchants, ep: undefined, sortBy: ItemListSortBy.ILVL } };

export const EpHidden: Story = { args: { ep: { showValues: false, tooltipId: 'ep-tooltip' } } };
