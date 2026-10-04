import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemSwapPickerView } from './ItemSwapPicker';

const SLOTS = [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotRanged];

const EmptySlot = ({ slot }: { slot: ItemSlot }) => (
	<span className="ui-icon-picker-swatch block size-16 grayscale" style={{ backgroundImage: `url('${getEmptySlotIconUrl(slot)}')` }} />
);

const meta = {
	title: 'Features/ItemSwap/ItemSwapPicker',
	component: ItemSwapPickerView,
	tags: ['autodocs'],
	args: {
		onSwap: fn(),
		children: SLOTS.map(slot => <EmptySlot key={slot} slot={slot} />),
	},
	argTypes: { children: { control: false } },
	decorators: [
		Story => (
			<div className="grid w-96 gap-2">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ItemSwapPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('gear-swap-icon'));
		await expect(args.onSwap).toHaveBeenCalledOnce();
	},
};

export const WithNote: Story = { args: { note: 'Only the main hand swaps back after the opener.' } };

export const OneSlot: Story = { args: { children: <EmptySlot slot={ItemSlot.ItemSlotTrinket1} /> } };
