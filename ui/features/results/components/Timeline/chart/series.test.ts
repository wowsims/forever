import { ResourceType } from '@generated/proto/spell';
import type { ResourceGroupLog } from '@sim/proto/combat_log';
import type { UnitMetrics } from '@sim/proto/sim_result';
import { describe, expect, it } from 'vitest';

import { manaDataset } from './series';

const group = (timestamp: number, valueBefore: number, valueAfter: number, maxValue: number) =>
	({ timestamp, valueBefore, valueAfter, maxValue, logs: [] }) as unknown as ResourceGroupLog;

const unit = (manaLogs: Array<ResourceGroupLog>) => ({ groupedResourceLogs: { [ResourceType.ResourceTypeMana]: manaLogs } }) as unknown as UnitMetrics;

describe('manaDataset', () => {
	it('scales to the maximum mana when mana was spent before the pull', () => {
		const mana = manaDataset(unit([group(-1, 3565, 2865, 3565), group(0, 2865, 2800, 3565), group(5, 2800, 2600, 3565)]));
		expect(mana?.maxMana).toBe(3565);
	});

	it('scales to the highest maximum mana when the maximum rises during the fight', () => {
		const mana = manaDataset(unit([group(0, 3565, 3400, 3565), group(5, 3400, 3500, 3800)]));
		expect(mana?.maxMana).toBe(3800);
	});
});
