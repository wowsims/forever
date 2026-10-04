import { ActionId } from '@sim/proto/action_id';
import type { CastLog } from '@sim/proto/combat_log';
import type { UnitMetrics } from '@sim/proto/sim_result';
import { describe, expect, it } from 'vitest';

import { buildRotationModel } from './build';
import type { CastItem, CastRow } from './types';

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

describe('cast stacks', () => {
	it('runs casts closer than the window into one stack, each carrying the casts before it', () => {
		expect(stacks([1, 1, 1.15, 5])).toEqual([[1], [1, 1], [1, 1, 1.15], [5]]);
	});

	it('starts a new stack after a gap of the window or more', () => {
		expect(stacks([1, 1.22])).toEqual([[1], [1.22]]);
	});
});
