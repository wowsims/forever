import { Faction, ItemQuality } from '@generated/proto/common';
import { UIItemSource } from '@generated/proto/ui';
import type { Sim } from '@sim/sim';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { makeUIItem } from '@testing/storybook/fixtures/protos';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemNoticeIconView } from '../GearPicker/ItemNoticeIcon';
import { ItemListRowView } from './ItemListRow';
import { ItemSource } from './ItemSource';

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

export const NewlyDiscovered: Story = {
	args: {
		notice: (
			<ItemNoticeIconView level="info">
				<p className="mb-0">New: added to the sim on 11/3/2026</p>
			</ItemNoticeIconView>
		),
	},
};

export const EnchantsTab: Story = {
	args: {
		name: 'Enchant Weapon - Crusader',
		quality: ItemQuality.ItemQualityUncommon,
		iconUrl: wowheadIconUrl('spell_holy_greaterheal'),
		ilvl: undefined,
		source: (
			<small>
				Western Plaguelands
				<br />
				Scarlet Spellbinder
			</small>
		),
		compare: undefined,
		ep: { ep: 9.2, equippedEP: 12 },
	},
};

export const EnchantWithoutSource: Story = {
	args: {
		name: 'Enchant Chest - Inferior Stamina',
		quality: ItemQuality.ItemQualityCommon,
		iconUrl: wowheadIconUrl('inv_misc_enchantedscroll'),
		ilvl: 0,
		source: null,
		compare: undefined,
		ep: { ep: 1.2, equippedEP: 12 },
	},
};

export const OtherSources: Story = {
	args: {
		source: (
			<ItemSource
				item={makeUIItem({
					id: 19019,
					sources: [
						UIItemSource.create({
							source: { oneofKind: 'soldBy', soldBy: { npcId: 12944, npcName: 'Lokhtos Darkbargainer', zoneId: 1584, faction: Faction.Unknown } },
						}),
						UIItemSource.create({ source: { oneofKind: 'quest', quest: { id: 7, name: 'Attunement to the Core', faction: Faction.Unknown } } }),
						UIItemSource.create({
							source: { oneofKind: 'drop', drop: { difficulty: 0, npcId: 0, zoneId: 0, otherName: 'World Drop', category: '' } },
						}),
					],
				})}
				sim={{ db: { getZone: () => null, getNpc: () => null } } as unknown as Sim}
			/>
		),
	},
	play: async ({ args, canvasElement }) => {
		const more = within(canvasElement).getByTestId('item-source-more');
		await userEvent.hover(more);
		await within(canvasElement.ownerDocument.body).findByRole('dialog');
		await userEvent.click(more);
		await expect(args.onEquip).not.toHaveBeenCalled();
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
