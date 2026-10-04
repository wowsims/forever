import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ConfirmPopover, type ConfirmPopoverProps } from './ConfirmPopover';

const ConfirmPopoverStory = ({ open: initialOpen, onOpenChange, ...props }: ConfirmPopoverProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<ConfirmPopover
			{...props}
			open={open}
			onOpenChange={next => {
				setOpen(next);
				onOpenChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Kit/ConfirmPopover',
	component: ConfirmPopover,
	tags: ['autodocs'],
	args: {
		open: false,
		onOpenChange: fn(),
		onConfirm: fn(),
		trigger: 'Remove',
		triggerClassName: 'border border-danger px-4 py-2 text-danger',
		confirmLabel: 'Delete',
		children: "Delete saved gear 'Pre-raid'?",
	},
	argTypes: {
		side: { control: 'select', options: ['top', 'right', 'bottom', 'left'] },
		backdrop: { control: 'boolean' },
		children: { control: 'text' },
		trigger: { control: 'text' },
		anchor: { control: false },
	},
	render: args => (
		<div className="pt-24">
			<ConfirmPopoverStory {...args} />
		</div>
	),
} satisfies Meta<typeof ConfirmPopover>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const OkOnly: Story = {
	args: {
		onConfirm: undefined,
		confirmLabel: 'OK',
		trigger: 'Save',
		triggerClassName: 'border border-primary px-4 py-2 text-white',
		children: 'Enter a name first.',
	},
};

export const Confirms: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Remove' }));
		const popup = await canvas.findByRole('dialog');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(popup)).toBe(true);
		await userEvent.click(within(popup).getByRole('button', { name: 'Delete' }));
		await expect(args.onConfirm).toHaveBeenCalledTimes(1);
		await expect(args.onOpenChange).toHaveBeenLastCalledWith(false);
	},
};
