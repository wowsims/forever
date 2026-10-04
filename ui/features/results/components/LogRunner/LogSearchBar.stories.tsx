import type { Meta, StoryObj } from '@storybook/react-vite';
import { LOG_SUGGESTIONS } from '@testing/storybook/fixtures/results-views';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { LogSearchBar, type LogSearchBarProps } from './LogSearchBar';

const StatefulBar = ({ groups: initial, onChange, ...props }: LogSearchBarProps) => {
	const [groups, setGroups] = useState(initial);
	return (
		<LogSearchBar
			{...props}
			groups={groups}
			onChange={next => {
				setGroups(next);
				onChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Results/LogRunner/Components/LogSearchBar',
	component: LogSearchBar,
	tags: ['autodocs'],
	args: {
		groups: [
			{ id: 0, field: 'spell', join: 'or', values: ['Mortal Strike', 'Heroic Strike'] },
			{ id: 1, field: 'outcome', join: 'or', values: ['crit'] },
		],
		suggestions: LOG_SUGGESTIONS,
		onChange: fn(),
	},
	argTypes: { groups: { control: false }, suggestions: { control: false } },
	render: args => <StatefulBar {...args} />,
} satisfies Meta<typeof LogSearchBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TwoGroups: Story = {};

export const NoGroups: Story = { args: { groups: [] } };
