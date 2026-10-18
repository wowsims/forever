import { flexRender, type Row } from '@tanstack/react-table';
import { FakeLink } from '@ui-kit/FakeLink';
import { TableCell, TableRow } from '@ui-kit/Table';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import type { MetricsTableFeatures } from '../../hooks/useMetricsTable';
import type { MetricRow } from '../../model/grouping';

// The Name cell's indent by row depth: a spell, its ranks, a rank's tags.
const CHILD_INDENT = [undefined, 'pl-4', 'pl-8'];

export interface MetricsTableRowProps<T> {
	row: Row<MetricsTableFeatures, MetricRow<T>>;
	rowThreatOnly?: (metric: T) => boolean;
}

export const MetricsTableRow = <T,>({ row, rowThreatOnly }: MetricsTableRowProps<T>) => {
	const isParent = row.getCanExpand();
	const toggle = isParent ? row.getToggleExpandedHandler() : undefined;
	return (
		<TableRow
			testId={rowThreatOnly?.(row.original.metric) ? 'threat-metrics' : undefined}
			data-parent={isParent ? '' : undefined}
			data-child={row.depth > 0 ? '' : undefined}
			data-expanded={isParent && row.getIsExpanded() ? '' : undefined}>
			{row.getAllCells().map((cell, index) => {
				const tooltipId = cell.column.columnDef.meta?.tooltipId;
				return (
					// The model value, which the display string may round or abbreviate away.
					<TableCell
						key={cell.id}
						data-primary-metric={cell.column.columnDef.meta?.primaryColumn ? '' : undefined}
						className={[
							'ui-metrics-cell',
							toggle && 'relative',
							index === 0 && CHILD_INDENT[row.depth],
							cell.column.columnDef.meta?.columnClass,
							cell.column.columnDef.meta?.dataClass,
						]}
						data-text={String(cell.getValue())}
						{...tooltipAnchorProps(tooltipId)}
						data-row-id={tooltipId ? row.id : undefined}>
						{flexRender(cell.column.columnDef.cell, cell.getContext())}
						{toggle && <FakeLink onActivate={toggle} />}
					</TableCell>
				);
			})}
		</TableRow>
	);
};
