import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { SavedEncounter } from '@generated/proto/ui';
import type { Encounter } from '@sim/raid/encounter';
import type { IndividualSimHost } from '@sim/sim_host';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';

export const encounterData = (encounter: Encounter): SavedEncounter => SavedEncounter.create({ encounter: encounter.toProto() });

export const serializeEncounter = (data: SavedEncounter): string => JSON.stringify(SavedEncounter.toJson(data));

export const encounterPresets = (presets: SpecPresets): Array<SavedDataPanelEntry<SavedEncounter>> =>
	(presets.encounters ?? []).map(preset => {
		const data = SavedEncounter.create({ encounter: preset.encounter });
		return { name: preset.name, data, json: serializeEncounter(data), tooltip: preset.tooltip, isPreset: true };
	});

export const loadEncounter = ({ data }: SavedDataPanelEntry<SavedEncounter>, { sim }: IndividualSimHost<any>) => sim.encounter.fromProto(data.encounter!);
