import clsx from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface ButtonGroupProps extends ComponentPropsWithoutRef<'div'> {
	/** Pads every direct `Button` child as that `size`. */
	size?: 'sm' | 'xs';
	/** Spaces the buttons apart instead of joining their borders. */
	gap?: boolean;
	testId?: string;
}

export const ButtonGroup = forwardRef<HTMLDivElement, ButtonGroupProps>(({ size, gap = false, role = 'group', className, testId, children, ...props }, ref) => (
	<div ref={ref} role={role} data-size={size} data-gap={gap ? '' : undefined} data-testid={testId} className={clsx('ui-button-group', className)} {...props}>
		{children}
	</div>
));
ButtonGroup.displayName = 'ButtonGroup';
