import type { PlayerSpec } from '@sim/player/player_spec';
import type { Sim } from '@sim/sim';
import clsx from 'clsx';

import { SimTitleDropdown } from './header/SimTitleDropdown';
import type { SimHostObject } from './individual_sim_ui';
import { SpecReleaseInfo } from './ReleaseInfo';
import { SimRunActions } from './SimRunCard';
import { SimSidebarBody } from './SimSidebarBody';
import { useSimSide } from './SimTabs';
import { useUpgradesSlotRefs } from './UpgradesSlots';

export interface SimSidebarProps {
	/** Null on the render that builds the containers it is constructed from. */
	host: SimHostObject<any> | null;
	sim: Sim;
	spec: PlayerSpec<any>;
	simReady: boolean;
}

// The single sim's run card, warnings and stats.
const SingleSidebar = ({ host, sim, simReady, hidden }: Omit<SimSidebarProps, 'spec'> & { hidden: boolean }) => (
	<div
		className={clsx('@container flex-1 flex-col overflow-y-auto p-sidebar [&>*:not(:last-child)]:mb-4', hidden ? 'hidden' : 'flex')}
		data-testid="sim-sidebar-content">
		<SimSidebarBody host={host} sim={sim} simReady={simReady} actions={host && <SimRunActions host={host} sim={sim} wrap className="w-full" />} />
	</div>
);

// The Upgrades run panel's slot; the Upgrades tab body fills it.
const UpgradesSidebar = ({ hidden }: { hidden: boolean }) => (
	<div
		ref={useUpgradesSlotRefs().panel}
		className={clsx('flex-1 flex-col overflow-y-auto p-sidebar', hidden ? 'hidden' : 'flex')}
		data-testid="sim-sidebar-upgrades"
	/>
);

/** The aside from `lg` up. Both sides stay mounted and the other one hides, so flipping back keeps each side's state. */
export const SimSidebar = ({ host, sim, spec, simReady }: SimSidebarProps) => {
	const side = useSimSide();
	return (
		<aside className="sticky -top-px z-sidebar flex h-dvh max-w-100 flex-1 flex-col items-stretch bg-background" data-testid="sim-sidebar">
			<SimTitleDropdown currentSpec={spec} />
			<SingleSidebar host={host} sim={sim} simReady={simReady} hidden={side !== 'single'} />
			<UpgradesSidebar hidden={side !== 'upgrades'} />
			{host && <SpecReleaseInfo spec={spec} className="border-t border-border px-sidebar py-2 text-xs text-muted" />}
		</aside>
	);
};
