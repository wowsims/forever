import type { IconName } from '@ui-kit/Icon';
import type { ReactNode } from 'react';

export type TreeDensity = 'default' | 'compact';
/** `inline` puts the body on the header row, between the title and the trigger. */
export type TreeLayout = 'default' | 'inline';

export interface TreeRef {
	group: string;
	id: string;
	index: number;
	parentId: string | null;
}

export interface TreeDropRef {
	group: string;
	parentId: string | null;
	/** The insertion index, read against the list as it is before the dragged item leaves it. */
	index: number;
}

export interface TreeMove {
	from: TreeRef;
	to: TreeDropRef;
}

export interface TreeSortable {
	/** Called once per accepted drop. */
	onMove: (move: TreeMove) => void;
	/**
	 * Defaults to `from.group === to.group`. A drop onto the dragged item, into its own subtree, or
	 * back where it already is stays refused whatever this returns.
	 */
	canDrop?: (from: TreeRef, to: TreeDropRef) => boolean;
}

interface TreePartProps {
	className?: string;
	testId?: string;
}

export interface TreeRootProps extends TreePartProps {
	/**
	 * A root in children inherits its parent's; a top-level root, or a root in a body, without one gets a
	 * group of its own, so a body list is never the same list as its item's `Children`.
	 */
	group?: string;
	/** Defaults to the parent root's depth + 1, and 0 at the top. */
	depth?: number;
	/** Nested roots inherit it. */
	density?: TreeDensity;
	layout?: TreeLayout;
	/** Makes the items draggable. The topmost sortable root owns the drag; the roots nested under it share it and ignore their own. */
	sortable?: TreeSortable;
	/** Nested roots inherit it. */
	disabled?: boolean;
	children: ReactNode;
}

export interface TreeItemsProps extends TreePartProps {
	children: ReactNode;
}

export interface TreeItemProps extends TreePartProps {
	/** Stable identity, used for moves and as the `parentId` of the items nested under it. */
	id: string;
	disabled?: boolean;
	/** Fades the handle and title while the item stays movable and openable; for a row whose body is switched off. */
	muted?: boolean;
	children?: ReactNode;
}

export interface TreeItemHeaderProps extends TreePartProps {
	children: ReactNode;
}

export interface TreeItemHandleProps extends TreePartProps {
	label: string;
}

export interface TreeItemTitleProps extends TreePartProps {
	as?: 'span' | 'h6';
	children: ReactNode;
}

export interface TreeItemTriggerProps extends TreePartProps {
	label: string;
}

export interface TreeItemActionsProps extends TreePartProps {
	/** The menu button's accessible name. */
	label: string;
	children: ReactNode;
}

export interface TreeItemActionProps extends TreePartProps {
	label: string;
	icon?: IconName;
	tone?: 'default' | 'danger';
	disabled?: boolean;
	onClick?: () => void;
}

/** Status and toggles on the header row, just before the trigger. */
export interface TreeItemAsideProps extends TreePartProps {
	children: ReactNode;
}

/** A `Tree.Root` inside a body is a list of its own: it indents like `Children` and takes an `into` drop when empty. */
export interface TreeItemBodyProps extends TreePartProps {
	children: ReactNode;
}

export interface TreeItemChildrenProps extends TreePartProps {
	children: ReactNode;
}

export interface TreeAddButtonProps extends TreePartProps {
	/** The button text after a plus, or with `icon` its accessible name and tooltip. */
	label: string;
	icon?: IconName;
	/** `success` colours the button's text the way a create action reads elsewhere; meant for the icon form. */
	tone?: 'default' | 'success';
	disabled?: boolean;
	onClick: () => void;
}
