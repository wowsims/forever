import { enchantPresetPool, saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ArmorType, Class, Faction, ItemSlot, ItemType } from '@generated/proto/common';
import { BulkEnchantPresetSelection, ContentPhase, DatabaseFilters, DungeonDifficulty, UIEnchant, type UIItem, UIItemSource } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Gear } from '@sim/proto/gear';
import {
	bulkState,
	loadStoredBulkDropsEnchantPresets,
	loadStoredBulkDropsExcluded,
	loadStoredBulkDropsSources,
	loadStoredBulkDropsView,
} from '@sim/settings/bulk_settings';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIItem } from '@testing/storybook/fixtures/protos';
import { beforeEach, describe, expect, it } from 'vitest';

import {
	dropsRequestSettings,
	dropsSetupFor,
	setDropsEnchantPresetChecked,
	setDropsExcluded,
	setDropsItemsIncluded,
	setDropsSources,
	setDropsView,
} from './settings';

const DRAKKISATH = 'drop:1583:10363';
const WYRMTHALAK = 'drop:1583:9568';
const CLOAK_AGILITY = 849;
const CLOAK_ARMOR = 884;

const drop = (npcId: number, zoneId = 1583) =>
	UIItemSource.create({
		source: { oneofKind: 'drop', drop: { difficulty: DungeonDifficulty.DifficultyDungeon, npcId, zoneId, otherName: '', category: '' } },
	});

const item = (id: number, type: ItemType, npcId = 10363, fields: Partial<UIItem> = {}): UIItem =>
	makeUIItem({
		id,
		name: `Item ${id}`,
		type,
		armorType: ArmorType.ArmorTypePlate,
		phase: ContentPhase.ContentPhase1,
		sources: [drop(npcId)],
		ilvl: 60,
		...fields,
	});

const CLOAK = item(101, ItemType.ItemTypeBack);
const CHEST = item(205, ItemType.ItemTypeChest);
const WORN_HELM = item(150, ItemType.ItemTypeHead);
const BOOTS = item(310, ItemType.ItemTypeFeet, 9568, { ilvl: 25 });
const ITEMS = [CHEST, WORN_HELM, CLOAK, BOOTS];

let player: Player<any>;
let filters: DatabaseFilters;
let items: UIItem[];

beforeEach(() => {
	filters = DatabaseFilters.create();
	items = ITEMS;
	player = makeBulkPlayer({
		getGear: () => new Gear({ [ItemSlot.ItemSlotHead]: makeEquippedItem({ item: WORN_HELM }) }),
		getFaction: () => Faction.Alliance,
		getSpec: () => player.getPlayerSpec().specID,
		getClass: () => player.getPlayerSpec().classID,
		getEnchants: (slot: ItemSlot) =>
			slot === ItemSlot.ItemSlotBack ? [UIEnchant.create({ effectId: CLOAK_AGILITY }), UIEnchant.create({ effectId: CLOAK_ARMOR })] : [],
		sim: {
			getPhase: () => ContentPhase.ContentPhase1,
			getFilters: () => filters,
			db: { getAllItems: () => items, getZone: () => undefined, getNpc: () => undefined, getRepFaction: () => undefined },
		},
	});
	saveEnchantPreset(player, { name: 'Agility', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY] }] });
	saveEnchantPreset(player, { name: 'Armor', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_ARMOR] }] });
});

const PICKABLE = ['Agility', 'Armor'];
const sentIds = () => dropsRequestSettings(player).items.map(spec => spec.id);
const settingsVersion = () => bulkState(player).v.settings;

describe('dropsRequestSettings', () => {
	it('sends the picked items sorted, one change at a time, with Drops’ own enchant pool', () => {
		setDropsSources(player, [DRAKKISATH]);
		setDropsEnchantPresetChecked(player, 'Agility', true, PICKABLE);

		const settings = dropsRequestSettings(player);

		expect(settings.items.map(spec => spec.id)).toEqual([101, 205]);
		expect(settings.singleChange).toBe(true);
		expect(settings.enchants).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY] })]);
		expect(settings.freezeRingSlot).toBe(0);
	});

	it('leaves Gear’s checked presets out of a Drops request and Drops’ out of Gear’s', () => {
		setDropsSources(player, [DRAKKISATH]);
		setEnchantPresetChecked(player, 'Armor', true);
		setDropsEnchantPresetChecked(player, 'Agility', true, PICKABLE);

		expect(dropsRequestSettings(player).enchants).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY] })]);
		expect(enchantPresetPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_ARMOR] })]);
	});

	it('sends nothing worn and nothing the user unticked', () => {
		setDropsSources(player, [DRAKKISATH, WYRMTHALAK]);
		setDropsItemsIncluded(player, [205], false);

		expect(dropsSetupFor(player).list.items.map(entry => [entry.item.id, entry.worn, entry.included])).toEqual([
			[101, false, true],
			[150, true, true],
			[205, false, false],
			[310, false, true],
		]);
		expect(sentIds()).toEqual([101, 310]);
	});
});

describe('dropsSetupFor level filters', () => {
	const leafKeys = () => dropsSetupFor(player).tree.flatMap(node => node.leafKeys);

	it('lists every source while the gear picker sets no level bound', () => {
		expect(leafKeys()).toEqual([DRAKKISATH, WYRMTHALAK]);
	});

	it('drops an item below Min ilvl from the list and its source from the tree when it was the source’s only item', () => {
		setDropsSources(player, [DRAKKISATH, WYRMTHALAK]);
		filters.minIlvl = 30;

		const { tree, list } = dropsSetupFor(player);

		expect(tree.flatMap(node => node.leafKeys)).toEqual([DRAKKISATH]);
		expect(list.items.map(entry => entry.item.id)).not.toContain(310);
	});

	it('follows Max ilvl and Required level as they change', () => {
		filters.maxIlvl = 30;
		expect(leafKeys()).toEqual([WYRMTHALAK]);
		filters.maxIlvl = 0;
		filters.maxRequiredLevel = 10;
		expect(leafKeys()).toEqual([]);
	});
});

describe('dropsSetupFor level 45 cutoff', () => {
	it('exposes the level of every dungeon the database holds, the cut ones too', () => {
		items = [...ITEMS, item(400, ItemType.ItemTypeLegs, 2748, { sources: [drop(2748, 1337)], reqLvl: 44 })];

		expect(dropsSetupFor(player).dungeonLevels.get(1337)).toBe(44);
	});

	const ULDAMAN = 1337;
	const ARCHAEDAS = 'drop:1337:2748';
	const MARAUDON = 2100;
	const PRINCESS = 'drop:2100:12201';
	const leafKeys = () => dropsSetupFor(player).tree.flatMap(node => node.leafKeys);

	it('leaves a dungeon whose loot needs at most level 44 out of the tree, and a stored tick on it sends nothing', () => {
		items = [
			...ITEMS,
			item(400, ItemType.ItemTypeLegs, 2748, { sources: [drop(2748, ULDAMAN)], reqLvl: 44 }),
			item(401, ItemType.ItemTypeWrist, 2748, { sources: [drop(2748, ULDAMAN), drop(9568)], reqLvl: 44 }),
		];
		setDropsSources(player, [DRAKKISATH, ARCHAEDAS]);

		expect(leafKeys()).toEqual([DRAKKISATH, WYRMTHALAK]);
		expect(sentIds()).toEqual([101, 205]);
		// The item Wyrmthalak drops too stays under him.
		setDropsSources(player, [WYRMTHALAK]);
		expect(sentIds()).toEqual([310, 401]);
	});

	it('reads a dungeon’s level from every item it drops, those the class or the level filters leave out included', () => {
		items = [
			...ITEMS,
			item(400, ItemType.ItemTypeLegs, 2748, { sources: [drop(2748, ULDAMAN)], reqLvl: 40 }),
			item(401, ItemType.ItemTypeChest, 2748, { sources: [drop(2748, ULDAMAN)], reqLvl: 50, classAllowlist: [Class.ClassMage] }),
			item(402, ItemType.ItemTypeLegs, 12201, { sources: [drop(12201, MARAUDON)], reqLvl: 40 }),
			item(403, ItemType.ItemTypeChest, 12201, { sources: [drop(12201, MARAUDON)], reqLvl: 50 }),
		];
		setDropsSources(player, [ARCHAEDAS, PRINCESS]);
		filters.maxRequiredLevel = 45;

		expect(leafKeys()).toEqual([ARCHAEDAS, WYRMTHALAK, PRINCESS]);
		expect(sentIds()).toEqual([400, 402]);
	});
});

describe('Drops setters', () => {
	it('stores the sources, excluded ids and preset pick, each bumping the count', () => {
		const steps: Array<[() => void, () => unknown, unknown]> = [
			[() => setDropsSources(player, [WYRMTHALAK, DRAKKISATH]), () => loadStoredBulkDropsSources(player), [WYRMTHALAK, DRAKKISATH].sort()],
			[() => setDropsExcluded(player, [310, 101, 310]), () => loadStoredBulkDropsExcluded(player), [101, 310]],
			[() => setDropsItemsIncluded(player, [205, 101], false), () => loadStoredBulkDropsExcluded(player), [101, 205, 310]],
			[() => setDropsItemsIncluded(player, [101], true), () => loadStoredBulkDropsExcluded(player), [205, 310]],
			[
				() => setDropsEnchantPresetChecked(player, 'Armor', true, PICKABLE),
				() => loadStoredBulkDropsEnchantPresets(player),
				BulkEnchantPresetSelection.create({ names: ['Armor'] }),
			],
			[
				() => setDropsEnchantPresetChecked(player, 'Agility', true, PICKABLE),
				() => loadStoredBulkDropsEnchantPresets(player),
				BulkEnchantPresetSelection.create({ names: ['Agility', 'Armor'] }),
			],
		];
		for (const [step, stored, expected] of steps) {
			const before = settingsVersion();
			step();
			expect(stored()).toEqual(expected);
			expect(settingsVersion()).toBe(before + 1);
		}
		const { dropsSources, dropsExcluded, dropsEnchantPresets } = bulkState(player);
		expect([dropsSources, dropsExcluded, dropsEnchantPresets]).toEqual([[DRAKKISATH, WYRMTHALAK].sort(), [205, 310], ['Agility', 'Armor']]);
	});

	it('stores the view without recounting', () => {
		const before = settingsVersion();
		setDropsView(player, 'items');
		expect([bulkState(player).dropsView, loadStoredBulkDropsView(player), settingsVersion()]).toEqual(['items', 'items', before]);
	});
});
