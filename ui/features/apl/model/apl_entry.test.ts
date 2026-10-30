import { RaidSimRequest } from '@generated/proto/api';
import { APLAction, APLListItem, APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { parseAplEntryPath } from '@sim/proto/combat_log';
import type { SimResult, UnitMetrics } from '@sim/proto/sim_result';
import { describe, expect, it } from 'vitest';

import { resolveAplEntry, runRotation } from './apl_entry';

const item = (notes: string, action?: APLAction) => APLListItem.create({ notes, action });
const sequence = (count: number, nested?: APLAction): APLAction =>
	APLAction.create({
		action: {
			oneofKind: 'sequence',
			sequence: { name: '', actions: Array.from({ length: count }, (_, i) => (i === 1 && nested ? nested : APLAction.create())) },
		},
	});
const strictSequence = (count: number): APLAction =>
	APLAction.create({
		action: { oneofKind: 'strictSequence', strictSequence: { actions: Array.from({ length: count }, () => APLAction.create()) } },
	});

const rotation = APLRotation.create({
	prepullActions: [{ hide: false }],
	priorityList: [item('p0', sequence(3)), item('p1'), item('p2', sequence(2)), item('p3', sequence(3, strictSequence(2)))],
	groups: [
		{ name: 'first', actions: [item('g0a0')], variables: [] },
		{ name: 'second', actions: [item('g1a0'), item('g1a1'), item('g1a2', sequence(4))], variables: [] },
	],
});

const resolve = (text: string) => resolveAplEntry(rotation, parseAplEntryPath(text)!);

describe('resolveAplEntry', () => {
	it('names a priority row', () => {
		const resolved = resolve('priority 2')!;
		expect(resolved.kind).toBe('priority');
		expect(resolved.item).toBe(rotation.priorityList[1]);
		expect(resolved.stepIndexes).toEqual([]);
		expect(resolved.stepActions).toEqual([]);
	});

	it('names a group row and carries the group name in the crumbs', () => {
		const resolved = resolve('priority 3 > group 2.2')!;
		expect(resolved.kind).toBe('group');
		expect(resolved.item).toBe(rotation.groups[1].actions[1]);
		expect(resolved.crumbs).toEqual([
			{ kind: 'priority', index: 2 },
			{ kind: 'group', group: 1, index: 1, groupName: 'second' },
		]);
	});

	it('keeps trailing steps apart from the row of the sequence', () => {
		const resolved = resolve('priority 1 > step 3')!;
		expect(resolved.item).toBe(rotation.priorityList[0]);
		expect(resolved.stepIndexes).toEqual([2]);
	});

	it('names the group action holding the sequence for a group row with a step', () => {
		const resolved = resolve('priority 3 > group 2.3 > step 3')!;
		expect(resolved.kind).toBe('group');
		expect(resolved.item).toBe(rotation.groups[1].actions[2]);
		expect(resolved.stepIndexes).toEqual([2]);
	});

	it('takes the row from a group reference that is a step of a sequence', () => {
		const resolved = resolve('priority 3 > step 2 > group 2.2')!;
		expect(resolved.kind).toBe('group');
		expect(resolved.item).toBe(rotation.groups[1].actions[1]);
		expect(resolved.stepIndexes).toEqual([]);
		expect(resolved.stepActions).toEqual([]);
	});

	it('walks nested sequences, outermost step first, naming the rotation’s own step actions', () => {
		const resolved = resolve('priority 4 > step 2 > step 2')!;
		const outer = rotation.priorityList[3].action!.action as { sequence: { actions: Array<APLAction> } };
		const inner = outer.sequence.actions[1].action as { strictSequence: { actions: Array<APLAction> } };
		expect(resolved.item).toBe(rotation.priorityList[3]);
		expect(resolved.stepIndexes).toEqual([1, 1]);
		expect(resolved.stepActions).toHaveLength(2);
		expect(resolved.stepActions[0]).toBe(outer.sequence.actions[1]);
		expect(resolved.stepActions[1]).toBe(inner.strictSequence.actions[1]);
	});

	it('walks into a sequence a Schedule wraps', () => {
		const scheduled = APLRotation.create({
			priorityList: [item('s0', APLAction.create({ action: { oneofKind: 'schedule', schedule: { schedule: '0s', innerAction: sequence(3) } } }))],
		});
		const resolved = resolveAplEntry(scheduled, parseAplEntryPath('priority 1 > step 2')!)!;
		expect(resolved.item).toBe(scheduled.priorityList[0]);
		expect(resolved.stepIndexes).toEqual([1]);
	});

	it('names a prepull row', () => {
		const resolved = resolve('prepull 1')!;
		expect(resolved.kind).toBe('prepull');
		expect(resolved.item).toBe(rotation.prepullActions[0]);
	});

	it.each([
		'priority 5',
		'prepull 2',
		'group 3.1',
		'group 1.2',
		'step 1',
		'priority 99 > group 1.1',
		'priority 1 > group 3.1',
		'priority 1 > step 4',
		'priority 2 > step 1',
		'priority 3 > step 3',
		'priority 4 > step 2 > step 3',
		'priority 3 > step 3 > group 1.1',
	])('gives null for %s, out of range', text => {
		expect(resolve(text)).toBeNull();
	});
});

describe('runRotation', () => {
	const ran = (type: APLRotation_Type, notes: string) => APLRotation.create({ type, priorityList: [{ notes }] });
	const result = (...players: Array<APLRotation>) =>
		({
			request: RaidSimRequest.create({
				raid: { parties: [{ players: [] }, { players: players.map(rotation => ({ rotation })) }] },
			}),
		}) as unknown as SimResult;
	const unit = (index: number) => ({ index }) as UnitMetrics;

	it('takes the rotation of the request player at the unit’s raid index', () => {
		const run = result(ran(APLRotation_Type.TypeAPL, 'first'), ran(APLRotation_Type.TypeAPL, 'second'));
		expect(runRotation(run, unit(6))?.priorityList[0].notes).toBe('second');
	});

	it.each([APLRotation_Type.TypeAuto, APLRotation_Type.TypeSimple, APLRotation_Type.TypeUnknown])('takes the rotation of a run of type %s too', type => {
		expect(runRotation(result(ran(type, 'ran')), unit(5))?.priorityList[0].notes).toBe('ran');
	});

	it('gives nothing for a unit the request holds no player for', () => {
		expect(runRotation(result(ran(APLRotation_Type.TypeAPL, 'only')), unit(7))).toBeUndefined();
	});
});
