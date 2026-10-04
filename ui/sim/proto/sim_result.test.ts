import {
	ActionMetrics as ActionMetricsProto,
	TargetedActionMetrics as TargetedActionMetricsProto,
	UnitMetrics as UnitMetricsProto,
} from '@generated/proto/api';
import { makeSimRun } from '@testing/storybook/fixtures/results';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { SimResult } from './sim_result';

// makeSimRun()'s warrior and its target.
const PLAYER = 0;
const TARGET = 1;

type Entry = Partial<TargetedActionMetricsProto>;

// The warrior's main-hand action with these entries on the target and on the warrior. The sim lists
// every unit on every action, the action's own unit included.
const playerAction = async (onTarget: Entry, onPlayer: Entry) => {
	const { request, result } = makeSimRun();
	result.raidMetrics!.parties[0].players[0].actions[0].targets = [
		TargetedActionMetricsProto.create({ unitIndex: TARGET, hits: 1, ...onTarget }),
		TargetedActionMetricsProto.create({ unitIndex: PLAYER, hits: 1, ...onPlayer }),
	];
	return (await SimResult.makeNew(request, result)).getFirstPlayer()!;
};

describe('damage done and damage taken from self-inflicted damage', () => {
	it('counts only the damage on the other side as damage done', async () => {
		const [sapper] = (await playerAction({ damage: 1000 }, { damage: 500 })).getDamageActions();

		expect(sapper.damage).toBe(1000);
		expect(sapper.forTarget({ target: TARGET }).damage).toBe(1000);
	});

	it('leaves an action that puts nothing on its own side as it is', async () => {
		const player = await playerAction({ damage: 1000 }, { casts: 3 });

		expect(player.getDamageActions()).toEqual(player.actions);
	});

	it('drops an action that only damages its own side from damage done and threat', async () => {
		const player = await playerAction({}, { damage: 500, threat: 500 });

		expect(player.getDamageActions()).toEqual([]);
		expect(player.getThreatActions()).toEqual([]);
	});

	it('keeps an action that damages the other side and heals its own out of damage done', async () => {
		const player = await playerAction({ damage: 1000 }, { healing: 300 });

		expect(player.getDamageActions()).toEqual([]);
	});

	it("reports the damage its own side deals a unit, its own hits and its pet's, as that unit's damage taken", async () => {
		const { request, result } = makeSimRun();
		const warrior = result.raidMetrics!.parties[0].players[0];
		warrior.actions[0].targets = [
			TargetedActionMetricsProto.create({ unitIndex: TARGET, hits: 1, damage: 1000 }),
			TargetedActionMetricsProto.create({ unitIndex: PLAYER, hits: 1, damage: 500 }),
		];
		warrior.pets = [
			UnitMetricsProto.create({
				name: 'Pet',
				unitIndex: 2,
				actions: [
					ActionMetricsProto.create({
						id: warrior.actions[0].id,
						targets: [
							TargetedActionMetricsProto.create({ unitIndex: TARGET }),
							TargetedActionMetricsProto.create({ unitIndex: PLAYER, hits: 1, damage: 300 }),
						],
					}),
				],
			}),
		];
		const simResult = await SimResult.makeNew(request, result);

		expect(simResult.getOwnSideDamageOn(simResult.getFirstPlayer()!).map(hit => hit.damage)).toEqual([500, 300]);
	});

	it('leaves a hit on its own side out of the DPS and threat timelines', async () => {
		const { request, result } = makeSimRun();
		result.logs = [
			'[1.00] [Player (#1)] [Target 1] {OtherID: 1} Hit for 1000.000 damage (SpellSchool: 1). (Threat: 800.000)',
			'[2.00] [Player (#1)] [Player (#1)] {OtherID: 1} Hit for 500.000 damage (SpellSchool: 1). (Threat: 500.000)',
		].join('\n');
		const player = (await SimResult.makeNew(request, result)).getFirstPlayer()!;

		expect(player.damageDealtLogs.map(log => log.amount)).toEqual([1000]);
		expect(player.maxThreat).toBe(800);
	});

	it("keeps an enemy's damage on the player as damage done", async () => {
		const { request, result } = makeSimRun();
		result.encounterMetrics!.targets[0].actions = [
			ActionMetricsProto.create({
				id: result.raidMetrics!.parties[0].players[0].actions[0].id,
				targets: [
					TargetedActionMetricsProto.create({ unitIndex: PLAYER, hits: 1, damage: 800 }),
					TargetedActionMetricsProto.create({ unitIndex: TARGET }),
				],
			}),
		];
		const [target] = (await SimResult.makeNew(request, result)).getTargets();

		expect(target.getDamageActions().map(melee => melee.damage)).toEqual([800]);
	});
});

// This proto calls a blocked critical hit `blockedCrit*` where MoP's names the same three fields
// `critBlock*`. Both names exist in the two codebases and neither is a type error against the
// other's proto, so pin this one's.
describe('TargetedActionMetrics blocked-crit and crush getters', () => {
	const metrics = async (over: Partial<TargetedActionMetricsProto>) => {
		const { request, result } = makeSimRun();
		request.simOptions!.iterations = 2;
		result.raidMetrics!.parties[0].players[0].actions[0].targets = [TargetedActionMetricsProto.create({ unitIndex: TARGET, ...over })];
		const [action] = (await SimResult.makeNew(request, result)).getFirstPlayer()!.actions;
		return action.targets[0];
	};

	it('averages blockedCrits and blockedCritDamage over the iterations', async () => {
		const m = await metrics({ blockedCrits: 10, blockedCritDamage: 5000 });
		expect(m.blockedCrits).toBe(5);
		expect(m.blockedCritDamage).toBe(5000);
		expect(m.avgBlockedCritDamage).toBe(2500);
	});

	it('takes blockedCritPercent against every hit attempt', async () => {
		// 10 blocked crits out of 40 attempts (10 hits + 10 crits + 10 misses + 10 blocked crits).
		const m = await metrics({ hits: 10, crits: 10, misses: 10, blockedCrits: 10 });
		expect(m.hitAttempts).toBe(40);
		expect(m.blockedCritPercent).toBe(25);
	});

	it('counts blocked crits and crushes as landed hits', async () => {
		const m = await metrics({ hits: 2, blockedCrits: 4, crushes: 6 });
		expect(m.landedHitsRaw).toBe(12);
	});

	it('exposes the crush family', async () => {
		const m = await metrics({ hits: 10, crushes: 10, crushDamage: 4000 });
		expect(m.crushes).toBe(5);
		expect(m.crushDamage).toBe(4000);
		expect(m.avgCrushDamage).toBe(2000);
		expect(m.crushPercent).toBe(50);
	});
});
