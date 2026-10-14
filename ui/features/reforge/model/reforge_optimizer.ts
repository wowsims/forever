// DOM-free half of the gear optimizer: settings access, EP / soft-cap math and the solve itself
// (cache lookup, sim request, abort). The rendering half lives in ../components/ReforgePanel and
// owns every button, tooltip, toast and modal.
//
// There is no item reforging: "reforge" is only the optimizer's name.
import { enchantPresetPool } from '@features/enchant-presets/model/enchant_presets';
import { ReforgeOptimizeRequest, ReforgeSettings } from '@generated/proto/api';
import { Profession, Spec, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { ReforgeGearCache } from '@sim/cache/reforge_cache';
import { Player } from '@sim/player/player';
import { Gear } from '@sim/proto/gear';
import { getReforgeCacheGearKey } from '@sim/proto/items';
import { StatCap, Stats, UnitStatPresets } from '@sim/proto/stats';
import { bulkState } from '@sim/settings/bulk_settings';
import { ReforgeSettings as ReforgeSettingsState } from '@sim/settings/reforge_settings';
import type { ReforgeOptimizeConfig, Sim } from '@sim/sim';
import type { IndividualSimUIConfig, StatTooltipContent } from '@sim/spec_config';
import { getReforgeConfigHash, makeReforgeConfigRequestFields } from '@sim/state/reforge_request';
import { SimRunKind } from '@sim/state/sim_store';
import { subscribeAll, subscribePlayerField, subscribeReforgeField } from '@sim/state/subscriptions';
import { isDevMode } from '@sim/utils/env';

import { buildEnchantOptions, classifyEnchantSlots, type EnchantSlot, type SkippedEnchantSlot, skippedEnchantSlots } from './enchant_options';
import { applyBreakpointLimits, clearSoftCappedStats, toRelativeSoftCaps } from './utils';

// Handed to the option callbacks below so a spec config can reach the optimizer and
// its own defaults without closing over the sim UI instance.
export type ReforgeOptimizerContext = {
	player: Player<any>;
	reforger: ReforgeOptimizerModel;
	defaults: IndividualSimUIConfig<any>['defaults'];
};

export type ReforgeOptimizerOptions = {
	statTooltips?: StatTooltipContent;
	statSelectionPresets?: UnitStatPresets[];
	// Allows you to enable breakpoint limits for Treshold type caps
	enableBreakpointLimits?: boolean;
	// Allows you to modify the stats before they are returned for the calculations
	// For example: Adding class specific Glyphs/Talents that are not added by the backend
	updateGearStatsModifier?: (baseStats: Stats) => Stats;
	// Allows you to get alternate default EPs
	// For example for Fury where you have SMF and TG EPs
	getEPDefaults?: (player: Player<any>, ctx: ReforgeOptimizerContext) => Stats;
	// Allows you to modify default softCaps
	// For example you wish to add breakpoints for Berserking if enabled
	updateSoftCaps?: (softCaps: StatCap[], player: Player<any>, ctx: ReforgeOptimizerContext) => StatCap[];
	// Allows you to specifiy additional information for the soft cap tooltips
	additionalSoftCapTooltipInformation?: StatTooltipContent;
};

// The spec-config values the model needs; the view reads them off `IndividualSimHost`.
export type ReforgeOptimizerModelOptions = ReforgeOptimizerOptions & {
	defaults: IndividualSimUIConfig<any>['defaults'];
	epStats: Stat[];
};

// Everything outside this module reaches the optimizer through this shape. Per-field settings
// writes are not on it: `settings` is the store-backed facade and takes them directly.
export interface ReforgeOptimizerModel {
	readonly settings: ReforgeSettingsState;
	readonly defaults: IndividualSimUIConfig<any>['defaults'];
	readonly enableBreakpointLimits: boolean;
	readonly statSelectionPresets: UnitStatPresets[] | undefined;
	/** The gear a solve started from — the diff toast reads it, and a cancel restores it. */
	readonly previousGear: Gear | null;
	readonly softCapsConfig: StatCap[];
	readonly softCapsConfigWithLimits: StatCap[];
	readonly preCapEPs: Stats;
	readonly isAllowedToOverrideStatCaps: boolean;
	readonly statCaps: Stats;
	readonly processedStatCaps: Stats;
	setUseSoftCapBreakpoints(newValue: boolean): void;
	/**
	 * The slots whose enchant the optimizer leaves alone: those whose item can take a proc enchant the sim
	 * implements (see classifyEnchantSlots), and with the enchant presets on, those no checked preset lists.
	 */
	getSkippedEnchantSlots(gear: Gear): SkippedEnchantSlot[];
	optimizeReforges(gear?: Gear): Promise<Gear>;
	abortReforgeOptimization(): Promise<void>;
	updateGear(gear: Gear): Promise<Stats>;
	computeReforgeSoftCaps(baseStats: Stats): StatCap[];
	fromProto(proto: ReforgeSettings): void;
	/** Applies a preset build's partial settings; see ReforgeSettingsState.applyPreset. */
	applyPresetSettings(proto: ReforgeSettings): void;
	applyDefaults(): void;
}

const HYBRID_CASTER_SPECS = [Spec.SpecBalanceDruid, Spec.SpecDpsPriest, Spec.SpecElementalShaman];

export const createReforgeOptimizer = (sim: Sim, player: Player<any>, options: ReforgeOptimizerModelOptions): ReforgeOptimizerModel => {
	const { defaults, epStats, getEPDefaults, updateSoftCaps, updateGearStatsModifier, statSelectionPresets } = options;
	const enableBreakpointLimits = !!options.enableBreakpointLimits;
	const isHybridCaster = HYBRID_CASTER_SPECS.includes(player.getSpec());
	const baseSoftCaps = defaults.softCapBreakpoints || [];
	const settings = new ReforgeSettingsState(player, defaults, epStats);
	let previousGear: Gear | null = null;

	const softCapsConfig = () => updateSoftCaps?.(StatCap.cloneSoftCaps(baseSoftCaps), player, ctx) || baseSoftCaps;

	const softCapsConfigWithLimits = () =>
		!enableBreakpointLimits || !settings.useSoftCapBreakpoints ? softCapsConfig() : applyBreakpointLimits(softCapsConfig(), settings.breakpointLimits);

	const preCapEPs = () => {
		let weights = player.getEpWeights();

		if (!settings.useCustomEPValues) {
			if (getEPDefaults) {
				weights = getEPDefaults(player, ctx);
			} else if (player.hasCustomEPWeights()) {
				weights = defaults.epWeights;
			}
		}

		// Replace Spirit EP for hybrid casters with a small value in order to break ties between Spirit and Hit
		if (isHybridCaster) {
			weights = weights.withStat(Stat.StatSpirit, 0.01);
		}

		return weights;
	};

	// `softCapsConfig()` is always an array and so always truthy; this reads as `!useSoftCapBreakpoints` today.
	const isAllowedToOverrideStatCaps = () => !(settings.useSoftCapBreakpoints && softCapsConfig());

	const processedStatCaps = () => (isAllowedToOverrideStatCaps() ? settings.statCaps : clearSoftCappedStats(settings.statCaps, softCapsConfigWithLimits()));

	// The skipped list the panel shows and the solve's enchant options share one classification, rebuilt only when
	// the gear, Enchanting, the frozen slots or the enchant presets change.
	let enchantSlotsMemo: { key: unknown[]; slots: EnchantSlot[] } | null = null;
	const enchantSlots = (gear: Gear): EnchantSlot[] => {
		const frozenSlots = settings.freezeItemSlots ? [...settings.frozenItemSlots] : [];
		const presets = settings.useEnchantPresets ? bulkState(player).enchantPresets : undefined;
		const key = [gear, player.hasProfession(Profession.Enchanting), frozenSlots.join(), presets];
		if (!enchantSlotsMemo || key.some((part, index) => part !== enchantSlotsMemo!.key[index])) {
			enchantSlotsMemo = { key, slots: classifyEnchantSlots(player, gear, frozenSlots, presets && enchantPresetPool(player)) };
		}
		return enchantSlotsMemo.slots;
	};

	const getReforgeOptimizeConfig = (gear: Gear): ReforgeOptimizeConfig => {
		const proto = settings.toProto();
		proto.statCaps = processedStatCaps().toProto();

		return {
			gear,
			preCapEPWeights: preCapEPs(),
			undershootCaps: settings.undershootCaps,
			settings: proto,
			softCaps: softCapsConfigWithLimits(),
			enchantOptions: buildEnchantOptions(enchantSlots(gear), preCapEPs(), settings.useEnchantPresets),
		};
	};

	const runReforgeOptimization = async (gear?: Gear) => {
		if (isDevMode()) console.log('Starting backend gear optimization...');
		previousGear = gear || player.getGear();

		const config = getReforgeOptimizeConfig(previousGear);
		if (!config.enchantOptions.length) return previousGear;
		const cache = ReforgeGearCache.get(player.getPlayerSpec(), player.sim.env);
		const configHash = await getReforgeConfigHash({
			player,
			// THE cache-key contract for the 14-day optimizer cache. `makeReforgeConfigRequestFields` declares
			// exactly which player/request state a solve depends on; when a new field that affects solve
			// output is added to Player or ReforgeOptimizeRequest, it must be reflected there or stale
			// results are served silently - and an irrelevant field left in busts every user's cache on
			// unrelated changes. Raid and gear are excluded: those are separate components of the key.
			reforgeRequest: ReforgeOptimizeRequest.create({ ...makeReforgeConfigRequestFields(config) }),
			raidBuffs: sim.raid.getBuffs(),
			partyBuffs: player.getParty()?.getBuffs(),
			debuffs: sim.raid.getDebuffs(),
		});
		const frozenItemSlots = config.settings.freezeItemSlots && config.settings.frozenItemSlots.length ? config.settings.frozenItemSlots : undefined;
		// The frozen slots are part of the gear key: the optimizer leaves a frozen slot alone, so
		// the optimized gear depends on which slots are frozen.
		const cacheKey = ReforgeGearCache.getKey(getReforgeCacheGearKey(previousGear.asSpec(), frozenItemSlots), configHash);
		const cachedGear = await cache.get(cacheKey);
		if (cachedGear) {
			if (isDevMode()) console.log('Gear optimization: cache hit.');
			return sim.db.lookupEquipmentSpec(cachedGear);
		}

		const result = await sim.reforgeOptimize(config);
		if (!result.optimizedGear) {
			throw new Error(i18n.t('sidebar.buttons.suggest_reforges.no_result'));
		}

		await cache.setGear(cacheKey, result.optimizedGear);

		return sim.db.lookupEquipmentSpec(result.optimizedGear);
	};

	const model: ReforgeOptimizerModel = {
		settings,
		defaults,
		enableBreakpointLimits,
		statSelectionPresets,
		get previousGear() {
			return previousGear;
		},
		get softCapsConfig() {
			return softCapsConfig();
		},
		get softCapsConfigWithLimits() {
			return softCapsConfigWithLimits();
		},
		get preCapEPs() {
			return preCapEPs();
		},
		get isAllowedToOverrideStatCaps() {
			return isAllowedToOverrideStatCaps();
		},
		get statCaps() {
			return settings.statCaps;
		},
		get processedStatCaps() {
			return processedStatCaps();
		},
		setUseSoftCapBreakpoints: newValue => settings.setUseSoftCapBreakpoints(newValue),
		getSkippedEnchantSlots: gear => skippedEnchantSlots(enchantSlots(gear)),
		optimizeReforges: gear => sim.runs.start(SimRunKind.ReforgeOptimize, () => runReforgeOptimization(gear)),
		// Every optimizer request starts through `optimizeReforges`, so the store-gated abort reaches each one.
		abortReforgeOptimization: () => sim.runs.abort(SimRunKind.ReforgeOptimize),
		updateGear: async gear => {
			const currentStats = await sim.getCharacterStatsForGear(gear);
			const baseStats = Stats.fromProto(currentStats.finalStats).add(player.getDebuffStats());
			return updateGearStatsModifier ? updateGearStatsModifier(baseStats) : baseStats;
		},
		computeReforgeSoftCaps: baseStats => (isAllowedToOverrideStatCaps() ? [] : toRelativeSoftCaps(softCapsConfigWithLimits(), baseStats)),
		fromProto: proto => settings.fromProto(proto),
		applyPresetSettings: proto => settings.applyPreset(proto),
		applyDefaults: () => settings.applyDefaults(),
	};

	// Closes the loop the option callbacks need; they only read it once a spec option runs.
	const ctx: ReforgeOptimizerContext = { player, reforger: model, defaults };

	subscribeAll([
		subscribeReforgeField(settings, 'useCustomEPValues'),
		subscribePlayerField(player, 'epWeights'),
		subscribeReforgeField(settings, 'statCaps'),
	])(() => {
		if (settings.useCustomEPValues && (player.hasCustomEPWeights() || !settings._statCaps.equals(defaults.statCaps || new Stats()))) {
			settings.setUseSoftCapBreakpoints(false);
		}
	});

	return model;
};
