import type { DistributionMetrics, Raid as RaidProto, RaidSimResult } from '@generated/proto/api';
import type { Spec, UnitStats } from '@generated/proto/common';
import i18n from '@i18n/config';
import { saveRun, type StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { Stats, type UnitStat } from '@sim/proto/stats';
import type { PrecisionSettings } from '@sim/settings/precision';
import { type Sim, SimError } from '@sim/sim';
import { RequestTypes } from '@sim/sim_signal_manager';
import { SimRunKind, type StatSweepMetric } from '@sim/state/sim_store';
import { randomUUID } from '@sim/utils/misc';
import { generateRequestId } from '@sim/workers/worker_pool';
import { toastManager } from '@ui-kit/Toast';
import { SimRequest } from '@worker/types';

import { SERIES_SLOT_COUNT } from './colors';
import {
	advanceSweepProgress,
	buildSweepOffsets,
	measureSweepStep,
	type MetricSample,
	runSweepStat,
	skipSweepSteps,
	startSweepProgress,
	type StatSweepProgress,
	type StatSweepSample,
	type StatSweepStep,
	type StatSweepStepSample,
	type SweepPassRunner,
	sweepSteps,
} from './stat_sweep';

export interface StatSweepDeps {
	sim: Sim;
	specId: Spec;
}

export interface StatSweepStatArgs {
	unitStat: UnitStat;
	label: string;
	/** The stat's current total. */
	baseStatValue: number;
}

export interface StatSweepArgs {
	/** Run one after another, each on its own and saved as its own run. */
	stats: StatSweepStatArgs[];
	range: number;
	step: number;
	/** Sizes Target ±. */
	metric: StatSweepMetric;
	precision: PrecisionSettings;
	/** The first stat's colour slot; each stat after it takes the next one. */
	firstColorSlot: number;
	/** Saved on every run of the sweep, with one batch id the sweep draws. */
	batchLabel: string;
}

export interface StatSweepCallbacks {
	onProgress: (progress: StatSweepProgress) => void;
	/** A stat's run as it starts and after each step. */
	onRunUpdate: (run: StatSweepRecord) => void;
	/** A stat's run once it ends; one with no finished step is not saved. */
	onRunEnd: (run: StatSweepRecord) => void;
}

const sampleOf = ({ avg, stdev, aggregatorData }: DistributionMetrics): MetricSample => ({ avg, stdev, n: aggregatorData!.n, sumSq: aggregatorData!.sumSq });

const sweepSampleOf = (result: RaidSimResult): StatSweepSample => {
	const { dps, hps, threat, dtps, tmi } = result.raidMetrics!.parties[0].players[0];
	return { dps: sampleOf(dps!), hps: sampleOf(hps!), tps: sampleOf(threat!), dtps: sampleOf(dtps!), tmi: sampleOf(tmi!) };
};

// A copy of `raid` that shares everything but the first player's bonus stats.
const withFirstPlayerBonus = (raid: RaidProto, bonusStats: UnitStats): RaidProto => {
	const [party, ...parties] = raid.parties;
	const [player, ...players] = party.players;
	return { ...raid, parties: [{ ...party, players: [{ ...player, bonusStats }, ...players] }, ...parties] };
};

/**
 * Sweeps each stat in turn, as one stat sweep run that a stat sweep abort or an abort of every request stops. Every pass
 * patches one request built from the setup as the sweep starts, and every step of every stat starts at its seed, so the
 * steps differ only by the stat. A stat's run is saved as it ends.
 */
export function runStatSweep({ sim, specId }: StatSweepDeps, args: StatSweepArgs, { onProgress, onRunUpdate, onRunEnd }: StatSweepCallbacks): Promise<void> {
	const { precision, metric } = args;
	const aborting = () => sim.runs.isAborting(SimRunKind.StatSweep);

	return sim.runs.start(SimRunKind.StatSweep, async () => {
		// Set when a step comes back aborted though the sweep itself was not cancelled, as when another run aborts every request.
		let stoppedElsewhere = false;
		try {
			await sim.waitForInit();
			const base = sim.makeRaidSimRequest();
			const seed = base.simOptions!.randomSeed;
			const ownBonus = Stats.fromProto(base.raid!.parties[0].players[0].bonusStats);
			const defaultIterations = sim.getIterations();
			const offsets = buildSweepOffsets(args.range, args.step);
			const batchId = randomUUID();
			const startedAt = Date.now();
			// Every stat at every offset until its totals are known, so the bar and the estimate only shrink as offsets drop out.
			let progress = startSweepProgress(args.stats.map(() => offsets.length));
			const report = (next: StatSweepProgress) => {
				progress = next;
				onProgress(next);
			};
			onProgress(progress);
			let stepsBefore = 0;
			// The last finished step's iterations: in Target ± mode a step's total is known only after its first pass, and
			// neighbouring steps need about as many, so the bar sizes a step by it until its own plan is larger.
			let lastStepIterations = 0;
			for (const [statIndex, { unitStat, label, baseStatValue }] of args.stats.entries()) {
				if (aborting()) break;
				const raidAt = (offset: number) => withFirstPlayerBonus(base.raid!, ownBonus.add(new Stats().withUnitStat(unitStat, offset)).toProto());
				const record: StatSweepRecord = {
					id: randomUUID(),
					createdAt: Date.now(),
					specId,
					label,
					colorSlot: (args.firstColorSlot + statIndex) % SERIES_SLOT_COUNT,
					unitStat: unitStat.toJson(),
					metric,
					baseStatValue,
					points: [],
					batchId,
					batchLabel: args.batchLabel,
				};
				onRunUpdate(record);
				let finished = record;
				let stopped = false;
				try {
					// The stat's totals as it starts, so one stat's stat computations are in flight at a time.
					const totals = await Promise.all(
						offsets.map(async offset => {
							if (offset === 0) return baseStatValue;
							const result = await sim.computeRaidStats(raidAt(offset));
							if (result.errorResult) throw new SimError(result.errorResult);
							return Stats.fromProto(result.raidStats!.parties[0].players[0].finalStats).getUnitStat(unitStat);
						}),
					);
					const steps = sweepSteps(offsets, totals);
					if (steps.length < offsets.length) report(skipSweepSteps(progress, statIndex, offsets.length - steps.length));
					const measure = async (step: StatSweepStep, index: number, confirmsCap: (firstPass: StatSweepStepSample['firstPass']) => boolean) => {
						if (aborting()) return null;
						const raid = raidAt(step.offset);
						let planned = 0;
						const runPass: SweepPassRunner = async (iterations, seedOffset) => {
							if (aborting()) return null;
							const result = await sim.runRaidSimRequest(
								{
									...base,
									requestId: generateRequestId(SimRequest.raidSimAsync),
									raid,
									simOptions: {
										...base.simOptions!,
										iterations,
										randomSeed: seed + BigInt(seedOffset),
										debugFirstIteration: false,
										useLabeledRands: true,
									},
								},
								RequestTypes.StatSweep,
								metrics =>
									report(
										advanceSweepProgress(
											progress,
											stepsBefore + index,
											seedOffset + metrics.completedIterations,
											Math.max(planned, lastStepIterations),
											(Date.now() - startedAt) / 1000,
										),
									),
							);
							if (!result.error) return sweepSampleOf(result);
							if (!aborting()) stoppedElsewhere = true;
							return null;
						};
						const sample = await measureSweepStep(
							runPass,
							precision,
							metric,
							defaultIterations,
							total => {
								planned = total;
							},
							confirmsCap,
						);
						if (sample) lastStepIterations = sample.metrics[metric].n;
						return sample;
					};
					const statRun = await runSweepStat(
						steps,
						measure,
						({ points, cappedStep }) => {
							finished = { ...record, points, ...(cappedStep && { cappedAt: cappedStep.offset }) };
							onRunUpdate(finished);
						},
						skipped => report(skipSweepSteps(progress, statIndex, skipped)),
					);
					stopped = statRun.stopped;
					stepsBefore += statRun.points.length;
				} finally {
					onRunEnd(finished);
					if (finished.points.length) saveRun(finished).catch(console.error);
				}
				if (stopped) break;
			}
		} catch (error: any) {
			console.error(error);
			toastManager.add({ variant: 'error', body: error?.message || 'Something went wrong running the stat sweep.' });
		}
		if (stoppedElsewhere) toastManager.add({ variant: 'info', body: i18n.t('sidebar.buttons.stat_sweep.stopped') });
	});
}
