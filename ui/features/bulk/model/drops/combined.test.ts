import { DistributionMetrics } from '@generated/proto/api';
import { Faction, HandType, ItemSlot, ItemType } from '@generated/proto/common';
import { type UIItem, UIItemSource } from '@generated/proto/ui';
import type { TopGearResult } from '@sim/bulk/types';
import type { Database } from '@sim/proto/database';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { CURRENT_ROTATION_SETUP, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { BLACKROCK_SPIRE, dropsNamesFixture, REND, RIVENDARE, STRATHOLME, WYRMTHALAK } from '@testing/storybook/fixtures/bulk-drops';
import { makeUIItem } from '@testing/storybook/fixtures/protos';
import { describe, expect, it } from 'vitest';

import { rankEncounterResults } from '../results/results';
import { dropsCombinedSets } from './combined';

const REND_KEY = `drop:${BLACKROCK_SPIRE}:${REND}`;
const RIVENDARE_KEY = `drop:${STRATHOLME}:${RIVENDARE}`;
const WYRMTHALAK_KEY = `drop:${BLACKROCK_SPIRE}:${WYRMTHALAK}`;

const drop = (zoneId: number, npcId: number) =>
	UIItemSource.create({ source: { oneofKind: 'drop', drop: { difficulty: 0, npcId, zoneId, otherName: '', category: '' } } });

const item = (id: number, type: ItemType, sources: UIItemSource[], fields: Partial<UIItem> = {}) =>
	makeUIItem({ id, name: `Item ${id}`, type, handType: HandType.HandTypeUnknown, reqLvl: 60, sources, ...fields });

const rend = [drop(BLACKROCK_SPIRE, REND)];
const WORN_MAIN_HAND = item(1, ItemType.ItemTypeWeapon, [], { handType: HandType.HandTypeOneHand });
const WORN_OFF_HAND = item(2, ItemType.ItemTypeWeapon, [], { handType: HandType.HandTypeOneHand });
const WORN_RING_1 = item(3, ItemType.ItemTypeFinger, []);
const WORN_RING_2 = item(4, ItemType.ItemTypeFinger, []);
const RING = item(101, ItemType.ItemTypeFinger, rend);
const OTHER_RING = item(102, ItemType.ItemTypeFinger, rend);
const FELSTRIKER = item(103, ItemType.ItemTypeWeapon, [...rend, drop(STRATHOLME, RIVENDARE)], { handType: HandType.HandTypeOneHand });
const TWO_HANDER = item(104, ItemType.ItemTypeWeapon, [...rend, drop(STRATHOLME, RIVENDARE)], { handType: HandType.HandTypeTwoHand });
const GLOVES = item(105, ItemType.ItemTypeHands, rend);
const TRINKET = item(106, ItemType.ItemTypeTrinket, rend, { limitCategory: 7 });
const SAME_CATEGORY_TRINKET = item(107, ItemType.ItemTypeTrinket, rend, { limitCategory: 7 });
const LOSING_CHEST = item(108, ItemType.ItemTypeChest, rend);
const BELT = item(109, ItemType.ItemTypeWaist, [drop(BLACKROCK_SPIRE, WYRMTHALAK)]);
const ITEMS = [
	WORN_MAIN_HAND,
	WORN_OFF_HAND,
	WORN_RING_1,
	WORN_RING_2,
	RING,
	OTHER_RING,
	FELSTRIKER,
	TWO_HANDER,
	GLOVES,
	TRINKET,
	SAME_CATEGORY_TRINKET,
	LOSING_CHEST,
	BELT,
];

const names = dropsNamesFixture();
const DB = {
	getAllItems: () => ITEMS,
	getItemById: (id: number) => ITEMS.find(entry => entry.id === id),
	getZone: (id: number) => ({ name: names.zone(id) }),
	getNpc: (id: number) => ({ name: names.npc(id) }),
	getRepFaction: () => undefined,
} as unknown as Database;

const equipped = (entry: UIItem) => new EquippedItem({ item: entry });

const CURRENT_GEAR = new Gear({
	[ItemSlot.ItemSlotMainHand]: equipped(WORN_MAIN_HAND),
	[ItemSlot.ItemSlotOffHand]: equipped(WORN_OFF_HAND),
	[ItemSlot.ItemSlotFinger1]: equipped(WORN_RING_1),
	[ItemSlot.ItemSlotFinger2]: equipped(WORN_RING_2),
});

const wearing = (changes: Partial<Record<ItemSlot, UIItem | null>>) =>
	new Gear({
		...CURRENT_GEAR.asMap(),
		...Object.fromEntries(Object.entries(changes).map(([slot, entry]) => [slot, entry ? equipped(entry) : null])),
	});

const CURRENT_DPS = 750;
const N = 10_000;
const metrics = (avg: number) => DistributionMetrics.create({ avg, stdev: 100, aggregatorData: { n: N, sumSq: 0 } });
const row = (gear: Gear, gain: number): TopGearResult => ({
	gear,
	dpsMetrics: metrics(CURRENT_DPS + gain),
	pairedErrorToBaseline: 1,
	pairedDpsDiffToBaseline: gain,
});

const resultsOf = (rows: TopGearResult[]) =>
	rankEncounterResults(
		[
			{
				reference: true,
				talents: CURRENT_TALENT_SETUP,
				rotation: CURRENT_ROTATION_SETUP,
				settings: null,
				setups: null,
				result: { referenceDpsMetrics: metrics(CURRENT_DPS), topGearResults: rows, iterations: N, metrics: {} },
			},
		],
		CURRENT_GEAR,
		0,
	);

// Rend drops two rings (the better one upgrades in either slot), Felstriker, which moves the worn main hand across, a
// two-hander, gloves, two trinkets of one limit category and a chest that loses. Rivendare drops Felstriker and the
// two-hander, and Wyrmthalak only the belt.
const RESULTS = resultsOf([
	row(wearing({ [ItemSlot.ItemSlotFinger1]: RING }), 20),
	row(wearing({ [ItemSlot.ItemSlotFinger2]: RING }), 18),
	row(wearing({ [ItemSlot.ItemSlotMainHand]: FELSTRIKER, [ItemSlot.ItemSlotOffHand]: WORN_MAIN_HAND }), 15),
	row(wearing({ [ItemSlot.ItemSlotFinger1]: OTHER_RING }), 12),
	row(wearing({ [ItemSlot.ItemSlotFinger2]: OTHER_RING }), 11),
	row(wearing({ [ItemSlot.ItemSlotHands]: GLOVES }), 10),
	row(wearing({ [ItemSlot.ItemSlotMainHand]: TWO_HANDER, [ItemSlot.ItemSlotOffHand]: null }), 8),
	row(wearing({ [ItemSlot.ItemSlotTrinket1]: TRINKET }), 6),
	row(wearing({ [ItemSlot.ItemSlotTrinket2]: SAME_CATEGORY_TRINKET }), 5),
	row(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), 9),
	row(wearing({ [ItemSlot.ItemSlotChest]: LOSING_CHEST }), -4),
]);
const SETS = dropsCombinedSets(DB, RESULTS, [REND_KEY, RIVENDARE_KEY, WYRMTHALAK_KEY], Faction.Alliance);

const ids = (gear: Gear) =>
	Object.fromEntries(gear.getItemSlots().flatMap(slot => (gear.getEquippedItem(slot) ? [[slot, gear.getEquippedItem(slot)!.id]] : [])));

describe('dropsCombinedSets', () => {
	it('wears a card’s upgrades together, largest gain first, leaving out what a slot or limit category conflict keeps out', () => {
		const set = SETS.get(REND_KEY)!;
		expect(set.itemIds).toEqual([RING.id, FELSTRIKER.id, OTHER_RING.id, GLOVES.id, TRINKET.id]);
		expect(set.leftOutIds).toEqual([TWO_HANDER.id, SAME_CATEGORY_TRINKET.id]);
		// The second ring takes the slot the first left free, and Felstriker brings the worn main hand into the off hand.
		expect(ids(set.gear)).toEqual({
			[ItemSlot.ItemSlotMainHand]: FELSTRIKER.id,
			[ItemSlot.ItemSlotOffHand]: WORN_MAIN_HAND.id,
			[ItemSlot.ItemSlotFinger1]: RING.id,
			[ItemSlot.ItemSlotFinger2]: OTHER_RING.id,
			[ItemSlot.ItemSlotHands]: GLOVES.id,
			[ItemSlot.ItemSlotTrinket1]: TRINKET.id,
		});
	});

	it('sims no set for a card whose upgrades all want the same hands, nor for a card with one upgrade', () => {
		expect([...SETS.keys()]).toEqual([REND_KEY]);
	});
});
