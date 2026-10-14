import {
	ErrorOutcome,
	ErrorOutcomeType,
	ProgressMetrics,
	RaidSimRequest,
	RaidSimResult,
	StatWeightsCalcRequest,
	StatWeightsPairedSums,
	StatWeightsRequest,
	StatWeightsResult,
	StatWeightsSidePairs,
	StatWeightsStatPairs,
	StatWeightsTopUpRequest,
} from '@generated/proto/api';
import { SimRequest } from '@worker/types';

import { SimSignals } from '../sim_signal_manager';
import { isDevMode } from '../utils/env';
import { sum } from '../utils/math';
import { generateRequestId, WorkerPool, WorkerProgressCallback } from '../workers/worker_pool';
import { runConcurrentSimQueue } from './sim';

const makeAndSendWeightsError = (err: string | ErrorOutcome, onProgress: WorkerProgressCallback): StatWeightsResult => {
	const errRes = StatWeightsResult.create();
	if (typeof err === 'string') {
		console.error(err);
		errRes.error = ErrorOutcome.create({ message: err });
	} else {
		if (err.message) console.error(err.message);
		errRes.error = err;
	}
	onProgress(ProgressMetrics.create({ finalWeightResult: errRes }));
	return errRes;
};

const firstPlayer = (result: RaidSimResult) => result.raidMetrics!.parties[0].players[0];

// The first player's per-iteration DPS, HPS, threat, DTPS and TMI, in this order. A run's splits list them in split order,
// which is the order their combined result lists them in.
const RUN_METRICS = ['dps', 'hps', 'threat', 'dtps', 'tmi'] as const;

const pushRunValues = (values: number[][], splits: RaidSimResult[]) =>
	RUN_METRICS.forEach((metric, m) => {
		for (const split of splits) for (const value of firstPlayer(split)[metric]!.allValues) values[m].push(value);
	});

// The first player's chance of death in the splits' combined result: each split's weighted by its share of the iterations,
// summed in split order, in the same order of operations as CombineConcurrentSimResults in sim/core/sim_concurrent.go.
const chanceOfDeath = (splits: RaidSimResult[]): number => {
	const iterations = sum(splits.map(split => split.iterationsDone));
	return splits.reduce((chance, split) => chance + firstPlayer(split).chanceOfDeath * (split.iterationsDone / iterations), 0);
};

// Pairs each iteration of a stat run with the baseline's iteration offset + i, which ran at the same
// seed, in the same order of operations as pairStatWeightRun in sim/core/statweight.go.
const pairRun = (baseline: number[][], splits: RaidSimResult[], offset: number): StatWeightsSidePairs => {
	const [dps, hps, tps, dtps, tmi] = RUN_METRICS.map((metric, m) => {
		let i = 0;
		let total = 0;
		let sumSq = 0;
		for (const split of splits) {
			for (const value of firstPlayer(split)[metric]!.allValues) {
				const diff = value - baseline[m][offset + i++];
				total += diff;
				sumSq += diff * diff;
			}
		}
		return StatWeightsPairedSums.create({ n: i, sum: total, sumSq });
	});
	return StatWeightsSidePairs.create({ dps, hps, tps, dtps, tmi, chanceOfDeath: chanceOfDeath(splits) });
};

// Adds a later pass's pairs to a side's. The chance of death stays the first run's.
const addPairs = (side: StatWeightsSidePairs, pass: StatWeightsSidePairs) => {
	for (const metric of ['dps', 'hps', 'tps', 'dtps', 'tmi'] as const) {
		side[metric]!.n += pass[metric]!.n;
		side[metric]!.sum += pass[metric]!.sum;
		side[metric]!.sumSq += pass[metric]!.sumSq;
	}
};

export const runConcurrentStatWeights = async (
	request: StatWeightsRequest,
	workerPool: WorkerPool,
	onProgress: WorkerProgressCallback,
	signals: SimSignals,
): Promise<StatWeightsResult> => {
	if (isDevMode()) {
		console.log('Getting stat weight sim requests.');
	}

	const newRaidSimRequestId = () => generateRequestId(SimRequest.raidSimAsync);

	const aborted = () => makeAndSendWeightsError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);

	const manualResponse = await workerPool.statWeightRequests(request);
	manualResponse.baseRequest!.requestId = newRaidSimRequestId();

	if (signals.abort.isTriggered()) return aborted();

	let iterationsTotal = manualResponse.baseRequest!.simOptions!.iterations;
	let iterationsDone = 0;
	let simsTotal = 1;
	let simsDone = 0;

	// The backend leaves RequestLow unset for the school-split hit/crit stats, whose low-side
	// modifier is always zero (sim/core/statweight.go). Skip the low sim when it is absent rather
	// than counting iterations for a request that does not exist.
	for (const statReqData of manualResponse.statSimRequests) {
		if (statReqData.requestLow) {
			iterationsTotal += statReqData.requestLow.simOptions!.iterations;
			simsTotal += 1;
		}
		iterationsTotal += statReqData.requestHigh!.simOptions!.iterations;
		simsTotal += 1;
	}

	if (isDevMode()) {
		console.log(`Need to run a total of ${simsTotal} sims and ${iterationsTotal} iterations.`);
	}

	// One per sim, as the sims of a pass run at once.
	const progressHandler = () => {
		let lastIterations = 0;
		return (pm: ProgressMetrics) => {
			iterationsDone += pm.completedIterations - lastIterations;
			lastIterations = pm.completedIterations;

			onProgress(
				ProgressMetrics.create({
					totalIterations: iterationsTotal,
					completedIterations: iterationsDone,
					totalSims: simsTotal,
					completedSims: simsDone,
				}),
			);
		};
	};

	// Runs the sims as one queue of splits across the workers. Each hands its splits to `take` as it ends, so no whole result
	// outlives its sim, and as each `take` writes its own slot the values do not depend on which sim ends first. The splits
	// stay uncombined: the weights read the first player's values alone, which the splits hold in full.
	const runSims = (sims: Array<{ request: RaidSimRequest; take: (splits: RaidSimResult[]) => void }>) =>
		runConcurrentSimQueue(
			sims.map(({ request, take }) => ({
				request,
				onProgress: progressHandler(),
				onResult: splits => {
					simsDone++;
					take(splits);
				},
			})),
			workerPool,
			signals,
		);

	// Only the baseline keeps its per-iteration values, for the stat runs of later passes to pair with.
	const baselineValues: number[][] = RUN_METRICS.map(() => []);
	let baseChanceOfDeath = 0;
	const baseError = await runSims([
		{
			request: manualResponse.baseRequest!,
			take: splits => {
				pushRunValues(baselineValues, splits);
				baseChanceOfDeath = chanceOfDeath(splits);
			},
		},
	]);
	if (baseError) return makeAndSendWeightsError(baseError, onProgress);

	const calcRequest = StatWeightsCalcRequest.create({
		epReferenceStat: manualResponse.epReferenceStat,
		hpsReferenceStat: manualResponse.hpsReferenceStat,
		probeIterations: baselineValues[0].length,
		baselineIterations: baselineValues[0].length,
		baseChanceOfDeath,
	});
	const statPairs = new Map<number, StatWeightsStatPairs>();

	const paired = (request: RaidSimRequest, offset: number, take: (pairs: StatWeightsSidePairs) => void) => ({
		request,
		take: (splits: RaidSimResult[]) => take(pairRun(baselineValues, splits, offset)),
	});

	const statSims = manualResponse.statSimRequests.flatMap(statReqData => {
		const pairs = StatWeightsStatPairs.create({ statData: statReqData.statData });
		calcRequest.statPairs.push(pairs);
		statPairs.set(statReqData.statData!.unitStat, pairs);
		statReqData.requestHigh!.requestId = newRaidSimRequestId();
		const high = paired(statReqData.requestHigh!, 0, high => (pairs.high = high));
		if (!statReqData.requestLow) return [high];
		statReqData.requestLow.requestId = newRaidSimRequestId();
		return [paired(statReqData.requestLow, 0, low => (pairs.low = low)), high];
	});
	const statError = await runSims(statSims);
	if (statError) return makeAndSendWeightsError(statError, onProgress);

	if (request.target) {
		const baseRequest = manualResponse.baseRequest!;
		const seed = baseRequest.simOptions!.randomSeed;
		const statRequests = new Map(manualResponse.statSimRequests.map(statReqData => [statReqData.statData!.unitStat, statReqData]));

		// Iteration i runs at seed randomSeed + i, so a run continued at the seed after its last iteration
		// repeats what one longer run would have simulated.
		const continued = (simRequest: RaidSimRequest, done: number, iterations: number): RaidSimRequest => ({
			...simRequest,
			requestId: newRaidSimRequestId(),
			simOptions: { ...simRequest.simOptions!, iterations, randomSeed: seed + BigInt(done) },
		});

		for (let passes = 0; ; passes++) {
			const plan = await workerPool.statWeightTopUpPlan(StatWeightsTopUpRequest.create({ calc: calcRequest, target: request.target, passes }));
			if (!plan.stats.length) break;

			const baselineIterations = calcRequest.baselineIterations;
			const baselineDelta = plan.baselineIterations - baselineIterations;
			if (baselineDelta > 0) {
				iterationsTotal += baselineDelta;
				simsTotal += 1;
			}
			for (const planned of plan.stats) {
				const runs = statRequests.get(planned.unitStat)!.requestLow ? 2 : 1;
				iterationsTotal += runs * (planned.iterations - statPairs.get(planned.unitStat)!.high!.dps!.n);
				simsTotal += runs;
			}

			if (baselineDelta > 0) {
				const baseError = await runSims([
					{
						request: continued(baseRequest, baselineIterations, baselineDelta),
						take: splits => pushRunValues(baselineValues, splits),
					},
				]);
				if (baseError) return makeAndSendWeightsError(baseError, onProgress);
				calcRequest.baselineIterations = baselineValues[0].length;
			}

			const topUps = plan.stats.flatMap(planned => {
				const statReqData = statRequests.get(planned.unitStat)!;
				const pairs = statPairs.get(planned.unitStat)!;
				const done = pairs.high!.dps!.n;
				const delta = planned.iterations - done;
				const high = paired(continued(statReqData.requestHigh!, done, delta), done, high => addPairs(pairs.high!, high));
				if (!statReqData.requestLow) return [high];
				return [paired(continued(statReqData.requestLow, done, delta), done, low => addPairs(pairs.low!, low)), high];
			});
			const topUpError = await runSims(topUps);
			if (topUpError) return makeAndSendWeightsError(topUpError, onProgress);
		}
	}

	if (isDevMode()) {
		console.log(`All ${simsTotal} sims finished successfully. Computing weights.`);
	}

	const weightResult = await workerPool.statWeightCompute(calcRequest);
	if (weightResult.error) return makeAndSendWeightsError(weightResult.error, onProgress);
	onProgress(ProgressMetrics.create({ finalWeightResult: weightResult }));
	return weightResult;
};
