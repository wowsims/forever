import { OtherAction } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { ActionId } from '@sim/proto/action_id';
import type { CastLog, ResourceGroupLog } from '@sim/proto/combat_log';
import type { UnitMetrics } from '@sim/proto/sim_result';
import { describe, expect, it } from 'vitest';

import { buildRotationModel } from './build';
import type { CastItem, CastRow, ResourceItem, ResourceRow } from './types';

const SEAL = ActionId.fromSpellId(20293);

const cast = (timestamp: number): CastLog =>
	({
		actionId: SEAL,
		timestamp,
		castTime: 0,
		gcd: 0,
		effectiveTime: 0,
		travelTime: 0,
		cancelTime: 0,
		delay: 0,
		castCancelledLog: null,
		damageDealtLogs: [],
	}) as unknown as CastLog;

const unit = (castLogs: Array<CastLog>) =>
	({
		castLogs,
		auraUptimeLogs: [],
		groupedResourceLogs: new Proxy({}, { get: () => [] }),
		pets: [],
		getMeleeActions: () => [],
		getSpellActions: () => [],
	}) as unknown as UnitMetrics;

const stacks = (timestamps: Array<number>) => {
	const model = buildRotationModel({ player: unit(timestamps.map(cast)), targets: [unit([])], duration: 30, showGcd: false });
	const row = model.rows.find(r => r.kind === 'cast') as CastRow;
	return (row.items as Array<CastItem>).map(item => item.stack.map(log => log.timestamp));
};

const castOf = (actionId: ActionId, timestamp: number, travelTime = 0): CastLog => ({ ...cast(timestamp), actionId, travelTime }) as CastLog;

const castRows = (castLogs: Array<CastLog>) =>
	buildRotationModel({ player: unit(castLogs), targets: [unit([])], duration: 30, showGcd: false }).rows.filter((r): r is CastRow => r.kind === 'cast');

describe('wand shots', () => {
	const channel = ActionId.fromSpellId(5019);
	const shot = ActionId.fromSpellId(5019, 0, 1);

	it('draws the shots in the Shoot row, with no row of their own', () => {
		const rows = castRows([castOf(channel, 1), castOf(shot, 1.5, 0.2), castOf(shot, 2.5, 0.2)]);
		expect(rows).toHaveLength(1);
		const items = rows[0].items.filter((item): item is CastItem => item.kind === 'cast');
		expect(items.map(item => item.log.actionId!.tag)).toEqual([0, 1, 1]);
		expect(items.slice(1).map(item => item.travelDuration)).toEqual([0.2, 0.2]);
	});

	it('keeps the main-hand and off-hand autos in rows of their own', () => {
		const attack = (tag: number) => ActionId.fromOtherId(OtherAction.OtherActionAttack, tag);
		expect(castRows([castOf(attack(1), 1), castOf(attack(2), 1.5)])).toHaveLength(2);
	});

	it('keeps the tags of a spell outside the grouped list in rows of their own', () => {
		expect(castRows([castOf(ActionId.fromSpellId(20293), 1), castOf(ActionId.fromSpellId(20293, 0, 1), 2)])).toHaveLength(2);
	});
});

describe('resource rows', () => {
	const group = (resourceType: ResourceType, timestamp: number, valueBefore: number, valueAfter: number): ResourceGroupLog =>
		({ resourceType, timestamp, valueBefore, valueAfter, maxValue: 1000, logs: [] }) as unknown as ResourceGroupLog;

	const resourceItems = (resourceType: ResourceType, groups: Array<ResourceGroupLog>) => {
		const player = {
			...unit([]),
			groupedResourceLogs: new Proxy({}, { get: (_, key) => (Number(key) === resourceType ? groups : []) }),
		} as unknown as UnitMetrics;
		const rows = buildRotationModel({ player, targets: [unit([])], duration: 30, showGcd: false }).rows.filter(
			(r): r is ResourceRow => r.kind === 'resource',
		);
		expect(rows).toHaveLength(1);
		return rows[0].items as Array<ResourceItem>;
	};

	const regen = (resourceType: ResourceType) => [group(resourceType, 0, 1000, 1000), group(resourceType, 0.1, 400, 450), group(resourceType, 0.2, 450, 500)];

	it('draws mana as a level bar per step, without a number or a percent in the block', () => {
		const items = resourceItems(ResourceType.ResourceTypeMana, regen(ResourceType.ResourceTypeMana));
		expect(items).toHaveLength(3);
		expect(items.map(item => item.display)).toEqual(['fill', 'fill', 'fill']);
		expect(items.map(item => item.fillPercent)).toEqual([100, 45, 50]);
		expect(items.map(item => item.text)).toEqual(['', '', '']);
	});

	it('draws energy and focus the same way as mana', () => {
		for (const resourceType of [ResourceType.ResourceTypeEnergy, ResourceType.ResourceTypeFocus]) {
			expect(resourceItems(resourceType, regen(resourceType)).map(item => item.display)).toEqual(['fill', 'fill', 'fill']);
		}
	});

	it('draws a credit between two regen ticks in the slice of the tick before it', () => {
		const mana = ResourceType.ResourceTypeMana;
		const items = resourceItems(mana, [group(mana, 7.41, 500, 520), group(mana, 7.5, 520, 530), group(mana, 7.51, 530, 540), group(mana, 7.61, 540, 560)]);
		expect(items).toHaveLength(3);
		expect(items.map(item => [item.start, item.log.valueBefore, item.log.valueAfter])).toEqual([
			[7.41, 500, 530],
			[7.51, 530, 540],
			[7.61, 540, 560],
		]);
		expect(items[0].end).toBe(items[1].start);
		expect(items[0].fillPercent).toBe(53);
	});

	it('keeps a spend between two regen ticks in the level and the events of its step', () => {
		const mana = ResourceType.ResourceTypeMana;
		const spend = { ...group(mana, 1.55, 1000, 700), logs: [{ isSpend: true }] } as unknown as ResourceGroupLog;
		const items = resourceItems(mana, [
			{ ...group(mana, 1.51, 1000, 1000), logs: [{ isSpend: false }] } as unknown as ResourceGroupLog,
			spend,
			group(mana, 1.61, 700, 710),
		]);
		expect(items).toHaveLength(2);
		expect(items[0].fillPercent).toBe(70);
		expect(items[0].log.logs).toHaveLength(2);
	});

	it('ends a slice before the pull where the next group starts, even at 0', () => {
		const mana = ResourceType.ResourceTypeMana;
		const items = resourceItems(mana, [group(mana, -0.1, 1000, 900), group(mana, 0, 900, 900), group(mana, 0.1, 900, 950)]);
		expect(items.map(item => [item.start, item.end])).toEqual([
			[-0.1, 0],
			[0, 0.1],
			[0.1, 30],
		]);
	});

	it('keeps every group of a number row, however close', () => {
		const rage = ResourceType.ResourceTypeRage;
		expect(resourceItems(rage, [group(rage, 1, 0, 10), group(rage, 1.02, 10, 20)])).toHaveLength(2);
	});

	it('keeps health as a percent and rage as a number', () => {
		expect(resourceItems(ResourceType.ResourceTypeHealth, regen(ResourceType.ResourceTypeHealth))[1]).toMatchObject({ display: 'percent', text: '45%' });
		expect(resourceItems(ResourceType.ResourceTypeRage, regen(ResourceType.ResourceTypeRage))[1]).toMatchObject({ display: 'number', text: '450' });
	});
});

describe('cast stacks', () => {
	it('runs casts closer than the window into one stack, each carrying the casts before it', () => {
		expect(stacks([1, 1, 1.15, 5])).toEqual([[1], [1, 1], [1, 1, 1.15], [5]]);
	});

	it('starts a new stack after a gap of the window or more', () => {
		expect(stacks([1, 1.22])).toEqual([[1], [1.22]]);
	});
});
