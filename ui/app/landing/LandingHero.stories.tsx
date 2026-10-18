import type { Meta, StoryObj } from '@storybook/react-vite';

import { LandingHero } from './LandingHero';

const meta = {
	title: 'App/Landing/Components/LandingHero',
	component: LandingHero,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
} satisfies Meta<typeof LandingHero>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
