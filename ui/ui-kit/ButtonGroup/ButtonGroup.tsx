import { Button } from '@ui-kit/Button';
import type { ControlSize } from '@ui-kit/Field';
import clsx from 'clsx';
import { Children, cloneElement, type ComponentPropsWithoutRef, forwardRef, isValidElement } from 'react';

export interface ButtonGroupProps extends ComponentPropsWithoutRef<'div'> {
	/** Sizes every direct `Button` child as that `size`, over the size the button sets itself. */
	size?: ControlSize;
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
			{size === undefined
				? children
				: Children.map(children, child =>
						isValidElement<{ size?: ControlSize }>(child) && child.type === Button ? cloneElement(child, { size }) : child,
					)}
		</div>
	),
);
ButtonGroup.displayName = 'ButtonGroup';
