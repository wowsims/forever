import type { Meta, StoryObj } from '@storybook/react-vite';

import { GearPlannerNotice } from './GearPlannerNotice';

const meta = {
	title: 'Features/Results/SimResultsPanel/Components/GearPlannerNotice',
	component: GearPlannerNotice,
	tags: ['autodocs'],
} satisfies Meta<typeof GearPlannerNotice>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
