import type { Meta, StoryObj } from '@storybook/react-vite';
import { LOG_SUGGESTIONS } from '@testing/storybook/fixtures/results-views';
import { ToolbarButton } from '@ui-kit/Toolbar';
import { fn } from 'storybook/test';

import { LogToolbar } from './LogToolbar';

const meta = {
	title: 'Features/Results/LogRunner/Components/LogToolbar',
	component: LogToolbar,
	tags: ['autodocs'],
	args: {
		groups: [],
		suggestions: LOG_SUGGESTIONS,
		onChange: fn(),
		children: <ToolbarButton className="whitespace-nowrap">Export</ToolbarButton>,
	},
	argTypes: { groups: { control: false }, suggestions: { control: false }, children: { control: false } },
} satisfies Meta<typeof LogToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NoFilters: Story = {};

export const Filtered: Story = {
	args: {
		groups: [
			{ id: 0, field: 'spell', join: 'or', values: ['Mortal Strike', 'Heroic Strike'] },
			{ id: 1, field: 'outcome', join: 'or', values: ['crit'] },
			{ id: 2, field: 'source', join: 'or', values: ['Player (#1)'] },
			{ id: 3, field: 'amount', join: 'and', values: ['>5000'] },
		],
	},
};
