import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { BulkEncounterTabs, type BulkEncounterTabsProps } from './BulkResults';

const StatefulTabs = (props: BulkEncounterTabsProps) => {
	const [selected, setSelected] = useState(props.selected);
	return (
		<BulkEncounterTabs
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
	title: 'Features/Bulk/BulkResults/BulkEncounterTabs',
	component: BulkEncounterTabs,
	tags: ['autodocs'],
	render: args => <StatefulTabs {...args} />,
	args: {
		entries: [
			{ name: '', isCurrent: true, durationLabel: '3:00 ±5s' },
			{ name: 'Patchwerk', isCurrent: false, durationLabel: '5:00 ±30s' },
			{ name: 'Council', isCurrent: false, durationLabel: '4:00 ±10s' },
		],
		selected: 0,
		onSelect: fn(),
	},
} satisfies Meta<typeof BulkEncounterTabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeEncounters: Story = {};

export const ManyEncountersWrap: Story = {
	args: {
		entries: [
			{ name: '', isCurrent: true, durationLabel: '3:00 ±5s' },
			...['Patchwerk', 'Council', 'Execute test', 'Long fight', 'Short fight', 'Cleave pack'].map((name, index) => ({
				name,
				isCurrent: false,
				durationLabel: index === 2 ? 'Health' : `${index + 2}:00 ±10s`,
			})),
		],
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
};
