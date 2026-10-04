import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export interface TableCellProps extends Omit<ComponentPropsWithoutRef<'td'>, 'className'> {
	className?: ClassValue;
	testId?: string;
}

export const TableCell = ({ className, testId, children, ...rest }: TableCellProps) => (
	<td {...rest} className={clsx('ui-table-cell', className)} data-testid={testId}>
		{children}
	</td>
);
