import type { Meta, StoryObj } from '@storybook/react-vite';

import { LandingActions } from './LandingActions';

const meta = {
	title: 'App/Landing/Components/LandingActions',
	component: LandingActions,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
} satisfies Meta<typeof LandingActions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
