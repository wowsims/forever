import { SavedGearSet } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

const savedGearCodec: SavedDataCodec<SavedGearSet> = {
	toJson: gearSet => SavedGearSet.toJson(gearSet),
	fromJson: json => SavedGearSet.fromJson(json, { ignoreUnknownFields: true }),
};

export const useSavedGear = () => useSavedData(useSimHost().getSavedGearStorageKey(), savedGearCodec);
