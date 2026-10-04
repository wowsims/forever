import type { Meta, StoryObj } from '@storybook/react-vite';

import { LandingPage } from './LandingPage';

const meta = {
	title: 'App/Landing/LandingPage',
	component: LandingPage,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
} satisfies Meta<typeof LandingPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
