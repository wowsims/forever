import clsx, { type ClassValue } from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface IconRowProps extends Omit<ComponentPropsWithoutRef<'div'>, 'className'> {
	className?: ClassValue;
	testId?: string;
}

export const IconRow = forwardRef<HTMLDivElement, IconRowProps>(({ className, testId, ...props }, ref) => (
	<div ref={ref} className={clsx('ui-icon-row', className)} data-testid={testId} {...props} />
));
IconRow.displayName = 'IconRow';
