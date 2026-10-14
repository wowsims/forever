import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { LOG_SUGGESTIONS } from '@testing/storybook/fixtures/results-views';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { LogSearchGroup, type LogSearchGroupProps } from './LogSearchGroup';

const StatefulGroup = ({ group: initial, onChange, ...props }: LogSearchGroupProps) => {
	const [group, setGroup] = useState(initial);
	return (
		<LogSearchGroup
			{...props}
			group={group}
			onChange={next => {
				setGroup(next);
				onChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Results/LogRunner/Components/LogSearchGroup',
	component: LogSearchGroup,
	tags: ['autodocs'],
	args: {
		group: { id: 0, field: 'spell', join: 'or', values: ['Mortal Strike'] },
		suggestions: LOG_SUGGESTIONS,
		onChange: fn(),
		onRemove: fn(),
	},
	argTypes: { group: { control: false }, suggestions: { control: false } },
	render: args => <StatefulGroup {...args} />,
} satisfies Meta<typeof LogSearchGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const PickedField: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: i18n.t('results_tab.details.logs.search_join.and') }));
		await expect(args.onChange).toHaveBeenCalledWith(expect.objectContaining({ join: 'and' }));
	},
};

export const TypedField: Story = { args: { group: { id: 0, field: 'amount', join: 'and', values: ['>5000'] } } };

export const Outcomes: Story = { args: { group: { id: 0, field: 'outcome', join: 'or', values: ['crit', 'critical-block'] } } };

export const Empty: Story = { args: { group: { id: 0, field: 'source', join: 'or', values: [] } } };
