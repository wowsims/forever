import { DistributionMetrics } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { ItemSlot, ItemType, Stat } from '@generated/proto/common';
import { SavedSettings, UIEnchant, UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, BulkGearSetup, BulkPlayerSettingsSetup, BulkRankedResult, BulkRotationSetup } from '@sim/bulk/types';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { Stats } from '@sim/proto/stats';
import { CURRENT_ROTATION_SETUP, CURRENT_SETTINGS_SETUP, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { describe, expect, it } from 'vitest';

import { changedEnchantLabels, changedItemLabels, compareMatrix, gainTint, type MatrixRow } from './compare';
import { type BulkFinishedRequest, rankEncounterResults } from './results';

const equipped = (id: number, enchant?: UIEnchant) =>
	new EquippedItem({ item: UIItem.create({ id, name: `Item ${id}`, type: ItemType.ItemTypeHead, scalingOptions: { 0: { ilvl: 60 } } }), enchant });

// A set wearing `head` and, when given, `hands`; two sets built from the same ids are the same set to the gear identity key.
const gearSet = (head: number, hands?: EquippedItem) => new Gear({ [ItemSlot.ItemSlotHead]: equipped(head), [ItemSlot.ItemSlotHands]: hands ?? null });

const result = (gear: Gear, avg: number, build: Partial<BulkRankedResult> = {}): BulkRankedResult => ({
	gear,
	dpsMetrics: DistributionMetrics.create({ avg, stdev: 1 }),
	talents: CURRENT_TALENT_SETUP,
	rotation: CURRENT_ROTATION_SETUP,
	gearSetup: null,
	settings: null,
	...build,
});

const CURRENT = gearSet(1);
const A = gearSet(2);
const B = gearSet(3);
const C = gearSet(4);

// An encounter ranking `ranked` and the current gear at 1000 DPS, best first, as the run stores it. `baseBuild` gives the
// current gear's row the parts of a Builds batch's reference.
const entry = (
	name: string,
	duration: number,
	ranked: Array<[Gear, number, Partial<BulkRankedResult>?]>,
	baseBuild: Partial<BulkRankedResult> = {},
): BulkEncounterResults => {
	const base = result(CURRENT, 1000, baseBuild);
	const chains = [...ranked.map(([gear, avg, build]) => result(gear, avg, build)), base]
		.sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg)
		.map(ranking => [ranking]);
	return {
		name,
		isCurrent: false,
		duration,
		durationVariation: 5,
		results: { chains, originalGearResults: base, referenceIsCurrent: true, currentGear: CURRENT, iterations: 100 },
	};
};

// A Builds request that sims `baseline` as its baseline and `candidates` beside it, with the Current bundle's parts.
const finishedOf = (reference: boolean, baseline: [BulkGearSetup, number] | null, candidates: Array<[BulkGearSetup, number]>): BulkFinishedRequest => ({
	reference,
	talents: CURRENT_TALENT_SETUP,
	rotation: CURRENT_ROTATION_SETUP,
	settings: CURRENT_SETTINGS_SETUP,
	setups: { byGearKey: new Map(candidates.map(([setup]) => [getGearIdentityKey(setup.gear.asSpec()), setup])), baseline: baseline?.[0] ?? null },
	result: {
		referenceDpsMetrics: DistributionMetrics.create({ avg: baseline?.[1] ?? 1000, stdev: 1 }),
		topGearResults: candidates.map(([setup, avg]) => ({ gear: setup.gear, dpsMetrics: DistributionMetrics.create({ avg, stdev: 1 }) })),
		iterations: 100,
		metrics: {},
	},
});

const rankedEntry = (name: string, duration: number, requests: BulkFinishedRequest[], topResults = 0): BulkEncounterResults => ({
	name,
	isCurrent: false,
	duration,
	durationVariation: 5,
	results: rankEncounterResults(requests, CURRENT, topResults),
});

const ENTRIES = [
	entry('Patchwerk', 300, [
		[C, 1100],
		[A, 1050],
	]),
	entry('Current', 180, [
		[B, 1030],
		[gearSet(2), 990],
	]),
	entry('Council', 90, [[A, 1020]]),
];

describe('compareMatrix', () => {
	const cellsOf = (row: MatrixRow) => row.cells.map(cell => (cell ? [cell.gainPct, cell.tint, cell.best] : null));

	it.each([
		[0, 7],
		[3, 3],
	])('gives each set the ranking keeps at top results %i a row of its own', (topResults, kept) => {
		const setups = [2, 3, 4, 5, 6, 7, 8].map((id): BulkGearSetup => ({ gearSet: `Set ${id}`, gear: gearSet(id), bonusStats: new Stats() }));
		const ranked = rankedEntry(
			'',
			180,
			[
				finishedOf(
					true,
					null,
					setups.map((setup, index): [BulkGearSetup, number] => [setup, 1100 - index * 10]),
				),
			],
			topResults,
		);

		expect(compareMatrix([ranked]).rows.map(row => row.gearSetup?.gearSet ?? 'current')).toEqual([
			'current',
			...setups.slice(0, kept).map(setup => setup.gearSet),
		]);
	});

	it('gives no current gear’s row when the batch did not sim the current setup, each column measured against its best set', () => {
		const [two, three, four] = [2, 3, 4].map((id): BulkGearSetup => ({ gearSet: `Set ${id}`, gear: gearSet(id), bonusStats: new Stats() }));
		const { rows, current } = compareMatrix([
			rankedEntry('Long', 120, [finishedOf(false, [three, 800], [[two, 600]])]),
			rankedEntry('Short', 60, [
				finishedOf(
					false,
					[two, 1000],
					[
						[three, 750],
						[four, 500],
					],
				),
			]),
		]);

		expect(current).toBe(CURRENT);
		expect(rows.map(row => row.set?.gearSetup?.gearSet ?? 'current')).toEqual(['Set 2', 'Set 3', 'Set 4']);
		expect(rows.map(cellsOf)).toEqual([
			[
				[0, 0, true],
				[-25, -5, false],
			],
			[
				[-25, -3, false],
				[0, 0, true],
			],
			[[-50, -5, false], null],
		]);
		expect(rows.map(row => row.cells.map(cell => (cell ? cell.rank : 'gap')))).toEqual([
			[1, 2],
			[2, 1],
			[3, 'gap'],
		]);
	});

	it('orders the columns by duration', () => {
		expect(compareMatrix(ENTRIES).columns.map(column => column.name)).toEqual(['Council', 'Current', 'Patchwerk']);
	});

	it('numbers each set where it first shows, reading the columns in order, and keeps the number in every column', () => {
		const { rows } = compareMatrix(ENTRIES);

		expect(rows.map(row => row.set && [row.set.number, row.set.gear.getEquippedItem(ItemSlot.ItemSlotHead)!.id])).toEqual([null, [1, 2], [2, 3], [3, 4]]);
	});

	it('ranks the sets in each column from 1, the current gear taking no rank', () => {
		const { rows } = compareMatrix(ENTRIES);

		expect(rows.map(row => row.cells.map(cell => (cell ? cell.rank : 'gap')))).toEqual([
			[null, null, null],
			[1, 2, 2],
			['gap', 1, 'gap'],
			['gap', 'gap', 1],
		]);
	});

	it('puts the current gear first, then every set by number, with a gap where an encounter did not rank the set', () => {
		const { rows } = compareMatrix(ENTRIES);

		expect(rows.map(row => row.set?.number ?? 'current')).toEqual(['current', 1, 2, 3]);
		expect(rows.map(cellsOf)).toEqual([
			[
				[0, 0, false],
				[0, 0, false],
				[0, 0, false],
			],
			[
				[2, 5, true],
				[-1, -2, false],
				[5, 3, false],
			],
			[null, [3, 5, true], null],
			[null, null, [10, 5, true]],
		]);
	});

	it('tints each encounter on its own scale, so its best gain takes the deepest tint and a loss stays a loss', () => {
		const { rows } = compareMatrix([
			entry('Short', 60, [
				[A, 1100],
				[B, 1050],
			]),
			entry('Long', 300, [
				[B, 1040],
				[A, 990],
			]),
		]);

		expect(rows.map(cellsOf)).toEqual([
			[
				[0, 0, false],
				[0, 0, false],
			],
			[
				[10, 5, true],
				[-1, -2, false],
			],
			[
				[5, 3, false],
				[4, 5, true],
			],
		]);
	});

	it('marks the current gear best where every set the encounter ranked does worse', () => {
		const [current, set] = compareMatrix([entry('Solo', 60, [[A, 990]])]).rows;

		expect([current.cells[0]!.best, set.cells[0]!.best]).toEqual([true, false]);
	});

	it('keys a set by gear setup, talents, rotation and settings, so the same gear under other settings is a set of its own', () => {
		const opener: BulkRotationSetup = {
			name: 'Opener',
			isCurrent: false,
			rotation: APLRotation.create({ type: APLRotationType.TypeSimple, simple: { specRotationJson: '{"useOpener":true}' } }),
		};
		const healer: BulkPlayerSettingsSetup = { name: 'Healer', isCurrent: false, settings: SavedSettings.create({ reactionTimeMs: 250 }) };
		const raid: BulkGearSetup = { gearSet: 'Raid', gear: A, bonusStats: new Stats().withStat(Stat.StatStrength, 20) };
		const saved: BulkGearSetup = { ...raid, gearSet: 'Saved', bonusStats: new Stats() };
		const raidOnCurrent: BulkGearSetup = { ...raid, gear: CURRENT };
		const current = { settings: CURRENT_SETTINGS_SETUP };
		const { rows } = compareMatrix([
			entry(
				'Short',
				60,
				[
					[A, 1100, { ...current, gearSetup: raid }],
					[A, 1080, { ...current, gearSetup: raid, rotation: opener }],
					[A, 1060, { ...current, gearSetup: saved }],
					[CURRENT, 1050, { ...current, gearSetup: raidOnCurrent }],
					[A, 1040, { gearSetup: raid, settings: healer }],
				],
				current,
			),
			entry(
				'Long',
				120,
				[
					[A, 1090, { gearSetup: raid, settings: healer }],
					[A, 1070, { ...current, gearSetup: raid, rotation: opener }],
				],
				current,
			),
		]);

		expect(rows.map(row => [row.set?.number ?? null, row.gearSetup, row.rotation.name, row.settings])).toEqual([
			[null, null, '', CURRENT_SETTINGS_SETUP],
			[1, raid, '', CURRENT_SETTINGS_SETUP],
			[2, raid, 'Opener', CURRENT_SETTINGS_SETUP],
			[3, saved, '', CURRENT_SETTINGS_SETUP],
			[4, raidOnCurrent, '', CURRENT_SETTINGS_SETUP],
			[5, raid, '', healer],
		]);
		expect(rows.map(row => row.set?.gearSetup ?? null)).toEqual([null, raid, raid, saved, raidOnCurrent, raid]);
		expect(rows.map(row => row.set?.settings ?? null)).toEqual([null, ...Array.from({ length: 4 }, () => CURRENT_SETTINGS_SETUP), healer]);
		expect(rows.map(row => row.cells.map(cell => (cell ? cell.rank : 'gap')))).toEqual([
			[null, null],
			[1, 'gap'],
			[2, 2],
			[3, 'gap'],
			[4, 'gap'],
			[5, 1],
		]);
	});
});

describe('gainTint', () => {
	it('tints next to no change neutral, then one step for each fifth of the encounter’s largest gain', () => {
		const scale = { low: 0, high: 10 };
		const tintOf = (gainPct: number) => gainTint(gainPct, scale);

		expect([0, 0.04, -0.04].map(tintOf)).toEqual([0, 0, 0]);
		expect([0.05, 2, 2.01, 6, 6.01, 8, 8.01, 10].map(tintOf)).toEqual([1, 1, 2, 3, 4, 4, 5, 5]);
	});

	it('measures losses and gains against the largest of either, so a smaller side never reaches the darkest step', () => {
		const scale = { low: -4, high: 2 };
		const tintOf = (gainPct: number) => gainTint(gainPct, scale);

		expect([-4, -1, 1, 2].map(tintOf)).toEqual([-5, -2, 2, 3]);
	});

	it('tints every cell neutral where no set changes anything', () => {
		expect(gainTint(0, { low: 0, high: 0 })).toBe(0);
	});
});

describe('changedItemLabels', () => {
	const pair = (first: ItemSlot, second: ItemSlot, ids: [number, number]) => new Gear({ [first]: equipped(ids[0]), [second]: equipped(ids[1]) });

	it('names the items worn where the current gear wears something else, an enchant change included', () => {
		const crusader = UIEnchant.create({ effectId: 1900, name: 'Crusader' });
		const current = gearSet(1, equipped(7));

		expect(changedItemLabels(gearSet(1, equipped(7)), current, false)).toEqual([]);
		expect(changedItemLabels(gearSet(2, equipped(7, crusader)), current, false)).toEqual(['Item 2', 'Item 7 (Crusader)']);
	});

	it.each([
		['rings', ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2],
		['trinkets', ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	])('names nothing where the %s only change places', (_, first, second) => {
		expect(changedItemLabels(pair(first, second, [11, 10]), pair(first, second, [10, 11]), false)).toEqual([]);
	});

	it('names a new item in a pair', () => {
		const current = pair(ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2, [10, 11]);

		expect(changedItemLabels(pair(ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2, [10, 12]), current, false)).toEqual(['Item 12']);
	});

	it('takes swapped weapons for no change only for a player who can dual wield', () => {
		const current = pair(ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, [20, 21]);
		const swapped = pair(ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, [21, 20]);

		expect(changedItemLabels(swapped, current, true)).toEqual([]);
		expect(changedItemLabels(swapped, current, false)).toEqual(['Item 21', 'Item 20']);
	});
});

describe('changedEnchantLabels', () => {
	const agility = UIEnchant.create({ effectId: 849, name: 'Agility' });
	const strength = UIEnchant.create({ effectId: 856, name: 'Strength' });
	const crusader = UIEnchant.create({ effectId: 1900, name: 'Crusader' });
	const enchanted = (head?: UIEnchant, hands?: UIEnchant) =>
		new Gear({ [ItemSlot.ItemSlotHead]: equipped(1, head), [ItemSlot.ItemSlotHands]: equipped(7, hands) });
	const own = enchanted(agility);
	const labelsOf = (gear: Gear) => changedEnchantLabels({ gear, gearSetup: { gearSet: 'Raid', gear, gearSetGear: own, bonusStats: new Stats() } });

	it('names no enchant when every slot wears the gear set’s own', () => {
		expect(labelsOf(enchanted(agility))).toEqual([]);
	});

	it('names an enchant on a slot the gear set leaves bare, and not one equal to the gear set’s own', () => {
		expect(labelsOf(enchanted(agility, crusader))).toEqual(['slots.hands: Crusader']);
	});

	it('names every slot whose enchant differs by slot, in slot order', () => {
		expect(labelsOf(enchanted(strength, crusader))).toEqual(['slots.head: Strength', 'slots.hands: Crusader']);
	});

	it('names none for a row with no gear setup', () => {
		expect(changedEnchantLabels({ gear: enchanted(strength, crusader), gearSetup: null })).toEqual([]);
	});
});
