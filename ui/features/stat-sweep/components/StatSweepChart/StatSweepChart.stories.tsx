import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeCritSweepPoints, makeDefenseSweepPoints, makeHitSweepPoints, sweepSeries } from '@testing/storybook/fixtures/stat-sweep';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { StatSweepChart, type StatSweepChartProps } from './StatSweepChart';

const CANVAS = { smoke: false, smokeReason: 'chart.js needs a canvas' };

// Holds the picked metric and spread, so the controls redraw the chart as they do in the dialog.
const Interactive = (props: StatSweepChartProps) => {
	const [metric, setMetric] = useState(props.metric);
	const [showSpread, setShowSpread] = useState(props.showSpread);
	return (
		<StatSweepChart
			{...props}
			metric={metric}
			onMetricChange={next => {
				setMetric(next);
				props.onMetricChange(next);
			}}
			showSpread={showSpread}
			onShowSpreadChange={next => {
				setShowSpread(next);
				props.onShowSpreadChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/StatSweep/StatSweepChart',
	component: StatSweepChart,
	tags: ['autodocs'],
	args: { series: [], metric: 'dps', metrics: ['dps', 'tps'], onMetricChange: fn(), showSpread: true, onShowSpreadChange: fn() },
	argTypes: { series: { control: false } },
	render: args => <Interactive {...args} />,
	decorators: [
		Story => (
			<div className="max-w-3xl">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof StatSweepChart>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const SingleRun: Story = {
	args: { series: [sweepSeries('current', 'Hit Rating', 120, makeHitSweepPoints(120))] },
	parameters: CANVAS,
};

export const SpreadOff: Story = {
	args: { ...SingleRun.args, showSpread: false },
	parameters: CANVAS,
};

export const Threat: Story = {
	args: { ...SingleRun.args, metric: 'tps' },
	parameters: CANVAS,
};

export const Overlay: Story = {
	args: {
		series: [
			sweepSeries('current', 'Hit Rating', 120, makeHitSweepPoints(120)),
			sweepSeries('older', 'Hit Rating', 90, makeHitSweepPoints(90, 10_000), 1, 'Hit Rating ±100 (old gear)'),
		],
	},
	parameters: CANVAS,
};

export const OverlayOtherStat: Story = {
	args: {
		series: [sweepSeries('hit', 'Hit Rating', 120, makeHitSweepPoints(120)), sweepSeries('crit', 'Crit Rating', 210, makeCritSweepPoints(210), 1)],
	},
	parameters: CANVAS,
};

export const TankDamageTaken: Story = {
	args: {
		series: [sweepSeries('defense', 'Defense Rating', 300, makeDefenseSweepPoints(300))],
		metric: 'dtps',
		metrics: ['dps', 'tps', 'dtps', 'tmi'],
	},
	parameters: CANVAS,
};

export const TankThreat: Story = {
	args: { ...TankDamageTaken.args, metric: 'tps' },
	parameters: CANVAS,
};
