import type { Meta, StoryObj } from '@storybook/react-vite';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { CharacterStatsHeaderView, type CharacterStatsHeaderViewProps, CharacterStatsPanelView, type CharacterStatsPanelViewProps } from './CharacterStats';

const ROWS = [
	['Strength', '320'],
	['Agility', '180'],
	['Stamina', '410'],
	['Attack Power', '1,240'],
	['Melee Crit', '24.50%'],
];

const SampleTable = () => (
	<Table variant="plain" flush>
		<tbody>
			{ROWS.map(([label, value]) => (
				<TableRow key={label} className="ui-character-stats-row">
					<TableCell className="ui-character-stats-label">{label}</TableCell>
					<TableCell className="ui-character-stats-value">{value}</TableCell>
				</TableRow>
			))}
		</tbody>
	</Table>
);

const StatefulPanel = ({ mode: initialMode, onModeChange, ...props }: CharacterStatsPanelViewProps) => {
	const [mode, setMode] = useState(initialMode);
	return (
		<CharacterStatsPanelView
			{...props}
			mode={mode}
			onModeChange={next => {
				setMode(next);
				onModeChange(next);
			}}>
			<SampleTable />
		</CharacterStatsPanelView>
	);
};

const StatefulHeader = ({ mode: initial, onModeChange, ...props }: CharacterStatsHeaderViewProps) => {
	const [mode, setMode] = useState(initial);
	return (
		<CharacterStatsHeaderView
			{...props}
			mode={mode}
			onModeChange={next => {
				setMode(next);
				onModeChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/CharacterStats/CharacterStats',
	component: CharacterStatsPanelView,
	subcomponents: { CharacterStatsHeaderView },
	tags: ['autodocs'],
	args: {
		mode: 'key',
		onModeChange: fn(),
		labels: { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' },
	},
	argTypes: { mode: { control: 'inline-radio', options: ['key', 'all'] }, children: { control: false } },
	render: args => <StatefulPanel {...args} />,
	decorators: [
		Story => (
			<div className="flex w-62 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof CharacterStatsPanelView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'All' }));
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
	},
};

export const HeaderKey: Story = {
	render: args => <StatefulHeader {...args} />,
	play: async ({ args, canvasElement }) => {
		const all = within(canvasElement).getByRole('button', { name: 'All' });
		await userEvent.click(all);
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
		await expect(all).toHaveAttribute('aria-pressed', 'true');
	},
};

export const HeaderAll: Story = { render: args => <StatefulHeader {...args} />, args: { mode: 'all' } };
