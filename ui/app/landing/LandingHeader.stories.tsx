import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';

import { LandingHeader } from './LandingHeader';

const meta = {
	title: 'App/Landing/Components/LandingHeader',
	component: LandingHeader,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
} satisfies Meta<typeof LandingHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ canvasElement }) => {
		const [open] = within(canvasElement).getAllByTestId('navbar-toggler');
		await userEvent.click(open);
		await expect(open.getAttribute('aria-expanded')).toBe('true');
	},
};
