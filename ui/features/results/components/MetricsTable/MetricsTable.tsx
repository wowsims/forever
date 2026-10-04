import type { SortDirection } from '@tanstack/react-table';
import { Button } from '@ui-kit/Button';
import { Table, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import { type MetricsColumnDef, useMetricsTable } from '../../hooks/useMetricsTable';
import type { MetricRow } from '../../model/grouping';
import { MetricsTableRow } from './MetricsTableRow';

const ariaSort = (direction: false | SortDirection) => (direction === 'asc' ? 'ascending' : direction === 'desc' ? 'descending' : 'none');

export interface MetricsTableProps<T> {
	/** The table's `data-testid`, which the gates address it by. */
	rootTestId: string;
	columns: Array<MetricsColumnDef<T>>;
	rows: Array<MetricRow<T>>;
	sortColumnId: string;
	/** Whether a run has completed. A table with no rows hides only once one has — never at load. */
	hasResult: boolean;
	rowThreatOnly?: (metric: T) => boolean;
}

export const MetricsTable = <T,>({ rootTestId, columns, rows, sortColumnId, hasResult, rowThreatOnly }: MetricsTableProps<T>) => {
	const table = useMetricsTable({ columns, rows, sortColumnId });

	if (hasResult && !rows.length) return null;

	return (
		<Table striped hover className="ui-metrics-table" testId={rootTestId}>
			<thead data-testid="metrics-table-header">
				{table.getHeaderGroups().map(headerGroup => (
					<TableRow testId="metrics-table-header-row" key={headerGroup.id}>
						{headerGroup.headers.map(header => (
							<TableHeaderCell
								key={header.id}
								testId="metrics-table-header-cell"
								data-primary-metric={header.column.columnDef.meta?.primaryColumn ? '' : undefined}
								className={[
									'ui-metrics-header-cell',
									'p-0',
									header.column.columnDef.meta?.columnClass,
									header.column.columnDef.meta?.headerCellClass,
								]}
								aria-sort={ariaSort(header.column.getIsSorted())}
								{...tooltipAnchorProps(header.column.columnDef.meta?.headerTooltipId, header.column.columnDef.meta?.headerTooltip)}>
								<Button
									data-testid="metrics-table-sort"
									variant="unstyled"
									className="flex h-full w-full items-center justify-center px-2 py-1 text-inherit [font:inherit] focus-visible:focus-ring"
									onClick={header.column.getToggleSortingHandler()}>
									<span>
										<table.FlexRender header={header} />
									</span>
								</Button>
							</TableHeaderCell>
						))}
					</TableRow>
				))}
			</thead>
			<tbody data-testid="metrics-table-body">
				{table.getRowModel().rows.map(row => (
					<MetricsTableRow key={row.id} row={row} rowThreatOnly={rowThreatOnly} />
				))}
			</tbody>
		</Table>
	);
};
