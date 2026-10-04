import { GemColor, ItemQuality, ScalingItemProperties } from '@generated/proto/common';
import { UIEnchant as Enchant, UIGem as Gem, UIItem as Item } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { describe, expect, it } from 'vitest';

import { type GearData, SelectorModalTabs } from '../../types';
import { buildSelectorTabs, eligibilityFor } from './utils';

const gearData: GearData = { equipItem: () => undefined, getEquippedItem: () => null, subscribe: () => () => undefined };

// `itemsTabData` builds a real `EquippedItem` per row, and that reads the base scaling entry.
const scalingOptions = { 0: ScalingItemProperties.create({ ilvl: 120 }) };

interface ItemStub {
	randomSuffixOptions?: number;
	randomSuffix?: object | null;
	sockets?: GemColor[];
}

const equippedItem = ({ randomSuffixOptions = 0, randomSuffix = null, sockets = [] }: ItemStub) =>
	({
		item: Item.create({
			id: 1,
			name: 'Stub',
			quality: ItemQuality.ItemQualityEpic,
			ilvl: 120,
			scalingOptions,
			randomSuffixOptions: Array.from({ length: randomSuffixOptions }, (_, index) => index + 1),
			gemSockets: sockets,
			socketBonus: [],
		}),
		randomSuffix,
		_randomSuffix: randomSuffix,
		hasRandomSuffixOptions: () => randomSuffixOptions > 0,
		numSockets: () => sockets.length,
		curSocketColors: () => sockets,
		withRandomSuffix: () => ({}),
	}) as unknown as EquippedItem;

interface PlayerStub {
	items?: number;
	enchants?: number;
	gems?: number;
	randomSuffixes?: number;
}

const player = ({ items = 1, enchants = 1, gems = 1, randomSuffixes = 1 }: PlayerStub) =>
	({
		getItems: () => Array.from({ length: items }, (_, index) => Item.create({ id: index + 1, name: `Item ${index}`, ilvl: 120, scalingOptions })),
		getEnchants: () => Array.from({ length: enchants }, (_, index) => Enchant.create({ effectId: index + 1, name: `Enchant ${index}` })),
		getGems: () => Array.from({ length: gems }, (_, index) => Gem.create({ id: index + 1, name: `Gem ${index}` })),
		getRandomSuffixes: () => Array.from({ length: randomSuffixes }, (_, index) => ({ id: index + 1, name: `Suffix ${index}` })),
		computeItemEP: () => 1,
		computeEnchantEP: () => 1,
		computeGemEP: () => 1,
		computeRandomSuffixEP: () => 1,
		computeStatsEP: () => 1,
	}) as unknown as Player<any>;

const tabs = (options: { player: PlayerStub; item: ItemStub | null; addOnly?: boolean }) =>
	buildSelectorTabs({
		player: player(options.player),
		slot: 0,
		gearData: { ...gearData, addOnly: options.addOnly },
		equippedItem: options.item ? equippedItem(options.item) : null,
	});

const labels = (options: { player: PlayerStub; item: ItemStub | null }) => tabs(options).map(tab => tab.label);

describe('buildSelectorTabs', () => {
	it('offers only the item and enchant tabs for an empty slot', () => {
		expect(labels({ player: {}, item: null })).toEqual([SelectorModalTabs.Items, SelectorModalTabs.Enchants]);
	});

	it('drops a tab whose data is empty rather than showing it blank', () => {
		expect(labels({ player: { enchants: 0 }, item: null })).toEqual([SelectorModalTabs.Items]);
	});

	it('orders the tabs items, enchants, suffix, gems', () => {
		expect(
			labels({
				player: {},
				item: { randomSuffixOptions: 2, randomSuffix: { id: 1 }, sockets: [GemColor.GemColorRed] },
			}),
		).toEqual([SelectorModalTabs.Items, SelectorModalTabs.Enchants, SelectorModalTabs.RandomSuffixes, SelectorModalTabs.Gem1]);
	});

	it('offers the suffix tab only to an item that has suffix options', () => {
		expect(labels({ player: {}, item: {} })).not.toContain(SelectorModalTabs.RandomSuffixes);
		expect(labels({ player: {}, item: { randomSuffixOptions: 2 } })).toContain(SelectorModalTabs.RandomSuffixes);
	});

	it('builds one gem tab per socket the item currently has', () => {
		expect(labels({ player: {}, item: { sockets: [GemColor.GemColorRed, GemColor.GemColorBlue] } })).toEqual([
			SelectorModalTabs.Items,
			SelectorModalTabs.Enchants,
			SelectorModalTabs.Gem1,
			SelectorModalTabs.Gem2,
		]);
	});

	it('offers an add-only picker the items tab alone, with nothing to remove', () => {
		const built = tabs({ player: {}, item: null, addOnly: true });
		expect(built.map(tab => tab.label)).toEqual([SelectorModalTabs.Items]);
		expect(built[0].onRemove).toBeUndefined();
		expect(tabs({ player: {}, item: null })[0].onRemove).toBeDefined();
	});
});

describe('eligibilityFor', () => {
	it('reports what an empty slot offers', () => {
		expect(eligibilityFor({ player: player({}), slot: 0, equippedItem: null })).toEqual({
			hasEnchants: true,
			socketCount: undefined,
		});
	});

	it('counts the sockets the item has right now', () => {
		expect(
			eligibilityFor({
				player: player({}),
				slot: 0,
				equippedItem: equippedItem({ sockets: [GemColor.GemColorRed] }),
			}).socketCount,
		).toBe(1);
	});
});
