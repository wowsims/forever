import { ProgressMetrics, RaidSimRequest, RaidSimRequestSplitRequest, RaidSimRequestSplitResult, RaidSimResult, SimOptions } from '@generated/proto/api';
import type { WorkerReceiveMessage, WorkerSendMessage } from '@worker/types';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../sim_signal_manager';
import { noop } from '../utils/misc';
import { WorkerPool } from './worker_pool';

// A sim worker that is ready at once and answers every raid sim with its final result.
class FakeWorker {
	static instances: FakeWorker[] = [];
	readonly posted: WorkerReceiveMessage[] = [];
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
		if (message.msg !== 'raidSimAsync') return;
		const outputData = ProgressMetrics.toBinary(ProgressMetrics.create({ finalRaidResult: RaidSimResult.create({ iterationsDone: 1 }) }));
		queueMicrotask(() => this.listener({ data: { msg: 'raidSimAsync', id: message.id, outputData } as WorkerSendMessage }));
	}

	terminate() {}
}

const signals = { abort: { onTrigger: () => noop } } as unknown as SimSignals;

const raidSimsPosted = () => FakeWorker.instances.map(worker => worker.posted.filter(message => message.msg === 'raidSimAsync').length);

const request = (requestId: string) => RaidSimRequest.create({ requestId, simOptions: SimOptions.create({ iterations: 1 }) });

beforeEach(() => {
	FakeWorker.instances = [];
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
