import { SavedStatSweepSettings, StatSweepMetric as SavedMetric } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import type { Env } from '../state/env';
import { patchKeyed, seedKeyed, type SimStore, type StatSweepMetric, type StatSweepSlice } from '../state/sim_store';
import { subscribeStatSweepChange } from '../state/subscriptions';
import { arrayEquals, equalsOrBothNull, isUnchangedPatch } from '../utils/collections';
import { precisionModeFromProto, precisionModeToProto } from './precision';

export type StatSweepValues = Omit<StatSweepSlice, 'v'>;

const DEFAULTS: StatSweepValues = {
	stats: null,
	range: 100,
	step: 10,
	metric: 'dps',
	showSpread: true,
	precisionMode: 'iterations',
	iterations: 0,
	targetCi: 1,
	maxIterationMultiplier: 0,
};

const SAVED_METRICS: Record<StatSweepMetric, SavedMetric> = {
	dps: SavedMetric.StatSweepMetricDps,
	hps: SavedMetric.StatSweepMetricHps,
	tps: SavedMetric.StatSweepMetricTps,
	dtps: SavedMetric.StatSweepMetricDtps,
	tmi: SavedMetric.StatSweepMetricTmi,
};

const metricFromProto = (saved: SavedMetric): StatSweepMetric =>
	(Object.keys(SAVED_METRICS) as StatSweepMetric[]).find(metric => SAVED_METRICS[metric] === saved) ?? DEFAULTS.metric;

// Proto3 cannot tell an unset field from 0, or an unset list from an empty one, so a zero range, step or target and
// no stored stat read as the default.
const fromProto = (saved: SavedStatSweepSettings): StatSweepValues => {
	const stats = saved.stats.map(stat => UnitStat.fromProto(stat)).filter(stat => stat.isStat() || stat.isPseudoStat());
	return {
		stats: stats.length ? stats : DEFAULTS.stats,
		range: saved.range || DEFAULTS.range,
		step: saved.step || DEFAULTS.step,
		metric: metricFromProto(saved.metric),
		showSpread: !saved.hideSpread,
		precisionMode: precisionModeFromProto(saved.precisionMode),
		iterations: saved.iterations,
		targetCi: saved.targetCi || DEFAULTS.targetCi,
		maxIterationMultiplier: saved.maxIterationMultiplier,
	};
};

const sameStats = (a: UnitStat[] | null, b: UnitStat[] | null) => equalsOrBothNull(a, b, (a, b) => arrayEquals(a, b, (x, y) => x.equals(y)));

// The stat sweep dialog's stored setup. Values live in the sim store (`statSweep[player.storeKey]`) with a version
// counter, and every change is saved to localStorage.
export class StatSweepSettings {
	private readonly storageKey: string;
	private readonly env: Env;
	readonly store: SimStore;
	readonly storeKey: number;

	constructor(player: Player<any>, storageKey: string) {
		this.storageKey = storageKey;
		this.env = player.sim.env;
		this.store = player.sim.store;
		this.storeKey = player.storeKey;

		seedKeyed(this.store, 'statSweep', this.storeKey, { ...DEFAULTS, v: { settings: 0 } });
		subscribeStatSweepChange(this)(() => {
			this.env.storage.setItem(this.storageKey, SavedStatSweepSettings.toJsonString(this.toProto()));
		});
	}

	private write(patch: Partial<StatSweepValues>) {
		patchKeyed(this.store, 'statSweep', this.storeKey, patch, ['settings']);
	}

	/** The stored values; the dialog plots the stored metric only while the display settings offer it. */
	get values(): StatSweepValues {
		const { v: _, ...values } = this.store.getState().statSweep[this.storeKey];
		return values;
	}

	set(patch: Partial<StatSweepValues>) {
		if (!isUnchangedPatch(this.values, patch, { stats: sameStats })) this.write(patch);
	}

	applyDefaults() {
		this.write(DEFAULTS);
	}

	load() {
		const storageValue = this.env.storage.getItem(this.storageKey);
		if (storageValue) {
			this.write(fromProto(SavedStatSweepSettings.fromJsonString(storageValue, { ignoreUnknownFields: true })));
		}
	}

	toProto(): SavedStatSweepSettings {
		const { stats, range, step, metric, showSpread, precisionMode, iterations, targetCi, maxIterationMultiplier } = this.values;
		return SavedStatSweepSettings.create({
			apiVersion: CURRENT_API_VERSION,
			stats: stats?.map(stat => stat.toProto()) ?? [],
			range,
			step,
			metric: SAVED_METRICS[metric],
			hideSpread: !showSpread,
			precisionMode: precisionModeToProto(precisionMode),
			iterations,
			targetCi,
			maxIterationMultiplier,
		});
	}
}
