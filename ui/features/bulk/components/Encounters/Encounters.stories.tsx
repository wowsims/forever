import { BulkEncounterRef } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { type EncounterRow, EncountersView, type EncountersViewProps } from './Encounters';

const StatefulEncounters = (props: EncountersViewProps) => {
	const [presets, setPresets] = useState(props.presets);
	const [saved, setSaved] = useState(props.saved);
	const toggle = (rows: readonly EncounterRow[], ref: BulkEncounterRef, checked: boolean) => rows.map(row => (row.ref === ref ? { ...row, checked } : row));
	return (
		<EncountersView
			{...props}
			presets={presets}
			saved={saved}
			onCheckedChange={(ref, checked) => {
				if (ref.preset) setPresets(toggle(presets, ref, checked));
				else setSaved(toggle(saved, ref, checked));
				props.onCheckedChange(ref, checked);
			}}
		/>
	);
};

const rowOf =
	(preset: boolean) =>
	(name: string, summary: string, checked = false, duplicate = false): EncounterRow => ({
		ref: BulkEncounterRef.create({ name, preset }),
		name,
		summary,
		checked,
		duplicate,
	});
const presetRow = rowOf(true);
const savedRow = rowOf(false);

const meta = {
	title: 'Features/Bulk/Encounters',
	component: EncountersView,
	tags: ['autodocs'],
	render: args => <StatefulEncounters {...args} />,
	args: {
		currentSummary: '3:00 ±5s · 1 target',
		presets: [presetRow('Patchwerk', '5:00 ±30s · 1 target', true)],
		saved: [
			savedRow('Council', '4:00 ±10s · 3 targets', true),
			savedRow('Council, saved twice', '4:00 ±10s · 3 targets', true, true),
			savedRow('Execute test', 'Health · 1 target'),
		],
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

export const SavedOnly: Story = { args: { presets: [], saved: [savedRow('Council', '4:00 ±10s · 3 targets')] } };

export const Empty: Story = { args: { presets: [], saved: [] } };
