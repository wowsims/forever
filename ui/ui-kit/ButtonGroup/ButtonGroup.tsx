import clsx from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface ButtonGroupProps extends ComponentPropsWithoutRef<'div'> {
	/** Pads every direct `Button` child as that `size`. */
	size?: 'sm' | 'xs';
	/** Spaces the buttons apart instead of joining their borders. */
	gap?: boolean;
	/** `vertical` stacks the buttons as full-width rows with their content on the left. */
	orientation?: 'horizontal' | 'vertical';
	testId?: string;
}

export const ButtonGroup = forwardRef<HTMLDivElement, ButtonGroupProps>(
	({ size, gap = false, orientation = 'horizontal', role = 'group', className, testId, children, ...props }, ref) => (
		<div
			ref={ref}
			role={role}
			data-size={size}
			data-gap={gap ? '' : undefined}
			data-orientation={orientation === 'vertical' ? 'vertical' : undefined}
			data-testid={testId}
			className={clsx('ui-button-group', className)}
			{...props}>
			{children}
		</div>
	),
);
ButtonGroup.displayName = 'ButtonGroup';
