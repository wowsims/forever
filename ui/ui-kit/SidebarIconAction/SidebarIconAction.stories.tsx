import type { Meta, StoryObj } from '@storybook/react-vite';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SidebarIconAction } from './SidebarIconAction';

const meta = {
	title: 'Kit/SidebarIconAction',
	component: SidebarIconAction,
	tags: ['autodocs'],
	args: { icon: 'scale-balanced', label: 'Calculate Stat Weights', testId: 'sidebar-icon-action-story', onClick: fn() },
	argTypes: { disabled: { control: 'boolean' } },
} satisfies Meta<typeof SidebarIconAction>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Idle: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Calculate Stat Weights' }));
		await expect(args.onClick).toHaveBeenCalledOnce();
	},
};

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
