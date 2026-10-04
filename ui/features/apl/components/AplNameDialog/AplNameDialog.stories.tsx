import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { AplNameDialog, type AplNameDialogProps } from './AplNameDialog';

const AplNameDialogStory = ({ open: initialOpen, onClose, ...props }: AplNameDialogProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<>
			<Button onClick={() => setOpen(true)}>{props.title}</Button>
			<AplNameDialog
				{...props}
				open={open}
				onClose={() => {
					setOpen(false);
					onClose();
				}}
			/>
		</>
	);
};

const meta = {
	title: 'Features/Apl/AplNameDialog',
	component: AplNameDialog,
	tags: ['autodocs'],
	args: {
		open: false,
		title: 'New Group',
		inputLabel: 'Name',
		existingNames: ['Execute', 'Opener'],
		onSubmit: fn(),
		onCancel: fn(),
		onClose: fn(),
	},
	render: args => <AplNameDialogStory {...args} />,
} satisfies Meta<typeof AplNameDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Open on a rename with a taken name in the list: type one of `existingNames` to see the conflict block the confirm button. */
export const Default: Story = {
	args: {
		open: true,
		title: 'Rename Group',
		confirmLabel: 'Rename',
		defaultValue: 'Opener',
		placeholder: 'rage_pooled',
		existingNames: ['Execute', 'Opener'],
	},
};
