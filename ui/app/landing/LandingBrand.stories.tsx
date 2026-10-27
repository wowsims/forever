import type { Meta, StoryObj } from '@storybook/react-vite';

import { LandingBrand } from './LandingBrand';

const meta = {
	title: 'App/Landing/Components/LandingBrand',
	component: LandingBrand,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
} satisfies Meta<typeof LandingBrand>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
