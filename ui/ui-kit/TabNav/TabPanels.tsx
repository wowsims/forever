import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface TabPanelsProps {
	className?: ClassValue;
	testId?: string;
	children?: ReactNode;
	[key: string]: unknown;
}

export const TabPanels = ({ className, testId, children, ...rest }: TabPanelsProps) => (
	<div className={clsx('grid grid-cols-1 items-start pt-section', className)} data-testid={testId} {...rest}>
		{children}
	</div>
);
