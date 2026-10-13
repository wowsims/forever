import { Encounter } from '@generated/proto/common';
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
		const override = Encounter.create({ duration: 90, durationVariation: 10, useHealth: true });

		await runCoreBulkSim(context(runBulkSim), new AbortController().signal, {} as never, { encounterOverride: override });

		expect(runBulkSim.mock.calls[0][3]).toEqual(override);
	});
});
