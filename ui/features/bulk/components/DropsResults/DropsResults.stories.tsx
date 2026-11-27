import { wowheadIconUrl } from '@sim/proto/wowhead';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsCardsFixture, type DropsRowFixture, dropsRowsFixture } from '@testing/storybook/fixtures/bulk-drops';
import { ItemCell, ItemCellIcon } from '@ui-kit/ItemCell';
import { expect, fn, userEvent, within } from 'storybook/test';

import { BulkResultRowView } from '../BulkResults/BulkResultRow';
import { DropsBossCardView } from '../DropsBossCard';
import { DropsItemLabel } from '../DropsItemLabel';
import { DropsSummaryView } from '../DropsSummary';
import { DropsResultsView } from './DropsResults';

const ORDERS = ['recommended', 'gain', 'upgrades', 'zone'] as const;

const cards = dropsCardsFixture();
const rows = dropsRowsFixture();

// The real row fills every slot from the database; a story draws only the new item's cell, beside its label.
const row = ({ label, ...props }: DropsRowFixture, key: number) => (
	<BulkResultRowView key={key} showDamage onEquip={fn()} {...props}>
		{label && (
			<div className="flex min-w-0 items-center gap-2">
				<ItemCell
					className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
					icon={
						<ItemCellIcon ilvl={label.item.ilvl}>
							<a className="ui-item-picker-icon" style={{ backgroundImage: `url('${wowheadIconUrl(label.item.item.icon, 'large')}')` }} />
						</ItemCellIcon>
					}
					name={label.item.item.name}
				/>
				<DropsItemLabel {...label} />
			</div>
		)}
	</BulkResultRowView>
);

const summary = (
	<DropsSummaryView
		current={{ avg: 750, plusMinus: 2 }}
		bestDrop={{ name: 'Felstriker', dps: 780, pct: 4 }}
		farmFirst={{ label: 'Warchief Rend Blackhand', detail: 'Blackrock Spire', upgrades: 2, total: 5 }}
	/>
);

const meta = {
	title: 'Features/Bulk/DropsResults',
	component: DropsResultsView,
	tags: ['autodocs'],
	args: {
		summary,
		view: 'bosses',
		onViewChange: fn(),
		order: 'recommended',
		orders: ORDERS,
		onOrderChange: fn(),
		filter: null,
		cards: cards.map(card => <DropsBossCardView key={card.label.label} {...card} />),
		rows: rows.map(row),
	},
	argTypes: { summary: { control: false }, cards: { control: false }, rows: { control: false }, filter: { control: false } },
} satisfies Meta<typeof DropsResultsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Bosses: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'bulk_tab.drops.results.view.items' }));
		await expect(args.onViewChange).toHaveBeenCalledWith('items');
	},
};

/** A run with Combined gain offers that order too; an order other than Recommended hides the cards' rank. */
export const OrderedByCombinedGain: Story = {
	args: {
		order: 'combined',
		orders: [...ORDERS, 'combined'],
		cards: cards.map(card => <DropsBossCardView key={card.label.label} {...card} rank={null} />),
	},
};

export const Items: Story = { args: { view: 'items' } };

/** An item that upgrades in both ring slots lists a row for each. */
export const ItemsTwoUpgradeRows: Story = {
	args: {
		view: 'items',
		rows: [rows[0], rows[1], { ...rows[1], label: { ...rows[1].label!, slot: 'Finger 2' }, avg: 762.1 }, ...rows.slice(2)].map(row),
	},
};

export const FilteredBySource: Story = {
	args: {
		view: 'items',
		filter: { label: 'Warchief Rend Blackhand', onClear: fn() },
		rows: [rows[0], rows[2]].map(row),
	},
};

/** Every card dimmed: no item beat the current gear. */
export const NoUpgrade: Story = {
	args: {
		summary: <DropsSummaryView current={{ avg: 750, plusMinus: 2 }} bestDrop={null} farmFirst={null} />,
		cards: cards.slice(2).map(card => <DropsBossCardView key={card.label.label} {...card} />),
	},
};
