import { useRef } from 'react';

import type { InputConfig } from '../input';
import { canDeleteAt, carryId, keyFor, moveItem, reconcileIds } from '../list_ids';
import { useConnectedField } from './useConnectedField';

export interface ListItemsConfig<M, T> extends InputConfig<M, Array<T>> {
	newItem: () => T;
	/** Without it, `copy` does nothing. */
	copyItem?: (item: T) => T;
	minimumItems?: number;
}

/**
 * Binds a list to a config: every operation reads the list off the source, writes one fresh array
 * through `config.setValue` and leaves the source array alone.
 *
 * `itemConfig` hands a row a binding to its own item. Its `getValue` re-reads the source and finds
 * the item by key, so a binding made before a reorder, or before the source replaced every row with
 * a clone, still reads its own row.
 */
export const useListItems = <M, T>(modObject: M, config: ListItemsConfig<M, T>) => {
	const configRef = useRef(config);
	configRef.current = config;
	const lastSeenRef = useRef<ReadonlyArray<T> | null>(null);

	const { value, setValue, hidden, disabled, frameProps } = useConnectedField(modObject, config);
	reconcileIds(lastSeenRef.current, value);
	lastSeenRef.current = value;

	const read = () => {
		const list = configRef.current.getValue(modObject);
		reconcileIds(lastSeenRef.current, list);
		lastSeenRef.current = list;
		return list;
	};

	const indexOfKey = (list: ReadonlyArray<T>, key: number) => list.findIndex((candidate, i) => keyFor(candidate, i) === key);

	// Indices name rows of the render the caller saw; each operation finds that row in a fresh read,
	// so a write the render has not caught up with moves the index along with its row.
	const rendered = value;
	const locate = (list: ReadonlyArray<T>, index: number) => (index >= 0 && index < rendered.length ? indexOfKey(list, keyFor(rendered[index], index)) : -1);

	const canRemove = (index: number) => canDeleteAt(index, configRef.current.minimumItems);

	const append = (item: T) => setValue(read().concat([item]));
	const create = () => append(configRef.current.newItem());

	const remove = (index: number) => {
		const list = read().slice();
		const at = locate(list, index);
		if (at === -1 || !canRemove(at)) return;
		list.splice(at, 1);
		setValue(list);
	};

	const copy = (index: number) => {
		const { copyItem } = configRef.current;
		if (!copyItem) return;
		const list = read().slice();
		const at = locate(list, index);
		if (at === -1) return;
		list.splice(at, 0, copyItem(list[at]));
		setValue(list);
	};

	/** `to` is a slot in `0..items.length`, read against the positions before `from` is taken out. */
	const move = (from: number, to: number) => {
		if (to < 0 || to > rendered.length) return;
		const list = read();
		const at = locate(list, from);
		const slot = to === rendered.length ? list.length : locate(list, to);
		if (at === -1 || slot === -1 || slot === at || slot === at + 1) return;
		setValue(moveItem(list, at, slot));
	};

	const replaceByKey = (key: number, next: T) => {
		const list = read().slice();
		const index = indexOfKey(list, key);
		if (index === -1) return;
		carryId(list[index], next);
		list[index] = next;
		setValue(list);
	};

	const itemConfig = (item: T, index: number): InputConfig<M, T> => {
		const key = keyFor(item, index);
		return {
			storeSubscribe: config.storeSubscribe,
			storeField: config.storeField,
			getValue: () => {
				const list = read();
				return list[indexOfKey(list, key)];
			},
			setValue: (_, next) => replaceByKey(key, next),
		};
	};

	return { items: value, itemConfig, create, append, remove, copy, move, canRemove, hidden, disabled, frameProps };
};
