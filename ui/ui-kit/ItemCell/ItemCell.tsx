import { Box, type BoxBorder, type BoxPadding } from '@ui-kit/Box';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface ItemCellProps {
	border?: BoxBorder;
	padding?: BoxPadding;
	icon: ReactNode;
	ilvl?: ReactNode;
	name: ReactNode;
	labels?: ReactNode;
	action?: ReactNode;
	/** Where the name and labels sit: `end` right-aligns them from `md` up, and below `md` they read from the start. */
	align?: 'start' | 'end';
	className?: ClassValue;
	testId?: string;
	rootDataAttributes?: Record<string, string>;
}

export const ItemCell = ({
	border = 'none',
	padding = 'none',
	icon,
	ilvl,
	name,
	labels,
	action,
	align,
	className,
	testId,
	rootDataAttributes,
}: ItemCellProps) => (
	<Box border={border} padding={padding} className={clsx('ui-item-picker-root', className)} testId={testId ?? 'item-picker-root'} {...rootDataAttributes}>
		<div className="ui-item-picker-icon-wrapper" data-testid="item-picker-icon-wrapper">
			{ilvl !== undefined && (
				<span className="ui-item-picker-ilvl" data-testid="item-picker-ilvl">
					{ilvl}
				</span>
			)}
			{icon}
		</div>
		<div className="ui-item-picker-labels-container" data-align={align} data-testid="item-picker-labels-container">
			<div className="ui-item-picker-name-row flex gap-1 tracking-normal" data-testid="item-picker-name-row">
				{name}
			</div>
			{labels}
		</div>
		{action}
	</Box>
);
