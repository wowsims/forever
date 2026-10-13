import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { type EncounterRow, EncountersView, type EncountersViewProps } from './Encounters';

const StatefulEncounters = (props: EncountersViewProps) => {
	const [presets, setPresets] = useState(props.presets);
	const [saved, setSaved] = useState(props.saved);
	const toggle = (rows: readonly EncounterRow[], name: string, checked: boolean) => rows.map(row => (row.name === name ? { ...row, checked } : row));
	return (
		<EncountersView
			{...props}
			presets={presets}
			saved={saved}
			onCheckedChange={(preset, name, checked) => {
				if (preset) setPresets(toggle(presets, name, checked));
				else setSaved(toggle(saved, name, checked));
				props.onCheckedChange(preset, name, checked);
			}}
		/>
	);
};

const row = (name: string, summary: string, checked = false, duplicate = false): EncounterRow => ({ name, summary, checked, duplicate });

const meta = {
	title: 'Features/Bulk/BulkSettings/Encounters',
	component: EncountersView,
	tags: ['autodocs'],
	render: args => <StatefulEncounters {...args} />,
	args: {
		currentSummary: '3:00 ±5s · 1 target',
		presets: [row('Patchwerk', '5:00 ±30s · 1 target', true)],
		saved: [row('Council', '4:00 ±10s · 3 targets', true), row('Short fight', '3:00 ±5s · 1 target', true, true), row('Execute test', 'Health · 1 target')],
		onCheckedChange: fn(),
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof EncountersView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithDuplicate: Story = {};

export const SavedOnly: Story = { args: { presets: [], saved: [row('Council', '4:00 ±10s · 3 targets')] } };

export const Empty: Story = { args: { presets: [], saved: [] } };
