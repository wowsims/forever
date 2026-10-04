import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export interface TableHeaderCellProps extends Omit<ComponentPropsWithoutRef<'th'>, 'className'> {
	/** Lets the label wrap; a header cell stays on one line otherwise. */
	wrap?: boolean;
	className?: ClassValue;
	testId?: string;
}

export const TableHeaderCell = ({ wrap, className, testId, children, ...rest }: TableHeaderCellProps) => (
	<th {...rest} className={clsx('ui-table-header-cell', className)} data-wrap={wrap ? '' : undefined} data-testid={testId}>
		{children}
	</th>
);
