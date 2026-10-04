import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SidebarDisabledContext } from './context/sidebar_disabled';
import { SidebarActionButton } from './SidebarActionButton';

const meta = {
	title: 'Kit/SidebarActionButton',
	component: SidebarActionButton,
	tags: ['autodocs'],
	args: {
		children: 'Simulate',
		onClick: fn(),
		testId: 'sidebar-action-story',
	},
	argTypes: {
		disabled: { control: 'boolean' },
		loading: { control: 'boolean' },
		children: { control: 'text' },
		className: { control: 'text' },
	},
	decorators: [
		Story => (
			<div className="flex w-60 flex-col items-center">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SidebarActionButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Idle: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Simulate' }));
		await expect(args.onClick).toHaveBeenCalledOnce();
	},
};
export const Loading: Story = { args: { children: 'Stat Weights', loading: true } };
export const Disabled: Story = { args: { disabled: true } };
export const ContextDisabled: Story = {
	decorators: [
		Story => (
			<SidebarDisabledContext value>
				<Story />
			</SidebarDisabledContext>
		),
	],
};
