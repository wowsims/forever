import { useSimStatus } from '@sim/hooks/useSimStatus';
import type { PlayerSpec } from '@sim/player/player_spec';
import type { Sim } from '@sim/sim';
import { StickyHeaderContext } from '@ui-kit/hooks/useStickyToolbar';
import { useWideLayout } from '@ui-kit/hooks/useWideLayout';
import { ToastArea, toastManager } from '@ui-kit/Toast';
import clsx from 'clsx';
import { type ReactNode, type RefObject, useEffect, useLayoutEffect, useRef, useState } from 'react';

import { ForeverNotice } from './ForeverNotice';
import { SimTitleDropdown } from './header/SimTitleDropdown';
import { SimToolbar } from './header/SimToolbar';
import type { SimHostObject } from './individual_sim_ui';
import { simUiAttributes, simUiClasses } from './shell_classes';
import { SimDock } from './SimDock';
import { SimImportExport } from './SimImportExport';
import { SimOnboarding } from './SimOnboarding';
import { SimSidebar } from './SimSidebar';
import { SimSideOnly, SimTabsPanels, SimTabsStrip } from './SimTabs';
import { SimTabsSection } from './SimTabsSection';
import type { ShellDom } from './types/shell_dom';
import { UpgradesSlotsProvider } from './UpgradesSlots';

export interface SimShellProps {
	domRef: RefObject<ShellDom | null>;
	/** Null on the render that builds the containers it is constructed from, so everything reaching through it waits a render. */
	host: SimHostObject<any> | null;
	sim: Sim;
	className: string;
	spec: PlayerSpec<any>;
	noticeText?: string;
	knownIssues: ReadonlyArray<ReactNode>;
	onOpenSettings: () => void;
}

export const SimShell = ({ domRef, host, sim, className, spec, noticeText, knownIssues, onOpenSettings }: SimShellProps) => {
	const root = useRef<HTMLDivElement>(null);
	const header = useRef<HTMLElement>(null);
	const [headerEl, setHeaderEl] = useState<HTMLElement | null>(null);
	const simReady = useSimStatus(sim).status === 'ready';
	// Below `lg` the sidebar becomes the dock and the tabs move into it. The two never mount together: the run's actions
	// keep their own state, and the tab strip registers its tabs with the one `Tabs.Root`.
	const compact = !useWideLayout();

	const [stuck, setStuck] = useState(false);
	useEffect(() => {
		const element = header.current;
		if (!element) return;
		// One delivery can carry several records, oldest first, so the last is the current state — reading `[entry]` leaves the bar stuck on a stale ratio.
		const observer = new IntersectionObserver(entries => setStuck(entries[entries.length - 1].intersectionRatio < 1), { threshold: [1] });
		observer.observe(element);
		return () => observer.disconnect();
	}, []);

	// A child's layout effect runs before its parent's, which is what lets `SimApp` construct against a populated bundle in the very same commit.
	useLayoutEffect(() => {
		domRef.current = { root: root.current! };
		setHeaderEl(header.current);
	}, [domRef]);

	return (
		<StickyHeaderContext value={headerEl}>
			<UpgradesSlotsProvider>
				<div
					ref={root}
					className={clsx(
						simUiClasses({ className, spec }),
						"max-h-sim-dock-inset-h overflow-y-auto before:fixed before:top-0 before:left-0 before:-z-1 before:h-screen before:w-screen before:bg-sim before:bg-cover before:bg-no-repeat before:content-[''] max-lg:overflow-x-hidden",
					)}
					data-testid="sim-ui"
					{...simUiAttributes({ spec })}>
					<div className="flex h-full min-h-sim-dock-inset-h flex-col" data-testid="sim-root">
						<ForeverNotice host={host} />
						{noticeText ? (
							<div className="relative mb-0 w-full border border-transparent bg-overlay p-4 text-center" data-testid="notices-banner">
								{noticeText}
							</div>
						) : null}
						{/* The tab root wraps the sidebar too, so the last run can open the Results tab; it renders as `display: contents`. */}
						<SimTabsSection host={host}>
							<div className="flex flex-1" data-testid="sim-container">
								{!compact && <SimSidebar host={host} sim={sim} spec={spec} simReady={simReady} />}
								<div
									className="z-1 mx-auto flex w-full min-w-[calc(275px+1vw)] flex-4 flex-col pt-0 pr-page pb-page pl-page"
									data-testid="sim-content">
									<header
										ref={header}
										className="sticky -top-px z-header -mx-page h-sim-header pt-6 pr-page pl-page whitespace-nowrap transition-colors duration-150 ease-in-out after:absolute after:inset-x-0 after:-bottom-px after:mx-auto after:h-px after:w-page-inset-w after:bg-border after:transition-[width] after:duration-150 after:ease-in-out after:content-[''] data-stuck:bg-background data-stuck:after:w-full max-lg:pt-0"
										data-testid="sim-header"
										data-stuck={stuck ? '' : undefined}>
										<div
											className="flex h-full flex-1 scrollbar-none items-stretch overflow-x-scroll [&::-webkit-scrollbar]:hidden"
											data-testid="sim-header-container">
											{compact ? <SimTitleDropdown currentSpec={spec} compact /> : <SimTabsStrip />}
											{host && (
												<SimSideOnly side="single">
													<SimImportExport />
												</SimSideOnly>
											)}
											<SimToolbar sim={sim} knownIssues={knownIssues} onOpenSettings={onOpenSettings} />
										</div>
									</header>
									<main className="flex h-4/5 grow" data-testid="sim-main">
										<SimTabsPanels />
									</main>
								</div>
							</div>
							{compact && <SimDock host={host} sim={sim} spec={spec} simReady={simReady} />}
							{/* Inside the tab root, which the tour opens tabs through. */}
							<SimOnboarding host={host} simReady={simReady} />
						</SimTabsSection>
					</div>
					<ToastArea manager={toastManager} />
				</div>
			</UpgradesSlotsProvider>
		</StickyHeaderContext>
	);
};
