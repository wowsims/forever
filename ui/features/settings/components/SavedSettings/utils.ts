import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { applySavedSettings, readSavedSettings } from '@features/settings/model/saved_settings';
import { Player } from '@generated/proto/api';
import { SavedSettings } from '@generated/proto/ui';
import { withSpec } from '@sim/proto/spec_functions';
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
			itemSwap: settings.playerOptions?.itemSwap,
			spec: settings.specOptions
				? withSpec(host.player.getSpec(), Player.create(), { ...host.player.specTypeFunctions.optionsCreate(), ...settings.specOptions }).spec
				: undefined,
		});
		return { name: settings.name, data, json: serializeSettings(data), tooltip: settings.tooltip, isPreset: true };
	});
	const itemSwapEntries = (presets.itemSwaps ?? []).map(presetItemSwap => {
		// Without spec options, so loading one keeps the live ones.
		const data = SavedSettings.create({
			...readSavedSettings(host),
			enableItemSwap: true,
			itemSwap: presetItemSwap.itemSwap,
			spec: { oneofKind: undefined },
		});
		return { name: presetItemSwap.name, data, json: serializeSettings(data), tooltip: presetItemSwap.tooltip, isPreset: true };
	});
	return [...settingsEntries, ...itemSwapEntries];
};

export const loadSettings = ({ data }: SavedDataPanelEntry<SavedSettings>, host: IndividualSimHost<any>) => applySavedSettings(host, data);
