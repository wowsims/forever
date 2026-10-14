import { describe, expect, it } from 'vitest';

import { makeSimResult } from './results';

describe('makeSimResult', () => {
	it('resolves one warrior, one target and the main-hand attack without the item database', async () => {
		const result = await makeSimResult();

		const player = result.getFirstPlayer()!;
		expect(player.name).toBe('Player');
		expect(player.dps.avg).toBeCloseTo(325.3, 1);
		expect(player.getDamageActions().map(action => action.name)).toEqual(['common.action_names.attack (common.action_tags.main_hand)']);
		expect(result.getTargets().map(target => target.name)).toEqual(['Raid Target']);
		expect(result.iterations).toBe(100);
		expect(result.duration).toBe(180);
	});

	it('scales the damage of a second run and never hands back the first from the cache', async () => {
		const first = await makeSimResult();
		const second = await makeSimResult({ dpsScale: 1.1 });

		expect(second).not.toBe(first);
		expect(second.getFirstPlayer()!.dps.avg).toBeCloseTo(first.getFirstPlayer()!.dps.avg * 1.1, 6);
	});
});
