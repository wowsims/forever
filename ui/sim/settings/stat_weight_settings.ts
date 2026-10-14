import { PseudoStat, Stat } from '@generated/proto/common';
import { SavedStatWeightSettings } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import type { Env } from '../state/env';
import { patchKeyed, seedKeyed, SimStore, StatWeightsSlice } from '../state/sim_store';
import { subscribeStatWeightsChange } from '../state/subscriptions';
import { isUnchangedPatch } from '../utils/collections';
import { precisionModeFromProto, precisionModeToProto, type PrecisionSettings } from './precision';

const PRECISION_DEFAULTS: PrecisionSettings = { precisionMode: 'iterations', iterations: 0, targetCi: 0.01, maxIterationMultiplier: 0 };

// Proto3 cannot tell an unset field from 0, so a zero target reads as the default.
const precisionFromProto = (saved: SavedStatWeightSettings): PrecisionSettings => ({
	precisionMode: precisionModeFromProto(saved.precisionMode),
	iterations: saved.iterations,
	targetCi: saved.targetCi || PRECISION_DEFAULTS.targetCi,
	maxIterationMultiplier: saved.maxIterationMultiplier,
});

// Stat-weight modal settings. Values live in the sim store
// (`statWeights[player.storeKey]`) with a version counter; persists itself to
// localStorage on every change.
export class StatWeightActionSettings {
	private readonly storageKey: string;
	private readonly env: Env;
	readonly store: SimStore;
	readonly storeKey: number;

	constructor(player: Player<any>, storageKey: string) {
		this.storageKey = storageKey;
		this.env = player.sim.env;
		this.store = player.sim.store;
		this.storeKey = player.storeKey;

		seedKeyed(this.store, 'statWeights', this.storeKey, {
			excludedStats: [],
			excludedPseudoStats: [],
			...PRECISION_DEFAULTS,
			v: { settings: 0 },
		});
		subscribeStatWeightsChange(this)(() => {
			this.env.storage.setItem(this.storageKey, SavedStatWeightSettings.toJsonString(this.toProto()));
		});
	}

	private get slice(): StatWeightsSlice {
		return this.store.getState().statWeights[this.storeKey];
	}

	// One store write per logical change.
	private write(patch: Partial<Omit<StatWeightsSlice, 'v'>>) {
		patchKeyed(this.store, 'statWeights', this.storeKey, patch, ['settings']);
	}

	get excludedStats(): Stat[] {
		return this.slice.excludedStats.slice();
	}

	get excludedPseudoStats(): PseudoStat[] {
		return this.slice.excludedPseudoStats.slice();
	}

	get precision(): PrecisionSettings {
		const { precisionMode, iterations, targetCi, maxIterationMultiplier } = this.slice;
		return { precisionMode, iterations, targetCi, maxIterationMultiplier };
	}

	setPrecision(patch: Partial<PrecisionSettings>) {
		if (!isUnchangedPatch(this.precision, patch)) this.write(patch);
	}

	applyDefaults() {
		this.write({ excludedStats: [], excludedPseudoStats: [], ...PRECISION_DEFAULTS });
	}

	load() {
		const storageValue = this.env.storage.getItem(this.storageKey);
		if (storageValue) {
			const settingsProto = SavedStatWeightSettings.fromJsonString(storageValue, { ignoreUnknownFields: true });
			this.write({
				excludedStats: settingsProto.excludedStats || [],
				excludedPseudoStats: settingsProto.excludedPseudoStats || [],
				...precisionFromProto(settingsProto),
			});
		}
	}

	toProto(): SavedStatWeightSettings {
		const { precisionMode, iterations, targetCi, maxIterationMultiplier } = this.precision;
		return SavedStatWeightSettings.create({
			apiVersion: CURRENT_API_VERSION,
			excludedStats: this.excludedStats,
			excludedPseudoStats: this.excludedPseudoStats,
			precisionMode: precisionModeToProto(precisionMode),
			iterations,
			targetCi,
			maxIterationMultiplier,
		});
	}

	/**
	 * Check if a stat should be excluded from weight calculation.
	 * @param stat
	 * @returns true if stat should be excluded.
	 */
	isStatExcludedFromCalc(stat: Stat): boolean {
		return !!this.excludedStats.includes(stat);
	}

	/**
	 * Check if a pseudostat should be excluded from weight calculation.
	 * @param pseudoStat
	 * @returns true if pseudostat should be excluded.
	 */
	isPseudoStatExcludedFromCalc(pseudoStat: PseudoStat): boolean {
		return !!this.excludedPseudoStats.includes(pseudoStat);
	}

	/**
	 * Check if a unitstat should be excluded from weight calculation.
	 * @param unitstat
	 * @returns true if unitstat should be excluded.
	 */
	isUnitStatExcludedFromCalc(unitstat: UnitStat): boolean {
		return unitstat.isStat() ? this.isStatExcludedFromCalc(unitstat.getStat()) : this.isPseudoStatExcludedFromCalc(unitstat.getPseudoStat());
	}

	/**
	 * Set whether a stat should be excluded from calculation.
	 * @param stat
	 * @param exclude
	 */
	setStatExcluded(stat: UnitStat, exclude: boolean) {
		const updateStatEntry = <T extends Stat | PseudoStat>(s: T, target: T[]) => {
			const currentIdx = target.indexOf(s);
			if (exclude) {
				if (currentIdx === -1) target.push(s);
			} else if (currentIdx !== -1) {
				target.splice(currentIdx, 1);
			}
			return target;
		};
		if (stat.isStat()) {
			this.write({ excludedStats: updateStatEntry(stat.getStat(), this.excludedStats) });
		} else {
			this.write({ excludedPseudoStats: updateStatEntry(stat.getPseudoStat(), this.excludedPseudoStats) });
		}
	}
}
