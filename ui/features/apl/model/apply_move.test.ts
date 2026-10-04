import { APLAction, APLGroup, APLListItem, APLPrepullAction, APLRotation, APLValue } from '@generated/proto/apl';
import { describe, expect, it } from 'vitest';

import {
	actionListGroup,
	APL_GROUPS,
	APL_ROWS,
	APL_VARIABLES,
	aplItemId,
	type AplMove,
	applyMove,
	canApplyMove,
	PREPULL,
	PRIORITY,
	resolveMoveSource,
	valueListGroup,
} from './apply_move';

// `create` copies every message it is handed, so nested messages are assigned to keep the test's references.
const build = <T extends object>(message: T, parts: Partial<T>): T => Object.assign(message, parts);
const rotationOf = (parts: Partial<APLRotation>) => build(APLRotation.create(), parts);
const reset = (sequenceName: string) => APLAction.create({ action: { oneofKind: 'resetSequence', resetSequence: { sequenceName } } });
const listItem = (action: APLAction) => build(APLListItem.create(), { action });
const row = (name: string) => listItem(reset(name));
const prepull = (name: string, doAt = '-3s') =>
	build(APLPrepullAction.create({ doAtValue: { value: { oneofKind: 'const', const: { val: doAt } } } }), { action: reset(name) });
const group = (name: string, actions: Array<APLListItem>) => build(APLGroup.create({ name }), { actions });
const condition = (value: APLValue) => build(APLAction.create(), { condition: value });
const names = (list: Array<{ action?: APLAction }>) => list.map(item => (item.action?.action as any).resetSequence.sequenceName);

const constant = (val: string, uuid: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val } } });
const valsOf = (value: APLValue) => (value.value as any).and.vals as Array<APLValue>;
const and = (uuid: string, vals: Array<APLValue>) => {
	const value = APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'and', and: { vals: [] } } });
	(value.value as any).and.vals = vals;
	return value;
};
const sequenceOf = (actions: Array<APLAction>) => {
	const action = APLAction.create({ action: { oneofKind: 'sequence', sequence: { name: 'opener', actions: [] } } });
	(action.action as any).sequence.actions = actions;
	return action;
};

/** The move the tree reports for `list[from]` dropped at `to` in the list `(toGroup, toParent)`. */
const move = (
	list: Array<unknown>,
	from: number,
	group: string,
	parentId: string | null,
	to: number,
	toGroup = group,
	toParent: string | null = parentId,
): AplMove => ({
	from: { group, id: aplItemId(list[from], from), index: from, parentId },
	to: { group: toGroup, parentId: toParent, index: to },
});

describe('applyMove', () => {
	describe('within one list, with the drop index read before the item leaves', () => {
		it.each([
			['down: dropped after the third row, it lands third', ['a', 'b', 'c', 'd'], 0, 3, ['b', 'c', 'a', 'd']],
			['up', ['a', 'b', 'c', 'd'], 3, 1, ['a', 'd', 'b', 'c']],
			['to the end', ['a', 'b', 'c'], 0, 3, ['b', 'c', 'a']],
		])('moves a row %s', (_case, start, from, to, expected) => {
			const rotation = rotationOf({ priorityList: start.map(row) });
			applyMove(rotation, move(rotation.priorityList, from, APL_ROWS, PRIORITY, to));
			expect(names(rotation.priorityList)).toEqual(expected);
		});

		it('writes one fresh array for the changed list and leaves every other list and every row as they were', () => {
			const rows = [row('a'), row('b'), row('c')];
			const rotation = rotationOf({ priorityList: rows.slice(), prepullActions: [prepull('p')], groups: [], valueVariables: [] });
			const before = {
				priorityList: rotation.priorityList,
				prepullActions: rotation.prepullActions,
				groups: rotation.groups,
				valueVariables: rotation.valueVariables,
			};

			const result = applyMove(rotation, move(rotation.priorityList, 2, APL_ROWS, PRIORITY, 0));

			expect(result).toBe(rotation);
			expect(rotation.priorityList).not.toBe(before.priorityList);
			expect(before.priorityList.map(item => item)).toEqual(rows);
			expect(rotation.priorityList).toEqual([rows[2], rows[0], rows[1]]);
			rotation.priorityList.forEach(item => expect(rows).toContain(item));
			expect(rotation.prepullActions).toBe(before.prepullActions);
			expect(rotation.groups).toBe(before.groups);
			expect(rotation.valueVariables).toBe(before.valueVariables);
		});

		it('reorders the rotation’s groups and its variables', () => {
			const groups = [APLGroup.create({ name: 'one' }), APLGroup.create({ name: 'two' })];
			const rotation = rotationOf({ groups: groups.slice(), valueVariables: [{ name: 'x' }, { name: 'y' }, { name: 'z' }] });
			applyMove(rotation, move(rotation.groups, 0, APL_GROUPS, null, 2));
			applyMove(rotation, move(rotation.valueVariables, 2, APL_VARIABLES, null, 0));
			expect(rotation.groups).toEqual([groups[1], groups[0]]);
			expect(rotation.valueVariables.map(variable => variable.name)).toEqual(['z', 'x', 'y']);
		});
	});

	describe('between lists', () => {
		it('moves a pre-pull action at the pull into the priority list as a list item holding the same action', () => {
			const source = prepull('p', '-1s');
			const rotation = rotationOf({ prepullActions: [source], priorityList: [row('a'), row('b')] });
			const before = rotation.priorityList;

			applyMove(rotation, move(rotation.prepullActions, 0, APL_ROWS, PREPULL, 1, APL_ROWS, PRIORITY));

			expect(rotation.prepullActions).toEqual([]);
			expect(rotation.priorityList).not.toBe(before);
			expect(names(rotation.priorityList)).toEqual(['a', 'p', 'b']);
			expect(rotation.priorityList[1].action).toBe(source.action);
			expect(Object.keys(rotation.priorityList[1]).sort()).toEqual(['action', 'hide', 'notes']);
		});

		it('moves a priority row without a condition into the pre-pull list with the pull as its time', () => {
			const source = row('a');
			source.hide = true;
			const rotation = rotationOf({ prepullActions: [prepull('p')], priorityList: [source] });

			applyMove(rotation, move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 0, APL_ROWS, PREPULL));

			expect(rotation.priorityList).toEqual([]);
			expect(names(rotation.prepullActions)).toEqual(['a', 'p']);
			const moved = rotation.prepullActions[0];
			expect(moved.action).toBe(source.action);
			expect(moved.hide).toBe(true);
			expect(moved.doAtValue?.value).toEqual({ oneofKind: 'const', const: { val: '-1s' } });
			expect(moved.doAtValue?.uuid?.value).toBeTruthy();
		});

		it('moves an action from one group into another', () => {
			const [x, y, z] = [row('x'), row('y'), row('z')];
			const one = group('one', [x, y]);
			const two = group('two', [z]);
			const rotation = rotationOf({ groups: [one, two], priorityList: [row('a')] });
			const priority = rotation.priorityList;
			const groups = rotation.groups;

			applyMove(rotation, move(one.actions, 0, APL_GROUPS, aplItemId(one, 0), 1, APL_GROUPS, aplItemId(two, 1)));

			expect(one.actions).toEqual([y]);
			expect(two.actions).toEqual([z, x]);
			expect(two.actions[1]).toBe(x);
			expect(rotation.groups).toBe(groups);
			expect(rotation.priorityList).toBe(priority);
		});

		it('moves a value into another row’s value list and keeps its uuid', () => {
			const first = constant('1', 'v1');
			const outerA = and('a', [first, constant('2', 'v2')]);
			const outerB = and('b', [constant('3', 'v3')]);
			const rowA = listItem(condition(outerA));
			const rowB = listItem(condition(outerB));
			const rotation = rotationOf({ priorityList: [rowA, rowB] });
			const bVals = valsOf(outerB);

			applyMove(rotation, move(valsOf(outerA), 0, valueListGroup(outerA), 'row-a', 0, valueListGroup(outerB), 'row-b'));

			expect(valsOf(outerA).map(value => value.uuid?.value)).toEqual(['v2']);
			expect(valsOf(outerB).map(value => value.uuid?.value)).toEqual(['v1', 'v3']);
			expect(valsOf(outerB)[0]).toBe(first);
			expect(valsOf(outerB)).not.toBe(bVals);
		});

		it('reorders a value list nested inside another value', () => {
			const inner = and('inner', [constant('1', 'x'), constant('2', 'y'), constant('3', 'z')]);
			const rotation = rotationOf({ priorityList: [listItem(condition(and('outer', [inner])))] });

			applyMove(rotation, move(valsOf(inner), 0, valueListGroup(inner), 'outer', 2));

			expect(valsOf(inner).map(value => value.uuid?.value)).toEqual(['y', 'x', 'z']);
		});

		it('reorders a sequence’s actions', () => {
			const sequence = sequenceOf([reset('a'), reset('b')]);
			const rotation = rotationOf({ priorityList: [listItem(sequence)] });
			const actions = (sequence.action as any).sequence.actions;

			applyMove(rotation, move(actions, 1, actionListGroup(sequence), 'row', 0));

			expect((sequence.action as any).sequence.actions.map((action: APLAction) => (action.action as any).resetSequence.sequenceName)).toEqual(['b', 'a']);
		});
	});

	describe('refusals leave the rotation untouched', () => {
		const lists = (rotation: APLRotation) => [rotation.priorityList, rotation.prepullActions, rotation.groups, rotation.valueVariables];
		const untouched = (rotation: APLRotation, change: () => void) => {
			const snapshot = APLRotation.clone(rotation);
			const before = lists(rotation);
			change();
			expect(rotation).toEqual(snapshot);
			lists(rotation).forEach((list, index) => expect(list).toBe(before[index]));
		};

		it('refuses a group dropped into its own actions', () => {
			const one = group('one', [row('x')]);
			const rotation = rotationOf({ groups: [one] });
			const into = move(rotation.groups, 0, APL_GROUPS, null, 0, APL_GROUPS, aplItemId(one, 0));
			expect(canApplyMove(rotation, into)).toBe(false);
			untouched(rotation, () => applyMove(rotation, into));
		});

		it('refuses a value dropped into a list inside itself', () => {
			const inner = and('inner', [constant('1', 'x')]);
			const outer = and('outer', [inner, constant('2', 'y')]);
			const rotation = rotationOf({ priorityList: [listItem(condition(outer))] });
			const into = move(valsOf(outer), 0, valueListGroup(outer), 'row', 1, valueListGroup(inner), 'inner');
			expect(canApplyMove(rotation, into)).toBe(false);
			untouched(rotation, () => applyMove(rotation, into));
		});

		it('refuses a move whose id no longer names the item at its index', () => {
			const rotation = rotationOf({ priorityList: [row('a'), row('b'), row('c')] });
			const stale = move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 3);
			stale.from.id = aplItemId(rotation.priorityList[1], 1);
			expect(canApplyMove(rotation, stale)).toBe(false);
			untouched(rotation, () => applyMove(rotation, stale));
		});

		it('refuses a list it cannot find, a row between the sections and a drop back in place', () => {
			const rotation = rotationOf({ priorityList: [row('a'), row('b')], groups: [group('one', [row('x')])] });
			const moves = [
				move(rotation.priorityList, 0, 'somewhere-else', PRIORITY, 2),
				move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 0, APL_ROWS, null),
				move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 1),
				move(rotation.priorityList, 0, valueListGroup(and('missing', [])), PRIORITY, 1),
				move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 0, APL_GROUPS, aplItemId(rotation.groups[0], 0)),
				move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 5),
			];
			for (const refused of moves) {
				expect(canApplyMove(rotation, refused)).toBe(false);
				untouched(rotation, () => applyMove(rotation, refused));
			}
		});

		it('refuses a pre-pull move for an action the other list does not offer', () => {
			const rotation = rotationOf({ priorityList: [row('a')], prepullActions: [] });
			const toPrepull = move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 0, APL_ROWS, PREPULL);
			expect(canApplyMove(rotation, toPrepull, (kind, isPrepull) => !(kind === 'resetSequence' && isPrepull))).toBe(false);
			expect(canApplyMove(rotation, toPrepull, () => true)).toBe(true);
		});

		it('answers the same with a source resolved once per drag, and re-resolves it once its list is replaced', () => {
			const rotation = rotationOf({ priorityList: [row('a'), row('b'), row('c')], prepullActions: [prepull('p', '-1s')] });
			const from = move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 0).from;
			const source = resolveMoveSource(rotation, from);
			const targets = [0, 1, 2, 3].flatMap(index => [
				{ group: APL_ROWS, parentId: PRIORITY, index },
				{ group: APL_ROWS, parentId: PREPULL, index: Math.min(index, 1) },
			]);
			for (const to of targets) expect(canApplyMove(rotation, { from, to }, undefined, source)).toBe(canApplyMove(rotation, { from, to }));

			applyMove(rotation, move(rotation.priorityList, 2, APL_ROWS, PRIORITY, 0));
			const stale = { from: move(rotation.priorityList, 1, APL_ROWS, PRIORITY, 0).from, to: { group: APL_ROWS, parentId: PRIORITY, index: 3 } };
			expect(canApplyMove(rotation, stale, undefined, source)).toBe(canApplyMove(rotation, stale));
			expect(canApplyMove(rotation, stale, undefined, source)).toBe(true);
		});

		it('refuses a priority row with a condition into the pre-pull list, which has no room for it', () => {
			const source = row('a');
			source.action!.condition = constant('false', 'never');
			const rotation = rotationOf({ priorityList: [source], prepullActions: [prepull('p')] });
			const toPrepull = move(rotation.priorityList, 0, APL_ROWS, PRIORITY, 0, APL_ROWS, PREPULL);
			expect(canApplyMove(rotation, toPrepull)).toBe(false);
			untouched(rotation, () => applyMove(rotation, toPrepull));
			expect(source.action!.condition?.uuid?.value).toBe('never');
		});

		it('refuses a pre-pull row timed away from the pull into the priority list, which has no room for its time', () => {
			const rotation = rotationOf({ prepullActions: [prepull('p', '-3s')], priorityList: [row('a')] });
			const toPriority = move(rotation.prepullActions, 0, APL_ROWS, PREPULL, 0, APL_ROWS, PRIORITY);
			expect(canApplyMove(rotation, toPriority)).toBe(false);
			untouched(rotation, () => applyMove(rotation, toPriority));
		});

		it('refuses an operand list whose owner’s uuid another value shares', () => {
			const first = and('same', [constant('1', 'x'), constant('2', 'y')]);
			const second = and('same', [constant('3', 'z')]);
			const rotation = rotationOf({ priorityList: [listItem(condition(first)), listItem(condition(second))] });
			const reorder = move(valsOf(first), 0, valueListGroup(first), 'row', 2);
			expect(canApplyMove(rotation, reorder)).toBe(false);
			untouched(rotation, () => applyMove(rotation, reorder));
		});
	});
});
