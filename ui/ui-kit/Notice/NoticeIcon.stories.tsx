import type { Meta, StoryObj } from '@storybook/react-vite';
import type { IconSize } from '@ui-kit/Icon';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { NoticeIcon } from './NoticeIcon';
import type { NoticeLevel } from './noticeLevel';

const LEVELS: NoticeLevel[] = ['info', 'warning', 'error'];
const SIZES: IconSize[] = ['2xs', 'xs', 'sm', '1x', 'lg', 'xl', '2xl', '2x', '3x', '4x', '5x'];

const meta = {
	title: 'Kit/NoticeIcon',
	component: NoticeIcon,
	tags: ['autodocs'],
	args: { level: 'warning', label: 'Warnings', children: <p>Whirlwind is never castable: its cooldown is longer than the fight.</p> },
	argTypes: {
		level: { control: 'select', options: LEVELS },
		label: { control: 'text' },
		size: { control: 'select', options: SIZES },
		hidden: { control: 'boolean' },
		className: { control: 'text' },
		children: { control: false },
	},
} satisfies Meta<typeof NoticeIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Info: Story = { args: { level: 'info', label: 'Additional information', children: <p>Set bonuses count every equipped piece.</p> } };
export const Warning: Story = {};
export const ErrorLevel: Story = { name: 'Error', args: { level: 'error', label: 'Errors', children: <p>The rotation has an invalid action.</p> } };

export const OpensOnHover: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.hover(within(canvasElement).getByRole('button'));
		const content = await within(canvasElement.ownerDocument.body).findByText('Whirlwind is never castable: its cooldown is longer than the fight.');
		await waitFor(() => expect(content).toBeVisible());
	},
};
