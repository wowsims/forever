import { ItemSlot } from '@generated/proto/common';
import { Gear } from '@sim/proto/gear';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { OpenSelectorModalContext } from '../../hooks/useSelectorModal';
import { GearPicker } from './GearPicker';

const meta = {
	title: 'Features/Gear/GearPicker/GearPicker',
	component: GearPicker,
	tags: ['autodocs'],
	args: { ready: true },
	decorators: [
		Story => (
			<OpenSelectorModalContext value={fn()}>
				<Story />
			</OpenSelectorModalContext>
		),
	],
} satisfies Meta<typeof GearPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EmptySlots: Story = { decorators: [withSimHost({ player: { getSpec: () => 0 } })] };

export const Equipped: Story = {
	decorators: [
		withSimHost({
			player: { getSpec: () => 0 },
			playerSlice: {
				gear: new Gear({
					[ItemSlot.ItemSlotMainHand]: makeEquippedItem(),
					[ItemSlot.ItemSlotHands]: makeEquippedItem({ item: { id: 16863, name: 'Gauntlets of Might', icon: 'inv_gauntlets_10' } }),
				}),
			},
		}),
	],
};
