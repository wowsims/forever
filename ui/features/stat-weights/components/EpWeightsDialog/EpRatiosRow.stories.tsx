import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEpColumns } from '@testing/storybook/fixtures/stat-weights';
import { Table } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { fn } from 'storybook/test';

import { EpRatiosRow } from './EpRatiosRow';
import { StatsType } from './types';
import { EP_TOOLTIP_ID } from './utils';

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/EpRatiosRow',
	component: EpRatiosRow,
	tags: ['autodocs'],
	args: {
		columns: makeEpColumns(),
		epRatios: [1, 0, 0, 0, 0, 0],
		onEpRatioChange: fn(),
		onComputeEp: fn(),
		showThreatMetrics: false,
	},
	argTypes: { columns: { control: false } },
	decorators: [
		Story => (
			<>
				<Table data-stats-type={StatsType.Ep}>
					<thead>
						<Story />
					</thead>
				</Table>
				<Tooltip id={EP_TOOLTIP_ID} />
			</>
		),
	],
} satisfies Meta<typeof EpRatiosRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ThreatMetrics: Story = {
	args: {
		columns: makeEpColumns({ damage: true, healing: true, threat: true }),
		epRatios: [1, 0, 0.5, 0.25, 0, 0],
		showThreatMetrics: true,
	},
};
