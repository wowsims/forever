import {
	ComputeStatsResult,
	DistributionMetrics,
	type ErrorOutcome,
	PlayerStats,
	type ProgressMetrics,
	type Raid as RaidProto,
	RaidSimRequest,
	RaidSimResult,
} from '@generated/proto/api';
import { Class, Spec, Stat } from '@generated/proto/common';
import type { MetricSummary, StatSweepPoint } from '@sim/cache/stat_sweep_history';
import { Stats, UnitStat } from '@sim/proto/stats';
import { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { SimRuns } from '@sim/sim_runs';
import { createSimStore, patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, type StatSweepMetric, zeroVersions } from '@sim/state/sim_store';

import type { StatSweepSeries } from '../../../features/stat-sweep/components/StatSweepChart';
import type { StatSweepHistoryEntry } from '../../../features/stat-sweep/components/StatSweepHistory';
import { buildSweepOffsets, type MetricSample, type StatSweepSample, type SweepBatch, sweepSteps } from '../../../features/stat-sweep/model/stat_sweep';

const HIT_CAP = 150;
const DPS_PER_POINT = 2.1;

// DPS climbs until the cap and then flattens.
const dpsAt = (value: number) => 1180 + DPS_PER_POINT * Math.min(value, HIT_CAP) - 0.004 * (Math.min(value, HIT_CAP) - 80) ** 2;

type MetricsAt = (value: number, samples: number) => Record<StatSweepMetric, MetricSummary>;

const pointsFrom = (base: number, samples: number, metricsAt: MetricsAt): StatSweepPoint[] => {
	const offsets = buildSweepOffsets(100, 10);
	return sweepSteps(
		UnitStat.fromStat(Stat.StatHitRating),
		offsets,
		offsets.map(offset => base + offset),
	).map(step => ({ ...step, metrics: metricsAt(step.statValue, samples) }));
};

// A damage dealer: threat follows damage, and nothing is healed or taken.
const dealerMetrics =
	(dps: (value: number) => number): MetricsAt =>
	(value, n) => ({
		dps: { avg: dps(value), stdev: 450, n },
		hps: { avg: 0, stdev: 0, n },
		tps: { avg: dps(value) * 0.71, stdev: 320, n },
		dtps: { avg: 0, stdev: 0, n },
		tmi: { avg: 0, stdev: 0, n },
	});

// A tank whose damage taken and TMI fall as the stat rises.
const tankMetrics: MetricsAt = (value, n) => ({
	dps: { avg: 640 + 0.2 * value, stdev: 90, n },
	hps: { avg: 0, stdev: 0, n },
	tps: { avg: 1450 + 0.35 * value, stdev: 160, n },
	dtps: { avg: 1250 - 1.6 * value, stdev: 210, n },
	tmi: { avg: 18 - 0.03 * value, stdev: 4.5, n },
});

/** A ±100 hit sweep in steps of 10 from `base`, against a hit cap at 150. */
export const makeHitSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] => pointsFrom(base, samples, dealerMetrics(dpsAt));

/** A ±100 crit sweep in steps of 10 from `base`, with no cap in reach. */
export const makeCritSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] =>
	pointsFrom(
		base,
		samples,
		dealerMetrics(value => 1300 + 1.1 * value),
	);

/** A ±100 defense sweep in steps of 10 from `base`, for a tank. */
export const makeDefenseSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] => pointsFrom(base, samples, tankMetrics);

export const sweepSeries = (
	id: string,
	statLabel: string,
	base: number,
	points: StatSweepPoint[],
	colorSlot = 0,
	label = `${statLabel} ±100`,
): StatSweepSeries => ({ id, label, colorSlot, statLabel, baseStatValue: base, points });

const historyEntry = (
	id: string,
	statLabel: string,
	createdAt: number,
	overlaid: boolean,
	colorSlot?: number,
	metric: StatSweepMetric = 'dps',
): StatSweepHistoryEntry => ({ id, label: `${statLabel} ±40`, statLabel, createdAt, overlaid, colorSlot, metric });

/** Newest first: a four-stat sweep drawn whole, a one-stat sweep, and an older TPS sweep with one of its two runs drawn. */
export const SWEEP_HISTORY_BATCHES: SweepBatch<StatSweepHistoryEntry>[] = [
	{
		id: 'melee',
		runs: [
			historyEntry('crit', 'Crit Rating', Date.UTC(2026, 9, 13, 14, 3), true, 0),
			historyEntry('haste', 'Haste Rating', Date.UTC(2026, 9, 13, 14, 5), true, 1),
			historyEntry('arp', 'Armor Penetration', Date.UTC(2026, 9, 13, 14, 7), true, 2),
			historyEntry('expertise', 'Expertise Rating', Date.UTC(2026, 9, 13, 14, 9), true, 3),
		],
	},
	{ id: 'hit', runs: [historyEntry('hit', 'Hit Rating', Date.UTC(2026, 9, 12, 21, 40), false)] },
	{
		id: 'threat',
		runs: [
			historyEntry('strength', 'Strength', Date.UTC(2026, 9, 10, 18, 15), true, 4, 'tps'),
			historyEntry('agility', 'Agility', Date.UTC(2026, 9, 10, 18, 17), false, undefined, 'tps'),
		],
	},
];

const flatSample = (avg: number, stdev: number, n: number): MetricSample => ({ avg, stdev, n, sumSq: n * (stdev ** 2 + avg ** 2) });

/** One sweep step's metrics over `iterations`, each moving with `dps`; TPS has the widest spread. */
export const makeSweepSample = (iterations: number, dps = 1000, tmi = dps / 100): StatSweepSample => ({
	dps: flatSample(dps, 100, iterations),
	hps: flatSample(10, 5, iterations),
	tps: flatSample(2 * dps, 400, iterations),
	dtps: flatSample(3000 - dps, 300, iterations),
	tmi: flatSample(tmi, 2, iterations),
});

/** One pass of a sweep step, as the fake sim reads it off the request. */
export interface SweepPass {
	iterations: number;
	seed: number;
	/** The first player's bonus stats, the step's offset included. */
	bonusStats: Stats;
	onProgress: (progress: ProgressMetrics) => void;
}

/** One DPS per bonus point of Strength or Agility, so every step's DPS differs. */
export const sweepStepDps = ({ bonusStats }: SweepPass) => 1000 + bonusStats.getStat(Stat.StatStrength) + bonusStats.getStat(Stat.StatAgility);

export interface StatSweepHostParts {
	simulatePass?: (pass: SweepPass) => Promise<StatSweepSample | ErrorOutcome>;
	abortType?: (typeMask: number) => Promise<void>;
	simDisabled?: boolean;
	showPowerTools?: boolean;
}

export interface StatSweepHostFixture {
	host: IndividualSimHost<any>;
	store: SimStore;
	settings: StatSweepSettings;
}

const firstPlayerBonus = (raid: RaidProto | undefined) => Stats.fromProto(raid!.parties[0].players[0].bonusStats);

const metricOf = ({ avg, stdev, n, sumSq }: MetricSample) => DistributionMetrics.create({ avg, stdev, aggregatorData: { n, sumSq } });

const resultOf = ({ dps, hps, tps, dtps, tmi }: StatSweepSample) =>
	RaidSimResult.create({
		raidMetrics: { parties: [{ players: [{ dps: metricOf(dps), hps: metricOf(hps), threat: metricOf(tps), dtps: metricOf(dtps), tmi: metricOf(tmi) }] }] },
	});

/**
 * A host for the stat sweep: a warrior weighing Strength and Agility with no stats of its own, so a step's totals are its
 * bonus, on a sim that draws seed 7 and runs `simulatePass` for each pass, by default at `sweepStepDps`.
 */
export const makeStatSweepHost = ({ simulatePass, abortType, simDisabled = false, showPowerTools = false }: StatSweepHostParts = {}): StatSweepHostFixture => {
	const store = createSimStore();
	patchSlice(store, 'ui', { showPowerTools });
	const storage = new Map<string, string>();
	const env = { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } };
	seedKeyed(store, 'players', 0, { v: zeroVersions(PLAYER_FIELDS) } as never);
	const simulate = simulatePass ?? (async (pass: SweepPass) => makeSweepSample(pass.iterations, sweepStepDps(pass)));
	const sim = {
		store,
		env,
		runs: new SimRuns(store, { abortType: abortType ?? (async () => {}) }),
		waitForInit: () => Promise.resolve(),
		getIterations: () => 1000,
		getShowDamageMetrics: () => store.getState().ui.showDamageMetrics,
		getShowHealingMetrics: () => store.getState().ui.showHealingMetrics,
		getShowThreatMetrics: () => store.getState().ui.showThreatMetrics,
		makeRaidSimRequest: () =>
			RaidSimRequest.create({
				raid: { parties: [{ players: [{ bonusStats: new Stats().toProto() }] }] },
				simOptions: { iterations: 1000, randomSeed: 7n },
			}),
		computeRaidStats: async (raid: RaidProto) =>
			ComputeStatsResult.create({ raidStats: { parties: [{ players: [{ finalStats: firstPlayerBonus(raid).toProto() }] }] } }),
		runRaidSimRequest: async ({ raid, simOptions }: RaidSimRequest, _type: number, onProgress: (progress: ProgressMetrics) => void) => {
			const sample = await simulate({
				iterations: simOptions!.iterations,
				seed: Number(simOptions!.randomSeed),
				bonusStats: firstPlayerBonus(raid),
				onProgress,
			});
			return 'dps' in sample ? resultOf(sample) : RaidSimResult.create({ error: sample });
		},
	};
	const player = {
		storeKey: 0,
		sim,
		getClass: () => Class.ClassWarrior,
		getSpec: () => Spec.SpecUnknown,
		getCurrentStats: () => PlayerStats.create(),
	};
	const settings = new StatSweepSettings(player as never, 'stat-sweep-fixture');
	const host = {
		player,
		sim,
		simDisabled,
		individualConfig: { epStats: [Stat.StatStrength, Stat.StatAgility], epPseudoStats: [], epReferenceStat: Stat.StatStrength },
	};
	return { host: host as never, store, settings };
};
