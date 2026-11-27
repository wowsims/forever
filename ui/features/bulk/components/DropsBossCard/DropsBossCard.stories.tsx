import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsCardsFixture } from '@testing/storybook/fixtures/bulk-drops';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DropsBossCardView } from './DropsBossCard';

const [top, second, none] = dropsCardsFixture();

const meta = {
	title: 'Features/Bulk/DropsBossCard',
	component: DropsBossCardView,
	tags: ['autodocs'],
	args: top,
	argTypes: { strip: { control: 'object' }, label: { control: 'object' }, best: { control: 'object' } },
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsBossCardView>;

export default meta;
type Story = StoryObj<typeof meta>;

const onEquip = fn();

// A worn item shows "Equipped" in its strip slot; every other icon carries a [+] that equips it.
export const Default: Story = {
	args: { onSelect: fn(), strip: top.strip.map(entry => (entry.gainPct === null ? entry : { ...entry, onEquip })) },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByTestId('drops-boss-select'));
		await expect(args.onSelect).toHaveBeenCalledOnce();
		await userEvent.click(canvas.getAllByTestId('drops-boss-equip')[0]);
		await expect(onEquip).toHaveBeenCalledOnce();
		await expect(args.onSelect).toHaveBeenCalledOnce();
	},
};

export const AllUpgrades: Story = { args: second };

/** The source's upgrades simmed together as one set; the (i) names the items in it. */
export const CombinedGain: Story = {
	args: { ...second, combined: { dps: 768.2, pct: 2.43, marginPct: 0.12, items: ['Mark of the Dragon Lord', 'Reiver Claws'], leftOut: [] } },
};

/** Two hand upgrades cannot both be worn, so the set leaves the smaller one out and says so. */
export const CombinedGainPartial: Story = {
	args: {
		...second,
		strip: [...second.strip, { itemId: 2004, icon: second.strip[1].icon, name: 'Dracorian Gauntlets', ilvl: 63, gainPct: 0.2, upgrade: true }],
		combined: { dps: 768.2, pct: 2.43, marginPct: 0.12, items: ['Mark of the Dragon Lord', 'Reiver Claws'], leftOut: ['Dracorian Gauntlets'] },
	},
};

export const Selected: Story = { args: { selected: true } };

/** An upgrade card with its rank hidden, as every order but Recommended shows it. */
export const Unranked: Story = { args: { rank: null } };

/** No upgrade: the card dims and names no best gain. */
export const NoUpgrade: Story = { args: none };

/** Items with no icon keep their tile and item level, so each % stays under its own item. */
export const MissingIcons: Story = {
	args: {
		...none,
		label: { label: 'Darkmaster Gandling', detail: 'Scholomance' },
		strip: [
			{ itemId: 1, icon: '', name: 'Missing icon', ilvl: 61, gainPct: -3.43, upgrade: false },
			{ itemId: 2, icon: second.strip[0].icon, name: second.strip[0].name, ilvl: second.strip[0].ilvl, gainPct: -18.2, upgrade: false },
			{ itemId: 3, icon: '', name: 'Missing icon too', ilvl: 63, gainPct: -25.35, upgrade: false },
			{ itemId: 4, icon: '', name: 'Worn, missing icon', ilvl: 58, gainPct: null, upgrade: false },
		],
	},
};
