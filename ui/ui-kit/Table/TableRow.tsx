import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export interface TableRowProps extends Omit<ComponentPropsWithoutRef<'tr'>, 'className'> {
	className?: ClassValue;
	testId?: string;
}

export const TableRow = ({ className, testId, children, ...rest }: TableRowProps) => (
	<tr {...rest} className={clsx('ui-table-row', className)} data-testid={testId}>
		{children}
	</tr>
);
