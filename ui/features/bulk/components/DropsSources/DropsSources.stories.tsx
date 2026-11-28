import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsTreeFixture } from '@testing/storybook/fixtures/bulk-drops';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DropsSourcesView } from './DropsSources';

const meta = {
	title: 'Features/Bulk/DropsSources',
	component: DropsSourcesView,
	tags: ['autodocs'],
	args: { tree: dropsTreeFixture(), ticked: new Set<string>(), onTick: fn(), onQuickPick: fn() },
	argTypes: { tree: { control: false }, ticked: { control: false } },
	decorators: [
		Story => (
			<div className="w-full max-w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsSourcesView>;

export default meta;
type Story = StoryObj<typeof meta>;

// Nothing ticked, every node closed.
export const Default: Story = {
	play: async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: i18n.t('bulk_tab.drops.setup.quick_picks.clear') }));
		await expect(args.onQuickPick).toHaveBeenCalledWith('clear');
	},
};

// Two of the three Blackrock Spire bosses ticked: the zone and Dungeons read mixed, and Raids and World drops stay closed.
export const Mixed: Story = {
	args: { ticked: new Set(['drop:1583:10429', 'drop:1583:10363']) },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: i18n.t('bulk_tab.drops.sources.categories.drop_dungeon') }));
		await userEvent.click(canvas.getByRole('button', { name: 'Blackrock Spire' }));
		await expect(canvas.getByRole('checkbox', { name: 'Blackrock Spire' }).getAttribute('aria-checked')).toBe('mixed');
		await expect(canvas.getByRole('checkbox', { name: i18n.t('bulk_tab.drops.sources.categories.drop_dungeon') }).getAttribute('aria-checked')).toBe(
			'mixed',
		);
		for (const category of ['drop_raid', 'drop_world']) {
			await expect(canvas.getByRole('checkbox', { name: i18n.t(`bulk_tab.drops.sources.categories.${category}`) }).getAttribute('aria-checked')).toBe(
				'false',
			);
		}
	},
};

export const Empty: Story = { args: { tree: [] } };
