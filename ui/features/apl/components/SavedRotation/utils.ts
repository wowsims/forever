import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { APLRotation } from '@generated/proto/apl';
import { SavedRotation } from '@generated/proto/ui';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';

export const serializeRotation = (data: SavedRotation): string => JSON.stringify(SavedRotation.toJson(data));

export const rotationPresets = (presets: SpecPresets, { player }: IndividualSimHost<any>): Array<SavedDataPanelEntry<SavedRotation>> =>
	(presets.rotations ?? []).map(preset => {
		// Defaulted so the panel's equality check always has something to compare.
		const data = SavedRotation.create({ ...preset.rotation, rotation: preset.rotation.rotation ?? APLRotation.create() });
		return {
			name: preset.name,
			data,
			json: serializeRotation(data),
			tooltip: preset.tooltip,
			isPreset: true,
			disabled: !!preset.enableWhen && !preset.enableWhen(player),
			afterLoad: () => preset.onLoad?.(player),
		};
	});

export const loadRotation = ({ data }: SavedDataPanelEntry<SavedRotation>, { player }: IndividualSimHost<any>) =>
	batch(() => player.setAplRotation(data.rotation || APLRotation.create()));
