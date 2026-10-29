import { ItemSlot, Race } from '@generated/proto/common';
import { FactionFilter, SimSettings, UIItem as Item, UIItem_FactionRestriction } from '@generated/proto/ui';
import { fromJsonStringLenient } from '@sim/proto/lenient_json';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const ALLIANCE = Item.create({ id: 1, factionRestriction: UIItem_FactionRestriction.ALLIANCE_ONLY });
const HORDE = Item.create({ id: 2, factionRestriction: UIItem_FactionRestriction.HORDE_ONLY });
const BOTH = Item.create({ id: 3 });

describe('Player.filterItemData faction filter', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		player.setRace(Race.RaceOrc);
		const listed = () => player.filterItemData([ALLIANCE, HORDE, BOTH], item => item, ItemSlot.ItemSlotTrinket1).map(item => item.id);
		const pick = (factionFilter: FactionFilter) => sim.setFilters({ ...sim.getFilters(), factionFilter });
		return { sim, player, listed, pick };
	};

	it("keeps the race's faction by default and follows a race change", async () => {
		const { player, listed } = await setup();
		expect(listed()).toEqual([2, 3]);

		player.setRace(Race.RaceHighOrderSkyborne);
		expect(listed()).toEqual([1, 3]);
	});

	it('keeps every item on any, through a race change, until the race is picked again', async () => {
		const { player, listed, pick } = await setup();
		pick(FactionFilter.FactionFilterAny);
		expect(listed()).toEqual([1, 2, 3]);

		player.setRace(Race.RaceHuman);
		expect(listed()).toEqual([1, 2, 3]);

		pick(FactionFilter.FactionFilterRace);
		expect(listed()).toEqual([1, 3]);

		pick(99 as FactionFilter);
		expect(listed()).toEqual([1, 2, 3]);
	});

	it('saves a pick with the other filters, and restoring the defaults follows the race again', async () => {
		const { sim, listed, pick } = await setup();
		pick(FactionFilter.FactionFilterAny);

		sim.fromProto(fromJsonStringLenient(SimSettings, SimSettings.toJsonString(sim.toProto())));
		expect(listed()).toEqual([1, 2, 3]);

		sim.applyDefaults(false, false);
		expect(listed()).toEqual([2, 3]);
	});
});
