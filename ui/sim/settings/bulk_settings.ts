// The bulk tab's store slice and its persisted settings blob, so the UI neither writes the
// store directly nor owns a localStorage key. Everything is derived from the player: the slice
// is `bulk[player.storeKey]` and the key is the player's spec prefix.
import { BulkSettings as BulkSettingsProto } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import {
	BulkBuildBundles,
	BulkEnchantPresets,
	BulkEnchantPresetSelection,
	BulkEncounterRef,
	BulkEncounterSelection,
	BulkTopResults,
} from '@generated/proto/ui';
import type { IMessageType } from '@protobuf-ts/runtime';

import { BulkSimItemSlot } from '../bulk/constants_auto_gen';
import type { Player } from '../player/player';
import { type BulkMode, BulkSlice, patchKeyed, seedKeyed } from '../state/sim_store';
import { specStorageKey } from '../state/storage_keys';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '../wasm/bulk_sim/constants_auto_gen';

const BULK_SETTINGS_STORAGE_KEY = 'bulk-settings.v2';
const LEGACY_BULK_SETTINGS_STORAGE_KEY = 'bulk-settings.v1';
const BULK_ENCHANT_PRESETS_STORAGE_KEY = 'bulk-enchant-presets.v1';
const BULK_BUILD_ENCHANT_PRESETS_STORAGE_KEY = 'bulk-build-enchant-presets.v1';
const BULK_BUILD_BUNDLES_STORAGE_KEY = 'bulk-build-bundles.v1';
const BULK_ENCOUNTERS_STORAGE_KEY = 'bulk-encounters.v1';
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
	mode: 'gear',
	gearTopResults: BULK_SIM_DEFAULT_TOP_RESULTS,
	buildsTopResults: 0,
	combinations: 0,
	combinationsError: null,
	combinationsPending: false,
	isRunning: false,
	gearStarted: false,
	buildsStarted: false,
	gearResults: null,
	buildsResults: null,
	runGear: null,
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
	return stored === 'gear' || stored === 'builds' ? stored : null;
};

export const storeBulkMode = (player: Player<any>, mode: BulkMode) => {
	try {
		player.sim.env.storage.setItem(specStorageKey(player.getPlayerSpec(), BULK_MODE_STORAGE_KEY), mode);
	} catch {}
};
