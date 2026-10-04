import { createEvent, fireEvent, render, screen } from '@testing-library/react';
import { type ReactNode, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { useTreeItem } from './hooks/useTreeItem';
import { useTreeRoot } from './hooks/useTreeRoot';
import { Tree } from './index';
import type { TreeSortable } from './types';

interface Row {
	id: string;
	children?: Array<Row>;
	/** Rows of another group, rendered in the row's body. */
	values?: Array<Row>;
}

const RowItem = ({ row }: { row: Row }) => (
	<Tree.Item id={row.id} testId={`item-${row.id}`}>
		<Tree.Item.Header>
			<Tree.Item.Handle label={`Move ${row.id}`} testId={`handle-${row.id}`} />
			<Tree.Item.Title testId={`title-${row.id}`}>{row.id}</Tree.Item.Title>
		</Tree.Item.Header>
		{row.values && (
			<Tree.Item.Body>
				<Tree.Root group="values" testId={`values-${row.id}`}>
					<Tree.Items>
						{row.values.map(value => (
							<RowItem key={value.id} row={value} />
						))}
					</Tree.Items>
				</Tree.Root>
			</Tree.Item.Body>
		)}
		{row.children && (
			<Tree.Item.Children testId={`children-${row.id}`}>
				<Tree.Items>
					{row.children.map(child => (
						<RowItem key={child.id} row={child} />
					))}
				</Tree.Items>
			</Tree.Item.Children>
		)}
	</Tree.Item>
);

const mount = (rows: Array<Row>, { sortable }: { sortable?: TreeSortable } = {}) =>
	render(
		<Tree.Root group="rows" sortable={sortable}>
			<Tree.Items>
				{rows.map(row => (
					<RowItem key={row.id} row={row} />
				))}
			</Tree.Items>
		</Tree.Root>,
	);

const flat = (...ids: Array<string>): Array<Row> => ids.map(id => ({ id }));

const item = (id: string) => screen.getByTestId(`item-${id}`);

// happy-dom lays nothing out, so each row is given a box: `index` rows of 20px from the top.
const place = (id: string, index: number) => vi.spyOn(item(id), 'getBoundingClientRect').mockReturnValue(new DOMRect(index * 20, index * 20, 20, 20));

const grab = (id: string) => {
	fireEvent.mouseDown(screen.getByTestId(`handle-${id}`));
	fireEvent.dragStart(item(id), { dataTransfer: new DataTransfer() });
};

// happy-dom's DragEvent is a plain Event, which drops `clientX`/`clientY` from its init.
const dragEvent = (type: 'dragOver' | 'drop', target: HTMLElement, at: { x?: number; y?: number }) => {
	const event = createEvent[type](target, { dataTransfer: new DataTransfer() });
	Object.defineProperty(event, 'clientX', { value: at.x ?? 0 });
	Object.defineProperty(event, 'clientY', { value: at.y ?? 0 });
	return fireEvent(target, event);
};
const dropOn = (target: HTMLElement, y = 0) => dragEvent('drop', target, { y });
/** Whether the target accepts the drag, i.e. cancelled the dragover. */
const accepts = (target: HTMLElement, y = 0) => !dragEvent('dragOver', target, { y });

const setup = () => {
	const onMove = vi.fn();
	return { onMove, sortable: { onMove } };
};

describe('Tree sortable', () => {
	it('moves an item before a row when dropped above its midpoint', () => {
		const { onMove, sortable } = setup();
		mount(flat('a', 'b', 'c'), { sortable });
		['a', 'b', 'c'].forEach(place);

		grab('c');
		dropOn(item('a'), 5);
		expect(onMove).toHaveBeenCalledTimes(1);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'rows', id: 'c', index: 2, parentId: null },
			to: { group: 'rows', parentId: null, index: 0 },
		});
	});

	it('moves an item after a row when dropped below its midpoint, with the index read before the item is removed', () => {
		const { onMove, sortable } = setup();
		mount(flat('a', 'b', 'c'), { sortable });
		['a', 'b', 'c'].forEach(place);

		grab('a');
		dropOn(item('b'), 35);
		expect(onMove).toHaveBeenCalledTimes(1);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'rows', id: 'a', index: 0, parentId: null },
			to: { group: 'rows', parentId: null, index: 2 },
		});
	});

	it('refuses a drop on the item itself or next to it, where nothing would move', () => {
		const { onMove, sortable } = setup();
		mount(flat('a', 'b', 'c'), { sortable });
		['a', 'b', 'c'].forEach(place);

		grab('b');
		expect(accepts(item('b'), 25)).toBe(false);
		expect(accepts(item('a'), 15)).toBe(false);
		expect(accepts(item('c'), 45)).toBe(false);
		for (const [id, y] of [
			['b', 25],
			['a', 15],
			['c', 45],
		] as const) {
			grab('b');
			dropOn(item(id), y);
		}
		expect(onMove).not.toHaveBeenCalled();
	});

	it("moves an item from one item's children into another's and calls onMove once", () => {
		const { onMove, sortable } = setup();
		mount(
			[
				{ id: 'A', children: flat('a1', 'a2') },
				{ id: 'B', children: flat('b1') },
			],
			{ sortable },
		);
		['A', 'a1', 'a2', 'B', 'b1'].forEach(place);

		grab('a1');
		expect(accepts(item('b1'), 95)).toBe(true);
		dropOn(item('b1'), 95);
		expect(onMove).toHaveBeenCalledTimes(1);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'rows', id: 'a1', index: 0, parentId: 'A' },
			to: { group: 'rows', parentId: 'B', index: 1 },
		});
	});

	it("drops into an open item's children area, at the position among its children", () => {
		const { onMove, sortable } = setup();
		mount([{ id: 'A', children: flat('a1', 'a2') }, { id: 'B' }], { sortable });
		['A', 'a1', 'a2', 'B'].forEach(place);

		grab('B');
		dropOn(screen.getByTestId('children-A'), 35);
		expect(onMove).toHaveBeenCalledTimes(1);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'rows', id: 'B', index: 1, parentId: null },
			to: { group: 'rows', parentId: 'A', index: 1 },
		});
	});

	it('refuses a drop into a descendant of the dragged item', () => {
		const { onMove, sortable } = setup();
		mount([{ id: 'A', children: [{ id: 'a1', children: flat('a11') }] }, { id: 'B' }], { sortable });

		const targets = [item('a1'), item('a11'), screen.getByTestId('children-A'), screen.getByTestId('children-a1')];
		grab('A');
		for (const target of targets) expect(accepts(target)).toBe(false);
		for (const target of targets) {
			grab('A');
			dropOn(target);
		}
		expect(onMove).not.toHaveBeenCalled();
	});

	it('refuses a drop into a list of another group', () => {
		const { onMove, sortable } = setup();
		mount([{ id: 'a', values: flat('v1', 'v2') }, { id: 'b' }], { sortable });

		grab('v1');
		expect(accepts(item('b'))).toBe(false);
		dropOn(item('b'));
		expect(onMove).not.toHaveBeenCalled();
	});

	const mountOuterAroundValues = (sortable: TreeSortable) => {
		mount([{ id: 'b' }, { id: 'a', values: flat('v1', 'v2') }], { sortable });
		place('b', 0);
		vi.spyOn(item('a'), 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 20, 20, 80));
		place('v1', 2);
		place('v2', 3);
	};

	it('hands a drop over a nested list of another group to the item around it', () => {
		const { onMove, sortable } = setup();
		mountOuterAroundValues(sortable);

		grab('b');
		expect(accepts(item('v2'), 70)).toBe(true);
		expect(item('a').getAttribute('data-drop')).toBe('after');
		expect(item('v2').hasAttribute('data-drag')).toBe(false);
		dropOn(item('v2'), 70);
		expect(onMove).toHaveBeenCalledTimes(1);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'rows', id: 'b', index: 0, parentId: null },
			to: { group: 'rows', parentId: null, index: 2 },
		});
	});

	it('refuses a drop over a nested list of another group when the item around it refuses too', () => {
		const onMove = vi.fn();
		mountOuterAroundValues({ onMove, canDrop: (from, to) => from.group === to.group && to.parentId !== null });

		grab('b');
		expect(accepts(item('v2'), 70)).toBe(false);
		expect(item('a').hasAttribute('data-drag')).toBe(false);
		dropOn(item('v2'), 70);
		expect(onMove).not.toHaveBeenCalled();
	});

	it('moves within a nested list of another group', () => {
		const { onMove, sortable } = setup();
		mount([{ id: 'a', values: flat('v1', 'v2') }], { sortable });
		['a', 'v1', 'v2'].forEach(place);

		grab('v1');
		dropOn(item('v2'), 55);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'values', id: 'v1', index: 0, parentId: 'a' },
			to: { group: 'values', parentId: 'a', index: 2 },
		});
	});

	it('drops into an empty list in another item’s body, marking the list', () => {
		const { onMove, sortable } = setup();
		mount(
			[
				{ id: 'a', values: flat('v1') },
				{ id: 'b', values: [] },
			],
			{ sortable },
		);
		place('v1', 1);

		grab('v1');
		expect(accepts(screen.getByTestId('values-b'), 0)).toBe(true);
		expect(screen.getByTestId('values-b').getAttribute('data-drop')).toBe('into');
		dropOn(screen.getByTestId('values-b'), 0);
		expect(onMove).toHaveBeenCalledTimes(1);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'values', id: 'v1', index: 0, parentId: 'a' },
			to: { group: 'values', parentId: 'b', index: 0 },
		});
		expect(screen.getByTestId('values-b').hasAttribute('data-drop')).toBe(false);
	});

	it('drops below the last item of a body list at the end of that list', () => {
		const { onMove, sortable } = setup();
		mount(
			[
				{ id: 'a', values: flat('v1', 'v2') },
				{ id: 'b', values: flat('w1') },
			],
			{ sortable },
		);
		place('v1', 0);
		place('v2', 1);
		place('w1', 2);

		grab('v1');
		dropOn(screen.getByTestId('values-b'), 100);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'values', id: 'v1', index: 0, parentId: 'a' },
			to: { group: 'values', parentId: 'b', index: 1 },
		});
	});

	it('refuses an item dropped into the list in its own body', () => {
		const { onMove, sortable } = setup();
		mount([{ id: 'a', values: [] }], { sortable: { ...sortable, canDrop: () => true } });

		grab('a');
		expect(accepts(screen.getByTestId('values-a'), 0)).toBe(false);
		dropOn(screen.getByTestId('values-a'), 0);
		expect(onMove).not.toHaveBeenCalled();
	});

	it('lets canDrop replace the group rule, but never allows a drop into a descendant', () => {
		const onMove = vi.fn();
		const canDrop = vi.fn(() => true);
		mount(
			[
				{ id: 'a', values: flat('v1') },
				{ id: 'B', children: flat('b1') },
			],
			{ sortable: { onMove, canDrop } },
		);
		['a', 'v1', 'B', 'b1'].forEach(place);

		grab('v1');
		dropOn(item('a'), 5);
		expect(onMove).toHaveBeenCalledWith({
			from: { group: 'values', id: 'v1', index: 0, parentId: 'a' },
			to: { group: 'rows', parentId: null, index: 0 },
		});
		expect(canDrop).toHaveBeenCalledWith({ group: 'values', id: 'v1', index: 0, parentId: 'a' }, { group: 'rows', parentId: null, index: 0 });

		onMove.mockClear();
		grab('B');
		dropOn(item('b1'), 65);
		expect(onMove).not.toHaveBeenCalled();
	});

	it('refuses a drop on the dragged item or next to it when canDrop allows everything, and keeps the parent out of it', () => {
		const onMove = vi.fn();
		mount([{ id: 'A', children: flat('a1', 'a2') }], { sortable: { onMove, canDrop: () => true } });
		['A', 'a1', 'a2'].forEach(place);

		grab('a1');
		expect(accepts(item('a1'), 25)).toBe(false);
		expect(item('A').hasAttribute('data-drag')).toBe(false);
		expect(accepts(item('a2'), 45)).toBe(false);
		for (const [id, y] of [
			['a1', 25],
			['a2', 45],
		] as const) {
			grab('a1');
			dropOn(item(id), y);
		}
		expect(onMove).not.toHaveBeenCalled();
	});

	it('refuses a same-group drop that canDrop turns down', () => {
		const onMove = vi.fn();
		mount(flat('a', 'b', 'c'), { sortable: { onMove, canDrop: (_from, to) => to.index !== 0 } });
		['a', 'b', 'c'].forEach(place);

		grab('c');
		expect(accepts(item('a'), 5)).toBe(false);
		dropOn(item('a'), 5);
		expect(onMove).not.toHaveBeenCalled();
		grab('c');
		dropOn(item('a'), 15);
		expect(onMove).toHaveBeenCalledWith(expect.objectContaining({ to: { group: 'rows', parentId: null, index: 1 } }));
	});

	it('marks the dragged item from and the hovered one to, with the drop position, and clears both on dragend', () => {
		const { sortable } = setup();
		mount([{ id: 'a' }, { id: 'b' }, { id: 'C', children: [] }], { sortable });
		['a', 'b', 'C'].forEach(place);

		grab('a');
		expect(item('a').getAttribute('data-drag')).toBe('from');
		dragEvent('dragOver', item('b'), { y: 35 });
		expect(item('b').getAttribute('data-drag')).toBe('to');
		expect(item('b').getAttribute('data-drop')).toBe('after');
		dragEvent('dragOver', screen.getByTestId('children-C'), { y: 0 });
		expect(item('b').hasAttribute('data-drag')).toBe(false);
		expect(item('C').getAttribute('data-drag')).toBe('to');
		expect(item('C').getAttribute('data-drop')).toBe('into');

		fireEvent.dragEnd(item('a'), { dataTransfer: new DataTransfer() });
		for (const id of ['a', 'b', 'C']) {
			expect(item(id).hasAttribute('data-drag')).toBe(false);
			expect(item(id).hasAttribute('data-drop')).toBe(false);
		}
	});

	it('clears the drag marks after a drop', () => {
		const { sortable } = setup();
		mount(flat('a', 'b'), { sortable });
		['a', 'b'].forEach(place);

		grab('a');
		dragEvent('dragOver', item('b'), { y: 35 });
		dropOn(item('b'), 35);
		expect(item('a').hasAttribute('data-drag')).toBe(false);
		expect(item('b').hasAttribute('data-drag')).toBe(false);
	});

	it('clears the hover mark when the pointer leaves the hovered item', () => {
		const { sortable } = setup();
		mount(flat('a', 'b'), { sortable });
		['a', 'b'].forEach(place);

		grab('a');
		dragEvent('dragOver', item('b'), { y: 35 });
		fireEvent.dragLeave(item('b'), { dataTransfer: new DataTransfer(), relatedTarget: document.body });
		expect(item('b').hasAttribute('data-drag')).toBe(false);
		expect(item('a').getAttribute('data-drag')).toBe('from');
	});

	it('keeps the hover mark when the pointer moves between parts of the hovered item', () => {
		const { sortable } = setup();
		mount(flat('a', 'b'), { sortable });
		['a', 'b'].forEach(place);

		grab('a');
		dragEvent('dragOver', item('b'), { y: 35 });
		const leave = createEvent.dragLeave(item('b'), { dataTransfer: new DataTransfer() });
		Object.defineProperty(leave, 'relatedTarget', { value: screen.getByTestId('title-b') });
		fireEvent(item('b'), leave);
		expect(item('b').getAttribute('data-drag')).toBe('to');
		expect(item('b').getAttribute('data-drop')).toBe('after');
	});

	it('turns an item draggable while its handle is pressed', () => {
		const { sortable } = setup();
		mount(flat('a', 'b'), { sortable });

		expect(item('a').hasAttribute('draggable')).toBe(false);
		fireEvent.mouseDown(screen.getByTestId('handle-a'));
		expect(item('a').getAttribute('draggable')).toBe('true');
		expect(item('b').hasAttribute('draggable')).toBe(false);
		fireEvent.mouseUp(document);
		expect(item('a').hasAttribute('draggable')).toBe(false);
	});

	it('arms an item on a press anywhere in it', () => {
		const { sortable } = setup();
		mount(flat('a'), { sortable });

		fireEvent.mouseDown(screen.getByTestId('title-a'));
		expect(item('a').getAttribute('draggable')).toBe('true');
	});

	it('arms only the innermost item on a press in a nested one', () => {
		const { sortable } = setup();
		mount([{ id: 'A', children: flat('a1') }], { sortable });

		fireEvent.mouseDown(screen.getByTestId('title-a1'));
		expect(item('a1').getAttribute('draggable')).toBe('true');
		expect(item('A').hasAttribute('draggable')).toBe(false);
	});

	it('does not arm an item on a press in a text field or on a disabled item', () => {
		const { sortable } = setup();
		render(
			<Tree.Root group="rows" sortable={sortable}>
				<Tree.Items>
					<Tree.Item id="a" testId="item-a">
						<Tree.Item.Body>
							<input data-testid="field-a" />
						</Tree.Item.Body>
					</Tree.Item>
					<Tree.Item id="b" testId="item-b" disabled>
						<Tree.Item.Title testId="title-b">b</Tree.Item.Title>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);

		fireEvent.mouseDown(screen.getByTestId('field-a'));
		expect(item('a').hasAttribute('draggable')).toBe(false);
		fireEvent.mouseDown(screen.getByTestId('title-b'));
		expect(item('b').hasAttribute('draggable')).toBe(false);
	});

	it('ignores a dragstart that bubbles up from a nested item', () => {
		const { onMove, sortable } = setup();
		mount([{ id: 'A', children: flat('a1', 'a2') }, { id: 'B' }], { sortable });
		['A', 'a1', 'a2', 'B'].forEach(place);

		grab('a2');
		expect(item('a2').getAttribute('data-drag')).toBe('from');
		expect(item('A').hasAttribute('data-drag')).toBe(false);
		dropOn(item('a1'), 25);
		expect(onMove).toHaveBeenCalledWith(expect.objectContaining({ from: { group: 'rows', id: 'a2', index: 1, parentId: 'A' } }));
	});

	it('does nothing without sortable', () => {
		mount([{ id: 'A', children: flat('a1') }, { id: 'B' }]);

		fireEvent.mouseDown(screen.getByTestId('handle-B'));
		expect(item('B').hasAttribute('draggable')).toBe(false);
		fireEvent.dragStart(item('B'), { dataTransfer: new DataTransfer() });
		expect(item('B').hasAttribute('data-drag')).toBe(false);
		expect(accepts(item('A'))).toBe(false);
		expect(accepts(screen.getByTestId('children-A'))).toBe(false);
	});

	it('does not re-render the root context consumers when the sortable object is recreated on each render', () => {
		const renders = vi.fn();
		const Probe = () => {
			renders(useTreeRoot().group, useTreeItem().id);
			return null;
		};
		const Host = ({ children }: { children: ReactNode }) => {
			const [count, setCount] = useState(0);
			return (
				<>
					<button type="button" data-testid="rerender" onClick={() => setCount(count + 1)} />
					<Tree.Root group="rows" sortable={{ onMove: () => {} }}>
						{children}
					</Tree.Root>
				</>
			);
		};
		render(
			<Host>
				<Tree.Items>
					<Tree.Item id="a">
						<Probe />
					</Tree.Item>
				</Tree.Items>
			</Host>,
		);
		renders.mockClear();

		fireEvent.click(screen.getByTestId('rerender'));
		expect(renders).not.toHaveBeenCalled();
	});

	it('hands onMove the handler of the latest render', () => {
		const first = vi.fn();
		const second = vi.fn();
		const { rerender } = render(
			<Tree.Root group="rows" sortable={{ onMove: first }}>
				<Tree.Items>
					{flat('a', 'b').map(row => (
						<RowItem key={row.id} row={row} />
					))}
				</Tree.Items>
			</Tree.Root>,
		);
		rerender(
			<Tree.Root group="rows" sortable={{ onMove: second }}>
				<Tree.Items>
					{flat('a', 'b').map(row => (
						<RowItem key={row.id} row={row} />
					))}
				</Tree.Items>
			</Tree.Root>,
		);
		['a', 'b'].forEach(place);

		grab('a');
		dropOn(item('b'), 35);
		expect(first).not.toHaveBeenCalled();
		expect(second).toHaveBeenCalledTimes(1);
	});
});
