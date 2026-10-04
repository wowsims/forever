import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SimToolbarView } from './SimToolbar';

const meta = {
	title: 'App/Header/SimToolbar',
	component: SimToolbarView,
	tags: ['autodocs'],
	args: {
		knownIssues: [],
		outdatedNativeSim: false,
		onOpenSettings: fn(),
	},
	argTypes: { knownIssues: { control: false } },
	decorators: [
		Story => (
			<div className="flex list-none flex-nowrap items-end pl-0 text-(length:--text-ui) font-bold">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SimToolbarView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('sim-options'));
		await expect(args.onOpenSettings).toHaveBeenCalledOnce();
	},
};

export const KnownIssues: Story = {
	args: { knownIssues: ['Bloodrage costs no health.', 'Deep Wounds ticks on its own schedule.'] },
};

export const OutdatedNativeSim: Story = { args: { outdatedNativeSim: true } };
