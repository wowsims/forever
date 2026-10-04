import { PlayerSpecs } from '@sim/player/specs';
import { SETTINGS_STORAGE_SUFFIX, SHARED_SAVED_ENCOUNTER_STORAGE_KEY } from '@sim/state/persistence';
import {
	SAVED_EP_WEIGHTS_STORAGE_KEY,
	SAVED_GEAR_STORAGE_KEY,
	SAVED_ROTATION_STORAGE_KEY,
	SAVED_SETTINGS_STORAGE_KEY,
	SAVED_TALENTS_STORAGE_KEY,
	specStorageKey,
} from '@sim/state/storage_keys';
import { describe, expect, it } from 'vitest';

import { PRESET_FILTER_STORAGE_KEY } from './storage_keys';

// The keys every saved set in a user's browser is filed under, spelled out for one spec. A key that
// changes name here loses every saved set the user has under the old one, and nothing else in the
// suite would notice.
const EXPECTED_KEYS = {
	getSettingsStorageKey: '__forever_balance_druid__currentSettings__',
	getSavedGearStorageKey: '__forever_balance_druid__savedGear__',
	getSavedTalentsStorageKey: '__forever_balance_druid__savedTalents__',
	getSavedRotationStorageKey: '__forever_balance_druid__savedRotation__',
	getSavedSettingsStorageKey: '__forever_balance_druid__savedSettings__',
	getSavedEPWeightsStorageKey: '__forever_balance_druid__savedEPWeights__',
	getSavedEncounterStorageKey: 'sharedData__savedEncounter__',
	getPresetFilterStorageKey: '__forever_balance_druid__presetFilters__',
};

const key = (part: string) => specStorageKey(PlayerSpecs.BalanceDruid, part);

describe('the storage keys SimHostObject builds', () => {
	it('builds the expected keys for all eight accessors', () => {
		expect({
			getSettingsStorageKey: key(SETTINGS_STORAGE_SUFFIX),
			getSavedGearStorageKey: key(SAVED_GEAR_STORAGE_KEY),
			getSavedTalentsStorageKey: key(SAVED_TALENTS_STORAGE_KEY),
			getSavedRotationStorageKey: key(SAVED_ROTATION_STORAGE_KEY),
			getSavedSettingsStorageKey: key(SAVED_SETTINGS_STORAGE_KEY),
			getSavedEPWeightsStorageKey: key(SAVED_EP_WEIGHTS_STORAGE_KEY),
			getSavedEncounterStorageKey: SHARED_SAVED_ENCOUNTER_STORAGE_KEY,
			getPresetFilterStorageKey: key(PRESET_FILTER_STORAGE_KEY),
		}).toEqual(EXPECTED_KEYS);
	});
});
