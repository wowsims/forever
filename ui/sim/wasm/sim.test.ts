import {
	ErrorOutcome,
	ErrorOutcomeType,
	RaidSimRequest,
	RaidSimRequestSplitRequest,
	RaidSimRequestSplitResult,
	RaidSimResult,
	SimOptions,
} from '@generated/proto/api';
import { describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../sim_signal_manager';
import { noop } from '../utils/misc';
import type { WorkerPool } from '../workers/worker_pool';
import { runConcurrentSimQueue } from './sim';

type QueuedSim = Parameters<typeof runConcurrentSimQueue>[0][number];

const makeSignals = () => {
	let triggered = false;
	return {
		abort: {
			isTriggered: () => triggered,
			trigger: async () => {
				triggered = true;
			},
		},
	} as unknown as SimSignals;
};

// Splits a request in two halves, each under the request's id and its index, as the worker pool does.
const splitInTwo = async ({ request }: RaidSimRequestSplitRequest) =>
	RaidSimRequestSplitResult.create({
		splitsDone: 2,
		requests: [0, 1].map(i =>
			RaidSimRequest.create({
				...request,
				requestId: `${request!.requestId}-${i}`,
				simOptions: SimOptions.create({ iterations: request!.simOptions!.iterations / 2 }),
			}),
		),
	});

// A pool of two workers whose splits end when the test ends them, in the order it chooses.
const makeHeldPool = () => {
	const held: Array<{ request: RaidSimRequest; workerIndex?: number; ended: boolean; end: (result: RaidSimResult) => void }> = [];
	const pool = {
		getNumWorkers: () => 2,
		raidSimRequestSplit: vi.fn(splitInTwo),
		raidSimAsync: vi.fn(
			(request: RaidSimRequest, _onProgress: unknown, _signals: unknown, workerIndex?: number) =>
				new Promise<RaidSimResult>(resolve => {
					const split = {
						request,
						workerIndex,
						ended: false,
						end: (result: RaidSimResult) => {
							split.ended = true;
							resolve(result);
						},
					};
					held.push(split);
				}),
		),
		raidSimResultCombination: vi.fn(),
	} as unknown as WorkerPool;
	return { pool, held };
};

const queuedSim = (iterations: number, requestId: string) => ({
	request: RaidSimRequest.create({ requestId, simOptions: SimOptions.create({ iterations }) }),
	onProgress: vi.fn(),
	onResult: vi.fn(),
});

describe('runConcurrentSimQueue', () => {
	it('keeps one split on each worker and sends the next, of the same sim or the next, as one comes back', async () => {
		const { pool, held } = makeHeldPool();
		const sims: QueuedSim[] = [queuedSim(100, 'a'), queuedSim(60, 'b')];

		const queue = runConcurrentSimQueue(sims, pool, makeSignals());
		await vi.waitFor(() => expect(held).toHaveLength(2));
		expect(held.map(({ request }) => request.requestId)).toEqual(['a-0', 'a-1']);

		held[1].end(RaidSimResult.create({ iterationsDone: 50 }));
		await vi.waitFor(() => expect(held).toHaveLength(3));
		expect(held[2].request.requestId).toBe('b-0');

		held[0].end(RaidSimResult.create({ iterationsDone: 50 }));
		await vi.waitFor(() => expect(held).toHaveLength(4));
		held[3].end(RaidSimResult.create({ iterationsDone: 30 }));
		held[2].end(RaidSimResult.create({ iterationsDone: 30 }));

		expect(await queue).toBeUndefined();
		expect(vi.mocked(pool.raidSimAsync).mock.calls).toHaveLength(4);
	});

	it('sends no worker a split while one runs on it, and the next to the worker that came back', async () => {
		const { pool, held } = makeHeldPool();
		const running = () => held.filter(split => !split.ended).map(split => split.workerIndex);

		const queue = runConcurrentSimQueue([queuedSim(100, 'a'), queuedSim(60, 'b'), queuedSim(40, 'c')], pool, makeSignals());
		await vi.waitFor(() => expect(held).toHaveLength(2));
		expect(running().sort()).toEqual([0, 1]);

		for (const freed of [1, 0, 3, 2, 4, 5]) {
			const sent = held.length;
			held[freed].end(RaidSimResult.create({ iterationsDone: 1 }));
			if (sent === 6) continue;
			await vi.waitFor(() => expect(held).toHaveLength(sent + 1));
			expect(held[sent].workerIndex).toBe(held[freed].workerIndex);
			expect(new Set(running()).size).toBe(running().length);
		}

		expect(await queue).toBeUndefined();
		expect(held.map(({ request }) => request.requestId)).toEqual(['a-0', 'a-1', 'b-0', 'b-1', 'c-0', 'c-1']);
	});

	it('sends no split claimed while an abort lands, and reports the abort', async () => {
		const { pool } = makeHeldPool();
		const signals = makeSignals();
		let returnSplit: () => void = noop;
		vi.mocked(pool.raidSimRequestSplit).mockImplementation(request => new Promise(resolve => (returnSplit = () => resolve(splitInTwo(request)))));

		const queue = runConcurrentSimQueue([queuedSim(100, 'a')], pool, signals);
		await vi.waitFor(() => expect(pool.raidSimRequestSplit).toHaveBeenCalled());
		await signals.abort.trigger();
		returnSplit();

		expect(await queue).toEqual(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }));
		expect(pool.raidSimAsync).not.toHaveBeenCalled();
	});

	it('hands each sim its split results in split order, uncombined, with a final progress', async () => {
		const { pool, held } = makeHeldPool();
		const sim = queuedSim(100, 'a');

		const queue = runConcurrentSimQueue([sim], pool, makeSignals());
		await vi.waitFor(() => expect(held).toHaveLength(2));
		held[1].end(RaidSimResult.create({ iterationsDone: 49 }));
		held[0].end(RaidSimResult.create({ iterationsDone: 51 }));
		await queue;

		expect(sim.onResult).toHaveBeenCalledWith([RaidSimResult.create({ iterationsDone: 51 }), RaidSimResult.create({ iterationsDone: 49 })]);
		expect(sim.onProgress.mock.calls.at(-1)![0].completedIterations).toBe(100);
		expect(sim.onResult.mock.invocationCallOrder[0]).toBeLessThan(sim.onProgress.mock.invocationCallOrder.at(-1)!);
		expect(pool.raidSimResultCombination).not.toHaveBeenCalled();
	});

	it('stops at a failure and reports it over the aborts it sets off', async () => {
		const { pool, held } = makeHeldPool();
		const signals = makeSignals();
		const failure = ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeError, message: 'failed' });

		const queue = runConcurrentSimQueue([queuedSim(100, 'a'), queuedSim(100, 'b')], pool, signals);
		await vi.waitFor(() => expect(held).toHaveLength(2));
		held[1].end(RaidSimResult.create({ error: failure }));
		held[0].end(RaidSimResult.create({ error: ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }) }));

		expect(await queue).toEqual(failure);
		expect(signals.abort.isTriggered()).toBe(true);
		expect(vi.mocked(pool.raidSimAsync).mock.calls).toHaveLength(2);
		expect(pool.raidSimResultCombination).not.toHaveBeenCalled();
	});
});
