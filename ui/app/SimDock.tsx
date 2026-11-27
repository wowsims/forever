import { SimResultReadout } from '@features/results/components/SimResultsPanel';
import i18n from '@i18n/config';
import type { PlayerSpec } from '@sim/player/player_spec';
import type { Sim } from '@sim/sim';
import { Button } from '@ui-kit/Button';
import { Drawer } from '@ui-kit/Drawer';
import { Icon } from '@ui-kit/Icon';
import { SidebarIconAction } from '@ui-kit/SidebarIconAction';
import clsx from 'clsx';
import { type ComponentPropsWithRef, useCallback, useEffect, useRef, useState } from 'react';

import type { SimHostObject } from './individual_sim_ui';
import { SpecReleaseInfo } from './ReleaseInfo';
import { SimPresetsDialog } from './SimPresetsDialog';
import { SimRunActions } from './SimRunCard';
import { SimSidebarBody } from './SimSidebarBody';
import { SimTabsStrip, useSimSide } from './SimTabs';
import { useUpgradesSlotRefs } from './UpgradesSlots';

export interface SimDockProps {
	/** Null on the render that builds the containers it is constructed from. */
	host: SimHostObject<any> | null;
	sim: Sim;
	spec: PlayerSpec<any>;
	simReady: boolean;
}

const SHEET_CLASS = 'bottom-sim-dock z-sidebar max-h-4/5 shadow-popup';

// A side's bar, hidden while the other side is open; its sheet closes with it.
const useDockSheet = (hidden: boolean) => {
	const bar = useRef<HTMLDivElement>(null);
	const [open, setOpen] = useState(false);
	useEffect(() => {
		if (hidden) setOpen(false);
	}, [hidden]);
	const barProps = { ref: bar, className: clsx('flex h-sim-dock-bar items-center gap-2 px-sidebar', hidden && 'hidden') };
	const sheetProps = {
		side: 'bottom' as const,
		open,
		onOpenChange: setOpen,
		modal: false,
		ignoreOutsidePress: (event: Event) => !!bar.current?.contains(event.target as Node),
		className: SHEET_CLASS,
	};
	return { open, setOpen, barProps, sheetProps };
};

// Base UI's trigger props (the press handler, the ref, the expanded state) pass through to the button.
const SheetToggle = ({ open, testId, children, ...trigger }: ComponentPropsWithRef<'button'> & { open: boolean; testId: string }) => (
	<Button {...trigger} variant="unstyled" className="flex h-full min-w-0 flex-1 items-center gap-2" data-testid={testId}>
		<Icon name={open ? 'chevron-down' : 'chevron-up'} size="xs" className="shrink-0 text-link" />
		{children}
	</Button>
);

// The single sim's result and actions; its sheet holds the rest of the sidebar.
const SingleDockBar = ({ host, sim, spec, simReady, hidden }: SimDockProps & { host: SimHostObject<any>; hidden: boolean }) => {
	const { open, barProps, sheetProps } = useDockSheet(hidden);
	const [presetsOpen, setPresetsOpen] = useState(false);
	return (
		<div {...barProps} data-testid="sim-dock-single-bar">
			<Drawer
				{...sheetProps}
				testId="sim-dock-sheet"
				trigger={
					<SheetToggle open={open} testId="sim-dock-toggle">
						<SimResultReadout panel={host.resultsPanel} results={host.raidSimResultsManager} />
					</SheetToggle>
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
		</div>
	);
};

// The Upgrades run's count and button; its sheet holds the run panel. The Upgrades tab body fills all three slots, and
// the sheet stays mounted while closed so the panel that portals the bar's button stays mounted with it.
const UpgradesDockBar = ({ hidden }: { hidden: boolean }) => {
	const slots = useUpgradesSlotRefs();
	const { open, setOpen, barProps, sheetProps } = useDockSheet(hidden);
	// The run button is portalled in from the tab body, so React carries its clicks up that tree, not this one; a DOM
	// listener hears them. A run's results fill the page behind the sheet, so starting one closes it.
	const action = useCallback(
		(element: HTMLElement | null) => {
			slots.action(element);
			element?.addEventListener('click', () => setOpen(false), true);
		},
		[slots, setOpen],
	);
	return (
		<div {...barProps} data-testid="sim-dock-upgrades-bar">
			<Drawer
				{...sheetProps}
				keepMounted
				testId="sim-dock-upgrades-sheet"
				trigger={
					<SheetToggle open={open} testId="sim-dock-upgrades-toggle">
						{/* Portalled content too: without pointer events a press on the count lands on the toggle. */}
						<span ref={slots.summary} className="pointer-events-none min-w-0 truncate" />
					</SheetToggle>
				}>
				<div ref={slots.panel} className="@container flex flex-col" />
			</Drawer>
			<div ref={action} className="shrink-0" />
		</div>
	);
};

/**
 * The sidebar below `lg`: a bar pinned to the bottom of the screen with the open side's run and the tabs, and a sheet
 * over the page with the rest of that side's sidebar. Both sides' bars stay mounted and the other one hides, as in the
 * aside. A sheet stays open through a press on its bar, so a run started from the bar fills the results it shows.
 */
export const SimDock = ({ host, sim, spec, simReady }: SimDockProps) => {
	const side = useSimSide();
	return (
		<div className="fixed inset-x-0 bottom-0 z-sim-dock flex h-sim-dock flex-col border-t border-border bg-background" data-testid="sim-dock">
			{host ? (
				<>
					<SingleDockBar host={host} sim={sim} spec={spec} simReady={simReady} hidden={side !== 'single'} />
					<UpgradesDockBar hidden={side !== 'upgrades'} />
				</>
			) : (
				<div className="h-sim-dock-bar" />
			)}
			<SimTabsStrip placement="bottom" className="h-sim-dock-nav border-t border-border" />
		</div>
	);
};
