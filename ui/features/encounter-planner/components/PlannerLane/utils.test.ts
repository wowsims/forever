import { EncounterPlan, PlanBossAbility_Kind, PlanBossAbility_Source } from '@generated/proto/common';
import { eventAt } from '@testing/encounter_plan';
import { describe, expect, it } from 'vitest';

import { defaultEvent } from '../../model/lanes';
import { laneEvents, withWindow } from './utils';

const downtime = (start: number, end: number) => ({ ...defaultEvent('downtime'), start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
const BASE = 180;
const rowsByStart = (plan: EncounterPlan, lane: Parameters<typeof laneEvents>[1]) => laneEvents(plan, lane, BASE).map(event => [event.window.start, event.row]);

describe('laneEvents', () => {
	it('puts windows that do not overlap on one row', () => {
		const plan = EncounterPlan.create({ downtimes: [downtime(0.5, 0.6), downtime(0, 0.2), downtime(0.3, 0.4)] });
		expect(rowsByStart(plan, 'boss')).toEqual([
			[0, 0],
			[0.3, 0],
			[0.5, 0],
		]);
	});

	it('moves an overlapping window to the next free row', () => {
		const plan = EncounterPlan.create({ downtimes: [downtime(0, 0.5), downtime(0.2, 0.3), downtime(0.4, 0.6), downtime(0.6, 0.7)] });
		expect(rowsByStart(plan, 'boss')).toEqual([
			[0, 0],
			[0.2, 1],
			[0.4, 1],
			[0.6, 0],
		]);
	});

	it('draws the boss health keyframes as segments on the first row and the downtimes below them', () => {
		const plan = EncounterPlan.create({
			downtimes: [eventAt('downtime', 0.4)],
			bossHealth: [
				{ ...eventAt('bossHealth', 0.5), healthPercent: 65 },
				{ ...eventAt('bossHealth', 0.25), healthPercent: 90 },
			],
		});
		const events = laneEvents(plan, 'boss', BASE);
		expect(events.map(event => [event.ref.kind, event.ref.index, event.window.start, event.window.end, event.row, event.startFixed])).toEqual([
			['bossHealth', 1, 0, 0.25, 0, true],
			['bossHealth', 0, 0.25, 0.5, 0, false],
			['downtime', 0, 0.4, expect.closeTo(0.5), 1, false],
		]);
	});

	it("puts a target's downtimes on its row, and the boss health and the downtimes of no target on the first target's", () => {
		const on = (targetIndex: number, start: number) => ({ ...downtime(start, start + 0.1), targetIndex });
		const plan = EncounterPlan.create({
			downtimes: [on(1, 0.1), on(0, 0.3), on(2, 0.5), on(-1, 0.7)],
			bossHealth: [{ ...eventAt('bossHealth', 0.5), healthPercent: 50 }],
		});
		const row = (index: number) => laneEvents(plan, 'boss', BASE, { index, count: 2 }).map(event => [event.ref.kind, event.ref.index, event.row]);
		expect(row(0)).toEqual([
			['bossHealth', 0, 0],
			['downtime', 1, 1],
			['downtime', 2, 1],
			['downtime', 3, 1],
		]);
		expect(row(1)).toEqual([['downtime', 0, 0]]);
	});

	it('draws a move from its start to the end of its walk, in seconds or in yards at the run speed', () => {
		const seconds = { ...eventAt('move', 0.1), yards: 0, seconds: 18 };
		const yards = { ...eventAt('move', 0.5), yards: 70 };
		const plan = EncounterPlan.create({ moves: [seconds, yards] });
		expect(laneEvents(plan, 'moves', BASE).map(event => [event.window.start, event.window.end])).toEqual([
			[0.1, expect.closeTo(0.2)],
			[0.5, expect.closeTo(0.5 + 10 / BASE)],
		]);
	});

	it('keeps the repeats after the first occurrence at the width of the first and packs around their full reach', () => {
		const repeating = { ...eventAt('move', 0.1), yards: 0, seconds: 18, repeat: { interval: 0.3, jitter: 0, count: 2 } };
		const plan = EncounterPlan.create({ moves: [repeating, eventAt('move', 0.5)] });
		const [first, second] = laneEvents(plan, 'moves', BASE);
		expect(first.repeats.map(repeat => [repeat.start, repeat.end])).toEqual([
			[expect.closeTo(0.4), expect.closeTo(0.5)],
			[expect.closeTo(0.7), expect.closeTo(0.8)],
		]);
		expect(second.row).toBe(1);
	});

	it('draws a spell hit, a strike or an area hit 10 s wide at its start whatever its end, a thrash from its start to its end, and every repeat as a ghost', () => {
		const repeat = { interval: 0.25, jitter: 0, count: 0 };
		const ability = (kind: PlanBossAbility_Kind) => ({ ...eventAt('ability', 0.2), kind, repeat, end: { at: 0.21, jitter: 0 } });
		const plan = EncounterPlan.create({
			bossAbilities: [
				ability(PlanBossAbility_Kind.SpellHit),
				ability(PlanBossAbility_Kind.Strike),
				ability(PlanBossAbility_Kind.AreaHit),
				{ ...eventAt('ability', 0.2), kind: PlanBossAbility_Kind.Thrash, repeat },
			],
			addWaves: [eventAt('wave', 0.2)],
		});
		const events = laneEvents(plan, 'abilities', BASE).sort((a, b) => a.ref.index - b.ref.index);
		const width = 10 / BASE;
		for (const instant of events.slice(0, 3)) {
			expect([instant.window.start, instant.window.end, instant.tickSeconds, instant.openEnd, instant.fixedWidth]).toEqual([
				0.2,
				expect.closeTo(0.2 + width),
				0,
				true,
				true,
			]);
			expect(instant.repeats.map(ghost => [ghost.start, ghost.end])).toEqual(
				[0.45, 0.7, 0.95].map(at => [expect.closeTo(at), expect.closeTo(at + width)]),
			);
		}
		const thrash = events[3];
		expect([thrash.window.end, thrash.repeats.length, thrash.tickSeconds, thrash.fixedWidth]).toEqual([expect.closeTo(0.3), 3, 0, false]);
		expect(laneEvents(plan, 'waves', BASE).map(event => event.openEnd)).toEqual([true]);
	});

	it("gives a pulse's window and each of its repeats the pulse's tick interval", () => {
		const pulse = {
			...defaultEvent('ability'),
			kind: PlanBossAbility_Kind.Pulse,
			start: { at: 0.1, jitter: 0 },
			end: { at: 0.14, jitter: 0 },
			repeat: { interval: 0.89, jitter: 0, count: 0 },
			tickInterval: 3,
		};
		const [ticking] = laneEvents(EncounterPlan.create({ bossAbilities: [pulse] }), 'abilities', BASE);
		expect(ticking.repeats).toHaveLength(1);
		expect([ticking.tickSeconds, ticking.castSeconds]).toEqual([3, 0]);
	});

	it("starts a boss pulse's ticks at its cast's end, and draws none when the cast outlasts the window", () => {
		// The window is 0.04 of a 180 s fight, 7.2 s long.
		const pulse = (castTime: number) => ({
			...defaultEvent('ability'),
			kind: PlanBossAbility_Kind.Pulse,
			source: PlanBossAbility_Source.Boss,
			castTime,
			start: { at: 0.1, jitter: 0 },
			end: { at: 0.14, jitter: 0 },
			tickInterval: 3,
		});
		const [casting, outlasting] = laneEvents(EncounterPlan.create({ bossAbilities: [pulse(2), pulse(8)] }), 'abilities', BASE).sort(
			(a, b) => a.ref.index - b.ref.index,
		);
		expect([casting.tickSeconds, casting.castSeconds]).toEqual([3, 2]);
		expect([outlasting.tickSeconds, outlasting.castSeconds]).toEqual([0, 0]);
	});
});

describe('withWindow', () => {
	it('writes the window back into the event the kind stores it in, on a copy', () => {
		const plan = EncounterPlan.create({ addWaves: [eventAt('wave', 0.2)], moves: [eventAt('move', 0.2)] });
		const wave = withWindow(plan, { kind: 'wave', index: 0 }, { start: 0.3, end: 0.45 }, BASE).addWaves[0];
		expect(wave.spawn?.at).toBe(0.3);
		expect(wave.lifetime).toBeCloseTo(0.15);
		expect(withWindow(plan, { kind: 'move', index: 0 }, { start: 0.6, end: 0.6 + 5 / 7 / BASE }, BASE).moves[0].at?.at).toBe(0.6);
		expect(plan.addWaves[0].spawn?.at).toBe(0.2);
	});

	it("sets a window ability's start and end from the window, and only a spell hit's start", () => {
		const plan = EncounterPlan.create({
			bossAbilities: [eventAt('ability', 0.2), { ...eventAt('ability', 0.2), kind: PlanBossAbility_Kind.SwingTiming }],
		});
		const moved = (index: number) => withWindow(plan, { kind: 'ability', index }, { start: 0.5, end: 0.7 }, BASE).bossAbilities[index];
		expect([moved(0).start?.at, moved(0).end?.at]).toEqual([0.5, expect.closeTo(0.3)]);
		expect([moved(1).start?.at, moved(1).end?.at]).toEqual([0.5, 0.7]);
	});

	it('stretches the walk of a move through the unit it holds and leaves it alone on a plain move', () => {
		const plan = EncounterPlan.create({ moves: [eventAt('move', 0.1), { ...eventAt('move', 0.1), yards: 0, seconds: 4 }] });
		const length = (seconds: number) => ({ start: 0.1, end: 0.1 + seconds / BASE });
		const yards = withWindow(plan, { kind: 'move', index: 0 }, length(2), BASE).moves[0];
		expect([yards.yards, yards.seconds]).toEqual([expect.closeTo(14), 0]);
		const timed = withWindow(plan, { kind: 'move', index: 1 }, length(6), BASE).moves[1];
		expect([timed.yards, timed.seconds]).toEqual([0, expect.closeTo(6)]);
		const moved = withWindow(plan, { kind: 'move', index: 0 }, { start: 0.3, end: 0.3 + 5 / 7 / BASE }, BASE).moves[0];
		expect([moved.at?.at, moved.yards]).toEqual([0.3, 5]);
	});

	describe('boss health segments', () => {
		const plan = EncounterPlan.create({
			bossHealth: [
				{ ...defaultEvent('bossHealth'), at: 0.25, healthPercent: 90 },
				{ ...defaultEvent('bossHealth'), at: 0.5, healthPercent: 65 },
				{ ...defaultEvent('bossHealth'), at: 0.75, healthPercent: 30 },
			],
		});
		const ats = (next: EncounterPlan) => next.bossHealth.map(keyframe => keyframe.at);

		it('moves the keyframe at the segment end with the end handle, between its neighbours', () => {
			expect(ats(withWindow(plan, { kind: 'bossHealth', index: 1 }, { start: 0.25, end: 0.6 }, BASE))).toEqual([0.25, 0.6, 0.75]);
			expect(ats(withWindow(plan, { kind: 'bossHealth', index: 1 }, { start: 0.25, end: 0.9 }, BASE))).toEqual([0.25, 0.75, 0.75]);
		});

		it('moves the keyframe before the segment with the start handle and keeps the implied start fixed', () => {
			expect(ats(withWindow(plan, { kind: 'bossHealth', index: 1 }, { start: 0.4, end: 0.5 }, BASE))).toEqual([0.4, 0.5, 0.75]);
			expect(ats(withWindow(plan, { kind: 'bossHealth', index: 0 }, { start: 0.1, end: 0.25 }, BASE))).toEqual([0.25, 0.5, 0.75]);
		});

		it('moves only the keyframe at the end when the whole segment moves', () => {
			expect(ats(withWindow(plan, { kind: 'bossHealth', index: 1 }, { start: 0.3, end: 0.55 }, BASE))).toEqual([0.25, 0.55, 0.75]);
		});
	});
});
