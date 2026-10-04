import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStatWeightsResult, STAT_WEIGHT_STATS } from '@testing/storybook/fixtures/stat-weights';
import { Table } from '@ui-kit/Table';
import { fn } from 'storybook/test';

import { StatWeightRowView } from './StatWeightRow';
import { StatsType } from './types';

const RESULT = makeStatWeightsResult();
const [STRENGTH, AGILITY] = STAT_WEIGHT_STATS;

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/StatWeightRow',
	component: StatWeightRowView,
	tags: ['autodocs'],
	args: {
		stat: AGILITY,
		fullName: 'Agility',
		shortName: 'Agi',
		metrics: [{ ratioIndex: 0, epRatio: 1 }],
		iterations: 1000,
		epDelta: 0,
		currentEp: 0.79,
		onCurrentEpChange: fn(),
		include: { value: true, disabled: false, onChange: fn() },
		showThreatMetrics: false,
	},
	argTypes: {
		stat: { control: false },
		metrics: { control: false },
	},
	decorators: [
		Story => (
			<Table data-stats-type={StatsType.Ep}>
				<tbody>
					<Story />
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof StatWeightRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Before a calculation every metric cell reads N/A. */
export const Default: Story = {};

/** After a calculation the EP cell is coloured by how far it sits from the current EP weight. */
export const Calculated: Story = {
	args: { metrics: [{ ratioIndex: 0, epRatio: 1, statWeights: RESULT.dps }], epDelta: 0.5, currentEp: 0.29 },
};

export const ExcludedFromCalculation: Story = {
	args: { include: { value: false, disabled: false, onChange: fn() } },
};

/** The reference stat is always part of the calculation, so its toggle is locked. */
export const ReferenceStat: Story = {
	args: {
		stat: STRENGTH,
		fullName: 'Strength',
		shortName: 'Str',
		currentEp: 1,
		metrics: [{ ratioIndex: 0, epRatio: 1, statWeights: RESULT.dps }],
		include: { value: true, disabled: true, onChange: fn() },
	},
};

export const OutsideTheEpSet: Story = { args: { include: undefined } };

export const GearPlanner: Story = { args: { metrics: [], include: undefined, showUpdateColumn: false } };
