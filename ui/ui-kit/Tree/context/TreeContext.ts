import { createContext } from 'react';

import type { SortableChildrenProps } from '../hooks/useSortable';
import type { TreeDensity, TreeLayout } from '../types';

export interface TreeRootContextValue {
	group: string;
	depth: number;
	density: TreeDensity;
	layout: TreeLayout;
	disabled: boolean;
	/** The id of the item this root is nested in, `null` at the top. */
	parentId: string | null;
}

export interface TreeItemContextValue {
	id: string;
	index: number;
	parentId: string | null;
	open: boolean;
	setOpen: (open: boolean) => void;
	drag: 'from' | 'to' | null;
	disabled: boolean;
	bodyId: string;
	childrenId: string;
	/** The ids of the body and children parts that are mounted, for the trigger's `aria-controls`; `undefined` with neither. */
	controls: string | undefined;
	/** Body and Children announce themselves while mounted. */
	mountPart: (part: 'body' | 'children') => () => void;
	childrenProps: SortableChildrenProps | undefined;
}

export interface TreeItemsContextValue {
	register: (key: string, element: HTMLElement) => void;
	unregister: (key: string) => void;
	/** -1 until the item's first layout pass has registered it. */
	indexOf: (key: string) => number;
	/** Recomputes the order when the element registered under `key` no longer sits at `index`. */
	check: (key: string, index: number) => void;
}

export const TreeRootContext = createContext<TreeRootContextValue | null>(null);
export const TreeItemsContext = createContext<TreeItemsContextValue | null>(null);
export const TreeItemContext = createContext<TreeItemContextValue | null>(null);
/** Closes the actions menu an action sits in; `null` for an action outside a menu. */
export const TreeActionsContext = createContext<(() => void) | null>(null);
/** Whether a root sits in an item's body rather than in its children; every item resets it. */
export const TreeBodyContext = createContext(false);
