import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { SavedTalents } from '@generated/proto/ui';
import type { IndividualSimHost } from '@sim/sim_host';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';

export const serializeTalents = (data: SavedTalents): string => JSON.stringify(SavedTalents.toJson(data));

export const talentPresets = (presets: SpecPresets, { player }: IndividualSimHost<any>): Array<SavedDataPanelEntry<SavedTalents>> =>
	presets.talents.map(preset => ({
		name: preset.name,
		data: preset.data,
		json: serializeTalents(preset.data),
		isPreset: true,
		afterLoad: () => preset.onLoad?.(player),
	}));

export const loadTalents = ({ data }: SavedDataPanelEntry<SavedTalents>, { player }: IndividualSimHost<any>) => player.setTalentsString(data.talentsString);
