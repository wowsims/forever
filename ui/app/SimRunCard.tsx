import { RESULTS_TAB_ID, useActivateTab } from '@features/hooks/useActivateTab';
import { IterationsPicker } from '@features/results/components/IterationsPicker';
import { SimResultsPanel } from '@features/results/components/SimResultsPanel';
import { SimulateAction } from '@features/results/components/SimulateAction';
import { StatWeightsAction } from '@features/stat-weights/components/StatWeightsAction';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import type { Sim } from '@sim/sim';
import { ButtonGroup } from '@ui-kit/ButtonGroup';
import { RunCard } from '@ui-kit/RunCard';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';

import type { SimHostObject } from './individual_sim_ui';
import { NoticeNativeSim } from './NoticeNativeSim';
import { SidebarActions } from './SidebarActions';

export interface SimRunCardProps {
	/** Null on the render that builds the containers it is constructed from. */
	host: SimHostObject<any> | null;
	sim: Sim;
}

export const SimRunCard = ({ host, sim }: SimRunCardProps) => {
	const activateTab = useActivateTab();
	const runs = !host?.simDisabled;
	const { status } = useSimStatus(sim);

	return (
		<SidebarDisabledContext value={!!host?.disabled}>
			<RunCard
				testId="sim-run-card"
				loading={status === 'loading'}
				controls={runs && <IterationsPicker sim={sim} />}
				action={
					host && (
						<>
							<ButtonGroup gap className="w-full justify-end" testId="sim-run-card-run-row">
								{runs && <SimulateAction />}
								<StatWeightsAction />
							</ButtonGroup>
							<div className="flex justify-end gap-2 empty:hidden" data-testid="sim-run-card-actions">
								<SidebarActions registry={host.sidebar} />
							</div>
						</>
					)
				}>
				{host && (
					<>
						<SimResultsPanel
							panel={host.resultsPanel}
							results={host.raidSimResultsManager}
							onOpenBreakdown={runs ? () => activateTab(RESULTS_TAB_ID) : undefined}
						/>
						<NoticeNativeSim />
					</>
				)}
			</RunCard>
		</SidebarDisabledContext>
	);
};
