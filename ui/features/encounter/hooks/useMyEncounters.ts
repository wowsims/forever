import { Encounter } from '@generated/proto/common';
import { CURRENT_API_VERSION } from '@sim/constants/other';
import { useSimHost } from '@sim/context/SimHostContext';
import { fromJsonLenient } from '@sim/proto/lenient_json';
import { MY_ENCOUNTERS_STORAGE_KEY } from '@sim/state/storage_keys';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

// Lenient so an entry saved with fields the protos no longer list still loads. An entry reads at the current api
// version, which the live encounter it is compared to always carries.
const myEncounterCodec: SavedDataCodec<Encounter> = {
	toJson: encounter => Encounter.toJson(encounter),
	fromJson: json => ({ ...fromJsonLenient(Encounter, json), apiVersion: CURRENT_API_VERSION }),
};

// The encounters the player built and named in the designer, one list per spec, each under its name.
export const useMyEncounters = () => useSavedData(useSimHost().getStorageKey(MY_ENCOUNTERS_STORAGE_KEY), myEncounterCodec);
