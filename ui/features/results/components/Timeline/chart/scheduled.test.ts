import { AddWave, DamageAmpType, EncounterPlan, PlanBossAbility, PlanBossAbility_Kind, PlanDamageAmp, PlanRepeat, PlanTime } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { scheduledMarkers } from './scheduled';

const at = (value: number): PlanTime => ({ at: value, jitter: 0 });
const repeat = (interval: number, count: number): PlanRepeat => ({ interval, jitter: 0, count });
const plan = (partial: Partial<EncounterPlan>): EncounterPlan => ({ ...EncounterPlan.create(), ...partial });

describe('scheduledMarkers', () => {
	it('draws nothing without a plan', () => {
		expect(scheduledMarkers(undefined, 100)).toEqual([]);
	});

	it('scales a downtime by the fight duration', () => {
		const markers = scheduledMarkers(plan({ downtimes: [{ name: '', targetIndex: 0, start: at(0.2), end: at(0.3), clearDebuffs: true }] }), 200);
		expect(markers).toEqual([{ kind: 'downtime', start: 40, end: 60 }]);
	});

	it('draws a wave for each occurrence, from spawn to spawn plus lifetime', () => {
		const wave = { ...AddWave.create(), name: 'w', count: 2, spawn: at(0.1), lifetime: 0.1, repeat: repeat(0.3, 2) };
		const markers = scheduledMarkers(plan({ addWaves: [wave] }), 100);
		expect(markers.map(m => [m.start, m.end])).toEqual([
			[10, 20],
			[40, 50],
			[70, 80],
		]);
	});

	it('repeats until the end of the fight when the count is 0 and cuts the last window there', () => {
		const amp = PlanDamageAmp.create({ type: DamageAmpType.DamageAmpTypeCasterBuff, percent: 10, start: at(0.4), end: at(0.6), repeat: repeat(0.4, 0) });
		const markers = scheduledMarkers(plan({ damageAmps: [amp] }), 100);
		expect(markers.map(m => [m.kind, m.start, m.end])).toEqual([
			['amp', 40, 60],
			['amp', 80, 100],
		]);
	});

	it('marks a move as a point', () => {
		const markers = scheduledMarkers(plan({ moves: [{ name: '', at: at(0.5), yards: 5, seconds: 0, repeat: repeat(0, 0) }] }), 60);
		expect(markers).toEqual([{ kind: 'move', start: 30, end: 30 }]);
	});

	it("draws a pulse's windows on their own row and leaves the other boss abilities out", () => {
		const ability = (kind: PlanBossAbility_Kind) => PlanBossAbility.create({ kind, start: at(0.1), end: at(0.2), repeat: repeat(0.5, 1) });
		const markers = scheduledMarkers(
			plan({ bossAbilities: [ability(PlanBossAbility_Kind.Thrash), ability(PlanBossAbility_Kind.Pulse), ability(PlanBossAbility_Kind.AreaHit)] }),
			100,
		);
		expect(markers).toEqual([
			{ kind: 'pulse', start: 10, end: 20 },
			{ kind: 'pulse', start: 60, end: 70 },
		]);
	});

	it('skips an inverted window', () => {
		const markers = scheduledMarkers(plan({ downtimes: [{ name: '', targetIndex: 0, start: at(0.5), end: at(0.3), clearDebuffs: true }] }), 100);
		expect(markers).toEqual([]);
	});
});
