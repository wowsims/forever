// The backend omits RequestLow for the school-split hit/crit stats: sim/core/statweight.go
// zeroes their low-side stat modifier and only populates lowSimRequest when it is non-zero.
// Nothing in the type system catches a missing guard here — a non-null assertion on the low
// request type-checks and throws at runtime — so the absent-low path is covered explicitly.
import {
	AggregatorData,
	DistributionMetrics,
	ErrorOutcome,
	ErrorOutcomeType,
	PartyMetrics,
	ProgressMetrics,
	RaidMetrics,
	RaidSimRequest,
	RaidSimRequestSplitRequest,
	RaidSimRequestSplitResult,
	RaidSimResult,
	SimOptions,
	StatWeightsCalcRequest,
	StatWeightsPairedSums,
	StatWeightsRequest,
	StatWeightsResult,
	StatWeightsSidePairs,
	StatWeightsStatData,
	StatWeightsTarget,
	StatWeightsTargetMetric,
	StatWeightsTopUpPlan,
	StatWeightsTopUpRequest,
	UnitMetrics,
} from '@generated/proto/api';
import { Encounter, Stat } from '@generated/proto/common';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../sim_signal_manager';
import { noop } from '../utils/misc';
import type { WorkerPool } from '../workers/worker_pool';
import { runConcurrentStatWeights } from './stat_weights';

type QueuedSim = Parameters<(typeof import('./sim'))['runConcurrentSimQueue']>[0][number];

const runConcurrentSimQueue = vi.hoisted(() => vi.fn());
vi.mock('./sim', () => ({ runConcurrentSimQueue }));

const queued = () => runConcurrentSimQueue.mock.calls.flatMap(([sims]) => (sims as QueuedSim[]).map(sim => sim.request));

// A request whose run adds worth to every iteration, carried in the encounter duration for the fake sim to read.
const requestWithIterations = (iterations: number, worth = 0, randomSeed = 0n) =>
	RaidSimRequest.create({ simOptions: SimOptions.create({ iterations, randomSeed }), encounter: Encounter.create({ duration: worth }) });

// Iteration i of a run has seed randomSeed + i, and every run shares the noise at that seed. Metric m
// scales the values by m + 1, so a pairing that mixes up metrics or iterations shows in the sums.
const simulate = (request: RaidSimRequest) => {
	const { iterations, randomSeed } = request.simOptions!;
	const worth = request.encounter!.duration;
	const metric = (scale: number) =>
		DistributionMetrics.create({
			allValues: Array.from({ length: iterations }, (_, i) => scale * (1000 + (((Number(randomSeed) + i) * 7) % 13) + worth)),
			aggregatorData: AggregatorData.create({ n: iterations }),
		});
	const player = UnitMetrics.create({
		dps: metric(1),
		hps: metric(2),
		threat: metric(3),
		dtps: metric(4),
		tmi: metric(5),
		chanceOfDeath: Number(randomSeed) / 1000,
	});
	return RaidSimResult.create({ raidMetrics: RaidMetrics.create({ parties: [PartyMetrics.create({ players: [player] })] }), iterationsDone: iterations });
};

// A run that the queue left in one split.
const simulateSplits = (request: RaidSimRequest) => [simulate(request)];

// The pairs of n iterations of a run that adds worth to every baseline iteration.
const expectedPairs = (n: number, worth: number, chanceOfDeath: number) => {
	const sums = (diff: number) => StatWeightsPairedSums.create({ n, sum: n * diff, sumSq: n * diff * diff });
	return StatWeightsSidePairs.create({
		dps: sums(worth),
		hps: sums(2 * worth),
		tps: sums(3 * worth),
		dtps: sums(4 * worth),
		tmi: sums(5 * worth),
		chanceOfDeath,
	});
};

const carriesPerIterationValues = (calc: StatWeightsCalcRequest) => JSON.stringify(StatWeightsCalcRequest.toJson(calc)).includes('allValues');

const signals = { abort: { isTriggered: () => false } } as unknown as SimSignals;

const makePool = (statSimRequests: Array<{ statData?: StatWeightsStatData; requestLow?: RaidSimRequest; requestHigh: RaidSimRequest }>) =>
	({
		statWeightRequests: vi.fn(async () => ({
			baseRequest: requestWithIterations(100),
			epReferenceStat: 0,
			statSimRequests: statSimRequests.map(statRequest => ({ statData: StatWeightsStatData.create(), ...statRequest })),
		})),
		statWeightCompute: vi.fn(async () => StatWeightsResult.create()),
	}) as unknown as WorkerPool;

describe('runConcurrentStatWeights', () => {
	beforeEach(() => {
		runConcurrentSimQueue.mockReset();
		// Report one iteration per sim so the caller's progress handler runs and exposes the totals.
		runConcurrentSimQueue.mockImplementation(async (sims: QueuedSim[]) => {
			for (const { request, onProgress, onResult } of sims) {
				onResult(simulateSplits(request));
				onProgress(ProgressMetrics.create({ completedIterations: request.simOptions!.iterations }));
			}
		});
	});

	it('counts every sim in the last progress', async () => {
		const { runConcurrentSimQueue: queue } = await vi.importActual<typeof import('./sim')>('./sim');
		runConcurrentSimQueue.mockImplementation(queue);
		const pool = Object.assign(makePool([{ requestLow: requestWithIterations(20), requestHigh: requestWithIterations(30) }]), {
			getNumWorkers: () => 1,
			raidSimRequestSplit: async ({ request }: RaidSimRequestSplitRequest) => RaidSimRequestSplitResult.create({ splitsDone: 1, requests: [request!] }),
			raidSimAsync: async (request: RaidSimRequest, onProgress: (pm: ProgressMetrics) => void) => {
				const result = simulate(request);
				onProgress(
					ProgressMetrics.create({ totalIterations: result.iterationsDone, completedIterations: result.iterationsDone, finalRaidResult: result }),
				);
				return result;
			},
		});
		const progress: ProgressMetrics[] = [];

		await runConcurrentStatWeights({} as never, pool, pm => progress.push(pm), signals);

		const last = progress.filter(pm => !pm.finalWeightResult).at(-1)!;
		expect([last.completedSims, last.totalSims]).toEqual([3, 3]);
		expect([last.completedIterations, last.totalIterations]).toEqual([150, 150]);
	});

	it('skips the low sim, and its iterations, when requestLow is unset', async () => {
		const pool = makePool([{ requestHigh: requestWithIterations(30) }]);
		const progress: ProgressMetrics[] = [];

		await runConcurrentStatWeights({} as never, pool, pm => progress.push(pm), signals);

		// Baseline + high only; the absent low sim is neither run nor counted.
		expect(queued()).toHaveLength(2);
		expect(progress[0].totalSims).toBe(2);
		expect(progress[0].totalIterations).toBe(130);

		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		expect(calcRequest.statPairs[0].low).toBeUndefined();
		expect(calcRequest.statPairs[0].high).toBeDefined();
	});

	it('runs the low sim when requestLow is present', async () => {
		const pool = makePool([{ requestLow: requestWithIterations(20), requestHigh: requestWithIterations(30) }]);
		const progress: ProgressMetrics[] = [];

		await runConcurrentStatWeights({} as never, pool, pm => progress.push(pm), signals);

		expect(queued()).toHaveLength(3);
		expect(progress[0].totalSims).toBe(3);
		expect(progress[0].totalIterations).toBe(150);

		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		expect(calcRequest.statPairs[0].low).toBeDefined();
	});

	it('pairs each stat run with the baseline at the same iterations and sends the sums alone', async () => {
		const pool = makePool([
			{ statData: StatWeightsStatData.create({ unitStat: 1 }), requestLow: requestWithIterations(100, -10), requestHigh: requestWithIterations(100, 10) },
		]);

		await runConcurrentStatWeights({} as never, pool, noop, signals);

		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		expect(calcRequest.statPairs[0].low).toEqual(expectedPairs(100, -10, 0));
		expect(calcRequest.statPairs[0].high).toEqual(expectedPairs(100, 10, 0));
		expect([calcRequest.probeIterations, calcRequest.baselineIterations]).toEqual([100, 100]);
		expect(carriesPerIterationValues(calcRequest)).toBe(false);
	});

	it('runs every stat sim in one queue and pairs each with the baseline whichever ends first', async () => {
		const pool = makePool([
			{ statData: StatWeightsStatData.create({ unitStat: 1 }), requestLow: requestWithIterations(100, -10), requestHigh: requestWithIterations(100, 10) },
			{ statData: StatWeightsStatData.create({ unitStat: 2 }), requestHigh: requestWithIterations(100, 20) },
		]);
		runConcurrentSimQueue.mockImplementation(async (sims: QueuedSim[]) => {
			for (const { request, onResult } of [...sims].reverse()) onResult(simulateSplits(request));
		});

		await runConcurrentStatWeights({} as never, pool, noop, signals);

		expect(runConcurrentSimQueue.mock.calls.map(([sims]) => sims.length)).toEqual([1, 3]);
		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		expect(calcRequest.statPairs.map(pairs => pairs.statData!.unitStat)).toEqual([1, 2]);
		expect(calcRequest.statPairs[0].low).toEqual(expectedPairs(100, -10, 0));
		expect(calcRequest.statPairs[0].high).toEqual(expectedPairs(100, 10, 0));
		expect(calcRequest.statPairs[1].high).toEqual(expectedPairs(100, 20, 0));
	});

	it("reads a run's splits as the run whole, with its chance of death weighted by each split's iterations", async () => {
		const seed = 7n;
		const pool = makePool([]);
		vi.mocked(pool.statWeightRequests).mockResolvedValue({
			baseRequest: requestWithIterations(100, 0, seed),
			epReferenceStat: 0,
			hpsReferenceStat: 0,
			statSimRequests: [{ statData: StatWeightsStatData.create({ unitStat: 1 }), requestHigh: requestWithIterations(100, 10, seed) }],
		});
		// The first 60 iterations on one worker and the last 40 on another, as the sim splits a run.
		runConcurrentSimQueue.mockImplementation(async (sims: QueuedSim[]) => {
			for (const { request, onResult } of sims) {
				const { iterations, randomSeed } = request.simOptions!;
				const split = (n: number, offset: number) =>
					simulate({ ...request, simOptions: SimOptions.create({ iterations: n, randomSeed: randomSeed + BigInt(offset) }) });
				onResult([split(60, 0), split(iterations - 60, 60)]);
			}
		});

		await runConcurrentStatWeights({} as never, pool, noop, signals);

		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		const chanceOfDeath = 0 + (7 / 1000) * (60 / 100) + (67 / 1000) * (40 / 100);
		expect(calcRequest.baseChanceOfDeath).toBe(chanceOfDeath);
		expect(calcRequest.statPairs[0].high).toEqual(expectedPairs(100, 10, chanceOfDeath));
	});

	it('stops at a failed sim and reports its error', async () => {
		vi.spyOn(console, 'error').mockImplementation(noop);
		const pool = makePool([{ requestHigh: requestWithIterations(100, 10) }]);
		const failure = ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeError, message: 'failed' });
		runConcurrentSimQueue.mockImplementation(async (sims: QueuedSim[]) => {
			if (runConcurrentSimQueue.mock.calls.length > 1) return failure;
			sims.forEach(({ request, onResult }) => onResult(simulateSplits(request)));
		});

		const result = await runConcurrentStatWeights({} as never, pool, noop, signals);

		expect(result.error).toEqual(failure);
		expect(pool.statWeightCompute).not.toHaveBeenCalled();
	});

	describe('with a target ±', () => {
		const probe = 100;
		const seed = 7n;
		const unitStat = 1;
		const target = StatWeightsTarget.create({
			epCi: 0.01,
			maxIterationMultiplier: 3,
			metrics: [StatWeightsTargetMetric.StatWeightsTargetMetricDps, StatWeightsTargetMetric.StatWeightsTargetMetricHps],
		});
		const request = StatWeightsRequest.create({ target, hpsReferenceStat: Stat.StatIntellect });

		const makeTargetPool = (plan: (call: number, topUpRequest: StatWeightsTopUpRequest) => StatWeightsTopUpPlan) => {
			let calls = 0;
			return {
				statWeightRequests: vi.fn(async () => ({
					baseRequest: requestWithIterations(probe, 0, seed),
					epReferenceStat: 0,
					hpsReferenceStat: Stat.StatIntellect,
					statSimRequests: [
						{
							statData: StatWeightsStatData.create({ unitStat }),
							requestLow: requestWithIterations(probe, -10, seed),
							requestHigh: requestWithIterations(probe, 10, seed),
						},
					],
				})),
				statWeightTopUpPlan: vi.fn(async (topUpRequest: StatWeightsTopUpRequest) => plan(calls++, topUpRequest)),
				statWeightCompute: vi.fn(async () => StatWeightsResult.create()),
			} as unknown as WorkerPool;
		};

		it('continues each planned run at the seed after its last iteration and adds its pairs to the probe', async () => {
			const pool = makeTargetPool(call =>
				call === 0
					? StatWeightsTopUpPlan.create({ baselineIterations: 3 * probe, stats: [{ unitStat, iterations: 2 * probe }] })
					: StatWeightsTopUpPlan.create(),
			);
			const progress: ProgressMetrics[] = [];

			await runConcurrentStatWeights(request, pool, pm => progress.push(pm), signals);

			const topUpRequest = vi.mocked(pool.statWeightTopUpPlan).mock.calls[0][0];
			expect(topUpRequest.target).toEqual(target);
			expect(topUpRequest.calc!.hpsReferenceStat).toBe(Stat.StatIntellect);
			const topUps = queued()
				.slice(3)
				.map(simRequest => simRequest.simOptions!);
			expect(topUps.map(options => options.iterations)).toEqual([2 * probe, probe, probe]);
			expect(topUps.map(options => options.randomSeed)).toEqual([seed + BigInt(probe), seed + BigInt(probe), seed + BigInt(probe)]);

			const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
			const probeChanceOfDeath = Number(seed) / 1000;
			expect(calcRequest.statPairs[0].low).toEqual(expectedPairs(2 * probe, -10, probeChanceOfDeath));
			expect(calcRequest.statPairs[0].high).toEqual(expectedPairs(2 * probe, 10, probeChanceOfDeath));
			expect([calcRequest.probeIterations, calcRequest.baselineIterations]).toEqual([probe, 3 * probe]);
			expect(calcRequest.baseChanceOfDeath).toBe(probeChanceOfDeath);

			const last = progress.filter(pm => !pm.finalWeightResult).at(-1)!;
			expect(last.totalSims).toBe(6);
			expect(last.totalIterations).toBe(3 * probe + 2 * (2 * probe));
		});

		it('asks for a plan after every pass, pairs a stat topped up past the probe with the baseline run on, and stops at the first empty plan', async () => {
			const pool = makeTargetPool(call =>
				call < 2
					? StatWeightsTopUpPlan.create({ baselineIterations: 3 * probe, stats: [{ unitStat, iterations: (call + 2) * probe }] })
					: StatWeightsTopUpPlan.create(),
			);

			await runConcurrentStatWeights(request, pool, noop, signals);

			expect(vi.mocked(pool.statWeightTopUpPlan).mock.calls.map(([topUpRequest]) => topUpRequest.passes)).toEqual([0, 1, 2]);
			// The probe's three runs, the first pass's baseline and stat runs, and the second pass's stat runs alone.
			expect(queued()).toHaveLength(3 + 3 + 2);
			const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
			expect(calcRequest.statPairs[0].low).toEqual(expectedPairs(3 * probe, -10, Number(seed) / 1000));
			expect(calcRequest.statPairs[0].high).toEqual(expectedPairs(3 * probe, 10, Number(seed) / 1000));
			expect([calcRequest.probeIterations, calcRequest.baselineIterations]).toEqual([probe, 3 * probe]);
		});

		it('sends no per-iteration values to the worker', async () => {
			// Each plan request as sent: the caller goes on adding to the calc request after the call.
			const planned: StatWeightsCalcRequest[] = [];
			const pool = makeTargetPool((call, topUpRequest) => {
				planned.push(StatWeightsCalcRequest.clone(topUpRequest.calc!));
				return call === 0
					? StatWeightsTopUpPlan.create({ baselineIterations: 2 * probe, stats: [{ unitStat, iterations: 2 * probe }] })
					: StatWeightsTopUpPlan.create();
			});

			await runConcurrentStatWeights(request, pool, noop, signals);

			expect(planned).toHaveLength(2);
			for (const calc of [...planned, vi.mocked(pool.statWeightCompute).mock.calls[0][0]]) {
				expect(carriesPerIterationValues(calc)).toBe(false);
			}
		});
	});
});
