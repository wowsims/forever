import { ItemSlot } from '@generated/proto/common';
import { DropSource, DungeonDifficulty, RaidFilterOption, SourceFilterOption, UIItem as Item, UIItemSource } from '@generated/proto/ui';
import { Sim } from '@sim/sim';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { Player } from './player';

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
const UNSORTED_DROP = Item.create({ id: 8, sources: [UIItemSource.create({ source: { oneofKind: 'drop', drop: DropSource.create() } })] });

describe('Player.filterItemData source filter', () => {
	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		const listed = () =>
			player
				.filterItemData(
					[NO_SOURCE, QUEST, QUEST_AND_VENDOR, RAID, DUNGEON, PVP_ONLY, PVP_VENDOR, UNSORTED_DROP],
					item => item,
					ItemSlot.ItemSlotTrinket1,
				)
				.map(item => item.id);
		const tick = (...sources: SourceFilterOption[]) => sim.setFilters({ ...sim.getFilters(), sources });
		return { listed, tick };
	};

	it('keeps every item while every source is ticked', async () => {
		const { listed } = await setup();
		expect(listed()).toEqual([1, 2, 3, 4, 5, 6, 7, 8]);
	});

	it('keeps an item while one of its sources is ticked', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceSoldBy, SourceFilterOption.SourcePvp);
		expect(listed()).toEqual([1, 3, 6, 7, 8]);

		tick(SourceFilterOption.SourceQuest);
		expect(listed()).toEqual([1, 2, 3, 8]);
	});

	it('sorts a drop into Raid or Dungeon by its difficulty, and no toggle hides a drop with neither', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceRaid);
		expect(listed()).toEqual([1, 4, 8]);

		tick(SourceFilterOption.SourceDungeon);
		expect(listed()).toEqual([1, 5, 8]);
	});

	it('hides every PvP item when PvP is unticked, whatever its other sources', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceSoldBy, SourceFilterOption.SourceQuest, SourceFilterOption.SourceRaid, SourceFilterOption.SourceDungeon);
		expect(listed()).toEqual([1, 2, 3, 4, 5, 8]);
	});
});

describe('Player.filterItemData raid toggles', () => {
	const MC_ZONE_ID = Player.RAID_IDS[RaidFilterOption.RaidMC];
	const mcDrop = UIItemSource.create({
		source: { oneofKind: 'drop', drop: DropSource.create({ difficulty: DungeonDifficulty.DifficultyRaid, zoneId: MC_ZONE_ID }) },
	});
	const MC_ONLY = Item.create({ id: 10, sources: [mcDrop] });
	const MC_AND_QUEST = Item.create({ id: 11, sources: [mcDrop, quest] });
	const mcZoneDrop = UIItemSource.create({ source: { oneofKind: 'drop', drop: DropSource.create({ zoneId: MC_ZONE_ID }) } });
	const MC_ZONE_ONLY = Item.create({ id: 12, sources: [mcZoneDrop] });

	const setup = async () => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		const listed = () => player.filterItemData([MC_ONLY, MC_AND_QUEST, MC_ZONE_ONLY], item => item, ItemSlot.ItemSlotTrinket1).map(item => item.id);
		const tickMC = (ticked: boolean) =>
			sim.setFilters({ ...sim.getFilters(), raids: Sim.ALL_RAIDS.filter(raid => ticked || raid != RaidFilterOption.RaidMC) });
		const tick = (sources: SourceFilterOption[], raids: RaidFilterOption[]) => sim.setFilters({ ...sim.getFilters(), sources, raids });
		return { listed, tickMC, tick };
	};

	it('hides an item that only drops in an unticked raid and shows it when the raid is ticked', async () => {
		const { listed, tickMC } = await setup();
		tickMC(false);
		expect(listed()).toEqual([11]);

		tickMC(true);
		expect(listed()).toEqual([10, 11, 12]);
	});

	it('hides a raid drop while Raid is unticked, though its raid is ticked', async () => {
		const { listed, tick } = await setup();
		tick([SourceFilterOption.SourceQuest], [RaidFilterOption.RaidMC]);
		expect(listed()).toEqual([11, 12]);
	});

	it('hides a drop with no difficulty in an unticked raid and shows it in a ticked one', async () => {
		const { listed, tickMC } = await setup();
		tickMC(false);
		expect(listed()).toEqual([11]);

		tickMC(true);
		expect(listed()).toEqual([10, 11, 12]);
	});
});
