import { DistributionMetrics } from '@generated/proto/api';
import { ItemSlot, ItemType } from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { describe, expect, it } from 'vitest';

import { bulkComparison, changedItemLabels, compareMatrix, encounterRanks, gainTint, type MatrixRow } from './compare';

const equipped = (id: number, enchant?: UIEnchant) =>
	new EquippedItem({ item: UIItem.create({ id, name: `Item ${id}`, type: ItemType.ItemTypeHead, scalingOptions: { 0: { ilvl: 60 } } }), enchant });

// A set wearing `head` and, when given, `hands`; two sets built from the same ids are the same set to the gear identity key.
const gearSet = (head: number, hands?: EquippedItem) => new Gear({ [ItemSlot.ItemSlotHead]: equipped(head), [ItemSlot.ItemSlotHands]: hands ?? null });

const result = (gear: Gear, avg: number): TopGearResult => ({ gear, dpsMetrics: DistributionMetrics.create({ avg, stdev: 1 }) });

const CURRENT = gearSet(1);
const A = gearSet(2);
const B = gearSet(3);
const C = gearSet(4);

// An encounter ranking `ranked` and the current gear at 1000 DPS, best first, as the run stores it.
const entry = (name: string, duration: number | null, ranked: Array<[Gear, number]>): BulkEncounterResults => {
	const base = result(CURRENT, 1000);
	const chains = [...ranked.map(([gear, avg]) => result(gear, avg)), base].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg).map(ranking => [ranking]);
	return {
		name,
		isCurrent: false,
		durationLabel: '',
		duration,
		durationVariation: duration === null ? null : 5,
		results: { chains, originalGearResults: base, iterations: 100 },
	};
};

const ENTRIES = [
	entry('Patchwerk', null, [
		[C, 1100],
		[A, 1050],
	]),
	entry('Current', 180, [
		[B, 1030],
		[gearSet(2), 990],
	]),
	entry('Council', 90, [[A, 1020]]),
];

describe('bulkComparison', () => {
	it('orders the encounters by duration, health-based last, each with its own ranking and the current gear in it', () => {
		const { encounters } = bulkComparison(ENTRIES);

		expect(encounters.map(encounter => encounter.entry.name)).toEqual(['Council', 'Current', 'Patchwerk']);
		expect(encounters.map(encounter => encounter.bars.map(bar => [bar.dpsMetrics.avg, bar.gain]))).toEqual([
			[
				[1020, 20],
				[1000, 0],
			],
			[
				[1030, 30],
				[1000, 0],
				[990, -10],
			],
			[
				[1100, 100],
				[1050, 50],
				[1000, 0],
			],
		]);
	});

	it('numbers each set where it first shows and keeps the number in every encounter, leaving the current gear out', () => {
		const { encounters, sets } = bulkComparison(ENTRIES);
		const numberOf = (key: string) => sets.find(set => set.key === key)?.number;

		expect(encounters.map(encounter => encounter.bars.map(bar => (bar.isCurrent ? 'current' : numberOf(bar.key))))).toEqual([
			[1, 'current'],
			[2, 'current', 1],
			[3, 1, 'current'],
		]);
		expect(sets.map(set => [set.number, set.gear.getEquippedItem(ItemSlot.ItemSlotHead)!.id])).toEqual([
			[1, 2],
			[2, 3],
			[3, 4],
		]);
	});
});

describe('encounterRanks', () => {
	it('ranks the sets an encounter ranked from 1, leaving the current gear out', () => {
		const [, current] = bulkComparison(ENTRIES).encounters;

		expect([...encounterRanks(current).values()]).toEqual([1, 2]);
	});
});

describe('gainTint', () => {
	it('tints next to no change neutral, then one step for each fifth of the table’s largest gain', () => {
		const scale = { low: 0, high: 10 };
		const tintOf = (gainPct: number) => gainTint(gainPct, scale);

		expect([0, 0.04, -0.04].map(tintOf)).toEqual([0, 0, 0]);
		expect([0.05, 2, 2.01, 6, 6.01, 8, 8.01, 10].map(tintOf)).toEqual([1, 1, 2, 3, 4, 4, 5, 5]);
	});

	it('measures losses and gains against the largest of either, so a smaller side never reaches the darkest step', () => {
		const scale = { low: -4, high: 2 };
		const tintOf = (gainPct: number) => gainTint(gainPct, scale);

		expect([-4, -1, 1, 2].map(tintOf)).toEqual([-5, -2, 2, 3]);
	});

	it('tints every cell neutral where no set changes anything', () => {
		expect(gainTint(0, { low: 0, high: 0 })).toBe(0);
	});
});

describe('compareMatrix', () => {
	const cellsOf = (row: MatrixRow) => row.cells.map(cell => (cell ? [cell.gainPct, cell.tint, cell.best] : null));

	it('puts the current gear first, then every set by number, with a gap where an encounter did not rank the set', () => {
		const { rows } = compareMatrix(bulkComparison(ENTRIES));

		expect(rows.map(row => row.set?.number ?? 'current')).toEqual(['current', 1, 2, 3]);
		expect(rows.map(cellsOf)).toEqual([
			[
				[0, 0, false],
				[0, 0, false],
				[0, 0, false],
			],
			[
				[2, 1, true],
				[-1, -1, false],
				[5, 3, false],
			],
			[null, [3, 2, true], null],
			[null, null, [10, 5, true]],
		]);
	});

	it('tints the whole table on one scale, from its lowest gain to its highest, 0 included', () => {
		expect(compareMatrix(bulkComparison(ENTRIES)).scale).toEqual({ low: -1, high: 10 });

		const gainsOnly = compareMatrix(bulkComparison([entry('Short', 60, [[A, 1100]]), entry('Long', 300, [[B, 1040]])]));

		expect(gainsOnly.scale).toEqual({ low: 0, high: 10 });
		expect(gainsOnly.rows.map(cellsOf)).toEqual([
			[
				[0, 0, false],
				[0, 0, false],
			],
			[[10, 5, true], null],
			[null, [4, 2, true]],
		]);
	});

	it('marks the current gear best where every set the encounter ranked does worse', () => {
		const [current, set] = compareMatrix(bulkComparison([entry('Solo', 60, [[A, 990]])])).rows;

		expect([current.cells[0]!.best, set.cells[0]!.best]).toEqual([true, false]);
	});
});

describe('changedItemLabels', () => {
	it('names the items worn where the current gear wears something else, an enchant change included', () => {
		const crusader = UIEnchant.create({ effectId: 1900, name: 'Crusader' });
		const current = gearSet(1, equipped(7));

		expect(changedItemLabels(gearSet(1, equipped(7)), current)).toEqual([]);
		expect(changedItemLabels(gearSet(2, equipped(7, crusader)), current)).toEqual(['Item 2', 'Item 7 (Crusader)']);
	});
});
