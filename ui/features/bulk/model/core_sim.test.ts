import { BulkSimStage, ProgressMetrics } from '@generated/proto/api';
import { describe, expect, it, vi } from 'vitest';

import { type CoreBulkSimContext, runCoreBulkSim } from './core_sim';

const RESULT = { baseline: { dpsMetrics: { avg: 1000, stdev: 1 } }, topResults: [], stageMetrics: [] };

const context = (runBulkSim: ReturnType<typeof vi.fn>, setSimProgress = vi.fn()) =>
	({
		simUI: { sim: { runBulkSim, db: { lookupEquipmentSpec: vi.fn() } } },
		throwIfBulkAborted: () => {},
		runWithBulkAbort: <T>(promise: Promise<T>) => promise,
		setSimProgress,
		debugOptimisationRound: () => {},
	}) as unknown as CoreBulkSimContext;

describe('runCoreBulkSim', () => {
	it('passes the encounter override to the sim', async () => {
		const runBulkSim = vi.fn(async (..._args: unknown[]) => RESULT);
		const override = { duration: 90, durationVariation: 10 };

		await runCoreBulkSim(context(runBulkSim), new AbortController().signal, {} as never, { encounterOverride: override });

		expect(runBulkSim.mock.calls[0][3]).toEqual(override);
	});

	it('decorates every stage title', async () => {
		const setSimProgress = vi.fn();
		const runBulkSim = vi.fn(async (onProgress: (progress: ProgressMetrics) => void) => {
			onProgress(ProgressMetrics.create({ totalIterations: 10, bulkStage: BulkSimStage.BulkSimStageLow }));
			onProgress(ProgressMetrics.create({ totalIterations: 10, bulkStage: BulkSimStage.BulkSimStageMedium }));
			return RESULT;
		});

		await runCoreBulkSim(context(runBulkSim, setSimProgress), new AbortController().signal, {} as never, {
			decorateTitle: title => `D2 · ${title}`,
		});

		expect(setSimProgress.mock.calls).toHaveLength(2);
		for (const call of setSimProgress.mock.calls) expect(call[1].title).toMatch(/^D2 · /);
	});
});
