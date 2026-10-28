import type { Meta, StoryObj } from '@storybook/react-vite';
import { MenuItem } from '@ui-kit/Menu';
import { expect, fireEvent, fn, userEvent, waitFor, within } from 'storybook/test';

import { ContextMenu } from './ContextMenu';

const meta = {
	title: 'Kit/ContextMenu',
	component: ContextMenu,
	tags: ['autodocs'],
	args: {
		onOpenChange: fn(),
		triggerProps: { className: 'border border-primary p-8 text-white' },
		children: 'Right click here',
		menu: (
			<>
				<MenuItem layout="block">Duplicate</MenuItem>
				<MenuItem layout="block">Delete</MenuItem>
			</>
		),
	},
	argTypes: { menu: { control: false }, children: { control: 'text' } },
} satisfies Meta<typeof ContextMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const OpensOnRightClick: Story = {
	play: async ({ canvasElement }) => {
		fireEvent.contextMenu(within(canvasElement).getByText('Right click here'), { clientX: 40, clientY: 40 });
		const body = within(canvasElement.ownerDocument.body);
		await waitFor(() => expect(body.getByRole('menuitem', { name: 'Duplicate' })).toBeVisible());
		await userEvent.keyboard('{Escape}');
		await waitFor(() => expect(body.queryByRole('menuitem', { name: 'Duplicate' })).toBeNull());
	},
};
