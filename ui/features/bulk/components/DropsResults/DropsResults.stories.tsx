import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsCardsFixture, dropsRowsFixture } from '@testing/storybook/fixtures/bulk-drops';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DropsBossCardView } from '../DropsBossCard';
import { DropsItemRowView } from '../DropsItemRow';
import { DropsSummaryView } from '../DropsSummary';
import { DropsResultsView } from './DropsResults';

const cards = dropsCardsFixture();
const rows = dropsRowsFixture();

const summary = (
	<DropsSummaryView
		current={{ avg: 750, plusMinus: 2 }}
		bestDrop={{ name: 'Felstriker', dps: 780, pct: 4 }}
		bestBoss={{ label: 'Warchief Rend Blackhand', detail: 'Blackrock Spire', pct: 4 }}
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
		showAll: false,
		onShowAllChange: fn(),
		filter: null,
		cards: cards.map(card => <DropsBossCardView key={card.label.label} {...card} />),
		rows: rows.map((row, index) => <DropsItemRowView key={index} {...row} />),
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

export const Items: Story = { args: { view: 'items' } };

export const ItemsAllVariations: Story = {
	args: {
		view: 'items',
		showAll: true,
		rows: [rows[0], { ...rows[1], hiddenVariations: 0 }, { ...rows[1], rank: 3, slot: 'Finger 2', avg: 762.1, plusMinus: 2 }, ...rows.slice(2)].map(
			(row, index) => <DropsItemRowView key={index} {...row} />,
		),
	},
};

export const FilteredBySource: Story = {
	args: {
		view: 'items',
		filter: { label: 'Warchief Rend Blackhand', onClear: fn() },
		rows: [rows[0], rows[2]].map((row, index) => <DropsItemRowView key={index} {...row} />),
	},
};

/** Every card dimmed: no item beat the current gear. */
export const NoUpgrade: Story = {
	args: {
		summary: <DropsSummaryView current={{ avg: 750, plusMinus: 2 }} bestDrop={null} bestBoss={null} />,
		cards: cards.slice(2).map(card => <DropsBossCardView key={card.label.label} {...card} />),
	},
};
