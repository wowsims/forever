import { AreaType, CreatureFamily, ItemSlot, MobType } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const PLAIN = Item.create({ id: 2 });
const FOREST = Item.create({ id: 3, areaTypes: [AreaType.AreaTypeForestGrassland] });
const DESERT = Item.create({ id: 4, areaTypes: [AreaType.AreaTypeDesert] });
const UNDEAD_AND_DEMON = Item.create({ id: 5, mobTypes: [MobType.MobTypeDemon, MobType.MobTypeUndead] });
const PLANT = Item.create({ id: 6, creatureFamilies: [CreatureFamily.CreatureFamilyPlant] });

describe('Player.filterItemData area, mob type and creature group filters', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		const listed = () =>
			player.filterItemData([PLAIN, FOREST, DESERT, UNDEAD_AND_DEMON, PLANT], item => item, ItemSlot.ItemSlotTrinket1).map(item => item.id);
		const pick = (areaTypes: AreaType[], mobTypes: MobType[], creatureFamilies: CreatureFamily[] = []) =>
			sim.setFilters({ ...sim.getFilters(), areaTypes, mobTypes, creatureFamilies });
		return { listed, pick };
	};

	it('keeps every item while nothing is picked', async () => {
		const { listed } = await setup();
		expect(listed()).toEqual([2, 3, 4, 5, 6]);
	});

	it('shows only the items with an effect in a picked area', async () => {
		const { listed, pick } = await setup();
		pick([AreaType.AreaTypeForestGrassland], []);
		expect(listed()).toEqual([3]);

		pick([AreaType.AreaTypeForestGrassland, AreaType.AreaTypeDesert], []);
		expect(listed()).toEqual([3, 4]);
	});

	it('shows an item when one of its mob types is picked', async () => {
		const { listed, pick } = await setup();
		pick([], [MobType.MobTypeUndead]);
		expect(listed()).toEqual([5]);

		pick([], [MobType.MobTypeBeast]);
		expect(listed()).toEqual([]);
	});

	it('shows only the items with an effect against a picked creature group', async () => {
		const { listed, pick } = await setup();
		pick([], [], [CreatureFamily.CreatureFamilyPlant]);
		expect(listed()).toEqual([6]);
	});

	it('shows the items of every picked kind together', async () => {
		const { listed, pick } = await setup();
		pick([AreaType.AreaTypeDesert], [MobType.MobTypeDemon], [CreatureFamily.CreatureFamilyPlant]);
		expect(listed()).toEqual([4, 5, 6]);
	});
});
