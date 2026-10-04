import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface SummaryTableRowProps {
	children?: ReactNode;
	className?: string;
	testId?: string;
}

/** A flex line, not a `<tr>` (that is `TableRow`): its children in a row, the last one pushed to the right edge. */
export const SummaryTableRow = ({ children, className, testId = 'summary-table-row' }: SummaryTableRowProps) => (
	<div className={clsx('flex items-center [&_*:last-child]:ml-auto', className)} data-testid={testId}>
		{children}
	</div>
);
