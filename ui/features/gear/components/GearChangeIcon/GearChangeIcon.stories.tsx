import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';

import { GearChangeIconView } from './GearChangeIcon';

const meta = {
	title: 'Features/Gear/GearChangeIcon',
	component: GearChangeIconView,
	tags: ['autodocs'],
	args: {
		slot: ItemSlot.ItemSlotHead,
		equipped: true,
		iconUrl: wowheadIconUrl('inv_helmet_36'),
		href: '',
	},
} satisfies Meta<typeof GearChangeIconView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Unchanged: Story = {};

export const EmptySlot: Story = { args: { equipped: false, iconUrl: '' } };

export const EnchantUnchanged: Story = {
	args: { enchant: { name: 'Enchant Chest - Greater Stats', iconUrl: wowheadIconUrl('inv_misc_note_01') } },
};

export const EnchantChanged: Story = {
	args: { enchant: { name: 'Enchant Chest - Greater Stats', iconUrl: wowheadIconUrl('inv_misc_note_01') }, enchantChanged: true },
};

export const EnchantChangedWithoutIcon: Story = { args: { enchant: { name: 'Enchant Chest - Greater Stats', iconUrl: '' }, enchantChanged: true } };
