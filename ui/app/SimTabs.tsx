import { Tabs } from '@base-ui/react/tabs';
import { TabActivationContext } from '@features/hooks/useActivateTab';
import { childProps } from '@ui-kit/child_props';
import { TabNav, type TabNavProps } from '@ui-kit/TabNav';
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

/** The open tab's id, apart from the list so that a tab change re-renders only what follows it. */
export const useActiveSimTab = () => use(SimActiveTabContext);

// `Tabs.Root` mounts with no `defs` at all (before a host exists) and keeps mounting through every later
// render: swapping it out for a plain fragment on the render `defs` first arrives would remount the sidebar,
// the header and `main`.
export const SimTabs = ({ defs, children }: SimTabsProps) => {
	const [activeId, setActiveId] = useState<string | null>(null);
	const tabs = useMemo(() => childProps<SimTabDefProps>(defs), [defs]);
	// The value is `null`, not `undefined`, so `Tabs.Root` is controlled from its first render (with no tabs yet).
	const value = tabs.find(tab => tab.id === activeId)?.id ?? tabs[0]?.id ?? null;

	// Base UI fires `onValueChange` for a user interaction only, so a pane opening a sibling tab
	// reports the page view through here instead. Stable, so a shell re-render unrelated to the tabs
	// does not re-render every mounted pane through `TabActivationContext`.
	const activate = useCallback(
		(id: string) => {
			setActiveId(id);
			trackPageView(tabs.find(tab => tab.id === id)?.title ?? id, id);
		},
		[tabs],
	);

	const listContextValue = useMemo(() => ({ tabs }), [tabs]);

	return (
		<TabActivationContext value={activate}>
			<Tabs.Root className="contents" value={value} onValueChange={next => activate(String(next))}>
				<SimTabsListContext value={listContextValue}>
					<SimActiveTabContext value={value}>{children}</SimActiveTabContext>
				</SimTabsListContext>
			</Tabs.Root>
		</TabActivationContext>
	);
};

export interface SimTabsStripProps {
	placement?: TabNavProps['placement'];
	className?: string;
}

export const SimTabsStrip = ({ placement, className }: SimTabsStripProps) => {
	const list = use(SimTabsListContext);
	if (!list?.tabs.length) return null;
	return (
		<TabNav
			variant="sim"
			placement={placement}
			className={className}
			testId="sim-tabs"
			tabs={list.tabs.map(tab => ({ id: tab.id, label: tab.title, badge: tab.badge, renderTab: tab.renderTab }))}
		/>
	);
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
