import { GemColor } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIGem } from '@testing/storybook/fixtures/protos';
import { cssVars } from '@ui-kit/utils/css';
import { expect, fn, userEvent, within } from 'storybook/test';

import { GemSocket } from './GemSocket';

const meta = {
	title: 'Features/Gear/ItemCell/Components/GemSocket',
	component: GemSocket,
	tags: ['autodocs'],
	args: { socketColor: GemColor.GemColorRed, gem: null, onActivate: fn() },
	argTypes: { gem: { control: false } },
	// `socketColor` is a `GemColor` number, not a CSS colour, so the preview's colour matcher is narrowed.
	parameters: { controls: { matchers: { color: /^(background|color)$/i } } },
	decorators: [
		Story => (
			<div className="flex" style={cssVars({ '--gem-width': '2rem' })}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof GemSocket>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EmptyRedSocket: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('gem-socket-container'));
		await expect(args.onActivate).toHaveBeenCalledOnce();
	},
};

export const EmptyYellowSocket: Story = { args: { socketColor: GemColor.GemColorYellow } };

export const Socketed: Story = { args: { gem: makeUIGem() } };
