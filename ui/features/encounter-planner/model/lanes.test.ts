import { EncounterPlan, PlanBossAbility_Kind, PlanTargets, SpellSchool } from '@generated/proto/common';
import { addEventAt } from '@testing/encounter_plan';
import { describe, expect, it } from 'vitest';

import type { PlanEventKind } from './lanes';
import {
	addEvent,
	defaultEvent,
	duplicateEvent,
	eventLabelKey,
	eventsOf,
	eventWindow,
	getEvent,
	healthAt,
	healthEndSegment,
	healthSegments,
	holdBlocker,
	holdDuringDowntime,
	holdSegment,
	laneLabelKey,
	moveSeconds,
	occurrences,
	pulseTicks,
	removeEvent,
	withAbilityKind,
} from './lanes';

const KINDS: PlanEventKind[] = ['downtime', 'bossHealth', 'ability', 'wave', 'move', 'amp'];

describe('lanes', () => {
	it('builds label keys under the planner namespace', () => {
		expect(laneLabelKey('waves')).toBe('settings_tab.encounter.planner.lanes.waves');
		expect(eventLabelKey('bossHealth')).toBe('settings_tab.encounter.planner.events.boss_health');
	});

	it('defaults every kind inside the fight', () => {
		for (const kind of KINDS) {
			const plan = EncounterPlan.create();
			const ref = addEvent(plan, kind);
			const window = eventWindow(plan, ref)!;
			const [from, to] = 'at' in window ? [window.at, window.at] : [window.start, window.end];
			expect(from).toBeGreaterThanOrEqual(0);
			expect(to).toBeLessThanOrEqual(1);
		}
	});

	it('uses the documented defaults', () => {
		const wave = defaultEvent('wave');
		expect([wave.adds.length, wave.lifetime, wave.executeProportion20, wave.executeProportion35, wave.killShare]).toEqual([1, 0.1, 0.2, 0.35, 0.25]);
		expect(defaultEvent('move').yards).toBe(5);
		expect(defaultEvent('amp').percent).toBe(10);
		expect(defaultEvent('ability').castTime).toBe(2);
		expect(defaultEvent('ability').weaponPercent).toBe(150);
		expect(defaultEvent('ability').procChance).toBe(20);
		expect(defaultEvent('ability').extraAttacks).toBe(1);
		const downtime = defaultEvent('downtime');
		expect(downtime.clearDebuffs).toBe(true);
		expect(downtime.end!.at - downtime.start!.at).toBeCloseTo(0.1);
	});

	it('adds a downtime on the target given, and the first one by default', () => {
		const plan = EncounterPlan.create();
		addEvent(plan, 'downtime', 1);
		addEvent(plan, 'downtime');
		expect(plan.downtimes.map(downtime => downtime.targetIndex)).toEqual([1, 0]);
	});

	it('adds, reads and removes events by list and index', () => {
		const plan = EncounterPlan.create();
		addEventAt(plan, 'move', 0.1);
		const ref = addEventAt(plan, 'move', 0.5);
		expect(ref).toEqual({ kind: 'move', index: 1 });
		expect(eventsOf(plan, 'move')).toHaveLength(2);
		expect(getEvent(plan, ref)).toBe(plan.moves[1]);
		removeEvent(plan, { kind: 'move', index: 0 });
		expect(plan.moves).toHaveLength(1);
		expect(getEvent(plan, ref)).toBeUndefined();
		removeEvent(plan, { kind: 'move', index: -1 });
		removeEvent(plan, { kind: 'move', index: 5 });
		expect(plan.moves).toHaveLength(1);
	});

	it('keeps the waves an amp names through a removal, and marks a removed one for validation', () => {
		const plan = EncounterPlan.create();
		for (let wave = 0; wave < 3; wave++) addEvent(plan, 'wave');
		addEvent(plan, 'amp');
		addEvent(plan, 'amp');
		plan.damageAmps[0].targets = PlanTargets.create({ targetIndices: [1], waveIndices: [0, 1, 2] });
		plan.damageAmps[1].targets = undefined;

		removeEvent(plan, { kind: 'wave', index: 1 });
		expect(plan.damageAmps[0].targets).toEqual(PlanTargets.create({ targetIndices: [1], waveIndices: [0, -1, 1] }));
		expect(plan.damageAmps[1].targets).toBeUndefined();

		// A copy goes at the end of the list, so no position moves.
		duplicateEvent(plan, { kind: 'wave', index: 0 }, 0.1);
		expect(plan.damageAmps[0].targets!.waveIndices).toEqual([0, -1, 1]);

		removeEvent(plan, { kind: 'wave', index: 9 });
		removeEvent(plan, { kind: 'move', index: 0 });
		expect(plan.damageAmps[0].targets!.waveIndices).toEqual([0, -1, 1]);
	});

	it('draws moves and keyframes as points and waves as spawn to spawn plus lifetime', () => {
		const plan = EncounterPlan.create();
		addEventAt(plan, 'move', 0.3);
		addEventAt(plan, 'bossHealth', 0.6);
		addEventAt(plan, 'wave', 0.2);
		expect(eventWindow(plan, { kind: 'move', index: 0 })).toEqual({ at: 0.3 });
		expect(eventWindow(plan, { kind: 'bossHealth', index: 0 })).toEqual({ at: 0.6 });
		expect(eventWindow(plan, { kind: 'wave', index: 0 })).toEqual({ start: 0.2, end: 0.30000000000000004 });
		expect(eventWindow(plan, { kind: 'wave', index: 4 })).toBeUndefined();
	});

	describe('occurrences', () => {
		it('returns one window without a repeat', () => {
			const plan = EncounterPlan.create();
			const ref = addEventAt(plan, 'amp', 0.2);
			expect(occurrences(plan, ref, 10)).toEqual([{ start: 0.2, end: 0.30000000000000004 }]);
		});

		it('repeats count times after the first, ignoring jitter', () => {
			const plan = EncounterPlan.create();
			const ref = addEventAt(plan, 'move', 0.1);
			plan.moves[0].repeat = { interval: 0.2, jitter: 0.05, count: 2 };
			expect(occurrences(plan, ref, 10).map(o => o.start)).toEqual([0.1, 0.30000000000000004, 0.5]);
		});

		it('fires a spell hit, a strike or an area hit as instants that repeat past its end, and a window as spans of its length', () => {
			const plan = EncounterPlan.create();
			const ref = addEventAt(plan, 'ability', 0.1);
			const ability = plan.bossAbilities[0];
			ability.end!.at = 0.15;
			for (const kind of [PlanBossAbility_Kind.SpellHit, PlanBossAbility_Kind.Strike, PlanBossAbility_Kind.AreaHit]) {
				ability.kind = kind;
				ability.repeat = { interval: 0.2, jitter: 0, count: 0 };
				expect(eventWindow(plan, ref)).toEqual({ at: 0.1 });
				expect(occurrences(plan, ref, 10).map(o => [o.start, o.end - o.start])).toEqual([0.1, 0.3, 0.5, 0.7, 0.9].map(at => [expect.closeTo(at), 0]));
				ability.repeat.count = 2;
				expect(occurrences(plan, ref, 10).map(o => o.start)).toEqual([0.1, 0.3, 0.5].map(at => expect.closeTo(at)));
			}

			ability.kind = PlanBossAbility_Kind.Thrash;
			ability.end!.at = 0.15;
			ability.repeat = { interval: 0.4, jitter: 0, count: 0 };
			expect(eventWindow(plan, ref)).toEqual({ start: 0.1, end: 0.15 });
			expect(occurrences(plan, ref, 10).map(o => [o.start, o.end - o.start])).toEqual([
				[0.1, expect.closeTo(0.05)],
				[expect.closeTo(0.5), expect.closeTo(0.05)],
				[expect.closeTo(0.9), expect.closeTo(0.05)],
			]);
		});

		it('fires a pulse as windows of its length', () => {
			const plan = EncounterPlan.create();
			const ref = addEventAt(plan, 'ability', 0.1);
			const ability = plan.bossAbilities[0];
			ability.end!.at = 0.25;
			withAbilityKind(ability, PlanBossAbility_Kind.Pulse);
			ability.repeat = { interval: 0.5, jitter: 0, count: 0 };
			expect(occurrences(plan, ref, 10).map(o => [o.start, o.end])).toEqual([
				[0.1, 0.25],
				[expect.closeTo(0.6), expect.closeTo(0.75)],
			]);
		});

		it('runs to the end of the fight with count 0 and caps at maxCount', () => {
			const plan = EncounterPlan.create();
			const ref = addEventAt(plan, 'move', 0.1);
			plan.moves[0].repeat = { interval: 0.25, jitter: 0, count: 0 };
			expect(occurrences(plan, ref, 10).map(o => o.start)).toEqual([0.1, 0.35, 0.6, 0.85]);
			expect(occurrences(plan, ref, 2)).toHaveLength(2);
			expect(occurrences(plan, ref, 0)).toEqual([]);
		});
	});

	describe('healthSegments', () => {
		const keyframes = (...points: Array<[number, number]>) =>
			EncounterPlan.create({ bossHealth: points.map(([at, healthPercent]) => ({ ...defaultEvent('bossHealth'), at, healthPercent })) });

		it('runs a segment from the keyframe before it in time, or the implied 0 s and 100 %, to the keyframe', () => {
			const plan = keyframes([0.5, 65], [0.25, 90]);
			expect(healthSegments(plan)).toEqual([
				{ index: 1, previous: undefined, fromPercent: 100, toPercent: 90, start: 0, end: 0.25 },
				{ index: 0, previous: 1, fromPercent: 90, toPercent: 65, start: 0.25, end: 0.5 },
			]);
		});

		it('clamps and flattens a rising curve the way the sim does', () => {
			const plan = keyframes([0.2, 50], [0.4, 80], [1.5, -5]);
			expect(healthSegments(plan).map(segment => [segment.toPercent, segment.end])).toEqual([
				[50, 0.2],
				[50, 0.4],
				[0, 1],
			]);
		});

		it('adds a read-only end segment only when the curve stops before the end above 0 %', () => {
			expect(healthEndSegment(keyframes([0.5, 65]))).toEqual({ start: 0.5, end: 1, fromPercent: 65 });
			expect(healthEndSegment(keyframes([0.5, 0]))).toBeUndefined();
			expect(healthEndSegment(keyframes([1, 20]))).toBeUndefined();
			expect(healthEndSegment(EncounterPlan.create())).toBeUndefined();
		});
	});

	it('walks a move through its yards at the player run speed, and its seconds otherwise', () => {
		expect(moveSeconds({ ...defaultEvent('move'), yards: 14, seconds: 9 })).toBe(2);
		expect(moveSeconds({ ...defaultEvent('move'), yards: 0, seconds: 9 })).toBe(9);
	});

	describe('duplicateEvent', () => {
		const KINDS_WITH_NAME: PlanEventKind[] = ['downtime', 'bossHealth', 'ability', 'wave', 'move', 'amp'];

		it("appends a copy shifted by the delta to the kind's list and returns its ref", () => {
			for (const kind of KINDS_WITH_NAME) {
				const plan = EncounterPlan.create();
				const original = addEventAt(plan, kind, 0.2);
				const copy = duplicateEvent(plan, original, 0.3);
				expect(copy).toEqual({ kind, index: 1 });
				const before = eventWindow(plan, original)!;
				const after = eventWindow(plan, copy)!;
				const start = (window: typeof before) => ('at' in window ? window.at : window.start);
				expect(start(after) - start(before)).toBeCloseTo(0.3);
				expect(eventsOf(plan, kind)[1]).not.toBe(eventsOf(plan, kind)[0]);
			}
		});

		it('keeps the copy inside the fight', () => {
			const plan = EncounterPlan.create();
			const ref = addEventAt(plan, 'downtime', 0.8);
			const copy = duplicateEvent(plan, ref, 0.5);
			expect(plan.downtimes[copy.index].end!.at).toBeCloseTo(1);
			expect(plan.downtimes[copy.index].start!.at).toBeCloseTo(0.9);
		});

		it('names the copy of a named event, and leaves an unnamed one unnamed', () => {
			const plan = EncounterPlan.create();
			const wave = addEventAt(plan, 'wave', 0.1);
			const plain = duplicateEvent(plan, wave, 0.1);
			expect(plan.addWaves[plain.index].name).toBe('');
			plan.addWaves[0].name = 'Add';
			const named = duplicateEvent(plan, wave, 0.1);
			expect(plan.addWaves[named.index].name).toContain('copy_name');
			const move = addEventAt(plan, 'move', 0.1);
			plan.moves[0].name = 'Run';
			expect(plan.moves[duplicateEvent(plan, move, 0.1).index].name).toContain('copy_name');
			expect(plan.addWaves[named.index].adds).toHaveLength(plan.addWaves[0].adds.length);
		});
	});

	describe('holding boss health', () => {
		const keyframes = (...points: Array<[number, number]>) =>
			EncounterPlan.create({ bossHealth: points.map(([at, healthPercent]) => ({ ...defaultEvent('bossHealth'), at, healthPercent })) });
		const withDowntime = (plan: EncounterPlan, start: number, end: number, targetIndex = 0) => {
			plan.downtimes.push({ ...defaultEvent('downtime'), targetIndex, start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
			return plan;
		};

		it('reads the curve with its own start and end points', () => {
			expect(healthAt(EncounterPlan.create(), 0.3)).toBe(100);
			const plan = keyframes([0.5, 50]);
			expect(healthAt(plan, 0)).toBe(100);
			expect(healthAt(plan, 0.25)).toBe(75);
			expect(healthAt(plan, 0.75)).toBe(25);
			expect(healthAt(plan, 1)).toBe(0);
		});

		it('cuts a flat part of at most 10 s, or half the segment, off a segment start at its start percent', () => {
			const plan = keyframes([0.25, 90], [0.3, 60]);
			const copy = holdSegment(plan, { kind: 'bossHealth', index: 1 }, 180)!;
			expect(copy).toEqual({ kind: 'bossHealth', index: 2 });
			expect([plan.bossHealth[2].at, plan.bossHealth[2].healthPercent]).toEqual([expect.closeTo(0.275), 90]);
			const long = holdSegment(plan, { kind: 'bossHealth', index: 0 }, 180)!;
			expect([plan.bossHealth[long.index].at, plan.bossHealth[long.index].healthPercent]).toEqual([expect.closeTo(10 / 180), 100]);
		});

		it('leaves a segment of no length alone', () => {
			const plan = keyframes([0.5, 50], [0.5, 40]);
			expect(holdSegment(plan, { kind: 'bossHealth', index: 1 }, 180)).toBeUndefined();
			expect(plan.bossHealth).toHaveLength(2);
		});

		it('holds the health the curve has at a boss downtime start across it, the implied start included', () => {
			const plan = withDowntime(keyframes([0.25, 90], [0.75, 30]), 0.4, 0.5);
			const refs = holdDuringDowntime(plan, 0);
			expect(refs).toEqual([
				{ kind: 'bossHealth', index: 2 },
				{ kind: 'bossHealth', index: 3 },
			]);
			expect(plan.bossHealth.slice(2).map(keyframe => [keyframe.at, keyframe.healthPercent])).toEqual([
				[0.4, expect.closeTo(72)],
				[0.5, expect.closeTo(72)],
			]);
			const implied = withDowntime(keyframes([0.5, 50]), 0.1, 0.2);
			holdDuringDowntime(implied, 0);
			expect(implied.bossHealth[1].healthPercent).toBeCloseTo(90);
		});

		it('names the keyframe in the way and holds nothing then, nor for another target than the boss', () => {
			const blocked = withDowntime(keyframes([0.45, 80]), 0.4, 0.5);
			expect(holdBlocker(blocked, 0)).toEqual({ kind: 'bossHealth', index: 0 });
			expect(holdDuringDowntime(blocked, 0)).toEqual([]);
			expect(blocked.bossHealth).toHaveLength(1);
			const atStart = withDowntime(keyframes([0.4, 80], [0.6, 70]), 0.4, 0.5);
			expect(holdBlocker(atStart, 0)).toBeUndefined();
			const heldToEnd = withDowntime(keyframes([0.4, 80], [0.5, 80]), 0.4, 0.5);
			expect(holdBlocker(heldToEnd, 0)).toBeUndefined();
			// The curve falls to the keyframe on the downtime's end, whatever a new keyframe beside it says.
			const fallsToEnd = withDowntime(keyframes([0.4, 80], [0.5, 70]), 0.4, 0.5);
			expect(holdBlocker(fallsToEnd, 0)).toEqual({ kind: 'bossHealth', index: 1 });
			expect(holdDuringDowntime(fallsToEnd, 0)).toEqual([]);
			const other = withDowntime(keyframes([0.9, 80]), 0.4, 0.5, 1);
			expect(holdDuringDowntime(other, 0)).toEqual([]);
		});
	});

	describe('area damage', () => {
		const switched = (kind: PlanBossAbility_Kind) => {
			const ability = defaultEvent('ability');
			withAbilityKind(ability, kind);
			return ability;
		};

		it('gives a new pulse and a new area hit fire damage, a pulse 1 s ticks ± 0.2 s, and both Resistible on and Avoidable off', () => {
			expect(switched(PlanBossAbility_Kind.Pulse)).toMatchObject({
				kind: PlanBossAbility_Kind.Pulse,
				school: SpellSchool.SpellSchoolFire,
				minDamage: 500,
				maxDamage: 700,
				tickInterval: 1,
				tickJitter: 0.2,
				resistible: true,
				avoidable: false,
			});
			expect(switched(PlanBossAbility_Kind.AreaHit)).toMatchObject({
				school: SpellSchool.SpellSchoolFire,
				minDamage: 1500,
				maxDamage: 2000,
				resistible: true,
				avoidable: false,
			});
		});

		it('keeps damage and an interval already set, and Resistible off on a switch between the two area kinds', () => {
			const ability = { ...defaultEvent('ability'), school: SpellSchool.SpellSchoolShadow, minDamage: 100, maxDamage: 200 };
			withAbilityKind(ability, PlanBossAbility_Kind.Pulse);
			expect([ability.school, ability.minDamage, ability.maxDamage]).toEqual([SpellSchool.SpellSchoolShadow, 100, 200]);
			ability.tickInterval = 3;
			ability.resistible = false;
			withAbilityKind(ability, PlanBossAbility_Kind.AreaHit);
			withAbilityKind(ability, PlanBossAbility_Kind.Pulse);
			expect([ability.tickInterval, ability.resistible]).toEqual([3, false]);
		});

		it('ends a window kind a default width after its start when the end it is switched with is not after the start', () => {
			for (const kind of [PlanBossAbility_Kind.Thrash, PlanBossAbility_Kind.SwingTiming, PlanBossAbility_Kind.Pulse]) {
				const ability = { ...defaultEvent('ability'), start: { at: 0.6, jitter: 0 }, end: { at: 0.5, jitter: 0 } };
				withAbilityKind(ability, kind);
				expect(ability.end.at).toBeCloseTo(0.7);
				ability.end.at = 0.9;
				withAbilityKind(ability, PlanBossAbility_Kind.SpellHit);
				withAbilityKind(ability, kind);
				expect(ability.end.at).toBe(0.9);
			}
		});

		it('ticks at a window start and every interval before its end, once for a window shorter than one interval', () => {
			expect(pulseTicks({ start: 0.1, end: 0.2 }, 3, 100, 10)).toEqual([0.1, expect.closeTo(0.13), expect.closeTo(0.16), expect.closeTo(0.19)]);
			expect(pulseTicks({ start: 0.1, end: 0.11 }, 3, 100, 10)).toEqual([0.1]);
			expect(pulseTicks({ start: 0.1, end: 0.2 }, 1, 100, 3)).toHaveLength(3);
			expect(pulseTicks({ start: 0.1, end: 0.2 }, 0, 100, 10)).toEqual([]);
		});
	});
});
