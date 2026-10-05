import { ItemSlot } from '@generated/proto/common';
import { Gear } from '@sim/proto/gear';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { expect, fireEvent, fn, within } from 'storybook/test';

import { SlotRail } from './SlotRail';

const PARTIAL_GEAR = new Gear({
	[ItemSlot.ItemSlotMainHand]: makeEquippedItem(),
	[ItemSlot.ItemSlotHands]: makeEquippedItem({ item: { id: 16863, name: 'Gauntlets of Might', icon: 'inv_gauntlets_10' } }),
});

const meta = {
	title: 'Features/Gear/SelectorModal/Components/SlotRail',
	component: SlotRail,
	tags: ['autodocs'],
	args: { gear: PARTIAL_GEAR, currentSlot: ItemSlot.ItemSlotMainHand, onOpen: fn() },
	argTypes: { gear: { control: false } },
	decorators: [
		Story => (
			<div className="relative ml-24">
				<Story />
			</div>
		),
		withSimHost(),
	],
} satisfies Meta<typeof SlotRail>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		// A click without the hover: the rail's tooltip opening on hover would land after the story ends.
		fireEvent.click(within(canvasElement).getAllByTestId('item-picker-icon')[2]);
		await expect(args.onOpen).toHaveBeenCalledWith(ItemSlot.ItemSlotShoulder);
	},
};
