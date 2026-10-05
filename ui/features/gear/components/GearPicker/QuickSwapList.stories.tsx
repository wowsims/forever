import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { expect, fn, userEvent, within } from 'storybook/test';

import { QuickSwapList } from './QuickSwapList';

const meta = {
	title: 'Features/Gear/GearPicker/Components/QuickSwapList',
	component: QuickSwapList,
	tags: ['autodocs'],
	args: {
		title: 'Favorite Enchants',
		emptyMessage: 'Favorite enchants to see them here.',
		entries: [],
		onItemClick: fn(),
		footerButton: { label: 'Open Enchants', onClick: fn() },
	},
	argTypes: { entries: { control: false } },
	decorators: [
		Story => (
			<div className="w-54 bg-gray-900">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof QuickSwapList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Open Enchants' }));
		await expect(args.footerButton.onClick).toHaveBeenCalledOnce();
	},
};

export const FavouriteEnchants: Story = {
	args: { entries: [{ item: makeUIEnchant(), active: true }] },
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};
