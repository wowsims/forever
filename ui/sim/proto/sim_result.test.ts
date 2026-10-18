import {
	ActionMetrics as ActionMetricsProto,
	HitKind,
	TargetedActionMetrics as TargetedActionMetricsProto,
	UnitMetrics as UnitMetricsProto,
} from '@generated/proto/api';
import { makeSimRun, valueRange as range } from '@testing/storybook/fixtures/results';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { ActionMetrics, SimResult } from './sim_result';

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

	it('takes the casts of a DoT as its hit attempts when every tick crits', async () => {
		const normalTicks = await metrics({ casts: 10, misses: 1, ticks: 50 });
		const critTicks = await metrics({ casts: 10, misses: 1, critTicks: 50 });
		expect(critTicks.hitAttempts).toBe(normalTicks.hitAttempts);
		expect(critTicks.missPercent).toBe(normalTicks.missPercent);
	});

	it('takes the casts of a DoT, its misses among them, as its hit attempts', async () => {
		const m = await metrics({ casts: 10, misses: 1, ticks: 50 });
		expect(m.hitAttempts).toBe(10);
		expect(m.missPercent).toBe(10);
	});

	it('takes the direct outcomes of a spell that also ticks as its hit attempts when every direct hit crits', async () => {
		const m = await metrics({ casts: 10, misses: 1, crits: 9, ticks: 40 });
		expect(m.hitAttempts).toBe(10);
		expect(m.missPercent).toBe(10);
	});
});

// One main-hand action per entry list, each listing its entries as targets, over two iterations.
const rangedActions = async (...perAction: Array<Array<Entry>>) => {
	const { request, result } = makeSimRun();
	request.simOptions!.iterations = 2;
	const player = result.raidMetrics!.parties[0].players[0];
	const template = player.actions[0];
	player.actions = perAction.map(entries =>
		ActionMetricsProto.create({
			...template,
			targets: entries.map(entry => TargetedActionMetricsProto.create({ unitIndex: TARGET, ...entry })),
		}),
	);
	return (await SimResult.makeNew(request, result)).getFirstPlayer()!.actions;
};

describe('per-value ranges', () => {
	it('merges the targets of an action by count, weighted mean and the widest bounds', async () => {
		const [action] = await rangedActions([
			{ ranges: { [HitKind.HitKindHit]: range(2, 10, 2, 8, 12) } },
			{ ranges: { [HitKind.HitKindHit]: range(6, 20, 0, 15, 30) } },
		]);

		const hit = action.range(HitKind.HitKindHit)!;
		expect(hit.aggregatorData!.n).toBe(8);
		expect(hit.avg).toBeCloseTo(17.5);
		expect(hit.min).toBe(8);
		expect(hit.max).toBe(30);
		expect(hit.stdev).toBeCloseTo(Math.sqrt(19.75));
		expect(hit.aggregatorData!.sumSq).toBe(0);
	});

	it('ignores a side that landed nothing when taking the bounds', async () => {
		const [action] = await rangedActions([
			{ ranges: { [HitKind.HitKindCrit]: range(0, 0, 0, 0, 0) } },
			{ ranges: { [HitKind.HitKindCrit]: range(3, 50, 0, 40, 60) } },
		]);

		const crit = action.range(HitKind.HitKindCrit)!;
		expect(crit.min).toBe(40);
		expect(crit.max).toBe(60);
		expect(action.range(HitKind.HitKindTick)).toBeUndefined();
	});

	it('merges the same kind across actions and keeps the kinds one side lacks', async () => {
		const actions = await rangedActions(
			[{ ranges: { [HitKind.HitKindHit]: range(2, 10, 0, 9, 11), [HitKind.HitKindTick]: range(4, 5, 0, 4, 6) } }],
			[{ ranges: { [HitKind.HitKindHit]: range(2, 30, 0, 25, 35) } }],
		);

		const merged = ActionMetrics.merge(actions);
		const hit = merged.range(HitKind.HitKindHit)!;
		expect(hit.aggregatorData!.n).toBe(4);
		expect(hit.avg).toBeCloseTo(20);
		expect(hit.min).toBe(9);
		expect(hit.max).toBe(35);
		expect(merged.range(HitKind.HitKindTick)!.aggregatorData!.n).toBe(4);
	});

	it('merges custom metrics by key and reports their amount per iteration', async () => {
		const actions = await rangedActions(
			[{ customMetrics: [{ key: 'spec_metric', range: range(4, 100, 0, 80, 120) }] }],
			[
				{
					customMetrics: [
						{ key: 'spec_metric', range: range(4, 200, 0, 150, 260) },
						{ key: 'divine_aegis', range: range(2, 50, 0, 50, 50) },
					],
				},
			],
		);

		const byKey = Object.fromEntries(ActionMetrics.merge(actions).customMetrics.map(metric => [metric.key, metric]));
		expect(byKey.spec_metric).toMatchObject({ value: 600, average: 150, min: 80, max: 260 });
		expect(byKey.divine_aegis).toMatchObject({ value: 50, average: 50 });
	});

	it('gives a custom metric its share of the damage of the action, or of its healing when it deals none', async () => {
		const [damaging, healing] = await rangedActions(
			[{ damage: 1000, customMetrics: [{ key: 'spec_metric', range: range(4, 100, 0, 80, 120) }] }],
			[{ healing: 1000, customMetrics: [{ key: 'divine_aegis', range: range(2, 50, 0, 50, 50) }] }],
		);

		expect(damaging.customMetrics[0].percentage).toBeCloseTo(40);
		expect(healing.customMetrics[0].percentage).toBeCloseTo(10);
	});

	it('gives each damage kind its bounds only when the kind landed values, and drops an empty custom metric', async () => {
		const [action] = await rangedActions([
			{
				hits: 4,
				ranges: { [HitKind.HitKindHit]: range(4, 10, 0, 8, 12), [HitKind.HitKindCrit]: range(0, 0, 0, 0, 0) },
				customMetrics: [{ key: 'divine_aegis', range: range(0, 0, 0, 0, 0) }],
			},
		]);

		expect(action.damageDone.hit).toMatchObject({ min: 8, max: 12 });
		expect(action.damageDone.critHit).not.toHaveProperty('min');
		expect(action.damageDone.critHit).not.toHaveProperty('max');
		expect(action.customMetrics).toEqual([]);
	});

	it('averages each healing kind and shares the total over healing and shielding', async () => {
		const [action] = await rangedActions([
			{
				healing: 1000,
				shielding: 100,
				ranges: {
					[HitKind.HitKindHeal]: range(4, 100, 0, 90, 110),
					[HitKind.HitKindCritHeal]: range(2, 300, 0, 280, 320),
					[HitKind.HitKindShield]: range(1, 100, 0, 100, 100),
				},
			},
		]);

		const done = action.healingDone;
		expect(done.heal).toMatchObject({ value: 200, average: 100, min: 90, max: 110 });
		expect(done.critHeal).toMatchObject({ value: 300, average: 300, min: 280, max: 320 });
		expect(done.shield).toMatchObject({ value: 50, average: 100 });
		expect(done.heal.percentage).toBeCloseTo((200 / 550) * 100);
		expect(done.tick).toEqual({ value: 0, percentage: 0, average: 0 });
	});

	it('shows a shield range as the shield row of an action that only shields', async () => {
		const [action] = await rangedActions([{ hits: 2, shielding: 1000, ranges: { [HitKind.HitKindShield]: range(4, 250, 0, 200, 300) } }]);

		const done = action.healingDone;
		expect(done.shield).toMatchObject({ value: 500, average: 250, min: 200, max: 300 });
		expect(done.shield.percentage).toBeCloseTo(100);
		expect(done.heal.value).toBe(0);
		expect(done.critHeal.value).toBe(0);
	});
});
