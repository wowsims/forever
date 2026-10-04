import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { useRef, useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Popover, type PopoverAlign, type PopoverSide } from './Popover';

const SIDES: Array<PopoverSide> = ['top', 'right', 'bottom', 'left'];
const ALIGNS: Array<PopoverAlign> = ['start', 'center', 'end'];

const AnchoredStory = () => {
	const anchor = useRef<HTMLDivElement>(null);
	const [open, setOpen] = useState(false);
	return (
		<div className="flex items-center gap-4">
			<Button onClick={() => setOpen(true)}>Show hint</Button>
			<div ref={anchor} className="border border-surface-border px-4 py-2">
				Anchor
			</div>
			<Popover anchor={anchor} open={open} onOpenChange={setOpen} side="right">
				The popup hangs off the box, not the button that opened it.
			</Popover>
		</div>
	);
};

const meta = {
	title: 'Kit/Popover',
	component: Popover,
	tags: ['autodocs'],
	args: {
		trigger: 'Settings',
		triggerClassName: 'border border-primary px-4 py-2 text-white',
		onOpenChange: fn(),
		children: <p className="m-0">Popover contents.</p>,
	},
	argTypes: {
		side: { control: 'select', options: SIDES },
		align: { control: 'select', options: ALIGNS },
		backdrop: { control: 'boolean' },
		children: { control: false },
		trigger: { control: 'text' },
		triggerRender: { control: false },
		anchor: { control: false },
		container: { control: false },
	},
} satisfies Meta<typeof Popover>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Settings' }));
		const popup = await canvas.findByRole('dialog');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(popup)).toBe(true);
	},
};

export const Sides: Story = {
	render: args => (
		<div className="grid grid-cols-3 place-items-center gap-x-60 gap-y-32 p-32">
			{SIDES.flatMap(side =>
				ALIGNS.map(align => (
					<Popover key={`${side}-${align}`} {...args} open side={side} align={align} trigger={`${side} ${align}`}>
						{side} / {align}
					</Popover>
				)),
			)}
		</div>
	),
};

export const ButtonTrigger: Story = {
	args: {
		triggerClassName: undefined,
		triggerRender: <Button variant="primary" />,
		triggerProps: { 'aria-label': 'Reforge settings' },
		trigger: <Icon name="gear" />,
	},
	play: async ({ canvasElement }) => {
		const trigger = within(canvasElement).getByRole('button', { name: 'Reforge settings' });
		await userEvent.click(trigger);
		await expect(trigger.hasAttribute('data-popup-open')).toBe(true);
	},
};

export const Anchored: Story = {
	render: () => <AnchoredStory />,
};
