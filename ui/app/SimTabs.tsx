import { Tabs } from '@base-ui/react/tabs';
import { BulkModeTabs } from '@features/bulk/components/BulkModeTabs';
import { BULK_TAB_ID, TabActivationContext } from '@features/hooks/useActivateTab';
import i18n from '@i18n/config';
import { childProps } from '@ui-kit/child_props';
import { Icon } from '@ui-kit/Icon';
import { TabNav, type TabNavProps } from '@ui-kit/TabNav';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import clsx from 'clsx';
import { createContext, type ReactNode, use, useCallback, useMemo, useState } from 'react';

import { trackPageView } from '../tracking/utils';
import type { SimTabDefProps } from './SimTabDef';

export interface SimTabsProps {
	defs: ReactNode;
	children: ReactNode;
}

interface SimTabsListContextValue {
	tabs: ReadonlyArray<SimTabDefProps>;
}

const SimTabsListContext = createContext<SimTabsListContextValue | null>(null);
const SimActiveTabContext = createContext<string | null>(null);
// The single sim tab open before the Upgrades side, which the switch returns to.
const SimLastSingleTabContext = createContext<string | null>(null);

/** The open tab's id, apart from the list so that a tab change re-renders only what follows it. */
export const useActiveSimTab = () => use(SimActiveTabContext);

/** Single sim or Upgrades: the Upgrades side is the batch tab, every other tab is the single sim's. */
export type SimSide = 'single' | 'upgrades';

export const useSimSide = (): SimSide => (useActiveSimTab() === BULK_TAB_ID ? 'upgrades' : 'single');

/** Renders its children on one side only, and unmounts them on the other. */
export const SimSideOnly = ({ side, children }: { side: SimSide; children: ReactNode }) => (useSimSide() === side ? children : null);

// `Tabs.Root` is the one React ancestor the strip, the panels and the sidebar's tab links all need; the
// shell renders it around the whole sim container, so a pane is a plain child of `main`. It mounts with no
// `defs` at all (before a host exists) and keeps mounting through every later render — swapping it out for a
// plain fragment on the render `defs` first arrives would remount the sidebar, the header and `main`.
export const SimTabs = ({ defs, children }: SimTabsProps) => {
	const [activeId, setActiveId] = useState<string | null>(null);
	const [lastSingleId, setLastSingleId] = useState<string | null>(null);
	const tabs = useMemo(() => childProps<SimTabDefProps>(defs), [defs]);
	// The first tab is the one open on load, and an id no tab carries falls back to it. `null`, not
	// `undefined`, so `Tabs.Root` is controlled from its first render (with no tabs yet) rather than
	// flipping from uncontrolled once one exists.
	const value = tabs.find(tab => tab.id === activeId)?.id ?? tabs[0]?.id ?? null;

	// Base UI fires `onValueChange` for a user interaction only, so a pane opening a sibling tab
	// reports the page view through here instead. Stable, so a shell re-render unrelated to the tabs
	// (the sticky header toggling) does not re-render every mounted pane through `TabActivationContext`.
	const activate = useCallback(
		(id: string) => {
			setActiveId(id);
			if (id !== BULK_TAB_ID) setLastSingleId(id);
			trackPageView(tabs.find(tab => tab.id === id)?.title ?? id, id);
		},
		[tabs],
	);

	const listContextValue = useMemo(() => ({ tabs }), [tabs]);

	return (
		<TabActivationContext value={activate}>
			<Tabs.Root className="contents" value={value} onValueChange={next => activate(String(next))}>
				<SimTabsListContext value={listContextValue}>
					<SimActiveTabContext value={value}>
						<SimLastSingleTabContext value={lastSingleId}>{children}</SimLastSingleTabContext>
					</SimActiveTabContext>
				</SimTabsListContext>
			</Tabs.Root>
		</TabActivationContext>
	);
};

export interface SimTabsStripProps {
	placement?: TabNavProps['placement'];
	className?: string;
}

// Single sim or Upgrades. The Upgrades side is the batch tab alone, so the strip leaves it out and this switch opens it.
// In the dock it shrinks to two icons at the start of the tab row, so it adds no row of its own.
const SimSideSwitch = ({ side, singleTabId, dock }: { side: SimSide; singleTabId: string; dock: boolean }) => {
	const activate = use(TabActivationContext);
	const label = (value: SimSide, icon: 'user' | 'arrow-up') =>
		dock ? <Icon name={icon} title={i18n.t(`sim_sides.${value}`)} /> : i18n.t(`sim_sides.${value}`);
	return (
		<ToggleGroup<SimSide>
			className={dock ? 'mx-2 shrink-0 self-center' : 'mr-6 shrink-0 self-center'}
			size="sm"
			aria-label={i18n.t('sim_sides.label')}
			testId="sim-side-switch"
			options={[
				{ value: 'single', label: label('single', 'user') },
				{ value: 'upgrades', label: label('upgrades', 'arrow-up') },
			]}
			value={side}
			onChange={next => activate(next === 'upgrades' ? BULK_TAB_ID : singleTabId)}
		/>
	);
};

export const SimTabsStrip = ({ placement, className }: SimTabsStripProps) => {
	const list = use(SimTabsListContext);
	const side = useSimSide();
	const lastSingle = use(SimLastSingleTabContext);
	if (!list?.tabs.length) return null;
	const singleTabs = list.tabs.filter(tab => tab.id !== BULK_TAB_ID);
	const hasUpgrades = singleTabs.length < list.tabs.length;
	const dock = placement === 'bottom';
	const strip = (
		<>
			{hasUpgrades && <SimSideSwitch side={side} singleTabId={lastSingle ?? singleTabs[0].id} dock={dock} />}
			{side === 'upgrades' && <BulkModeTabs placement={placement} className={dock ? 'min-w-0 flex-1' : undefined} />}
			{side === 'single' && (
				<TabNav
					variant="sim"
					placement={placement}
					className={dock ? 'min-w-0 flex-1' : className}
					testId="sim-tabs"
					tabs={singleTabs.map(tab => ({ id: tab.id, label: tab.title, badge: tab.badge, renderTab: tab.renderTab }))}
				/>
			)}
		</>
	);
	// The dock has a fixed height for one tab row, so the switch and the tabs share it.
	return dock ? <div className={clsx('flex items-stretch', className)}>{strip}</div> : strip;
};

export const SimTabsPanels = () => {
	const list = use(SimTabsListContext);
	if (!list?.tabs.length) return null;
	return (
		<>
			{list.tabs.map(tab => (
				// `keepMounted`: every pane is built once and three of them read the live document, so none may be unmounted.
				<Tabs.Panel
					key={tab.id}
					value={tab.id}
					keepMounted
					className="max-w-full grow pt-6 opacity-100 transition-opacity duration-150 ease-linear data-starting-style:opacity-0 max-lg:pt-4"
					data-testid="sim-tab-panel">
					{tab.children}
				</Tabs.Panel>
			))}
		</>
	);
};
