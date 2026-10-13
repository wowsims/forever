import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { type Ref, useImperativeHandle, useState } from 'react';

export interface ProgressStepListProps {
	labels: ReadonlyArray<string>;
	/** 0-based step being run; the ones before it are done. */
	running: number;
}

// Every row and its status slot are there from the start, so the progress dialog keeps its height through the run.
export const ProgressStepList = ({ labels, running }: ProgressStepListProps) => (
	<ul className="mt-2 mb-0 flex list-none flex-col items-center gap-1 p-0" data-testid="progress-step-list">
		{labels.map((label, index) => {
			const state = index < running ? 'done' : index === running ? 'running' : 'pending';
			return (
				<li
					key={index}
					className="flex items-center gap-2"
					data-state={state}
					aria-current={state === 'running' ? 'step' : undefined}
					data-testid="progress-step">
					{/* Balances the status slot, so the label sits on the dialog's centre line. */}
					<span className="w-4" aria-hidden />
					{label}
					<span className="flex w-4 items-center justify-center">
						{state === 'done' && <Icon name="circle-check" className="text-success" />}
						{state === 'running' && <Spinner size="xs" testId="progress-step-spinner" />}
					</span>
				</li>
			);
		})}
	</ul>
);

export interface ProgressStepListHandle {
	setRunning: (running: number) => void;
}

/** Takes the running step through `ref`, so a progress tick renders this leaf and never the dialog. */
export const HandleProgressStepList = ({ labels, ref }: Pick<ProgressStepListProps, 'labels'> & { ref?: Ref<ProgressStepListHandle> }) => {
	const [running, setRunning] = useState(0);
	useImperativeHandle(ref, () => ({ setRunning }), []);
	return <ProgressStepList labels={labels} running={running} />;
};
