import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { SavedEPWeights } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';

export const epWeightsData = (weights: Stats): SavedEPWeights => SavedEPWeights.create({ epWeights: weights.toProto() });

export const serializeEpWeights = (data: SavedEPWeights): string => JSON.stringify(SavedEPWeights.toJson(data));

export const epWeightsPresets = (presets: SpecPresets, { player }: IndividualSimHost<any>): Array<SavedDataPanelEntry<SavedEPWeights>> =>
	presets.epWeights.map(preset => {
		const data = epWeightsData(preset.epWeights);
		return {
			name: preset.name,
			data,
			json: serializeEpWeights(data),
			isPreset: true,
			disabled: !!preset.enableWhen && !preset.enableWhen(player),
			afterLoad: () => preset.onLoad?.(player),
		};
	});

export const loadEpWeights = ({ data }: SavedDataPanelEntry<SavedEPWeights>, { player }: IndividualSimHost<any>) =>
	player.setEpWeights(Stats.fromProto(data.epWeights));
