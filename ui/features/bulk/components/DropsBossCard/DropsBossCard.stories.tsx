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

// A worn item shows "worn" in its strip slot.
export const Default: Story = {
	args: { onSelect: fn() },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('drops-boss-card'));
		await expect(args.onSelect).toHaveBeenCalledOnce();
	},
};

export const AllUpgrades: Story = { args: second };

export const Selected: Story = { args: { selected: true } };

/** No upgrade: the card dims and names no best gain. */
export const NoUpgrade: Story = { args: none };
