import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ForeverNoticeView } from './ForeverNotice';

const meta = {
	title: 'App/ForeverNotice',
	component: ForeverNoticeView,
	tags: ['autodocs'],
	args: { onRestoreDefaults: fn() },
} satisfies Meta<typeof ForeverNoticeView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button'));
		await expect(args.onRestoreDefaults).toHaveBeenCalledTimes(1);
	},
};

export const BeforeTheHost: Story = { args: { onRestoreDefaults: undefined } };
