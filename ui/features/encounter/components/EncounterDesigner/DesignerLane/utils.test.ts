import { EncounterDesign } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { defaultEvent } from '../../../model/design/lanes';
import { laneEventKinds, laneEvents, withWindow } from './utils';

const downtime = (start: number, end: number) => ({ ...defaultEvent('downtime'), start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
const rowsByStart = (design: EncounterDesign, lane: Parameters<typeof laneEvents>[1]) => laneEvents(design, lane).map(event => [event.window.start, event.row]);

describe('laneEvents', () => {
	it('puts windows that do not overlap on one row', () => {
		const design = EncounterDesign.create({ downtimes: [downtime(0.5, 0.6), downtime(0, 0.2), downtime(0.3, 0.4)] });
		expect(rowsByStart(design, 'boss')).toEqual([
			[0, 0],
			[0.3, 0],
			[0.5, 0],
		]);
	});

	it('moves an overlapping window, or a point at the same instant, to the next free row', () => {
		const design = EncounterDesign.create({ downtimes: [downtime(0, 0.5), downtime(0.2, 0.3), downtime(0.4, 0.6), downtime(0.6, 0.7)] });
		expect(rowsByStart(design, 'boss')).toEqual([
			[0, 0],
			[0.2, 1],
			[0.4, 1],
			[0.6, 0],
		]);
		const points = EncounterDesign.create({ moves: [defaultEvent('move', 0.5), defaultEvent('move', 0.5)] });
		expect(rowsByStart(points, 'moves').map(([, row]) => row)).toEqual([0, 1]);
	});

	it('lists the boss lane downtimes and keyframes in time order', () => {
		const design = EncounterDesign.create({
			downtimes: [defaultEvent('downtime', 0.4)],
			bossHealth: [defaultEvent('bossHealth', 0.1), defaultEvent('bossHealth', 0.45)],
		});
		const events = laneEvents(design, 'boss');
		expect(laneEventKinds('boss')).toEqual(['downtime', 'bossHealth']);
		expect(events.map(event => [event.ref.kind, event.ref.index, event.shape, event.row])).toEqual([
			['bossHealth', 0, 'dot', 0],
			['downtime', 0, 'span', 0],
			['bossHealth', 1, 'dot', 1],
		]);
	});

	it('keeps the repeats after the first occurrence and packs around their full reach', () => {
		const repeating = { ...defaultEvent('move', 0.1), repeat: { interval: 0.3, jitter: 0, count: 2 } };
		const design = EncounterDesign.create({ moves: [repeating, defaultEvent('move', 0.5)] });
		const [first, second] = laneEvents(design, 'moves');
		expect(first.repeats.map(repeat => repeat.start)).toEqual([expect.closeTo(0.4), expect.closeTo(0.7)]);
		expect(second.row).toBe(1);
	});
});

describe('withWindow', () => {
	it('writes the window back into the event the kind stores it in, on a copy', () => {
		const design = EncounterDesign.create({ addWaves: [defaultEvent('wave', 0.2)], moves: [defaultEvent('move', 0.2)] });
		const wave = withWindow(design, { kind: 'wave', index: 0 }, { start: 0.3, end: 0.45 }).addWaves[0];
		expect(wave.spawn?.at).toBe(0.3);
		expect(wave.lifetime).toBeCloseTo(0.15);
		expect(withWindow(design, { kind: 'move', index: 0 }, { start: 0.6, end: 0.6 }).moves[0].at?.at).toBe(0.6);
		expect(design.addWaves[0].spawn?.at).toBe(0.2);
	});
});
