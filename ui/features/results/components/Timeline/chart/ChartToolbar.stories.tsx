import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ChartToolbar } from './ChartToolbar';

const meta = {
	title: 'Features/Results/Timeline/Components/ChartToolbar',
	component: ChartToolbar,
	tags: ['autodocs'],
	args: { onReset: fn(), onZoomIn: fn(), onZoomOut: fn(), onPanLeft: fn(), onPanRight: fn() },
	decorators: [
		Story => (
			<div className="flex">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ChartToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		for (const button of within(canvasElement).getAllByRole('button')) await userEvent.click(button);
		await expect([args.onReset, args.onZoomOut, args.onZoomIn, args.onPanLeft, args.onPanRight].every(handler => handler.mock.calls.length === 1)).toBe(
			true,
		);
	},
};
