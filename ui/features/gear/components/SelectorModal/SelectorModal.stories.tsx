import { ItemQuality } from '@generated/proto/common';
import { SortDirection } from '@sim/constants/other';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { fn } from 'storybook/test';

import { ItemListSortBy } from '../../model/item_sort';
import { SelectorModalTabs } from '../../types';
import { ItemListToolbarView } from './ItemList';
import { ItemListHeaderRow } from './ItemListHeaderRow';
import { ItemListRowView } from './ItemListRow';
import { SelectorModalView } from './SelectorModal';

const ROWS = [
	{ name: 'Thunderfury, Blessed Blade of the Windseeker', quality: ItemQuality.ItemQualityLegendary, icon: 'inv_sword_39', ilvl: 80, ep: 42.4 },
	{ name: 'Brutality Blade', quality: ItemQuality.ItemQualityEpic, icon: 'inv_sword_43', ilvl: 70, ep: 38 },
	{ name: 'Maladath, Runed Blade of the Black Flight', quality: ItemQuality.ItemQualityEpic, icon: 'inv_sword_47', ilvl: 75, ep: 36.1 },
];

const ItemsPane = () => (
	<>
		<ItemListToolbarView
			search=""
			onSearchChange={fn()}
			phase={1}
			onPhaseChange={fn()}
			toggles={[{ id: 'show-ep-values-selector', testId: 'selector-modal-show-ep-values', label: 'Show EP', value: true, onChange: fn() }]}
			remove={{ label: 'Unequip Item', onClick: fn() }}
		/>
		<ItemListHeaderRow
			label={SelectorModalTabs.Items}
			ep={{ showValues: true, tooltipId: 'ep-tooltip' }}
			sortBy={ItemListSortBy.EP}
			sortDirection={SortDirection.DESC}
			onSort={fn()}
		/>
		{ROWS.map((row, index) => (
			<div key={row.name} className="relative flex items-center gap-4 bg-gray-900 p-2">
				<ItemListRowView
					name={row.name}
					quality={row.quality}
					nameDescription=""
					iconUrl={wowheadIconUrl(row.icon)}
					href=""
					ilvl={row.ilvl}
					active={index === 1}
					onEquip={fn()}
					source={<small>Molten Core</small>}
					ep={{ ep: row.ep, equippedEP: 38 }}
					favourited={index === 0}
					favouriteTooltipId="favourite-tooltip"
					onToggleFavourite={fn()}
					compare={{ inBatch: false, tooltipId: 'compare-tooltip', onToggle: fn() }}
				/>
			</div>
		))}
	</>
);

const ENCHANTS = [
	{ name: 'Crusader', quality: ItemQuality.ItemQualityUncommon, icon: 'spell_holy_holysmite', lvl: 60 },
	{ name: 'Lifestealing', quality: ItemQuality.ItemQualityUncommon, icon: 'spell_shadow_lifedrain', lvl: 60 },
	{ name: 'Fiery Weapon', quality: ItemQuality.ItemQualityUncommon, icon: 'spell_fire_flameblades', lvl: 50 },
];

const EnchantsPane = () => (
	<>
		<ItemListToolbarView
			search=""
			onSearchChange={fn()}
			phase={1}
			onPhaseChange={fn()}
			toggles={[{ id: 'show-all-enchants-selector', testId: 'selector-modal-show-all-enchants', label: 'Show All', value: false, onChange: fn() }]}
			remove={{ label: 'Remove Enchant', onClick: fn() }}
		/>
		<ItemListHeaderRow label={SelectorModalTabs.Enchants} sortBy={ItemListSortBy.ILVL} sortDirection={SortDirection.DESC} onSort={fn()} />
		{ENCHANTS.map((row, index) => (
			<div key={row.name} className="relative flex items-center gap-4 bg-gray-900 p-2">
				<ItemListRowView
					name={row.name}
					quality={row.quality}
					nameDescription=""
					iconUrl={wowheadIconUrl(row.icon)}
					href=""
					ilvl={row.lvl}
					active={index === 0}
					onEquip={fn()}
					favourited={index === 1}
					favouriteTooltipId="favourite-tooltip"
					onToggleFavourite={fn()}
				/>
			</div>
		))}
	</>
);

const meta = {
	title: 'Features/Gear/SelectorModal/SelectorModal',
	component: SelectorModalView,
	tags: ['autodocs'],
	args: {
		id: 'storybook-selector-modal',
		open: true,
		onOpenChange: fn(),
		title: 'Main Hand',
		tabs: [
			{ label: SelectorModalTabs.Items, pane: <ItemsPane /> },
			{ label: SelectorModalTabs.Enchants, pane: <EnchantsPane /> },
		],
		activeTab: SelectorModalTabs.Items,
		onSelectTab: fn(),
		sequence: 1,
	},
	argTypes: { tabs: { control: false }, rail: { control: false } },
} satisfies Meta<typeof SelectorModalView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ItemsTab: Story = {};

export const EnchantsTab: Story = { args: { activeTab: SelectorModalTabs.Enchants } };
