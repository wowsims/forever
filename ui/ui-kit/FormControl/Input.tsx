import { Input as BaseInput, type InputProps as BaseInputProps } from '@base-ui/react/input';
import { type ControlSize, useControlSize } from '@ui-kit/Field';
import clsx, { type ClassValue } from 'clsx';
import { forwardRef } from 'react';

export interface InputProps extends Omit<BaseInputProps, 'className' | 'size'> {
	className?: ClassValue;
	size?: ControlSize;
	/** The native `size` attribute, the box's width in characters; `size` is the control size. */
	htmlSize?: number;
}

export const Input = forwardRef<HTMLElement, InputProps>(({ className, size, htmlSize, ...props }, ref) => {
	const controlSize = useControlSize(size);
	return <BaseInput ref={ref} className={clsx('ui-input', className)} data-size={controlSize} size={htmlSize} {...props} />;
});
Input.displayName = 'Input';
