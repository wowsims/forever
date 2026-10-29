import { GroupList } from '@features/apl/components/GroupList';
import { PriorityPane } from '@features/apl/components/PriorityPane';
import { aplClipToText } from '@features/apl/model/clipboard';
import { APLAction, APLGroup, APLListItem, APLPrepullAction, APLRotation, APLValue } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { dragEvent, FakeIntersectionObserver, place } from '@testing/apl/dom';
import { makePlayer } from '@testing/apl/player';
import { source } from '@testing/apl/source';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { toastManager } from '@ui-kit/Toast';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/proto/action_id', () => ({ ActionId: { replaceAllInString: async (str: string) => str } }));

const MOVE = 'common.list_picker.move_drag_drop';
const TOGGLE = 'common.list_picker.toggle_item';

const reset = (sequenceName: string) => APLAction.create({ action: { oneofKind: 'resetSequence', resetSequence: { sequenceName } } });
const row = (name: string) => APLListItem.create({ action: reset(name) });
const prepull = (name: string, doAt = '-2s') =>
	APLPrepullAction.create({ action: reset(name), doAtValue: { value: { oneofKind: 'const', const: { val: doAt } } } });
const sequenceNames = (list: Array<{ action?: APLAction }>) => list.map(item => (item.action?.action as any).resetSequence.sequenceName);

let player: ReturnType<typeof makePlayer>;
let notifications: number;

const setup = (rotation: Partial<APLRotation>) => {
	source.listeners.clear();
	notifications = 0;
	player = makePlayer(APLRotation.create(rotation), () => {
		notifications++;
		source.notify();
	});
};

const mount = (children = <PriorityPane />) =>
	render(<SimHostProvider host={{ player, rootElem: document.body, applyEmptyAplRotation: () => {} } as never}>{children}</SimHostProvider>);

const section = (testId: string) => document.querySelector(`[data-testid="${testId}"]`) as HTMLElement;
const rowsOf = (testId: string) => Array.from(section(testId).querySelectorAll<HTMLElement>('[data-testid="tree-item"]'));
const priorityRows = () => rowsOf('apl-priority-list-picker-root');
const prepullRows = () => rowsOf('apl-prepull-list-root');
const namesShown = () => Array.from(document.querySelectorAll<HTMLInputElement>('[data-kind="resetSequence"] input')).map(input => input.value);

/** A clipboard that reads back what was last written. */
const stubClipboard = () => {
	let clipboard = '';
	vi.stubGlobal('navigator', { clipboard: { writeText: async (text: string) => void (clipboard = text), readText: async () => clipboard } });
};

const grab = (element: HTMLElement) => {
	fireEvent.mouseDown(within(element).getAllByRole('button', { name: MOVE })[0]);
	fireEvent.dragStart(element, { dataTransfer: new DataTransfer() });
};

const accepts = (target: HTMLElement, y: number) => !dragEvent('dragOver', target, y);
const drop = (target: HTMLElement, y: number) =>
	act(() => {
		dragEvent('drop', target, y);
	});

// Async, so the popup's positioning settles inside `act`.
const openActions = async (element: HTMLElement) => {
	const trigger = within(element).getAllByTestId('tree-item-actions')[0];
	await act(async () => {
		trigger.focus();
		fireEvent.click(trigger);
	});
};
const openMenu = () => Array.from(document.querySelectorAll('[data-testid="tree-item-actions-popup"]')).find(popup => popup.hasAttribute('data-open'))!;
const clickAction = async (testId: string) => {
	await act(async () => {
		fireEvent.click(openMenu().querySelector(`[data-testid="${testId}"]`)!);
	});
};

beforeEach(() => {
	document.body.innerHTML = '';
	vi.stubGlobal('IntersectionObserver', FakeIntersectionObserver);
});

afterEach(() => {
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

describe('AplList on the priority pane', () => {
	it('renders one tree with the pre-pull and priority sections, one item per row', () => {
		setup({ prepullActions: [prepull('p')], priorityList: [row('a'), row('b')] });
		mount();

		expect(document.querySelectorAll('[data-testid="tree"][data-depth="0"]')).toHaveLength(1);
		expect(prepullRows()).toHaveLength(1);
		expect(priorityRows()).toHaveLength(2);
		expect(namesShown()).toEqual(['p', 'a', 'b']);
		expect(priorityRows()[0].querySelector('[data-testid="apl-list-item-picker-root"]')!.getAttribute('data-size')).toBe('xs');
	});

	it('hides a row through its eye with one notification, and keeps the row’s menu usable', () => {
		setup({ priorityList: [row('a')] });
		mount();

		act(() => {
			fireEvent.click(priorityRows()[0].querySelector('[data-testid="hide-picker-button"]')!);
		});

		expect(player.aplRotation.priorityList[0].hide).toBe(true);
		expect(notifications).toBe(1);
		expect(priorityRows()[0].querySelector('[data-testid="apl-list-item-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
		expect(within(priorityRows()[0]).getAllByTestId('tree-item-actions')[0].hasAttribute('disabled')).toBe(false);
	});

	it('copies a row next to itself and deletes one through the row’s menu', async () => {
		setup({ priorityList: [row('a'), row('b')] });
		mount();

		await openActions(priorityRows()[0]);
		await clickAction('apl-item-copy');
		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['a', 'a', 'b']);
		expect(player.aplRotation.priorityList[1]).not.toBe(player.aplRotation.priorityList[0]);

		await openActions(priorityRows()[2]);
		await clickAction('apl-item-delete');
		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['a', 'a']);
	});

	it('offers extract only for a row with a condition, and extracts it into a variable', async () => {
		const withCondition = row('a');
		withCondition.action!.condition = APLValue.create({ uuid: { value: 'c' }, value: { oneofKind: 'const', const: { val: '1' } } });
		setup({ priorityList: [withCondition, row('b')], valueVariables: [] });
		mount();

		await openActions(priorityRows()[1]);
		expect(openMenu().querySelector('[data-testid="apl-item-copy"]')).not.toBeNull();
		expect(openMenu().querySelector('[data-testid="apl-item-extract"]')).toBeNull();
		await act(async () => {
			fireEvent.keyDown(document.activeElement!, { key: 'Escape' });
		});

		await openActions(priorityRows()[0]);
		await clickAction('apl-item-extract');
		act(() => {
			fireEvent.change(screen.getByRole('textbox'), { target: { value: 'threshold' } });
		});
		act(() => {
			fireEvent.keyDown(screen.getByRole('textbox'), { key: 'Enter' });
		});

		expect(player.aplRotation.valueVariables.map(variable => variable.name)).toEqual(['threshold']);
		expect(player.aplRotation.priorityList[0].action!.condition!.value).toEqual({ oneofKind: 'variableRef', variableRef: { name: 'threshold' } });
	});

	it('appends a row through the toolbar’s new button', () => {
		setup({ priorityList: [] });
		mount();

		act(() => {
			fireEvent.click(document.querySelectorAll('[data-testid="apl-floating-action-bar-root"] button')[0]);
		});

		expect(player.aplRotation.priorityList).toHaveLength(1);
		expect(player.aplRotation.priorityList[0].action).toBeDefined();
	});

	it('keeps a collapsed row collapsed through an edit in another row and a reload of the list', () => {
		setup({ priorityList: [row('a'), row('b')] });
		mount();

		act(() => {
			fireEvent.click(within(priorityRows()[0]).getAllByRole('button', { name: TOGGLE })[0]);
		});
		expect(priorityRows()[0].hasAttribute('data-open')).toBe(false);

		act(() => {
			fireEvent.change(priorityRows()[1].querySelector('[data-kind="resetSequence"] input')!, { target: { value: 'typed' } });
		});
		act(() => {
			player.aplRotation.priorityList = player.aplRotation.priorityList.map(item => APLListItem.clone(item));
			source.notify();
		});

		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['a', 'typed']);
		expect(priorityRows()[0].hasAttribute('data-open')).toBe(false);
		expect(priorityRows()[1].hasAttribute('data-open')).toBe(true);
	});

	it('renders what a loaded rotation changed and keeps the rows it skipped editable', () => {
		setup({ priorityList: [row('a'), row('b')] });
		mount();
		expect(namesShown()).toEqual(['a', 'b']);

		act(() => {
			player.aplRotation = APLRotation.create({ priorityList: [row('a'), row('changed')] });
			source.notify();
		});
		expect(namesShown()).toEqual(['a', 'changed']);

		act(() => {
			fireEvent.change(document.querySelectorAll<HTMLInputElement>('[data-kind="resetSequence"] input')[0], { target: { value: 'typed' } });
		});
		expect(player.aplRotation.priorityList[0].action?.action).toEqual({ oneofKind: 'resetSequence', resetSequence: { sequenceName: 'typed' } });

		act(() => {
			player.aplRotation.priorityList[0].action!.action = { oneofKind: 'resetSequence', resetSequence: { sequenceName: 'edited' } };
			player.touchRotation();
		});
		expect(namesShown()).toEqual(['edited', 'changed']);
	});
});

describe('AplList drag and drop', () => {
	it('moves a row down with one notification, landing where it was dropped', () => {
		setup({ priorityList: [row('a'), row('b'), row('c'), row('d')] });
		mount();
		place(priorityRows());

		grab(priorityRows()[0]);
		drop(priorityRows()[2], 2 * 20 + 15);

		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['b', 'c', 'a', 'd']);
		expect(notifications).toBe(1);
		expect(namesShown()).toEqual(['b', 'c', 'a', 'd']);
	});

	it('edits a moved row through its own binding', () => {
		setup({ priorityList: [row('a'), row('b'), row('c')] });
		mount();
		place(priorityRows());
		const moved = player.aplRotation.priorityList[0];

		grab(priorityRows()[0]);
		drop(priorityRows()[2], 2 * 20 + 15);
		act(() => {
			fireEvent.change(priorityRows()[2].querySelector('[data-kind="resetSequence"] input')!, { target: { value: 'moved' } });
		});

		expect(player.aplRotation.priorityList[2]).toBe(moved);
		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['b', 'c', 'moved']);
	});

	it('refuses a pre-pull row timed before the pull into the priority list, keeping its time', () => {
		setup({ prepullActions: [prepull('p')], priorityList: [row('a')] });
		mount();
		place(prepullRows());
		place(priorityRows(), 1);

		grab(prepullRows()[0]);
		expect(accepts(priorityRows()[0], 20 + 5)).toBe(false);
		drop(priorityRows()[0], 20 + 5);

		expect(player.aplRotation.prepullActions).toHaveLength(1);
		expect(player.aplRotation.prepullActions[0].doAtValue?.value).toEqual({ oneofKind: 'const', const: { val: '-2s' } });
		expect(notifications).toBe(0);
	});

	it('moves a pre-pull row at the pull into the priority list, holding the same action', () => {
		setup({ prepullActions: [prepull('p', '-1s')], priorityList: [row('a'), row('b')] });
		mount();
		place(prepullRows());
		place(priorityRows(), 2);
		const action = player.aplRotation.prepullActions[0].action;

		grab(prepullRows()[0]);
		drop(priorityRows()[1], 3 * 20 + 5);

		expect(player.aplRotation.prepullActions).toEqual([]);
		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['a', 'p', 'b']);
		expect(player.aplRotation.priorityList[1].action).toBe(action);
		expect(notifications).toBe(1);
	});

	it('refuses a pre-pull drop for an action the pre-pull list does not offer', () => {
		const wait = APLListItem.create({ action: { action: { oneofKind: 'wait', wait: {} } } });
		setup({ prepullActions: [prepull('p')], priorityList: [wait] });
		mount();
		place(prepullRows());
		place(priorityRows(), 1);

		grab(priorityRows()[0]);
		expect(accepts(prepullRows()[0], 5)).toBe(false);
		drop(prepullRows()[0], 5);

		expect(player.aplRotation.prepullActions).toHaveLength(1);
		expect(player.aplRotation.priorityList).toHaveLength(1);
		expect(notifications).toBe(0);
	});

	it('moves a value from one row’s operand list into another row’s, keeping its uuid', () => {
		const constant = (uuid: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val: uuid } } });
		const and = (uuid: string, vals: Array<APLValue>) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'and', and: { vals } } });
		const withCondition = (condition: APLValue) => APLListItem.create({ action: { condition, action: reset('x').action } });
		setup({ priorityList: [withCondition(and('first', [constant('v1'), constant('v2')])), withCondition(and('second', [constant('v3')]))] });
		mount();
		const operands = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-value-list-item"]'));
		place(operands());

		fireEvent.mouseDown(within(operands()[0]).getAllByTestId('apl-value-list-handle')[0]);
		fireEvent.dragStart(operands()[0], { dataTransfer: new DataTransfer() });
		drop(operands()[2], 2 * 20 + 5);

		const vals = (index: number) =>
			(player.aplRotation.priorityList[index].action!.condition!.value as any).and.vals.map((value: APLValue) => value.uuid!.value);
		expect(vals(0)).toEqual(['v2']);
		expect(vals(1)).toEqual(['v1', 'v3']);
		expect(notifications).toBe(1);
	});

	it('reorders a nested operand list and leaves the list around it alone', () => {
		const constant = (uuid: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val: uuid } } });
		const list = (kind: 'and' | 'or', uuid: string, vals: Array<APLValue>) =>
			APLValue.create({ uuid: { value: uuid }, value: { oneofKind: kind, [kind]: { vals } } as APLValue['value'] });
		setup({
			priorityList: [
				APLListItem.create({
					action: { condition: list('and', 'outer', [list('or', 'o', [constant('x'), constant('y')]), constant('c')]), action: reset('r').action },
				}),
			],
		});
		mount();
		const [outerRows, innerRows] = [...document.querySelectorAll<HTMLElement>('[data-testid="apl-value-list"]')].map(element => [
			...element.querySelectorAll<HTMLElement>(':scope > ul > [data-testid="apl-value-list-item"]'),
		]);
		outerRows.forEach((el, i) => vi.spyOn(el, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, i * 100, 20, 100)));
		innerRows.forEach((el, i) => vi.spyOn(el, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 10 + i * 20, 20, 20)));

		fireEvent.mouseDown(within(innerRows[1]).getAllByTestId('apl-value-list-handle')[0]);
		fireEvent.dragStart(innerRows[1], { dataTransfer: new DataTransfer() });
		drop(innerRows[0], 12);

		const outer = player.aplRotation.priorityList[0].action!.condition!;
		expect((outer.value as any).and.vals.map((value: APLValue) => value.uuid!.value)).toEqual(['o', 'c']);
		expect(((outer.value as any).and.vals[0].value as any).or.vals.map((value: APLValue) => value.uuid!.value)).toEqual(['y', 'x']);
		expect(notifications).toBe(1);
	});

	it('drops a value into another row’s empty operand list', () => {
		const constant = (uuid: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val: uuid } } });
		const and = (uuid: string, vals: Array<APLValue>) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'and', and: { vals } } });
		const withCondition = (condition: APLValue) => APLListItem.create({ action: { condition, action: reset('x').action } });
		setup({ priorityList: [withCondition(and('first', [constant('v1')])), withCondition(and('second', []))] });
		mount();
		const [, empty] = Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-value-list"]'));
		const operand = document.querySelector<HTMLElement>('[data-testid="apl-value-list-item"]')!;

		fireEvent.mouseDown(within(operand).getAllByTestId('apl-value-list-handle')[0]);
		fireEvent.dragStart(operand, { dataTransfer: new DataTransfer() });
		expect(accepts(empty, 0)).toBe(true);
		drop(empty, 0);

		const vals = (index: number) =>
			(player.aplRotation.priorityList[index].action!.condition!.value as any).and.vals.map((value: APLValue) => value.uuid!.value);
		expect(vals(0)).toEqual([]);
		expect(vals(1)).toEqual(['v1']);
		expect(notifications).toBe(1);
	});

	it('reorders a sequence in a row a rotation load skipped', () => {
		const sequenceRow = () =>
			APLListItem.create({ action: { action: { oneofKind: 'sequence', sequence: { name: 'opener', actions: [reset('a'), reset('b')] } } } });
		setup({ priorityList: [sequenceRow(), row('x')] });
		mount();
		act(() => {
			player.aplRotation = APLRotation.create({ priorityList: [sequenceRow(), row('changed')] });
			source.notify();
		});
		notifications = 0;
		const steps = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-action-list-item"]'));
		place(steps());

		fireEvent.mouseDown(within(steps()[1]).getAllByTestId('apl-action-list-handle')[0]);
		fireEvent.dragStart(steps()[1], { dataTransfer: new DataTransfer() });
		drop(steps()[0], 5);

		const actions = (player.aplRotation.priorityList[0].action!.action as any).sequence.actions;
		expect(actions.map((action: APLAction) => (action.action as any).resetSequence.sequenceName)).toEqual(['b', 'a']);
		expect(notifications).toBe(1);
	});

	it('refuses a row dropped beside a section, outside every list', () => {
		setup({ prepullActions: [prepull('p')], priorityList: [row('a')] });
		mount();
		place(priorityRows(), 1);
		vi.spyOn(section('apl-prepull-list-root'), 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 0, 200, 20));

		grab(priorityRows()[0]);
		expect(accepts(section('apl-prepull-list-root'), 5)).toBe(false);
		drop(section('apl-prepull-list-root'), 5);

		expect(sequenceNames(player.aplRotation.priorityList)).toEqual(['a']);
		expect(notifications).toBe(0);
	});
});

describe('AplList on the action groups pane', () => {
	const groupItems = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-group-editor-root"]'));
	const actionRows = (group: HTMLElement) => within(group).queryAllByTestId('tree-item');

	it('moves an action from one group into another', () => {
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x'), row('y')] }), APLGroup.create({ name: 'two', actions: [row('z')] })] });
		mount(<GroupList />);
		const [one, two] = groupItems();
		place(actionRows(one));
		place(actionRows(two), 2);

		grab(actionRows(one)[0]);
		drop(actionRows(two)[0], 2 * 20 + 15);

		const [first, second] = player.aplRotation.groups;
		expect(sequenceNames(first.actions)).toEqual(['y']);
		expect(sequenceNames(second.actions)).toEqual(['z', 'x']);
		expect(notifications).toBe(1);
	});

	it('a group cannot be dropped into its own actions', () => {
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] }), APLGroup.create({ name: 'two', actions: [] })] });
		mount(<GroupList />);
		const [one] = groupItems();
		place(actionRows(one));
		const before = APLRotation.clone(player.aplRotation);

		grab(one);
		expect(accepts(actionRows(one)[0], 5)).toBe(false);
		drop(actionRows(one)[0], 5);

		expect(player.aplRotation).toEqual(before);
		expect(notifications).toBe(0);
	});

	it('reorders the groups themselves', () => {
		setup({ groups: [APLGroup.create({ name: 'one', actions: [] }), APLGroup.create({ name: 'two', actions: [] })] });
		mount(<GroupList />);
		place(groupItems());

		grab(groupItems()[1]);
		drop(groupItems()[0], 5);

		expect(player.aplRotation.groups.map(group => group.name)).toEqual(['two', 'one']);
	});

	const pasteFromToolbar = () =>
		act(async () => {
			fireEvent.click(within(screen.getByTestId('apl-floating-action-bar-root')).getByTestId('apl-paste'));
		});

	it('pastes a copied group under the name the dialog starts on, the next free one', async () => {
		stubClipboard();
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] })] });
		mount(<GroupList />);

		await openActions(groupItems()[0]);
		await clickAction('apl-item-clipboard');
		await pasteFromToolbar();
		const dialog = screen.getByRole('dialog');
		expect(within(dialog).getByRole('textbox')).toHaveProperty('value', 'one (2)');
		await act(async () => {
			fireEvent.click(within(dialog).getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.paste' }));
		});

		expect(player.aplRotation.groups.map(group => group.name)).toEqual(['one', 'one (2)']);
		expect(sequenceNames(player.aplRotation.groups[1].actions)).toEqual(['x']);
	});

	it('warns and adds nothing when the clipboard holds an action and the paste is for a group', async () => {
		stubClipboard();
		const toast = vi.spyOn(toastManager, 'add');
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] })] });
		mount(<GroupList />);

		await openActions(actionRows(groupItems()[0])[0]);
		await clickAction('apl-item-clipboard');
		await pasteFromToolbar();

		expect(toast).toHaveBeenLastCalledWith(expect.objectContaining({ variant: 'warning', body: 'rotation_tab.apl.clipboard.wrong_kind' }));
		expect(player.aplRotation.groups).toHaveLength(1);
		expect(screen.queryByRole('dialog')).toBeNull();
	});

	it("pastes an action copied from one group into another group's actions", async () => {
		stubClipboard();
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] }), APLGroup.create({ name: 'two', actions: [row('y')] })] });
		mount(<GroupList />);

		await openActions(actionRows(groupItems()[0])[0]);
		await clickAction('apl-item-clipboard');
		await act(async () => {
			fireEvent.click(within(groupItems()[1]).getByTestId('apl-paste'));
		});

		const [one, two] = player.aplRotation.groups;
		expect(sequenceNames(one.actions)).toEqual(['x']);
		expect(sequenceNames(two.actions)).toEqual(['y', 'x']);
	});

	it('takes the copied text in a dialog when the page cannot read the clipboard', async () => {
		vi.stubGlobal('navigator', {});
		const text = aplClipToText({ kind: 'action', item: row('z') });
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] })] });
		mount(<GroupList />);

		await act(async () => {
			fireEvent.click(within(groupItems()[0]).getByTestId('apl-paste'));
		});
		const dialog = screen.getByRole('dialog');
		fireEvent.change(within(dialog).getByRole('textbox'), { target: { value: text } });
		await act(async () => {
			fireEvent.click(within(dialog).getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.paste' }));
		});

		expect(sequenceNames(player.aplRotation.groups[0].actions)).toEqual(['x', 'z']);
	});

	it('goes on from the text dialog to the name dialog for a group', async () => {
		vi.stubGlobal('navigator', {});
		const text = aplClipToText({ kind: 'group', item: APLGroup.create({ name: 'one', actions: [row('z')] }) });
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] })] });
		mount(<GroupList />);

		await pasteFromToolbar();
		fireEvent.change(within(screen.getByRole('dialog')).getByRole('textbox'), { target: { value: text } });
		await act(async () => {
			fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.paste' }));
		});
		const dialog = screen.getByRole('dialog');
		expect(within(dialog).getByRole('textbox')).toHaveProperty('value', 'one (2)');
		await act(async () => {
			fireEvent.click(within(dialog).getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.paste' }));
		});

		expect(player.aplRotation.groups.map(group => group.name)).toEqual(['one', 'one (2)']);
		expect(sequenceNames(player.aplRotation.groups[1].actions)).toEqual(['z']);
	});

	it('adds an action to a group through the add button inside its children', () => {
		setup({ groups: [APLGroup.create({ name: 'one', actions: [] })] });
		mount(<GroupList />);

		act(() => {
			fireEvent.click(within(groupItems()[0]).getByTestId('tree-add'));
		});

		expect(player.aplRotation.groups[0].actions).toHaveLength(1);
	});

	it('drops an action into a group that has none, through its children area', () => {
		setup({ groups: [APLGroup.create({ name: 'one', actions: [row('x')] }), APLGroup.create({ name: 'two', actions: [] })] });
		mount(<GroupList />);
		const [one, two] = groupItems();
		place(actionRows(one));
		const empty = within(two).getByTestId('apl-group-actions');

		grab(actionRows(one)[0]);
		expect(accepts(empty, 0)).toBe(true);
		drop(empty, 0);

		const [first, second] = player.aplRotation.groups;
		expect(first.actions).toEqual([]);
		expect(sequenceNames(second.actions)).toEqual(['x']);
		expect(notifications).toBe(1);
	});
});
