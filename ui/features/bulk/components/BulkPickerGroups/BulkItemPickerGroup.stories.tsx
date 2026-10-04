import { BulkSimItemSlot } from '@sim/bulk/utils';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { BulkItemPickerGroup } from './BulkItemPickerGroup';

const meta = {
	title: 'Features/Bulk/BulkPickerGroups/Components/BulkItemPickerGroup',
	component: BulkItemPickerGroup,
	tags: ['autodocs'],
	args: { bulkSlot: BulkSimItemSlot.ItemSlotFinger, entries: [], onAddItem: fn(), addTooltipId: 'bulk-add-items-tooltip' },
	argTypes: { entries: { control: false } },
} satisfies Meta<typeof BulkItemPickerGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

/** A slot group with no items in the batch; a filled group's entries are `BulkItemPicker`s, storied as `BulkItemPickerView`. */
export const Empty: Story = {};
