import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { LogRunnerHeaderView, LogRunnerNoMatches } from './LogRunner';

const meta = {
	title: 'Features/Results/LogRunner/LogRunner',
	component: LogRunnerHeaderView,
	tags: ['autodocs'],
	args: { searchText: '', onSearchChange: fn() },
} satisfies Meta<typeof LogRunnerHeaderView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const Searching: Story = { args: { searchText: '"Mortal Strike" crit' } };

export const NoMatches: Story = { render: () => <LogRunnerNoMatches /> };
