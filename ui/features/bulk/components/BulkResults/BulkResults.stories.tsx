import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { BulkDurationPills, type BulkDurationPillsProps } from './BulkResults';

const StatefulPills = (props: BulkDurationPillsProps) => {
	const [selected, setSelected] = useState(props.selected);
	return (
		<BulkDurationPills
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
	title: 'Features/Bulk/BulkResults/BulkDurationPills',
	component: BulkDurationPills,
	tags: ['autodocs'],
	render: args => <StatefulPills {...args} />,
	args: {
		entries: [
			{ duration: 180, durationVariation: 5, isEncounter: true },
			{ duration: 90, durationVariation: 10, isEncounter: false },
			{ duration: 300, durationVariation: 30, isEncounter: false },
		],
		selected: 0,
		onSelect: fn(),
	},
} satisfies Meta<typeof BulkDurationPills>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeDurations: Story = {};

export const ManyDurationsWrap: Story = {
	args: {
		entries: [180, 45, 90, 120, 300, 480, 600].map((duration, index) => ({ duration, durationVariation: 10, isEncounter: index === 0 })),
	},
};
