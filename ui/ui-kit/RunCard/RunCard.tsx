import { Spinner } from '@ui-kit/Spinner';
import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface RunCardProps {
	/** The run's inputs, above the action: an iterations field, say. */
	controls?: ReactNode;
	/** The one button that starts (or stops) the run. */
	action?: ReactNode;
	/** What the run produced or will produce, under the action. */
	children?: ReactNode;
	/** Covers the whole card with a spinner and blocks it, while what it drives is not ready yet. */
	loading?: boolean;
	className?: string;
	testId?: string;
}

export const RunCard = ({ controls, action, children, loading = false, className, testId = 'run-card' }: RunCardProps) => (
	<section className={clsx('relative flex w-full flex-col gap-4', className)} data-testid={testId} aria-busy={loading || undefined}>
		{(controls || action) && (
			<div className="flex flex-col gap-2" data-testid="run-card-controls">
				{controls}
				{action}
			</div>
		)}
		{children && (
			<div className={clsx('flex flex-col gap-4 empty:hidden')} data-testid="run-card-body">
				{children}
			</div>
		)}
		{loading && (
			<div className="absolute inset-0 z-1 flex items-center justify-center bg-background/60" data-testid="run-card-loading">
				<Spinner size="md" />
			</div>
		)}
	</section>
);
