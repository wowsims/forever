import { Input as BaseInput, type InputProps as BaseInputProps } from '@base-ui/react/input';
import clsx, { type ClassValue } from 'clsx';
import { forwardRef } from 'react';

/** `md` shares the button's control scale, so the input sits flush with a default-size button beside it. */
export type InputSize = 'sm' | 'md';

export interface InputProps extends Omit<BaseInputProps, 'className'> {
	className?: ClassValue;
	inputSize?: InputSize;
}

export const Input = forwardRef<HTMLElement, InputProps>(({ className, inputSize = 'sm', ...props }, ref) => (
	<BaseInput ref={ref} className={clsx('ui-input', className)} data-size={inputSize} {...props} />
));
Input.displayName = 'Input';
