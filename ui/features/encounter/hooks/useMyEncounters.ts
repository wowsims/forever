import { Encounter } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { fromJsonLenient } from '@sim/proto/lenient_json';
import { MY_ENCOUNTERS_STORAGE_KEY } from '@sim/state/storage_keys';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

// Lenient so an entry saved with fields the protos no longer list still loads.
const myEncounterCodec: SavedDataCodec<Encounter> = {
	toJson: encounter => Encounter.toJson(encounter),
	fromJson: json => fromJsonLenient(Encounter, json),
};

// The encounters the player built and named in the designer, one list per spec, each under its name.
export const useMyEncounters = () => useSavedData(useSimHost().getStorageKey(MY_ENCOUNTERS_STORAGE_KEY), myEncounterCodec);
