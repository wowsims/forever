import type { BulkResultRowViewProps } from '@features/bulk/components/BulkResults/BulkResultRow';
import type { DropsBossCardViewProps } from '@features/bulk/components/DropsBossCard';
import type { DropsItemLabelProps } from '@features/bulk/components/DropsItemLabel';
import { type DropsItemList, dropsItemList, dropsUsableItems } from '@features/bulk/model/drops/items';
import { dropsItemSourceKeys, type DropsNames, type DropsSourceNode, dropsSourceTree } from '@features/bulk/model/drops/sources';
import { rankEncounterResults } from '@features/bulk/model/results/results';
import { BulkSimStage, DistributionMetrics } from '@generated/proto/api';
import { ArmorType, Class, Faction, HandType, ItemQuality, ItemSlot, ItemType, WeaponType } from '@generated/proto/common';
import { ContentPhase, DungeonDifficulty, type UIItem, UIItem_FactionRestriction, UIItemSource } from '@generated/proto/ui';
import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import { PlayerSpecs } from '@sim/player/specs';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { wowheadIconUrl } from '@sim/proto/wowhead';

import { CURRENT_ROTATION_SETUP, CURRENT_TALENT_SETUP } from './bulk-compare';
import { makeUIItem } from './protos';

export const BLACKROCK_SPIRE = 1583;
export const STRATHOLME = 2017;
export const MOLTEN_CORE = 2717;

export const REND = 10429;
export const DRAKKISATH = 10363;
export const WYRMTHALAK = 9568;
export const RIVENDARE = 10440;
export const BARTHILAS = 10435;
export const RAGNAROS = 11502;

const ZONES: Record<number, string> = {
	[BLACKROCK_SPIRE]: 'Blackrock Spire',
	1584: 'Blackrock Depths',
	2557: 'Dire Maul',
	2057: 'Scholomance',
	[STRATHOLME]: 'Stratholme',
	[MOLTEN_CORE]: 'Molten Core',
};

const NPCS: Record<number, string> = {
	[REND]: 'Warchief Rend Blackhand',
	[DRAKKISATH]: 'General Drakkisath',
	[WYRMTHALAK]: 'Overlord Wyrmthalak',
	[RIVENDARE]: 'Baron Rivendare',
	[BARTHILAS]: 'Magistrate Barthilas',
	[RAGNAROS]: 'Ragnaros',
};

export const dropsNamesFixture = (): DropsNames => ({
	zone: id => ZONES[id] ?? null,
	npc: id => NPCS[id] ?? null,
	faction: () => null,
});

const drop = (zoneId: number, npcId: number) =>
	UIItemSource.create({
		source: {
			oneofKind: 'drop',
			drop: {
				difficulty: zoneId === MOLTEN_CORE ? DungeonDifficulty.DifficultyRaid : DungeonDifficulty.DifficultyDungeon,
				npcId,
				zoneId,
				otherName: '',
				category: '',
			},
		},
	});
const worldDrop = UIItemSource.create({
	source: { oneofKind: 'drop', drop: { difficulty: DungeonDifficulty.DifficultyUnknown, npcId: 0, zoneId: 0, otherName: 'World Drop', category: '' } },
});

const gear = (id: number, name: string, type: ItemType, sources: UIItemSource[], fields: Partial<UIItem> = {}) =>
	makeUIItem({
		id,
		name,
		type,
		icon: 'inv_misc_questionmark',
		armorType: ArmorType.ArmorTypePlate,
		weaponType: WeaponType.WeaponTypeUnknown,
		handType: HandType.HandTypeUnknown,
		quality: ItemQuality.ItemQualityRare,
		phase: ContentPhase.ContentPhase1,
		sources,
		...fields,
	});

const weapon = (id: number, name: string, handType: HandType, sources: UIItemSource[], fields: Partial<UIItem> = {}) =>
	gear(id, name, ItemType.ItemTypeWeapon, sources, { armorType: ArmorType.ArmorTypeUnknown, weaponType: WeaponType.WeaponTypeSword, handType, ...fields });

export const WORN_ITEM_ID = 2003;
export const UNTICKED_ITEM_ID = 2007;
export const ALLIANCE_ITEM_ID = 2011;
export const WEARABLE_ITEM_ID = 2005;

/** What the database holds for the fixture: a two-source weapon, a worn weapon, a class and a random-suffix item the list leaves out, an Alliance-only item, a raid boss and a world drop. */
export const dropsItemsFixture = (): UIItem[] => [
	weapon(2001, 'Felstriker', HandType.HandTypeOneHand, [drop(BLACKROCK_SPIRE, REND), drop(STRATHOLME, RIVENDARE)], {
		quality: ItemQuality.ItemQualityEpic,
		ilvl: 63,
		reqLvl: 58,
	}),
	weapon(2002, 'Blackblade of Shahram', HandType.HandTypeTwoHand, [drop(BLACKROCK_SPIRE, DRAKKISATH)], {
		quality: ItemQuality.ItemQualityEpic,
		ilvl: 63,
		reqLvl: 58,
	}),
	weapon(WORN_ITEM_ID, 'Blackhand Doomsaw', HandType.HandTypeTwoHand, [drop(BLACKROCK_SPIRE, REND)], { ilvl: 63, reqLvl: 58 }),
	gear(2004, 'Dracorian Gauntlets', ItemType.ItemTypeHands, [drop(STRATHOLME, RIVENDARE)], { ilvl: 63, reqLvl: 58 }),
	gear(2005, 'Reiver Claws', ItemType.ItemTypeHands, [drop(BLACKROCK_SPIRE, WYRMTHALAK)], { ilvl: 60, reqLvl: 55 }),
	gear(2006, 'Mark of the Dragon Lord', ItemType.ItemTypeFinger, [drop(BLACKROCK_SPIRE, WYRMTHALAK)], {
		armorType: ArmorType.ArmorTypeUnknown,
		quality: ItemQuality.ItemQualityEpic,
		ilvl: 63,
		reqLvl: 58,
	}),
	gear(UNTICKED_ITEM_ID, 'Death Grips', ItemType.ItemTypeHands, [drop(STRATHOLME, BARTHILAS)], { ilvl: 61, reqLvl: 56 }),
	gear(2008, 'Robe of the Archmage', ItemType.ItemTypeChest, [drop(BLACKROCK_SPIRE, DRAKKISATH)], {
		classAllowlist: [Class.ClassMage],
		ilvl: 62,
		reqLvl: 57,
	}),
	gear(2009, 'Cloak of the Eagle', ItemType.ItemTypeBack, [drop(STRATHOLME, RIVENDARE)], { randomSuffixOptions: [1, 2], ilvl: 60, reqLvl: 55 }),
	gear(ALLIANCE_ITEM_ID, 'Lionheart Gauntlets', ItemType.ItemTypeHands, [drop(BLACKROCK_SPIRE, REND)], {
		factionRestriction: UIItem_FactionRestriction.ALLIANCE_ONLY,
		ilvl: 63,
		reqLvl: 58,
	}),
	gear(2010, 'Sulfuras Hand', ItemType.ItemTypeWaist, [drop(MOLTEN_CORE, RAGNAROS)], { ilvl: 66, reqLvl: 60 }),
	weapon(2012, 'Hand of Edward the Odd', HandType.HandTypeOneHand, [worldDrop], { weaponType: WeaponType.WeaponTypeMace, ilvl: 62, reqLvl: 57 }),
];

const usableItems = () => dropsUsableItems(dropsItemsFixture(), PlayerSpecs.DpsWarrior, { minIlvl: 0, maxIlvl: 0, maxRequiredLevel: 0 });

/** The Sources tree of the items the class can use, as the setup builds it. */
export const dropsTreeFixture = (): DropsSourceNode[] => dropsSourceTree(usableItems(), dropsNamesFixture());

/** Every dungeon boss ticked and the raid not: the worn weapon, a boss's unticked item, two items the list leaves out and an item from two ticked sources. */
export const dropsItemListFixture = (): DropsItemList => {
	const ticked = new Set(['drop:1583:10429', 'drop:1583:10363', 'drop:1583:9568', 'drop:2017:10440', 'drop:2017:10435']);
	return dropsItemList(
		usableItems().filter(item => item.id !== 2010),
		ticked,
		new Set([UNTICKED_ITEM_ID]),
		{
			playerSpec: PlayerSpecs.DpsWarrior,
			canDualWield: true,
			phase: ContentPhase.ContentPhase1,
			faction: Faction.Alliance,
			wornIds: new Set([WORN_ITEM_ID]),
		},
		item => dropsItemSourceKeys(item, dropsNamesFixture()),
	);
};

/** The Sources tree with the dungeon bosses ticked. */
export const dropsTickedFixture = (): ReadonlySet<string> =>
	new Set(['drop:1583:10429', 'drop:1583:10363', 'drop:1583:9568', 'drop:2017:10440', 'drop:2017:10435']);

const DROPS_CURRENT_DPS = 750;
const DROPS_ITERATIONS = 10_000;

const dropsMetrics = (avg: number) => DistributionMetrics.create({ avg, stdev: 100, aggregatorData: { n: DROPS_ITERATIONS, sumSq: 0 } });

/**
 * A finished Drops run with the dungeon bosses ticked, for a warrior wearing Blackhand Doomsaw. Felstriker tops Rend's
 * and Rivendare's cards, Mark of the Dragon Lord has a row per ring slot, and Reiver Claws was culled with an average
 * below the current gear's but a paired difference of +3, an upgrade. Dracorian Gauntlets was culled below the current
 * gear, and Drakkisath's Blackblade of Shahram loses, so his card has no upgrade.
 */
export const dropsResultsFixture = (): { dropsResults: BulkEncounterResults[]; dropsRunSources: string[] } => {
	const items = new Map(dropsItemsFixture().map(item => [item.id, item]));
	const equipped = (id: number) => new EquippedItem({ item: items.get(id)! });
	const ring = (id: number) => new EquippedItem({ item: makeUIItem({ id, name: `Band ${id}`, type: ItemType.ItemTypeFinger, sources: [] }) });
	const currentGear = new Gear({
		[ItemSlot.ItemSlotMainHand]: equipped(WORN_ITEM_ID),
		[ItemSlot.ItemSlotFinger1]: ring(2101),
		[ItemSlot.ItemSlotFinger2]: ring(2102),
	});
	const row = (slot: ItemSlot, id: number, avg: number, extra: Partial<TopGearResult> = {}): TopGearResult => ({
		gear: currentGear.withEquippedItem(slot, equipped(id)),
		dpsMetrics: dropsMetrics(avg),
		pairedErrorToBaseline: 1,
		pairedDpsDiffToBaseline: avg - DROPS_CURRENT_DPS,
		...extra,
	});
	const results = rankEncounterResults(
		[
			{
				reference: true,
				talents: CURRENT_TALENT_SETUP,
				rotation: CURRENT_ROTATION_SETUP,
				settings: null,
				setups: null,
				result: {
					referenceDpsMetrics: dropsMetrics(DROPS_CURRENT_DPS),
					topGearResults: [
						row(ItemSlot.ItemSlotMainHand, 2001, 780, { backendRank: 0 }),
						row(ItemSlot.ItemSlotFinger1, 2006, 765, { backendRank: 1 }),
						row(ItemSlot.ItemSlotFinger2, 2006, 762.1, { backendRank: 2 }),
						row(ItemSlot.ItemSlotMainHand, 2002, 748, { backendRank: 3 }),
						row(ItemSlot.ItemSlotHands, 2005, 749, { culledStage: BulkSimStage.BulkSimStageMedium, pairedDpsDiffToBaseline: 3 }),
						row(ItemSlot.ItemSlotHands, 2004, 745, { culledStage: BulkSimStage.BulkSimStageLow, pairedErrorToBaseline: 2 }),
					],
					iterations: DROPS_ITERATIONS,
					metrics: {},
				},
			},
		],
		currentGear,
		0,
	);
	return {
		dropsResults: [{ name: '', isCurrent: true, duration: 180, durationVariation: 5, results }],
		dropsRunSources: [...dropsTickedFixture()].sort(),
	};
};

const PCT_DELTA = (text: string, tone: 'positive' | 'negative' | null) => ({ text, tone });

/** The Bosses view's cards: a top card, a card with a worn item, and one with no upgrade. */
export const dropsCardsFixture = (): DropsBossCardViewProps[] => {
	const icon = (name: string) => wowheadIconUrl(name, 'medium');
	const base = { selected: false, onSelect: () => {} };
	return [
		{
			...base,
			rank: 1,
			label: { label: 'Warchief Rend Blackhand', detail: 'Blackrock Spire' },
			strip: [
				{ itemId: 2001, icon: icon('inv_sword_39'), name: 'Felstriker', ilvl: 63, gainPct: 4, upgrade: true },
				{ itemId: 2003, icon: icon('inv_weapon_halberd_04'), name: 'Blackhand Doomsaw', ilvl: 63, gainPct: null, upgrade: false },
			],
			best: { dps: 780, pct: 4 },
		},
		{
			...base,
			rank: 2,
			label: { label: 'Overlord Wyrmthalak', detail: 'Blackrock Spire' },
			strip: [
				{ itemId: 2006, icon: icon('inv_jewelry_ring_36'), name: 'Mark of the Dragon Lord', ilvl: 63, gainPct: 2, upgrade: true },
				{ itemId: 2005, icon: icon('inv_gauntlets_29'), name: 'Reiver Claws', ilvl: 60, gainPct: 0.4, upgrade: true },
			],
			best: { dps: 765, pct: 2 },
		},
		{
			...base,
			rank: null,
			label: { label: 'General Drakkisath', detail: 'Blackrock Spire' },
			strip: [{ itemId: 2002, icon: icon('inv_sword_43'), name: 'Blackblade of Shahram', ilvl: 63, gainPct: -0.27, upgrade: false }],
			best: null,
		},
	];
};

/** A Drops result row's view props, and the label of the item it adds; null on the current gear's row. */
export type DropsRowFixture = Pick<BulkResultRowViewProps, 'avg' | 'plusMinus' | 'isBase' | 'delta' | 'significance' | 'lowPrecision'> & {
	label: DropsItemLabelProps | null;
};

const RULED_OUT_EARLY = { note: 'Ruled out early', tooltip: 'Clearly below your current gear, so the run stopped simming it before the final stage.' };

/** The Items view's rows: an item upgrading in a slot, the current gear's row, a row ruled out early and a loss. */
export const dropsRowsFixture = (): DropsRowFixture[] => {
	const items = new Map(dropsItemsFixture().map(item => [item.id, item]));
	const equipped = (id: number) => new EquippedItem({ item: items.get(id)! });
	const base = { isBase: false, significance: 'Difference is significantly different (Z = 30.000).' };
	return [
		{
			...base,
			label: { item: equipped(2001), slot: 'Main Hand', source: 'Warchief Rend Blackhand', moreSources: ['Baron Rivendare'] },
			avg: 780,
			plusMinus: 2,
			delta: PCT_DELTA('+30.00 (4.00%)', 'positive'),
		},
		{
			...base,
			label: { item: equipped(2006), slot: 'Finger 1', source: 'Overlord Wyrmthalak', moreSources: [] },
			avg: 765,
			plusMinus: 2,
			delta: PCT_DELTA('+15.00 (2.00%)', 'positive'),
		},
		{ ...base, isBase: true, label: null, avg: 750, plusMinus: 2, delta: PCT_DELTA('+0.00 (0.00%)', null), significance: '' },
		{
			...base,
			label: { item: equipped(2005), slot: 'Hands', source: 'Overlord Wyrmthalak', moreSources: [] },
			avg: 749,
			plusMinus: 0,
			lowPrecision: RULED_OUT_EARLY,
			delta: PCT_DELTA('+3.00 (0.40%)', 'positive'),
		},
		{
			...base,
			label: { item: equipped(2004), slot: 'Hands', source: 'Baron Rivendare', moreSources: [] },
			avg: 745,
			plusMinus: 0,
			lowPrecision: RULED_OUT_EARLY,
			delta: PCT_DELTA('-5.00 (0.67%)', 'negative'),
			significance: 'Difference is significantly different (Z = 2.500).',
		},
	];
};
