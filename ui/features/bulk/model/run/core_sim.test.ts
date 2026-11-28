import { BulkSimStage } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Encounter, UnitStats } from '@generated/proto/common';
import { Gear } from '@sim/proto/gear';
import { describe, expect, it, vi } from 'vitest';

import { type CoreBulkSimContext, runCoreBulkSim } from './core_sim';

const RESULT = { baseline: { dpsMetrics: { avg: 1000, stdev: 1 } }, topResults: [], culledResults: [], stageMetrics: [] };

const context = (runBulkSim: ReturnType<typeof vi.fn>, setSimProgress = vi.fn()) =>
	({
		simUI: { sim: { runBulkSim, db: { lookupEquipmentSpec: vi.fn() } } },
		throwIfBulkAborted: () => {},
		runWithBulkAbort: <T>(promise: Promise<T>) => promise,
		setSimProgress,
		debugOptimisationRound: () => {},
	}) as unknown as CoreBulkSimContext;

describe('runCoreBulkSim', () => {
	it('passes the encounter to the sim as its override', async () => {
		const runBulkSim = vi.fn(async (..._args: unknown[]) => RESULT);
		const signal = new AbortController().signal;
		const encounter = Encounter.create({ duration: 90, durationVariation: 10 });

		await runCoreBulkSim(context(runBulkSim), signal, {} as never, { encounter });

		expect(runBulkSim.mock.calls[0].slice(2)).toEqual([signal, { encounter }]);
	});

	it('passes the build override to the sim', async () => {
		const runBulkSim = vi.fn(async (..._args: unknown[]) => RESULT);
		const signal = new AbortController().signal;
		const encounter = Encounter.create();
		const build = { talentsString: '05-3', rotation: APLRotation.create({ type: APLRotationType.TypeAuto }), bonusStats: UnitStats.create() };

		await runCoreBulkSim(context(runBulkSim), signal, {} as never, { encounter, build });

		expect(runBulkSim.mock.calls[0].slice(2)).toEqual([signal, { encounter, build }]);
	});

	it('passes the candidates to the sim as they are', async () => {
		const runBulkSim = vi.fn(async (..._args: unknown[]) => RESULT);
		const candidates = [new Gear({})];

		await runCoreBulkSim(context(runBulkSim), new AbortController().signal, {} as never, { encounter: Encounter.create(), candidates });

		expect((runBulkSim.mock.calls[0][3] as { candidates: Gear[] }).candidates).toBe(candidates);
	});

	it('appends the culled rows after the top results, with their stage and no rank', async () => {
		const row = (avg: number, fields: object = {}) => ({ gear: { items: [] }, dpsMetrics: { avg, stdev: 1 }, ...fields });
		const runBulkSim = vi.fn(async (..._args: unknown[]) => ({
			...RESULT,
			topResults: [
				row(1100, { pairedErrorToNextResult: 2, pairedErrorToBaseline: 3, pairedDpsDiffToBaseline: 100 }),
				row(1050, { pairedErrorToBaseline: 4, pairedDpsDiffToBaseline: 50 }),
			],
			culledResults: [row(990, { pairedErrorToBaseline: 5, pairedDpsDiffToBaseline: -12, culledStage: BulkSimStage.BulkSimStageLow })],
		}));

		const { topGearResults } = await runCoreBulkSim(context(runBulkSim), new AbortController().signal, {} as never, { encounter: Encounter.create() });

		expect(topGearResults.map(({ gear: _gear, dpsMetrics, ...rest }) => [dpsMetrics.avg, rest])).toEqual([
			[1100, { backendRank: 0, pairedErrorToNextResult: 2, pairedErrorToBaseline: 3, pairedDpsDiffToBaseline: 100 }],
			[1050, { backendRank: 1, pairedErrorToNextResult: undefined, pairedErrorToBaseline: 4, pairedDpsDiffToBaseline: 50 }],
			[990, { culledStage: BulkSimStage.BulkSimStageLow, pairedErrorToBaseline: 5, pairedDpsDiffToBaseline: -12 }],
		]);
		expect(Object.keys(topGearResults[2])).not.toContain('backendRank');
		expect(Object.keys(topGearResults[2])).not.toContain('pairedErrorToNextResult');
	});
});
