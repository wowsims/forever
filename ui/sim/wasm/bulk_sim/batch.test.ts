import { BulkSimRequest, BulkSimStage, RaidSimRequest } from '@generated/proto/api';
import { describe, expect, it } from 'vitest';

import type { SimSignals } from '../../sim_signal_manager';
import type { WorkerPool } from '../../workers/worker_pool';
import { runBulkSimCandidateBatchOnWorkers } from './batch';
import { makeBulkSimStageProgressEmitter } from './progress';

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

describe('runBulkSimCandidateBatchOnWorkers', () => {
	it('resolves a batch of no candidates at once', async () => {
		const emitter = makeBulkSimStageProgressEmitter(() => {}, BulkSimStage.BulkSimStageHigh, 0, 0);
		await expect(
			runBulkSimCandidateBatchOnWorkers(
				BulkSimRequest.create({ baseRequest: RaidSimRequest.create() }),
				[],
				100,
				{ getNumWorkers: () => 2 } as unknown as WorkerPool,
				makeSignals(),
				{ completedSimsBase: 0, completedIterationsBase: 0, emitter },
			),
		).resolves.toEqual([]);
	});
});
