import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { type DurationRow, DurationsView, type DurationsViewProps } from './Durations';

const StatefulDurations = (props: DurationsViewProps) => {
	const [rows, setRows] = useState<readonly DurationRow[]>(props.rows);
	return (
		<DurationsView
			{...props}
			rows={rows}
			onAdd={() => {
				const last = rows.at(-1) ?? props.encounter;
				setRows([...rows, { duration: last.duration + 60, durationVariation: last.durationVariation, duplicate: false }]);
				props.onAdd();
			}}
			onChange={(index, patch) => {
				setRows(rows.map((row, at) => (at === index ? { ...row, ...patch } : row)));
				props.onChange(index, patch);
			}}
			onRemove={index => {
				setRows(rows.filter((_, at) => at !== index));
				props.onRemove(index);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Bulk/BulkSettings/Durations',
	component: DurationsView,
	tags: ['autodocs'],
	render: args => <StatefulDurations {...args} />,
	args: {
		encounter: { duration: 180, durationVariation: 5 },
		rows: [],
		healthFight: false,
		onAdd: fn(),
		onChange: fn(),
		onRemove: fn(),
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DurationsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EncounterOnly: Story = {};

export const TwoExtras: Story = {
	args: {
		rows: [
			{ duration: 90, durationVariation: 10, duplicate: false },
			{ duration: 300, durationVariation: 30, duplicate: false },
		],
	},
};

export const DuplicateRow: Story = {
	args: { rows: [{ duration: 180, durationVariation: 5, duplicate: true }] },
};

export const HealthFight: Story = {
	args: { rows: [{ duration: 90, durationVariation: 10, duplicate: false }], healthFight: true },
};
