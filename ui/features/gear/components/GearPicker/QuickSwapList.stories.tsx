import { GemColor } from '@generated/proto/common';
import type { UIEnchant, UIGem } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIEnchant, makeUIGem } from '@testing/storybook/fixtures/protos';
import { expect, fn, userEvent, within } from 'storybook/test';

import { QuickSwapList, type QuickSwapListProps } from './QuickSwapList';

type Props = QuickSwapListProps<UIGem | UIEnchant>;

const meta = {
	title: 'Features/Gear/GearPicker/Components/QuickSwapList',
	component: QuickSwapList as (props: Props) => ReturnType<typeof QuickSwapList>,
	tags: ['autodocs'],
	args: {
		title: 'Favorite Gems',
		emptyMessage: 'Favorite gems to see them here.',
		entries: [],
		onItemClick: fn(),
		footerButton: { label: 'Open Gems', onClick: fn() },
	},
	argTypes: { entries: { control: false } },
	decorators: [
		Story => (
			<div className="w-54 bg-gray-900">
				<Story />
			</div>
		),
	],
} satisfies Meta<(props: Props) => ReturnType<typeof QuickSwapList>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Open Gems' }));
		await expect(args.footerButton.onClick).toHaveBeenCalledOnce();
	},
};

export const FavouriteGems: Story = {
	args: {
		entries: [
			{ item: makeUIGem(), active: true },
			{ item: makeUIGem({ id: 2, name: 'Citrine', icon: 'inv_misc_gem_topaz_01', color: GemColor.GemColorYellow }), active: false },
		],
	},
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};

export const FavouriteEnchants: Story = {
	args: {
		title: 'Favorite Enchants',
		emptyMessage: 'Favorite enchants to see them here.',
		entries: [{ item: makeUIEnchant(), active: true }],
		footerButton: { label: 'Open Enchants', onClick: fn() },
	},
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};
