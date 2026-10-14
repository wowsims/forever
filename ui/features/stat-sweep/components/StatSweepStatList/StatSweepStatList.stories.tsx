import type { Meta, StoryObj } from '@storybook/react-vite';

import { StatSweepStatList } from './StatSweepStatList';

const meta = {
	title: 'Features/StatSweep/StatSweepStatList',
	component: StatSweepStatList,
	tags: ['autodocs'],
	args: { labels: ['Strength', 'Agility', 'Melee Hit', 'Melee Crit'], running: 1 },
} satisfies Meta<typeof StatSweepStatList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MidRun: Story = {};

export const FirstStat: Story = { args: { running: 0 } };

export const OneStat: Story = { args: { labels: ['Strength'], running: 0 } };
