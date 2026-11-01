import clsx, { type ClassValue } from 'clsx';

export interface SpinnerProps {
	size?: 'xs' | 'sm' | 'md' | 'lg';
	/** `strong` draws a thicker ring with a foreground head, for a spinner that marks the running row in dense content. */
	tone?: 'default' | 'strong';
	/** The spinner's accessible name; without one it is decoration. */
	label?: string;
	className?: ClassValue;
	testId?: string;
}

export const Spinner = ({ size = 'md', tone = 'default', label, className, testId = 'loader' }: SpinnerProps) => (
	<div
		data-testid={testId}
		data-size={size}
		data-tone={tone === 'default' ? undefined : tone}
		role={label ? 'img' : undefined}
		aria-label={label}
		className={clsx('ui-spinner rounded-full', className)}
	/>
);
