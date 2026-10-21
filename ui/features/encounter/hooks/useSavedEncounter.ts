import { SavedEncounter } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import { fromJsonLenient } from '@sim/proto/lenient_json';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

// Lenient so an entry saved with fields the protos no longer list still loads.
export const savedEncounterCodec: SavedDataCodec<SavedEncounter> = {
	toJson: encounter => SavedEncounter.toJson(encounter),
	fromJson: json => fromJsonLenient(SavedEncounter, json),
};

export const useSavedEncounter = () => useSavedData(useSimHost().getSavedEncounterStorageKey(), savedEncounterCodec);
