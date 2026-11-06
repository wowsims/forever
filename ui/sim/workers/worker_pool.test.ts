import {
	ComputeStatsRequest,
	ComputeStatsResult,
	ProgressMetrics,
	RaidSimRequest,
	RaidSimRequestSplitRequest,
	RaidSimRequestSplitResult,
	RaidSimResult,
	SimOptions,
} from '@generated/proto/api';
import type { WorkerReceiveMessage, WorkerSendMessage } from '@worker/types';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../sim_signal_manager';
import { noop } from '../utils/misc';
import { WorkerPool } from './worker_pool';

// A sim worker that is ready at once and answers every raid sim with its final result, or every request the
// way `answer` says.
class FakeWorker {
	static instances: FakeWorker[] = [];
	static answerRaidSims = true;
	static answer: (message: WorkerReceiveMessage) => WorkerSendMessage | undefined = () => undefined;
	readonly posted: WorkerReceiveMessage[] = [];
	terminated = false;
	private listener: (event: { data: WorkerSendMessage }) => void = noop;

	constructor() {
		FakeWorker.instances.push(this);
	}

	addEventListener(_type: 'message', listener: (event: { data: WorkerSendMessage }) => void) {
		this.listener = listener;
		queueMicrotask(() => listener({ data: { msg: 'ready', outputData: new Uint8Array([1]) } as WorkerSendMessage }));
	}

	postMessage(message: WorkerReceiveMessage) {
		this.posted.push(message);
		const answer = FakeWorker.answer(message);
		if (answer) {
			queueMicrotask(() => this.listener({ data: answer }));
			return;
		}
		if (message.msg !== 'raidSimAsync' || !FakeWorker.answerRaidSims) return;
		const outputData = ProgressMetrics.toBinary(ProgressMetrics.create({ finalRaidResult: RaidSimResult.create({ iterationsDone: 1 }) }));
		queueMicrotask(() => this.listener({ data: { msg: 'raidSimAsync', id: message.id, outputData } as WorkerSendMessage }));
	}

	terminate() {
		this.terminated = true;
	}
}

const signals = { abort: { onTrigger: () => noop } } as unknown as SimSignals;

const raidSimsPosted = () => FakeWorker.instances.map(worker => worker.posted.filter(message => message.msg === 'raidSimAsync').length);

const request = (requestId: string) => RaidSimRequest.create({ requestId, simOptions: SimOptions.create({ iterations: 1 }) });

beforeEach(() => {
	FakeWorker.instances = [];
	FakeWorker.answerRaidSims = true;
	FakeWorker.answer = () => undefined;
	vi.stubGlobal('Worker', FakeWorker);
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('WorkerPool.raidSimAsync', () => {
	it('runs on the worker at the index it is given, whichever worker is least busy', async () => {
		const pool = new WorkerPool(3);

		const result = await pool.raidSimAsync(request('pinned'), noop, signals, 2);

		expect(result.iterationsDone).toBe(1);
		expect(raidSimsPosted()).toEqual([0, 0, 1]);
	});

	it('runs on the least busy worker when the pool no longer has the index', async () => {
		const pool = new WorkerPool(3);
		pool.setNumWorkers(2);

		await pool.raidSimAsync(request('dropped'), noop, signals, 2);

		expect(raidSimsPosted()).toEqual([1, 0, 0]);
	});
});

describe('WorkerPool.raidSimRequestSplit', () => {
	const splitOf = async (requestId: string) => {
		const pool = new WorkerPool(1);
		// The sim clones the request into each split, its id included.
		const split = request(requestId);
		const splits = RaidSimRequestSplitResult.create({ splitsDone: 2, requests: [split, RaidSimRequest.clone(split)] });
		vi.spyOn(pool, 'makeApiCall').mockResolvedValue(RaidSimRequestSplitResult.toBinary(splits));
		const result = await pool.raidSimRequestSplit(RaidSimRequestSplitRequest.create({ splitCount: 2, request: split }));
		return result.requests.map(split => split.requestId);
	};

	it('gives each split its own id, as one worker can be handed two splits of a sim', async () => {
		expect(await splitOf('raidSimAsync-1')).toEqual(['raidSimAsync-1-0', 'raidSimAsync-1-1']);
	});

	it('leaves the splits of a request with no id without one', async () => {
		expect(await splitOf('')).toEqual(['', '']);
	});
});

describe('WorkerPool on a worker whose sim runtime has stopped', () => {
	// The worker answers one computeStats the way WorkerInterface reports a stopped runtime, and the
	// raid sim it was handed before that stays unanswered.
	const stoppedAnswer = (message: WorkerReceiveMessage): WorkerSendMessage | undefined =>
		message.msg === 'computeStats'
			? ({ msg: 'computeStats', id: message.id, outputData: new Uint8Array(), error: 'stopped', stopped: true } as WorkerSendMessage)
			: undefined;

	it('fails the request and every request waiting on the worker, then runs the next one on a fresh worker', async () => {
		FakeWorker.answer = stoppedAnswer;
		const pool = new WorkerPool(1);
		const [first] = FakeWorker.instances;
		const raidSim = pool.raidSimAsync(request('waiting'), noop, signals);
		FakeWorker.answerRaidSims = false;

		await expect(pool.computeStats(ComputeStatsRequest.create())).rejects.toThrow('stopped');
		await expect(raidSim).rejects.toThrow('stopped before answering');
		expect(first.terminated).toBe(true);
		expect(FakeWorker.instances).toHaveLength(2);

		FakeWorker.answer = message =>
			message.msg === 'computeStats'
				? ({ msg: 'computeStats', id: message.id, outputData: ComputeStatsResult.toBinary(ComputeStatsResult.create()) } as WorkerSendMessage)
				: undefined;
		await expect(pool.computeStats(ComputeStatsRequest.create())).resolves.toBeDefined();
		expect(FakeWorker.instances[1].posted.filter(message => message.msg === 'computeStats')).toHaveLength(1);
	});

	it('treats an answer with no result as a stopped runtime', async () => {
		FakeWorker.answer = message => (message.msg === 'computeStats' ? ({ msg: 'computeStats', id: message.id } as WorkerSendMessage) : undefined);
		const pool = new WorkerPool(1);

		await expect(pool.computeStats(ComputeStatsRequest.create())).rejects.toThrow('no result');
		expect(FakeWorker.instances).toHaveLength(2);
	});
});
