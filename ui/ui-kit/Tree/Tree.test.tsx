import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { type ReactNode, useSyncExternalStore } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { useTreeItem } from './hooks/useTreeItem';
import { useTreeRoot } from './hooks/useTreeRoot';
import { Tree } from './index';

const ItemProbe = () => {
	const { id, index, parentId } = useTreeItem();
	return <span data-testid={`probe-${id}`}>{`${index}:${parentId ?? '-'}`}</span>;
};

const RootProbe = ({ testId }: { testId: string }) => {
	const { group, depth, density, disabled } = useTreeRoot();
	return <span data-testid={testId}>{`${group}:${depth}:${density}:${disabled}`}</span>;
};

const probe = (id: string) => screen.getByTestId(`probe-${id}`).textContent;

const Row = ({ id, children }: { id: string; children?: ReactNode }) => (
	<Tree.Item id={id} testId={`item-${id}`}>
		<Tree.Item.Header>
			<Tree.Item.Title>
				<ItemProbe />
			</Tree.Item.Title>
		</Tree.Item.Header>
		{children}
	</Tree.Item>
);

const createOrder = (initial: Array<string>) => {
	let ids = initial;
	const listeners = new Set<() => void>();
	return {
		get: () => ids,
		set: (next: Array<string>) => {
			ids = next;
			listeners.forEach(listener => listener());
		},
		subscribe: (listener: () => void) => {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
	};
};

// Subscribes on its own, so a reorder re-renders this component and not the Items around it.
const StoreRows = ({ order }: { order: ReturnType<typeof createOrder> }) => {
	const ids = useSyncExternalStore(order.subscribe, order.get);
	return ids.map(id => <Row key={id} id={id} />);
};

const Collapsible = () => (
	<Tree.Root>
		<Tree.Items>
			<Tree.Item id="a">
				<Tree.Item.Header>
					<Tree.Item.Title>Row A</Tree.Item.Title>
					<Tree.Item.Trigger label="Toggle row A" />
				</Tree.Item.Header>
				<Tree.Item.Body testId="body-a">Body A</Tree.Item.Body>
			</Tree.Item>
		</Tree.Items>
	</Tree.Root>
);

const trigger = () => screen.getByRole('button', { name: 'Toggle row A' });

describe('Tree', () => {
	it('renders a list of list items, with the depth of each nested root', () => {
		render(
			<Tree.Root group="rotation">
				<Tree.Items>
					<Row id="a">
						<Tree.Item.Children testId="children-a">
							<Tree.Items>
								<Row id="a1">
									<Tree.Item.Children testId="children-a1">
										<Tree.Items>
											<Row id="a1x" />
										</Tree.Items>
									</Tree.Item.Children>
								</Row>
							</Tree.Items>
						</Tree.Item.Children>
					</Row>
				</Tree.Items>
			</Tree.Root>,
		);

		const roots = screen.getAllByTestId('tree');
		expect(roots.map(root => root.getAttribute('data-depth'))).toEqual(['0', '1', '2']);
		const lists = screen.getAllByRole('list');
		expect(lists.map(list => list.tagName)).toEqual(['UL', 'UL', 'UL']);
		expect(screen.getAllByRole('listitem').map(item => item.tagName)).toEqual(['LI', 'LI', 'LI']);
		expect(within(screen.getByTestId('children-a')).getAllByTestId('tree')[0].getAttribute('data-depth')).toBe('1');
		expect(probe('a')).toBe('0:-');
		expect(probe('a1')).toBe('0:a');
		expect(probe('a1x')).toBe('0:a1');
	});

	it('gives a nested root the group, density and disabled state of its parent', () => {
		render(
			<Tree.Root group="rotation" density="compact" disabled>
				<RootProbe testId="outer" />
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Children>
							<RootProbe testId="inner" />
						</Tree.Item.Children>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		expect(screen.getByTestId('outer').textContent).toBe('rotation:0:compact:true');
		expect(screen.getByTestId('inner').textContent).toBe('rotation:1:compact:true');
		expect(screen.getAllByTestId('tree')[1].getAttribute('data-density')).toBe('compact');
	});

	it('puts density, layout and orientation on the root as data attributes', () => {
		const { rerender } = render(
			<Tree.Root>
				<Tree.Items>{null}</Tree.Items>
			</Tree.Root>,
		);
		const root = screen.getByTestId('tree');
		expect(root.getAttribute('data-depth')).toBe('0');
		expect(root.getAttribute('data-density')).toBe('default');
		expect(root.getAttribute('data-layout')).toBe('default');
		expect(root.getAttribute('data-orientation')).toBe('vertical');

		rerender(
			<Tree.Root density="compact" layout="inline">
				<Tree.Items>{null}</Tree.Items>
			</Tree.Root>,
		);
		expect(root.getAttribute('data-density')).toBe('compact');
		expect(root.getAttribute('data-layout')).toBe('inline');
	});

	// The index comes from DOM order, so wrappers, fragments and a wrapper that reorders or drops
	// its rows on a store notification all give the right index.
	it('indexes items rendered through wrappers', () => {
		const order = createOrder(['c', 'd', 'e']);
		render(
			<Tree.Root>
				<Tree.Items>
					<Row id="a" />
					<>
						<Row id="b" />
					</>
					<StoreRows order={order} />
					<Row id="f" />
				</Tree.Items>
			</Tree.Root>,
		);
		const indices = () => ['a', 'b', 'c', 'd', 'e', 'f'].map(id => screen.queryByTestId(`probe-${id}`)?.textContent ?? null);
		expect(indices()).toEqual(['0:-', '1:-', '2:-', '3:-', '4:-', '5:-']);

		act(() => order.set(['e', 'c', 'd']));
		expect(screen.getAllByRole('listitem').map(item => item.getAttribute('data-testid'))).toEqual([
			'item-a',
			'item-b',
			'item-e',
			'item-c',
			'item-d',
			'item-f',
		]);
		expect(indices()).toEqual(['0:-', '1:-', '3:-', '4:-', '2:-', '5:-']);

		act(() => order.set(['e', 'd']));
		expect(indices()).toEqual(['0:-', '1:-', null, '3:-', '2:-', '4:-']);

		act(() => order.set(['x', 'e', 'd']));
		expect(probe('x')).toBe('2:-');
		expect(indices()).toEqual(['0:-', '1:-', null, '4:-', '3:-', '5:-']);
	});

	it('toggles an item from its trigger, and hides the body while closed', () => {
		render(<Collapsible />);
		const item = screen.getByRole('listitem');
		const body = screen.getByTestId('body-a');
		expect(item.hasAttribute('data-open')).toBe(true);
		expect(trigger().getAttribute('aria-expanded')).toBe('true');
		expect(trigger().getAttribute('aria-controls')).toBe(body.id);
		expect(body.hidden).toBe(false);

		fireEvent.click(trigger());
		expect(item.hasAttribute('data-open')).toBe(false);
		expect(trigger().getAttribute('aria-expanded')).toBe('false');
		expect(body.hidden).toBe(true);
		expect(trigger().getAttribute('aria-controls')).toBe(body.id);

		fireEvent.click(trigger());
		expect(item.hasAttribute('data-open')).toBe(true);
		expect(body.hidden).toBe(false);
	});

	it('hides the children with the body when an item closes', () => {
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Header>
							<Tree.Item.Trigger label="Toggle row A" />
						</Tree.Item.Header>
						<Tree.Item.Children testId="children-a">
							<Tree.Items>
								<Row id="a1" />
							</Tree.Items>
						</Tree.Item.Children>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		expect(screen.getByTestId('children-a').hidden).toBe(false);

		fireEvent.click(trigger());
		expect(screen.getByTestId('children-a').hidden).toBe(true);
	});

	it('opens the actions menu, runs an action and closes', () => {
		const onDelete = vi.fn();
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Header>
							<Tree.Item.Title>Row A</Tree.Item.Title>
							<Tree.Item.Actions label="Row A actions">
								<Tree.Item.Action icon="copy" label="Copy" />
								<Tree.Item.Action icon="xmark" label="Delete" tone="danger" onClick={onDelete} testId="delete-a" />
							</Tree.Item.Actions>
						</Tree.Item.Header>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		const actions = screen.getByRole('button', { name: 'Row A actions' });
		expect(actions.getAttribute('data-testid')).toBe('tree-item-actions');
		actions.focus();
		fireEvent.click(actions);
		const menu = screen.getByRole('dialog');
		expect(within(menu).getByRole('button', { name: 'Copy' }).getAttribute('data-testid')).toBe('tree-item-action');
		const remove = within(menu).getByRole('button', { name: 'Delete' });
		expect(remove.getAttribute('data-tone')).toBe('danger');
		expect(remove.getAttribute('data-testid')).toBe('delete-a');

		fireEvent.click(remove);
		expect(onDelete).toHaveBeenCalledTimes(1);
		expect(menu.hasAttribute('data-open')).toBe(false);
	});

	it('runs an action rendered outside a menu', () => {
		const onHide = vi.fn();
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Action icon="eye-slash" label="Hide" onClick={onHide} />
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		fireEvent.click(screen.getByRole('button', { name: 'Hide' }));
		expect(onHide).toHaveBeenCalledTimes(1);
	});

	it('calls the add button handler, with a text label or an icon and an accessible name', () => {
		const onAdd = vi.fn();
		const { rerender } = render(
			<Tree.Root>
				<Tree.AddButton label="New action" onClick={onAdd} />
			</Tree.Root>,
		);
		const add = screen.getByRole('button', { name: 'New action' });
		expect(add.getAttribute('data-testid')).toBe('tree-add');
		expect(add.textContent).toBe('New action');
		fireEvent.click(add);
		expect(onAdd).toHaveBeenCalledTimes(1);

		rerender(
			<Tree.Root>
				<Tree.AddButton label="New action" icon="plus" onClick={onAdd} />
			</Tree.Root>,
		);
		const iconAdd = screen.getByRole('button', { name: 'New action' });
		expect(iconAdd.getAttribute('aria-label')).toBe('New action');
		expect(iconAdd.textContent).toBe('');
		fireEvent.click(iconAdd);
		expect(onAdd).toHaveBeenCalledTimes(2);
	});

	it('disables the item controls and marks the items of a disabled root', () => {
		const onAdd = vi.fn();
		render(
			<Tree.Root disabled>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Header>
							<Tree.Item.Handle label="Move row A" />
							<Tree.Item.Trigger label="Toggle row A" />
							<Tree.Item.Actions label="Row A actions">
								<Tree.Item.Action label="Delete" />
							</Tree.Item.Actions>
						</Tree.Item.Header>
					</Tree.Item>
				</Tree.Items>
				<Tree.AddButton label="New action" onClick={onAdd} />
			</Tree.Root>,
		);
		expect(screen.getByRole('listitem').hasAttribute('data-disabled')).toBe(true);
		for (const name of ['Move row A', 'Toggle row A', 'Row A actions', 'New action']) {
			expect((screen.getByRole('button', { name }) as HTMLButtonElement).disabled).toBe(true);
		}
	});

	it('marks a muted item and keeps its handle and trigger usable', () => {
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a" muted>
						<Tree.Item.Header>
							<Tree.Item.Handle label="Move row A" />
							<Tree.Item.Trigger label="Toggle row A" />
						</Tree.Item.Header>
					</Tree.Item>
					<Tree.Item id="b">
						<Tree.Item.Header>
							<Tree.Item.Handle label="Move row B" />
						</Tree.Item.Header>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		const [a, b] = screen.getAllByRole('listitem');
		expect(a.hasAttribute('data-muted')).toBe(true);
		expect(a.hasAttribute('data-disabled')).toBe(false);
		expect(b.hasAttribute('data-muted')).toBe(false);
		for (const name of ['Move row A', 'Toggle row A']) {
			expect((screen.getByRole('button', { name }) as HTMLButtonElement).disabled).toBe(false);
		}
	});

	it('marks a disabled item without disabling its siblings', () => {
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a" disabled testId="item-a" />
					<Tree.Item id="b" testId="item-b" />
				</Tree.Items>
			</Tree.Root>,
		);
		expect(screen.getByTestId('item-a').hasAttribute('data-disabled')).toBe(true);
		expect(screen.getByTestId('item-b').hasAttribute('data-disabled')).toBe(false);
	});

	it('forwards testId on every part', () => {
		render(
			<Tree.Root testId="t-root">
				<Tree.Items testId="t-items">
					<Tree.Item id="a" testId="t-item">
						<Tree.Item.Header testId="t-header">
							<Tree.Item.Handle label="Move" testId="t-handle" />
							<Tree.Item.Title testId="t-title" as="h6">
								Title
							</Tree.Item.Title>
							<Tree.Item.Aside testId="t-aside">Badge</Tree.Item.Aside>
							<Tree.Item.Trigger label="Toggle" testId="t-trigger" />
							<Tree.Item.Actions label="Actions" testId="t-actions">
								<Tree.Item.Action label="Delete" testId="t-action" />
							</Tree.Item.Actions>
						</Tree.Item.Header>
						<Tree.Item.Body testId="t-body">Body</Tree.Item.Body>
						<Tree.Item.Children testId="t-children">
							<Tree.Items>{null}</Tree.Items>
						</Tree.Item.Children>
					</Tree.Item>
				</Tree.Items>
				<Tree.AddButton label="Add" testId="t-add" onClick={() => {}} />
			</Tree.Root>,
		);
		fireEvent.click(screen.getByTestId('t-actions'));
		const parts = ['root', 'items', 'item', 'header', 'handle', 'title', 'aside', 'trigger', 'actions', 'action', 'body', 'children', 'add'];
		for (const part of parts) {
			expect(screen.queryByTestId(`t-${part}`), part).not.toBeNull();
		}
		expect(screen.getByTestId('t-title').tagName).toBe('H6');
	});

	it('gives the parts their default testids', () => {
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Header>
							<Tree.Item.Title>Title</Tree.Item.Title>
							<Tree.Item.Actions label="Actions">
								<Tree.Item.Action label="Delete" />
							</Tree.Item.Actions>
						</Tree.Item.Header>
					</Tree.Item>
				</Tree.Items>
				<Tree.AddButton label="Add" onClick={() => {}} />
			</Tree.Root>,
		);
		fireEvent.click(screen.getByRole('button', { name: 'Actions' }));
		for (const testId of ['tree', 'tree-item', 'tree-item-header', 'tree-item-actions', 'tree-item-action', 'tree-add']) {
			expect(screen.getAllByTestId(testId).length, testId).toBe(1);
		}
		expect(screen.getByText('Title').tagName).toBe('SPAN');
	});

	// The item grid places each part by its own grid area, so the header renders only the parts it
	// is given: an omitted Handle leaves no element behind to hold a track open.
	it('renders only the header parts it is given, directly inside the header', () => {
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Header>
							<Tree.Item.Title>Title</Tree.Item.Title>
							<Tree.Item.Trigger label="Toggle" />
						</Tree.Item.Header>
						<Tree.Item.Body>Body</Tree.Item.Body>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		const header = screen.getByTestId('tree-item-header');
		expect(Array.from(header.children)).toEqual([screen.getByText('Title'), screen.getByRole('button', { name: 'Toggle' })]);
		expect(screen.getByRole('listitem').children[0]).toBe(header);
	});

	it('marks a root in a body as a list of its own, and not a root in children or one nested in its items', () => {
		render(
			<Tree.Root>
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Body>
							<Tree.Root group="values" testId="body-list">
								<Tree.Items>
									<Tree.Item id="v">
										<Tree.Item.Children testId="value-children">
											<Tree.Items>{null}</Tree.Items>
										</Tree.Item.Children>
									</Tree.Item>
								</Tree.Items>
							</Tree.Root>
						</Tree.Item.Body>
						<Tree.Item.Children testId="a-children">
							<Tree.Items>{null}</Tree.Items>
						</Tree.Item.Children>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		expect(screen.getByTestId('body-list').hasAttribute('data-in-body')).toBe(true);
		expect(within(screen.getByTestId('a-children')).getByTestId('tree').hasAttribute('data-in-body')).toBe(false);
		expect(within(screen.getByTestId('value-children')).getByTestId('tree').hasAttribute('data-in-body')).toBe(false);
		expect(screen.getAllByTestId('tree')[0].hasAttribute('data-in-body')).toBe(false);
	});

	it('points the trigger at the parts the item mounts: body, children, both or none', () => {
		const Toggle = ({ id, body, kids }: { id: string; body?: boolean; kids?: boolean }) => (
			<Tree.Item id={id}>
				<Tree.Item.Header>
					<Tree.Item.Trigger label={`Toggle ${id}`} />
				</Tree.Item.Header>
				{body && <Tree.Item.Body testId={`body-${id}`}>Body</Tree.Item.Body>}
				{kids && (
					<Tree.Item.Children testId={`children-${id}`}>
						<Tree.Items>{null}</Tree.Items>
					</Tree.Item.Children>
				)}
			</Tree.Item>
		);
		render(
			<Tree.Root>
				<Tree.Items>
					<Toggle id="b" body />
					<Toggle id="c" kids />
					<Toggle id="both" body kids />
					<Toggle id="none" />
				</Tree.Items>
			</Tree.Root>,
		);
		const controls = (id: string) => screen.getByRole('button', { name: `Toggle ${id}` }).getAttribute('aria-controls');
		expect(controls('b')).toBe(screen.getByTestId('body-b').id);
		expect(controls('c')).toBe(screen.getByTestId('children-c').id);
		expect(controls('both')).toBe(`${screen.getByTestId('body-both').id} ${screen.getByTestId('children-both').id}`);
		expect(controls('none')).toBeNull();
		for (const id of ['b', 'c', 'both'])
			expect(
				controls(id)!
					.split(' ')
					.every(part => document.getElementById(part)),
			).toBe(true);
	});

	it('gives a root in a body without a group of its own a group apart from its parent', () => {
		const GroupProbe = ({ testId }: { testId: string }) => <span data-testid={testId}>{useTreeRoot().group}</span>;
		render(
			<Tree.Root group="rows">
				<Tree.Items>
					<Tree.Item id="a">
						<Tree.Item.Body>
							<Tree.Root>
								<GroupProbe testId="body-group" />
							</Tree.Root>
						</Tree.Item.Body>
						<Tree.Item.Children>
							<GroupProbe testId="children-group" />
						</Tree.Item.Children>
					</Tree.Item>
				</Tree.Items>
			</Tree.Root>,
		);
		expect(screen.getByTestId('children-group').textContent).toBe('rows');
		expect(screen.getByTestId('body-group').textContent).not.toBe('rows');
		expect(screen.getByTestId('body-group').textContent).toBeTruthy();
	});

	it('puts a plus before a text add button and marks the success tone', () => {
		render(
			<Tree.Root>
				<Tree.AddButton label="New row" onClick={() => {}} />
				<Tree.AddButton label="New value" icon="plus" tone="success" testId="icon-add" onClick={() => {}} />
			</Tree.Root>,
		);
		const text = screen.getByRole('button', { name: 'New row' });
		expect(text.querySelector('svg')?.getAttribute('data-icon')).toBe('plus');
		expect(text.hasAttribute('data-tone')).toBe(false);
		expect(screen.getByTestId('icon-add').getAttribute('data-tone')).toBe('success');
	});

	it('throws when a part is rendered outside its root or item', () => {
		const error = vi.spyOn(console, 'error').mockImplementation(() => {});
		expect(() => render(<Tree.Items>{null}</Tree.Items>)).toThrow(/Tree\.Root/);
		expect(() => render(<Tree.Item.Body>Body</Tree.Item.Body>)).toThrow(/Tree\.Item/);
		expect(() =>
			render(
				<Tree.Root>
					<Tree.Item id="a" />
				</Tree.Root>,
			),
		).toThrow(/Tree\.Items/);
		error.mockRestore();
	});
});
