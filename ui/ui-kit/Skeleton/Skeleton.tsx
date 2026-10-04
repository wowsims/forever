import clsx, { type ClassValue } from 'clsx';

export interface SkeletonProps {
	className?: ClassValue;
	testId?: string;
}

export const Skeleton = ({ className, testId = 'skeleton' }: SkeletonProps) => <span className={clsx('ui-skeleton', className)} data-testid={testId} />;
