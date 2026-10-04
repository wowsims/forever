import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeEpColumns } from '@testing/storybook/fixtures/stat-weights';
import { Table } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';

import { EpWeightsHeader } from './EpWeightsHeader';
import { StatsType } from './types';
import { EP_TOOLTIP_ID } from './utils';

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/EpWeightsHeader',
	component: EpWeightsHeader,
	tags: ['autodocs'],
	args: {
		columns: makeEpColumns(),
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
} satisfies Meta<typeof EpWeightsHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ThreatMetrics: Story = {
	args: { columns: makeEpColumns({ damage: true, healing: true, threat: true }), showThreatMetrics: true },
};

export const GearPlanner: Story = {
	args: { columns: makeEpColumns({ damage: false, healing: false, threat: false }), showUpdateColumn: false },
};
