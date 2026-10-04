import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { keyFor } from '../list_ids';
import { type ListItemsConfig, useListItems } from './useListItems';

interface Row {
	name: string;
}

// A stand-in facade: a list plus the (onChange) => unsubscribe contract of state/subscriptions.ts.
class Rows {
	private listeners = new Set<() => void>();
	constructor(public value: Array<Row>) {}
	set(next: Array<Row>) {
		this.value = next;
		Array.from(this.listeners).forEach(listener => listener());
	}
	readonly subscribe: StoreSubscribe = listener => {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	};
}

const rowsOf = (...names: Array<string>) => new Rows(names.map(name => ({ name })));
const names = (list: ReadonlyArray<Row>) => list.map(row => row.name);

const setup = (rows: Rows, extra: Partial<ListItemsConfig<Rows, Row>> = {}) => {
	const setValue = vi.fn((model: Rows, next: Array<Row>) => model.set(next));
	const config: ListItemsConfig<Rows, Row> = {
		storeSubscribe: model => model.subscribe,
		getValue: model => model.value,
		setValue,
		newItem: () => ({ name: 'new' }),
		copyItem: row => ({ name: `${row.name} copy` }),
		...extra,
	};
	const hook = renderHook(() => useListItems(rows, config));
	return { ...hook, setValue, config };
};

describe('useListItems', () => {
	it('reads the bound list and follows the source', () => {
		const rows = rowsOf('a', 'b');
		const { result } = setup(rows);
		expect(names(result.current.items)).toEqual(['a', 'b']);

		act(() => rows.set([...rows.value, { name: 'c' }]));
		expect(names(result.current.items)).toEqual(['a', 'b', 'c']);
	});

	it('create appends a new item to a copy of the list, in one write', () => {
		const rows = rowsOf('a');
		const before = rows.value;
		const { result, setValue } = setup(rows);

		act(() => result.current.create());

		expect(setValue).toHaveBeenCalledOnce();
		expect(names(rows.value)).toEqual(['a', 'new']);
		expect(rows.value).not.toBe(before);
		expect(names(before)).toEqual(['a']);
		expect(names(result.current.items)).toEqual(['a', 'new']);
	});

	it('remove drops the row from a copy of the list, in one write', () => {
		const rows = rowsOf('a', 'b', 'c');
		const before = rows.value;
		const { result, setValue } = setup(rows);

		act(() => result.current.remove(1));

		expect(setValue).toHaveBeenCalledOnce();
		expect(names(rows.value)).toEqual(['a', 'c']);
		expect(names(before)).toEqual(['a', 'b', 'c']);
	});

	it('copy inserts the copied row at its index, in one write, and does nothing without copyItem', () => {
		const rows = rowsOf('a', 'b');
		const before = rows.value;
		const { result, setValue } = setup(rows);
		const keyOfB = keyFor(result.current.items[1], 1);

		act(() => result.current.copy(1));

		expect(setValue).toHaveBeenCalledOnce();
		expect(names(rows.value)).toEqual(['a', 'b copy', 'b']);
		expect(names(before)).toEqual(['a', 'b']);
		expect(keyFor(rows.value[2], 2)).toBe(keyOfB);
		expect(keyFor(rows.value[1], 1)).not.toBe(keyOfB);

		const bare = rowsOf('a');
		const { result: withoutCopy, setValue: bareWrite } = setup(bare, { copyItem: undefined });
		act(() => withoutCopy.current.copy(0));
		expect(bareWrite).not.toHaveBeenCalled();
	});

	it('move reorders a copy of the list with the destination read before removal, in one write', () => {
		const rows = rowsOf('a', 'b', 'c', 'd');
		const before = rows.value;
		const { result, setValue } = setup(rows);

		act(() => result.current.move(0, 3));

		expect(setValue).toHaveBeenCalledOnce();
		expect(names(rows.value)).toEqual(['b', 'c', 'a', 'd']);
		expect(names(before)).toEqual(['a', 'b', 'c', 'd']);
	});

	it('does not write for a move that leaves the order as it is', () => {
		const rows = rowsOf('a', 'b', 'c');
		const { result, setValue } = setup(rows);

		act(() => result.current.move(1, 1));
		act(() => result.current.move(1, 2));

		expect(setValue).not.toHaveBeenCalled();
	});

	it('canRemove respects the minimum, and remove refuses a row the minimum keeps', () => {
		const rows = rowsOf('only');
		const { result, setValue } = setup(rows, { minimumItems: 1 });

		expect(result.current.canRemove(0)).toBe(false);
		act(() => result.current.remove(0));
		expect(setValue).not.toHaveBeenCalled();
		expect(names(rows.value)).toEqual(['only']);

		act(() => result.current.create());
		expect(result.current.canRemove(0)).toBe(false);
		expect(result.current.canRemove(1)).toBe(true);

		act(() => result.current.remove(1));
		expect(names(rows.value)).toEqual(['only']);
	});

	it('bases each operation on the source, not on the last render', () => {
		const rows = rowsOf('a');
		const { result, setValue } = setup(rows);
		const { create } = result.current;

		// A write that bypasses the hook and is not yet rendered.
		rows.value = [...rows.value, { name: 'b' }];
		act(() => create());

		expect(setValue).toHaveBeenCalledOnce();
		expect(names(rows.value)).toEqual(['a', 'b', 'new']);
	});

	it('acts on the row an index named at render, after a write the render has not seen', () => {
		const removeCase = rowsOf('a', 'b', 'c');
		const { result: removing } = setup(removeCase);
		const { remove } = removing.current;
		removeCase.value = [{ name: 'x' }, ...removeCase.value];
		act(() => remove(1));
		expect(names(removeCase.value)).toEqual(['x', 'a', 'c']);

		const copyCase = rowsOf('a', 'b');
		const { result: copying } = setup(copyCase);
		const { copy } = copying.current;
		copyCase.value = [{ name: 'x' }, ...copyCase.value];
		act(() => copy(0));
		expect(names(copyCase.value)).toEqual(['x', 'a copy', 'a', 'b']);

		const moveCase = rowsOf('a', 'b', 'c');
		const { result: moving } = setup(moveCase);
		const { move } = moving.current;
		moveCase.value = [{ name: 'x' }, ...moveCase.value];
		act(() => move(0, 2));
		expect(names(moveCase.value)).toEqual(['x', 'b', 'a', 'c']);
	});

	it('writes nothing for an index outside the rendered list, or for a row that is gone', () => {
		const rows = rowsOf('a', 'b');
		const { result, setValue } = setup(rows);
		const { remove, copy, move } = result.current;

		act(() => remove(2));
		act(() => copy(-1));
		act(() => move(0, 3));
		act(() => move(2, 0));
		act(() => move(-1, 1));
		expect(setValue).not.toHaveBeenCalled();

		rows.value = rows.value.slice(1);
		act(() => remove(0));
		act(() => move(0, 2));
		expect(setValue).not.toHaveBeenCalled();
		expect(names(rows.value)).toEqual(['b']);
	});

	it('keeps each row key across a render that leaves the list alone', () => {
		const rows = rowsOf('a', 'b');
		const { result, rerender } = setup(rows);
		const items = result.current.items;
		const keys = items.map(keyFor);

		rerender();

		expect(result.current.items).toBe(items);
		expect(result.current.items.map(keyFor)).toEqual(keys);
	});
});

describe('useListItems itemConfig', () => {
	it('inherits the list source', () => {
		const rows = rowsOf('a');
		const { result, config } = setup(rows);
		const item = result.current.itemConfig(result.current.items[0], 0);
		expect(item.storeSubscribe).toBe(config.storeSubscribe);
		expect(item).toHaveProperty('storeField', config.storeField);
	});

	it('re-reads its row by key after a reorder', () => {
		const rows = rowsOf('a', 'b', 'c');
		const { result } = setup(rows);
		const b = result.current.itemConfig(result.current.items[1], 1);

		act(() => result.current.move(1, 0));

		expect(names(rows.value)).toEqual(['b', 'a', 'c']);
		expect(b.getValue(rows)).toBe(rows.value[0]);
	});

	it('keeps its row when the source replaces every row with a clone', () => {
		const rows = rowsOf('a', 'b');
		const { result } = setup(rows);
		const b = result.current.itemConfig(result.current.items[1], 1);

		act(() => rows.set(rows.value.map(row => ({ ...row }))));

		expect(b.getValue(rows)).toBe(rows.value[1]);
	});

	it('replaces its row in a copy of the list, in one write, and keeps the row key', () => {
		const rows = rowsOf('a', 'b');
		const before = rows.value;
		const { result, setValue } = setup(rows);
		const key = keyFor(result.current.items[1], 1);
		const b = result.current.itemConfig(result.current.items[1], 1);

		act(() => b.setValue(rows, { name: 'b2' }));

		expect(setValue).toHaveBeenCalledOnce();
		expect(names(rows.value)).toEqual(['a', 'b2']);
		expect(names(before)).toEqual(['a', 'b']);
		expect(keyFor(rows.value[1], 1)).toBe(key);
		expect(b.getValue(rows)).toBe(rows.value[1]);
	});

	it('writes nothing for a row that is gone', () => {
		const rows = rowsOf('a', 'b');
		const { result, setValue } = setup(rows);
		const b = result.current.itemConfig(result.current.items[1], 1);

		act(() => result.current.remove(1));
		setValue.mockClear();
		act(() => b.setValue(rows, { name: 'b2' }));

		expect(setValue).not.toHaveBeenCalled();
		expect(names(rows.value)).toEqual(['a']);
	});
});

describe('useListItems frame', () => {
	it('passes the field frame, hidden and disabled through from the config', () => {
		const rows = rowsOf('a');
		const { result } = setup(rows, { label: 'Targets', id: 'targets', showWhen: () => true, enableWhen: () => false });
		expect(result.current.frameProps).toMatchObject({ id: 'targets', label: 'Targets' });
		expect(result.current.hidden).toBe(false);
		expect(result.current.disabled).toBe(true);
	});
});
