import type { Meta, StoryObj } from '@storybook/react-vite';

import { LandingFooter } from './LandingFooter';

const meta = {
	title: 'App/Landing/Components/LandingFooter',
	component: LandingFooter,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
} satisfies Meta<typeof LandingFooter>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
