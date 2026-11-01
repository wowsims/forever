import { type ItemSlot, ItemSpec } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import type {
	BulkEncounterResults,
	BulkGearSetup,
	BulkPlayerSettingsSetup,
	BulkRankedResult,
	BulkRotationSetup,
	BulkTalentSetup,
	TopGearResult,
} from '@sim/bulk/types';
import type { Gear } from '@sim/proto/gear';

import { getSwappableItemSlotPair, itemSpecPairsEqualUnordered } from '../gear/swappable_slots';
import { formatBulkItemLabel } from './format';
import { bulkResultKey } from './results';

interface ComparedSet {
	key: string;
	gear: Gear;
	talents: BulkTalentSetup;
	rotation: BulkRotationSetup;
	gearSetup: BulkGearSetup | null;
	settings: BulkPlayerSettingsSetup | null;
	number: number;
}

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

// The enchants a Builds set wears where its gear set's own gear wears another or none, in slot order.
const changedEnchants = (gear: Gear, gearSetGear: Gear): Array<{ slot: ItemSlot; name: string }> =>
	gear.getItemSlots().flatMap(slot => {
		const enchant = gear.getEquippedItem(slot)?.enchant;
		const own = gearSetGear.getEquippedItem(slot)?.enchant;
		return enchant && enchant.effectId !== own?.effectId ? [{ slot, name: enchant.name }] : [];
	});

// Each as "Slot: Enchant"; none for a row with no gear setup.
export const changedEnchantLabels = ({ gear, gearSetup }: Pick<BulkRankedResult, 'gear' | 'gearSetup'>): string[] =>
	gearSetup?.gearSetGear ? changedEnchants(gear, gearSetup.gearSetGear).map(({ slot, name }) => `${translateSlotName(slot)}: ${name}`) : [];

const percentOf = (value: number, of: number) => (of ? (value / of) * 100 : 0);

/** A matrix cell's tint: the gain's sign and its size against its encounter's largest, 0 for next to no change, ±5 for the largest. */
export type GainTint = -5 | -4 | -3 | -2 | -1 | 0 | 1 | 2 | 3 | 4 | 5;

/** The gains in percent an encounter's column spans: its lowest, 0 at most, and its highest, 0 at least. */
export interface GainScale {
	low: number;
	high: number;
}

const TINT_STEPS = 5;
// A gain in percent below this, either way, tints neutral.
const NO_CHANGE_PCT = 0.05;

// Each fifth of the encounter's largest gain either way is one step: ±1 up to a fifth of it, ±5 past four fifths. The steps
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
	/** DPS over the reference's, in percent of its DPS: the current gear's, or the encounter's best set's when the batch did not sim the current setup. */
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
	rotation: BulkRotationSetup;
	gearSetup: BulkGearSetup | null;
	settings: BulkPlayerSettingsSetup | null;
	/** One per column; null where the encounter did not rank the set. */
	cells: Array<MatrixCell | null>;
}

interface CompareMatrix {
	/** One per encounter by duration, shortest first, equal lengths in run order. */
	columns: BulkEncounterResults[];
	/** The current gear's row when the batch simmed the current setup, then every set's by number. */
	rows: MatrixRow[];
	current: Gear | null;
}

// A set is a gear set, with its gear setup when it has one, and the talents, rotation and settings it ran with, so the
// same gear under any other of them is a set of its own. It is numbered where it first shows, reading the columns in
// order and each one's ranking top down, and keeps that number in every other column that ranks it. With no current
// gear's row, each column's reference is its best set, so every gain is 0 or a loss.
export const compareMatrix = (entries: readonly BulkEncounterResults[]): CompareMatrix => {
	const first = entries[0]?.results;
	const reference = first?.referenceIsCurrent ? first.originalGearResults : null;
	const current = first?.currentGear ?? null;
	const currentKey = reference ? bulkResultKey(reference) : '';
	const columns = [...entries].sort((a, b) => a.duration - b.duration);
	const sets = new Map<string, ComparedSet>();
	const rankings = columns.map(({ results: { chains, originalGearResults } }) => {
		const results = chains.flat();
		let rank = 0;
		const byKey = new Map<string, Pick<MatrixCell, 'result' | 'rank'>>(
			results.map(result => {
				const key = bulkResultKey(result);
				const isCurrent = key === currentKey;
				if (!isCurrent && !sets.has(key)) {
					const { gear, talents, rotation, gearSetup, settings } = result;
					sets.set(key, { key, gear, talents, rotation, gearSetup, settings, number: sets.size + 1 });
				}
				return [key, { result, rank: isCurrent ? null : ++rank }];
			}),
		);
		return { base: originalGearResults.dpsMetrics.avg, top: Math.max(...results.map(result => result.dpsMetrics.avg)), byKey };
	});
	const rowOf = (
		key: string,
		set: ComparedSet | null,
		{ talents, rotation, gearSetup, settings }: Pick<BulkRankedResult, 'talents' | 'rotation' | 'gearSetup' | 'settings'>,
	) => ({
		key,
		set,
		talents,
		rotation,
		gearSetup,
		settings,
		cells: rankings.map(({ base, top, byKey }) => {
			const ranked = byKey.get(key);
			if (!ranked) return null;
			const { avg } = ranked.result.dpsMetrics;
			return { ...ranked, gainPct: percentOf(avg - base, base), best: avg === top };
		}),
	});
	const rows = [...(reference ? [rowOf(currentKey, null, reference)] : []), ...[...sets.values()].map(set => rowOf(set.key, set, set))];
	// Each encounter's column tints on its own span, so its largest gain takes the deepest tint whatever the others gain.
	const scales = rankings.map((_, index): GainScale => {
		const gains = rows.flatMap(row => row.cells[index]?.gainPct ?? []);
		return { low: Math.min(0, ...gains), high: Math.max(0, ...gains) };
	});
	const tinted = (cell: Omit<MatrixCell, 'tint'> | null, index: number): MatrixCell | null =>
		cell && { ...cell, tint: gainTint(cell.gainPct, scales[index]) };
	return { columns, rows: rows.map((row): MatrixRow => ({ ...row, cells: row.cells.map(tinted) })), current };
};
