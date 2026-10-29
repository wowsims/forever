import { BulkSimStage, type ProgressMetrics } from '@generated/proto/api';
import type { BulkSimProgressConfig } from '@sim/bulk/types';
import { describe, expect, it } from 'vitest';

import { batchSecondsRemaining, candidateGearProgress, roundSecondsRemaining, simProgress, smoothSecondsRemaining } from './progress';

const metrics = (fields: Partial<ProgressMetrics>) =>
	({ completedIterations: 0, totalIterations: 0, completedSims: 0, totalSims: 0, bulkStage: BulkSimStage.BulkSimStageUnknown, ...fields }) as ProgressMetrics;

const config = (fields: Partial<BulkSimProgressConfig> = {}): BulkSimProgressConfig => ({ currentRound: 1, totalRounds: 2, ...fields });

describe('candidateGearProgress', () => {
	it('reports the stage and title alone until it can be measured', () => {
		expect(candidateGearProgress({ now: 1000 })).toEqual({ stage: 'preparing', title: expect.any(String) });
	});

	it('estimates the time left from the rate so far', () => {
		const progress = candidateGearProgress({ completed: 25, total: 100, startedAt: 0, now: 10_000 });
		// 10s for 25 of 100, so 30s for the remaining 75.
		expect(progress.secondsRemaining).toBe(30);
		expect(progress).toMatchObject({ current: 25, total: 100 });
	});

	it('leaves the estimate off before the first candidate lands, so it cannot divide by zero', () => {
		expect(candidateGearProgress({ completed: 0, total: 100, startedAt: 0, now: 10_000 }).secondsRemaining).toBeUndefined();
	});
});

describe('simProgress', () => {
	it('measures its rounds by the aggregate iteration counts when the caller supplies them', () => {
		const progress = simProgress(
			metrics({ totalIterations: 10, completedIterations: 5 }),
			config({ aggregateCompletedIterations: 400, aggregateTotalIterations: 1000, aggregateStartedAt: 0 }),
			0,
			10_000,
		);
		expect(progress?.total).toBe(2);
		expect(progress?.current).toBeCloseTo(0.8);
	});

	it('switches to sim counts when the caller asks for them', () => {
		const progress = simProgress(
			metrics({ totalIterations: 100, completedIterations: 50, totalSims: 8, completedSims: 3 }),
			config({ useSimCountProgress: true }),
			0,
			1000,
		);
		expect(progress?.total).toBe(8);
		expect(progress?.current).toBe(4);
	});

	it('returns null on an unusable estimate, so the last frame stays up', () => {
		const unusable = config({ aggregateCompletedIterations: 50, aggregateTotalIterations: 100, aggregateStartedAt: NaN });
		expect(simProgress(metrics({ totalIterations: 100, completedIterations: 50 }), unusable, 0, 1000)).toBeNull();
	});
});

// Seconds per step of a real batch with the same plan for every step: four encounters (1:00, 1:00, 2:00, and 3:00 with
// two targets), each running bundle groups A, B and C.
const MEASURED = [5.7, 5.5, 20.8, 5.5, 5.7, 19.8, 5.5, 5.8, 19.3, 4.2, 4.3, 8.0];

// Replays `times` with each step's progress moving linearly in time: the estimate and the true time left at four points
// of every step.
const replay = (times: readonly number[], groupsPerEncounter: number) => {
	const groups = times.map((_, step) => step % groupsPerEncounter);
	const weights = times.map(() => 40_000);
	const total = times.reduce((sum, time) => sum + time, 0);
	const points: Array<{ running: number; estimate: number | undefined; truth: number }> = [];
	let elapsed = 0;
	for (let running = 0; running < times.length; running++) {
		for (const progress of [0, 0.25, 0.5, 0.75]) {
			const estimate = batchSecondsRemaining({ groups, weights, seconds: times.slice(0, running), running, progress });
			points.push({ running, estimate, truth: total - elapsed - times[running] * progress });
		}
		elapsed += times[running];
	}
	return points;
};

describe('batch time left', () => {
	it('shows nothing until every bundle group has finished once, which is after the first encounter', () => {
		const points = replay(MEASURED, 3);
		expect(points.filter(point => point.running < 3).every(point => point.estimate === undefined)).toBe(true);
		expect(points.filter(point => point.running >= 3).every(point => point.estimate !== undefined)).toBe(true);
	});

	it('predicts each step from its bundle group’s last step, within 25% while the encounters cost alike', () => {
		for (const { estimate, truth } of replay(MEASURED.slice(0, 9), 3).filter(point => point.running >= 3)) {
			expect(Math.abs(estimate! - truth) / truth).toBeLessThanOrEqual(0.25);
		}
	});

	// The last encounter costs half of what the earlier ones did, which nothing before it shows, so the estimate
	// runs over by up to 18 seconds there and its relative error grows as the time left shrinks.
	it('stays within 20 seconds on the measured batch', () => {
		for (const { estimate, truth } of replay(MEASURED, 3).filter(point => point.running >= 3)) {
			expect(Math.abs(estimate! - truth)).toBeLessThanOrEqual(20);
		}
	});

	it('scales a bundle group’s measured time by the planned iterations of the step it predicts', () => {
		expect(batchSecondsRemaining({ groups: [0, 0], weights: [10_000, 30_000], seconds: [4], running: 1, progress: 0.5 })).toBe(6);
	});

	it('shows nothing for a single encounter or a Gear-mode batch, where no bundle group finishes before the last step', () => {
		expect(replay(MEASURED.slice(0, 3), 3).every(point => point.estimate === undefined)).toBe(true);
		expect(replay([12], 1).every(point => point.estimate === undefined)).toBe(true);
	});
});

describe('shown time left', () => {
	it('starts at the first estimate', () => {
		expect(smoothSecondsRemaining(undefined, 42, 1)).toBe(42);
	});

	it('counts down a second a second while the estimate does not move', () => {
		const shown = [40.4];
		for (let second = 0; second < 3; second++) shown.push(smoothSecondsRemaining(shown.at(-1), 40.4, 1));
		expect(shown.map(roundSecondsRemaining)).toEqual([40, 39, 38, 37]);
	});

	it('glides toward a lower estimate rather than jumping to it', () => {
		const first = smoothSecondsRemaining(60, 30, 1);
		const second = smoothSecondsRemaining(first, 29, 1);
		expect(first).toBeLessThan(59);
		expect(first).toBeGreaterThan(30);
		expect(second).toBeLessThan(first);
		expect(second).toBeGreaterThan(29);
	});

	it('rises only for an estimate more than 10 s or a quarter above it, and glides up then', () => {
		expect(smoothSecondsRemaining(25, 33, 1)).toBe(24);
		const raised = smoothSecondsRemaining(30, 45, 1);
		expect(raised).toBeGreaterThan(30);
		expect(raised).toBeLessThan(45);
	});

	it('counts down at most 3 s a second when the estimate collapses, so it takes several seconds to reach the floor', () => {
		const shown = [15];
		for (let second = 0; second < 3; second++) shown.push(smoothSecondsRemaining(shown.at(-1), 0, 1));
		expect(shown).toEqual([15, 12, 9, 6]);
	});

	it('shows whole seconds, and never less than 1 s while running', () => {
		expect([42.4, 42.6, 0.2, 0].map(roundSecondsRemaining)).toEqual([42, 43, 1, 1]);
	});
});
