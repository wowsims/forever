import { ItemSlot } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());
afterEach(() => vi.restoreAllMocks());

const NOW_MS = Date.UTC(2026, 10, 11, 12);
const DAY_SECONDS = 24 * 60 * 60;
const FIRST_SEEN: Record<number, number | undefined> = {
	2: NOW_MS / 1000 - 3 * DAY_SECONDS,
	3: NOW_MS / 1000 - 8 * DAY_SECONDS,
	4: undefined,
};

describe('Player.filterItemData newly discovered filter', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		vi.spyOn(Date, 'now').mockReturnValue(NOW_MS);
		vi.spyOn(sim.db, 'getItemFirstSeen').mockImplementation(id => FIRST_SEEN[id]);
		const listed = () =>
			player
				.filterItemData(
					[2, 3, 4].map(id => Item.create({ id })),
					item => item,
					ItemSlot.ItemSlotTrinket1,
				)
				.map(item => item.id);
		return { sim, listed };
	};

	it('keeps every item while the filter is off', async () => {
		const { listed } = await setup();
		expect(listed()).toEqual([2, 3, 4]);
	});

	it('shows only the items first seen within the last week, and none with no date', async () => {
		const { sim, listed } = await setup();
		sim.setFilters({ ...sim.getFilters(), newlyDiscoveredOnly: true });
		expect(listed()).toEqual([2]);
	});
});
