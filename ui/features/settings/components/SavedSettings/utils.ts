import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { applySavedSettings, readSavedSettings } from '@features/settings/model/saved_settings';
import { SavedSettings } from '@generated/proto/ui';
import type { IndividualSimHost } from '@sim/sim_host';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';

export const serializeSettings = (data: SavedSettings): string => JSON.stringify(SavedSettings.toJson(data));

export const settingsPresets = (presets: SpecPresets, host: IndividualSimHost<any>): Array<SavedDataPanelEntry<SavedSettings>> => {
	const settingsEntries = (presets.settings ?? []).map(settings => {
		const data = SavedSettings.create({
			race: settings.race,
			raidBuffs: settings.raidBuffs,
			partyBuffs: settings.partyBuffs,
			playerBuffs: settings.buffs,
			debuffs: settings.debuffs,
			consumables: settings.consumables,
			professions:
				settings.playerOptions?.profession1 && settings.playerOptions?.profession2
					? [settings.playerOptions.profession1, settings.playerOptions.profession2]
					: undefined,
			distanceFromTarget: settings.playerOptions?.distanceFromTarget,
			reactionTimeMs: settings.playerOptions?.reactionTimeMs,
			channelClipDelayMs: settings.playerOptions?.channelClipDelayMs,
			inFrontOfTarget: settings.playerOptions?.inFrontOfTarget,
			enableItemSwap: settings.playerOptions?.enableItemSwap,
		});
		return { name: settings.name, data, json: serializeSettings(data), tooltip: settings.tooltip, isPreset: true };
	});
	const itemSwapEntries = (presets.itemSwaps ?? []).map(presetItemSwap => {
		const data = SavedSettings.create({ ...readSavedSettings(host), enableItemSwap: true, itemSwap: presetItemSwap.itemSwap });
		return { name: presetItemSwap.name, data, json: serializeSettings(data), tooltip: presetItemSwap.tooltip, isPreset: true };
	});
	return [...settingsEntries, ...itemSwapEntries];
};

export const loadSettings = ({ data }: SavedDataPanelEntry<SavedSettings>, host: IndividualSimHost<any>) => applySavedSettings(host, data);
