import { DistributionMetrics } from '@generated/proto/api';
import { ItemSlot, ItemType } from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { describe, expect, it } from 'vitest';

import { bulkComparison, changedItemLabels, dpsAxisRange } from './compare';

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
const entry = (name: string, duration: number | null, ranked: Array<[Gear, number]>, iterations = 100): BulkEncounterResults => {
	const base = result(CURRENT, 1000);
	const chains = [...ranked.map(([gear, avg]) => result(gear, avg)), base].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg).map(ranking => [ranking]);
	return {
		name,
		isCurrent: false,
		durationLabel: '',
		summaryLabel: '',
		duration,
		durationVariation: duration === null ? null : 5,
		results: { chains, originalGearResults: base, iterations },
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
	entry('Council', 90, [[A, 1020]], 400),
];

describe('bulkComparison', () => {
	it('orders the encounters by duration, health-based last, each with its own ranking and the current gear in it', () => {
		const { encounters } = bulkComparison(ENTRIES);

		expect(encounters.map(encounter => encounter.entry.name)).toEqual(['Council', 'Current', 'Patchwerk']);
		expect(encounters.map(encounter => encounter.bars.map(bar => [bar.dpsMetrics.avg, bar.gain, bar.iterations]))).toEqual([
			[
				[1020, 20, 400],
				[1000, 0, 400],
			],
			[
				[1030, 30, 100],
				[1000, 0, 100],
				[990, -10, 100],
			],
			[
				[1100, 100, 100],
				[1050, 50, 100],
				[1000, 0, 100],
			],
		]);
	});

	it('numbers each set where it first shows and keeps the number in every encounter, leaving the current gear out', () => {
		const { encounters, sets } = bulkComparison(ENTRIES);

		expect(encounters.map(encounter => encounter.bars.map(bar => (bar.isCurrent ? 'current' : bar.number)))).toEqual([
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

describe('dpsAxisRange', () => {
	it('ends at the longest bar and starts where the shortest keeps a third of its length', () => {
		const { min, max } = dpsAxisRange([1000, 1060, 1030]);

		expect([min, max]).toEqual([970, 1060]);
		expect((1000 - min) / (max - min)).toBeCloseTo(1 / 3);
	});

	it('gives a lone value a bar of its own', () => {
		expect(dpsAxisRange([800])).toEqual({ min: 792, max: 800 });
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
