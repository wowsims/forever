import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { withSimHost } from '@testing/storybook/fixtures/models';

import { ReforgeDoneToast } from './ReforgeDoneToast';

const ITEM_SLOTS = [ItemSlot.ItemSlotHead, ItemSlot.ItemSlotNeck, ItemSlot.ItemSlotShoulder, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeDoneToast',
	component: ReforgeDoneToast,
	tags: ['autodocs'],
	decorators: [withSimHost()],
	args: { itemSlots: ITEM_SLOTS, changedSlots: new Map(), previousGear: null },
	argTypes: { itemSlots: { control: false }, changedSlots: { control: false }, previousGear: { control: false } },
} satisfies Meta<typeof ReforgeDoneToast>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Every slot renders its empty frame when the run moved nothing in it. */
export const NoSlotChanged: Story = {};

export const MainHandChanged: Story = {
	args: { changedSlots: new Map([[ItemSlot.ItemSlotMainHand, makeEquippedItem()]]) },
};
