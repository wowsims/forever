/** The first `minimumItems` rows cannot be removed; every row after them can. */
export const canDeleteAt = (index: number, minimumItems: number | undefined): boolean => !minimumItems || index + 1 > minimumItems;

/** Moves `from` to `to` in a copy of `list`, with `to` read against the pre-removal positions. */
export const moveItem = <T>(list: ReadonlyArray<T>, from: number, to: number): Array<T> => {
	const next = list.slice();
	const [item] = next.splice(from, 1);
	next.splice(to > from ? to - 1 : to, 0, item);
	return next;
};

const itemIds = new WeakMap<object, number>();
let nextItemId = 1;

export const keyFor = (item: unknown, fallbackIndex: number): number => {
	if (item === null || typeof item !== 'object') return -(fallbackIndex + 1);
	let id = itemIds.get(item);
	if (id === undefined) {
		id = nextItemId++;
		itemIds.set(item, id);
	}
	return id;
};

export const carryId = (oldItem: unknown, newItem: unknown): void => {
	if (oldItem === null || typeof oldItem !== 'object') return;
	if (newItem === null || typeof newItem !== 'object') return;
	const id = itemIds.get(oldItem);
	if (id !== undefined) itemIds.set(newItem, id);
};

export const hasId = (item: unknown): boolean => item !== null && typeof item === 'object' && itemIds.has(item);

/**
 * Hands each id-less object in `list` the id of the object `prev` held at the same index, when that
 * object is gone from `list`: a source that replaces rows with clones keeps every row's key.
 * Idempotent, so two readers may both run it on the same pair.
 */
export const reconcileIds = (prev: ReadonlyArray<unknown> | null, list: ReadonlyArray<unknown>): void => {
	if (prev === null || prev === list) return;
	const stillPresent = new Set(list);
	for (let index = 0; index < list.length; index++) {
		const item = list[index];
		if (hasId(item)) continue;
		const old = prev[index];
		if (old !== undefined && !stillPresent.has(old)) carryId(old, item);
	}
};
