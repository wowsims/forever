import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEquippedItem, wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SelectorModalTabs } from '../../types';
import { ItemDetailCellView } from './ItemDetailCell';

const meta = {
	title: 'Features/Gear/ItemCell/ItemDetailCell',
	component: ItemDetailCellView,
	tags: ['autodocs'],
	args: {
		slot: ItemSlot.ItemSlotMainHand,
		item: makeEquippedItem(),
		iconUrl: wowheadIconUrl('inv_sword_39'),
		href: '',
		onOpen: fn(),
	},
	argTypes: { item: { control: false }, notice: { control: false }, extraLabels: { control: false }, action: { control: false } },
	decorators: [
		Story => (
			<div className="ui-gear-picker w-120">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ItemDetailCellView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Equipped: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('item-picker-name-container'));
		await expect(args.onOpen).toHaveBeenCalledWith(SelectorModalTabs.Items);
	},
};

export const EmptySlot: Story = { args: { item: null, iconUrl: '' } };

export const RightColumn: Story = {
	args: { align: 'end', slot: ItemSlot.ItemSlotHands, item: null, iconUrl: '' },
	decorators: [
		Story => (
			<div className="*:flex-row-reverse">
				<Story />
			</div>
		),
	],
};

export const Enchanted: Story = {
	args: { item: makeEquippedItem({ enchant: makeUIEnchant() }) },
	parameters: { smoke: false, smokeReason: 'The enchant label fetches /forever/assets/enchants/descriptions.json' },
};
