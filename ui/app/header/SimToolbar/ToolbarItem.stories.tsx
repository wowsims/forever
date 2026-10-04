import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ToolbarItem } from './ToolbarItem';

const meta = {
	title: 'App/Header/Components/ToolbarItem',
	component: ToolbarItem,
	tags: ['autodocs'],
	args: { icon: 'gear', tooltip: 'Sim options', testId: 'toolbar-item', onClick: fn() },
	decorators: [
		Story => (
			<div className="flex items-end font-bold">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ToolbarItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const IconButton: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('toolbar-item'));
		await expect(args.onClick).toHaveBeenCalledOnce();
	},
};

export const Link: Story = { args: { icon: 'bug', tooltip: 'Report a bug', href: 'https://github.com/wowsims/forever/issues/new/choose', onClick: undefined } };

export const TextEntry: Story = {
	args: { icon: undefined, className: 'text-link-danger', tooltip: 'Rage decay is off', children: 'Known Issues' },
};

export const Hidden: Story = { args: { hidden: true } };
