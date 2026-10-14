import { APLRotation } from '@generated/proto/apl';
import { makePlayer } from '@testing/apl/player';
import { source } from '@testing/apl/source';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StrictMode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { GroupVariablesField } from './GroupVariablesField';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const placeholderAction = (name: string) => ({
	hide: false,
	notes: '',
	action: { value: { oneofKind: 'variablePlaceholder', variablePlaceholder: { name } } } as any,
});

let player: ReturnType<typeof makePlayer>;
let parentValue: { groupName?: string; variables: Array<any> };

/** `groupName` names a group in `rotation.groups`; the referenced group defines the placeholders. */
const setup = (opts: { groupName?: string; placeholders?: Array<string>; existingVariables?: Array<any>; variableNames?: Array<string> } = {}) => {
	source.listeners.clear();
	// Built as a plain object, not `APLRotation.create(...)`: `placeholderAction`'s `action.value`
	// shape is what `visitObjects`/`isPlaceholder` (see placeholders.test.ts) walk, and `.create()`
	// only knows the real `APLAction` field set — it would drop the fixture's `value` key.
	const rotation = {
		groups: opts.placeholders ? [{ name: 'g1', actions: opts.placeholders.map(placeholderAction), variables: [] }] : [],
		valueVariables: (opts.variableNames || []).map(name => ({ name })),
	} as unknown as APLRotation;
	player = makePlayer(rotation, () => source.notify());
	parentValue = { groupName: opts.groupName, variables: opts.existingVariables || [] };
};

const mount = ({ strict = false } = {}) => {
	const field = (
		<GroupVariablesField
			player={player as never}
			config={{ id: 'gv', getValue: () => parentValue.variables, setValue: () => player.touchRotation() } as never}
			groupNameField="groupName"
			getParentValue={() => parentValue}
		/>
	);
	return render(strict ? <StrictMode>{field}</StrictMode> : field);
};

const container = () => document.querySelector('[data-testid="group-reference-variables-container"]') as HTMLElement;
const rows = () => Array.from(document.querySelectorAll('[data-testid="apl-group-variable-picker-root"]'));
const labels = () => rows().map(row => row.querySelector('label')!.textContent);

const openMenu = (index: number) => {
	const item = document.querySelectorAll('[data-testid="apl-group-variables-item"]')[index];
	const trigger = item.querySelector<HTMLButtonElement>('[data-testid="tree-item-actions"]')!;
	trigger.focus();
	act(() => void fireEvent.click(trigger));
	return document.querySelector('[data-testid="tree-item-actions-popup"][data-open]') as HTMLElement;
};

const clickItemAction = (index: number, action: string) => {
	const button = openMenu(index).querySelector(action)!;
	act(() => void fireEvent.click(button));
};

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('GroupVariablesField', () => {
	it('derives one row per placeholder the selected group defines, labelled by name', () => {
		setup({ groupName: 'g1', placeholders: ['a', 'b'] });
		mount();

		expect(rows()).toHaveLength(2);
		expect(rows()[0].querySelector('label')!.textContent).toBe('a:');
		expect(rows()[1].querySelector('label')!.textContent).toBe('b:');
	});

	it.each([
		['with no group selected', { placeholders: ['a'] }],
		['when the selected group defines no placeholders', { groupName: 'g1', placeholders: [] }],
	])('is hidden %s', (_case, opts) => {
		setup(opts);
		mount();

		expect(container().hidden).toBe(true);
		expect(rows()).toHaveLength(0);
	});

	it('keeps an existing variables entry by identity, and creates the missing one', () => {
		const existingA = { name: 'a', value: { uuid: { value: 'keep-a' }, value: { oneofKind: 'variableRef', variableRef: { name: 'x' } } } };
		setup({ groupName: 'g1', placeholders: ['a', 'b'], existingVariables: [existingA] });
		mount();

		expect(parentValue.variables[0]).toBe(existingA);
		expect(parentValue.variables[1]).not.toBe(existingA);
		expect(parentValue.variables[1].name).toBe('b');
	});

	it('syncs parentValue.variables to exactly the derived rows after render', () => {
		setup({ groupName: 'g1', placeholders: ['a', 'b'], existingVariables: [] });
		mount();

		expect(parentValue.variables.map(entry => entry.name)).toEqual(['a', 'b']);
	});

	it('reconciles a deleted row back, so no label ever names another row entry', () => {
		const existingA = { name: 'a', value: { uuid: { value: 'keep-a' }, value: { oneofKind: 'variableRef', variableRef: { name: 'x' } } } };
		setup({ groupName: 'g1', placeholders: ['a', 'b'], existingVariables: [existingA] });
		mount();

		clickItemAction(0, '[data-testid="apl-group-variables-delete"]');

		expect(labels()).toEqual(['a:', 'b:']);
		expect(parentValue.variables.map(entry => entry.name)).toEqual(['a', 'b']);
		// The deleted row comes back empty: delete resets the assignment, it does not drop the row.
		expect(parentValue.variables[0]).not.toBe(existingA);
		expect(parentValue.variables[0].value.value.variableRef.name).toBe('');
	});

	it('offers delete and nothing else: no copy, because a duplicate entry is dropped by the next reconcile, and no add or drag, because the group fixes the rows', () => {
		setup({ groupName: 'g1', placeholders: ['a', 'b'] });
		mount();

		const menu = openMenu(0);
		expect([...menu.querySelectorAll('button')].map(button => button.getAttribute('data-testid'))).toEqual(['apl-group-variables-delete']);
		expect(container().querySelector('[data-testid="tree-add"]')).toBeNull();
		expect(container().querySelector('[aria-label="common.list_picker.move_drag_drop"]')).toBeNull();
	});

	it('titles an inline list with a caption that labels no control', () => {
		setup({ groupName: 'g1', placeholders: ['a'] });
		mount();

		const caption = container().querySelector('[data-testid="form-label"]')!;
		expect(caption.tagName).toBe('SPAN');
		expect(caption.textContent).toBe('rotation_tab.apl.actions.group_reference.variables');
		expect(container().querySelector('[data-testid="apl-group-variables"]')!.getAttribute('data-layout')).toBe('inline');
	});

	it('writes a variableRef and touches the rotation when a row picks a variable', async () => {
		setup({ groupName: 'g1', placeholders: ['a'], variableNames: ['x', 'y'] });
		const touched = vi.spyOn(player, 'touchRotation');
		mount();

		await userEvent.click(within(rows()[0] as HTMLElement).getByRole('combobox'));
		const items = screen.getAllByRole('option');
		expect(items.map(el => el.textContent)).toEqual(['x', 'y']);
		await userEvent.click(items[1]);

		expect(parentValue.variables[0].value).toEqual({ uuid: expect.anything(), value: { oneofKind: 'variableRef', variableRef: { name: 'y' } } });
		expect(touched).toHaveBeenCalled();
	});

	it('gives the row label an htmlFor naming the choice trigger', () => {
		setup({ groupName: 'g1', placeholders: ['a'] });
		mount();

		const label = rows()[0].querySelector('label') as HTMLLabelElement;
		const trigger = within(rows()[0] as HTMLElement).getByRole('combobox');
		expect(label.getAttribute('for')).toBe(trigger.id);
		expect(label.getAttribute('for')).toBeTruthy();
	});
});

// The reference outlives what it points at: deleting a group, or renaming one, leaves a reference
// naming nothing, and a selected group can hold an existing variableRef assignment while declaring
// no placeholder for it. None of those may cost the user the assignments they typed.
describe('GroupVariablesField — nothing to derive from', () => {
	const assigned = () => ({
		name: 'TargetDemonOrUndead',
		value: { uuid: { value: 'keep' }, value: { oneofKind: 'variableRef', variableRef: { name: 'No' } } },
	});

	it.each([
		['with no group selected', { placeholders: ['a'] }],
		['when the named group no longer exists', { groupName: 'deleted', placeholders: ['a'] }],
		['when the selected group declares no placeholder', { groupName: 'g1', placeholders: [] }],
	])('leaves the stored variables alone %s', (_case, opts) => {
		const existing = [assigned()];
		setup({ ...opts, existingVariables: existing });
		mount();

		expect(parentValue.variables).toBe(existing);
		expect(parentValue.variables[0].value.value.variableRef.name).toBe('No');
		expect(container().hidden).toBe(true);
		expect(rows()).toHaveLength(0);
	});

	it('still rewrites the list once the group does declare placeholders', () => {
		setup({ groupName: 'g1', placeholders: ['a'], existingVariables: [assigned()] });
		mount();

		expect(parentValue.variables.map(entry => entry.name)).toEqual(['a']);
	});

	it('reconciles on read without writing to the store, on mount and on every re-read', () => {
		setup({ groupName: 'g1', placeholders: ['a', 'b'] });
		const touched = vi.spyOn(player, 'touchRotation');
		const modified = vi.spyOn(player, 'modifyAplRotation');
		mount();
		act(() => source.notify());

		expect(parentValue.variables.map(entry => entry.name)).toEqual(['a', 'b']);
		expect(touched).not.toHaveBeenCalled();
		expect(modified).not.toHaveBeenCalled();
	});
});

describe('GroupVariablesField under StrictMode (bug 2 — reconcile is a write on a read path)', () => {
	it('mounts with exactly one row per placeholder, no duplicates from the double-invoked read', () => {
		setup({ groupName: 'g1', placeholders: ['a', 'b'] });
		mount({ strict: true });

		expect(labels()).toEqual(['a:', 'b:']);
		expect(parentValue.variables.map(entry => entry.name)).toEqual(['a', 'b']);
	});

	it('keeps existing entries by identity through the StrictMode double render', () => {
		const existingA = { name: 'a', value: { uuid: { value: 'keep-a' }, value: { oneofKind: 'variableRef', variableRef: { name: 'x' } } } };
		setup({ groupName: 'g1', placeholders: ['a', 'b'], existingVariables: [existingA] });
		mount({ strict: true });

		expect(parentValue.variables[0]).toBe(existingA);
		expect(parentValue.variables).toHaveLength(2);
	});

	it('reconciles a delete correctly under StrictMode, with no duplicate or dropped rows', () => {
		const existingA = { name: 'a', value: { uuid: { value: 'keep-a' }, value: { oneofKind: 'variableRef', variableRef: { name: 'x' } } } };
		setup({ groupName: 'g1', placeholders: ['a', 'b'], existingVariables: [existingA] });
		mount({ strict: true });

		clickItemAction(0, '[data-testid="apl-group-variables-delete"]');

		expect(labels()).toEqual(['a:', 'b:']);
		expect(parentValue.variables.map(entry => entry.name)).toEqual(['a', 'b']);
	});
});

describe('GroupVariableRow — a row only exists while its group is selected', () => {
	it('deselecting the group unmounts every row rather than leaving one mounted', () => {
		setup({ groupName: 'g1', placeholders: ['a', 'b'] });
		mount();
		expect(rows()).toHaveLength(2);

		act(() => {
			parentValue.groupName = undefined;
			player.touchRotation();
		});

		expect(rows()).toHaveLength(0);
	});

	it('reselecting the group mounts fresh row instances, which start with stale=true and read correctly on their first render', () => {
		setup({ groupName: 'g1', placeholders: ['a'] });
		mount();
		expect(rows()).toHaveLength(1);

		act(() => {
			parentValue.groupName = undefined;
			player.touchRotation();
		});
		expect(rows()).toHaveLength(0);

		act(() => {
			parentValue.groupName = 'g1';
			player.touchRotation();
		});

		expect(rows()).toHaveLength(1);
	});
});
