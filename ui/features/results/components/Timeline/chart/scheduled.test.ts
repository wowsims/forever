import { AddWave, DamageAmpType, DesignDamageAmp, DesignRepeat, DesignTime, EncounterDesign } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { scheduledMarkers } from './scheduled';

const at = (value: number): DesignTime => ({ at: value, jitter: 0 });
const repeat = (interval: number, count: number): DesignRepeat => ({ interval, jitter: 0, count });
const design = (partial: Partial<EncounterDesign>): EncounterDesign => ({ ...EncounterDesign.create(), ...partial });

describe('scheduledMarkers', () => {
	it('draws nothing without a design', () => {
		expect(scheduledMarkers(undefined, 100)).toEqual([]);
	});

	it('scales a downtime by the fight duration', () => {
		const markers = scheduledMarkers(design({ downtimes: [{ targetIndex: 0, start: at(0.2), end: at(0.3), clearDebuffs: true }] }), 200);
		expect(markers).toEqual([{ kind: 'downtime', start: 40, end: 60 }]);
	});

	it('draws a wave for each occurrence, from spawn to spawn plus lifetime', () => {
		const wave = { ...AddWave.create(), name: 'w', count: 2, spawn: at(0.1), lifetime: 0.1, repeat: repeat(0.3, 2) };
		const markers = scheduledMarkers(design({ addWaves: [wave] }), 100);
		expect(markers.map(m => [m.start, m.end])).toEqual([
			[10, 20],
			[40, 50],
			[70, 80],
		]);
	});

	it('repeats until the end of the fight when the count is 0 and cuts the last window there', () => {
		const amp = DesignDamageAmp.create({ type: DamageAmpType.DamageAmpTypeCasterBuff, percent: 10, start: at(0.4), end: at(0.6), repeat: repeat(0.4, 0) });
		const markers = scheduledMarkers(design({ damageAmps: [amp] }), 100);
		expect(markers.map(m => [m.kind, m.start, m.end])).toEqual([
			['amp', 40, 60],
			['amp', 80, 100],
		]);
	});

	it('marks a move as a point', () => {
		const markers = scheduledMarkers(design({ moves: [{ at: at(0.5), yards: 5, seconds: 0, repeat: repeat(0, 0) }] }), 60);
		expect(markers).toEqual([{ kind: 'move', start: 30, end: 30 }]);
	});

	it('skips an inverted window', () => {
		const markers = scheduledMarkers(design({ downtimes: [{ targetIndex: 0, start: at(0.5), end: at(0.3), clearDebuffs: true }] }), 100);
		expect(markers).toEqual([]);
	});
});
