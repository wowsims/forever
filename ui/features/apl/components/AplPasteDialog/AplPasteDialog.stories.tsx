import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { AplPasteDialog, type AplPasteDialogProps } from './AplPasteDialog';

const AplPasteDialogStory = ({ open: initialOpen, onClose, ...props }: AplPasteDialogProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<>
			<Button onClick={() => setOpen(true)}>{props.title}</Button>
			<AplPasteDialog
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
	title: 'Features/Apl/AplPasteDialog',
	component: AplPasteDialog,
	tags: ['autodocs'],
	args: { open: false, title: 'Paste Action', onSubmit: fn(), onClose: fn() },
	render: args => <AplPasteDialogStory {...args} />,
} satisfies Meta<typeof AplPasteDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The page could not read the clipboard, so the paste button opens this: the confirm button waits for text. */
export const CannotReadClipboard: Story = { args: { open: true } };
