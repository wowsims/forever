import clsx, { type ClassValue } from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface FieldGridProps extends Omit<ComponentPropsWithoutRef<'div'>, 'className'> {
	className?: ClassValue;
	testId?: string;
}

export const FieldGrid = forwardRef<HTMLDivElement, FieldGridProps>(({ className, testId, ...props }, ref) => (
	<div ref={ref} className={clsx('ui-field-grid', className)} data-testid={testId} {...props} />
));
FieldGrid.displayName = 'FieldGrid';
