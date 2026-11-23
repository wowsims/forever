import { enchantPresetPool } from '@features/enchant-presets/model/enchant_presets';
import { BulkSettings } from '@generated/proto/api';
import { ItemSpec } from '@generated/proto/common';
import { BulkEnchantPresetSelection, type UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import {
	bulkState,
	patchBulkState,
	storeBulkDropsEnchantPresets,
	storeBulkDropsExcluded,
	storeBulkDropsSources,
	storeBulkDropsView,
} from '@sim/settings/bulk_settings';
import type { BulkDropsView } from '@sim/state/sim_store';

import { bulkEnchantPickEntries, bulkEnchantPickRuns } from '../builds/enchant_picks';
import { checkedRefsAfter, type SavedRef } from '../builds/saved_refs';
import { dropsFactionItems, dropsItemContext, type DropsItemList, dropsItemList, dropsPhaseItems, dropsPickedItemIds } from './items';
import { type DropsNames, dropsNamesOf } from './sources';

export const setDropsSources = (player: Player<any>, keys: readonly string[]) => {
	const dropsSources = [...keys].sort();
	patchBulkState(player, { dropsSources }, ['settings']);
	storeBulkDropsSources(player, dropsSources);
};

export const setDropsExcluded = (player: Player<any>, ids: readonly number[]) => {
	const dropsExcluded = [...new Set(ids)].sort((a, b) => a - b);
	patchBulkState(player, { dropsExcluded }, ['settings']);
	storeBulkDropsExcluded(player, dropsExcluded);
};

export const setDropsItemsIncluded = (player: Player<any>, ids: readonly number[], included: boolean) => {
	const { dropsExcluded } = bulkState(player);
	const changed = new Set(ids);
	setDropsExcluded(player, included ? dropsExcluded.filter(id => !changed.has(id)) : [...dropsExcluded, ...ids]);
};

const refOf = (name: string): SavedRef => ({ name, preset: false });

// Names go in list order, and one no longer pickable goes, as in the Builds pick.
export const setDropsEnchantPresetChecked = (player: Player<any>, name: string, checked: boolean, pickable: readonly string[]) => {
	const stored = bulkState(player).dropsEnchantPresets;
	const dropsEnchantPresets = checkedRefsAfter(stored.map(refOf), refOf(name), checked, pickable.map(refOf)).map(ref => ref.name);
	patchBulkState(player, { dropsEnchantPresets }, ['settings']);
	storeBulkDropsEnchantPresets(player, BulkEnchantPresetSelection.create({ names: dropsEnchantPresets }));
};

export const setDropsView = (player: Player<any>, view: BulkDropsView) => {
	patchBulkState(player, { dropsView: view });
	storeBulkDropsView(player, view);
};

export interface DropsSetupData {
	/** The items the phase and the player's side let through: what the Sources tree is built from. */
	items: UIItem[];
	list: DropsItemList;
	names: DropsNames;
}

/** The one read of the database, the live gear, the phase and the Drops picks that the Setup pane and the request both use. */
export const dropsSetupFor = (player: Player<any>): DropsSetupData => {
	const { db } = player.sim;
	const context = dropsItemContext(player);
	const { dropsSources, dropsExcluded } = bulkState(player);
	const names = dropsNamesOf(db);
	const items = dropsFactionItems(dropsPhaseItems(db.getAllItems(), context.phase), context.faction);
	return { items, list: dropsItemList(items, new Set(dropsSources), new Set(dropsExcluded), context, names), names };
};

export const dropsItemListFor = (player: Player<any>): DropsItemList => dropsSetupFor(player).list;

// One new item per candidate, sent with Drops' own preset pick and none of the Gear mode's frozen slots. The server
// knows only the items a request ships, so every picked id goes in `items`.
export const dropsRequestSettings = (player: Player<any>): BulkSettings => {
	const { enchantPresets, dropsEnchantPresets } = bulkState(player);
	const presets = bulkEnchantPickRuns(bulkEnchantPickEntries(enchantPresets), dropsEnchantPresets).map(pick => pick.preset);
	return BulkSettings.create({
		items: dropsPickedItemIds(dropsItemListFor(player)).map(id => ItemSpec.create({ id })),
		singleChange: true,
		enchants: enchantPresetPool(player, presets),
	});
};
