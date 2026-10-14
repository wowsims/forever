import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Dialog, type DialogProps, type DialogSize } from './Dialog';

const SIZES: Array<DialogSize> = ['sm', 'md', 'lg', 'xl', 'full'];

interface DialogStoryProps extends DialogProps {
	trigger?: string;
}

const DialogStory = ({ trigger = 'Open dialog', open: initialOpen, onOpenChange, ...props }: DialogStoryProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<>
			<Button onClick={() => setOpen(true)}>{trigger}</Button>
			<Dialog
				{...props}
				open={open}
				onOpenChange={next => {
					setOpen(next);
					onOpenChange(next);
				}}
			/>
		</>
	);
};

const PreventCloseStory = ({ open: initialOpen, onOpenChange, ...props }: DialogProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<>
			<Button onClick={() => setOpen(true)}>Open dialog</Button>
			<Dialog
				{...props}
				open={open}
				onOpenChange={onOpenChange}
				preventClose
				footer={
					<Button variant="cancel" onClick={() => setOpen(false)}>
						Stop
					</Button>
				}
			/>
		</>
	);
};

const meta = {
	title: 'Kit/Dialog',
	component: Dialog,
	tags: ['autodocs'],
	args: {
		open: false,
		onOpenChange: fn(),
		title: 'Options',
		children: <p>Dialog body.</p>,
	},
	argTypes: {
		size: { control: 'select', options: SIZES },
		children: { control: false },
		footer: { control: false },
		headerChildren: { control: false },
	},
	render: args => <DialogStory {...args} />,
} satisfies Meta<typeof Dialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Sizes: Story = {
	render: args => (
		<div className="flex gap-2">
			{SIZES.map(size => (
				<DialogStory key={size} {...args} size={size} title={`${size} dialog`} trigger={size} />
			))}
		</div>
	),
};

export const WithFooter: Story = {
	args: {
		footer: (
			<>
				<Button variant="cancel">Cancel</Button>
				<Button>Save</Button>
			</>
		),
	},
};

export const Elevated: Story = {
	args: {
		children: <DialogStory open={false} onOpenChange={fn()} elevated title="Elevated" trigger="Open elevated dialog" />,
	},
};

export const VerticalAlignTop: Story = {
	args: { verticalAlign: 'top' },
};

export const VerticalAlignCenter: Story = {
	args: { verticalAlign: 'center' },
};

export const ScrollContents: Story = {
	args: {
		scrollContents: true,
		title: 'Long list',
		children: (
			<ol className="m-0">
				{Array.from({ length: 60 }, (_, index) => (
					<li key={index}>Row {index + 1}</li>
				))}
			</ol>
		),
	},
};

/** The viewport inset in both directions, for a tool that wants all the room it can get; the body scrolls inside it. */
export const Full: Story = {
	args: { size: 'full', scrollContents: true, title: 'Stat sweep', footer: <Button>Run</Button> },
};

export const PreventClose: Story = {
	render: args => <PreventCloseStory {...args} />,
	args: { title: 'Simulating', children: <p>No close button, no Escape, no backdrop press: only the footer ends it.</p> },
};

export const Opens: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Open dialog' }));
		const dialog = await canvas.findByRole('dialog');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(dialog)).toBe(true);
	},
};
