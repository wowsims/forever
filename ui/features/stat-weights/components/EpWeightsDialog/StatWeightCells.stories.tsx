import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStatWeightsResult, STAT_WEIGHT_STATS } from '@testing/storybook/fixtures/stat-weights';
import { Table, TableRow } from '@ui-kit/Table';

import { StatWeightCells } from './StatWeightCells';

const RESULT = makeStatWeightsResult();

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/StatWeightCells',
	component: StatWeightCells,
	tags: ['autodocs'],
	args: {
		stat: STAT_WEIGHT_STATS[1],
		statWeights: RESULT.dps,
		iterations: 1000,
		epRatio: 1,
		epDelta: 0,
		cellClassName: 'ui-ep-weights-table-cell',
	},
	argTypes: {
		stat: { control: false },
		statWeights: { control: false },
	},
	decorators: [
		Story => (
			<Table>
				<tbody>
					<TableRow>
						<Story />
					</TableRow>
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof StatWeightCells>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const AboveCurrentEp: Story = { args: { epDelta: 0.4 } };
export const BelowCurrentEp: Story = { args: { epDelta: -0.4 } };
export const UnusedMetric: Story = { args: { epRatio: 0 } };
export const NotApplicable: Story = { args: { statWeights: undefined } };
