import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import type { ProgressTrackerHandle } from '@ui-kit/ProgressTrackerDialog';
import { ProgressTrackerDialog } from '@ui-kit/ProgressTrackerDialog';
import { useEffect, useRef, useState } from 'react';

import { bulkProgress, cancelBulkBatch, subscribeBulkProgress } from '../../model/run/run';
import { BulkIterations } from './BulkIterations';
import { type BulkProgressStepsHandle, HandleBulkProgressSteps } from './BulkProgressSteps';
import { BulkTimeLeft } from './BulkTimeLeft';

/** One element, so an iterations tick commits only the leaf. */
const ITERATIONS = <BulkIterations />;

/** Mounted only while a batch is in flight, the way the reforge tracker is: at rest there is nothing here. */
export const BulkProgressDialog = () => {
	const host = useSimHost();

	const [state, setState] = useState({ stage: 'preparing', hasIterations: false });
	// The run emits its first progress before it is marked running, so the encounter list is complete at mount.
	const [rows] = useState(() => bulkProgress(host.player)?.steps.rows ?? []);
	const barRef = useRef<ProgressTrackerHandle>(null);
	const stepsRef = useRef<BulkProgressStepsHandle>(null);

	useEffect(
		() =>
			subscribeBulkProgress(host.player, progress => {
				const hasIterations = rows.length <= 1 && !!progress.iterations;
				setState(previous =>
					previous.stage === progress.stage && previous.hasIterations === hasIterations ? previous : { stage: progress.stage, hasIterations },
				);
				// A batch of one step, which a Gear-mode batch always is, has no finished step to estimate from, so its bar
				// shows the stage's own time left; a longer batch shows its estimate beside the elapsed time.
				barRef.current?.setProgress({
					title: progress.title,
					current: progress.current,
					total: progress.total,
					secondsRemaining: rows.length > 1 ? undefined : (progress.secondsRemaining ?? null),
				});
				stepsRef.current?.setSteps(progress.steps);
			}),
		[host, rows],
	);

	return (
		<ProgressTrackerDialog
			open
			testId="bulk-sim-progress-tracker"
			title={i18n.t('bulk_tab.progress.title')}
			state={{ stage: state.stage, message: state.hasIterations ? ITERATIONS : undefined }}
			hasProgressBar
			belowBar={rows.length > 1 && <HandleBulkProgressSteps ref={stepsRef} rows={rows} />}
			afterElapsed={<BulkTimeLeft />}
			onCancel={() => void cancelBulkBatch(host)}
			ref={barRef}
		/>
	);
};
