import clsx from 'clsx';
import { useCallback, useId, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { TreeBodyContext, TreeItemContext, type TreeItemContextValue } from './context/TreeContext';
import { useItemSortable } from './hooks/useItemSortable';
import { useTreeItems } from './hooks/useTreeItems';
import { useTreeRoot } from './hooks/useTreeRoot';
import type { TreeItemProps } from './types';

export const Item = ({ id, disabled, muted, className, testId, children }: TreeItemProps) => {
	const root = useTreeRoot();
	const { register, unregister, indexOf, check } = useTreeItems();
	const key = useId();
	const element = useRef<HTMLLIElement>(null);
	const index = indexOf(key);

	useLayoutEffect(() => {
		register(key, element.current!);
		return () => unregister(key);
	}, [register, unregister, key]);

	useLayoutEffect(() => check(key, index));

	const [open, setOpen] = useState(true);

	const [parts, setParts] = useState({ body: false, children: false });
	const mountPart = useCallback((part: 'body' | 'children') => {
		setParts(current => ({ ...current, [part]: true }));
		return () => setParts(current => ({ ...current, [part]: false }));
	}, []);
	const bodyId = `${key}-body`;
	const childrenId = `${key}-children`;
	const controls = [parts.body && bodyId, parts.children && childrenId].filter(Boolean).join(' ') || undefined;

	const itemDisabled = root.disabled || !!disabled;
	const { armed, drag, drop, itemProps, armProps, childrenProps } = useItemSortable({
		key,
		element,
		place: { group: root.group, id, index, parentId: root.parentId },
		disabled: itemDisabled,
	});
	const value = useMemo<TreeItemContextValue>(
		() => ({
			id,
			index,
			parentId: root.parentId,
			open,
			setOpen,
			drag,
			disabled: itemDisabled,
			bodyId,
			childrenId,
			controls,
			mountPart,
			childrenProps,
		}),
		[id, index, root.parentId, open, setOpen, drag, itemDisabled, bodyId, childrenId, controls, mountPart, childrenProps],
	);

	return (
		<TreeItemContext.Provider value={value}>
			<li
				ref={element}
				className={clsx('ui-tree-item', className)}
				data-open={open ? '' : undefined}
				data-drag={drag ?? undefined}
				data-drop={drop ?? undefined}
				data-disabled={itemDisabled ? '' : undefined}
				data-muted={muted ? '' : undefined}
				data-testid={testId ?? 'tree-item'}
				draggable={armed || undefined}
				{...itemProps}
				{...armProps}>
				<TreeBodyContext.Provider value={false}>{children}</TreeBodyContext.Provider>
			</li>
		</TreeItemContext.Provider>
	);
};
