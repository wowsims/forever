import { DesignBossAbility_Kind, DesignTargets, EncounterDesign } from '@generated/proto/common';
import { addEventAt } from '@testing/encounter_design';
import { describe, expect, it } from 'vitest';

import type { DesignEventKind } from './lanes';
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
	removeEvent,
} from './lanes';

const KINDS: DesignEventKind[] = ['downtime', 'bossHealth', 'ability', 'wave', 'move', 'amp'];

describe('lanes', () => {
	it('builds label keys under the designer namespace', () => {
		expect(laneLabelKey('waves')).toBe('settings_tab.encounter.designer.lanes.waves');
		expect(eventLabelKey('bossHealth')).toBe('settings_tab.encounter.designer.events.boss_health');
	});

	it('defaults every kind inside the fight', () => {
		for (const kind of KINDS) {
			const design = EncounterDesign.create();
			const ref = addEvent(design, kind);
			const window = eventWindow(design, ref)!;
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

	it('adds, reads and removes events by list and index', () => {
		const design = EncounterDesign.create();
		addEventAt(design, 'move', 0.1);
		const ref = addEventAt(design, 'move', 0.5);
		expect(ref).toEqual({ kind: 'move', index: 1 });
		expect(eventsOf(design, 'move')).toHaveLength(2);
		expect(getEvent(design, ref)).toBe(design.moves[1]);
		removeEvent(design, { kind: 'move', index: 0 });
		expect(design.moves).toHaveLength(1);
		expect(getEvent(design, ref)).toBeUndefined();
		removeEvent(design, { kind: 'move', index: -1 });
		removeEvent(design, { kind: 'move', index: 5 });
		expect(design.moves).toHaveLength(1);
	});

	it('keeps the waves an amp names through a removal, and marks a removed one for validation', () => {
		const design = EncounterDesign.create();
		for (let wave = 0; wave < 3; wave++) addEvent(design, 'wave');
		addEvent(design, 'amp');
		addEvent(design, 'amp');
		design.damageAmps[0].targets = DesignTargets.create({ targetIndices: [1], waveIndices: [0, 1, 2] });
		design.damageAmps[1].targets = undefined;

		removeEvent(design, { kind: 'wave', index: 1 });
		expect(design.damageAmps[0].targets).toEqual(DesignTargets.create({ targetIndices: [1], waveIndices: [0, -1, 1] }));
		expect(design.damageAmps[1].targets).toBeUndefined();

		// A copy goes at the end of the list, so no position moves.
		duplicateEvent(design, { kind: 'wave', index: 0 }, 0.1);
		expect(design.damageAmps[0].targets!.waveIndices).toEqual([0, -1, 1]);

		removeEvent(design, { kind: 'wave', index: 9 });
		removeEvent(design, { kind: 'move', index: 0 });
		expect(design.damageAmps[0].targets!.waveIndices).toEqual([0, -1, 1]);
	});

	it('draws moves and keyframes as points and waves as spawn to spawn plus lifetime', () => {
		const design = EncounterDesign.create();
		addEventAt(design, 'move', 0.3);
		addEventAt(design, 'bossHealth', 0.6);
		addEventAt(design, 'wave', 0.2);
		expect(eventWindow(design, { kind: 'move', index: 0 })).toEqual({ at: 0.3 });
		expect(eventWindow(design, { kind: 'bossHealth', index: 0 })).toEqual({ at: 0.6 });
		expect(eventWindow(design, { kind: 'wave', index: 0 })).toEqual({ start: 0.2, end: 0.30000000000000004 });
		expect(eventWindow(design, { kind: 'wave', index: 4 })).toBeUndefined();
	});

	describe('occurrences', () => {
		it('returns one window without a repeat', () => {
			const design = EncounterDesign.create();
			const ref = addEventAt(design, 'amp', 0.2);
			expect(occurrences(design, ref, 10)).toEqual([{ start: 0.2, end: 0.30000000000000004 }]);
		});

		it('repeats count times after the first, ignoring jitter', () => {
			const design = EncounterDesign.create();
			const ref = addEventAt(design, 'move', 0.1);
			design.moves[0].repeat = { interval: 0.2, jitter: 0.05, count: 2 };
			expect(occurrences(design, ref, 10).map(o => o.start)).toEqual([0.1, 0.30000000000000004, 0.5]);
		});

		it('fires a spell hit or a strike as instants that stop at its end, and a window as spans of its length', () => {
			const design = EncounterDesign.create();
			const ref = addEventAt(design, 'ability', 0.1);
			const ability = design.bossAbilities[0];
			ability.end!.at = 0.35;
			ability.repeat = { interval: 0.1, jitter: 0, count: 0 };
			for (const kind of [DesignBossAbility_Kind.SpellHit, DesignBossAbility_Kind.Strike]) {
				ability.kind = kind;
				expect(eventWindow(design, ref)).toEqual({ start: 0.1, end: 0.35 });
				expect(occurrences(design, ref, 10).map(o => [o.start, o.end])).toEqual([
					[0.1, 0.1],
					[expect.closeTo(0.2), expect.closeTo(0.2)],
					[expect.closeTo(0.3), expect.closeTo(0.3)],
				]);
			}
			ability.end!.at = 0.05;
			expect(occurrences(design, ref, 10)).toEqual([]);

			ability.kind = DesignBossAbility_Kind.Thrash;
			ability.end!.at = 0.15;
			ability.repeat = { interval: 0.4, jitter: 0, count: 0 };
			expect(eventWindow(design, ref)).toEqual({ start: 0.1, end: 0.15 });
			expect(occurrences(design, ref, 10).map(o => [o.start, o.end - o.start])).toEqual([
				[0.1, expect.closeTo(0.05)],
				[expect.closeTo(0.5), expect.closeTo(0.05)],
				[expect.closeTo(0.9), expect.closeTo(0.05)],
			]);
		});

		it('runs to the end of the fight with count 0 and caps at maxCount', () => {
			const design = EncounterDesign.create();
			const ref = addEventAt(design, 'move', 0.1);
			design.moves[0].repeat = { interval: 0.25, jitter: 0, count: 0 };
			expect(occurrences(design, ref, 10).map(o => o.start)).toEqual([0.1, 0.35, 0.6, 0.85]);
			expect(occurrences(design, ref, 2)).toHaveLength(2);
			expect(occurrences(design, ref, 0)).toEqual([]);
		});
	});

	describe('healthSegments', () => {
		const keyframes = (...points: Array<[number, number]>) =>
			EncounterDesign.create({ bossHealth: points.map(([at, healthPercent]) => ({ ...defaultEvent('bossHealth'), at, healthPercent })) });

		it('runs a segment from the keyframe before it in time, or the implied 0 s and 100 %, to the keyframe', () => {
			const design = keyframes([0.5, 65], [0.25, 90]);
			expect(healthSegments(design)).toEqual([
				{ index: 1, previous: undefined, fromPercent: 100, toPercent: 90, start: 0, end: 0.25 },
				{ index: 0, previous: 1, fromPercent: 90, toPercent: 65, start: 0.25, end: 0.5 },
			]);
		});

		it('clamps and flattens a rising curve the way the sim does', () => {
			const design = keyframes([0.2, 50], [0.4, 80], [1.5, -5]);
			expect(healthSegments(design).map(segment => [segment.toPercent, segment.end])).toEqual([
				[50, 0.2],
				[50, 0.4],
				[0, 1],
			]);
		});

		it('adds a read-only end segment only when the curve stops before the end above 0 %', () => {
			expect(healthEndSegment(keyframes([0.5, 65]))).toEqual({ start: 0.5, end: 1, fromPercent: 65 });
			expect(healthEndSegment(keyframes([0.5, 0]))).toBeUndefined();
			expect(healthEndSegment(keyframes([1, 20]))).toBeUndefined();
			expect(healthEndSegment(EncounterDesign.create())).toBeUndefined();
		});
	});

	it('walks a move through its yards at the player run speed, and its seconds otherwise', () => {
		expect(moveSeconds({ ...defaultEvent('move'), yards: 14, seconds: 9 })).toBe(2);
		expect(moveSeconds({ ...defaultEvent('move'), yards: 0, seconds: 9 })).toBe(9);
	});

	describe('duplicateEvent', () => {
		const KINDS_WITH_NAME: DesignEventKind[] = ['downtime', 'bossHealth', 'ability', 'wave', 'move', 'amp'];

		it("appends a copy shifted by the delta to the kind's list and returns its ref", () => {
			for (const kind of KINDS_WITH_NAME) {
				const design = EncounterDesign.create();
				const original = addEventAt(design, kind, 0.2);
				const copy = duplicateEvent(design, original, 0.3);
				expect(copy).toEqual({ kind, index: 1 });
				const before = eventWindow(design, original)!;
				const after = eventWindow(design, copy)!;
				const start = (window: typeof before) => ('at' in window ? window.at : window.start);
				expect(start(after) - start(before)).toBeCloseTo(0.3);
				expect(eventsOf(design, kind)[1]).not.toBe(eventsOf(design, kind)[0]);
			}
		});

		it('keeps the copy inside the fight', () => {
			const design = EncounterDesign.create();
			const ref = addEventAt(design, 'downtime', 0.8);
			const copy = duplicateEvent(design, ref, 0.5);
			expect(design.downtimes[copy.index].end!.at).toBeCloseTo(1);
			expect(design.downtimes[copy.index].start!.at).toBeCloseTo(0.9);
		});

		it('names the copy of a named event, and leaves an unnamed one unnamed', () => {
			const design = EncounterDesign.create();
			const wave = addEventAt(design, 'wave', 0.1);
			const plain = duplicateEvent(design, wave, 0.1);
			expect(design.addWaves[plain.index].name).toBe('');
			design.addWaves[0].name = 'Add';
			const named = duplicateEvent(design, wave, 0.1);
			expect(design.addWaves[named.index].name).toContain('copy_name');
			const move = addEventAt(design, 'move', 0.1);
			design.moves[0].name = 'Run';
			expect(design.moves[duplicateEvent(design, move, 0.1).index].name).toContain('copy_name');
			expect(design.addWaves[named.index].adds).toHaveLength(design.addWaves[0].adds.length);
		});
	});

	describe('holding boss health', () => {
		const keyframes = (...points: Array<[number, number]>) =>
			EncounterDesign.create({ bossHealth: points.map(([at, healthPercent]) => ({ ...defaultEvent('bossHealth'), at, healthPercent })) });
		const withDowntime = (design: EncounterDesign, start: number, end: number, targetIndex = 0) => {
			design.downtimes.push({ ...defaultEvent('downtime'), targetIndex, start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
			return design;
		};

		it('reads the curve with its own start and end points', () => {
			expect(healthAt(EncounterDesign.create(), 0.3)).toBe(100);
			const design = keyframes([0.5, 50]);
			expect(healthAt(design, 0)).toBe(100);
			expect(healthAt(design, 0.25)).toBe(75);
			expect(healthAt(design, 0.75)).toBe(25);
			expect(healthAt(design, 1)).toBe(0);
		});

		it('cuts a flat part of at most 10 s, or half the segment, off a segment start at its start percent', () => {
			const design = keyframes([0.25, 90], [0.3, 60]);
			const copy = holdSegment(design, { kind: 'bossHealth', index: 1 }, 180)!;
			expect(copy).toEqual({ kind: 'bossHealth', index: 2 });
			expect([design.bossHealth[2].at, design.bossHealth[2].healthPercent]).toEqual([expect.closeTo(0.275), 90]);
			const long = holdSegment(design, { kind: 'bossHealth', index: 0 }, 180)!;
			expect([design.bossHealth[long.index].at, design.bossHealth[long.index].healthPercent]).toEqual([expect.closeTo(10 / 180), 100]);
		});

		it('leaves a segment of no length alone', () => {
			const design = keyframes([0.5, 50], [0.5, 40]);
			expect(holdSegment(design, { kind: 'bossHealth', index: 1 }, 180)).toBeUndefined();
			expect(design.bossHealth).toHaveLength(2);
		});

		it('holds the health the curve has at a boss downtime start across it, the implied start included', () => {
			const design = withDowntime(keyframes([0.25, 90], [0.75, 30]), 0.4, 0.5);
			const refs = holdDuringDowntime(design, 0);
			expect(refs).toEqual([
				{ kind: 'bossHealth', index: 2 },
				{ kind: 'bossHealth', index: 3 },
			]);
			expect(design.bossHealth.slice(2).map(keyframe => [keyframe.at, keyframe.healthPercent])).toEqual([
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
});
