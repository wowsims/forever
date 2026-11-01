import type { TabNavTab } from '@ui-kit/TabNav';
import type { ReactNode } from 'react';

export interface SimTabDefProps {
	id: string;
	title: string;
	badge?: string;
	renderTab?: TabNavTab['renderTab'];
	/** The whole pane, not just its body: `RotationTabPane` picks its own wrapper class and the results pane has no wrapper at all. */
	children: ReactNode;
}

// A declaration, never rendered: `SimTabs` reads these props off its children and hands them to
// `SimTabsStrip` (the tab) and `SimTabsPanels` (the panel), which render wherever the shell places them.
export const SimTabDef = (_props: SimTabDefProps): null => null;
