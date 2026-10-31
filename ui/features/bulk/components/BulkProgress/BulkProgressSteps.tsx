import i18n from '@i18n/config';
import { Icon } from '@ui-kit/Icon';
import { ProgressTrackerBar } from '@ui-kit/ProgressTrackerDialog';
import { Spinner } from '@ui-kit/Spinner';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { type Ref, useImperativeHandle, useState } from 'react';

import type { BulkBuildStep } from '../../model/builds/builds';

export interface BulkProgressStepsProps {
	rows: readonly BulkBuildStep[];
	/** 0-based step being run; the ones before it are done. */
	running: number;
	/** How far the running step is, 0 to 1. */
	progress: number;
}

const COLUMNS = ['encounter', 'settings', 'talents', 'rotation'] as const;

// Every row and its status cell are there from the start, so the progress dialog keeps its height through the run. The
// wrapper scrolls sideways when the cells cannot wrap to the dialog's width. Each row's own progress is a hairline laid
// over its collapsed bottom border, half of which lies in the row, so it takes none of the cells' padding; a pending
// row shows the empty track in place of the separator. A pending row fades its step cells only, so the track keeps its
// colour, and fades them in when the row starts.
export const BulkProgressSteps = ({ rows, running, progress }: BulkProgressStepsProps) => (
	<div className="w-full overflow-clip text-left">
		<Table dense flush size="sm" variant="rows" testId="bulk-progress-steps">
			<thead>
				<TableRow>
					{COLUMNS.map(column => (
						<TableHeaderCell key={column} scope="col" className="text-left">
							{i18n.t(`bulk_tab.builds.parts.${column}`)}
						</TableHeaderCell>
					))}
					<TableHeaderCell scope="col">
						<span className="sr-only">{i18n.t('bulk_tab.progress.status')}</span>
					</TableHeaderCell>
				</TableRow>
			</thead>
			<tbody>
				{rows.map((row, index) => {
					const state = index < running ? 'done' : index === running ? 'running' : 'pending';
					return (
						<TableRow
							key={index}
							className="relative"
							data-state={state}
							aria-current={state === 'running' ? 'step' : undefined}
							testId="progress-step">
							{COLUMNS.map(column => (
								<TableCell
									key={column}
									className={['transition-opacity duration-200 ease-out motion-reduce:transition-none', state === 'pending' && 'opacity-50']}>
									{row[column]}
									{column === 'encounter' && (
										<span className="block text-xs font-normal text-muted" data-testid="progress-step-encounter-summary">
											{row.encounterSummary}
										</span>
									)}
								</TableCell>
							))}
							<TableCell className="w-6 text-right">
								{state === 'done' && <Icon name="circle-check" className="text-success" title={i18n.t('bulk_tab.progress.done')} />}
								{state === 'running' && (
									<Spinner
										size="xs"
										tone="strong"
										className="ml-auto"
										label={i18n.t('bulk_tab.progress.running')}
										testId="progress-step-spinner"
									/>
								)}
								{state === 'pending' && <span className="sr-only">{i18n.t('bulk_tab.progress.pending')}</span>}
								<ProgressTrackerBar
									running
									variant="line"
									progress={state === 'done' ? 1 : state === 'running' ? progress : 0}
									label={i18n.t('bulk_tab.progress.step_progress', { encounter: row.encounter, talents: row.talents })}
									className="absolute inset-x-0 -bottom-px"
									testId="progress-step-line"
								/>
							</TableCell>
						</TableRow>
					);
				})}
			</tbody>
		</Table>
	</div>
);

export interface BulkProgressStepsHandle {
	setSteps: (steps: Pick<BulkProgressStepsProps, 'running' | 'progress'>) => void;
}

/** Takes the running step and its progress through `ref`, so a progress tick renders this leaf and never the dialog. */
export const HandleBulkProgressSteps = ({ rows, ref }: Pick<BulkProgressStepsProps, 'rows'> & { ref?: Ref<BulkProgressStepsHandle> }) => {
	const [{ running, progress }, setSteps] = useState<Pick<BulkProgressStepsProps, 'running' | 'progress'>>({ running: 0, progress: 0 });
	useImperativeHandle(ref, () => ({ setSteps }), []);
	return <BulkProgressSteps rows={rows} running={running} progress={progress} />;
};
