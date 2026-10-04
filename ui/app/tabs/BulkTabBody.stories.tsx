import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { BulkTabActionsView } from './BulkTabBody';

const meta = {
	title: 'App/Tabs/BulkTabBody',
	component: BulkTabActionsView,
	tags: ['autodocs'],
	args: {
		showNativeDownload: true,
		onImportBags: fn(),
		onImportFavorites: fn(),
		onClear: fn(),
	},
	decorators: [
		Story => (
			<div className="grid gap-6">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkTabActionsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const actions = within(within(canvasElement).getByTestId('bulk-gear-actions')).getAllByRole('button');
		await userEvent.click(actions[actions.length - 1]);
		await expect(args.onClear).toHaveBeenCalledOnce();
	},
};
