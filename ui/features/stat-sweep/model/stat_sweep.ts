import type { AggregatorData, DistributionMetrics } from '@generated/proto/api';
import type { MetricSummary, StatSweepPoint, StatSweepRecord } from '@sim/cache/stat_sweep_history';
import type { UnitStat } from '@sim/proto/stats';
import { precisionRunnable, type PrecisionSettings } from '@sim/settings/precision';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { distinct } from '@sim/utils/collections';
import { toCsv } from '@sim/utils/csv';
import { mergeAggregates, roundFloat, sum, timeRemainingAtRate, Z_90 } from '@sim/utils/math';
import { STAT_WEIGHT_MAX_TOP_UP_PASSES, STAT_WEIGHT_TARGET_PROBE_ITERATIONS } from '@sim/wasm/stat_weights_constants_auto_gen';

export type StatSweepStep = Pick<StatSweepPoint, 'offset' | 'bonus' | 'statValue'>;

export type MetricSample = Pick<DistributionMetrics, 'avg' | 'stdev'> & AggregatorData;

/** The first player's DPS, HPS, TPS, DTPS and TMI over one pass. */
export type StatSweepSample = Record<StatSweepMetric, MetricSample>;

/** Every metric a step records, in the order the toggle and the CSV list them. */
export const STAT_SWEEP_METRICS: ReadonlyArray<StatSweepMetric> = ['dps', 'hps', 'tps', 'dtps', 'tmi'];

/** DPS and TPS always, HPS with the healing metrics shown, DTPS and TMI with the threat and tank options on. */
export const offeredSweepMetrics = ({ healing, threat }: { healing: boolean; threat: boolean }): StatSweepMetric[] =>
	STAT_SWEEP_METRICS.filter(metric => metric === 'dps' || metric === 'tps' || (metric === 'hps' ? healing : threat));

/** The stored metric while it is offered, else DPS. */
export const shownSweepMetric = (stored: StatSweepMetric, offered: ReadonlyArray<StatSweepMetric>): StatSweepMetric =>
	offered.includes(stored) ? stored : 'dps';

/**
 * Multiples of `step` from −`range` to +`range`, ascending, always with 0. ±`range` itself is left out when `range` is
 * not a multiple of `step`.
 */
export function buildSweepOffsets(range: number, step: number): number[] {
	if (!(step > 0) || !(range > 0)) return [0];
	const offsets = [0];
	const count = Math.floor(roundFloat(range / step));
	for (let k = 1; k <= count; k++) {
		const offset = roundFloat(k * step);
		offsets.push(offset, -offset);
	}
	return offsets.sort((a, b) => a - b);
}

/**
 * The bonus a step adds to `unitStat`. Range and step are rating, so a % stat moves by the % that rating gives (10 hit rating
 * = 1% hit), as the EP calculation scales them; a stat with no rating form takes the offset as is.
 */
export const sweepBonus = (unitStat: UnitStat, offset: number): number => unitStat.convertRatingToDefaultUnits(offset) ?? offset;

/** Each offset with the bonus it adds to `unitStat` and the stat total it reaches, `totals` in the same order; an offset whose total falls below 0 is dropped. */
export const sweepSteps = (unitStat: UnitStat, offsets: readonly number[], totals: readonly number[]): StatSweepStep[] =>
	offsets.map((offset, index) => ({ offset, bonus: sweepBonus(unitStat, offset), statValue: roundFloat(totals[index]) })).filter(step => step.statValue >= 0);

/** The stored runs plus any run in memory they do not hold yet, newest first. */
export const mergeSweepHistory = (inMemory: readonly StatSweepRecord[], stored: readonly StatSweepRecord[]): StatSweepRecord[] =>
	distinct([...stored, ...inMemory], (a, b) => a.id === b.id).sort((a, b) => b.createdAt - a.createdAt);

export interface SweepBatch<R> {
	id: string;
	/** In run order. */
	runs: R[];
}

/** `runs`, given newest first, grouped by sweep, newest sweep first; a run with no batch id is a batch of its own, keyed by its id. */
export function sweepBatches<R extends Pick<StatSweepRecord, 'id' | 'batchId'>>(runs: readonly R[]): SweepBatch<R>[] {
	const batches = new Map<string, R[]>();
	for (const run of runs) {
		const id = run.batchId ?? run.id;
		const batch = batches.get(id);
		if (batch) batch.push(run);
		else batches.set(id, [run]);
	}
	return [...batches].map(([id, runs]) => ({ id, runs: runs.reverse() }));
}

// A step's first pass in Target ± mode, as many as the EP calculation's probe; the iterations setting plays no part there.
const PROBE_ITERATIONS = STAT_WEIGHT_TARGET_PROBE_ITERATIONS;
// Passes after the first in Target ± mode, as many as the EP calculation's Target ± runs: the first top-up, which sizes
// from the first pass alone, and two that resize from the iterations so far.
const MAX_TOP_UP_PASSES = STAT_WEIGHT_MAX_TOP_UP_PASSES;
// The sim takes a request's iteration count as an int32.
const MAX_ITERATIONS = 2 ** 31 - 1;

// Two runs' metrics as one run over both runs' iterations reports them.
const mergeSweepSamples = (a: StatSweepSample, b: StatSweepSample): StatSweepSample =>
	Object.fromEntries(STAT_SWEEP_METRICS.map(metric => [metric, mergeAggregates(a[metric], b[metric])])) as StatSweepSample;

// The iterations at which a metric spread of `stdev` has a 90% ± of `targetCi`; a spread that is not a number asks for none.
const iterationsForCi = (stdev: number, targetCi: number) => Math.ceil(((Z_90 * (Number.isFinite(stdev) ? stdev : 0)) / targetCi) ** 2);

/** Runs `iterations` more from `seedOffset` past the run's seed; null when the run stopped. */
export type SweepPassRunner = (iterations: number, seedOffset: number) => Promise<StatSweepSample | null>;

/** A step's metrics over all its passes, with each metric's average over its first pass, which every step runs at the same seed and size. */
export interface StatSweepStepSample {
	metrics: StatSweepSample;
	firstPass: Record<StatSweepMetric, number>;
}

type FirstPass = StatSweepStepSample['firstPass'];

const averages = (sample: StatSweepSample): FirstPass => Object.fromEntries(STAT_SWEEP_METRICS.map(metric => [metric, sample[metric].avg])) as FirstPass;

/**
 * One step's metrics. Iterations mode runs once. Target ± mode runs a fixed first pass, then up to three more, each up to
 * the total the ± of `metric` needs and capped at Max × the first pass. A pass continues at the seed after the last
 * iteration, so the passes add up to one longer run. `onPlanned` hears each total before its pass. Without a target above
 * 0, or when `confirmsCap` says the first pass shows the stat capped, the first pass is all the step runs.
 */
export async function measureSweepStep(
	run: SweepPassRunner,
	precision: PrecisionSettings,
	metric: StatSweepMetric,
	defaultIterations: number,
	onPlanned?: (total: number) => void,
	confirmsCap?: (firstPass: FirstPass) => boolean,
): Promise<StatSweepStepSample | null> {
	if (precision.precisionMode !== 'target') {
		const iterations = precision.iterations || defaultIterations;
		onPlanned?.(iterations);
		const sample = await run(iterations, 0);
		return sample && { metrics: sample, firstPass: averages(sample) };
	}
	onPlanned?.(PROBE_ITERATIONS);
	let sample = await run(PROBE_ITERATIONS, 0);
	if (!sample) return null;
	const firstPass = averages(sample);
	if (!precisionRunnable(precision) || confirmsCap?.(firstPass)) return { metrics: sample, firstPass };
	const cap = Math.min(precision.maxIterationMultiplier > 0 ? PROBE_ITERATIONS * precision.maxIterationMultiplier : Infinity, MAX_ITERATIONS);
	for (let pass = 0; sample && pass < MAX_TOP_UP_PASSES; pass++) {
		const done = sample[metric].n;
		const total = Math.min(iterationsForCi(sample[metric].stdev, precision.targetCi), cap);
		if (total <= done) break;
		onPlanned?.(total);
		const more = await run(total - done, done);
		sample = more && mergeSweepSamples(sample, more);
	}
	return sample && { metrics: sample, firstPass };
}

const summaries = (sample: StatSweepSample): Record<StatSweepMetric, MetricSummary> =>
	Object.fromEntries(STAT_SWEEP_METRICS.map(metric => [metric, { avg: sample[metric].avg, stdev: sample[metric].stdev, n: sample[metric].n }])) as Record<
		StatSweepMetric,
		MetricSummary
	>;

// Up from change 0 to the top, then down from the first step below 0.
const outwardOrder = (steps: readonly StatSweepStep[]): StatSweepStep[] => [
	...steps.filter(step => step.offset >= 0).sort((a, b) => a.offset - b.offset),
	...steps.filter(step => step.offset < 0).sort((a, b) => b.offset - a.offset),
];

const sameFirstPass = (a: FirstPass, b: FirstPass) => STAT_SWEEP_METRICS.every(metric => a[metric] === b[metric]);

export interface StatSweepStatRun {
	/** The finished steps, sorted by change. */
	points: StatSweepPoint[];
	/** The highest step that changed a metric: the two steps above it each gave the same value of every metric as the step below, so more of the stat does nothing. */
	cappedStep?: StatSweepStep;
	/** Set when a step came back with no result, which ends the run. */
	stopped: boolean;
}

/**
 * Runs one stat's steps up from change 0 to the top, then down from the first step below 0. Every step runs at the same
 * seed, so a stat that no longer changes anything gives the exact first-pass average of every metric of the step below
 * it. One such repeat can be a step too small to change any roll, so the stat stops going up after two in a row: the step
 * before the first repeat is the capped step, and the steps above the second are skipped, with `onSkip` told how many.
 * The steps below 0 always all run. `measure` hears each step with the count of steps run before it and whether a first
 * pass would make that step the second repeat, and `onUpdate` hears the run after each step.
 */
export async function runSweepStat(
	steps: readonly StatSweepStep[],
	measure: (step: StatSweepStep, ran: number, confirmsCap: (firstPass: FirstPass) => boolean) => Promise<StatSweepStepSample | null>,
	onUpdate?: (run: StatSweepStatRun) => void,
	onSkip?: (count: number) => void,
): Promise<StatSweepStatRun> {
	const ordered = outwardOrder(steps);
	const run: StatSweepStatRun = { points: [], stopped: false };
	const upward: Array<{ step: StatSweepStep; firstPass: FirstPass }> = [];
	for (const [index, step] of ordered.entries()) {
		if (run.cappedStep && step.offset > 0) continue;
		const up = step.offset >= 0;
		const beforeLast = upward.at(-2);
		const last = upward.at(-1);
		const wouldCap = (firstPass: FirstPass) =>
			up && !!beforeLast && !!last && sameFirstPass(beforeLast.firstPass, last.firstPass) && sameFirstPass(last.firstPass, firstPass);
		const sample = await measure(step, run.points.length, wouldCap);
		if (!sample) return { ...run, stopped: true };
		run.points = [...run.points, { ...step, metrics: summaries(sample.metrics) }].sort((a, b) => a.offset - b.offset);
		if (wouldCap(sample.firstPass)) {
			run.cappedStep = beforeLast!.step;
			const skipped = ordered.slice(index + 1).filter(next => next.offset > 0).length;
			if (skipped) onSkip?.(skipped);
		}
		if (up) upward.push({ step, firstPass: sample.firstPass });
		onUpdate?.({ ...run });
	}
	return run;
}

export interface StatSweepProgress {
	/** Each stat's steps, in run order. */
	stepCounts: number[];
	/** The step being simulated, 0-based and counted across every stat. */
	running: number;
	/** Steps done across every stat, with the running step's done share; never falls. */
	completed: number;
	/** The running step's expected iterations, 0 until one is known. */
	iterationsPerStep: number;
	secondsRemaining?: number;
}

export const startSweepProgress = (stepCounts: number[]): StatSweepProgress => ({ stepCounts, running: 0, completed: 0, iterationsPerStep: 0 });

/** The 0-based stat that step `index`, counted across every stat, belongs to, and that stat's 0-based step. */
export function sweepStepPosition(stepCounts: readonly number[], index: number): { stat: number; step: number } {
	let stat = 0;
	let step = index;
	while (stat < stepCounts.length - 1 && step >= stepCounts[stat]) step -= stepCounts[stat++];
	return { stat, step };
}

/**
 * `progress` with step `index` (0-based, counted across every stat) `done` iterations into the `expected` it is likely
 * to run, `elapsedSeconds` into the run.
 */
export function advanceSweepProgress(progress: StatSweepProgress, index: number, done: number, expected: number, elapsedSeconds: number): StatSweepProgress {
	const share = expected > 0 ? Math.min(1, done / expected) : 0;
	const completed = Math.max(progress.completed, index + share);
	return {
		...progress,
		running: index,
		completed,
		iterationsPerStep: expected || progress.iterationsPerStep,
		secondsRemaining: timeRemainingAtRate(elapsedSeconds, completed, sum(progress.stepCounts)) ?? progress.secondsRemaining,
	};
}

/** `progress` with `count` steps of stat `statIndex` (0-based) that will not run taken out of its step count. */
export const skipSweepSteps = (progress: StatSweepProgress, statIndex: number, count: number): StatSweepProgress => ({
	...progress,
	stepCounts: progress.stepCounts.map((steps, index) => (index === statIndex ? steps - count : steps)),
});

/** The slot a new run takes: the one after the newest saved run's, wrapping at `slotCount`. */
export const nextColorSlot = (history: readonly Pick<StatSweepRecord, 'colorSlot'>[], slotCount: number): number =>
	history.length ? (history[0].colorSlot + 1) % slotCount : 0;

/**
 * The slot each shown run is drawn in, given oldest first. A run keeps its own slot unless an earlier run holds it, which
 * happens once the slots wrap around; it then takes the next free one.
 */
export function displayColorSlots(runs: readonly { id: string; colorSlot: number }[], slotCount: number): Record<string, number> {
	const shown: Record<string, number> = {};
	const taken = new Set<number>();
	for (const { id, colorSlot } of runs) {
		let slot = colorSlot;
		for (let tried = 0; taken.has(slot) && tried < slotCount; tried++) slot = (slot + 1) % slotCount;
		shown[id] = slot;
		taken.add(slot);
	}
	return shown;
}

export interface StatSweepCsvRun {
	label: string;
	/** Empty for a run saved with no batch. */
	batchLabel: string;
	statLabel: string;
	baseStatValue: number;
	points: ReadonlyArray<StatSweepPoint>;
}

export const STAT_SWEEP_CSV_COLUMNS = [
	'run',
	'batch',
	'stat',
	'base_total',
	'added',
	'bonus',
	'total',
	...STAT_SWEEP_METRICS.flatMap(metric => [metric, `${metric}_stdev`]),
	'samples',
] as const;

export function statSweepCsv(runs: ReadonlyArray<StatSweepCsvRun>): string {
	const rows = runs.flatMap(run =>
		run.points.map(point => [
			run.label,
			run.batchLabel,
			run.statLabel,
			run.baseStatValue,
			point.offset,
			point.bonus,
			point.statValue,
			...STAT_SWEEP_METRICS.flatMap(metric => [point.metrics[metric].avg, point.metrics[metric].stdev]),
			// Every metric comes from the same iterations.
			point.metrics.dps.n,
		]),
	);
	return toCsv([STAT_SWEEP_CSV_COLUMNS, ...rows]);
}
