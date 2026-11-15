import { SimResultReadout } from '@features/results/components/SimResultsPanel';
import i18n from '@i18n/config';
import type { PlayerSpec } from '@sim/player/player_spec';
import type { Sim } from '@sim/sim';
import { Button } from '@ui-kit/Button';
import { Drawer } from '@ui-kit/Drawer';
import { SidebarIconAction } from '@ui-kit/SidebarIconAction';
import { useRef, useState } from 'react';

import type { SimHostObject } from './individual_sim_ui';
import { SpecReleaseInfo } from './ReleaseInfo';
import { SimPresetsDialog } from './SimPresetsDialog';
import { SimRunActions } from './SimRunCard';
import { SimSidebarBody } from './SimSidebarBody';
import { SimTabsStrip } from './SimTabs';

export interface SimDockProps {
	/** Null on the render that builds the containers it is constructed from. */
	host: SimHostObject<any> | null;
	sim: Sim;
	spec: PlayerSpec<any>;
	simReady: boolean;
}

/**
 * The sidebar below `lg`: a bar pinned to the bottom of the screen with the run's result and actions and the tabs, and a
 * sheet over the page with the rest of the sidebar. The sheet stays open through a press on the bar, so a run started
 * from the bar fills the results it shows.
 */
export const SimDock = ({ host, sim, spec, simReady }: SimDockProps) => {
	const bar = useRef<HTMLDivElement>(null);
	const [open, setOpen] = useState(false);
	const [presetsOpen, setPresetsOpen] = useState(false);

	return (
		<div className="fixed inset-x-0 bottom-0 z-sidebar flex h-sim-dock flex-col border-t border-border bg-background" data-testid="sim-dock">
			<div ref={bar} className="flex h-sim-dock-bar items-center gap-2 px-sidebar">
				{host && (
					<>
						<Drawer
							side="bottom"
							open={open}
							onOpenChange={setOpen}
							modal={false}
							ignoreOutsidePress={event => !!bar.current?.contains(event.target as Node)}
							className="bottom-sim-dock z-sidebar max-h-4/5 shadow-popup"
							testId="sim-dock-sheet"
							trigger={
								<Button variant="unstyled" className="flex h-full min-w-0 flex-1 items-center" data-testid="sim-dock-toggle">
									<SimResultReadout panel={host.resultsPanel} results={host.raidSimResultsManager} />
								</Button>
							}>
							<div className="@container flex flex-col gap-4">
								<SimSidebarBody host={host} sim={sim} simReady={simReady} />
								<SpecReleaseInfo spec={spec} className="border-t border-border pt-2 text-xs text-muted" />
							</div>
						</Drawer>
						<SidebarIconAction
							icon="bookmark"
							label={i18n.t('sidebar.buttons.presets.title')}
							testId="sim-presets-action"
							onClick={() => setPresetsOpen(true)}
						/>
						<SimPresetsDialog open={presetsOpen} onOpenChange={setPresetsOpen} />
						<SimRunActions host={host} sim={sim} className="shrink-0" />
					</>
				)}
			</div>
			<SimTabsStrip placement="bottom" className="h-sim-dock-nav border-t border-border" />
		</div>
	);
};
