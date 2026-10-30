import { EquipmentSpec, UnitStats } from '@generated/proto/common';
import { BulkGearSetRef, SavedGearSet } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { PresetGear } from '@sim/presets/types';
import type { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';

import { storedSavedData } from './saved_refs';

export interface BulkGearSetEntry {
	ref: BulkGearSetRef;
	name: string;
	gear: Gear;
	bonusStats: Stats;
}

export interface BulkGearSetPick {
	name: string;
	gear: Gear;
	bonusStats: Stats;
}

const entry = (name: string, preset: boolean, gear: Gear, bonusStats: Stats, group = ''): BulkGearSetEntry => ({
	ref: BulkGearSetRef.create({ name, preset, group }),
	name,
	gear,
	bonusStats,
});

// Presets first, the ones the player can load, then the user's saved sets in the order the Saved Gear panel lists them.
// A preset brings no bonus stats, and its `onLoad` never runs here. Its ref carries its group, so presets of one name
// in two groups stay apart.
export const bulkGearSetEntries = (
	presets: readonly PresetGear[],
	saved: ReadonlyArray<{ name: string; data: SavedGearSet }>,
	player: Player<any>,
): BulkGearSetEntry[] => [
	...presets
		.filter(preset => !preset.enableWhen || preset.enableWhen(player))
		.map(preset => entry(preset.name, true, player.sim.db.lookupEquipmentSpec(preset.gear), new Stats(), preset.group)),
	...saved.map(({ name, data }) =>
		entry(name, false, player.sim.db.lookupEquipmentSpec(data.gear ?? EquipmentSpec.create()), Stats.fromProto(data.bonusStatsStats ?? UnitStats.create())),
	),
];

export const pickableBulkGearSets = (host: IndividualSimHost<any>): BulkGearSetEntry[] =>
	bulkGearSetEntries(
		host.individualConfig.presets.gear,
		storedSavedData(host, host.getSavedGearStorageKey(), SavedGearSet, { ignoreUnknownFields: true }),
		host.player,
	);

// A grouped preset is named "<name> (<group>)", so the results, Compare and the CSV tell two presets of one name apart.
export const bulkGearSetName = ({ ref, name }: BulkGearSetEntry): string =>
	ref.group ? i18n.t('bulk_tab.results.grouped_gear_set', { name, group: ref.group }) : name;
