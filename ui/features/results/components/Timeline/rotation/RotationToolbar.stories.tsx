import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { RotationToolbar } from './RotationToolbar';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationToolbar',
	component: RotationToolbar,
	tags: ['autodocs'],
	args: { onZoomOut: fn(), onZoomIn: fn(), onFit: fn(), onReset: fn() },
	decorators: [
		Story => (
			<div className="flex h-6 text-white [--label-w:12rem]">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof RotationToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const [zoomOut, zoomIn, fit, reset] = within(canvasElement).getAllByTestId('rotation-zoom-button');
		await userEvent.click(zoomOut);
		await userEvent.click(zoomIn);
		await userEvent.click(fit);
		await userEvent.click(reset);
		await expect([args.onZoomOut, args.onZoomIn, args.onFit, args.onReset].every(handler => handler.mock.calls.length === 1)).toBe(true);
	},
};
