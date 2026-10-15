import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { type BulkCompareBar, BulkCompareChart, type BulkCompareChartProps } from './BulkCompareChart';

const CANVAS = { smoke: false, smokeReason: 'chart.js needs a canvas' };

const bar = (number: number, avg: number, gain = ''): BulkCompareBar => {
	const label = number ? `Set ${number}` : 'Current gear';
	return {
		key: number ? `set-${number}` : 'current',
		label,
		isCurrent: number === 0,
		avg,
		details: `${label}: ${avg.toFixed(2)} ±3 DPS${gain && `, ${gain} over current gear`}`,
	};
};

// Holds the picked-out set, so hovering a bar dims the others in every group as it does in the dialog.
const Interactive = (props: BulkCompareChartProps) => {
	const [highlighted, setHighlighted] = useState(props.highlighted);
	return (
		<BulkCompareChart
			{...props}
			highlighted={highlighted}
			onHighlight={key => {
				setHighlighted(key);
				props.onHighlight(key);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Bulk/BulkCompareChart',
	component: BulkCompareChart,
	tags: ['autodocs'],
	render: args => <Interactive {...args} />,
	args: {
		groups: [
			{
				label: 'Council',
				summary: '1:30 ±10s · 1 target',
				bars: [bar(1, 838.1, '+25.70 (3.16%)'), bar(2, 830.92, '+18.52 (2.28%)'), bar(0, 812.4), bar(3, 809.05, '-3.35 (0.41%)')],
			},
			{
				label: 'Current',
				summary: '3:00 ±5s · 1 target',
				bars: [bar(4, 801.02, '+24.72 (3.18%)'), bar(1, 797.41, '+21.11 (2.72%)'), bar(0, 776.3)],
			},
			{
				label: 'Patchwerk',
				summary: 'Health · 1 target',
				bars: [bar(5, 12031.25, '+271.25 (2.31%)'), bar(1, 11981.5, '+221.50 (1.88%)'), bar(0, 11760)],
			},
		],
		highlighted: null,
		onHighlight: fn(),
	},
	argTypes: { groups: { control: false } },
	parameters: CANVAS,
	decorators: [
		Story => (
			<div className="max-w-3xl">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkCompareChart>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeEncounters: Story = {};

export const SetPickedOut: Story = { args: { highlighted: 'set-1' } };
