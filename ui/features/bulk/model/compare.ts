import type { DistributionMetrics } from '@generated/proto/api';
import { ItemSpec } from '@generated/proto/common';
import type { BulkEncounterResults } from '@sim/bulk/types';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';

import { formatBulkItemLabel } from './format';

export interface ComparedBar {
	key: string;
	isCurrent: boolean;
	dpsMetrics: DistributionMetrics;
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
			const { chains, originalGearResults } = entry.results;
			const bars = chains.flat().map((result): ComparedBar => {
				const key = gearKey(result.gear);
				const isCurrent = key === currentKey;
				if (!isCurrent && !sets.has(key)) sets.set(key, { key, gear: result.gear, number: sets.size + 1 });
				return {
					key,
					isCurrent,
					dpsMetrics: result.dpsMetrics,
					gain: result.dpsMetrics.avg - originalGearResults.dpsMetrics.avg,
				};
			});
			return { entry, bars };
		});
	return { encounters, sets: [...sets.values()], current };
};

export const changedItemLabels = (gear: Gear, current: Gear): string[] =>
	gear.getItemSlots().flatMap(slot => {
		const equipped = gear.getEquippedItem(slot);
		const worn = current.getEquippedItem(slot);
		return equipped && !(worn && ItemSpec.equals(equipped.asSpec(), worn.asSpec())) ? [formatBulkItemLabel(equipped)] : [];
	});

// The rank of each set the encounter ranked, 1 for its best; the current gear takes no rank.
export const encounterRanks = (encounter: ComparedEncounter): Map<string, number> =>
	new Map(encounter.bars.filter(bar => !bar.isCurrent).map((bar, index) => [bar.key, index + 1]));

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
	bar: ComparedBar;
	/** DPS over the current gear's, in percent of the current gear's DPS. */
	gainPct: number;
	tint: GainTint;
	/** The encounter's highest DPS, the current gear's included. */
	best: boolean;
}

export interface MatrixRow {
	key: string;
	/** Null for the current gear's row. */
	set: ComparedSet | null;
	/** One per encounter, in `BulkComparison.encounters` order; null where the encounter did not rank the set. */
	cells: Array<MatrixCell | null>;
}

export interface CompareMatrix {
	rows: MatrixRow[];
	/** The span every cell's tint is measured against, one for the whole table. */
	scale: GainScale;
}

// The current gear's row, then every set's by number, with a column for every encounter, health-based ones too.
export const compareMatrix = ({ encounters, sets }: BulkComparison): CompareMatrix => {
	const columns = encounters.map(({ entry, bars }) => ({
		base: entry.results.originalGearResults.dpsMetrics.avg,
		top: Math.max(...bars.map(bar => bar.dpsMetrics.avg)),
		byKey: new Map(bars.map(bar => [bar.key, bar])),
	}));
	const rowOf = (key: string, set: ComparedSet | null) => ({
		key,
		set,
		cells: columns.map(({ base, top, byKey }) => {
			const bar = byKey.get(key);
			if (!bar) return null;
			return { bar, gainPct: percentOf(bar.gain, base), best: bar.dpsMetrics.avg === top };
		}),
	});
	const currentKey = encounters.flatMap(encounter => encounter.bars).find(bar => bar.isCurrent)?.key;
	const rows = [...(currentKey === undefined ? [] : [rowOf(currentKey, null)]), ...sets.map(set => rowOf(set.key, set))];
	const gains = rows.flatMap(row => row.cells.flatMap(cell => (cell ? [cell.gainPct] : [])));
	const scale = { low: Math.min(0, ...gains), high: Math.max(0, ...gains) };
	const tinted = (cell: Omit<MatrixCell, 'tint'> | null): MatrixCell | null => cell && { ...cell, tint: gainTint(cell.gainPct, scale) };
	return { rows: rows.map((row): MatrixRow => ({ ...row, cells: row.cells.map(tinted) })), scale };
};
