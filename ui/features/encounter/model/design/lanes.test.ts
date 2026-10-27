import { EncounterDesign } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import type { DesignEventKind } from './lanes';
import { addEvent, defaultEvent, eventLabelKey, eventsOf, eventWindow, getEvent, laneLabelKey, laneOf, occurrences, removeEvent } from './lanes';

const KINDS: DesignEventKind[] = ['downtime', 'bossHealth', 'ability', 'wave', 'move', 'amp'];

describe('lanes', () => {
	it('places every event kind in a lane', () => {
		expect(KINDS.map(laneOf)).toEqual(['boss', 'boss', 'abilities', 'waves', 'moves', 'amps']);
	});

	it('builds label keys under the designer namespace', () => {
		expect(laneLabelKey('waves')).toBe('settings_tab.encounter.designer.lanes.waves');
		expect(eventLabelKey('bossHealth')).toBe('settings_tab.encounter.designer.events.bossHealth');
	});

	it('defaults every kind inside the fight', () => {
		for (const kind of KINDS) {
			const design = EncounterDesign.create();
			const ref = addEvent(design, kind, 0.98);
			const window = eventWindow(design, ref)!;
			const [from, to] = 'at' in window ? [window.at, window.at] : [window.start, window.end];
			expect(from).toBeGreaterThanOrEqual(0);
			expect(to).toBeLessThanOrEqual(1);
		}
	});

	it('uses the documented defaults', () => {
		const wave = defaultEvent('wave');
		expect([wave.count, wave.lifetime, wave.executeProportion20, wave.executeProportion35, wave.killShare]).toEqual([1, 0.1, 0.2, 0.35, 0.25]);
		expect(defaultEvent('move').yards).toBe(5);
		expect(defaultEvent('amp').percent).toBe(10);
		expect(defaultEvent('ability').castTime).toBe(2);
		const downtime = defaultEvent('downtime');
		expect(downtime.clearDebuffs).toBe(true);
		expect(downtime.end!.at - downtime.start!.at).toBeCloseTo(0.1);
	});

	it('adds, reads and removes events by list and index', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'move', 0.1);
		const ref = addEvent(design, 'move', 0.5);
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

	it('draws moves and keyframes as points and waves as spawn to spawn plus lifetime', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'move', 0.3);
		addEvent(design, 'bossHealth', 0.6);
		addEvent(design, 'wave', 0.2);
		expect(eventWindow(design, { kind: 'move', index: 0 })).toEqual({ at: 0.3 });
		expect(eventWindow(design, { kind: 'bossHealth', index: 0 })).toEqual({ at: 0.6 });
		expect(eventWindow(design, { kind: 'wave', index: 0 })).toEqual({ start: 0.2, end: 0.30000000000000004 });
		expect(eventWindow(design, { kind: 'wave', index: 4 })).toBeUndefined();
	});

	describe('occurrences', () => {
		it('returns one window without a repeat', () => {
			const design = EncounterDesign.create();
			const ref = addEvent(design, 'amp', 0.2);
			expect(occurrences(design, ref, 10)).toEqual([{ start: 0.2, end: 0.30000000000000004 }]);
		});

		it('repeats count times after the first, ignoring jitter', () => {
			const design = EncounterDesign.create();
			const ref = addEvent(design, 'move', 0.1);
			design.moves[0].repeat = { interval: 0.2, jitter: 0.05, count: 2 };
			expect(occurrences(design, ref, 10).map(o => o.start)).toEqual([0.1, 0.30000000000000004, 0.5]);
		});

		it('runs to the end of the fight with count 0 and caps at maxCount', () => {
			const design = EncounterDesign.create();
			const ref = addEvent(design, 'move', 0.1);
			design.moves[0].repeat = { interval: 0.25, jitter: 0, count: 0 };
			expect(occurrences(design, ref, 10).map(o => o.start)).toEqual([0.1, 0.35, 0.6, 0.85]);
			expect(occurrences(design, ref, 2)).toHaveLength(2);
			expect(occurrences(design, ref, 0)).toEqual([]);
		});
	});
});
