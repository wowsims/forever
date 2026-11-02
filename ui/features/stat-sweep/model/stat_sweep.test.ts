import { EP_UNIT_STATS } from '@features/stat-weights/model/ep_unit_stats';
import { PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { UnitStat } from '@sim/proto/stats';
import type { PrecisionSettings } from '@sim/settings/precision';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { Z_90 } from '@sim/utils/math';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../assets/locales/en/translation.json';
import {
	advanceSweepProgress,
	buildSweepOffsets,
	displayColorSlots,
	measureSweepStep,
	mergeSweepHistory,
	type MetricSample,
	nextColorSlot,
	offeredSweepMetrics,
	runSweepStat,
	shownSweepMetric,
	skipSweepSteps,
	startSweepProgress,
	statSweepCsv,
	type StatSweepSample,
	type StatSweepStepSample,
	SWEEP_POINTS_PER_PERCENT,
	sweepBatches,
	sweepBonus,
	sweepHasTotal,
	type SweepPassRunner,
	sweepRangeText,
	sweepStepPosition,
	sweepSteps,
} from './stat_sweep';

// The first pass of a Target ± step.
const PROBE = 1000;
// The total that brings a 90% ± of a spread of `stdev` down to `targetCi`.
const totalFor = (stdev: number, targetCi: number) => Math.ceil(((Z_90 * stdev) / targetCi) ** 2);

// A run of `values`, as the sim reports it: population stdev and the sum of squares.
const sampleOf = (values: number[]): MetricSample => {
	const n = values.length;
	const avg = values.reduce((total, value) => total + value, 0) / n;
	const sumSq = values.reduce((total, value) => total + value * value, 0);
	return { avg, stdev: Math.sqrt(sumSq / n - avg * avg), n, sumSq };
};

// `iterations` fights of each metric at its average and spread.
const flat = (avg: number, stdev: number, iterations: number): MetricSample => ({ avg, stdev, n: iterations, sumSq: iterations * (stdev ** 2 + avg ** 2) });

// Every metric at its own average, with the DPS spread or one given per metric.
const sweepSample = (iterations: number, stdevs: Partial<Record<StatSweepMetric, number>> = {}, dps = 1000): StatSweepSample => ({
	dps: flat(dps, stdevs.dps ?? 100, iterations),
	hps: flat(50, stdevs.hps ?? 10, iterations),
	tps: flat(dps * 2, stdevs.tps ?? 200, iterations),
	dtps: flat(400, stdevs.dtps ?? 40, iterations),
	tmi: flat(12, stdevs.tmi ?? 3, iterations),
});

const precision = (patch: Partial<PrecisionSettings> = {}): PrecisionSettings => ({
	precisionMode: 'iterations',
	iterations: 0,
	targetCi: 1,
	maxIterationMultiplier: 0,
	...patch,
});

describe('offeredSweepMetrics', () => {
	it('offers DPS and TPS always, HPS with the healing metrics, and DTPS and TMI with the threat and tank options', () => {
		expect(offeredSweepMetrics({ healing: false, threat: false })).toEqual(['dps', 'tps']);
		expect(offeredSweepMetrics({ healing: true, threat: false })).toEqual(['dps', 'hps', 'tps']);
		expect(offeredSweepMetrics({ healing: false, threat: true })).toEqual(['dps', 'tps', 'dtps', 'tmi']);
		expect(offeredSweepMetrics({ healing: true, threat: true })).toEqual(['dps', 'hps', 'tps', 'dtps', 'tmi']);
	});

	it('shows the stored metric while it is offered, else DPS', () => {
		expect(shownSweepMetric('tmi', ['dps', 'tps', 'dtps', 'tmi'])).toBe('tmi');
		expect(shownSweepMetric('tmi', ['dps', 'tps'])).toBe('dps');
	});
});

describe('buildSweepOffsets', () => {
	it('steps from -range to +range through 0', () => {
		expect(buildSweepOffsets(30, 10)).toEqual([-30, -20, -10, 0, 10, 20, 30]);
	});

	it('stops at the last multiple of step inside the range', () => {
		expect(buildSweepOffsets(25, 10)).toEqual([-20, -10, 0, 10, 20]);
		expect(buildSweepOffsets(5, 10)).toEqual([0]);
	});

	it('keeps fractional steps exact', () => {
		expect(buildSweepOffsets(0.3, 0.1)).toEqual([-0.3, -0.2, -0.1, 0, 0.1, 0.2, 0.3]);
	});

	it('returns only 0 for a step or range that cannot step', () => {
		expect(buildSweepOffsets(50, 0)).toEqual([0]);
		expect(buildSweepOffsets(50, -10)).toEqual([0]);
		expect(buildSweepOffsets(-50, 10)).toEqual([0]);
	});
});

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);
const CRIT_RATING = UnitStat.fromStat(Stat.StatCritRating);
const HIT_RATING = UnitStat.fromStat(Stat.StatHitRating);
const DEFENSE_RATING = UnitStat.fromStat(Stat.StatDefenseRating);
const RANGED_CRIT_PERCENT = UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent);
const SHADOW_HIT_PERCENT = UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow);
const MAIN_HAND_DPS = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps);

describe('sweepSteps', () => {
	it('places each offset at its total and drops those whose total falls below 0, keeping one that lands on 0', () => {
		expect(sweepSteps(STRENGTH, [-3, -2, -1, 0, 1], [-13, 0, 11, 22, 33])).toEqual([
			{ offset: -2, bonus: -20, statValue: 0 },
			{ offset: -1, bonus: -10, statValue: 11 },
			{ offset: 0, bonus: 0, statValue: 22 },
			{ offset: 1, bonus: 10, statValue: 33 },
		]);
	});

	it('rounds float noise out of the totals', () => {
		expect(sweepSteps(STRENGTH, [1], [0.1 + 0.2])).toEqual([{ offset: 1, bonus: 10, statValue: 0.3 }]);
	});

	it('keeps the percent as the offset, with the bonus in the unit the stat stores', () => {
		expect(sweepSteps(UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellCritPercent), [2], [5])).toEqual([{ offset: 2, bonus: 2, statValue: 5 }]);
		expect(sweepSteps(CRIT_RATING, [2], [100])).toEqual([{ offset: 2, bonus: 28, statValue: 100 }]);
	});
});

describe('measureSweepStep', () => {
	const runner = (stdevs: Partial<Record<StatSweepMetric, number>>) => vi.fn<SweepPassRunner>(async iterations => sweepSample(iterations, stdevs));

	it('runs the iterations setting once from the seed, or the sim iterations when it is 0', async () => {
		const run = runner({});

		expect((await measureSweepStep(run, precision({ iterations: 3000 }), 'dps', 12500))?.metrics.dps.n).toBe(3000);
		expect((await measureSweepStep(run, precision(), 'dps', 12500))?.metrics.dps.n).toBe(12500);
		expect(run.mock.calls).toEqual([
			[3000, 0],
			[12500, 0],
		]);
	});

	it('probes, then tops up to the total the target needs, continuing at the seed after the probe', async () => {
		const run = runner({ dps: 10 });
		const planned: number[] = [];

		const sample = await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.5, iterations: 3000 }), 'dps', 12500, total =>
			planned.push(total),
		);

		const total = totalFor(10, 0.5);
		expect(run.mock.calls).toEqual([
			[PROBE, 0],
			[total - PROBE, PROBE],
		]);
		expect(planned).toEqual([PROBE, total]);
		expect(sample?.metrics.dps.n).toBe(total);
		expect(sample?.metrics.tmi.n).toBe(total);
	});

	it("sizes Target ± by the run's metric, whose spread sets the total", async () => {
		const stdevs = { dps: 1, tps: 20 };
		const byDps = runner(stdevs);
		const byTps = runner(stdevs);

		await measureSweepStep(byDps, precision({ precisionMode: 'target', targetCi: 1 }), 'dps', 12500);
		const sample = await measureSweepStep(byTps, precision({ precisionMode: 'target', targetCi: 1 }), 'tps', 12500);

		expect(byDps).toHaveBeenCalledTimes(1);
		expect(byTps.mock.calls[1][0]).toBe(totalFor(20, 1) - PROBE);
		expect(sample?.metrics.tps.n).toBe(totalFor(20, 1));
	});

	it('merges the passes into one run over every iteration, each metric on its own', async () => {
		const pass = (dps: number[], tmi: number[]): StatSweepSample => ({
			...sweepSample(dps.length),
			dps: sampleOf(dps),
			tmi: sampleOf(tmi),
		});
		const run = vi
			.fn<SweepPassRunner>()
			.mockResolvedValueOnce(pass([1000, 1100, 1250], [10, 14, 6]))
			.mockResolvedValue(pass([900, 1300], [8, 12]));

		const sample = await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.001 }), 'dps', 12500);

		expect(run).toHaveBeenCalledTimes(4);
		const dps = sampleOf([1000, 1100, 1250, 900, 1300, 900, 1300, 900, 1300]);
		const tmi = sampleOf([10, 14, 6, 8, 12, 8, 12, 8, 12]);
		expect(sample?.metrics.dps.n).toBe(dps.n);
		expect(sample?.metrics.dps.avg).toBeCloseTo(dps.avg, 9);
		expect(sample?.metrics.dps.stdev).toBeCloseTo(dps.stdev, 9);
		expect(sample?.metrics.dps.sumSq).toBeCloseTo(dps.sumSq, 6);
		expect(sample?.metrics.tmi.avg).toBeCloseTo(tmi.avg, 9);
		expect(sample?.metrics.tmi.stdev).toBeCloseTo(tmi.stdev, 9);
	});

	it('runs the first pass alone without a target above 0, or with a spread that is not a number', async () => {
		for (const targetCi of [0, -1, NaN]) {
			const run = runner({ dps: 100 });
			const sample = await measureSweepStep(run, precision({ precisionMode: 'target', targetCi }), 'dps', 12500);
			expect(run.mock.calls).toEqual([[PROBE, 0]]);
			expect(sample?.metrics.dps.n).toBe(PROBE);
		}
		const noSpread = runner({ dps: NaN });
		await measureSweepStep(noSpread, precision({ precisionMode: 'target', targetCi: 0.5 }), 'dps', 12500);
		expect(noSpread).toHaveBeenCalledTimes(1);
	});

	it('asks for no more iterations than the sim takes in one request, with no cap set', async () => {
		const run = runner({ dps: 1e6 });

		await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 1e-6 }), 'dps', 12500);

		expect(run.mock.calls[1]).toEqual([2 ** 31 - 1 - PROBE, PROBE]);
		expect(run.mock.calls.every(([iterations]) => Number.isInteger(iterations) && iterations > 0)).toBe(true);
	});

	it('runs the first pass alone when it confirms the stat capped', async () => {
		const run = runner({ dps: 100 });
		const confirmsCap = vi.fn(() => true);

		const sample = await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.1 }), 'dps', 12500, undefined, confirmsCap);

		expect(confirmsCap).toHaveBeenCalledExactlyOnceWith({ dps: 1000, hps: 50, tps: 2000, dtps: 400, tmi: 12 });
		expect(run).toHaveBeenCalledTimes(1);
		expect(sample?.metrics.dps.n).toBe(PROBE);
	});

	it('stops after the probe when it already meets the target', async () => {
		const run = runner({ dps: 1 });

		await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 1 }), 'dps', 12500);

		expect(run).toHaveBeenCalledTimes(1);
	});

	it('caps a step at Max × the probe', async () => {
		const run = runner({ dps: 100 });

		const sample = await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.1, maxIterationMultiplier: 4 }), 'dps', 12500);

		expect(sample?.metrics.dps.n).toBe(4 * PROBE);
	});

	it('resizes each top-up from every iteration so far, and stops after the third, as the EP calculation does', async () => {
		// The spread grows with each pass, so every plan asks for more.
		let stdev = 10;
		const run = vi.fn<SweepPassRunner>(async iterations => {
			stdev *= 2;
			return sweepSample(iterations, { dps: stdev });
		});

		await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.5 }), 'dps', 12500);

		expect(run).toHaveBeenCalledTimes(4);
		const seedOffsets = run.mock.calls.map(call => call[1]);
		const iterations = run.mock.calls.map(call => call[0]);
		expect(seedOffsets).toEqual([0, iterations[0], iterations[0] + iterations[1], iterations[0] + iterations[1] + iterations[2]]);
	});

	it("keeps every metric's first-pass average beside its average over every pass", async () => {
		const pass = (tps: number[]) => ({ ...sweepSample(PROBE), tps: { ...sampleOf(tps), n: PROBE } });
		const run = vi
			.fn<SweepPassRunner>()
			.mockResolvedValueOnce(pass([1900, 2100]))
			.mockResolvedValueOnce(pass([2200, 2400]));

		const sample = await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.1, maxIterationMultiplier: 2 }), 'tps', 12500);

		expect(sample?.firstPass).toEqual({ dps: 1000, hps: 50, tps: 2000, dtps: 400, tmi: 12 });
		expect(sample?.metrics.tps.avg).not.toBe(2000);
		const once = await measureSweepStep(vi.fn<SweepPassRunner>().mockResolvedValue(sweepSample(10)), precision(), 'dtps', 12500);
		expect(once?.firstPass).toEqual({ dps: 1000, hps: 50, tps: 2000, dtps: 400, tmi: 12 });
	});

	it('reports nothing when a pass stops', async () => {
		const run = vi.fn<SweepPassRunner>().mockResolvedValueOnce(sweepSample(PROBE)).mockResolvedValueOnce(null);

		expect(await measureSweepStep(run, precision({ precisionMode: 'target', targetCi: 0.1 }), 'dps', 12500)).toBeNull();
	});
});

describe('runSweepStat', () => {
	const stepsOf = (...offsets: number[]) => offsets.map(offset => ({ offset, bonus: offset, statValue: 100 + offset }));
	// DPS moves TPS with it; TMI is set on its own.
	const metricsAt = (dps: number, tmi: number): StatSweepSample => ({ ...sweepSample(1000, {}, dps), tmi: flat(tmi, 3, 1000) });
	const stepSample = (avg: number, firstPassDps = avg, tmi = 12): StatSweepStepSample => {
		const firstPass = metricsAt(firstPassDps, tmi);
		return {
			metrics: metricsAt(avg, tmi),
			firstPass: { dps: firstPass.dps.avg, hps: firstPass.hps.avg, tps: firstPass.tps.avg, dtps: firstPass.dtps.avg, tmi: firstPass.tmi.avg },
		};
	};
	const measureBy = (value: (offset: number) => number) =>
		vi.fn(async ({ offset }: { offset: number }, _ran: number, _confirmsCap: (firstPass: StatSweepStepSample['firstPass']) => boolean) =>
			stepSample(value(offset)),
		);

	it('runs up from 0 to the top, then down from the first step below 0, and stops going up after two steps in a row repeat every metric of the step below', async () => {
		const measure = measureBy(offset => 1000 + Math.min(offset, 20));
		const onSkip = vi.fn();

		const run = await runSweepStat(stepsOf(-20, -10, 0, 10, 20, 30, 40, 50, 60), measure, undefined, onSkip);

		expect(measure.mock.calls.map(([step, ran]) => [step.offset, ran])).toEqual([
			[0, 0],
			[10, 1],
			[20, 2],
			[30, 3],
			[40, 4],
			[-10, 5],
			[-20, 6],
		]);
		expect(onSkip).toHaveBeenCalledExactlyOnceWith(2);
		expect(run.cappedStep).toEqual({ offset: 20, bonus: 20, statValue: 120 });
		expect(run.points.map(point => point.offset)).toEqual([-20, -10, 0, 10, 20, 30, 40]);
		expect(run.stopped).toBe(false);
	});

	it('keeps going up past one step that repeats the step below, which may be a step too small to change any roll', async () => {
		const measure = measureBy(offset => 1000 + (offset >= 20 ? offset : 0));
		const onSkip = vi.fn();

		const run = await runSweepStat(stepsOf(0, 10, 20, 30, 40), measure, undefined, onSkip);

		expect(measure).toHaveBeenCalledTimes(5);
		expect(onSkip).not.toHaveBeenCalled();
		expect(run.cappedStep).toBeUndefined();
	});

	it('tells measure that a first pass confirms the cap only on the step that would be the second repeat in a row', async () => {
		const repeat = stepSample(1010).firstPass;
		const confirms: Array<[number, boolean, boolean]> = [];
		const measure = vi.fn(async ({ offset }: { offset: number }, _ran: number, confirmsCap: (firstPass: StatSweepStepSample['firstPass']) => boolean) => {
			confirms.push([offset, confirmsCap(repeat), confirmsCap(stepSample(1011).firstPass)]);
			return stepSample(1000 + Math.min(offset, 10));
		});

		await runSweepStat(stepsOf(-10, 0, 10, 20, 30), measure);

		expect(confirms).toEqual([
			[0, false, false],
			[10, false, false],
			[20, false, false],
			[30, true, false],
			[-10, false, false],
		]);
	});

	it('runs every step when a metric changes by any amount', async () => {
		const measure = measureBy(offset => 1000 + Math.min(offset, 20) + offset * 1e-9);
		const onSkip = vi.fn();

		const run = await runSweepStat(stepsOf(0, 10, 20, 30, 40), measure, undefined, onSkip);

		expect(measure).toHaveBeenCalledTimes(5);
		expect(onSkip).not.toHaveBeenCalled();
		expect(run.cappedStep).toBeUndefined();
	});

	it('keeps going up while one metric still changes, though every other one matches the step below it exactly', async () => {
		const measure = vi.fn(async ({ offset }: { offset: number }) => stepSample(1000, 1000, 12 - offset / 10));
		const onSkip = vi.fn();

		const run = await runSweepStat(stepsOf(0, 10, 20, 30), measure, undefined, onSkip);

		expect(measure).toHaveBeenCalledTimes(4);
		expect(onSkip).not.toHaveBeenCalled();
		expect(run.cappedStep).toBeUndefined();
	});

	it('runs every step below 0, even where every metric stops changing', async () => {
		const measure = measureBy(offset => 1000 + Math.max(offset, -10));

		const run = await runSweepStat(stepsOf(-40, -30, -20, -10, 0, 10), measure);

		expect(run.points).toHaveLength(6);
		expect(run.cappedStep).toBeUndefined();
	});

	it("compares each step's first pass, not its average over every pass", async () => {
		const measure = vi.fn(async ({ offset }: { offset: number }) => stepSample(1000 + offset, 1000));

		const run = await runSweepStat(stepsOf(0, 10, 20, 30), measure);

		expect(measure).toHaveBeenCalledTimes(3);
		expect(run.cappedStep?.offset).toBe(0);
		expect(run.points.map(point => point.metrics.dps.avg)).toEqual([1000, 1010, 1020]);
	});

	it('keeps every metric of each step without the sums of squares', async () => {
		const run = await runSweepStat(
			stepsOf(0),
			measureBy(() => 1000),
		);

		expect(run.points[0].metrics).toEqual({
			dps: { avg: 1000, stdev: 100, n: 1000 },
			hps: { avg: 50, stdev: 10, n: 1000 },
			tps: { avg: 2000, stdev: 200, n: 1000 },
			dtps: { avg: 400, stdev: 40, n: 1000 },
			tmi: { avg: 12, stdev: 3, n: 1000 },
		});
	});

	it('reports the run after each step, and stops with the finished steps when a step comes back empty', async () => {
		const measure = vi
			.fn<(step: { offset: number }) => Promise<StatSweepStepSample | null>>()
			.mockResolvedValueOnce(stepSample(1000))
			.mockResolvedValueOnce(null);
		const onUpdate = vi.fn();

		const run = await runSweepStat(stepsOf(-10, 0, 10), measure, onUpdate);

		expect(run.stopped).toBe(true);
		expect(run.points.map(point => point.offset)).toEqual([0]);
		expect(onUpdate).toHaveBeenCalledTimes(1);
	});
});

describe('advanceSweepProgress', () => {
	it('counts the steps done plus the running step share of its expected iterations, and estimates the rest at the rate so far', () => {
		const progress = advanceSweepProgress(startSweepProgress([4]), 1, 500, 1000, 30);

		expect(progress).toEqual({ stepCounts: [4], running: 1, completed: 1.5, iterationsPerStep: 1000, secondsRemaining: 50 });
	});

	it("places a step counted across every stat at its stat and that stat's own step", () => {
		const start = startSweepProgress([3, 2, 4]);

		expect(advanceSweepProgress(start, 5, 500, 1000, 55)).toMatchObject({ running: 5, completed: 5.5, secondsRemaining: 35 });
		expect(sweepStepPosition([3, 2, 4], 5)).toEqual({ stat: 2, step: 0 });
		expect(sweepStepPosition([3, 2, 4], 4)).toEqual({ stat: 1, step: 1 });
	});

	it('never falls back when a step plans more iterations than it expected', () => {
		const before = advanceSweepProgress(startSweepProgress([4]), 1, 900, 1000, 30);
		const grown = advanceSweepProgress(before, 1, 1000, 5000, 31);

		expect(grown.completed).toBe(before.completed);
		expect(grown.iterationsPerStep).toBe(5000);
	});

	it("takes skipped steps out of their stat and the total, so the bar only moves on and the next stat's steps place after the ones that ran", () => {
		const before = advanceSweepProgress(startSweepProgress([6, 4]), 3, 500, 1000, 35);

		const skipped = skipSweepSteps(before, 0, 2);

		expect(skipped).toMatchObject({ running: 3, stepCounts: [4, 4], completed: 3.5 });
		expect(sweepStepPosition(skipped.stepCounts, skipped.running)).toEqual({ stat: 0, step: 3 });
		expect(skipped.completed / 8).toBeGreaterThan(before.completed / 10);
		const next = advanceSweepProgress(skipped, 4, 0, 1000, 50);
		expect(next).toMatchObject({ running: 4, completed: 4, secondsRemaining: 50 });
		expect(sweepStepPosition(next.stepCounts, next.running)).toEqual({ stat: 1, step: 0 });
	});

	it('has no estimate before anything completes', () => {
		expect(advanceSweepProgress(startSweepProgress([4]), 0, 0, 1000, 5).secondsRemaining).toBeUndefined();
	});
});

describe('mergeSweepHistory', () => {
	const sweepRun = (id: string, createdAt: number, label = id) => ({ id, createdAt, label }) as StatSweepRecord;

	it('keeps a run finished before the stored list arrived, newest first, and lists a run held in both once, as stored', () => {
		const merged = mergeSweepHistory([sweepRun('new', 3), sweepRun('a', 2, 'memory')], [sweepRun('a', 2, 'stored'), sweepRun('older', 0)]);

		expect(merged.map(run => [run.id, run.label])).toEqual([
			['new', 'new'],
			['a', 'stored'],
			['older', 'older'],
		]);
	});
});

describe('sweepBatches', () => {
	it('groups runs by sweep, newest sweep first with its runs in run order, and makes a run with no batch a batch of its own', () => {
		const run = (id: string, batchId?: string) => ({ id, batchId });

		const batches = sweepBatches([
			run('agility', 'new'),
			run('strength', 'new'),
			run('saved-before-batches'),
			run('haste', 'old'),
			run('crit', 'old'),
			run('hit', 'old'),
		]);

		expect(batches.map(({ id, runs }) => [id, runs.map(run => run.id)])).toEqual([
			['new', ['strength', 'agility']],
			['saved-before-batches', ['saved-before-batches']],
			['old', ['hit', 'crit', 'haste']],
		]);
	});
});

describe('color slots', () => {
	it('starts at 0 and follows the newest saved run, wrapping at the slot count', () => {
		expect(nextColorSlot([], 8)).toBe(0);
		expect(nextColorSlot([{ colorSlot: 2 }, { colorSlot: 5 }], 8)).toBe(3);
		expect(nextColorSlot([{ colorSlot: 7 }], 8)).toBe(0);
	});

	it('moves the later of two runs sharing a slot to the next free one', () => {
		expect(
			displayColorSlots(
				[
					{ id: 'a', colorSlot: 1 },
					{ id: 'b', colorSlot: 2 },
					{ id: 'c', colorSlot: 1 },
				],
				4,
			),
		).toEqual({ a: 1, b: 2, c: 3 });
	});
});

describe('statSweepCsv', () => {
	const metrics = (dps: number, dpsStdev: number, n: number) => ({
		dps: { avg: dps, stdev: dpsStdev, n },
		hps: { avg: 5, stdev: 1, n },
		tps: { avg: dps * 2, stdev: dpsStdev * 2, n },
		dtps: { avg: 300, stdev: 30, n },
		tmi: { avg: 9.5, stdev: 0.25, n },
	});

	it('writes one row per point of every run, with its batch, the stat added and its unit, every metric and its spread, and a quoted label', () => {
		const csv = statSweepCsv([
			{
				label: 'Strength, ±100',
				unitStat: STRENGTH,
				batchLabel: 'Crit, Strength ±10%',
				statLabel: 'Strength',
				baseStatValue: 120,
				points: [
					{ offset: -2, bonus: -20, statValue: 100, metrics: metrics(1500.5, 100, 10000) },
					{ offset: 0, bonus: 0, statValue: 120, metrics: metrics(1520, 0, 10000) },
				],
			},
			{
				label: 'Crit',
				unitStat: RANGED_CRIT_PERCENT,
				batchLabel: '',
				statLabel: 'Ranged "Crit"',
				baseStatValue: 5,
				points: [{ offset: 1, bonus: 1, statValue: 6, metrics: metrics(1600, 200, 400) }],
			},
		]);

		expect(csv.split('\n')).toEqual([
			'run,batch,stat,base_total,added,unit,total,dps,dps_stdev,hps,hps_stdev,tps,tps_stdev,dtps,dtps_stdev,tmi,tmi_stdev,samples',
			'"Strength, ±100","Crit, Strength ±10%",Strength,120,-20,points,100,1500.5,100,5,1,3001,200,300,30,9.5,0.25,10000',
			'"Strength, ±100","Crit, Strength ±10%",Strength,120,0,points,120,1520,0,5,1,3040,0,300,30,9.5,0.25,10000',
			'Crit,,"Ranged ""Crit""",5,1,percent,6,1600,200,5,1,3200,400,300,30,9.5,0.25,400',
			'',
		]);
	});

	it('leaves the totals of a rating stat empty and writes its percent added, with no rating anywhere in the row', () => {
		const csv = statSweepCsv([
			{
				label: 'Crit ±10%',
				unitStat: CRIT_RATING,
				batchLabel: 'Crit, Strength ±10%',
				statLabel: 'Crit',
				baseStatValue: 70,
				points: [{ offset: 2, bonus: 28, statValue: 98, metrics: metrics(1600, 200, 400) }],
			},
		]);

		expect(csv.split('\n')[1]).toBe('Crit ±10%,"Crit, Strength ±10%",Crit,,2,percent,,1600,200,5,1,3200,400,300,30,9.5,0.25,400');
	});
});

describe('sweepBonus', () => {
	it('adds the percent to a percent pseudo stat', () => {
		expect(sweepBonus(RANGED_CRIT_PERCENT, 1)).toBe(1);
		expect(sweepBonus(SHADOW_HIT_PERCENT, 2)).toBe(2);
		expect(sweepBonus(UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent), 1)).toBe(1);
	});

	it('adds the rating that buys the percent to a rating stat', () => {
		expect(sweepBonus(CRIT_RATING, 1)).toBe(14);
		expect(sweepBonus(CRIT_RATING, -2)).toBe(-28);
		expect(sweepBonus(HIT_RATING, 1)).toBe(10);
	});

	it('moves any other stat 10 points per 1%, defense too, which buys skill and no percent', () => {
		expect(SWEEP_POINTS_PER_PERCENT).toBe(10);
		expect(sweepBonus(STRENGTH, 1)).toBe(10);
		expect(sweepBonus(DEFENSE_RATING, 1)).toBe(10);
		expect(sweepBonus(MAIN_HAND_DPS, 1)).toBe(10);
		expect(sweepBonus(STRENGTH, 0.7)).toBe(7);
	});

	it('adds the percent to every percent pseudo stat the sweep offers', () => {
		const percentStats = EP_UNIT_STATS.filter(stat => stat.isPseudoStat() && PseudoStat[stat.getPseudoStat()].includes('Percent'));
		expect(percentStats.length).toBeGreaterThan(0);
		for (const stat of percentStats) expect(sweepBonus(stat, 3), stat.getKey()).toBe(3);
	});
});

describe('sweepHasTotal', () => {
	it('holds for a stat with one true total, and not for a rating stat, whose percent differs by melee, ranged and spell', () => {
		for (const stat of [CRIT_RATING, HIT_RATING, UnitStat.fromStat(Stat.StatDodgeRating)]) expect(sweepHasTotal(stat), stat.getKey()).toBe(false);
		for (const stat of [STRENGTH, DEFENSE_RATING, RANGED_CRIT_PERCENT, SHADOW_HIT_PERCENT, MAIN_HAND_DPS])
			expect(sweepHasTotal(stat), stat.getKey()).toBe(true);
	});
});

describe('sweepRangeText', () => {
	it('reads a percent stat in percent and any other stat in points', () => {
		expect(sweepRangeText([CRIT_RATING], 10)).toBe('10%');
		expect(sweepRangeText([RANGED_CRIT_PERCENT], 2.5)).toBe('2.5%');
		expect(sweepRangeText([STRENGTH], 10)).toBe('100');
		expect(sweepRangeText([DEFENSE_RATING], 0.7)).toBe('7');
	});

	it('reads a batch in percent while it holds a percent stat, and in points while it holds none', () => {
		expect(sweepRangeText([CRIT_RATING, STRENGTH], 10)).toBe('10%');
		expect(sweepRangeText([STRENGTH, DEFENSE_RATING], 10)).toBe('100');
	});

	describe('in the English labels', () => {
		beforeAll(() => {
			i18n.addResourceBundle('en', 'translation', english, true, true);
		});
		afterAll(() => {
			i18n.removeResourceBundle('en', 'translation');
		});
		const runLabel = (stat: string, unitStat: UnitStat) => i18n.t('sidebar.buttons.stat_sweep.run_label', { stat, range: sweepRangeText([unitStat], 10) });

		it('labels a mixed batch in percent and each of its runs in its own unit', () => {
			expect(runLabel('Crit', CRIT_RATING)).toBe('Crit ±10%');
			expect(runLabel('Strength', STRENGTH)).toBe('Strength ±100');
			expect(i18n.t('sidebar.buttons.stat_sweep.batch_label', { stats: 'Crit, Strength', range: sweepRangeText([CRIT_RATING, STRENGTH], 10) })).toBe(
				'Crit, Strength ±10%',
			);
		});
	});
});
