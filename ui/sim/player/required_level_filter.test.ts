import { ItemSlot, ScalingItemProperties } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const LEVEL_40 = Item.create({ id: 2, reqLvl: 40 });
const LEVEL_60 = Item.create({ id: 3, reqLvl: 60 });
// No requirement in the game files, so the item level places it.
const QUEST_ILVL_60 = Item.create({ id: 4, scalingOptions: { 0: ScalingItemProperties.create({ ilvl: 60 }) } });
const QUEST_ILVL_34 = Item.create({ id: 5, scalingOptions: { 0: ScalingItemProperties.create({ ilvl: 34 }) } });

describe('Player.filterItemData required level filter', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		const listed = () =>
			player.filterItemData([LEVEL_40, LEVEL_60, QUEST_ILVL_60, QUEST_ILVL_34], item => item, ItemSlot.ItemSlotTrinket1).map(item => item.id);
		const pick = (maxRequiredLevel: number) => sim.setFilters({ ...sim.getFilters(), maxRequiredLevel });
		return { listed, pick };
	};

	it('keeps every item until a level is set', async () => {
		const { listed } = await setup();
		expect(listed()).toEqual([2, 3, 4, 5]);
	});

	it('hides items that need a higher level', async () => {
		const { listed, pick } = await setup();
		pick(40);
		expect(listed()).toEqual([2, 5]);

		pick(29);
		expect(listed()).toEqual([5]);
	});

	it('places an item with no required level by its item level', async () => {
		const { listed, pick } = await setup();
		pick(59);
		expect(listed()).not.toContain(4);

		pick(60);
		expect(listed()).toContain(4);

		pick(29);
		expect(listed()).toContain(5);
		pick(28);
		expect(listed()).not.toContain(5);
	});
});
