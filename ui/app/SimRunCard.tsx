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
import clsx from 'clsx';
import type { ReactNode } from 'react';

import type { SimHostObject } from './individual_sim_ui';
import { NoticeNativeSim } from './NoticeNativeSim';
import { SidebarActions } from './SidebarActions';

export interface SimRunActionsProps {
	host: SimHostObject<any>;
	sim: Sim;
	className?: string;
}

/**
 * Stat weights, the actions features register and Simulate, on one row: in the run card, or in the dock below `lg`. Mounted
 * once, because `SimulateAction` holds whether its run is going; off while the sim loads.
 */
export const SimRunActions = ({ host, sim, className }: SimRunActionsProps) => {
	const { status } = useSimStatus(sim);
	return (
		<SidebarDisabledContext value={host.disabled || status === 'loading'}>
			<ButtonGroup gap className={clsx('items-center justify-end', className)} testId="sim-run-card-run-row">
				<StatWeightsAction />
				<SidebarActions registry={host.sidebar} />
				{/* Last although a registered action orders itself (`order-20`), and as wide as the row leaves it. */}
				{!host.simDisabled && (
					<div className="order-last min-w-24 flex-1">
						<SimulateAction />
					</div>
				)}
			</ButtonGroup>
		</SidebarDisabledContext>
	);
};

export interface SimRunCardProps {
	/** Null on the render that builds the containers it is constructed from. */
	host: SimHostObject<any> | null;
	sim: Sim;
	/** Under the iterations; the dock carries them itself below `lg`. */
	actions?: ReactNode;
}

export const SimRunCard = ({ host, sim, actions }: SimRunCardProps) => {
	const activateTab = useActivateTab();
	const runs = !host?.simDisabled;
	const { status } = useSimStatus(sim);

	return (
		<SidebarDisabledContext value={!!host?.disabled}>
			<RunCard testId="sim-run-card" loading={status === 'loading'} controls={runs && <IterationsPicker sim={sim} />} action={actions}>
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
