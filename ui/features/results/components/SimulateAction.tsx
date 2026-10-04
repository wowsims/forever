import type { ProgressMetrics } from '@generated/proto/api';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { SimRunKind } from '@sim/state/sim_store';
import { Icon } from '@ui-kit/Icon';
import { SidebarActionButton } from '@ui-kit/SidebarActionButton';
import { useRef, useState } from 'react';
import { flushSync } from 'react-dom';

import { trackEvent } from '../../../tracking/utils';

export const SimulateAction = () => {
	const host = useSimHost();
	const runs = host.sim.runs;
	const [running, setRunning] = useState(false);
	const [stopping, setStopping] = useState(false);
	// Not `runs.isAborting`: that tracks the kind's run, which the facade clears as the run settles,
	// while this tracks whether *this click's* abort is still in flight. Substituting one for the
	// other leaves the button on Stop after the next run completes.
	const waitAbort = useRef(false);

	const abort = async () => {
		if (waitAbort.current) return;
		try {
			waitAbort.current = true;
			await runs.abort(SimRunKind.IndividualSim);
		} catch (error) {
			console.error('Error on sim abort!');
			console.error(error);
		} finally {
			waitAbort.current = false;
			setStopping(false);
			if (!runs.isRunning(SimRunKind.IndividualSim)) setRunning(false);
		}
	};

	const simulate = async () => {
		trackEvent({
			action: 'sim',
			category: 'simulate',
			label: 'simulate',
			value: host.sim.getIterations(),
		});
		if (runs.isRunning(SimRunKind.IndividualSim)) return;
		setRunning(true);

		host.resultsViewer.addAbortButton(() => void abort());

		await host.runIndividualSim((progress: ProgressMetrics) => {
			if (progress.finalRaidResult?.error) {
				host.resultsViewer.hideAll();
				return;
			}
			host.resultsViewer.setProgress(progress);
		});

		host.resultsViewer.removeAbortButton();
		if (!waitAbort.current) setRunning(false);
	};

	if (running) {
		return (
			<SidebarActionButton
				testId="sim-stop-action"
				variant="cancel"
				disabled={stopping}
				onClick={() => {
					// Synchronous: the run action reads the button back in the click's own task.
					flushSync(() => setStopping(true));
					void abort();
				}}>
				<Icon name="xmark" style="base" />
				{stopping ? i18n.t('sidebar.results.stopping') : i18n.t('sidebar.results.stop')}
			</SidebarActionButton>
		);
	}

	return (
		<SidebarActionButton testId="dps-action" onClick={() => void simulate()}>
			{i18n.t('sidebar.buttons.simulate')}
		</SidebarActionButton>
	);
};
