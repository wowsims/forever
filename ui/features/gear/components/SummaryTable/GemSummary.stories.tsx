import { GemColor, ItemQuality } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { makeUIGem } from '@testing/storybook/fixtures/protos';
import { expect, fn, userEvent, within } from 'storybook/test';

import { GemSummaryView } from './GemSummary';

const meta = {
	title: 'Features/Gear/SummaryTable/Components/GemSummary',
	component: GemSummaryView,
	tags: ['autodocs'],
	args: {
		rows: [
			{ gem: makeUIGem(), count: 4, iconUrl: wowheadIconUrl('inv_misc_gem_ruby_01') },
			{
				gem: makeUIGem({ id: 2, name: 'Citrine', color: GemColor.GemColorYellow, quality: ItemQuality.ItemQualityUncommon }),
				count: 2,
				iconUrl: wowheadIconUrl('inv_misc_gem_topaz_01'),
			},
		],
		onReset: fn(),
	},
	argTypes: { rows: { control: false } },
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof GemSummaryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('summary-table-reset-button'));
		await expect(args.onReset).toHaveBeenCalledOnce();
	},
};

export const IconsResolving: Story = { args: { rows: [{ gem: makeUIGem(), count: 1 }] } };
