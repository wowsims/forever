import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { type Ref, useImperativeHandle, useState } from 'react';

export interface StatSweepStatListProps {
	labels: ReadonlyArray<string>;
	/** 0-based stat being run; the ones before it are done. */
	running: number;
}

// Every row and its status slot are there from the start, so the progress dialog keeps its height through the run.
export const StatSweepStatList = ({ labels, running }: StatSweepStatListProps) => (
	<ul className="mt-2 mb-0 flex list-none flex-col items-center gap-1 p-0" data-testid="stat-sweep-stat-list">
		{labels.map((label, index) => {
			const state = index < running ? 'done' : index === running ? 'running' : 'pending';
			return (
				<li
					key={label}
					className="flex items-center gap-2"
					data-state={state}
					aria-current={state === 'running' ? 'step' : undefined}
					data-testid="stat-sweep-stat">
					{/* Balances the status slot, so the label sits on the dialog's centre line. */}
					<span className="w-4" aria-hidden />
					{label}
					<span className="flex w-4 items-center justify-center">
						{state === 'done' && <Icon name="circle-check" className="text-success" />}
						{state === 'running' && <Spinner size="xs" testId="stat-sweep-stat-spinner" />}
					</span>
				</li>
			);
		})}
	</ul>
);

export interface StatSweepStatListHandle {
	setRunning: (running: number) => void;
}

/** Takes the running stat through `ref`, so a progress tick renders this leaf and never the dialog. */
export const HandleStatSweepStatList = ({ labels, ref }: Pick<StatSweepStatListProps, 'labels'> & { ref?: Ref<StatSweepStatListHandle> }) => {
	const [running, setRunning] = useState(0);
	useImperativeHandle(ref, () => ({ setRunning }), []);
	return <StatSweepStatList labels={labels} running={running} />;
};
