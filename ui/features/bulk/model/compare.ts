import type { DistributionMetrics } from '@generated/proto/api';
import { ItemSpec } from '@generated/proto/common';
import { translateProtoStatName } from '@i18n/localization';
import type { BulkEncounterResults } from '@sim/bulk/types';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';

export interface ComparedBar {
	key: string;
	isCurrent: boolean;
	/** The set's number across the whole comparison; 0 for the current gear. */
	number: number;
	dpsMetrics: DistributionMetrics;
	iterations: number;
	/** DPS over the current gear's in the same encounter. */
	gain: number;
}

export interface ComparedEncounter {
	entry: BulkEncounterResults;
	/** The encounter's own ranking, the current gear included, best first. */
	bars: ComparedBar[];
}

export interface ComparedSet {
	key: string;
	gear: Gear;
	number: number;
}

export interface BulkComparison {
	/** By duration, shortest first; health-based ones last, in run order. */
	encounters: ComparedEncounter[];
	/** Every ranked set but the current gear, by number. */
	sets: ComparedSet[];
	current: Gear | null;
}

const gearKey = (gear: Gear) => getGearIdentityKey(gear.asSpec());

const durationOrder = (entry: BulkEncounterResults) => entry.duration ?? Infinity;

// A set is numbered where it first shows, reading the encounters in display order and each one's ranking top down, and
// keeps that number in every other encounter that ranks it.
export const bulkComparison = (entries: readonly BulkEncounterResults[]): BulkComparison => {
	const sets = new Map<string, ComparedSet>();
	const current = entries[0]?.results.originalGearResults.gear ?? null;
	const currentKey = current ? gearKey(current) : '';
	const encounters = [...entries]
		.sort((a, b) => durationOrder(a) - durationOrder(b))
		.map((entry): ComparedEncounter => {
			const { chains, originalGearResults, iterations } = entry.results;
			const bars = chains.flat().map((result): ComparedBar => {
				const key = gearKey(result.gear);
				const isCurrent = key === currentKey;
				if (!isCurrent && !sets.has(key)) sets.set(key, { key, gear: result.gear, number: sets.size + 1 });
				return {
					key,
					isCurrent,
					number: isCurrent ? 0 : sets.get(key)!.number,
					dpsMetrics: result.dpsMetrics,
					iterations,
					gain: result.dpsMetrics.avg - originalGearResults.dpsMetrics.avg,
				};
			});
			return { entry, bars };
		});
	return { encounters, sets: [...sets.values()], current };
};

// "Item Suffix (Enchant)", as the result rows name it.
const itemLabel = (equipped: EquippedItem): string => {
	const suffix = equipped.randomSuffix ? ` ${translateProtoStatName(equipped.randomSuffix.name)}` : '';
	const enchant = equipped.enchant ? ` (${equipped.enchant.name})` : '';
	return `${equipped.item.name}${suffix}${enchant}`;
};

export const changedItemLabels = (gear: Gear, current: Gear): string[] =>
	gear.getItemSlots().flatMap(slot => {
		const equipped = gear.getEquippedItem(slot);
		const worn = current.getEquippedItem(slot);
		return equipped && !(worn && ItemSpec.equals(equipped.asSpec(), worn.asSpec())) ? [itemLabel(equipped)] : [];
	});

// The x range of one encounter's bars: the shortest bar keeps a third of the longest one's length, so the difference
// between them shows and no bar draws empty.
export const dpsAxisRange = (values: readonly number[]): { min: number; max: number } => {
	if (!values.length) return { min: 0, max: 1 };
	const low = Math.min(...values);
	const high = Math.max(...values);
	const pad = high > low ? (high - low) / 2 : Math.max(1, Math.abs(high) / 100);
	return { min: Math.max(0, low - pad), max: high };
};
