import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEquippedItem, wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemSwapIconView } from './ItemSwapIcon';

const meta = {
	title: 'Features/ItemSwap/ItemSwapIcon',
	component: ItemSwapIconView,
	tags: ['autodocs'],
	args: {
		slot: ItemSlot.ItemSlotMainHand,
		item: null,
		iconUrl: '',
		name: '',
		href: '',
		onOpen: fn(),
	},
	argTypes: { item: { control: false }, wowhead: { control: false } },
} satisfies Meta<typeof ItemSwapIconView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EmptySlot: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('icon-picker-button'));
		await expect(args.onOpen).toHaveBeenCalledOnce();
	},
};

export const Filled: Story = {
	args: { item: makeEquippedItem(), iconUrl: wowheadIconUrl('inv_sword_39'), name: 'Thunderfury, Blessed Blade of the Windseeker' },
};
