// The bulk tab's store slice and its persisted settings blob, so the UI neither writes the
// store directly nor owns a localStorage key. Everything is derived from the player: the slice
// is `bulk[player.storeKey]` and the key is the player's spec prefix.
import { BulkSettings as BulkSettingsProto } from '@generated/proto/api';
import { Faction, ItemSlot } from '@generated/proto/common';
import {
	BulkBuildBundles,
	BulkEnchantPresets,
	BulkEnchantPresetSelection,
	BulkEncounterRef,
	BulkEncounterSelection,
	BulkTopResults,
	DatabaseFilters,
} from '@generated/proto/ui';
import type { IMessageType } from '@protobuf-ts/runtime';

import { BulkSimItemSlot } from '../bulk/constants_auto_gen';
import type { Player } from '../player/player';
import { BULK_DROPS_ORDERS, type BulkDropsOrder, type BulkDropsView, type BulkMode, BulkSlice, patchKeyed, seedKeyed } from '../state/sim_store';
import { specStorageKey } from '../state/storage_keys';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '../wasm/bulk_sim/constants_auto_gen';

const BULK_SETTINGS_STORAGE_KEY = 'bulk-settings.v2';
const LEGACY_BULK_SETTINGS_STORAGE_KEY = 'bulk-settings.v1';
const BULK_ENCHANT_PRESETS_STORAGE_KEY = 'bulk-enchant-presets.v1';
const BULK_BUILD_ENCHANT_PRESETS_STORAGE_KEY = 'bulk-build-enchant-presets.v1';
const BULK_BUILD_BUNDLES_STORAGE_KEY = 'bulk-build-bundles.v1';
const BULK_ENCOUNTERS_STORAGE_KEY = 'bulk-encounters.v1';
const BULK_BATCH_ENCOUNTER_STORAGE_KEY = 'bulk-batch-encounter.v1';
const BULK_DROPS_SOURCES_STORAGE_KEY = 'bulk-drops-sources.v1';
const BULK_DROPS_EXCLUDED_STORAGE_KEY = 'bulk-drops-excluded.v1';
const BULK_DROPS_VIEW_STORAGE_KEY = 'bulk-drops-view.v1';
const BULK_DROPS_ORDER_STORAGE_KEY = 'bulk-drops-order.v1';
const BULK_DROPS_FILTERS_STORAGE_KEY = 'bulk-drops-filters.v1';
const BULK_DROPS_COMBINED_GAIN_STORAGE_KEY = 'bulk-drops-combined-gain.v1';
const BULK_MODE_STORAGE_KEY = 'bulk-mode.v1';
const BULK_TOP_RESULTS_STORAGE_KEY = 'bulk-top-results.v1';

const initialBulkSlice = (): BulkSlice => ({
	items: [],
	pickerGroups: new Map(),
	frozenItems: new Map([
		[BulkSimItemSlot.ItemSlotFinger, null],
		[BulkSimItemSlot.ItemSlotTrinket, null],
	]),
	frozenWeaponSlot: undefined,
	weaponTypeFilters: new Map([
		[ItemSlot.ItemSlotMainHand, []],
		[ItemSlot.ItemSlotOffHand, []],
	]),
	enchantPresets: [],
	buildEnchantPresets: [],
	bundles: [],
	currentBundleChecked: true,
	// The current encounter's ref, so it is checked until unticked.
	encounters: [BulkEncounterRef.create()],
	batchEncounter: BulkEncounterRef.create(),
	dropsSources: [],
	dropsExcluded: [],
	dropsView: 'bosses',
	dropsOrder: 'recommended',
	dropsFilters: null,
	dropsCombinedGain: false,
	mode: 'gear',
	gearTopResults: BULK_SIM_DEFAULT_TOP_RESULTS,
	buildsTopResults: 0,
	combinations: 0,
	combinationsError: null,
	combinationsPending: false,
	isRunning: false,
	gearStarted: false,
	dropsStarted: false,
	buildsStarted: false,
	gearResults: null,
	dropsResults: null,
	buildsResults: null,
	runGear: null,
	dropsRunSources: [],
	dropsRunFaction: Faction.Unknown,
	dropsCombined: null,
	v: { settings: 0, items: 0 },
});

// Seeds the slice before any subscriber exists (emit-less).
export const seedBulkSettings = (player: Player<any>) => seedKeyed(player.sim.store, 'bulk', player.storeKey, initialBulkSlice());

export const bulkState = (player: Player<any>): BulkSlice => player.sim.store.getState().bulk[player.storeKey];

// A bump with an empty patch is a bare emit; a patch with no bumps is a silent field write.
export const patchBulkState = (player: Player<any>, patch: Partial<Omit<BulkSlice, 'v'>>, bumps: ReadonlyArray<keyof BulkSlice['v']> = []) =>
	patchKeyed(player.sim.store, 'bulk', player.storeKey, patch, bumps);

// Null when nothing is stored, empty defaults when the blob is unparseable.
const loadStoredProto = <T extends object>(player: Player<any>, key: string, type: IMessageType<T>): T | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), key));
	if (stored == null) return null;
	try {
		return type.fromJsonString(stored, { ignoreUnknownFields: true });
	} catch {
		return type.create();
	}
};

// Over the quota the previously stored value stays, rather than the key being dropped with it.
const storeProto = <T extends object>(player: Player<any>, key: string, type: IMessageType<T>, value: T) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), key), type.toJsonString(value, { enumAsInteger: true }));
	} catch {}
};

// Drops the v1 key on the way.
export const loadStoredBulkSettings = (player: Player<any>): BulkSettingsProto | null => {
	player.sim.env.storage.removeItem(specStorageKey(player.getPlayerSpec(), LEGACY_BULK_SETTINGS_STORAGE_KEY));
	return loadStoredProto(player, BULK_SETTINGS_STORAGE_KEY, BulkSettingsProto);
};

// A batch's item list can outgrow the storage quota; drop the key rather than leave a
// half-written blob behind, matching what the pre-React bulk tab did.
export const storeBulkSettings = (player: Player<any>, settings: BulkSettingsProto) => {
	const { storage } = player.sim.env;
	const key = specStorageKey(player.getPlayerSpec(), BULK_SETTINGS_STORAGE_KEY);
	try {
		storage.setItem(key, BulkSettingsProto.toJsonString(settings, { enumAsInteger: true }));
	} catch (e) {
		if (e instanceof DOMException && e.name === 'QuotaExceededError') {
			storage.removeItem(key);
		}
	}
};

export const loadStoredEnchantPresets = (player: Player<any>): BulkEnchantPresets | null =>
	loadStoredProto(player, BULK_ENCHANT_PRESETS_STORAGE_KEY, BulkEnchantPresets);

export const storeEnchantPresets = (player: Player<any>, presets: BulkEnchantPresets) =>
	storeProto(player, BULK_ENCHANT_PRESETS_STORAGE_KEY, BulkEnchantPresets, presets);

export const loadStoredBulkBuildEnchantPresets = (player: Player<any>): BulkEnchantPresetSelection | null =>
	loadStoredProto(player, BULK_BUILD_ENCHANT_PRESETS_STORAGE_KEY, BulkEnchantPresetSelection);

export const storeBulkBuildEnchantPresets = (player: Player<any>, selection: BulkEnchantPresetSelection) =>
	storeProto(player, BULK_BUILD_ENCHANT_PRESETS_STORAGE_KEY, BulkEnchantPresetSelection, selection);

export const loadStoredBulkBundles = (player: Player<any>): BulkBuildBundles | null =>
	loadStoredProto(player, BULK_BUILD_BUNDLES_STORAGE_KEY, BulkBuildBundles);

export const storeBulkBundles = (player: Player<any>, value: BulkBuildBundles) => storeProto(player, BULK_BUILD_BUNDLES_STORAGE_KEY, BulkBuildBundles, value);

export const loadStoredBulkEncounters = (player: Player<any>): BulkEncounterSelection | null =>
	loadStoredProto(player, BULK_ENCOUNTERS_STORAGE_KEY, BulkEncounterSelection);

export const storeBulkEncounters = (player: Player<any>, value: BulkEncounterSelection) =>
	storeProto(player, BULK_ENCOUNTERS_STORAGE_KEY, BulkEncounterSelection, value);

// Null when nothing is stored; a blob that does not parse loads as the empty ref, which is Current encounter.
export const loadStoredBatchEncounter = (player: Player<any>): BulkEncounterRef | null =>
	loadStoredProto(player, BULK_BATCH_ENCOUNTER_STORAGE_KEY, BulkEncounterRef);

export const storeBatchEncounter = (player: Player<any>, value: BulkEncounterRef) =>
	storeProto(player, BULK_BATCH_ENCOUNTER_STORAGE_KEY, BulkEncounterRef, value);

// Null when nothing is stored and when the blob does not parse: an empty message would read as All for both modes. A
// negative value loads as its mode's default, 5 for Gear and All for Builds.
export const loadStoredBulkTopResults = (player: Player<any>): BulkTopResults | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), BULK_TOP_RESULTS_STORAGE_KEY));
	if (stored == null) return null;
	try {
		const { gear, builds } = BulkTopResults.fromJsonString(stored, { ignoreUnknownFields: true });
		return BulkTopResults.create({ gear: gear < 0 ? BULK_SIM_DEFAULT_TOP_RESULTS : gear, builds: builds < 0 ? 0 : builds });
	} catch {
		return null;
	}
};

export const storeBulkTopResults = (player: Player<any>, value: BulkTopResults) => storeProto(player, BULK_TOP_RESULTS_STORAGE_KEY, BulkTopResults, value);

// Null when nothing or an unknown mode is stored.
export const loadStoredBulkMode = (player: Player<any>): BulkMode | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), BULK_MODE_STORAGE_KEY));
	return stored === 'gear' || stored === 'drops' || stored === 'builds' ? stored : null;
};

export const storeBulkMode = (player: Player<any>, mode: BulkMode) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), BULK_MODE_STORAGE_KEY), mode);
	} catch {}
};

// Null when nothing is stored, when the blob does not parse and when it is not an array of what `isItem` accepts.
const loadStoredJsonArray = <T>(player: Player<any>, key: string, isItem: (value: unknown) => value is T): T[] | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), key));
	if (stored == null) return null;
	try {
		const parsed: unknown = JSON.parse(stored);
		return Array.isArray(parsed) && parsed.every(isItem) ? parsed : null;
	} catch {
		return null;
	}
};

const storeJsonArray = (player: Player<any>, key: string, values: readonly unknown[]) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), key), JSON.stringify(values));
	} catch {}
};

export const loadStoredBulkDropsSources = (player: Player<any>): string[] | null =>
	loadStoredJsonArray(player, BULK_DROPS_SOURCES_STORAGE_KEY, (value): value is string => typeof value === 'string');

export const storeBulkDropsSources = (player: Player<any>, keys: readonly string[]) => storeJsonArray(player, BULK_DROPS_SOURCES_STORAGE_KEY, keys);

export const loadStoredBulkDropsExcluded = (player: Player<any>): number[] | null =>
	loadStoredJsonArray(player, BULK_DROPS_EXCLUDED_STORAGE_KEY, (value): value is number => Number.isInteger(value));

export const storeBulkDropsExcluded = (player: Player<any>, ids: readonly number[]) => storeJsonArray(player, BULK_DROPS_EXCLUDED_STORAGE_KEY, ids);

// Null when nothing or an unknown view is stored.
export const loadStoredBulkDropsView = (player: Player<any>): BulkDropsView | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_VIEW_STORAGE_KEY));
	return stored === 'bosses' || stored === 'items' ? stored : null;
};

export const storeBulkDropsView = (player: Player<any>, view: BulkDropsView) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_VIEW_STORAGE_KEY), view);
	} catch {}
};

// Null when nothing or an unknown order is stored.
export const loadStoredBulkDropsOrder = (player: Player<any>): BulkDropsOrder | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_ORDER_STORAGE_KEY));
	return BULK_DROPS_ORDERS.find(order => order === stored) ?? null;
};

export const storeBulkDropsOrder = (player: Player<any>, order: BulkDropsOrder) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_ORDER_STORAGE_KEY), order);
	} catch {}
};

// Null when nothing is stored and when the blob does not parse: empty filters would hide every item.
export const loadStoredBulkDropsFilters = (player: Player<any>): DatabaseFilters | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_FILTERS_STORAGE_KEY));
	if (stored == null) return null;
	try {
		return DatabaseFilters.fromJsonString(stored, { ignoreUnknownFields: true });
	} catch {
		return null;
	}
};

export const storeBulkDropsFilters = (player: Player<any>, filters: DatabaseFilters) =>
	storeProto(player, BULK_DROPS_FILTERS_STORAGE_KEY, DatabaseFilters, filters);

// Null when nothing or anything but `true` or `false` is stored.
export const loadStoredBulkDropsCombinedGain = (player: Player<any>): boolean | null => {
	const stored = player.sim.env.storage.getItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_COMBINED_GAIN_STORAGE_KEY));
	return stored === 'true' ? true : stored === 'false' ? false : null;
};

export const storeBulkDropsCombinedGain = (player: Player<any>, on: boolean) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), BULK_DROPS_COMBINED_GAIN_STORAGE_KEY), String(on));
	} catch {}
};
