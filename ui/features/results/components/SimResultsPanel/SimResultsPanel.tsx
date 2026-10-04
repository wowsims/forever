import type { SimResultsManager } from '@features/results/model/results_manager';
import { useSimHost } from '@sim/context/SimHostContext';
import { useSyncExternalStore } from 'react';

import { GearPlannerNotice } from './GearPlannerNotice';
import { ResultsPanelStage, type ResultsPanelStore } from './results_panel_store';
import { SimProgress } from './SimProgress';
import { SimResultSummary } from './SimResultSummary';
import { UnlaunchedNotice } from './UnlaunchedNotice';

export interface SimResultsPanelProps {
	panel: ResultsPanelStore;
	results: SimResultsManager | null;
	onOpenBreakdown?: () => void;
}

export const SimResultsPanel = ({ panel, results, onOpenBreakdown }: SimResultsPanelProps) => {
	const host = useSimHost();
	const stage = useSyncExternalStore(panel.subscribe, panel.getStage);

	const summary = stage === ResultsPanelStage.Result ? results : null;
	const showProgress = stage === ResultsPanelStage.Pending || stage === ResultsPanelStage.Running;
	if (!summary && !showProgress && !host.simDisabled) return null;

	return (
		<div className="flex flex-col gap-2" data-testid="results-viewer">
			{summary && <SimResultSummary results={summary} onOpenBreakdown={onOpenBreakdown} />}
			{showProgress && <SimProgress panel={panel} />}
			{host.disabled && <UnlaunchedNotice />}
			{host.simDisabled && !host.disabled && <GearPlannerNotice />}
		</div>
	);
};
