import type { StatSweepRecord } from '@sim/cache/stat_sweep_history';
import type { PrecisionSettings } from '@sim/settings/precision';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { Z_90 } from '@sim/utils/math';
import { describe, expect, it, vi } from 'vitest';

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
	sweepBatches,
	type SweepPassRunner,
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

describe('sweepSteps', () => {
	it('places each offset at its total and drops those whose total falls below 0, keeping one that lands on 0', () => {
		expect(sweepSteps([-30, -20, -10, 0, 10], [-13, 0, 11, 22, 33])).toEqual([
			{ offset: -20, statValue: 0 },
			{ offset: -10, statValue: 11 },
			{ offset: 0, statValue: 22 },
			{ offset: 10, statValue: 33 },
		]);
	});

	it('rounds float noise out of the totals', () => {
		expect(sweepSteps([10], [0.1 + 0.2])).toEqual([{ offset: 10, statValue: 0.3 }]);
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
	const stepsOf = (...offsets: number[]) => offsets.map(offset => ({ offset, statValue: 100 + offset }));
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
		expect(run.cappedStep).toEqual({ offset: 20, statValue: 120 });
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

	it('writes one row per point of every run, with its batch, the stat added, every metric and its spread, and a quoted label', () => {
		const csv = statSweepCsv([
			{
				label: 'Hit, ±100',
				batchLabel: 'Hit Rating, Crit Rating ±100',
				statLabel: 'Hit Rating',
				baseStatValue: 120,
				points: [
					{ offset: -20, statValue: 100, metrics: metrics(1500.5, 100, 10000) },
					{ offset: 0, statValue: 120, metrics: metrics(1520, 0, 10000) },
				],
			},
			{
				label: 'Crit',
				batchLabel: '',
				statLabel: 'Crit "Rating"',
				baseStatValue: 50,
				points: [{ offset: 10, statValue: 61, metrics: metrics(1600, 200, 400) }],
			},
		]);

		expect(csv.split('\n')).toEqual([
			'run,batch,stat,base_total,added,total,dps,dps_stdev,hps,hps_stdev,tps,tps_stdev,dtps,dtps_stdev,tmi,tmi_stdev,samples',
			'"Hit, ±100","Hit Rating, Crit Rating ±100",Hit Rating,120,-20,100,1500.5,100,5,1,3001,200,300,30,9.5,0.25,10000',
			'"Hit, ±100","Hit Rating, Crit Rating ±100",Hit Rating,120,0,120,1520,0,5,1,3040,0,300,30,9.5,0.25,10000',
			'Crit,,"Crit ""Rating""",50,10,61,1600,200,5,1,3200,400,300,30,9.5,0.25,400',
			'',
		]);
	});
});
