import { ArmorType, Class, Faction, HandType, ItemType, Profession, ScalingItemProperties, WeaponType } from '@generated/proto/common';
import { ContentPhase, DungeonDifficulty, UIItem, UIItem_FactionRestriction, UIItemSource } from '@generated/proto/ui';
import { BulkSimItemSlot } from '@sim/bulk/utils';
import { PlayerSpecs } from '@sim/player/specs';
import { describe, expect, it, vi } from 'vitest';

import {
	dropsDungeonLevels,
	dropsFactionItems,
	type DropsItemContext,
	dropsItemList,
	dropsItemsBySlot,
	dropsItemsBySource,
	dropsLevelItems,
	dropsPhaseItems,
	dropsPickedItemIds,
	dropsUsableItems,
} from './items';
import { dropsItemSourceKeys, type DropsNames, dropsSourceTree } from './sources';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string) => key },
}));

const NAMES: DropsNames = {
	zone: id => ({ 1583: 'Blackrock Spire', 1584: 'Blackrock Depths' })[id] ?? null,
	npc: id => ({ 10363: 'General Drakkisath', 9568: 'Overlord Wyrmthalak' })[id] ?? null,
	faction: () => null,
};

const drop = (zoneId: number, npcId: number) =>
	UIItemSource.create({ source: { oneofKind: 'drop', drop: { difficulty: 0, npcId, zoneId, otherName: '', category: '' } } });
const quest = (faction: Faction) => UIItemSource.create({ source: { oneofKind: 'quest', quest: { id: 1, name: 'q', faction, zoneId: 1583 } } });
const soldBy = (faction: Faction) => UIItemSource.create({ source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'v', faction, zoneId: 1583 } } });
const rep = (factionId: Faction) => UIItemSource.create({ source: { oneofKind: 'rep', rep: { repFactionId: 529, repLevel: 0, factionId } } });
const crafted = () => UIItemSource.create({ source: { oneofKind: 'crafted', crafted: { profession: Profession.Blacksmithing, spellId: 0 } } });

// Every database item carries a base scaling entry, here at the item's own item level.
const item = (fields: Partial<UIItem> & { id: number; name: string }) =>
	UIItem.create({
		type: ItemType.ItemTypeChest,
		armorType: ArmorType.ArmorTypePlate,
		phase: ContentPhase.ContentPhaseNone,
		sources: [drop(1583, 10363)],
		scalingOptions: { 0: ScalingItemProperties.create({ ilvl: fields.ilvl ?? 0 }) },
		...fields,
	});
const weapon = (id: number, name: string, handType: HandType, weaponType = WeaponType.WeaponTypeSword) =>
	item({ id, name, type: ItemType.ItemTypeWeapon, armorType: ArmorType.ArmorTypeUnknown, handType, weaponType });

const DUAL: DropsItemContext = {
	playerSpec: PlayerSpecs.DpsWarrior,
	canDualWield: true,
	phase: ContentPhase.ContentPhase1,
	faction: Faction.Alliance,
	wornIds: new Set(),
};
const ctx = (fields: Partial<DropsItemContext> = {}): DropsItemContext => ({ ...DUAL, ...fields });
const BOTH = new Set(['drop:1583:10363', 'drop:1584:9568']);
const list = (items: UIItem[], context = DUAL, ticked: ReadonlySet<string> = BOTH, excluded: ReadonlySet<number> = new Set()) =>
	dropsItemList(items, ticked, excluded, context, item => dropsItemSourceKeys(item, NAMES));

const TWO_SOURCE = item({ id: 50, name: 'Two Sources', sources: [drop(1583, 10363), drop(1584, 9568)] });

const NO_LEVEL_FILTERS = { minIlvl: 0, maxIlvl: 0, maxRequiredLevel: 0 };
const usable = (items: UIItem[], filters = NO_LEVEL_FILTERS) => dropsUsableItems(items, PlayerSpecs.DpsWarrior, filters).map(i => i.id);

describe('dropsUsableItems', () => {
	it('leaves out what the class cannot equip and random-suffix items', () => {
		const wrongClass = item({ id: 60, name: 'Wrong Class', classAllowlist: [Class.ClassMage] });
		const suffixed = item({ id: 61, name: 'Suffixed', randomSuffixOptions: [1, 2] });
		expect(usable([wrongClass, suffixed, TWO_SOURCE])).toEqual([50]);
	});

	it('leaves out an item that fits no gear slot, such as a toy', () => {
		const toy = item({ id: 62, name: 'Piccolo of the Flaming Fire', type: ItemType.ItemTypeUnknown });
		expect(usable([toy, TWO_SOURCE])).toEqual([50]);
	});

	it('applies the gear picker’s item level and required level bounds, 0 leaving one unset', () => {
		const low = item({ id: 70, name: 'Low', ilvl: 25, reqLvl: 20 });
		const high = item({ id: 71, name: 'High', ilvl: 66, reqLvl: 60 });
		const scaled = item({ id: 72, name: 'Scaled', ilvl: 10, reqLvl: 60, scalingOptions: { 0: ScalingItemProperties.create({ ilvl: 70 }) } });
		const all = [low, high, scaled];
		expect(usable(all)).toEqual([70, 71, 72]);
		expect(usable(all, { ...NO_LEVEL_FILTERS, minIlvl: 30 })).toEqual([71, 72]);
		expect(usable(all, { ...NO_LEVEL_FILTERS, maxIlvl: 66 })).toEqual([70, 71]);
		expect(usable(all, { ...NO_LEVEL_FILTERS, maxRequiredLevel: 30 })).toEqual([70]);
	});
});

describe('dropsItemList', () => {
	it('lists an item once when two of its sources are ticked, with both keys', () => {
		const result = list([TWO_SOURCE]);
		expect(result.items).toHaveLength(1);
		expect(result.items[0].sourceKeys).toEqual(['drop:1583:10363', 'drop:1584:9568']);
	});

	it('lists an item with one ticked source of two under that source only', () => {
		const bySource = dropsItemsBySource(list([TWO_SOURCE], DUAL, new Set(['drop:1583:10363'])), dropsSourceTree([TWO_SOURCE], NAMES), NAMES);
		expect(bySource.map(group => group.key)).toEqual(['drop:1583:10363']);
	});

	it('skips an item none of whose sources is ticked', () => {
		expect(list([TWO_SOURCE], DUAL, new Set(['crafted:1'])).items).toEqual([]);
	});

	it('sorts the list by item name', () => {
		const result = list([item({ id: 1, name: 'Zed' }), item({ id: 2, name: 'Abe' })]);
		expect(result.items.map(i => i.item.name)).toEqual(['Abe', 'Zed']);
	});

	it('greys out an item worn in either ring slot and does not pick it', () => {
		const ring = item({ id: 70, name: 'Ring', type: ItemType.ItemTypeFinger, armorType: ArmorType.ArmorTypeUnknown });
		const result = list([ring], ctx({ wornIds: new Set([70]) }));
		expect(result.items[0].worn).toBe(true);
		expect(dropsPickedItemIds(result)).not.toContain(ring.id);
	});

	it('brings a worn item back ticked once it is taken off, unless the user unticked it', () => {
		const a = item({ id: 71, name: 'A' });
		const b = item({ id: 72, name: 'B' });
		const worn = list([a, b], ctx({ wornIds: new Set([71, 72]) }), BOTH, new Set([72]));
		expect(dropsPickedItemIds(worn)).toEqual([]);
		const off = list([a, b], ctx(), BOTH, new Set([72]));
		expect(off.items.map(i => [i.item.id, i.included])).toEqual([
			[71, true],
			[72, false],
		]);
		expect(dropsPickedItemIds(off)).toEqual([71]);
	});

	it('picks the included items sorted by id', () => {
		const items = [item({ id: 330, name: 'A' }), item({ id: 101, name: 'B' }), item({ id: 205, name: 'C' }), item({ id: 99, name: 'D' })];
		expect(dropsPickedItemIds(list(items, ctx(), BOTH, new Set([99])))).toEqual([101, 205, 330]);
	});

	it('tags a dual wielder’s one-hander for both hands and a two-hander as 2H', () => {
		const tags = (i: UIItem, context = DUAL) => list([i], context).items[0].tags;
		expect(tags(weapon(1, 'One', HandType.HandTypeOneHand))).toEqual(['mh', 'oh']);
		expect(tags(weapon(2, 'Main', HandType.HandTypeMainHand))).toEqual(['mh']);
		expect(tags(weapon(3, 'Off', HandType.HandTypeOffHand))).toEqual(['oh']);
		expect(tags(weapon(4, 'Two', HandType.HandTypeTwoHand))).toEqual(['2h']);
		const ring = item({ id: 5, name: 'Ring', type: ItemType.ItemTypeFinger, armorType: ArmorType.ArmorTypeUnknown });
		const trinket = item({ id: 6, name: 'Trinket', type: ItemType.ItemTypeTrinket, armorType: ArmorType.ArmorTypeUnknown });
		expect(tags(ring)).toEqual([]);
		expect(tags(trinket)).toEqual([]);
		expect(tags(item({ id: 7, name: 'Chest' }))).toEqual([]);
	});

	it('tags a one-hander for the main hand only when the player cannot dual wield', () => {
		const single = ctx({ canDualWield: false });
		expect(list([weapon(1, 'One', HandType.HandTypeOneHand)], single).items[0].tags).toEqual(['mh']);
	});

	it('groups by slot in slot order, weapons together for a dual wielder', () => {
		const groups = dropsItemsBySlot(
			list([weapon(1, 'Sword', HandType.HandTypeOneHand), item({ id: 2, name: 'Chest' }), weapon(3, 'Two', HandType.HandTypeTwoHand)]),
		);
		expect(groups.map(g => [g.slot, g.items.map(i => i.item.id)])).toEqual([
			[BulkSimItemSlot.ItemSlotChest, [2]],
			[BulkSimItemSlot.ItemSlotHandWeapon, [1, 3]],
		]);
	});

	it('groups by source in tree order and lists an item under each ticked source', () => {
		const items = [TWO_SOURCE, item({ id: 51, name: 'Another' })];
		const groups = dropsItemsBySource(list(items), dropsSourceTree(items, NAMES), NAMES);
		expect(groups.map(g => [g.key, g.label.label, g.items.map(i => i.item.id)])).toEqual([
			['drop:1584:9568', 'Overlord Wyrmthalak', [50]],
			['drop:1583:10363', 'General Drakkisath', [51, 50]],
		]);
	});
});

describe('dropsPhaseItems', () => {
	it('lists only items at or below the picked phase', () => {
		const P1 = item({ id: 1, name: 'P1', phase: ContentPhase.ContentPhase1 });
		const UNKNOWN = item({ id: 2, name: 'U', phase: ContentPhase.ContentPhaseUnknown });
		const NONE = item({ id: 3, name: 'N', phase: ContentPhase.ContentPhaseNone });
		expect(dropsPhaseItems([P1, UNKNOWN, NONE], ContentPhase.ContentPhase1).map(i => i.id)).toEqual([P1.id, NONE.id]);
	});
});

describe('dropsFactionItems', () => {
	it('keeps an Alliance player off Horde-only items and Horde quest, vendor and reputation sources', () => {
		const HORDE_ONLY = item({ id: 1, name: 'H', factionRestriction: UIItem_FactionRestriction.HORDE_ONLY });
		const MIXED = item({ id: 2, name: 'M', sources: [quest(Faction.Horde), rep(Faction.Horde), soldBy(Faction.Alliance)] });
		const HORDE_QUEST_ONLY = item({ id: 3, name: 'Q', sources: [quest(Faction.Horde)] });
		const HORDE_REP_ONLY = item({ id: 4, name: 'R', sources: [rep(Faction.Horde)] });
		const NEUTRAL = item({ id: 5, name: 'N', sources: [quest(Faction.Unknown), crafted(), drop(1583, 10363)] });
		const ALLIANCE_ONLY = item({ id: 6, name: 'A', factionRestriction: UIItem_FactionRestriction.ALLIANCE_ONLY });
		expect(
			dropsFactionItems([HORDE_ONLY, MIXED, HORDE_QUEST_ONLY, HORDE_REP_ONLY, NEUTRAL, ALLIANCE_ONLY], Faction.Alliance).map(i => [
				i.id,
				i.sources.length,
			]),
		).toEqual([
			[MIXED.id, 1],
			[NEUTRAL.id, NEUTRAL.sources.length],
			[ALLIANCE_ONLY.id, ALLIANCE_ONLY.sources.length],
		]);
	});
});

describe('dropsDungeonLevels', () => {
	it('rates a dungeon by its highest required level over all its drops and leaves raids and other sources out', () => {
		const dungeon = (zoneId: number) =>
			UIItemSource.create({
				source: { oneofKind: 'drop', drop: { difficulty: DungeonDifficulty.DifficultyDungeon, npcId: 1, zoneId, otherName: '', category: '' } },
			});
		const raid = UIItemSource.create({
			source: { oneofKind: 'drop', drop: { difficulty: DungeonDifficulty.DifficultyRaid, npcId: 2, zoneId: 2717, otherName: '', category: '' } },
		});
		const items = [
			item({ id: 1, name: 'Low', reqLvl: 40, sources: [dungeon(1583)] }),
			item({ id: 2, name: 'High', reqLvl: 58, sources: [dungeon(1583), raid] }),
			item({ id: 3, name: 'Other', reqLvl: 50, sources: [dungeon(1337), quest(Faction.Unknown)] }),
		];
		expect(dropsDungeonLevels(items)).toEqual(
			new Map([
				[1583, 58],
				[1337, 50],
			]),
		);
	});
});

describe('dropsLevelItems', () => {
	const ULDAMAN = 1337;
	const ZUL_FARRAK = 1176;
	const BLACKROCK_SPIRE = 1583;
	const MOLTEN_CORE = 2717;
	const instanceDrop = (difficulty: DungeonDifficulty, zoneId: number, npcId: number) =>
		UIItemSource.create({ source: { oneofKind: 'drop', drop: { difficulty, npcId, zoneId, otherName: '', category: '' } } });
	const dungeon = (zoneId: number, npcId: number) => instanceDrop(DungeonDifficulty.DifficultyDungeon, zoneId, npcId);
	const kept = (items: UIItem[]) => dropsLevelItems(items, dropsDungeonLevels(items)).map(i => [i.id, i.sources]);

	it('leaves out a dungeon whose loot needs at most level 44, an item it shares with a kept source keeping that one', () => {
		const archaedas = item({ id: 1, name: 'Archaedas Loot', reqLvl: 44, sources: [dungeon(ULDAMAN, 2748)] });
		const shared = item({ id: 2, name: 'Shared', reqLvl: 40, sources: [dungeon(ULDAMAN, 6910), dungeon(BLACKROCK_SPIRE, 9568)] });
		const spire = item({ id: 3, name: 'Spire Loot', reqLvl: 58, sources: [dungeon(BLACKROCK_SPIRE, 10363)] });
		expect(kept([archaedas, shared, spire])).toEqual([
			[2, [dungeon(BLACKROCK_SPIRE, 9568)]],
			[3, spire.sources],
		]);
	});

	it('keeps every item of a dungeon whose loot reaches level 45, the low ones too', () => {
		const high = item({ id: 1, name: 'High', reqLvl: 45, sources: [dungeon(ZUL_FARRAK, 7267)] });
		const low = item({ id: 2, name: 'Low', reqLvl: 40, sources: [dungeon(ZUL_FARRAK, 7795)] });
		expect(kept([high, low]).map(([id]) => id)).toEqual([1, 2]);
	});

	it('keeps a world drop, quest reward, crafted, reputation or vendor source only on an item that needs level 45', () => {
		const sources = [drop(0, 0), drop(ULDAMAN, 0), quest(Faction.Unknown), crafted(), rep(Faction.Unknown), soldBy(Faction.Unknown)];
		const items = sources.flatMap((source, index) => [
			item({ id: 10 + index, name: `Low ${index}`, reqLvl: 44, sources: [source] }),
			item({ id: 20 + index, name: `High ${index}`, reqLvl: 45, sources: [source] }),
		]);
		expect(kept(items).map(([id]) => id)).toEqual([20, 21, 22, 23, 24, 25]);
	});

	it('keeps a raid drop whatever level it needs', () => {
		const raid = item({
			id: 1,
			name: 'Raid',
			reqLvl: 30,
			sources: [instanceDrop(DungeonDifficulty.DifficultyRaid, MOLTEN_CORE, 11502), quest(Faction.Unknown)],
		});
		expect(kept([raid])).toEqual([[1, [raid.sources[0]]]]);
	});
});
