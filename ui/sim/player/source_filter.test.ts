import { ItemSlot } from '@generated/proto/common';
import { type DatabaseFilters, DropSource, DungeonDifficulty, RaidFilterOption, SourceFilterOption, UIItem as Item, UIItemSource } from '@generated/proto/ui';
import { Sim } from '@sim/sim';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { Player } from './player';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const quest = UIItemSource.create({ source: { oneofKind: 'quest', quest: { id: 1, name: 'A Quest', faction: 0 } } });
const soldBy = UIItemSource.create({ source: { oneofKind: 'soldBy', soldBy: { npcId: 2, npcName: 'A Vendor', zoneId: 0, faction: 0 } } });
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

	it('sorts a drop into Raid or Dungeon by its difficulty, and a drop with neither stays while World Drop is ticked', async () => {
		const { listed, tick } = await setup();
		tick(SourceFilterOption.SourceRaid);
		expect(listed()).toEqual([1, 4, 8]);

		tick(SourceFilterOption.SourceDungeon);
		expect(listed()).toEqual([1, 5, 8]);
	});

	describe('the World Drop toggle', () => {
		const worldDrop = UIItemSource.create({ source: { oneofKind: 'drop', drop: DropSource.create({ otherName: 'World Drop' }) } });
		const openWorldNpc = UIItemSource.create({ source: { oneofKind: 'drop', drop: DropSource.create({ npcId: 4494, zoneId: 28 }) } });
		const zoneDrop = UIItemSource.create({
			source: {
				oneofKind: 'drop',
				drop: DropSource.create({ zoneId: 721, otherName: 'Multiple NPCs', difficulty: DungeonDifficulty.DifficultyDungeon }),
			},
		});
		const crafted = UIItemSource.create({ source: { oneofKind: 'crafted', crafted: { profession: 0, spellId: 1 } } });
		const items = [
			NO_SOURCE,
			Item.create({ id: 20, sources: [worldDrop] }),
			Item.create({ id: 21, sources: [openWorldNpc] }),
			Item.create({ id: 22, sources: [crafted, openWorldNpc] }),
			Item.create({ id: 23, sources: [zoneDrop] }),
		];
		const setupToggle = async () => {
			const { sim, player } = await makeIndividualSim();
			sim.applyDefaults(false, false);
			const set = (edit: Partial<DatabaseFilters>) => sim.setFilters({ ...sim.getFilters(), ...edit });
			return { sim, set, listed: () => player.filterItemData(items, item => item, ItemSlot.ItemSlotTrinket1).map(item => item.id) };
		};

		it('is on by default', async () => {
			const { listed } = await setupToggle();
			expect(listed()).toEqual([1, 20, 21, 22, 23]);
		});

		it('hides a world drop and an open-world NPC when off, and keeps what another ticked source or no source shows', async () => {
			const { listed, set } = await setupToggle();
			set({ hideWorldDrops: true });
			expect(listed()).toEqual([1, 22, 23]);
		});

		it('hides an item only when every source is off', async () => {
			const { sim, listed, set } = await setupToggle();
			const withoutCrafting = sim.getFilters().sources.filter(source => source !== SourceFilterOption.SourceCrafting);
			set({ sources: withoutCrafting });
			expect(listed()).toContain(22);

			set({ sources: withoutCrafting, hideWorldDrops: true });
			expect(listed()).not.toContain(22);
		});

		it('leaves a zone drop to Dungeon or Raid', async () => {
			const { sim, listed, set } = await setupToggle();
			set({ sources: sim.getFilters().sources.filter(source => source !== SourceFilterOption.SourceDungeon) });
			expect(listed()).toEqual([1, 20, 21, 22]);
		});

		it('reads as on from filters saved with a partial sources list and no setting of its own', async () => {
			const { sim, listed } = await setupToggle();
			const settings = sim.toProto();
			settings.filters!.sources = [SourceFilterOption.SourceQuest];
			sim.fromProto(settings);
			expect(sim.getFilters().hideWorldDrops).toBe(false);
			expect(listed()).toEqual([1, 20, 21, 22]);
		});
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
