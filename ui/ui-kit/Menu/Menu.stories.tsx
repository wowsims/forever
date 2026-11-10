import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { Menu } from './Menu';
import { MenuItem } from './MenuItem';
import { MenuLinkItem } from './MenuLinkItem';

const ITEMS = (
	<>
		<MenuItem layout="block">Import</MenuItem>
		<MenuItem layout="block">Export</MenuItem>
		<MenuItem layout="block" disabled>
			Unsupported
		</MenuItem>
		<MenuLinkItem layout="block" href="#menu-link">
			Open link
		</MenuLinkItem>
	</>
);

const meta = {
	title: 'Kit/Menu',
	component: Menu,
	tags: ['autodocs'],
	args: {
		trigger: 'Open menu',
		triggerProps: { className: 'border border-primary px-4 py-2 text-white' },
		surface: 'menu',
		align: 'start',
		children: ITEMS,
	},
	argTypes: {
		surface: { control: 'select', options: ['menu', 'plain', 'none'] },
		side: { control: 'select', options: [undefined, 'top', 'right', 'bottom', 'left'] },
		align: { control: 'select', options: [undefined, 'start', 'center', 'end'] },
		children: { control: false },
		trigger: { control: 'text' },
		container: { control: false },
		popupRender: { control: false },
	},
} satisfies Meta<typeof Menu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Submenu: Story = {
	args: {
		children: (
			<>
				<MenuItem layout="block">Import</MenuItem>
				<Menu submenu surface="menu" side="right" align="start" trigger="More" triggerRender={<button type="button" className="ui-menu-item" />}>
					<MenuItem layout="block">Nested item</MenuItem>
					<MenuItem layout="block">Another nested item</MenuItem>
				</Menu>
			</>
		),
	},
};

export const SurfacePlain: Story = {
	args: { surface: 'plain' },
};

export const SurfaceNone: Story = {
	args: { surface: 'none' },
};

export const BareItems: Story = {
	args: {
		children: (
			<MenuLinkItem
				layout="row"
				href="#bare-link"
				bare
				className="gap-2 p-2 hover:bg-primary-dampened data-highlighted:bg-primary-dampened"
				testId="bare-link">
				Bare link
			</MenuLinkItem>
		),
	},
};

export const Opens: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const trigger = canvas.getByRole('button', { name: 'Open menu' });
		await userEvent.click(trigger);
		const menu = await canvas.findByRole('menu');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(menu)).toBe(true);
		await expect(trigger.hasAttribute('data-popup-open')).toBe(true);
		await userEvent.click(within(menu).getByRole('menuitem', { name: 'Export' }));
		await waitFor(() => expect(trigger.hasAttribute('data-popup-open')).toBe(false));
	},
};
