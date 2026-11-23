import { Faction, Profession } from '@generated/proto/common';
import { UIItem, UIItemSource } from '@generated/proto/ui';
import { professionNames } from '@sim/proto/names';
import { describe, expect, it, vi } from 'vitest';

import {
	DROPS_L60_DUNGEON_ZONE_IDS,
	dropsItemSourceKeys,
	type DropsNames,
	dropsNodeState,
	dropsQuickPickTicks,
	dropsSourceLabel,
	type DropsSourceNode,
	dropsSourceTree,
	dropsTicksAfter,
} from './sources';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string) => key },
}));

const ZONES = new Map([
	[1583, 'Blackrock Spire'],
	[1584, 'Blackrock Depths'],
	[2557, 'Dire Maul'],
	[491, 'Razorfen Kraul'],
	[139, 'Eastern Plaguelands'],
]);
const NPCS = new Map([
	[10363, 'General Drakkisath'],
	[9568, 'Overlord Wyrmthalak'],
	[9016, 'Bael’Gar'],
]);
const FACTIONS = new Map([[529, 'Argent Dawn']]);
const NAMES: DropsNames = {
	zone: id => ZONES.get(id) ?? null,
	npc: id => NPCS.get(id) ?? null,
	faction: id => FACTIONS.get(id) ?? null,
};

const drop = (fields: { zoneId?: number; npcId?: number; otherName?: string }) =>
	UIItemSource.create({ source: { oneofKind: 'drop', drop: { difficulty: 0, npcId: 0, zoneId: 0, otherName: '', category: '', ...fields } } });
const quest = (fields: { zoneId: number; faction?: Faction }) =>
	UIItemSource.create({ source: { oneofKind: 'quest', quest: { id: 1, name: 'q', faction: Faction.Unknown, ...fields } } });
const crafted = (profession: Profession) => UIItemSource.create({ source: { oneofKind: 'crafted', crafted: { profession, spellId: 0 } } });
const rep = (repFactionId: number) => UIItemSource.create({ source: { oneofKind: 'rep', rep: { repFactionId, repLevel: 0, factionId: Faction.Unknown } } });
const soldBy = (fields: { zoneId: number }) =>
	UIItemSource.create({ source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'v', faction: Faction.Unknown, ...fields } } });
const pvp = () => UIItemSource.create({ source: { oneofKind: 'pvp', pvp: {} } });

const item = (id: number, ...sources: UIItemSource[]) => UIItem.create({ id, sources });
const treeNode = (key: string, items: UIItem[] = [item(1, quest({ zoneId: 139 }), crafted(Profession.Alchemy))]): DropsSourceNode =>
	dropsSourceTree(items, NAMES).find(node => node.key === key)!;

describe('drops source tree', () => {
	it('keys each source kind as the shared contract names them, and PvP not at all', () => {
		const keys = (source: UIItemSource) => dropsItemSourceKeys(UIItem.create({ id: 1, sources: [source] }), NAMES);
		expect(keys(drop({ zoneId: 1583, npcId: 10363 }))).toEqual(['drop:1583:10363']);
		expect(keys(drop({ zoneId: 491, otherName: 'Trash' }))).toEqual(['drop:491:other:Trash']);
		expect(keys(drop({ otherName: 'World Drop' }))).toEqual(['drop:0:other:World Drop']);
		expect(keys(quest({ zoneId: 139 }))).toEqual(['quest:139']);
		expect(keys(quest({ zoneId: -264 }))).toEqual(['quest:other']);
		expect(keys(quest({ zoneId: 99999 }))).toEqual(['quest:other']);
		expect(keys(crafted(Profession.Blacksmithing))).toEqual([`crafted:${Profession.Blacksmithing}`]);
		expect(keys(rep(529))).toEqual(['rep:529']);
		expect(keys(soldBy({ zoneId: 0 }))).toEqual(['soldBy:0']);
		expect(keys(soldBy({ zoneId: 99999 }))).toEqual(['soldBy:0']);
		expect(keys(soldBy({ zoneId: 491 }))).toEqual(['soldBy:491']);
		expect(keys(pvp())).toEqual([]);
	});

	it('lists each key of an item once, in item order', () => {
		const sources = [quest({ zoneId: 139 }), drop({ zoneId: 1583, npcId: 10363 }), quest({ zoneId: 139 })];
		expect(dropsItemSourceKeys(UIItem.create({ id: 1, sources }), NAMES)).toEqual(['quest:139', 'drop:1583:10363']);
	});

	it('groups drops by zone and boss, with world drops after the zones, and counts each item once', () => {
		const tree = dropsSourceTree(
			[
				item(1, drop({ zoneId: 1583, npcId: 10363 }), drop({ zoneId: 1583, npcId: 9568 })),
				item(2, drop({ zoneId: 1583, otherName: 'Trash' })),
				item(3, drop({ zoneId: 1584, npcId: 9016 })),
				item(4, drop({ otherName: 'World Drop' })),
			],
			NAMES,
		);
		expect(tree.map(node => node.key)).toEqual(['drop']);
		const drops = tree[0];
		expect(drops.children.map(node => node.key)).toEqual(['drop:1584', 'drop:1583', 'drop:0:other:World Drop']);
		const spire = drops.children[1];
		expect(spire.itemIds).toEqual([1, 2]);
		expect(spire.leafKeys).toEqual(['drop:1583:10363', 'drop:1583:9568', 'drop:1583:other:Trash']);
		expect(spire.children.map(node => node.label)).toEqual(['General Drakkisath', 'Overlord Wyrmthalak', 'Trash']);
		expect(drops.itemIds).toEqual([3, 1, 2, 4]);
		expect(drops.children[2].leafKeys).toEqual(['drop:0:other:World Drop']);
	});

	it('orders the categories and puts the catch-all quest and vendor leaves last', () => {
		const tree = dropsSourceTree(
			[
				item(1, soldBy({ zoneId: 0 }), soldBy({ zoneId: 491 })),
				item(2, rep(529), crafted(Profession.Alchemy)),
				item(3, quest({ zoneId: 99999 }), quest({ zoneId: 139 }), drop({ zoneId: 491, otherName: 'Trash' })),
			],
			NAMES,
		);
		expect(tree.map(node => node.key)).toEqual(['drop', 'quest', 'crafted', 'rep', 'soldBy']);
		expect(tree[1].children.map(node => node.key)).toEqual(['quest:139', 'quest:other']);
		expect(tree[4].children.map(node => node.key)).toEqual(['soldBy:491', 'soldBy:0']);
		expect(tree[4].leafKeys).toEqual(['soldBy:491', 'soldBy:0']);
	});

	it('leaves out a category no item reaches', () => {
		const tree = dropsSourceTree([item(1, crafted(Profession.Alchemy), pvp())], NAMES);
		expect(tree.map(node => node.key)).toEqual(['crafted']);
		expect(dropsSourceTree([item(2, pvp())], NAMES)).toEqual([]);
	});

	it('reads a node as checked, unchecked or mixed from its leaves', () => {
		const zone = dropsSourceTree([item(1, drop({ zoneId: 1583, npcId: 10363 }), drop({ zoneId: 1583, npcId: 9568 }))], NAMES)[0].children[0];
		expect(dropsNodeState(zone, new Set())).toBe('unchecked');
		expect(dropsNodeState(zone, new Set(['drop:1583:10363']))).toBe('mixed');
		expect(dropsNodeState(zone, new Set(['drop:1583:10363', 'drop:1583:9568', 'quest:139']))).toBe('checked');
	});

	it('ticks and unticks a whole zone, sorted, keeping ticks of other nodes', () => {
		const zone = dropsSourceTree([item(1, drop({ zoneId: 1583, npcId: 10363 }), drop({ zoneId: 1583, npcId: 9568 }))], NAMES)[0].children[0];
		const ticked = dropsTicksAfter(['quest:139', 'crafted:1'], zone, true);
		expect(ticked).toEqual(['crafted:1', 'drop:1583:10363', 'drop:1583:9568', 'quest:139']);
		expect(dropsTicksAfter([...ticked, 'drop:1583:10363'], zone, false)).toEqual(['crafted:1', 'quest:139']);
	});

	it('keeps a tick whose node the tree does not list', () => {
		expect(dropsTicksAfter(['drop:2717:11502'], treeNode('quest'), true)).toContain('drop:2717:11502');
	});

	it('adds the five L60 dungeons to the ticks and ignores a zone the tree lacks', () => {
		expect(DROPS_L60_DUNGEON_ZONE_IDS).toEqual([1584, 1583, 2557, 2057, 2017]);
		const tree = dropsSourceTree(
			[
				item(1, drop({ zoneId: 1584, npcId: 9016 })),
				item(2, drop({ zoneId: 2557, otherName: 'Trash' })),
				item(3, drop({ zoneId: 491, otherName: 'Trash' })),
			],
			NAMES,
		);
		expect(dropsQuickPickTicks(['quest:139', 'drop:1584:9016'], tree, DROPS_L60_DUNGEON_ZONE_IDS)).toEqual([
			'drop:1584:9016',
			'drop:2557:other:Trash',
			'quest:139',
		]);
	});

	it('labels a boss by its NPC under its zone, and a quest zone with no row as Other quests', () => {
		expect(dropsSourceLabel('drop:1583:10363', NAMES)).toEqual({ label: 'General Drakkisath', detail: 'Blackrock Spire' });
		expect(dropsSourceLabel('drop:491:other:Trash', NAMES)).toEqual({ label: 'Trash', detail: 'Razorfen Kraul' });
		expect(dropsSourceLabel('drop:0:other:World Drop', NAMES)).toEqual({
			label: 'bulk_tab.drops.sources.world_drop',
			detail: 'bulk_tab.drops.sources.categories.drop',
		});
		expect(dropsSourceLabel('quest:139', NAMES)).toEqual({ label: 'Eastern Plaguelands', detail: 'bulk_tab.drops.sources.categories.quest' });
		expect(dropsSourceLabel('quest:other', NAMES)).toEqual({
			label: 'bulk_tab.drops.sources.other_quests',
			detail: 'bulk_tab.drops.sources.categories.quest',
		});
		expect(dropsSourceLabel(`crafted:${Profession.Alchemy}`, NAMES)).toEqual({
			label: professionNames.get(Profession.Alchemy),
			detail: 'bulk_tab.drops.sources.categories.crafted',
		});
		expect(dropsSourceLabel('rep:529', NAMES)).toEqual({ label: 'Argent Dawn', detail: 'bulk_tab.drops.sources.categories.rep' });
		expect(dropsSourceLabel('soldBy:0', NAMES)).toEqual({
			label: 'bulk_tab.drops.sources.other_vendors',
			detail: 'bulk_tab.drops.sources.categories.soldBy',
		});
	});
});
