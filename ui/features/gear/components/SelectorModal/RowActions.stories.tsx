import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { RowActions } from './RowActions';

const meta = {
	title: 'Features/Gear/SelectorModal/Components/RowActions',
	component: RowActions,
	tags: ['autodocs'],
	args: {
		favourited: false,
		favouriteTooltipId: 'favourite-tooltip',
		onToggleFavourite: fn(),
		compare: { inBatch: false, tooltipId: 'compare-tooltip', onToggle: fn() },
	},
	decorators: [
		Story => (
			<div className="flex gap-4">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof RowActions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('selector-modal-list-item-favorite'));
		await expect(args.onToggleFavourite).toHaveBeenCalledOnce();
	},
};

export const FavouriteInBatch: Story = { args: { favourited: true, compare: { inBatch: true, tooltipId: 'compare-tooltip', onToggle: fn() } } };
