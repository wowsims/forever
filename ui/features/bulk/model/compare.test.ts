import { DistributionMetrics } from '@generated/proto/api';
import { ItemSlot, ItemType } from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, BulkRankedResult } from '@sim/bulk/types';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { BULK_COMPARE_TALENT_ENTRIES, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { describe, expect, it } from 'vitest';

import { changedItemLabels, compareMatrix, gainTint, type MatrixRow } from './compare';

const equipped = (id: number, enchant?: UIEnchant) =>
	new EquippedItem({ item: UIItem.create({ id, name: `Item ${id}`, type: ItemType.ItemTypeHead, scalingOptions: { 0: { ilvl: 60 } } }), enchant });

// A set wearing `head` and, when given, `hands`; two sets built from the same ids are the same set to the gear identity key.
const gearSet = (head: number, hands?: EquippedItem) => new Gear({ [ItemSlot.ItemSlotHead]: equipped(head), [ItemSlot.ItemSlotHands]: hands ?? null });

const result = (gear: Gear, avg: number): BulkRankedResult => ({
	gear,
	dpsMetrics: DistributionMetrics.create({ avg, stdev: 1 }),
	talents: CURRENT_TALENT_SETUP,
});

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
		duration,
		durationVariation: duration === null ? null : 5,
		results: { chains, originalGearResults: base, iterations: 100, talentSetups: 1 },
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

describe('compareMatrix', () => {
	const cellsOf = (row: MatrixRow) => row.cells.map(cell => (cell ? [cell.gainPct, cell.tint, cell.best] : null));

	it('orders the columns by duration, health-based last', () => {
		expect(compareMatrix(ENTRIES).columns.map(column => column.name)).toEqual(['Council', 'Current', 'Patchwerk']);
	});

	it('numbers each set where it first shows, reading the columns in order, and keeps the number in every column', () => {
		const { rows } = compareMatrix(ENTRIES);

		expect(rows.map(row => row.set && [row.set.number, row.set.gear.getEquippedItem(ItemSlot.ItemSlotHead)!.id])).toEqual([null, [1, 2], [2, 3], [3, 4]]);
	});

	it('ranks the sets in each column from 1, the current gear taking no rank', () => {
		const { rows } = compareMatrix(ENTRIES);

		expect(rows.map(row => row.cells.map(cell => (cell ? cell.rank : 'gap')))).toEqual([
			[null, null, null],
			[1, 2, 2],
			['gap', 1, 'gap'],
			['gap', 'gap', 1],
		]);
	});

	it('puts the current gear first, then every set by number, with a gap where an encounter did not rank the set', () => {
		const { rows } = compareMatrix(ENTRIES);

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
		expect(compareMatrix(ENTRIES).scale).toEqual({ low: -1, high: 10 });

		const gainsOnly = compareMatrix([entry('Short', 60, [[A, 1100]]), entry('Long', 300, [[B, 1040]])]);

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
		const [current, set] = compareMatrix([entry('Solo', 60, [[A, 990]])]).rows;

		expect([current.cells[0]!.best, set.cells[0]!.best]).toEqual([true, false]);
	});

	it('keys a set by its gear and talents, so the current gear with another setup is a set of its own', () => {
		const { rows } = compareMatrix(BULK_COMPARE_TALENT_ENTRIES);
		const helm = (row: MatrixRow) => (row.set ? row.set.gear.getEquippedItem(ItemSlot.ItemSlotHead)!.id : null);

		expect(rows.map(row => [row.set?.number ?? null, helm(row), row.talents.name])).toEqual([
			[null, null, ''],
			[1, 2, 'Arms'],
			[2, 1, 'Arms'],
			[3, 3, ''],
		]);
		expect(rows[0].talents).toBe(CURRENT_TALENT_SETUP);
		expect(rows.map(row => row.cells.map(cell => (cell ? cell.rank : 'gap')))).toEqual([
			[null, null],
			[1, 2],
			[2, 'gap'],
			[3, 1],
		]);
	});

	it('shows talents only when an encounter ran more than one setup', () => {
		expect(compareMatrix(ENTRIES).showTalents).toBe(false);
		expect(compareMatrix(BULK_COMPARE_TALENT_ENTRIES).showTalents).toBe(true);
		expect(compareMatrix([...ENTRIES, BULK_COMPARE_TALENT_ENTRIES[0]]).showTalents).toBe(true);
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

describe('changedItemLabels', () => {
	const pair = (first: ItemSlot, second: ItemSlot, ids: [number, number]) => new Gear({ [first]: equipped(ids[0]), [second]: equipped(ids[1]) });

	it('names the items worn where the current gear wears something else, an enchant change included', () => {
		const crusader = UIEnchant.create({ effectId: 1900, name: 'Crusader' });
		const current = gearSet(1, equipped(7));

		expect(changedItemLabels(gearSet(1, equipped(7)), current, false)).toEqual([]);
		expect(changedItemLabels(gearSet(2, equipped(7, crusader)), current, false)).toEqual(['Item 2', 'Item 7 (Crusader)']);
	});

	it.each([
		['rings', ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2],
		['trinkets', ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	])('names nothing where the %s only change places', (_, first, second) => {
		expect(changedItemLabels(pair(first, second, [11, 10]), pair(first, second, [10, 11]), false)).toEqual([]);
	});

	it('names a new item in a pair', () => {
		const current = pair(ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2, [10, 11]);

		expect(changedItemLabels(pair(ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2, [10, 12]), current, false)).toEqual(['Item 12']);
	});

	it('takes swapped weapons for no change only for a player who can dual wield', () => {
		const current = pair(ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, [20, 21]);
		const swapped = pair(ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, [21, 20]);

		expect(changedItemLabels(swapped, current, true)).toEqual([]);
		expect(changedItemLabels(swapped, current, false)).toEqual(['Item 21', 'Item 20']);
	});
});
