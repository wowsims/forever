import type { Faction } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import type { BulkDropsCombined, BulkRankedResult, BulkResults, TopGearResult } from '@sim/bulk/types';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import type { BulkDropsOrder } from '@sim/state/sim_store';
import { pushTo } from '@sim/utils/collections';

import { buildTieChains, hasPairedDiffToBaseline, isTiedWithBaseline } from '../results/tie_chains';
import { dropsFactionItems, wornIdSet } from './items';
import { dropsItemSourceKeys, type DropsNames, type DropsSourceLabel, dropsSourceLabel } from './sources';

/** Resolves through the faction filter but not the phase filter, as `dropsItemLookup` builds it. */
export type DropsItemLookup = (id: number) => UIItem | undefined;

/**
 * Each item as the faction filter leaves it, phase ignored: a run's rows and the worn gear stay resolvable after a phase
 * switch. Resolved on first use, so only the ids asked for are filtered.
 */
export const dropsItemLookup = (itemById: (id: number) => UIItem | undefined, faction: Faction): DropsItemLookup => {
	const resolved = new Map<number, UIItem | undefined>();
	return id => {
		if (!resolved.has(id)) {
			const item = itemById(id);
			resolved.set(id, item && dropsFactionItems([item], faction)[0]);
		}
		return resolved.get(id);
	};
};

// The first entry with the largest value.
const largest = <T>(entries: readonly T[], value: (entry: T) => number): T | undefined =>
	entries.reduce<T | undefined>((top, entry) => (top === undefined || value(entry) > value(top) ? entry : top), undefined);

/** The one item a single-change row wears that is not among the current gear's `wornIdSet`; null for the current gear's row. */
export const dropsNewItem = (gear: Gear, worn: ReadonlySet<number>): EquippedItem | null =>
	gear.asArray().find(equipped => equipped !== null && !worn.has(equipped.id)) ?? null;

/**
 * DPS over the current gear's: the paired difference on a row with one, which may pair with an earlier stage's baseline
 * than the shown one, so the averages never decide it; the gap between the averages on a row with none.
 */
export const dropsGap = (row: TopGearResult, current: TopGearResult): number =>
	hasPairedDiffToBaseline(row) ? row.pairedDpsDiffToBaseline : row.dpsMetrics.avg - current.dpsMetrics.avg;

/** Above the current gear by `dropsGap` and told apart from it. */
const isDropsUpgrade = (row: TopGearResult, current: TopGearResult, iterations: number): boolean =>
	dropsGap(row, current) > 0 && !isTiedWithBaseline(row, current, iterations);

export interface DropsResultItem {
	itemId: number;
	/** The item's best row: slot, hand and enchant. Its proven upgrade with the largest gain, else its best avg. */
	best: BulkRankedResult;
	/** Every row of the item that is a proven upgrade, best avg first; empty when none is. */
	upgradeRows: readonly BulkRankedResult[];
	upgrade: boolean;
	/** `dropsGap` of the best row, negative for a loss. */
	gain: number;
}

/** By the best row's avg, best first; the current gear's row left out. */
export const dropsResultItems = (results: BulkResults): DropsResultItem[] => {
	const { originalGearResults: current, iterations } = results;
	const worn = wornIdSet(results.currentGear);
	// The chains hold the rows best avg first, so each item's rows keep that order.
	const rowsById = new Map<number, BulkRankedResult[]>();
	for (const row of results.chains.flat()) {
		const newItem = dropsNewItem(row.gear, worn);
		if (newItem) pushTo(rowsById, newItem.id, row);
	}
	return [...rowsById]
		.map(([itemId, rows]) => {
			const proven = rows.filter(row => isDropsUpgrade(row, current, iterations));
			const best = largest(proven, row => dropsGap(row, current)) ?? rows[0];
			return { itemId, best, upgradeRows: proven, upgrade: proven.length > 0, gain: dropsGap(best, current) };
		})
		.sort((a, b) => b.best.dpsMetrics.avg - a.best.dpsMetrics.avg);
};

/** The item's source keys among the run's ticked ones. */
export const runSourceKeys = (id: number, runSources: ReadonlySet<string>, names: DropsNames, itemById: DropsItemLookup): string[] => {
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
}

/** The upgrade with the largest gain, or undefined when none upgrades. */
export const topUpgrade = (items: readonly DropsResultItem[]) =>
	largest(
		items.filter(entry => entry.upgrade),
		entry => entry.gain,
	);

// The largest gain among a card's upgrades, so a tied row above them does not lead the card; its first item's on a card with none.
const bestGain = (group: DropsSourceGroup) => (topUpgrade(group.items) ?? group.items[0]).gain;

// Best gain first, groups with no upgrade last.
const byBestGain = (a: DropsSourceGroup, b: DropsSourceGroup) => Number(b.upgrades > 0) - Number(a.upgrades > 0) || bestGain(b) - bestGain(a);

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
		.map(([key, groupItems]) => ({
			key,
			label: dropsSourceLabel(key, names),
			items: groupItems,
			wornIds: wornByKey.get(key) ?? [],
			upgrades: groupItems.filter(entry => entry.upgrade).length,
		}))
		.sort(byBestGain);
};

// The card's upgrades over its strip, worn items included, as its "N / M" reads.
const upgradeShare = (group: DropsSourceGroup) => group.upgrades / (group.items.length + group.wornIds.length);

const byShare = (a: DropsSourceGroup, b: DropsSourceGroup) => upgradeShare(b) - upgradeShare(a) || bestGain(b) - bestGain(a);

/**
 * Best gain first, groups with no upgrade last. The cards whose best upgrade the results cannot tell apart from the best
 * upgrade of the card leading them, by the test that brackets two rows as tied (`buildTieChains` on the pair), go by
 * upgrade share and then best gain. Each tie is with the leader, never chained through a neighbour, so a card can only
 * pass cards within the error of its own gain.
 */
const recommendedOrder = (groups: readonly DropsSourceGroup[], iterations: number): DropsSourceGroup[] => {
	const sorted = [...groups].sort(byBestGain);
	const runs: Array<{ leader: TopGearResult; groups: DropsSourceGroup[] }> = [];
	for (const group of sorted.filter(entry => entry.upgrades > 0)) {
		const top = topUpgrade(group.items)!.best;
		const run = runs.at(-1);
		if (run && buildTieChains([run.leader, top], null, iterations).length === 1) run.groups.push(group);
		else runs.push({ leader: top, groups: [group] });
	}
	return [...runs.flatMap(run => run.groups.sort(byShare)), ...sorted.filter(group => !group.upgrades)];
};

/**
 * The cards in the Bosses view's order. `recommended` is `recommendedOrder`; `gain` is `dropsSourceGroups`' own order;
 * `upgrades` goes by upgrade count, then share, then best gain; `zone` by the label's detail (a drop's zone, any other
 * source's category), `recommended` within one; `combined` by Combined gain, the cards without one after by best gain.
 * Groups with no upgrade come last under every order but `zone`, where they come last in their zone.
 */
export const dropsOrderedGroups = (
	groups: readonly DropsSourceGroup[],
	order: BulkDropsOrder,
	iterations: number,
	combined: ReadonlyMap<string, BulkDropsCombined> | null,
): DropsSourceGroup[] => {
	switch (order) {
		case 'recommended':
			return recommendedOrder(groups, iterations);
		case 'gain':
			return [...groups].sort(byBestGain);
		case 'upgrades':
			return [...groups].sort((a, b) => b.upgrades - a.upgrades || byShare(a, b));
		case 'zone':
			return recommendedOrder(groups, iterations).sort((a, b) => a.label.detail.localeCompare(b.label.detail));
		case 'combined': {
			const gainOf = (group: DropsSourceGroup) => combined?.get(group.key)?.gain;
			return [...groups].sort((a, b) => {
				const [gainA, gainB] = [gainOf(a), gainOf(b)];
				if (gainA === undefined || gainB === undefined) return Number(gainA === undefined) - Number(gainB === undefined) || byBestGain(a, b);
				return gainB - gainA;
			});
		}
	}
};

/**
 * The Items view's chains: the current gear's row and, for each item, every row that is a proven upgrade (a ring that
 * upgrades in both slots lists both) or its best row when none is; one source's items only when given, as its card in
 * `groups` holds them.
 */
export const dropsItemChains = (
	results: BulkResults,
	items: readonly DropsResultItem[],
	options: { source: string | null; groups: readonly DropsSourceGroup[] },
): BulkRankedResult[][] => {
	const { source, groups } = options;
	const shown = source === null ? items : (groups.find(group => group.key === source)?.items ?? []);
	const kept = new Set([results.originalGearResults, ...shown.flatMap(entry => (entry.upgrade ? entry.upgradeRows : [entry.best]))]);
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
	/** The Recommended order's first card, whatever order the view shows; null when no card holds an upgrade. */
	farmFirst: DropsSourceGroup | null;
}

export const dropsSummary = (results: BulkResults, items: readonly DropsResultItem[], groups: readonly DropsSourceGroup[]): DropsSummary => {
	const [first] = recommendedOrder(groups, results.iterations);
	return { current: results.originalGearResults, best: topUpgrade(items) ?? null, farmFirst: first?.upgrades ? first : null };
};
