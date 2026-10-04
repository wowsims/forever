import clsx from 'clsx';
import { useCallback, useContext, useId, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { SortableContext } from './context/SortableContext';
import {
	TreeBodyContext,
	TreeItemContext,
	TreeItemsContext,
	type TreeItemsContextValue,
	TreeRootContext,
	type TreeRootContextValue,
} from './context/TreeContext';
import { useSortable, useSortableStatus } from './hooks/useSortable';
import { useTreeRoot } from './hooks/useTreeRoot';
import type { TreeItemsProps, TreeRootProps } from './types';

export const Root = ({ group, depth, density, layout = 'default', sortable, disabled, className, testId, children }: TreeRootProps) => {
	const parent = useContext(TreeRootContext);
	const parentId = useContext(TreeItemContext)?.id ?? null;
	const ownGroup = useId();
	const parentSortable = useContext(SortableContext);
	const ownSortable = useSortable(parentSortable ? undefined : sortable);
	const inBody = useContext(TreeBodyContext);
	const listKey = useId();
	const element = useRef<HTMLDivElement>(null);
	const listSortable = inBody ? parentSortable : null;
	const value = useMemo<TreeRootContextValue>(
		() => ({
			group: group ?? (inBody ? ownGroup : (parent?.group ?? ownGroup)),
			depth: depth ?? (parent ? parent.depth + 1 : 0),
			density: density ?? parent?.density ?? 'default',
			layout,
			disabled: disabled ?? parent?.disabled ?? false,
			parentId,
		}),
		[group, inBody, depth, density, layout, disabled, parent, ownGroup, parentId],
	);

	// A root in a body is a list of its own, not an item's children, so its area takes an `into` drop.
	useLayoutEffect(() => {
		listSortable?.registerList(listKey, { group: value.group, parentId, element: element.current! });
	});
	useLayoutEffect(() => {
		if (!listSortable) return;
		return () => listSortable.unregisterList(listKey);
	}, [listSortable, listKey]);
	const { drop } = useSortableStatus(listSortable, listKey);
	const listProps = useMemo(() => listSortable?.listProps(listKey), [listSortable, listKey]);

	return (
		<SortableContext.Provider value={parentSortable ?? ownSortable}>
			<TreeRootContext.Provider value={value}>
				<div
					ref={element}
					className={clsx('ui-tree', className)}
					data-depth={value.depth}
					data-density={value.density}
					data-layout={layout}
					data-orientation="vertical"
					data-in-body={inBody ? '' : undefined}
					data-drop={drop ?? undefined}
					data-testid={testId ?? 'tree'}
					{...listProps}>
					{children}
				</div>
			</TreeRootContext.Provider>
		</SortableContext.Provider>
	);
};

const inDocumentOrder = (a: HTMLElement, b: HTMLElement) => (a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING ? -1 : 1);

const sameOrder = (a: ReadonlyArray<string>, b: ReadonlyArray<string>) => a.length === b.length && a.every((key, i) => key === b[i]);

// Each item registers its element, and the order is read off the DOM, so an item rendered through a
// wrapper component or a fragment counts like a direct child. A registration, a removal or an item
// found out of place bumps `version`, which re-renders this list to recompute the order even when a
// wrapper re-rendered on its own.
export const Items = ({ className, testId, children }: TreeItemsProps) => {
	useTreeRoot();
	const elements = useRef(new Map<string, HTMLElement>());
	const [order, setOrder] = useState<ReadonlyArray<string>>([]);
	const [version, setVersion] = useState(0);
	const invalidate = useCallback(() => setVersion(current => current + 1), []);

	useLayoutEffect(() => {
		const next = Array.from(elements.current.entries())
			.sort(([, a], [, b]) => inDocumentOrder(a, b))
			.map(([key]) => key);
		if (!sameOrder(order, next)) setOrder(next);
	}, [children, version, order]);

	const register = useCallback(
		(key: string, element: HTMLElement) => {
			elements.current.set(key, element);
			invalidate();
		},
		[invalidate],
	);
	const unregister = useCallback(
		(key: string) => {
			elements.current.delete(key);
			invalidate();
		},
		[invalidate],
	);
	const value = useMemo<TreeItemsContextValue>(
		() => ({
			register,
			unregister,
			indexOf: key => order.indexOf(key),
			// Two comparisons with the neighbours the order names; any mismatch re-sorts the whole list.
			check: (key, index) => {
				const element = elements.current.get(key);
				if (!element) return;
				if (order[index] !== key) {
					invalidate();
					return;
				}
				const before = index > 0 ? elements.current.get(order[index - 1]) : undefined;
				const after = elements.current.get(order[index + 1]);
				if ((before && inDocumentOrder(before, element) > 0) || (after && inDocumentOrder(element, after) > 0)) invalidate();
			},
		}),
		[register, unregister, invalidate, order],
	);

	return (
		<TreeItemsContext.Provider value={value}>
			<ul className={clsx('ui-tree-items', className)} role="list" data-testid={testId}>
				{children}
			</ul>
		</TreeItemsContext.Provider>
	);
};
