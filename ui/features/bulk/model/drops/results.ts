import { BulkSimStage } from '@generated/proto/api';
import type { Faction } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import type { BulkRankedResult, BulkResults, TopGearResult } from '@sim/bulk/types';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';

import { buildTieChains, isTiedWithBaseline } from '../results/tie_chains';
import { dropsFactionItems } from './items';
import { dropsItemSourceKeys, type DropsNames, type DropsSourceLabel, dropsSourceLabel } from './sources';

/** Resolves through the faction filter but not the phase filter, as `dropsItemLookup` builds it. */
export type DropsItemLookup = (id: number) => UIItem | undefined;

/** The items as the faction filter leaves them, phase ignored: a run's rows and the worn gear stay resolvable after a phase switch. */
export const dropsItemLookup = (items: readonly UIItem[], faction: Faction): DropsItemLookup => {
	const byId = new Map(dropsFactionItems(items, faction).map(item => [item.id, item]));
	return id => byId.get(id);
};

const wornIdSet = (gear: Gear): Set<number> => new Set(gear.asArray().flatMap(equipped => (equipped ? [equipped.id] : [])));

const pushTo = <K, V>(map: Map<K, V[]>, key: K, value: V) => {
	const list = map.get(key);
	if (list) list.push(value);
	else map.set(key, [value]);
};

// The first entry with the largest value.
const largest = <T>(entries: readonly T[], value: (entry: T) => number): T | undefined =>
	entries.reduce<T | undefined>((top, entry) => (top === undefined || value(entry) > value(top) ? entry : top), undefined);

/** The one item a single-change row wears that the current gear does not, by id; null for the current gear's row. */
export const dropsNewItem = (gear: Gear, current: Gear): EquippedItem | null => {
	const worn = wornIdSet(current);
	return gear.asArray().find(equipped => equipped !== null && !worn.has(equipped.id)) ?? null;
};

/** Culled before the finalist set: no ± and no tight bracket. */
export const isDropsLowPrecision = (row: TopGearResult): boolean => row.culledStage !== undefined && row.culledStage !== BulkSimStage.BulkSimStageUnknown;

/**
 * DPS over the current gear's: the paired difference on a row with a paired error, which may pair with an earlier
 * stage's baseline than the shown one, so the averages never decide it; the gap between the averages on a row with none.
 */
export const dropsGap = (row: TopGearResult, current: TopGearResult): number =>
	row.pairedErrorToBaseline && row.pairedDpsDiffToBaseline !== undefined ? row.pairedDpsDiffToBaseline : row.dpsMetrics.avg - current.dpsMetrics.avg;

/** Above the current gear by `dropsGap` and told apart from it. */
export const isDropsUpgrade = (row: TopGearResult, current: TopGearResult): boolean => dropsGap(row, current) > 0 && !isTiedWithBaseline(row, current);

export interface DropsResultItem {
	itemId: number;
	/** The item's best row: slot, hand and enchant. Its proven upgrade with the largest gain, else its best avg. */
	best: BulkRankedResult;
	/** The item's other rows, best avg first. */
	variants: readonly BulkRankedResult[];
	upgrade: boolean;
	/** `dropsGap` of the best row, negative for a loss. */
	gain: number;
}

/** By the best row's avg, best first; the current gear's row left out. */
export const dropsResultItems = (results: BulkResults): DropsResultItem[] => {
	const current = results.originalGearResults;
	// The chains hold the rows best avg first, so each item's rows keep that order.
	const rowsById = new Map<number, BulkRankedResult[]>();
	for (const row of results.chains.flat()) {
		const newItem = dropsNewItem(row.gear, results.currentGear);
		if (newItem) pushTo(rowsById, newItem.id, row);
	}
	return [...rowsById]
		.map(([itemId, rows]) => {
			const proven = rows.filter(row => isDropsUpgrade(row, current));
			const best = largest(proven, row => dropsGap(row, current)) ?? rows[0];
			return { itemId, best, variants: rows.filter(row => row !== best), upgrade: proven.length > 0, gain: dropsGap(best, current) };
		})
		.sort((a, b) => b.best.dpsMetrics.avg - a.best.dpsMetrics.avg);
};

const runSourceKeys = (id: number, runSources: ReadonlySet<string>, names: DropsNames, itemById: DropsItemLookup): string[] => {
	const item = itemById(id);
	return item ? dropsItemSourceKeys(item, names).filter(key => runSources.has(key)) : [];
};

export interface DropsSourceGroup {
	key: string;
	label: DropsSourceLabel;
	/** Best first. */
	items: readonly DropsResultItem[];
	/** Worn items this source drops, shown as "worn". */
	wornIds: readonly number[];
	upgrades: number;
	/** The items and the worn ones. */
	total: number;
}

/** The upgrade with the largest gain, or undefined when none upgrades. */
export const topUpgrade = (items: readonly DropsResultItem[]) =>
	largest(
		items.filter(entry => entry.upgrade),
		entry => entry.gain,
	);

// The largest gain among a card's upgrades, so a tied row above them does not lead the card; its first item's on a card with none.
const bestGain = (group: DropsSourceGroup) => (topUpgrade(group.items) ?? group.items[0]).gain;

/** One card per ticked leaf the run's rows reach, best gain first, groups with no upgrade last. `itemById` is a `DropsItemLookup`. */
export const dropsSourceGroups = (
	items: readonly DropsResultItem[],
	runSources: ReadonlySet<string>,
	currentGear: Gear,
	itemById: DropsItemLookup,
	names: DropsNames,
): DropsSourceGroup[] => {
	const itemsByKey = new Map<string, DropsResultItem[]>();
	for (const entry of items) {
		for (const key of runSourceKeys(entry.itemId, runSources, names, itemById)) pushTo(itemsByKey, key, entry);
	}
	const wornByKey = new Map<string, number[]>();
	for (const id of wornIdSet(currentGear)) {
		for (const key of runSourceKeys(id, runSources, names, itemById)) pushTo(wornByKey, key, id);
	}
	return [...itemsByKey]
		.map(([key, groupItems]) => {
			const wornIds = wornByKey.get(key) ?? [];
			return {
				key,
				label: dropsSourceLabel(key, names),
				items: groupItems,
				wornIds,
				upgrades: groupItems.filter(entry => entry.upgrade).length,
				total: groupItems.length + wornIds.length,
			};
		})
		.sort((a, b) => Number(b.upgrades > 0) - Number(a.upgrades > 0) || bestGain(b) - bestGain(a));
};

/**
 * The Items view's chains: each item's best row and the current gear's, or every row with showAll; one source's items only
 * when given. `itemById` is a `DropsItemLookup`.
 */
export const dropsItemChains = (
	results: BulkResults,
	items: readonly DropsResultItem[],
	options: { showAll: boolean; source: string | null; runSources: ReadonlySet<string>; names: DropsNames; itemById: DropsItemLookup },
): BulkRankedResult[][] => {
	const { showAll, source, runSources, names, itemById } = options;
	if (showAll && source === null) return results.chains;
	const shown = source === null ? items : items.filter(entry => runSourceKeys(entry.itemId, runSources, names, itemById).includes(source));
	const kept = new Set([results.originalGearResults, ...shown.flatMap(entry => (showAll ? [entry.best, ...entry.variants] : [entry.best]))]);
	return buildTieChains(
		results.chains.flat().filter(row => kept.has(row)),
		results.originalGearResults,
		results.iterations,
	);
};

export interface DropsSummary {
	current: BulkRankedResult;
	/** The upgrade with the largest gain; null when nothing upgrades. */
	best: DropsResultItem | null;
	/** The first card holding `best`; null when nothing upgrades, or when `best` has no card among the run's sources. */
	bestGroup: DropsSourceGroup | null;
}

export const dropsSummary = (results: BulkResults, items: readonly DropsResultItem[], groups: readonly DropsSourceGroup[]): DropsSummary => {
	const best = topUpgrade(items) ?? null;
	return { current: results.originalGearResults, best, bestGroup: (best && groups.find(group => group.items.includes(best))) ?? null };
};
