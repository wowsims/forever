import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, screen, userEvent, within } from 'storybook/test';

import { Chip } from './Chip';

const meta = {
	title: 'Kit/Chip',
	component: Chip,
	tags: ['autodocs'],
	args: {
		label: 'Phase 1 Raid',
		onSelect: fn(),
	},
	argTypes: {
		label: { control: 'text' },
		active: { control: 'boolean' },
		disabled: { control: 'boolean' },
		confirmDelete: { control: 'boolean' },
		className: { control: 'text' },
		as: { control: false },
		nameAs: { control: false },
		rootProps: { control: false },
		nameProps: { control: false },
		deleteMessage: { control: false },
	},
	render: args => (
		<div className="flex">
			<Chip {...args} />
		</div>
	),
} satisfies Meta<typeof Chip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Active: Story = { args: { active: true } };
export const Disabled: Story = {
	args: { disabled: true },
	parameters: { docs: { description: { story: 'A disabled chip is hidden, so the canvas looks empty.' } } },
};
export const Deletable: Story = {
	args: { onDelete: fn(), confirmDelete: false, deleteLabel: 'Delete' },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Delete' }));
		await expect(args.onDelete).toHaveBeenCalledOnce();
	},
};
export const ConfirmDelete: Story = {
	args: {
		onDelete: fn(),
		deleteLabel: 'Delete',
		deleteMessage: "Delete saved Gear Set 'Phase 1 Raid'?",
		deleteConfirmLabel: 'Confirm',
	},
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Delete' }));
		await expect(args.onDelete).not.toHaveBeenCalled();
		await userEvent.click(await screen.findByRole('button', { name: 'Confirm' }));
		await expect(args.onDelete).toHaveBeenCalledOnce();
	},
};
