import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { PlannerToolbar } from './PlannerToolbar';

const meta = {
	title: 'Features/EncounterPlanner/PlannerToolbar',
	component: PlannerToolbar,
	tags: ['autodocs'],
	args: { snapSeconds: 1, onSnapChange: fn(), onZoomIn: fn(), onZoomOut: fn(), onFit: fn(), onAdd: fn() },
} satisfies Meta<typeof PlannerToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const SnapOff: Story = { args: { snapSeconds: 0 } };
