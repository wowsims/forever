import { type ItemSlot, ItemSpec } from '@generated/proto/common';
import type { BulkEncounterResults, BulkTalentSetup, TopGearResult } from '@sim/bulk/types';
import type { Gear } from '@sim/proto/gear';

import { formatBulkItemLabel } from './format';
import { bulkResultKey } from './results';
import { getSwappableItemSlotPair, itemSpecPairsEqualUnordered } from './swappable_slots';

interface ComparedSet {
	key: string;
	gear: Gear;
	talents: BulkTalentSetup;
	number: number;
}

const durationOrder = (entry: BulkEncounterResults) => entry.duration ?? Infinity;

// The items the set wears where the current gear wears something else, an enchant change included. A pair of slots
// whose two items only change places changes nothing.
export const changedItemLabels = (gear: Gear, current: Gear, canDualWield: boolean): string[] => {
	const items = gear.asSpec().items;
	const worn = current.asSpec().items;
	const unchanged = (slot: ItemSlot) => {
		const pair = getSwappableItemSlotPair(slot, canDualWield);
		return ItemSpec.equals(items[slot], worn[slot]) || (pair !== undefined && itemSpecPairsEqualUnordered(items, worn, pair));
	};
	return gear.getItemSlots().flatMap(slot => {
		const equipped = gear.getEquippedItem(slot);
		return equipped && !unchanged(slot) ? [formatBulkItemLabel(equipped)] : [];
	});
};

const percentOf = (value: number, of: number) => (of ? (value / of) * 100 : 0);

/** A matrix cell's tint: the gain's sign and its size against the table's largest, 0 for next to no change, ±5 for the largest. */
export type GainTint = -5 | -4 | -3 | -2 | -1 | 0 | 1 | 2 | 3 | 4 | 5;

/** The gains in percent a table spans: its lowest, 0 at most, and its highest, 0 at least. */
export interface GainScale {
	low: number;
	high: number;
}

const TINT_STEPS = 5;
// A gain in percent below this, either way, tints neutral.
const NO_CHANGE_PCT = 0.05;

// Each fifth of the table's largest gain either way is one step: ±1 up to a fifth of it, ±5 past four fifths. The steps
// are counted by products, as a quotient such as 0.6 * 5 lands past 3 and would take the next step.
export const gainTint = (gainPct: number, { low, high }: GainScale): GainTint => {
	const size = Math.abs(gainPct);
	if (size < NO_CHANGE_PCT) return 0;
	const largest = Math.max(-low, high);
	const steps = Array.from({ length: TINT_STEPS }, (_, step) => step).filter(step => size * TINT_STEPS > step * largest).length;
	return (Math.sign(gainPct) * steps) as GainTint;
};

export interface MatrixCell {
	result: TopGearResult;
	/** DPS over the current gear's, in percent of the current gear's DPS. */
	gainPct: number;
	/** The set's place in the encounter's ranking, 1 for its best; null on the current gear's row. */
	rank: number | null;
	tint: GainTint;
	/** The encounter's highest DPS, the current gear's included. */
	best: boolean;
}

export interface MatrixRow {
	key: string;
	/** Null for the current gear's row. */
	set: ComparedSet | null;
	talents: BulkTalentSetup;
	/** One per column; null where the encounter did not rank the set. */
	cells: Array<MatrixCell | null>;
}

interface CompareMatrix {
	/** One per encounter by duration, shortest first; health-based ones last, in run order. */
	columns: BulkEncounterResults[];
	/** The current gear's row, then every set's by number. */
	rows: MatrixRow[];
	/** The span every cell's tint is measured against, one for the whole table. */
	scale: GainScale;
	current: Gear | null;
	/** Some encounter ran more than one talent setup, so every row names its own. */
	showTalents: boolean;
}

// A set is a gear set with the talent setup it ran with, so the current gear with other talents is a set of its own. It
// is numbered where it first shows, reading the columns in order and each one's ranking top down, and keeps that number
// in every other column that ranks it.
export const compareMatrix = (entries: readonly BulkEncounterResults[]): CompareMatrix => {
	const reference = entries[0]?.results.originalGearResults ?? null;
	const current = reference?.gear ?? null;
	const currentKey = reference ? bulkResultKey(reference) : '';
	const columns = [...entries].sort((a, b) => durationOrder(a) - durationOrder(b));
	const sets = new Map<string, ComparedSet>();
	const rankings = columns.map(({ results: { chains, originalGearResults } }) => {
		const results = chains.flat();
		let rank = 0;
		const byKey = new Map<string, Pick<MatrixCell, 'result' | 'rank'>>(
			results.map(result => {
				const key = bulkResultKey(result);
				const isCurrent = key === currentKey;
				if (!isCurrent && !sets.has(key)) sets.set(key, { key, gear: result.gear, talents: result.talents, number: sets.size + 1 });
				return [key, { result, rank: isCurrent ? null : ++rank }];
			}),
		);
		return { base: originalGearResults.dpsMetrics.avg, top: Math.max(...results.map(result => result.dpsMetrics.avg)), byKey };
	});
	const rowOf = (key: string, set: ComparedSet | null, talents: BulkTalentSetup) => ({
		key,
		set,
		talents,
		cells: rankings.map(({ base, top, byKey }) => {
			const ranked = byKey.get(key);
			if (!ranked) return null;
			const { avg } = ranked.result.dpsMetrics;
			return { ...ranked, gainPct: percentOf(avg - base, base), best: avg === top };
		}),
	});
	const rows = [...(reference ? [rowOf(currentKey, null, reference.talents)] : []), ...[...sets.values()].map(set => rowOf(set.key, set, set.talents))];
	const gains = rows.flatMap(row => row.cells.flatMap(cell => (cell ? [cell.gainPct] : [])));
	const scale = { low: Math.min(0, ...gains), high: Math.max(0, ...gains) };
	const tinted = (cell: Omit<MatrixCell, 'tint'> | null): MatrixCell | null => cell && { ...cell, tint: gainTint(cell.gainPct, scale) };
	const showTalents = entries.some(entry => entry.results.talentSetups > 1);
	return { columns, rows: rows.map((row): MatrixRow => ({ ...row, cells: row.cells.map(tinted) })), scale, current, showTalents };
};
