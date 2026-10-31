import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { settingsPresets } from '@features/settings/components/SavedSettings/utils';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel, ItemSpec, Profession } from '@generated/proto/common';
import { BulkSettingsRef, SavedSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';

import type { BulkPart } from './bundles';
import { keyedOnce, storedSavedData } from './saved_refs';

export interface BulkSettingsEntry {
	ref: BulkSettingsRef;
	name: string;
	settings: SavedSettings;
}

const entry = (name: string, preset: boolean, settings: SavedSettings): BulkSettingsEntry => ({
	ref: BulkSettingsRef.create({ name, preset }),
	name,
	settings,
});

// The Saved Settings panel's presets first, the spec's settings and then its item swaps, then the user's saved entries in
// the order the panel lists them.
export const bulkSettingsEntries = (
	presets: SpecPresets,
	saved: ReadonlyArray<{ name: string; data: SavedSettings }>,
	host: IndividualSimHost<any>,
): BulkSettingsEntry[] => [
	...settingsPresets(presets, host).map(({ name, data }) => entry(name, true, data)),
	...saved.map(({ name, data }) => entry(name, false, data)),
];

// Read leniently, as the panel reads them, so an entry the panel lists is pickable here too.
export const pickableBulkSettings = (host: IndividualSimHost<any>): BulkSettingsEntry[] =>
	bulkSettingsEntries(
		host.individualConfig.presets,
		storedSavedData(host, host.getSavedSettingsStorageKey(), SavedSettings, { ignoreUnknownFields: true }),
		host,
	);

// As `Sim.applyBulkSettings` sends an entry: an absent message as an empty one, the item swap as the items it holds and
// its prepull bonus stats, and the two fields it never sends left out. The sim reads the item swap only while it is on.
// An entry without spec options keys without them; `withLiveSpecOptions` gives it the live ones first where those are
// known.
export const settingsContentKey = keyedOnce((settings: SavedSettings) => {
	const itemSwap = settings.enableItemSwap ? settings.itemSwap : undefined;
	return JSON.stringify([
		SavedSettings.toJson({
			...settings,
			raidBuffs: settings.raidBuffs ?? RaidBuffs.create(),
			partyBuffs: settings.partyBuffs ?? PartyBuffs.create(),
			playerBuffs: settings.playerBuffs ?? IndividualBuffs.create(),
			debuffs: settings.debuffs ?? Debuffs.create(),
			consumables: settings.consumables ?? ConsumesSpec.create(),
			professions: [settings.professions[0] || Profession.ProfessionUnknown, settings.professions[1] || Profession.ProfessionUnknown],
			itemSwap: undefined,
			healingModel: settings.healingModel ?? HealingModel.create(),
			darkIntentUptime: 0,
			challengeMode: false,
		}),
		(itemSwap?.items ?? []).flatMap((item, slot) => (item.id ? [[slot, ItemSpec.toJson(item)]] : [])),
		Stats.fromProto(itemSwap?.prepullBonusStats).toJson(),
	]);
});

// The current settings run unnamed; every view calls them "Current settings".
export const settingsLabel = ({ isCurrent, name }: BulkPart): string => (isCurrent ? i18n.t('bulk_tab.results.current_settings') : name);
