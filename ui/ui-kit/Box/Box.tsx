import type { ColorToken } from '@ui-kit/utils/color_tokens';
import clsx from 'clsx';
import type { ComponentPropsWithRef, CSSProperties } from 'react';

/** A colour token to draw the border in, or `none` for no border. */
export type BoxBorder = 'none' | ColorToken;
/** `xs` is the dense row padding (2 across, 1 down); `sm` and `md` are even all round. */
export type BoxPadding = 'none' | 'xs' | 'sm' | 'md';

export interface BoxProps extends ComponentPropsWithRef<'div'> {
	border?: BoxBorder;
	padding?: BoxPadding;
	testId?: string;
}

export const Box = ({ border = 'border', padding = 'sm', className, style, testId = 'box', children, ...props }: BoxProps) => (
	<div
		className={clsx('ui-box', className)}
		data-border={border}
		data-padding={padding}
		data-testid={testId}
		style={border === 'none' ? style : ({ '--box-border-color': `var(--color-${border})`, ...style } as CSSProperties)}
		{...props}>
		{children}
	</div>
);
