import { Button } from '@ui-kit/Button';
import { type ComponentType, useState } from 'react';

export interface DialogStoryProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

interface DialogTriggerProps extends DialogStoryProps {
	dialog: ComponentType<DialogStoryProps>;
}

/** A button that opens `dialog`, which starts open or closed as the story's `open` arg says. */
export const DialogTrigger = ({ dialog: Dialog, open: initialOpen, onOpenChange }: DialogTriggerProps) => {
	const [open, setOpen] = useState(initialOpen);
	const change = (next: boolean) => {
		setOpen(next);
		onOpenChange(next);
	};
	return (
		<>
			<Button onClick={() => change(true)}>Open dialog</Button>
			<Dialog open={open} onOpenChange={change} />
		</>
	);
};
