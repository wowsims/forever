import { enchantPresetPool, saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ArmorType, Faction, ItemSlot, ItemType } from '@generated/proto/common';
import { BulkEnchantPresetSelection, ContentPhase, UIEnchant, type UIItem, UIItemSource } from '@generated/proto/ui';
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
	dropsItemListFor,
	dropsRequestSettings,
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

const drop = (npcId: number) =>
	UIItemSource.create({ source: { oneofKind: 'drop', drop: { difficulty: 0, npcId, zoneId: 1583, otherName: '', category: '' } } });

const item = (id: number, type: ItemType, npcId = 10363): UIItem =>
	makeUIItem({ id, name: `Item ${id}`, type, armorType: ArmorType.ArmorTypePlate, phase: ContentPhase.ContentPhase1, sources: [drop(npcId)] });

const CLOAK = item(101, ItemType.ItemTypeBack);
const CHEST = item(205, ItemType.ItemTypeChest);
const WORN_HELM = item(150, ItemType.ItemTypeHead);
const BOOTS = item(310, ItemType.ItemTypeFeet, 9568);
const ITEMS = [CHEST, WORN_HELM, CLOAK, BOOTS];

let player: Player<any>;

beforeEach(() => {
	player = makeBulkPlayer({
		getGear: () => new Gear({ [ItemSlot.ItemSlotHead]: makeEquippedItem({ item: WORN_HELM }) }),
		getFaction: () => Faction.Alliance,
		getSpec: () => player.getPlayerSpec().specID,
		getClass: () => player.getPlayerSpec().classID,
		getEnchants: (slot: ItemSlot) =>
			slot === ItemSlot.ItemSlotBack ? [UIEnchant.create({ effectId: CLOAK_AGILITY }), UIEnchant.create({ effectId: CLOAK_ARMOR })] : [],
		sim: {
			getPhase: () => ContentPhase.ContentPhase1,
			db: { getAllItems: () => ITEMS, getZone: () => undefined, getNpc: () => undefined, getRepFaction: () => undefined },
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

		expect(dropsItemListFor(player).items.map(entry => [entry.item.id, entry.worn, entry.included])).toEqual([
			[101, false, true],
			[150, true, true],
			[205, false, false],
			[310, false, true],
		]);
		expect(sentIds()).toEqual([101, 310]);
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
