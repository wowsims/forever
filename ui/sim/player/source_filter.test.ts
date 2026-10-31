import { ItemSlot } from '@generated/proto/common';
import { DropSource, DungeonDifficulty, SourceFilterOption, UIItem as Item, UIItemSource } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const quest = UIItemSource.create({ source: { oneofKind: 'quest', quest: { id: 1, name: 'A Quest' } } });
const soldBy = UIItemSource.create({ source: { oneofKind: 'soldBy', soldBy: { npcId: 2, npcName: 'A Vendor', zoneId: 0 } } });
const raidDrop = UIItemSource.create({ source: { oneofKind: 'drop', drop: DropSource.create({ difficulty: DungeonDifficulty.DifficultyRaid }) } });
const dungeonDrop = UIItemSource.create({ source: { oneofKind: 'drop', drop: DropSource.create({ difficulty: DungeonDifficulty.DifficultyDungeon }) } });
const pvp = UIItemSource.create({ source: { oneofKind: 'pvp', pvp: {} } });

const NO_SOURCE = Item.create({ id: 1 });
const QUEST = Item.create({ id: 2, sources: [quest] });
const QUEST_AND_VENDOR = Item.create({ id: 3, sources: [quest, soldBy] });
const RAID = Item.create({ id: 4, sources: [raidDrop] });
const DUNGEON = Item.create({ id: 5, sources: [dungeonDrop] });
const PVP_ONLY = Item.create({ id: 6, sources: [pvp] });
const PVP_VENDOR = Item.create({ id: 7, sources: [soldBy, pvp] });

describe('Player.filterItemData source filter', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		const listed = () =>
			player
				.filterItemData([NO_SOURCE, QUEST, QUEST_AND_VENDOR, RAID, DUNGEON, PVP_ONLY, PVP_VENDOR], item => item, ItemSlot.ItemSlotTrinket1)
				.map(item => item.id);
		const tick = (...sources: SourceFilterOption[]) => sim.setFilters({ ...sim.getFilters(), sources });
		return { listed, tick };
	};

	it('keeps every item while every source is ticked', async () => {
		const { listed } = await setup();
		expect(listed()).toEqual([1, 2, 3, 4, 5, 6, 7]);
	});

	it('keeps an item while one of its sources is ticked', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceSoldBy, SourceFilterOption.SourcePvp);
		expect(listed()).toEqual([1, 3, 6, 7]);

		tick(SourceFilterOption.SourceQuest);
		expect(listed()).toEqual([1, 2, 3]);
	});

	it('sorts a drop into Raid or Dungeon by its difficulty', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceRaid);
		expect(listed()).toEqual([1, 4]);

		tick(SourceFilterOption.SourceDungeon);
		expect(listed()).toEqual([1, 5]);
	});

	it('hides every PvP item when PvP is unticked, whatever its other sources', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceSoldBy, SourceFilterOption.SourceQuest, SourceFilterOption.SourceRaid, SourceFilterOption.SourceDungeon);
		expect(listed()).toEqual([1, 2, 3, 4, 5]);
	});
});
