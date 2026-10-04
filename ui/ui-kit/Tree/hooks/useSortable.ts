import { type DragEvent, type MouseEvent, type RefObject, useLayoutEffect, useMemo, useRef, useSyncExternalStore } from 'react';

import type { TreeDropRef, TreeRef, TreeSortable } from '../types';

export type SortableDrop = 'before' | 'after' | 'into';

/** An item as a drag reads it: its place in its list and its element. */
export interface SortableEntry extends TreeRef {
	element: HTMLElement;
}

export interface SortableArmProps {
	onMouseDown: (event: MouseEvent<HTMLElement>) => void;
}

export interface SortableItemProps {
	onDragStart: (event: DragEvent<HTMLElement>) => void;
	onDragEnd: () => void;
	onDragOver: (event: DragEvent<HTMLElement>) => void;
	onDragLeave: (event: DragEvent<HTMLElement>) => void;
	onDrop: (event: DragEvent<HTMLElement>) => void;
}

export type SortableChildrenProps = Omit<SortableItemProps, 'onDragStart' | 'onDragEnd'>;

interface SortableState {
	armed: string | null;
	from: string | null;
	over: { key: string; drop: SortableDrop } | null;
}

export interface SortableController {
	register: (key: string, entry: SortableEntry) => void;
	unregister: (key: string) => void;
	subscribe: (listener: () => void) => () => void;
	getState: () => SortableState;
	/** Arms the item on a press anywhere in it, outside its nested items and its text fields. */
	armProps: (key: string) => SortableArmProps;
	/** The drag source and before/after drop target handlers for the item's element. */
	itemProps: (key: string) => SortableItemProps;
	/** The `into` drop target handlers for the area that holds the item's children. */
	childrenProps: (key: string) => SortableChildrenProps;
	registerList: (key: string, list: SortableList) => void;
	unregisterList: (key: string) => void;
	/** Drop target handlers for a list's own area, so a list with no items, or the space below its last item, takes a drop. */
	listProps: (key: string) => SortableChildrenProps;
}

/** A list that is not an item's children: a root in an item's body. */
export interface SortableList {
	group: string;
	parentId: string | null;
	element: HTMLElement;
}

type TargetMode = 'item' | 'into' | 'list';

const TRANSFER_TYPE = 'application/x-ui-tree';

const TEXT_FIELDS = 'input, textarea, select, [contenteditable="true"]';

const sameGroup = (from: TreeRef, to: TreeDropRef) => from.group === to.group;

const pastMidpoint = (event: DragEvent, { element }: SortableEntry) => {
	const rect = element.getBoundingClientRect();
	return event.clientY > rect.top + rect.height / 2;
};

const refOf = ({ group, id, index, parentId }: SortableEntry): TreeRef => ({ group, id, index, parentId });

const createSortable = (options: RefObject<TreeSortable | undefined>): SortableController => {
	const entries = new Map<string, SortableEntry>();
	const lists = new Map<string, SortableList>();
	const listeners = new Set<() => void>();
	let state: SortableState = { armed: null, from: null, over: null };
	let releaseMouseUp: (() => void) | null = null;

	const set = (next: Partial<SortableState>) => {
		const merged = { ...state, ...next };
		if (merged.armed === state.armed && merged.from === state.from && merged.over?.key === state.over?.key && merged.over?.drop === state.over?.drop)
			return;
		state = merged;
		for (const listener of [...listeners]) listener();
	};

	const end = () => {
		releaseMouseUp?.();
		releaseMouseUp = null;
		set({ armed: null, from: null, over: null });
	};

	// A press that never turns into a drag disarms on its release.
	const arm = (key: string, doc: Document) => {
		releaseMouseUp?.();
		const onMouseUp = () => {
			if (state.from === null) end();
		};
		doc.addEventListener('mouseup', onMouseUp);
		releaseMouseUp = () => doc.removeEventListener('mouseup', onMouseUp);
		set({ armed: key });
	};

	// The innermost item around the press; a press in a portalled popup has none.
	const ownerOf = (target: Element) => {
		for (let element: Element | null = target; element; element = element.parentElement) {
			for (const [key, entry] of entries) if (entry.element === element) return key;
		}
		return null;
	};

	// `refuse`: the drop would land on or inside the dragged item, whatever `canDrop` says (a move into
	// its own subtree would detach it), or leave it where it is. The target still owns the event, so
	// hovering the dragged item never reads as a drop around its parent. `pass`: the group rule or
	// `canDrop` turns this target down, and the enclosing item or children area decides instead.
	const verdict = (from: SortableEntry, target: HTMLElement, to: TreeDropRef) => {
		if (from.element.contains(target)) return 'refuse';
		if (to.group === from.group && to.parentId === from.parentId && (to.index === from.index || to.index === from.index + 1)) return 'refuse';
		return (options.current?.canDrop ?? sameGroup)(refOf(from), to) ? 'accept' : 'pass';
	};

	const countPast = (event: DragEvent, group: string, parentId: string | null) => {
		let index = 0;
		for (const entry of entries.values()) {
			if (entry.parentId === parentId && entry.group === group && pastMidpoint(event, entry)) index++;
		}
		return index;
	};

	const resolve = (key: string, event: DragEvent, mode: TargetMode) => {
		const from = state.from === null ? undefined : entries.get(state.from);
		if (!from) return 'pass';
		if (mode === 'list') {
			const list = lists.get(key);
			if (!list) return 'pass';
			const to = { group: list.group, parentId: list.parentId, index: countPast(event, list.group, list.parentId) };
			const result = verdict(from, list.element, to);
			return result === 'accept' ? { from, drop: 'into' as const, to } : result;
		}
		const over = entries.get(key);
		if (!over) return 'pass';
		let drop: SortableDrop;
		let to: TreeDropRef;
		if (mode === 'into') {
			drop = 'into';
			to = { group: over.group, parentId: over.id, index: countPast(event, over.group, over.id) };
		} else {
			const after = pastMidpoint(event, over);
			drop = after ? 'after' : 'before';
			to = { group: over.group, parentId: over.parentId, index: after ? over.index + 1 : over.index };
		}
		const result = verdict(from, over.element, to);
		return result === 'accept' ? { from, drop, to } : result;
	};

	// The innermost target that accepts or refuses stops the event; a passing one lets it bubble.
	const onDragOver = (key: string, mode: TargetMode) => (event: DragEvent<HTMLElement>) => {
		if (state.from === null) return;
		const target = resolve(key, event, mode);
		if (target === 'pass' || target === 'refuse') set({ over: null });
		if (target === 'pass') return;
		event.stopPropagation();
		if (target === 'refuse') return;
		event.preventDefault();
		event.dataTransfer.dropEffect = 'move';
		set({ over: { key, drop: target.drop } });
	};

	const onDragLeave = (key: string) => (event: DragEvent<HTMLElement>) => {
		if (state.over?.key !== key) return;
		if (event.relatedTarget instanceof Node && event.currentTarget.contains(event.relatedTarget)) return;
		set({ over: null });
	};

	// The drop ends the drag here: a move to another list unmounts the source item, and a
	// detached element's dragend never reaches React.
	const onDrop = (key: string, mode: TargetMode) => (event: DragEvent<HTMLElement>) => {
		if (state.from === null) return;
		const target = resolve(key, event, mode);
		if (target === 'pass') return;
		event.stopPropagation();
		event.preventDefault();
		end();
		if (target !== 'refuse') options.current?.onMove({ from: refOf(target.from), to: target.to });
	};

	return {
		register: (key, entry) => {
			entries.set(key, entry);
		},
		unregister: key => {
			entries.delete(key);
		},
		subscribe: listener => {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		getState: () => state,
		armProps: key => ({
			onMouseDown: event => {
				const target = event.target as Element;
				if (ownerOf(target) !== key || target.closest(TEXT_FIELDS)) return;
				arm(key, event.currentTarget.ownerDocument);
			},
		}),
		itemProps: key => ({
			onDragStart: event => {
				const entry = entries.get(key);
				// A nested item's dragstart bubbles through its ancestors.
				if (!entry || event.target !== entry.element) return;
				event.dataTransfer.effectAllowed = 'move';
				event.dataTransfer.setData(TRANSFER_TYPE, entry.id);
				set({ from: key, over: null });
			},
			onDragEnd: end,
			onDragOver: onDragOver(key, 'item'),
			onDragLeave: onDragLeave(key),
			onDrop: onDrop(key, 'item'),
		}),
		childrenProps: key => ({
			onDragOver: onDragOver(key, 'into'),
			onDragLeave: onDragLeave(key),
			onDrop: onDrop(key, 'into'),
		}),
		registerList: (key, list) => {
			lists.set(key, list);
		},
		unregisterList: key => {
			lists.delete(key);
		},
		listProps: key => ({
			onDragOver: onDragOver(key, 'list'),
			onDragLeave: onDragLeave(key),
			onDrop: onDrop(key, 'list'),
		}),
	};
};

/**
 * Native drag and drop over a set of registered items. Returns `null` without `sortable`. The
 * controller is created only while `sortable` is given, once, and reads it from the latest render, so a `sortable` object
 * written inline changes nothing downstream.
 */
export const useSortable = (sortable: TreeSortable | undefined): SortableController | null => {
	const options = useRef(sortable);
	useLayoutEffect(() => {
		options.current = sortable;
	});
	const enabled = sortable !== undefined;
	return useMemo(() => (enabled ? createSortable(options) : null), [enabled]);
};

const inert = () => () => {};

/** The item's drag marks. Only an item whose marks change re-renders. */
export const useSortableStatus = (controller: SortableController | null, key: string) => {
	const subscribe = controller?.subscribe ?? inert;
	const armed = useSyncExternalStore(subscribe, () => controller?.getState().armed === key);
	const drag = useSyncExternalStore<'from' | 'to' | null>(subscribe, () => {
		const state = controller?.getState();
		if (state?.from === key) return 'from';
		return state?.over?.key === key ? 'to' : null;
	});
	const drop = useSyncExternalStore(subscribe, () => {
		const over = controller?.getState().over;
		return over?.key === key ? over.drop : null;
	});
	return { armed, drag, drop };
};
