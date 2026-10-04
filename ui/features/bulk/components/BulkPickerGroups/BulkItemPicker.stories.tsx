import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { BulkItemPickerView } from './BulkItemPicker';

const meta = {
	title: 'Features/Bulk/BulkPickerGroups/BulkItemPicker',
	component: BulkItemPickerView,
	tags: ['autodocs'],
	decorators: [withSimHost({ player: { getSpec: () => 0 } })],
	args: { slot: ItemSlot.ItemSlotMainHand, item: makeEquippedItem(), state: 'editable', onOpen: fn(), onRemove: fn() },
	argTypes: { item: { control: false }, state: { control: 'select', options: ['frozen', 'equipped', 'editable'] } },
} satisfies Meta<typeof BulkItemPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Editable: Story = {};

export const Equipped: Story = { args: { state: 'equipped', onRemove: undefined } };

export const Frozen: Story = { args: { state: 'frozen', onRemove: undefined } };
