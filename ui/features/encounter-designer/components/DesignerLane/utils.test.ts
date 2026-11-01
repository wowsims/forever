import { DesignBossAbility_Kind, EncounterDesign } from '@generated/proto/common';
import { eventAt } from '@testing/encounter_design';
import { describe, expect, it } from 'vitest';

import { defaultEvent } from '../../model/lanes';
import { laneEvents, withWindow } from './utils';

const downtime = (start: number, end: number) => ({ ...defaultEvent('downtime'), start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
const BASE = 180;
const rowsByStart = (design: EncounterDesign, lane: Parameters<typeof laneEvents>[1]) =>
	laneEvents(design, lane, BASE).map(event => [event.window.start, event.row]);

describe('laneEvents', () => {
	it('puts windows that do not overlap on one row', () => {
		const design = EncounterDesign.create({ downtimes: [downtime(0.5, 0.6), downtime(0, 0.2), downtime(0.3, 0.4)] });
		expect(rowsByStart(design, 'boss')).toEqual([
			[0, 0],
			[0.3, 0],
			[0.5, 0],
		]);
	});

	it('moves an overlapping window to the next free row', () => {
		const design = EncounterDesign.create({ downtimes: [downtime(0, 0.5), downtime(0.2, 0.3), downtime(0.4, 0.6), downtime(0.6, 0.7)] });
		expect(rowsByStart(design, 'boss')).toEqual([
			[0, 0],
			[0.2, 1],
			[0.4, 1],
			[0.6, 0],
		]);
	});

	it('draws the boss health keyframes as segments on the first row and the downtimes below them', () => {
		const design = EncounterDesign.create({
			downtimes: [eventAt('downtime', 0.4)],
			bossHealth: [
				{ ...eventAt('bossHealth', 0.5), healthPercent: 65 },
				{ ...eventAt('bossHealth', 0.25), healthPercent: 90 },
			],
		});
		const events = laneEvents(design, 'boss', BASE);
		expect(events.map(event => [event.ref.kind, event.ref.index, event.window.start, event.window.end, event.row, event.startFixed])).toEqual([
			['bossHealth', 1, 0, 0.25, 0, true],
			['bossHealth', 0, 0.25, 0.5, 0, false],
			['downtime', 0, 0.4, expect.closeTo(0.5), 1, false],
		]);
	});

	it('draws a move from its start to the end of its walk, in seconds or in yards at the run speed', () => {
		const seconds = { ...eventAt('move', 0.1), yards: 0, seconds: 18 };
		const yards = { ...eventAt('move', 0.5), yards: 70 };
		const design = EncounterDesign.create({ moves: [seconds, yards] });
		expect(laneEvents(design, 'moves', BASE).map(event => [event.window.start, event.window.end])).toEqual([
			[0.1, expect.closeTo(0.2)],
			[0.5, expect.closeTo(0.5 + 10 / BASE)],
		]);
	});

	it('keeps the repeats after the first occurrence at the width of the first and packs around their full reach', () => {
		const repeating = { ...eventAt('move', 0.1), yards: 0, seconds: 18, repeat: { interval: 0.3, jitter: 0, count: 2 } };
		const design = EncounterDesign.create({ moves: [repeating, eventAt('move', 0.5)] });
		const [first, second] = laneEvents(design, 'moves', BASE);
		expect(first.repeats.map(repeat => [repeat.start, repeat.end])).toEqual([
			[expect.closeTo(0.4), expect.closeTo(0.5)],
			[expect.closeTo(0.7), expect.closeTo(0.8)],
		]);
		expect(second.row).toBe(1);
	});

	it('draws every ability from its start to its end, with the occurrences of a spell hit or a strike as ticks and a thrash repeating as ghosts', () => {
		const repeat = { interval: 0.1, jitter: 0, count: 0 };
		const ability = (kind: DesignBossAbility_Kind) => ({ ...eventAt('ability', 0.2), kind, repeat });
		const design = EncounterDesign.create({
			bossAbilities: [ability(DesignBossAbility_Kind.SpellHit), ability(DesignBossAbility_Kind.Strike), ability(DesignBossAbility_Kind.Thrash)],
			addWaves: [eventAt('wave', 0.2)],
		});
		const [hit, strike, thrash] = laneEvents(design, 'abilities', BASE);
		expect([hit.window.end - hit.window.start, hit.repeats.length, hit.ticks.length]).toEqual([expect.closeTo(0.1), 0, 2]);
		expect(strike.ticks).toHaveLength(2);
		expect([thrash.ticks.length, thrash.repeats.length > 0]).toEqual([0, true]);
		expect(laneEvents(design, 'waves', BASE).map(event => event.openEnd)).toEqual([true]);
	});
});

describe('withWindow', () => {
	it('writes the window back into the event the kind stores it in, on a copy', () => {
		const design = EncounterDesign.create({ addWaves: [eventAt('wave', 0.2)], moves: [eventAt('move', 0.2)] });
		const wave = withWindow(design, { kind: 'wave', index: 0 }, { start: 0.3, end: 0.45 }, BASE).addWaves[0];
		expect(wave.spawn?.at).toBe(0.3);
		expect(wave.lifetime).toBeCloseTo(0.15);
		expect(withWindow(design, { kind: 'move', index: 0 }, { start: 0.6, end: 0.6 + 5 / 7 / BASE }, BASE).moves[0].at?.at).toBe(0.6);
		expect(design.addWaves[0].spawn?.at).toBe(0.2);
	});

	it('sets an ability from the window, a spell hit as well as a thrash', () => {
		const design = EncounterDesign.create({
			bossAbilities: [eventAt('ability', 0.2), { ...eventAt('ability', 0.2), kind: DesignBossAbility_Kind.SwingTiming }],
		});
		for (const index of [0, 1]) {
			const moved = withWindow(design, { kind: 'ability', index }, { start: 0.5, end: 0.7 }, BASE).bossAbilities[index];
			expect([moved.start?.at, moved.end?.at]).toEqual([0.5, 0.7]);
		}
	});

	it('stretches the walk of a move through the unit it holds and leaves it alone on a plain move', () => {
		const design = EncounterDesign.create({ moves: [eventAt('move', 0.1), { ...eventAt('move', 0.1), yards: 0, seconds: 4 }] });
		const length = (seconds: number) => ({ start: 0.1, end: 0.1 + seconds / BASE });
		const yards = withWindow(design, { kind: 'move', index: 0 }, length(2), BASE).moves[0];
		expect([yards.yards, yards.seconds]).toEqual([expect.closeTo(14), 0]);
		const timed = withWindow(design, { kind: 'move', index: 1 }, length(6), BASE).moves[1];
		expect([timed.yards, timed.seconds]).toEqual([0, expect.closeTo(6)]);
		const moved = withWindow(design, { kind: 'move', index: 0 }, { start: 0.3, end: 0.3 + 5 / 7 / BASE }, BASE).moves[0];
		expect([moved.at?.at, moved.yards]).toEqual([0.3, 5]);
	});

	describe('boss health segments', () => {
		const design = EncounterDesign.create({
			bossHealth: [
				{ ...defaultEvent('bossHealth'), at: 0.25, healthPercent: 90 },
				{ ...defaultEvent('bossHealth'), at: 0.5, healthPercent: 65 },
				{ ...defaultEvent('bossHealth'), at: 0.75, healthPercent: 30 },
			],
		});
		const ats = (next: EncounterDesign) => next.bossHealth.map(keyframe => keyframe.at);

		it('moves the keyframe at the segment end with the end handle, between its neighbours', () => {
			expect(ats(withWindow(design, { kind: 'bossHealth', index: 1 }, { start: 0.25, end: 0.6 }, BASE))).toEqual([0.25, 0.6, 0.75]);
			expect(ats(withWindow(design, { kind: 'bossHealth', index: 1 }, { start: 0.25, end: 0.9 }, BASE))).toEqual([0.25, 0.75, 0.75]);
		});

		it('moves the keyframe before the segment with the start handle and keeps the implied start fixed', () => {
			expect(ats(withWindow(design, { kind: 'bossHealth', index: 1 }, { start: 0.4, end: 0.5 }, BASE))).toEqual([0.4, 0.5, 0.75]);
			expect(ats(withWindow(design, { kind: 'bossHealth', index: 0 }, { start: 0.1, end: 0.25 }, BASE))).toEqual([0.25, 0.5, 0.75]);
		});

		it('moves only the keyframe at the end when the whole segment moves', () => {
			expect(ats(withWindow(design, { kind: 'bossHealth', index: 1 }, { start: 0.3, end: 0.55 }, BASE))).toEqual([0.25, 0.55, 0.75]);
		});
	});
});
