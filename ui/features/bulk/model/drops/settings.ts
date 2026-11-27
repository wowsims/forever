import { enchantPresetPool } from '@features/enchant-presets/model/enchant_presets';
import { BulkSettings } from '@generated/proto/api';
import { ItemSpec } from '@generated/proto/common';
import type { DatabaseFilters, UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import {
	bulkState,
	patchBulkState,
	storeBulkDropsCombinedGain,
	storeBulkDropsExcluded,
	storeBulkDropsFilters,
	storeBulkDropsOrder,
	storeBulkDropsSources,
	storeBulkDropsView,
} from '@sim/settings/bulk_settings';
import type { BulkDropsOrder, BulkDropsView } from '@sim/state/sim_store';
import { memoLast } from '@sim/utils/misc';

import { dropsDefaultFilters } from './filters';
import {
	dropsDungeonLevels,
	dropsFactionItems,
	type DropsItemContext,
	dropsItemContext,
	type DropsItemList,
	dropsItemList,
	dropsLevelItems,
	dropsPhaseItems,
	dropsPickedItemIds,
	dropsUsableItems,
} from './items';
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

export const setDropsView = (player: Player<any>, view: BulkDropsView) => {
	patchBulkState(player, { dropsView: view });
	storeBulkDropsView(player, view);
};

export const setDropsOrder = (player: Player<any>, order: BulkDropsOrder) => {
	patchBulkState(player, { dropsOrder: order });
	storeBulkDropsOrder(player, order);
};

// No bump: the setting adds a request after the run and leaves the candidate count alone.
export const setDropsCombinedGain = (player: Player<any>, on: boolean) => {
	patchBulkState(player, { dropsCombinedGain: on });
	storeBulkDropsCombinedGain(player, on);
};

export const dropsFiltersOf = (player: Player<any>): DatabaseFilters => bulkState(player).dropsFilters ?? dropsDefaultFilters(player.getPlayerClass());

// No bump: the filters only change what the items dialog shows.
export const setDropsFilters = (player: Player<any>, filters: DatabaseFilters) => {
	patchBulkState(player, { dropsFilters: filters });
	storeBulkDropsFilters(player, filters);
};

export interface DropsSetupData {
	/** Built from the items the level cutoff, the phase, the player's side and the class let through, so a source with none is absent. */
	tree: DropsSourceNode[];
	list: DropsItemList;
	names: DropsNames;
	/** The level of each dungeon zone the database holds, whatever the class, phase or filters. */
	dungeonLevels: ReadonlyMap<number, number>;
}

interface DropsSourceLayer {
	items: UIItem[];
	sourceKeys: ReadonlyMap<UIItem, readonly string[]>;
	tree: DropsSourceNode[];
	names: DropsNames;
}

const levelLayer = memoLast<{ items: UIItem[]; byId: ReadonlyMap<number, UIItem>; dungeonLevels: ReadonlyMap<number, number> }>();
const sourceLayer = memoLast<DropsSourceLayer>();

// The level cutoff reads only the database, so a dungeon keeps its level whatever the class, phase or filters. The item
// count is part of the key because loading the leftover database adds items to the same `Database`.
const dropsLevelLayer = (db: Database, all: UIItem[]) =>
	levelLayer([db, all.length], () => {
		const dungeonLevels = dropsDungeonLevels(all);
		const items = dropsLevelItems(all, dungeonLevels);
		return { items, byId: new Map(items.map(item => [item.id, item])), dungeonLevels };
	});

/** An item with only the sources the level cutoff keeps, so a run's results name no source the Sources tree hides; an item the cutoff leaves out entirely, such as a worn one, as the database holds it. */
export const dropsLevelItemById =
	(db: Database) =>
	(id: number): UIItem | undefined =>
		dropsLevelLayer(db, db.getAllItems()).byId.get(id) ?? db.getItemById(id);

// The part of the setup that only the database, the phase, the side and the class decide, kept until one of them
// changes. The item count is part of the key for the same reason as in `dropsLevelLayer`.
const dropsSourceLayer = (db: Database, context: DropsItemContext): DropsSourceLayer => {
	const all = db.getAllItems();
	const { playerSpec, phase, faction } = context;
	return sourceLayer([db, all.length, phase, faction, playerSpec], () => {
		const names = dropsNamesOf(db);
		const items = dropsUsableItems(dropsFactionItems(dropsPhaseItems(dropsLevelLayer(db, all).items, phase), faction), playerSpec);
		const sourceKeys = new Map(items.map(item => [item, dropsItemSourceKeys(item, names)]));
		return { items, sourceKeys, tree: dropsSourceTree(items, names), names };
	});
};

/** The one read of the database, the live gear, the phase and the Drops picks that the Setup pane and the request both use. */
export const dropsSetupFor = (player: Player<any>): DropsSetupData => {
	const context = dropsItemContext(player);
	const { dropsSources, dropsExcluded } = bulkState(player);
	const { items, sourceKeys, tree, names } = dropsSourceLayer(player.sim.db, context);
	const { dungeonLevels } = dropsLevelLayer(player.sim.db, player.sim.db.getAllItems());
	const list = dropsItemList(items, new Set(dropsSources), new Set(dropsExcluded), context, item => sourceKeys.get(item)!);
	return { tree, list, names, dungeonLevels };
};

// One new item per candidate, sent with the enchant presets Gear mode has ticked and none of its frozen slots. The server
// knows only the items a request ships, so every picked id goes in `items`.
export const dropsRequestSettings = (player: Player<any>): BulkSettings =>
	BulkSettings.create({
		items: dropsPickedItemIds(dropsSetupFor(player).list).map(id => ItemSpec.create({ id })),
		singleChange: true,
		enchants: enchantPresetPool(player),
	});
