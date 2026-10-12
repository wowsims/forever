// The bulk tab's store slice and its persisted settings blob, so the UI neither writes the
// store directly nor owns a localStorage key. Everything is derived from the player: the slice
// is `bulk[player.storeKey]` and the key is the player's spec prefix.
import { BulkSettings as BulkSettingsProto } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkDurations, BulkEnchantPresets } from '@generated/proto/ui';
import type { IMessageType } from '@protobuf-ts/runtime';

import { BulkSimItemSlot } from '../bulk/constants_auto_gen';
import type { Player } from '../player/player';
import { BulkSlice, patchKeyed, seedKeyed } from '../state/sim_store';
import { specStorageKey } from '../state/storage_keys';

const BULK_SETTINGS_STORAGE_KEY = 'bulk-settings.v2';
const LEGACY_BULK_SETTINGS_STORAGE_KEY = 'bulk-settings.v1';
const BULK_ENCHANT_PRESETS_STORAGE_KEY = 'bulk-enchant-presets.v1';
const BULK_DURATIONS_STORAGE_KEY = 'bulk-durations.v1';

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
	durations: [],
	combinations: 0,
	iterations: 0,
	combinationsPending: false,
	isRunning: false,
	started: false,
	results: null,
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

// Over the quota the previously stored presets stay, rather than the key being dropped with them.
export const storeEnchantPresets = (player: Player<any>, presets: BulkEnchantPresets) => {
	try {
		player.sim.env.storage.setItem(
			specStorageKey(player.getPlayerSpec(), BULK_ENCHANT_PRESETS_STORAGE_KEY),
			BulkEnchantPresets.toJsonString(presets, { enumAsInteger: true }),
		);
	} catch {}
};

export const loadStoredBulkDurations = (player: Player<any>): BulkDurations | null => loadStoredProto(player, BULK_DURATIONS_STORAGE_KEY, BulkDurations);

// Over the quota the previously stored durations stay.
export const storeBulkDurations = (player: Player<any>, durations: BulkDurations) => {
	try {
		player.sim.env.storage.setItem(
			specStorageKey(player.getPlayerSpec(), BULK_DURATIONS_STORAGE_KEY),
			BulkDurations.toJsonString(durations, { enumAsInteger: true }),
		);
	} catch {}
};
