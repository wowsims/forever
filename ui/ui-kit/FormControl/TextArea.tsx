import { type ControlSize, useControlSize } from '@ui-kit/Field';
import clsx, { type ClassValue } from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface TextAreaProps extends Omit<ComponentPropsWithoutRef<'textarea'>, 'className'> {
	className?: ClassValue;
	size?: ControlSize;
}

export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(({ className, size, ...props }, ref) => {
	const controlSize = useControlSize(size);
	return <textarea ref={ref} className={clsx('ui-input', className)} data-size={controlSize} {...props} />;
});
TextArea.displayName = 'TextArea';
