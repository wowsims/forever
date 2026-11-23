import { enchantPresetPool } from '@features/enchant-presets/model/enchant_presets';
import { BulkSettings } from '@generated/proto/api';
import { type Faction, ItemSpec } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import {
	bulkState,
	patchBulkState,
	storeBulkDropsEnchantPresets,
	storeBulkDropsExcluded,
	storeBulkDropsSources,
	storeBulkDropsView,
} from '@sim/settings/bulk_settings';
import type { BulkDropsView } from '@sim/state/sim_store';
import { memoLast } from '@sim/utils/misc';

import { bulkEnchantPickEntries, bulkEnchantPickRuns, enchantPickSetter } from '../builds/enchant_picks';
import { dropsFactionItems, dropsItemContext, type DropsItemList, dropsItemList, dropsPhaseItems, dropsPickedItemIds } from './items';
import { dropsItemSourceKeys, type DropsNames, dropsNamesOf, type DropsSourceNode, dropsSourceTree } from './sources';

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

// Bumps `settings`: the Drops count reads its enchant pool.
export const setDropsEnchantPresetChecked = enchantPickSetter('dropsEnchantPresets', storeBulkDropsEnchantPresets, ['settings']);

export const setDropsView = (player: Player<any>, view: BulkDropsView) => {
	patchBulkState(player, { dropsView: view });
	storeBulkDropsView(player, view);
};

export interface DropsSetupData {
	/** Built from the items the phase and the player's side let through. */
	tree: DropsSourceNode[];
	list: DropsItemList;
	names: DropsNames;
}

interface DropsSourceLayer {
	items: UIItem[];
	sourceKeys: ReadonlyMap<UIItem, readonly string[]>;
	tree: DropsSourceNode[];
	names: DropsNames;
}

const sourceLayer = memoLast<DropsSourceLayer>();

// The part of the setup that only the database, the phase and the side decide, kept until one of them changes. The item
// count is part of the key because loading the leftover database adds items to the same `Database`.
const dropsSourceLayer = (db: Database, phase: number, faction: Faction): DropsSourceLayer => {
	const all = db.getAllItems();
	return sourceLayer([db, all.length, phase, faction], () => {
		const names = dropsNamesOf(db);
		const items = dropsFactionItems(dropsPhaseItems(all, phase), faction);
		const sourceKeys = new Map(items.map(item => [item, dropsItemSourceKeys(item, names)]));
		return { items, sourceKeys, tree: dropsSourceTree(items, names), names };
	});
};

/** The one read of the database, the live gear, the phase and the Drops picks that the Setup pane and the request both use. */
export const dropsSetupFor = (player: Player<any>): DropsSetupData => {
	const context = dropsItemContext(player);
	const { dropsSources, dropsExcluded } = bulkState(player);
	const { items, sourceKeys, tree, names } = dropsSourceLayer(player.sim.db, context.phase, context.faction);
	const list = dropsItemList(items, new Set(dropsSources), new Set(dropsExcluded), context, item => sourceKeys.get(item)!);
	return { tree, list, names };
};

// One new item per candidate, sent with Drops' own preset pick and none of the Gear mode's frozen slots. The server
// knows only the items a request ships, so every picked id goes in `items`.
export const dropsRequestSettings = (player: Player<any>): BulkSettings => {
	const { enchantPresets, dropsEnchantPresets } = bulkState(player);
	const presets = bulkEnchantPickRuns(bulkEnchantPickEntries(enchantPresets), dropsEnchantPresets).map(pick => pick.preset);
	return BulkSettings.create({
		items: dropsPickedItemIds(dropsSetupFor(player).list).map(id => ItemSpec.create({ id })),
		singleChange: true,
		enchants: enchantPresetPool(player, presets),
	});
};
