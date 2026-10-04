import type { ActionMetrics } from '@sim/proto/sim_result';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { type AttackRowsFixture, loadAttackRows } from '@testing/storybook/fixtures/results-views';
import { Table } from '@ui-kit/Table';

import { createMetricsColumnHelper, useMetricsTable } from '../../hooks/useMetricsTable';
import type { MetricRow } from '../../model/grouping';
import { attackMetricsColumns } from '../AttackMetricsColumns';
import { MetricsTableRow, type MetricsTableRowProps } from './MetricsTableRow';

const helper = createMetricsColumnHelper<ActionMetrics>();
const COLUMNS = helper.columns([attackMetricsColumns.name(), attackMetricsColumns.casts()]);

interface RowsProps extends Pick<MetricsTableRowProps<ActionMetrics>, 'rowThreatOnly'> {
	rows: Array<MetricRow<ActionMetrics>>;
}

// The row takes a table-model row, which only `useMetricsTable` builds.
const Rows = ({ rows, rowThreatOnly }: RowsProps) => {
	const table = useMetricsTable({ columns: COLUMNS, rows, sortColumnId: 'casts' });
	return (
		<Table striped hover className="ui-metrics-table">
			<tbody>
				{table.getRowModel().rows.map(row => (
					<MetricsTableRow key={row.id} row={row} rowThreatOnly={rowThreatOnly} />
				))}
			</tbody>
		</Table>
	);
};

const meta = {
	title: 'Features/Results/MetricsTable/Components/MetricsTableRow',
	component: Rows,
	tags: ['autodocs'],
	args: { rows: [] },
	argTypes: { rows: { control: false } },
	loaders: [loadAttackRows],
	render: (args, { loaded }) => <Rows {...args} rows={(loaded as AttackRowsFixture).rows} />,
} satisfies Meta<typeof Rows>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ParentAndChildren: Story = {};

export const ThreatOnlyMarked: Story = { args: { rowThreatOnly: () => true } };
