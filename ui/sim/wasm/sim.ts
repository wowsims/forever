import {
	ErrorOutcome,
	ErrorOutcomeType,
	ProgressMetrics,
	RaidSimRequest,
	RaidSimRequestSplitRequest,
	type RaidSimRequestSplitResult,
	RaidSimResult,
	RaidSimResultCombinationRequest,
} from '@generated/proto/api';

import { SimSignals } from '../sim_signal_manager';
import { isDevMode } from '../utils/env';
import { sum } from '../utils/math';
import { noop } from '../utils/misc';
import { WorkerPool, WorkerProgressCallback } from '../workers/worker_pool';

class ConcurrentSimProgress {
	readonly concurrency: number;
	readonly iterationsTotal: number;
	private readonly iterationsDone: number[];
	private readonly dpsValues: number[];
	private readonly hpsValues: number[];
	readonly finalResults: RaidSimResult[];

	constructor(concurrency: number, totalIterations: number) {
		this.concurrency = concurrency;
		this.iterationsTotal = totalIterations;
		this.iterationsDone = Array(this.concurrency).fill(0);
		this.dpsValues = Array(this.concurrency).fill(0);
		this.hpsValues = Array(this.concurrency).fill(0);
		this.finalResults = Array(this.concurrency);
	}

	getIterationsDone(): number {
		let total = 0;
		for (const done of this.iterationsDone) {
			total += done;
		}
		return total;
	}

	getDpsAvg(): number {
		let total = 0;
		for (const done of this.dpsValues) {
			total += done;
		}
		return total / this.concurrency;
	}

	getHpsAvg(): number {
		let total = 0;
		for (const done of this.hpsValues) {
			total += done;
		}
		return total / this.concurrency;
	}

	updateProgress(idx: number, msg: ProgressMetrics) {
		this.iterationsDone[idx] = msg.completedIterations;
		this.dpsValues[idx] = msg.dps;
		this.hpsValues[idx] = msg.hps;

		if (msg.finalRaidResult) {
			this.finalResults[idx] = msg.finalRaidResult;
		}
	}

	makeProgressMetrics(): ProgressMetrics {
		return ProgressMetrics.create({
			totalIterations: this.iterationsTotal,
			completedIterations: this.getIterationsDone(),
			dps: this.getDpsAvg(),
			hps: this.getHpsAvg(),
		});
	}
}

interface SimRunResult {
	errorResult?: RaidSimResult;
	results: RaidSimResult[];
	progressMetricsFinal: ProgressMetrics;
}

const runSims = (
	requests: RaidSimRequest[],
	totalIterations: number,
	wp: WorkerPool,
	onProgress: WorkerProgressCallback,
	signals: SimSignals,
): Promise<SimRunResult> => {
	return new Promise(resolve => {
		const csp = new ConcurrentSimProgress(requests.length, totalIterations);
		let progressCounter = 0;
		let running = requests.length;

		const resolveWithError = (idx: number, error: unknown) => {
			if (!running) return;

			console.error(`Worker ${idx} had an error!`, error);
			running = 0;
			signals.abort.trigger();

			const errorResult = RaidSimResult.create({
				error: ErrorOutcome.create({ message: error instanceof Error ? error.message : String(error) }),
			});
			const finalProgressMetrics = csp.makeProgressMetrics();
			finalProgressMetrics.finalRaidResult = errorResult;
			onProgress(finalProgressMetrics);
			resolve({
				errorResult,
				results: csp.finalResults,
				progressMetricsFinal: finalProgressMetrics,
			});
		};

		const progressHandler = (idx: number, pm: ProgressMetrics) => {
			if (!running) return;

			csp.updateProgress(idx, pm);

			progressCounter++;
			if (progressCounter % running == 0) {
				onProgress(csp.makeProgressMetrics());
			}

			if (pm.finalRaidResult) {
				running--;
				let errorResult: RaidSimResult | undefined;

				if (pm.finalRaidResult.error) {
					if (pm.finalRaidResult.error.type == ErrorOutcomeType.ErrorOutcomeError) {
						console.error(`Worker ${idx} had an error!`);
					}
					errorResult = pm.finalRaidResult;
					signals.abort.trigger();
				}

				if (errorResult || running == 0) {
					running = 0;
					const finalProgressMetrics = csp.makeProgressMetrics();
					finalProgressMetrics.finalRaidResult = errorResult;
					onProgress(finalProgressMetrics);
					resolve({
						errorResult: errorResult,
						results: csp.finalResults,
						progressMetricsFinal: finalProgressMetrics,
					});
					return;
				}
			}
		};

		for (let i = 0; i < requests.length; i++) {
			wp.raidSimAsync(requests[i], pm => progressHandler(i, pm), signals).catch(error => resolveWithError(i, error));
		}
	});
};

const makeAndSendRaidSimError = (err: string | ErrorOutcome, onProgress: WorkerProgressCallback): RaidSimResult => {
	const errRes = RaidSimResult.create();
	if (typeof err === 'string') {
		console.error(err);
		errRes.error = ErrorOutcome.create({ message: err });
	} else {
		if (err.message) console.error(err.message);
		errRes.error = err;
	}
	onProgress(ProgressMetrics.create({ finalRaidResult: errRes }));
	return errRes;
};

export const runConcurrentSim = async (
	request: RaidSimRequest,
	workerPool: WorkerPool,
	onProgress: WorkerProgressCallback,
	signals: SimSignals,
): Promise<RaidSimResult> => {
	if (isDevMode()) {
		console.log(`Sending requests split for ${workerPool.getNumWorkers()} splits.`);
	}

	const splitResult = await workerPool.raidSimRequestSplit(
		RaidSimRequestSplitRequest.create({
			splitCount: workerPool.getNumWorkers(),
			request: request,
		}),
	);

	if (splitResult.errorResult) {
		return makeAndSendRaidSimError(splitResult.errorResult, onProgress);
	}

	if (signals.abort.isTriggered()) {
		return makeAndSendRaidSimError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);
	}

	if (isDevMode()) {
		console.log(`Running ${request.simOptions!.iterations} iterations on ${splitResult.splitsDone} concurrent sims...`);
	}

	const simRes = await runSims(splitResult.requests, request.simOptions!.iterations, workerPool, onProgress, signals);

	if (simRes.errorResult && simRes.errorResult.error) {
		return makeAndSendRaidSimError(simRes.errorResult.error, onProgress);
	}

	if (signals.abort.isTriggered()) {
		return makeAndSendRaidSimError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);
	}

	if (isDevMode()) {
		console.log(`All ${splitResult.splitsDone} sims finished successfully. Combining ${simRes.results.length} results.`);
	}
	const combiResult = await workerPool.raidSimResultCombination(
		RaidSimResultCombinationRequest.create({
			results: simRes.results,
			debug: request.simOptions?.debug,
		}),
	);

	if (combiResult.error) {
		return makeAndSendRaidSimError(combiResult.error, onProgress);
	}

	simRes.progressMetricsFinal.finalRaidResult = combiResult;
	onProgress(simRes.progressMetricsFinal);

	return combiResult;
};

interface QueuedSim {
	request: RaidSimRequest;
	onProgress: WorkerProgressCallback;
	/** The sim's split results in split order, uncombined. Runs before the sim's final progress. */
	onResult: (splits: RaidSimResult[]) => void;
}

/**
 * Runs the sims, each split across the workers as runConcurrentSim splits it, as one queue of splits: a worker that hands
 * back a split takes the next one, of the same sim or the next, so no worker waits for the slowest split of a sim. Each
 * worker of the pool, as many as it has when the queue starts, takes its own splits one at a time, so no worker runs two
 * at once: a worker running two holds every call to it, an abort included, until one of them ends. When the pool shrinks
 * meanwhile, the splits of a worker it dropped go to the least busy one. Each sim's split results go to its onResult
 * uncombined, for a caller that reads only part of a result to put that part together itself. Returns the first error,
 * preferring a failure to the aborts it sets off.
 */
export const runConcurrentSimQueue = async (sims: QueuedSim[], workerPool: WorkerPool, signals: SimSignals): Promise<ErrorOutcome | undefined> => {
	const splitCount = workerPool.getNumWorkers();
	// Each sim's split requests, until its last split goes out.
	const splits: Array<Promise<RaidSimRequestSplitResult> | undefined> = [];
	const splitOf = (sim: number) =>
		(splits[sim] ??= workerPool.raidSimRequestSplit(RaidSimRequestSplitRequest.create({ splitCount, request: sims[sim].request })));
	const runs = sims.map(() => ({ results: [] as RaidSimResult[], completed: [] as number[], returned: 0, splitsDone: 0 }));

	const errors: ErrorOutcome[] = [];
	const fail = (error: ErrorOutcome) => {
		errors.push(error);
		signals.abort.trigger();
	};
	const failWith = (error: unknown) => fail(ErrorOutcome.create({ message: error instanceof Error ? error.message : String(error) }));

	let nextSim = 0;
	let nextSplit = 0;
	// The next split in queue order. A worker that waited for a sim's split looks again, as other workers may have taken
	// that sim's splits meanwhile.
	const claim = async (): Promise<{ sim: number; split: number; request: RaidSimRequest } | null> => {
		while (nextSim < sims.length) {
			const sim = nextSim;
			const splitResult = await splitOf(sim);
			if (splitResult.errorResult) throw new Error(splitResult.errorResult);
			if (sim !== nextSim) continue;
			runs[sim].splitsDone = splitResult.splitsDone;
			const split = nextSplit++;
			if (nextSplit === splitResult.splitsDone) {
				splits[sim] = undefined;
				nextSim++;
				nextSplit = 0;
			}
			// The next sim's split is asked for as this one starts, so it is back before a worker frees up for it.
			if (split === 0 && sim + 1 < sims.length) splitOf(sim + 1).catch(noop);
			return { sim, split, request: splitResult.requests[split] };
		}
		return null;
	};

	let progressCounter = 0;
	const finish = (sim: number) => {
		const { results } = runs[sim];
		runs[sim].results = [];
		const iterations = sims[sim].request.simOptions!.iterations;
		sims[sim].onResult(results);
		sims[sim].onProgress(ProgressMetrics.create({ totalIterations: iterations, completedIterations: iterations }));
	};

	// Checked again once a split is claimed: an abort sent before its sim starts never reaches the sim.
	const stopped = () => {
		if (errors.length) return true;
		if (!signals.abort.isTriggered()) return false;
		fail(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }));
		return true;
	};

	const worker = async (workerIndex: number) => {
		while (!stopped()) {
			const claimed = await claim();
			if (!claimed || stopped()) return;
			const { sim, split, request } = claimed;
			const run = runs[sim];
			const onProgress = (pm: ProgressMetrics) => {
				run.completed[split] = pm.completedIterations;
				// As often as runSims reports a sim, once a message from each worker.
				if (++progressCounter % splitCount) return;
				const completedIterations = sum(run.completed);
				sims[sim].onProgress(ProgressMetrics.create({ totalIterations: sims[sim].request.simOptions!.iterations, completedIterations }));
			};
			const result = await workerPool.raidSimAsync(request, onProgress, signals, workerIndex);
			if (result.error) return fail(result.error);
			run.results[split] = result;
			if (++run.returned === run.splitsDone) finish(sim);
		}
	};

	await Promise.all(Array.from({ length: splitCount }, (_, workerIndex) => worker(workerIndex).catch(failWith)));
	return errors.find(error => error.type !== ErrorOutcomeType.ErrorOutcomeAborted) ?? errors[0];
};
