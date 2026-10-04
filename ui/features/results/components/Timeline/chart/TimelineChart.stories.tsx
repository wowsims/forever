import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, dpsLog, HEROIC_STRIKE } from '@testing/storybook/fixtures/results-views';

import type { TimelineChartSpec } from '../../../model/timeline/chart/types';
import { DPS_SERIES_ID, dpsColor, dpsScale, timeScale, Y_DPS } from './series';
import { TimelineChart } from './TimelineChart';

const DURATION = 30;

const dpsSpec = (): TimelineChartSpec => {
	const points = Array.from({ length: DURATION + 1 }, (_, second) => {
		const log = dpsLog(second, [damageLog(second, 900 + second * 40), damageLog(second, 300, { actionId: HEROIC_STRIKE })]);
		return { x: second, y: log.dps, log };
	});
	return {
		datasets: [
			{
				label: 'DPS',
				seriesId: DPS_SERIES_ID,
				tooltipSpec: { kind: 'dps' },
				yAxisID: Y_DPS,
				borderColor: dpsColor(),
				borderWidth: 2,
				pointRadius: 0,
				data: points,
			},
		],
		scales: { x: timeScale(DURATION, 'Time (s)'), [Y_DPS]: dpsScale(Math.max(...points.map(point => point.y))) },
		annotations: null,
		duration: DURATION,
	};
};

const meta = {
	title: 'Features/Results/Timeline/Components/TimelineChart',
	component: TimelineChart,
	tags: ['autodocs'],
	args: { spec: null },
	argTypes: { spec: { control: false } },
	decorators: [
		Story => (
			<div className="h-128">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TimelineChart>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WaitingForData: Story = {};

export const Dps: Story = {
	args: { spec: dpsSpec() },
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
};
