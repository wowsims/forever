import { AddWave, DamageAmpType, EncounterPlan, HealthKeyframe, PlanBossAbility_Kind, Target } from '@generated/proto/common';
import { addEventAt } from '@testing/encounter_plan';
import { describe, expect, it } from 'vitest';

import { addEvent, removeEvent, withAbilityKind } from './lanes';
import { PLAN_ISSUE_KEYS, validatePlan } from './validate';

const keys = (plan: EncounterPlan, targetCount = 1) =>
	validatePlan(plan, targetCount).map(issue => `${issue.ref.kind}${issue.ref.index}:${issue.severity}:${issue.key}`);

const downtime = (plan: EncounterPlan, start: number, end: number, targetIndex = 0) => {
	addEvent(plan, 'downtime');
	const added = plan.downtimes[plan.downtimes.length - 1];
	added.start!.at = start;
	added.end!.at = end;
	added.targetIndex = targetIndex;
};

describe('validatePlan', () => {
	it('accepts an empty plan and the defaults', () => {
		const plan = EncounterPlan.create();
		expect(validatePlan(plan, 1)).toEqual([]);
		for (const kind of ['bossHealth', 'ability', 'wave', 'move', 'amp'] as const) addEvent(plan, kind);
		expect(validatePlan(plan, 1)).toEqual([]);
	});

	it('flags an end at or before the start as an error', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.5, 0.5);
		expect(keys(plan)).toContain(`downtime0:error:${PLAN_ISSUE_KEYS.endBeforeStart}`);
	});

	it('warns when the window or its jitter leaves the fight', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.2, 0.4);
		plan.downtimes[0].start!.jitter = 0.25;
		expect(keys(plan)).toContain(`downtime0:warning:${PLAN_ISSUE_KEYS.outsideFight}`);
		addEventAt(plan, 'wave', 0.9);
		plan.addWaves[0].spawn!.jitter = 0.15;
		expect(keys(plan)).toContain(`wave0:warning:${PLAN_ISSUE_KEYS.outsideFight}`);
	});

	it('lets a wave live past the fight end, which only its spawn must stay inside', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'wave');
		Object.assign(plan.addWaves[0], { spawn: { at: 0.95, jitter: 0 }, lifetime: 0.2 });
		expect(keys(plan)).toEqual([]);
	});

	it('warns about overlapping downtimes on one target but not on different targets', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.2, 0.4);
		downtime(plan, 0.3, 0.5);
		expect(keys(plan, 2)).toEqual([`downtime1:warning:${PLAN_ISSUE_KEYS.downtimeOverlap}`]);
		plan.downtimes[1].targetIndex = 1;
		expect(keys(plan, 2).filter(k => k.endsWith(PLAN_ISSUE_KEYS.downtimeOverlap))).toEqual([]);
	});

	it('flags what the sim skips: an amp of 0% or less, a wave of more than 64 adds, a move of no yards and no seconds', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'amp');
		plan.damageAmps[0].percent = 0;
		addEvent(plan, 'wave');
		plan.addWaves[0].adds = Array.from({ length: 65 }, () => Target.create());
		addEvent(plan, 'move');
		plan.moves[0].yards = 0;
		expect(keys(plan)).toEqual([
			`amp0:error:${PLAN_ISSUE_KEYS.ampNotPositive}`,
			`wave0:error:${PLAN_ISSUE_KEYS.tooManyAdds}`,
			`move0:error:${PLAN_ISSUE_KEYS.moveGoesNowhere}`,
		]);
		plan.damageAmps[0].percent = -10;
		plan.addWaves[0].adds = Array.from({ length: 64 }, () => Target.create());
		plan.moves[0].seconds = 2;
		expect(keys(plan)).toEqual([`amp0:error:${PLAN_ISSUE_KEYS.ampNotPositive}`]);
	});

	it('flags a wave that spawns no adds', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'wave');
		plan.addWaves[0].adds = [];
		expect(keys(plan)).toEqual([`wave0:error:${PLAN_ISSUE_KEYS.noAdds}`]);
	});

	it('warns about a wave that spends more of its life below 20% than below 35%', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'wave');
		Object.assign(plan.addWaves[0], { executeProportion20: 0.4, executeProportion35: 0.3 });
		expect(keys(plan)).toEqual([`wave0:warning:${PLAN_ISSUE_KEYS.executeOrder}`]);
	});

	it('reads no end on a spell hit, a strike or an area hit, while a window must end after its start', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'ability');
		const ability = plan.bossAbilities[0];
		withAbilityKind(ability, PlanBossAbility_Kind.AreaHit);
		ability.end!.at = ability.start!.at - 0.2;
		ability.end!.jitter = 0.5;
		for (const kind of [PlanBossAbility_Kind.SpellHit, PlanBossAbility_Kind.Strike, PlanBossAbility_Kind.AreaHit]) {
			ability.kind = kind;
			expect(keys(plan)).toEqual([]);
		}
		ability.kind = PlanBossAbility_Kind.Thrash;
		ability.end!.at = ability.start!.at;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.endBeforeStart}`]);
	});

	it('flags two moves at the same instant', () => {
		const plan = EncounterPlan.create();
		addEventAt(plan, 'move', 0.3);
		addEventAt(plan, 'move', 0.3);
		expect(keys(plan)).toEqual([`move1:error:${PLAN_ISSUE_KEYS.overlap}`]);
	});

	it('flags boss health that rises over time, in time order not list order', () => {
		const plan = EncounterPlan.create();
		addEventAt(plan, 'bossHealth', 0.5);
		plan.bossHealth[0].healthPercent = 40;
		addEventAt(plan, 'bossHealth', 0.2);
		plan.bossHealth[1].healthPercent = 80;
		expect(keys(plan)).toEqual([]);
		plan.bossHealth[1].healthPercent = 30;
		expect(keys(plan)).toEqual([`bossHealth0:error:${PLAN_ISSUE_KEYS.healthIncreases}`]);
	});

	it('flags the boss health keyframes the sim drops past its 16 curve points, and warns when its own end point is what falls off', () => {
		const curve = (plan: EncounterPlan) =>
			keys(plan).filter(k => k.endsWith(PLAN_ISSUE_KEYS.tooManyKeyframes) || k.endsWith(PLAN_ISSUE_KEYS.healthStopsShort));
		const dropped = (index: number) => `bossHealth${index}:error:${PLAN_ISSUE_KEYS.tooManyKeyframes}`;
		const stopsShort = (index: number) => `bossHealth${index}:warning:${PLAN_ISSUE_KEYS.healthStopsShort}`;
		// Listed latest first, so list order and time order differ.
		const plan = EncounterPlan.create();
		for (let step = 14; step >= 1; step--) {
			addEvent(plan, 'bossHealth');
			Object.assign(plan.bossHealth[plan.bossHealth.length - 1], { at: step / 20, healthPercent: 100 - step * 5 });
		}
		// 14 keyframes and the sim's own start and end points fill the 16.
		expect(curve(plan)).toEqual([]);

		plan.bossHealth.unshift(HealthKeyframe.create({ at: 15 / 20, healthPercent: 25 }));
		expect(curve(plan)).toEqual([stopsShort(0)]);

		plan.bossHealth.unshift(HealthKeyframe.create({ at: 16 / 20, healthPercent: 20 }));
		expect(curve(plan)).toEqual([dropped(0), stopsShort(1)]);

		// A first keyframe at the pull leaves the sim no start point to add.
		plan.bossHealth[15].at = 0;
		expect(curve(plan)).toEqual([stopsShort(0)]);

		// A last keyframe at 0% needs no end point.
		plan.bossHealth[0].healthPercent = 0;
		expect(curve(plan)).toEqual([]);

		addEventAt(plan, 'bossHealth', 0.95);
		plan.bossHealth[16].healthPercent = 0;
		expect(curve(plan)).toEqual([dropped(16)]);
	});

	it('flags the abilities the sim skips for doing nothing, and accepts a new one of any kind', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'ability');
		const ability = plan.bossAbilities[0];
		for (const kind of [PlanBossAbility_Kind.SpellHit, PlanBossAbility_Kind.Strike, PlanBossAbility_Kind.Thrash, PlanBossAbility_Kind.SwingTiming]) {
			ability.kind = kind;
			expect(keys(plan)).toEqual([]);
		}

		ability.kind = PlanBossAbility_Kind.Strike;
		ability.weaponPercent = 0;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.noWeaponDamage}`]);

		ability.kind = PlanBossAbility_Kind.Thrash;
		expect(keys(plan)).toEqual([]);
		ability.procChance = 0;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.noExtraAttacks}`]);
		ability.procChance = 20;
		ability.extraAttacks = 0;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.noExtraAttacks}`]);
	});

	it('accepts a new pulse and area hit, and flags the ones the sim skips or that roll min above max', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'ability');
		addEvent(plan, 'ability');
		const [pulse, areaHit] = plan.bossAbilities;
		withAbilityKind(pulse, PlanBossAbility_Kind.Pulse);
		withAbilityKind(areaHit, PlanBossAbility_Kind.AreaHit);
		expect(keys(plan)).toEqual([]);

		pulse.tickInterval = 0;
		areaHit.maxDamage = 0;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.noTickInterval}`, `ability1:error:${PLAN_ISSUE_KEYS.noDamage}`]);

		pulse.tickInterval = 1;
		pulse.minDamage = 800;
		areaHit.maxDamage = 2000;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.minAboveMax}`]);

		pulse.minDamage = 500;
		pulse.end!.at = pulse.start!.at;
		expect(keys(plan)).toEqual([`ability0:error:${PLAN_ISSUE_KEYS.endBeforeStart}`]);
	});

	it('names no target for a pulse or an area hit, whose target index the sim ignores', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'ability');
		addEvent(plan, 'ability');
		for (const [ability, kind] of [
			[plan.bossAbilities[0], PlanBossAbility_Kind.Pulse],
			[plan.bossAbilities[1], PlanBossAbility_Kind.SpellHit],
		] as const) {
			withAbilityKind(ability, kind);
			ability.targetIndex = -1;
		}
		expect(keys(plan)).toEqual([`ability1:error:${PLAN_ISSUE_KEYS.targetOutOfRange}`]);
	});

	it('warns about a wave at the very end of the fight, which never spawns', () => {
		const plan = EncounterPlan.create();
		addEventAt(plan, 'wave', 1);
		expect(keys(plan)).toEqual([`wave0:warning:${PLAN_ISSUE_KEYS.outsideFight}`]);
	});

	it("warns about a wave whose copies alive at once need more than the sim's 64 slots", () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'wave');
		// A spawn every 0.1 for a life of 0.25 keeps three copies up.
		Object.assign(plan.addWaves[0], {
			adds: Array.from({ length: 21 }, () => Target.create()),
			spawn: { at: 0, jitter: 0 },
			lifetime: 0.25,
			repeat: { interval: 0.1, jitter: 0, count: 0 },
		});
		expect(keys(plan)).toEqual([]);

		plan.addWaves[0].adds.push(Target.create());
		expect(keys(plan)).toEqual([`wave0:warning:${PLAN_ISSUE_KEYS.waveOverPool}`]);

		// Two occurrences are all a count of 1 has, however long they live.
		plan.addWaves[0].repeat!.count = 1;
		expect(keys(plan)).toEqual([]);

		// Jitter on the repeat brings the spawns closer, so a fourth copy overlaps.
		plan.addWaves[0].repeat!.count = 0;
		plan.addWaves[0].adds.pop();
		plan.addWaves[0].repeat!.jitter = 0.01;
		expect(keys(plan)).toEqual([`wave0:warning:${PLAN_ISSUE_KEYS.waveOverPool}`]);
	});

	it('flags a debuff amp whose wave was removed, and keeps the wave after it', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'wave');
		addEvent(plan, 'wave');
		addEvent(plan, 'amp');
		plan.damageAmps[0].type = DamageAmpType.DamageAmpTypeTargetDebuff;
		plan.damageAmps[0].targets!.waveIndices = [1];
		removeEvent(plan, { kind: 'wave', index: 0 });
		expect(keys(plan)).toEqual([]);
		removeEvent(plan, { kind: 'wave', index: 0 });
		expect(keys(plan)).toEqual([`amp0:error:${PLAN_ISSUE_KEYS.waveOutOfRange}`]);
	});

	it('flags targets out of range, including -1', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.2, 0.4, 2);
		addEvent(plan, 'ability');
		plan.bossAbilities[0].targetIndex = -1;
		addEvent(plan, 'amp');
		plan.damageAmps[0].targets!.targetIndices = [0, 3];
		expect(keys(plan, 2)).not.toContain(`amp0:error:${PLAN_ISSUE_KEYS.targetOutOfRange}`);
		plan.damageAmps[0].type = DamageAmpType.DamageAmpTypeTargetDebuff;
		plan.damageAmps[0].targets!.waveIndices = [0];
		const result = keys(plan, 2);
		expect(result).toContain(`downtime0:error:${PLAN_ISSUE_KEYS.targetOutOfRange}`);
		expect(result).toContain(`ability0:error:${PLAN_ISSUE_KEYS.targetOutOfRange}`);
		expect(result).toContain(`amp0:error:${PLAN_ISSUE_KEYS.targetOutOfRange}`);
		expect(result).toContain(`amp0:error:${PLAN_ISSUE_KEYS.waveOutOfRange}`);
	});

	it('reports info when every target is down at the same instant', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.2, 0.4, 0);
		downtime(plan, 0.3, 0.5, 1);
		const result = keys(plan, 2);
		expect(result).toContain(`downtime0:info:${PLAN_ISSUE_KEYS.allTargetsDown}`);
		expect(result).toContain(`downtime1:info:${PLAN_ISSUE_KEYS.allTargetsDown}`);
		expect(idleSpans(plan, 2)).toEqual([
			[0, 0.3, 0.4],
			[1, 0.3, 0.4],
		]);
		plan.downtimes[1].start!.at = 0.45;
		expect(keys(plan, 2).filter(k => k.includes(':info:'))).toEqual([]);
	});

	const idleSpans = (plan: EncounterPlan, targetCount = 1) =>
		validatePlan(plan, targetCount)
			.filter(issue => issue.key === PLAN_ISSUE_KEYS.allTargetsDown)
			.map(issue => [issue.ref.index, issue.span!.start, issue.span!.end]);

	const wave = (plan: EncounterPlan, at: number, lifetime: number, { adds = 1, interval = 0 } = {}) =>
		plan.addWaves.push(
			AddWave.create({
				adds: Array.from({ length: adds }, () => Target.create()),
				spawn: { at, jitter: 0 },
				lifetime,
				repeat: { interval, count: 0, jitter: 0 },
			}),
		);

	it('reports no idle stretch where an add is up for the whole downtime', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.4, 0.5);
		wave(plan, 0.35, 0.2);
		expect(idleSpans(plan)).toEqual([]);
	});

	it('names the part of a downtime no add covers, and each part when an add splits it', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.4, 0.5);
		wave(plan, 0.45, 0.1);
		expect(idleSpans(plan)).toEqual([[0, 0.4, 0.45]]);

		plan.addWaves[0].spawn!.at = 0.4375;
		plan.addWaves[0].lifetime = 0.03125;
		expect(idleSpans(plan)).toEqual([
			[0, 0.4, 0.4375],
			[0, 0.46875, 0.5],
		]);
	});

	it("counts each occurrence of a repeating wave as an add's life", () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.375, 0.5);
		wave(plan, 0.125, 0.0625, { interval: 0.25 });
		expect(idleSpans(plan)).toEqual([[0, 0.4375, 0.5]]);
	});

	it('counts no add of a wave the sim skips', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.4, 0.5);
		wave(plan, 0.4, 0.1, { adds: 0 });
		wave(plan, 0.4, 0.1, { adds: 65 });
		wave(plan, 0.4, 0, {});
		expect(idleSpans(plan)).toEqual([[0, 0.4, 0.5]]);
	});

	it('flags a wave with no lifetime', () => {
		const plan = EncounterPlan.create();
		addEventAt(plan, 'wave', 0.2);
		plan.addWaves[0].lifetime = 0;
		expect(keys(plan)).toEqual([`wave0:error:${PLAN_ISSUE_KEYS.endBeforeStart}`]);
	});

	it('applies each time its own jitter to the outside-fight check', () => {
		const plan = EncounterPlan.create();
		downtime(plan, 0.4, 0.9);
		plan.downtimes[0].start!.jitter = 0.3;
		expect(keys(plan).filter(k => k.includes(':warning:'))).toEqual([]);
		plan.downtimes[0].end!.jitter = 0.2;
		expect(keys(plan).filter(k => k.includes(':warning:'))).toEqual([`downtime0:warning:${PLAN_ISSUE_KEYS.outsideFight}`]);
	});
});
