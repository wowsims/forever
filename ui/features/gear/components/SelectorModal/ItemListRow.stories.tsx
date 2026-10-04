import { ItemQuality } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemListRowView } from './ItemListRow';

const meta = {
	title: 'Features/Gear/SelectorModal/Components/ItemListRow',
	component: ItemListRowView,
	tags: ['autodocs'],
	args: {
		name: 'Thunderfury, Blessed Blade of the Windseeker',
		quality: ItemQuality.ItemQualityLegendary,
		nameDescription: '',
		iconUrl: wowheadIconUrl('inv_sword_39'),
		href: 'https://www.wowhead.com/classic/item=19019',
		ilvl: 80,
		active: false,
		onEquip: fn(),
		source: <small>Molten Core</small>,
		ep: { ep: 42.4, equippedEP: 38 },
		favourited: false,
		favouriteTooltipId: 'favourite-tooltip',
		onToggleFavourite: fn(),
		compare: { inBatch: false, tooltipId: 'compare-tooltip', onToggle: fn() },
	},
	argTypes: { notice: { control: false }, source: { control: false } },
	decorators: [
		Story => (
			<div className="relative flex w-4xl items-center gap-4 bg-gray-900 p-2">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ItemListRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ItemsTab: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('selector-modal-list-item-link'));
		await expect(args.onEquip).toHaveBeenCalledOnce();
	},
};

export const EquippedFavourite: Story = {
	args: { active: true, favourited: true, ep: { ep: 38, equippedEP: 38 }, compare: { inBatch: true, tooltipId: 'compare-tooltip', onToggle: fn() } },
};

export const EnchantsTab: Story = {
	args: {
		name: 'Enchant Weapon - Crusader',
		quality: ItemQuality.ItemQualityUncommon,
		iconUrl: wowheadIconUrl('spell_holy_greaterheal'),
		ilvl: undefined,
		source: undefined,
		compare: undefined,
		ep: { ep: 9.2, equippedEP: 12 },
	},
};

export const TrinketWithoutEp: Story = {
	args: {
		name: 'Hand of Justice',
		quality: ItemQuality.ItemQualityRare,
		iconUrl: wowheadIconUrl('inv_jewelry_talisman_01'),
		ilvl: 58,
		nameDescription: 'Heroic',
		ep: undefined,
	},
};
