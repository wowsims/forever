import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export interface TableProps extends Omit<ComponentPropsWithoutRef<'table'>, 'className'> {
	/** `true` shades every second body row; `solid` paints odd and even body rows in the two table row colours. */
	striped?: boolean | 'solid';
	/** Highlights the body row under the pointer. */
	hover?: boolean;
	/** Halves the cell padding. */
	dense?: boolean;
	/** `plain` draws no row or column rules and keeps the cell padding; `rows` draws the row rules alone. */
	variant?: 'default' | 'plain' | 'rows';
	/** `sm` sets the header and cell text to the small content size; the default inherits the surrounding text. */
	size?: 'default' | 'sm';
	/** The first and last cell of every row drop their outer padding, so the row's content runs edge to edge. */
	flush?: boolean;
	className?: ClassValue;
	testId?: string;
}

const flag = (on: boolean | undefined) => (on ? '' : undefined);

export const Table = ({ striped, hover, dense, variant, size, flush, className, testId = 'table', children, ...rest }: TableProps) => (
	<table
		{...rest}
		className={clsx('ui-table', className)}
		data-striped={striped === 'solid' ? 'solid' : flag(striped)}
		data-hover={flag(hover)}
		data-dense={flag(dense)}
		data-size={size === 'sm' ? 'sm' : undefined}
		data-flush={flag(flush)}
		data-variant={variant === 'default' ? undefined : variant}
		data-testid={testId}>
		{children}
	</table>
);
