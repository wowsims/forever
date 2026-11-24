import { BulkSimStage, DistributionMetrics } from '@generated/proto/api';
import { Faction, HandType, ItemSlot, ItemType } from '@generated/proto/common';
import { type UIItem, UIItem_FactionRestriction, UIItemSource } from '@generated/proto/ui';
import type { TopGearResult } from '@sim/bulk/types';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { CURRENT_ROTATION_SETUP, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { BARTHILAS, BLACKROCK_SPIRE, DRAKKISATH, dropsNamesFixture, REND, RIVENDARE, STRATHOLME, WYRMTHALAK } from '@testing/storybook/fixtures/bulk-drops';
import { makeUIItem } from '@testing/storybook/fixtures/protos';
import { describe, expect, it } from 'vitest';

import { rankEncounterResults } from '../results/results';
import { wornIdSet } from './items';
import { dropsGap, dropsItemChains, dropsItemLookup, dropsNewItem, dropsResultItems, dropsSourceGroups, dropsSummary } from './results';

const REND_KEY = `drop:${BLACKROCK_SPIRE}:${REND}`;
const RIVENDARE_KEY = `drop:${STRATHOLME}:${RIVENDARE}`;
const DRAKKISATH_KEY = `drop:${BLACKROCK_SPIRE}:${DRAKKISATH}`;
const WYRMTHALAK_KEY = `drop:${BLACKROCK_SPIRE}:${WYRMTHALAK}`;
const BARTHILAS_KEY = `drop:${STRATHOLME}:${BARTHILAS}`;
// Barthilas is ticked now but was not when the run started.
const RUN_SOURCES: ReadonlySet<string> = new Set([REND_KEY, RIVENDARE_KEY, DRAKKISATH_KEY, WYRMTHALAK_KEY]);

const drop = (zoneId: number, npcId: number) =>
	UIItemSource.create({ source: { oneofKind: 'drop', drop: { difficulty: 0, npcId, zoneId, otherName: '', category: '' } } });

const item = (id: number, type: ItemType, sources: UIItemSource[] = [], handType = HandType.HandTypeUnknown) =>
	makeUIItem({ id, name: `Item ${id}`, type, handType, sources });

const WORN_MAIN_HAND = item(1, ItemType.ItemTypeWeapon, [], HandType.HandTypeOneHand);
const WORN_OFF_HAND = item(2, ItemType.ItemTypeWeapon, [], HandType.HandTypeOneHand);
const WORN_GLOVES = item(3, ItemType.ItemTypeHands, [drop(BLACKROCK_SPIRE, REND)]);
const WORN_RING_1 = item(4, ItemType.ItemTypeFinger);
const WORN_RING_2 = item(5, ItemType.ItemTypeFinger);
const FELSTRIKER = item(101, ItemType.ItemTypeWeapon, [drop(BLACKROCK_SPIRE, REND), drop(STRATHOLME, RIVENDARE)], HandType.HandTypeOneHand);
const GLOVES = item(102, ItemType.ItemTypeHands, [drop(BLACKROCK_SPIRE, REND)]);
const RING = item(103, ItemType.ItemTypeFinger, [drop(BLACKROCK_SPIRE, REND)]);
const TWO_HANDER = item(105, ItemType.ItemTypeWeapon, [drop(STRATHOLME, RIVENDARE)], HandType.HandTypeTwoHand);
const CHEST = item(106, ItemType.ItemTypeChest, [drop(BLACKROCK_SPIRE, DRAKKISATH)]);
const HELM = item(107, ItemType.ItemTypeHead, [drop(STRATHOLME, BARTHILAS)]);
const BELT = item(108, ItemType.ItemTypeWaist, [drop(BLACKROCK_SPIRE, WYRMTHALAK)]);
const CLOAK = item(109, ItemType.ItemTypeBack, [drop(BLACKROCK_SPIRE, WYRMTHALAK)]);

const ITEMS = new Map<number, UIItem>(
	[WORN_MAIN_HAND, WORN_OFF_HAND, WORN_GLOVES, WORN_RING_1, WORN_RING_2, FELSTRIKER, GLOVES, RING, TWO_HANDER, CHEST, HELM, BELT, CLOAK].map(entry => [
		entry.id,
		entry,
	]),
);
const itemById = (id: number) => ITEMS.get(id);
const names = dropsNamesFixture();

const equipped = (entry: UIItem) => new EquippedItem({ item: entry });

const CURRENT_GEAR = new Gear({
	[ItemSlot.ItemSlotMainHand]: equipped(WORN_MAIN_HAND),
	[ItemSlot.ItemSlotOffHand]: equipped(WORN_OFF_HAND),
	[ItemSlot.ItemSlotHands]: equipped(WORN_GLOVES),
	[ItemSlot.ItemSlotFinger1]: equipped(WORN_RING_1),
	[ItemSlot.ItemSlotFinger2]: equipped(WORN_RING_2),
});

const wearing = (changes: Partial<Record<ItemSlot, UIItem | null>>) =>
	new Gear({
		...CURRENT_GEAR.asMap(),
		...Object.fromEntries(Object.entries(changes).map(([slot, entry]) => [slot, entry ? equipped(entry) : null])),
	});

// A dual wielder's single-change rows: Felstriker in the main hand moves the worn main hand to the off hand.
const FELSTRIKER_GEAR = wearing({ [ItemSlot.ItemSlotMainHand]: FELSTRIKER, [ItemSlot.ItemSlotOffHand]: WORN_MAIN_HAND });
const TWO_HANDER_GEAR = wearing({ [ItemSlot.ItemSlotMainHand]: TWO_HANDER, [ItemSlot.ItemSlotOffHand]: null });

const CURRENT_DPS = 750;
const N = 10_000;

const metrics = (avg: number, n = N, stdev = 100) => DistributionMetrics.create({ avg, stdev, aggregatorData: { n, sumSq: 0 } });

// A row paired with the baseline at an error of `error`, its paired difference the gap to the current gear's DPS.
const paired = (gear: Gear, avg: number, extra: Partial<TopGearResult> = {}, error = 1): TopGearResult => ({
	gear,
	dpsMetrics: metrics(avg),
	pairedErrorToBaseline: error,
	pairedDpsDiffToBaseline: avg - CURRENT_DPS,
	...extra,
});

const rank = (rows: TopGearResult[]) =>
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

// Rend drops the ring, Felstriker and the gloves, which are culled but above the current gear, and the worn gloves.
// Felstriker drops from Rivendare too, as does a two-hander that loses. Drakkisath's chest sits inside its paired error,
// Wyrmthalak's belt is a small upgrade, and Barthilas' helm upgrades but its source was not ticked when the run started.
const RESULTS = rank([
	paired(wearing({ [ItemSlot.ItemSlotFinger1]: RING }), 767.5, { backendRank: 0 }),
	paired(FELSTRIKER_GEAR, 765, { backendRank: 1 }),
	paired(wearing({ [ItemSlot.ItemSlotFinger2]: RING }), 762.1, { backendRank: 2 }),
	paired(wearing({ [ItemSlot.ItemSlotChest]: CHEST }), 758, { backendRank: 3 }, 10),
	paired(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), 756, { backendRank: 4 }),
	paired(wearing({ [ItemSlot.ItemSlotHead]: HELM }), 755, { backendRank: 5 }),
	paired(wearing({ [ItemSlot.ItemSlotHands]: GLOVES }), 760, { culledStage: BulkSimStage.BulkSimStageMedium }),
	paired(TWO_HANDER_GEAR, 740, { culledStage: BulkSimStage.BulkSimStageLow }),
]);
const CURRENT = RESULTS.originalGearResults;
const ITEM_RESULTS = dropsResultItems(RESULTS);
const GROUPS = dropsSourceGroups(ITEM_RESULTS, RUN_SOURCES, CURRENT_GEAR, itemById, names);

const avgs = (chains: TopGearResult[][]) => chains.flat().map(row => row.dpsMetrics.avg);

const chainOptions = { groups: GROUPS };

describe('dropsNewItem', () => {
	it('finds the new item of a row that moved the worn main hand to the off hand', () => {
		expect(dropsNewItem(FELSTRIKER_GEAR, wornIdSet(CURRENT_GEAR))?.id).toBe(FELSTRIKER.id);
	});

	it('finds a two-hander that replaced a dual wielder’s pair', () => {
		expect(dropsNewItem(TWO_HANDER_GEAR, wornIdSet(CURRENT_GEAR))?.id).toBe(TWO_HANDER.id);
		expect(dropsNewItem(CURRENT_GEAR, wornIdSet(CURRENT_GEAR))).toBeNull();
	});
});

describe('dropsResultItems', () => {
	it('gathers an item’s rows into its best and its proven upgrade rows, best first', () => {
		const items = ITEM_RESULTS;
		expect(items[0].best.dpsMetrics.avg).toBe(767.5);
		expect(items[0].upgradeRows.map(row => row.dpsMetrics.avg)).toEqual([767.5, 762.1]);
		expect(items.find(entry => entry.itemId === GLOVES.id)!.upgradeRows.map(row => row.dpsMetrics.avg)).toEqual([760]);
		expect(items[0].gain).toBeCloseTo(17.5);
		expect(items.map(entry => entry.itemId)).toEqual([RING.id, FELSTRIKER.id, GLOVES.id, CHEST.id, BELT.id, HELM.id, TWO_HANDER.id]);
		expect(items.at(-1)!.gain).toBeCloseTo(-10);
		expect(items.filter(entry => entry.upgrade).every(entry => entry.gain > 0)).toBe(true);
	});

	it('measures an item’s gain by its paired difference, so an upgrade never shows a loss', () => {
		const [belt] = dropsResultItems(rank([paired(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), 748, { pairedDpsDiffToBaseline: 5 })]));
		expect([belt.upgrade, belt.gain]).toEqual([true, 5]);
	});

	it('counts an item inside the paired error of the current gear as no upgrade', () => {
		const beltAt = (avg: number) => dropsResultItems(rank([paired(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), avg, {}, 1.5)]))[0].upgrade;
		expect([beltAt(752), beltAt(756)]).toEqual([false, true]);
	});

	it('takes an item’s upgrade from the sign of its paired difference, not from the gap between the averages', () => {
		const [belt] = dropsResultItems(rank([paired(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), 752, { pairedDpsDiffToBaseline: -5 })]));
		expect([belt.upgrade, belt.gain]).toEqual([false, -5]);
	});

	it('judges a culled item with no paired error at its own iteration count', () => {
		// stdev 100: a 3 DPS gap separates at 50,000 iterations against the current gear's 10,000, not at 1,000.
		const beltAt = (n: number) =>
			dropsResultItems(
				rank([
					{
						gear: wearing({ [ItemSlot.ItemSlotWaist]: BELT }),
						dpsMetrics: metrics(753, n),
						pairedErrorToBaseline: 0,
						culledStage: BulkSimStage.BulkSimStageLow,
					},
				]),
			)[0].upgrade;
		expect([beltAt(1_000), beltAt(50_000)]).toEqual([false, true]);
	});

	it('leads an item with its proven upgrade, a tied row above it no upgrade row', () => {
		const [ring] = dropsResultItems(
			rank([
				paired(wearing({ [ItemSlot.ItemSlotFinger1]: RING }), 770, { culledStage: BulkSimStage.BulkSimStageLow }, 15),
				paired(wearing({ [ItemSlot.ItemSlotFinger2]: RING }), 760),
			]),
		);
		expect([ring.upgrade, ring.best.dpsMetrics.avg, ring.gain]).toEqual([true, 760, 10]);
		expect(ring.upgradeRows.map(row => row.dpsMetrics.avg)).toEqual([760]);
	});
});

describe('dropsGap', () => {
	it('takes the paired difference, not the gap between the averages', () => {
		expect(dropsGap(paired(CURRENT_GEAR, 748, { pairedDpsDiffToBaseline: 5 }), CURRENT)).toBe(5);
	});

	it('measures a row with a paired error but no paired difference by the averages', () => {
		// A decoded response holds 0 for a difference the backend never set.
		expect(dropsGap(paired(CURRENT_GEAR, 760, { pairedDpsDiffToBaseline: 0 }), CURRENT)).toBe(10);
	});
});

describe('dropsSourceGroups', () => {
	const group = (key: string) => GROUPS.find(entry => entry.key === key)!;

	it('puts an item with two ticked sources on both cards', () => {
		expect(group(REND_KEY).items.map(entry => entry.itemId)).toContain(FELSTRIKER.id);
		expect(group(RIVENDARE_KEY).items.map(entry => entry.itemId)).toEqual([FELSTRIKER.id, TWO_HANDER.id]);
		expect(group(RIVENDARE_KEY).label).toEqual({ label: 'Baron Rivendare', detail: 'Stratholme' });
	});

	it('groups by the sources the run started with', () => {
		expect(GROUPS.map(entry => entry.key)).not.toContain(BARTHILAS_KEY);
		const nowTicked = new Set([...RUN_SOURCES, BARTHILAS_KEY]);
		expect(dropsSourceGroups(ITEM_RESULTS, nowTicked, CURRENT_GEAR, itemById, names).map(entry => entry.key)).toContain(BARTHILAS_KEY);
	});

	it('sorts the cards by best gain, the ones with no upgrade last', () => {
		// Drakkisath's chest gains more than Wyrmthalak's belt but sits inside its paired error.
		expect(GROUPS.map(entry => entry.key)).toEqual([REND_KEY, RIVENDARE_KEY, WYRMTHALAK_KEY, DRAKKISATH_KEY]);
		expect([group(DRAKKISATH_KEY).upgrades, group(DRAKKISATH_KEY).items.length, group(DRAKKISATH_KEY).wornIds]).toEqual([0, 1, []]);
	});

	it('ranks a card by its best upgrade, not by a tied row above it', () => {
		const results = rank([
			paired(wearing({ [ItemSlot.ItemSlotBack]: CLOAK }), 770, {}, 15),
			paired(wearing({ [ItemSlot.ItemSlotFinger1]: RING }), 760),
			paired(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), 756),
		]);
		const items = dropsResultItems(results);
		const groups = dropsSourceGroups(items, RUN_SOURCES, CURRENT_GEAR, itemById, names);
		expect(groups.map(entry => entry.key)).toEqual([REND_KEY, WYRMTHALAK_KEY]);
		const summary = dropsSummary(results, items, groups);
		expect([summary.best?.itemId, summary.bestGroup?.key]).toEqual([RING.id, REND_KEY]);
	});

	it('counts a worn item of a ticked source in N / M without an upgrade', () => {
		const rend = group(REND_KEY);
		expect(rend.items.map(entry => entry.itemId)).toEqual([RING.id, FELSTRIKER.id, GLOVES.id]);
		expect(rend.wornIds).toEqual([WORN_GLOVES.id]);
		expect(rend.upgrades).toBe(3);
	});
});

describe('dropsItemChains', () => {
	it('lists every proven upgrade row of an item, the best row of one with none, and the current gear at its rank', () => {
		const chains = dropsItemChains(RESULTS, ITEM_RESULTS, { source: null, ...chainOptions });
		// The ring upgrades in both slots and lists both; the two-hander has one row and loses, so it lists that one.
		expect(avgs(chains)).toEqual([767.5, 765, 762.1, 760, 758, 756, 755, CURRENT_DPS, 740]);
		expect(chains.flat()).toContain(CURRENT);
	});

	it('lists only the best row of an item none of whose rows upgrade', () => {
		const results = rank([
			paired(wearing({ [ItemSlot.ItemSlotFinger1]: RING }), 749, { pairedDpsDiffToBaseline: -1 }),
			paired(wearing({ [ItemSlot.ItemSlotFinger2]: RING }), 745, { pairedDpsDiffToBaseline: -5 }),
		]);
		const items = dropsResultItems(results);
		expect(avgs(dropsItemChains(results, items, { source: null, groups: [] }))).toEqual([CURRENT_DPS, 749]);
	});

	it('narrows the chains to one source’s items, the current gear kept', () => {
		expect(avgs(dropsItemChains(RESULTS, ITEM_RESULTS, { source: RIVENDARE_KEY, ...chainOptions }))).toEqual([765, CURRENT_DPS, 740]);
		expect(avgs(dropsItemChains(RESULTS, ITEM_RESULTS, { source: REND_KEY, ...chainOptions }))).toEqual([767.5, 765, 762.1, 760, CURRENT_DPS]);
		// The chest's paired error ties it with the current gear once nothing ranks between them.
		const drakkisath = dropsItemChains(RESULTS, ITEM_RESULTS, { source: DRAKKISATH_KEY, ...chainOptions });
		expect(drakkisath.map(chain => chain.map(row => row.dpsMetrics.avg))).toEqual([[758, CURRENT_DPS]]);
	});
});

describe('dropsSummary', () => {
	it('names the best drop and the best boss', () => {
		const summary = dropsSummary(RESULTS, ITEM_RESULTS, GROUPS);
		expect(summary.current).toBe(CURRENT);
		expect(summary.best?.itemId).toBe(RING.id);
		expect(summary.bestGroup?.key).toBe(REND_KEY);
		expect(summary.bestGroup?.items).toContain(summary.best);
	});

	it('names the upgrade with the largest gain as the best drop, on the best boss’s card', () => {
		// The ring ranks higher by its average, but its paired difference is the smaller one.
		const results = rank([
			paired(wearing({ [ItemSlot.ItemSlotFinger1]: RING }), 770, { pairedDpsDiffToBaseline: 4 }),
			paired(wearing({ [ItemSlot.ItemSlotWaist]: BELT }), 765),
		]);
		const items = dropsResultItems(results);
		const groups = dropsSourceGroups(items, RUN_SOURCES, CURRENT_GEAR, itemById, names);
		const summary = dropsSummary(results, items, groups);
		expect([summary.best?.itemId, summary.bestGroup?.key]).toEqual([BELT.id, WYRMTHALAK_KEY]);
		expect(summary.bestGroup?.items).toContain(summary.best);
		expect(groups.map(entry => entry.key)).toEqual([WYRMTHALAK_KEY, REND_KEY]);
	});

	it('names no best drop or boss when nothing upgrades', () => {
		const results = rank([paired(wearing({ [ItemSlot.ItemSlotChest]: CHEST }), 752, {}, 1.5), paired(TWO_HANDER_GEAR, 740)]);
		const items = dropsResultItems(results);
		const groups = dropsSourceGroups(items, RUN_SOURCES, CURRENT_GEAR, itemById, names);
		expect(groups.map(entry => entry.upgrades)).toEqual([0, 0]);
		expect(dropsSummary(results, items, groups)).toEqual({ current: results.originalGearResults, best: null, bestGroup: null });
	});
});

describe('dropsItemLookup', () => {
	it('resolves items through the faction filter but not the phase filter', () => {
		const quest = (faction: Faction) =>
			UIItemSource.create({ source: { oneofKind: 'quest', quest: { id: 1, name: 'q', faction, zoneId: BLACKROCK_SPIRE } } });
		const hordeOnly = makeUIItem({ id: 201, factionRestriction: UIItem_FactionRestriction.HORDE_ONLY, sources: [drop(BLACKROCK_SPIRE, REND)] });
		const mixed = makeUIItem({ id: 202, sources: [quest(Faction.Horde), drop(BLACKROCK_SPIRE, REND)] });
		const later = makeUIItem({ id: 203, phase: 6, sources: [drop(BLACKROCK_SPIRE, REND)] });
		const lookup = dropsItemLookup(id => [hordeOnly, mixed, later].find(entry => entry.id === id), Faction.Alliance);
		expect(lookup(hordeOnly.id)).toBeUndefined();
		expect(lookup(mixed.id)?.sources).toEqual([drop(BLACKROCK_SPIRE, REND)]);
		expect(lookup(later.id)).toBe(later);
	});
});
