import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import type { ProgressStepListHandle, ProgressTrackerHandle } from '@ui-kit/ProgressTrackerDialog';
import { HandleProgressStepList, ProgressTrackerDialog } from '@ui-kit/ProgressTrackerDialog';
import { useEffect, useRef, useState } from 'react';

import { bulkProgress, cancelBulkBatch, subscribeBulkProgress } from '../../model/run';
import { BulkProgressMessage } from './BulkProgressMessage';

/** One element, so the dialog's own state changes only when the stage does. */
const MESSAGE = <BulkProgressMessage />;

/** Mounted only while a batch is in flight, the way the reforge tracker is: at rest there is nothing here. */
export const BulkProgressDialog = () => {
	const host = useSimHost();

	const [state, setState] = useState({ stage: 'preparing', hasMessage: false });
	// The run emits its first progress before it is marked running, so the encounter list is complete at mount.
	const [labels] = useState(() => bulkProgress(host.player)?.steps.labels ?? []);
	const barRef = useRef<ProgressTrackerHandle>(null);
	const stepsRef = useRef<ProgressStepListHandle>(null);

	useEffect(
		() =>
			subscribeBulkProgress(host.player, progress => {
				const hasMessage = progress.secondsRemaining !== undefined || !!progress.iterations;
				setState(previous =>
					previous.stage === progress.stage && previous.hasMessage === hasMessage ? previous : { stage: progress.stage, hasMessage },
				);
				barRef.current?.setProgress({ title: progress.title, current: progress.current, total: progress.total });
				stepsRef.current?.setRunning(progress.steps.running);
			}),
		[host],
	);

	return (
		<ProgressTrackerDialog
			open
			testId="bulk-sim-progress-tracker"
			title={i18n.t('bulk_tab.progress.title')}
			state={{ stage: state.stage, message: state.hasMessage ? MESSAGE : undefined }}
			hasProgressBar
			belowBar={labels.length > 1 && <HandleProgressStepList ref={stepsRef} labels={labels} />}
			onCancel={() => void cancelBulkBatch(host)}
			ref={barRef}
		/>
	);
};
