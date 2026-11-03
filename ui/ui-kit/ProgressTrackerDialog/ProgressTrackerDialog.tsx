import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import { type ReactNode, type Ref, useState } from 'react';

import { ElapsedTime } from './ElapsedTime';
import { HandleProgressBar } from './HandleProgressBar';
import type { ProgressTrackerHandle, ProgressTrackerState } from './types';

export interface ProgressTrackerDialogProps {
	open: boolean;
	title: string;
	state: ProgressTrackerState;
	className?: string;
	warning?: ReactNode;
	hasProgressBar?: boolean;
	/** Rendered right under the progress bar, above the elapsed time. */
	belowBar?: ReactNode;
	/** Rendered after the elapsed time, on its line. */
	afterElapsed?: ReactNode;
	onCancel?: () => void;
	ref?: Ref<ProgressTrackerHandle>;
	testId?: string;
}

export const ProgressTrackerDialog = ({
	open,
	title,
	state,
	className,
	warning,
	hasProgressBar,
	belowBar,
	afterElapsed,
	onCancel,
	ref,
	testId,
}: ProgressTrackerDialogProps) => {
	const [run, setRun] = useState(() => ({ open, startedAt: Date.now() }));
	if (run.open !== open) setRun({ open, startedAt: open ? Date.now() : run.startedAt });

	return (
		<Dialog
			open={open}
			onOpenChange={() => {}}
			className={clsx('max-h-modal-inset-h', className)}
			bodyClassName="min-h-0 overflow-hidden"
			size="md"
			verticalAlign="center"
			title={title}
			preventClose
			keepMounted
			elevated
			testId={testId ?? 'progress-tracker-dialog'}>
			<div
				className="flex min-h-0 flex-1 flex-col items-center gap-4 text-center text-sm"
				data-testid="progress-tracker-modal-content"
				data-stage={state.stage}>
				{warning && (
					<div className="border border-warning p-4 text-sm" data-testid="progress-tracker-modal-warning">
						{warning}
					</div>
				)}
				{hasProgressBar && <HandleProgressBar running={open} ref={ref} />}
				{/* Title, bar and Cancel stay in view; a long step list or message scrolls here. */}
				<div className="flex min-h-0 w-full flex-col items-center gap-4 overflow-y-auto" data-testid="progress-tracker-modal-scroll">
					{belowBar}
					<div
						className={!state.message ? 'hidden' : undefined}
						data-testid="progress-tracker-modal-message"
						data-stage={state.stage}
						hidden={!state.message}>
						{state.message}
					</div>
				</div>
				<div className="shrink-0">
					<strong>{i18n.t('common.elapsed_time')}:</strong> <ElapsedTime running={open} startedAt={run.startedAt} />
					{afterElapsed}
				</div>
				{onCancel && (
					<Button variant="outline-cancel" data-testid="progress-tracker-modal-cancel-btn" onClick={onCancel}>
						<Icon name="ban" style="base" />
						{i18n.t('sidebar.results.reference.cancel')}
					</Button>
				)}
			</div>
		</Dialog>
	);
};
