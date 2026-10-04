import { type RefObject, useContext, useLayoutEffect, useMemo } from 'react';

import { SortableContext } from '../context/SortableContext';
import type { TreeRef } from '../types';
import { useSortableStatus } from './useSortable';

interface ItemSortableOptions {
	key: string;
	element: RefObject<HTMLElement | null>;
	place: TreeRef;
	disabled: boolean;
}

/** Registers the item with the sortable above it and reads its drag marks; inert without one. */
export const useItemSortable = ({ key, element, place, disabled }: ItemSortableOptions) => {
	const sortable = useContext(SortableContext);
	const { group, id, index, parentId } = place;

	useLayoutEffect(() => {
		sortable?.register(key, { group, id, index, parentId, element: element.current! });
	});
	useLayoutEffect(() => {
		if (!sortable) return;
		return () => sortable.unregister(key);
	}, [sortable, key]);

	const status = useSortableStatus(sortable, key);
	const armProps = useMemo(() => (sortable && !disabled ? sortable.armProps(key) : undefined), [sortable, disabled, key]);
	const childrenProps = useMemo(() => sortable?.childrenProps(key), [sortable, key]);
	const itemProps = useMemo(() => sortable?.itemProps(key), [sortable, key]);
	return { ...status, itemProps, childrenProps, armProps };
};
