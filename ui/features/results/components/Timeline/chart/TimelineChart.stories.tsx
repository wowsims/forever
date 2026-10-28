import { AddWave, DesignDamageAmp, EncounterDesign } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, dpsLog, HEROIC_STRIKE } from '@testing/storybook/fixtures/results-views';

import type { TimelineChartSpec } from '../../../model/timeline/chart/types';
import { scheduledMarkers } from './scheduled';
import { DPS_SERIES_ID, dpsColor, dpsScale, timeScale, Y_DPS } from './series';
import { TimelineChart } from './TimelineChart';

const DURATION = 30;

const dpsSpec = (withDesign = false): TimelineChartSpec => {
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
		annotations: withDesign
			? {
					bands: [],
					icons: [],
					scheduled: scheduledMarkers(
						{
							...EncounterDesign.create(),
							downtimes: [{ name: '', targetIndex: 0, start: { at: 0.3, jitter: 0 }, end: { at: 0.4, jitter: 0 }, clearDebuffs: true }],
							addWaves: [
								{
									...AddWave.create(),
									count: 2,
									spawn: { at: 0.1, jitter: 0 },
									lifetime: 0.1,
									repeat: { interval: 0.4, count: 1, jitter: 0 },
								},
							],
							damageAmps: [{ ...DesignDamageAmp.create(), percent: 20, start: { at: 0.6, jitter: 0 }, end: { at: 0.8, jitter: 0 } }],
							moves: [{ name: '', at: { at: 0.5, jitter: 0 }, yards: 8, seconds: 0, repeat: { interval: 0, count: 0, jitter: 0 } }],
						},
						DURATION,
					),
				}
			: null,
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

export const DpsWithEncounterDesign: Story = {
	args: { spec: dpsSpec(true) },
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
};
