import { ItemRandomSuffix, ItemSlot, Stat } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());
afterEach(() => vi.restoreAllMocks());

const suffixStats = (stats: Partial<Record<Stat, number>>) => Object.assign(new Array<number>(Stat.StatAgility + 10).fill(0), stats);
const SUFFIXES: Record<number, ItemRandomSuffix | undefined> = {
	100: ItemRandomSuffix.create({ id: 100, stats: suffixStats({ [Stat.StatStrength]: 10, [Stat.StatStamina]: 10 }) }),
	101: ItemRandomSuffix.create({ id: 101, stats: suffixStats({ [Stat.StatAgility]: 10 }) }),
};

const ITEMS = [
	// Strength and stamina of its own.
	Item.create({ id: 2, scalingOptions: { 0: { stats: { [Stat.StatStrength]: 12, [Stat.StatStamina]: 8 } } } }),
	// Strength alone.
	Item.create({ id: 3, scalingOptions: { 0: { stats: { [Stat.StatStrength]: 15 } } } }),
	// Nothing of its own; one suffix gives strength and stamina, the other agility.
	Item.create({ id: 4, randomSuffixOptions: [100, 101] }),
	// Stamina of its own, and agility from its only suffix.
	Item.create({ id: 5, scalingOptions: { 0: { stats: { [Stat.StatStamina]: 6 } } }, randomSuffixOptions: [101] }),
];

describe('Player.filterItemData item stats filter', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		vi.spyOn(sim.db, 'getRandomSuffixById').mockImplementation(id => SUFFIXES[id]);
		const listed = (stats: Stat[]) => {
			sim.setFilters({ ...sim.getFilters(), itemStats: stats });
			return player
				.filterItemData(
					ITEMS.map(item => Item.clone(item)),
					item => item,
					ItemSlot.ItemSlotTrinket1,
				)
				.map(item => item.id);
		};
		return { listed };
	};

	it('keeps every item with no stat picked', async () => {
		const { listed } = await setup();
		expect(listed([])).toEqual([2, 3, 4, 5]);
	});

	it('keeps the items that carry every picked stat, by themselves or with one suffix', async () => {
		const { listed } = await setup();
		expect(listed([Stat.StatStrength, Stat.StatStamina])).toEqual([2, 4]);
		expect(listed([Stat.StatStrength])).toEqual([2, 3, 4]);
	});

	it('counts an item whose own stats and a suffix together carry the picks, but not two suffixes together', async () => {
		const { listed } = await setup();
		expect(listed([Stat.StatAgility, Stat.StatStamina])).toEqual([5]);
		expect(listed([Stat.StatAgility, Stat.StatStrength])).toEqual([]);
	});
});
