import type { Meta, StoryObj } from '@storybook/react-vite';

import { ProgressStepList } from './ProgressStepList';

const meta = {
	title: 'Kit/ProgressStepList',
	component: ProgressStepList,
	tags: ['autodocs'],
	args: { labels: ['Current · 3:00 ±5s · 1 target', 'Patchwerk · 5:00 ±30s · 1 target', 'Council · 4:00 ±10s · 3 targets'], running: 1 },
} satisfies Meta<typeof ProgressStepList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MidRun: Story = {};

export const FirstStep: Story = { args: { running: 0 } };

export const LastStep: Story = { args: { running: 2 } };
