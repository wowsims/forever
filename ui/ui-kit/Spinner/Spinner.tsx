import clsx, { type ClassValue } from 'clsx';

export interface SpinnerProps {
	size?: 'xs' | 'sm' | 'md' | 'lg';
	className?: ClassValue;
	testId?: string;
}

export const Spinner = ({ size = 'md', className, testId = 'loader' }: SpinnerProps) => (
	<div data-testid={testId} data-size={size} className={clsx('ui-spinner rounded-full', className)} />
);
