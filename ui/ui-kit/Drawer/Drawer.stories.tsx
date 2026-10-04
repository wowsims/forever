import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Drawer } from './Drawer';

const SIDES = ['left', 'bottom', 'right'] as const;

const meta = {
	title: 'Kit/Drawer',
	component: Drawer,
	tags: ['autodocs'],
	args: {
		trigger: <Button>Open drawer</Button>,
		onOpenChange: fn(),
		side: 'bottom',
		children: <p className="m-0">Drawer contents.</p>,
	},
	argTypes: {
		side: { control: 'select', options: SIDES },
		trigger: { control: false },
		children: { control: false },
		style: { control: false },
		ignoreOutsidePress: { control: false },
	},
} satisfies Meta<typeof Drawer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Sides: Story = {
	render: args => (
		<div className="flex gap-2">
			{SIDES.map(side => (
				<Drawer key={side} {...args} side={side} trigger={<Button>Open {side}</Button>} />
			))}
		</div>
	),
};

export const NonModal: Story = {
	args: { modal: false, children: <p className="m-0">The page behind stays interactive.</p> },
};

export const Opens: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Open drawer' }));
		const drawer = await canvas.findByRole('dialog');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(drawer)).toBe(true);
		await expect(args.onOpenChange).toHaveBeenCalledWith(true);
	},
};
