import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { BulkCompareButtonView } from '../BulkCompareButton';
import { BulkResultsCsvButtonView } from '../BulkResultsCsvButton';
import { BulkEncounterResultsView, type BulkEncounterResultsViewProps } from './BulkEncounterResults';

const StatefulResults = (props: BulkEncounterResultsViewProps) => {
	const [selected, setSelected] = useState(props.selected);
	return (
		<BulkEncounterResultsView
			{...props}
			selected={selected}
			onSelect={index => {
				setSelected(index);
				props.onSelect(index);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Bulk/BulkEncounterResults',
	component: BulkEncounterResultsView,
	tags: ['autodocs'],
	render: args => <StatefulResults {...args} />,
	args: {
		rows: [
			{ label: 'Current', durationLabel: '3:00 ±5s', bestDps: '788.23', gain: '+13.45 (1.75%)' },
			{ label: 'Patchwerk', durationLabel: '5:00 ±30s', bestDps: '798.10', gain: '+30.20 (3.93%)' },
			{ label: 'Council', durationLabel: '4:00 ±10s', bestDps: '1,104.90', gain: null },
		],
		selected: 0,
		onSelect: fn(),
		footer: (
			<div className="flex flex-wrap gap-2">
				<BulkResultsCsvButtonView onExport={fn()} />
				<BulkCompareButtonView onOpen={fn()} />
			</div>
		),
	},
	argTypes: { footer: { control: false } },
	decorators: [
		Story => (
			<div className="flex w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkEncounterResultsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeEncounters: Story = {};

export const LongNames: Story = {
	args: {
		rows: [
			{ label: 'Current', durationLabel: '3:00 ±5s', bestDps: '788.23', gain: '+13.45 (1.75%)' },
			{ label: 'Ragnaros, execute phase with adds', durationLabel: 'Health', bestDps: '702.00', gain: '+12.80 (1.86%)' },
		],
	},
};
