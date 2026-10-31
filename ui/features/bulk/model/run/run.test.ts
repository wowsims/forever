import { saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions, BulkSettings, BulkSimStage, ProgressMetrics } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Encounter, type EquipmentSpec, ItemSlot, ItemSpec, ItemType, Profession, Race, Stat, Target } from '@generated/proto/common';
import {
	BulkBuildBundle,
	BulkEnchantPreset,
	BulkEncounterRef,
	BulkGearSetRef,
	BulkRotationRef,
	BulkSettingsRef,
	BulkTalentRef,
	SavedEncounter,
	SavedGearSet,
	SavedRotation,
	SavedSettings,
	SavedTalents,
	UIEnchant,
} from '@generated/proto/ui';
import type { BulkEncounterResults, BulkSimOverrides } from '@sim/bulk/types';
import { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { Stats } from '@sim/proto/stats';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { type BulkMode, patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { beforeEach, describe, expect, it, type MockInstance, vi } from 'vitest';

import { setBulkMode } from '../mode';
import { compareMatrix } from '../results/compare';
import { runCoreBulkSim } from './core_sim';
import { bulkProgress, cancelBulkBatch, refreshBulkCombinations, runBulkBatch } from './run';
import { setBulkTopResults } from './top_results';

// i18n is not mocked: the test loader has no resources, so every label reads as its key.
vi.mock('./core_sim', () => ({ runCoreBulkSim: vi.fn() }));
vi.mock('../../../../tracking/utils', () => ({ trackEvent: () => {} }));
// The stub host has none of the getters the live settings are read from, so they read as the defaults.
vi.mock('@features/settings/model/saved_settings', async importOriginal => {
	const { SavedSettings: Settings } = await import('@generated/proto/ui');
	return { ...(await importOriginal<object>()), readSavedSettings: () => Settings.create() };
});

const AGILITY = 10;

describe('bulk requests', () => {
	it('carry the enchant pool', async () => {
		const player = makeBulkPlayer({
			getGear: () => ({ getEquippedItem: () => null }),
			getEnchants: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [UIEnchant.create({ effectId: AGILITY })] : []),
		});
		saveEnchantPreset(player, { name: 'Undead', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }] });
		setEnchantPresetChecked(player, 'Undead', true);
		const getBulkCombinationCount = vi.fn(async (_settings: BulkSettings) => ({ combinations: 2 }));

		await refreshBulkCombinations({
			player,
			sim: { getBulkCombinationCount },
			handleCrash: (e: unknown) => Promise.reject(e),
		} as unknown as IndividualSimHost<any>);

		expect(getBulkCombinationCount.mock.calls[0][0].enchants).toEqual([
			BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }),
		]);
	});

	it('keep a selection with no combination as its reason beside the count, not a crash, and clear it once it gives some', async () => {
		const player = makeBulkPlayer({ getGear: () => ({ getEquippedItem: () => null }) });
		const reason = 'no equippable pair of items available for grouped bulk slot 11';
		const getBulkCombinationCount = vi.fn(async () => ({ combinations: 0, error: { type: 0, message: reason } }));
		const handleCrash = vi.fn();
		const host = { player, sim: { getBulkCombinationCount }, handleCrash } as unknown as IndividualSimHost<any>;

		await refreshBulkCombinations(host);
		expect(handleCrash).not.toHaveBeenCalled();
		expect(bulkState(player)).toMatchObject({ combinations: 0, combinationsError: reason, combinationsPending: false });

		getBulkCombinationCount.mockResolvedValue({ combinations: 4, error: undefined } as never);
		await refreshBulkCombinations(host);
		expect(bulkState(player)).toMatchObject({ combinations: 4, combinationsError: null });
	});
});

const CORE_RESULT = { referenceDpsMetrics: { avg: 1000, stdev: 1 }, topGearResults: [], iterations: 100, metrics: {} };
const SAVED_KEY = 'saved-encounters';
const SAVED_TALENTS_KEY = 'saved-talents';
const SAVED_GEAR_KEY = 'saved-gear';
const SAVED_SETTINGS_KEY = 'saved-settings';
const SAVED_ROTATIONS_KEY = 'saved-rotations';
const LIVE = '30305001302';
const ARMS = '3030500130';
const FURY = '-0505000540501';
const FURY_EDITED = '-0505000530501';

const CURRENT = Encounter.create({ duration: 180, durationVariation: 5, targets: [Target.create()] });
const PATCHWERK = Encounter.create({ duration: 300, durationVariation: 30, useHealth: true, targets: [Target.create()] });
const COUNCIL = Encounter.create({ duration: 90, durationVariation: 10, targets: [Target.create()] });

const HEALER = SavedSettings.create({ race: Race.RaceTroll, professions: [Profession.Alchemy, Profession.Herbalism] });

const CLOAK_AGILITY = makeUIEnchant({ effectId: 849, name: 'Cloak agility', type: ItemType.ItemTypeBack });
const CLOAK_ARMOR = makeUIEnchant({ effectId: 884, name: 'Cloak armor', type: ItemType.ItemTypeBack });
const RING_STAMINA = makeUIEnchant({ effectId: 7301, name: 'Ring stamina', type: ItemType.ItemTypeFinger, requiredProfession: Profession.Enchanting });
const SLOT_ENCHANTS = (slot: ItemSlot) =>
	slot === ItemSlot.ItemSlotBack ? [CLOAK_AGILITY, CLOAK_ARMOR] : slot === ItemSlot.ItemSlotFinger1 ? [RING_STAMINA] : [];

const cloak = (enchant?: UIEnchant) => makeEquippedItem({ item: { id: 101, type: ItemType.ItemTypeBack, unique: false }, enchant });
const chest = () => makeEquippedItem({ item: { id: 102, type: ItemType.ItemTypeChest, unique: false } });
const ring = () => makeEquippedItem({ item: { id: 103, type: ItemType.ItemTypeFinger, unique: false } });

// The current gear already wears the Agility preset's cloak enchant; the Raid set does not.
const CURRENT_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(CLOAK_AGILITY) });
const CURRENT_WITH_ARMOR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(CLOAK_ARMOR) });
const RAID_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(), [ItemSlot.ItemSlotChest]: chest(), [ItemSlot.ItemSlotFinger1]: ring() });
const RAID_WITH_AGILITY = new Gear({ [ItemSlot.ItemSlotBack]: cloak(CLOAK_AGILITY), [ItemSlot.ItemSlotChest]: chest(), [ItemSlot.ItemSlotFinger1]: ring() });
const RAID_WITH_ARMOR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(CLOAK_ARMOR), [ItemSlot.ItemSlotChest]: chest(), [ItemSlot.ItemSlotFinger1]: ring() });
const KNOWN_GEAR = [CURRENT_GEAR, RAID_GEAR];

const STRENGTH_BONUS = new Stats().withStat(Stat.StatStrength, 10);

const AGILITY_PRESET = BulkEnchantPreset.create({
	name: 'Agility',
	slots: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY.effectId] })],
});
const ARMOR_PRESET = BulkEnchantPreset.create({
	name: 'Armor',
	slots: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_ARMOR.effectId] })],
});
const RINGS_PRESET = BulkEnchantPreset.create({
	name: 'Rings',
	slots: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotFinger1, effectIds: [RING_STAMINA.effectId] })],
});

const LIVE_ROTATION = APLRotation.create({ type: APLRotationType.TypeAPL, priorityList: [APLListItem.create({ notes: 'live' })] });
const OPENER = APLRotation.create({ type: APLRotationType.TypeAPL, priorityList: [APLListItem.create({ notes: 'opener' })] });

const resolvedFor = (notes: string) => APLRotation.create({ priorityList: [APLListItem.create({ notes })] });

const CURRENT_ENCOUNTER_REF = BulkEncounterRef.create();
const eref = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });
const COUNCIL_REF = eref('Council');

const RAID_SET = BulkGearSetRef.create({ name: 'Raid', preset: true });
const BUFFED_SET = BulkGearSetRef.create({ name: 'Buffed' });
const ARMS_REF = BulkTalentRef.create({ name: 'Arms', preset: true });
const HEALER_REF = BulkSettingsRef.create({ name: 'Healer' });

// A part left out is the current one.
const bundleOf = (parts: Partial<Pick<BulkBuildBundle, 'gearSet' | 'talents' | 'rotation' | 'settings'>>) =>
	BulkBuildBundle.create({ ...parts, checked: true });

const RAID_BUNDLE = bundleOf({ gearSet: RAID_SET });
const ARMS_BUNDLE = bundleOf({ talents: ARMS_REF });

const writeSaved = <T extends object>(host: IndividualSimHost<any>, key: string, type: { toJson: (message: T) => unknown }, record: Record<string, T>) =>
	host.sim.env.storage.setItem(key, JSON.stringify(Object.fromEntries(Object.entries(record).map(([name, data]) => [name, type.toJson(data)]))));

const writeSavedTalents = (host: IndividualSimHost<any>, record: Record<string, string>) =>
	writeSaved(
		host,
		SAVED_TALENTS_KEY,
		SavedTalents,
		Object.fromEntries(Object.entries(record).map(([name, talentsString]) => [name, SavedTalents.create({ talentsString })])),
	);

const BUFFED = SavedGearSet.create({ gear: RAID_GEAR.asSpec(), bonusStatsStats: STRENGTH_BONUS.toProto() });

interface BatchOptions {
	mode?: BulkMode;
	powerTools?: boolean;
	bundles?: BulkBuildBundle[];
	currentBundleChecked?: boolean;
	encounters?: BulkEncounterRef[];
	buildEnchantPresets?: string[];
	enchanter?: boolean;
}

// A non-enchanter unless `enchanter`, with live bonus stats of zero. The resolver gives a rotation by the talents alone.
const batchHost = ({
	mode = 'builds',
	powerTools = true,
	bundles = [RAID_BUNDLE],
	currentBundleChecked = true,
	encounters = [CURRENT_ENCOUNTER_REF],
	buildEnchantPresets = ['Agility'],
	enchanter = false,
}: BatchOptions = {}) => {
	const party = { setBuffs: vi.fn() };
	const player = makeBulkPlayer({
		getGear: () => CURRENT_GEAR,
		getTalentsString: () => LIVE,
		aplRotation: LIVE_ROTATION,
		getBonusStats: () => new Stats(),
		hasProfession: (profession: Profession) => enchanter && profession === Profession.Enchanting,
		getEnchants: SLOT_ENCHANTS,
		getResolvedAplRotationForBuild: vi.fn(({ talentsString }: { talentsString: string }) => resolvedFor(talentsString)),
		getParty: () => party,
		setGearAsync: vi.fn(async () => {}),
		setGear: vi.fn(),
		setTalentsString: vi.fn(),
		setAplRotation: vi.fn(),
		setBonusStats: vi.fn(),
		setRace: vi.fn(),
		setProfessions: vi.fn(),
		setConsumes: vi.fn(),
		setBuffs: vi.fn(),
		sim: {
			db: {
				getEnchants: SLOT_ENCHANTS,
				lookupEquipmentSpec: (spec: EquipmentSpec) =>
					KNOWN_GEAR.find(known => getGearIdentityKey(known.asSpec()) === getGearIdentityKey(spec)) ?? new Gear({}),
			},
		},
	});
	patchSlice(player.sim.store, 'ui', { showPowerTools: powerTools });
	patchBulkState(player, { mode, bundles, currentBundleChecked, encounters, enchantPresets: [AGILITY_PRESET], buildEnchantPresets });
	const host = {
		player,
		sim: {
			store: player.sim.store,
			env: player.sim.env,
			encounter: { toProto: () => Encounter.clone(CURRENT) },
			waitForInit: async () => {},
			signalManager: { abortType: vi.fn(async () => {}) },
			getBulkCombinationCount: vi.fn(async () => ({ combinations: 1 })),
			getShowPowerTools: () => player.sim.store.getState().ui.showPowerTools,
			raid: { setBuffs: vi.fn() },
		},
		individualConfig: {
			presets: {
				encounters: [{ name: 'Patchwerk', encounter: PATCHWERK }],
				talents: [{ name: 'Arms', data: SavedTalents.create({ talentsString: ARMS }) }],
				gear: [{ name: 'Raid', gear: RAID_GEAR.asSpec() }],
				rotations: [{ name: 'Opener', rotation: SavedRotation.create({ rotation: OPENER }) }],
			},
		},
		getSavedEncounterStorageKey: () => SAVED_KEY,
		getSavedTalentsStorageKey: () => SAVED_TALENTS_KEY,
		getSavedGearStorageKey: () => SAVED_GEAR_KEY,
		getSavedSettingsStorageKey: () => SAVED_SETTINGS_KEY,
		getSavedRotationStorageKey: () => SAVED_ROTATIONS_KEY,
		handleCrash: (e: unknown) => Promise.reject(e),
	} as unknown as IndividualSimHost<any>;
	writeSaved(host, SAVED_KEY, SavedEncounter, { Council: SavedEncounter.create({ encounter: COUNCIL }) });
	writeSavedTalents(host, { Fury: FURY });
	writeSaved(host, SAVED_GEAR_KEY, SavedGearSet, { Buffed: BUFFED });
	writeSaved(host, SAVED_SETTINGS_KEY, SavedSettings, { Healer: HEALER });
	return host;
};

const calls = () => vi.mocked(runCoreBulkSim).mock.calls;

const sentEncounters = () => calls().map(call => call[3].encounter);

const sentTalents = () => calls().map(call => call[3].build?.talentsString);

const specsOf = (gears: readonly Gear[] | undefined) => (gears ?? []).map(gear => gear.asSpec());

// Each call's talents, candidates and baseline gear, by spec: a request without Enchanting sends a copy of each.
const sentBuilds = () =>
	calls().map(call => [call[3].build?.talentsString, specsOf(call[3].candidates), call[3].baselineGear ? call[3].baselineGear.asSpec() : undefined]);

const steps = (host: IndividualSimHost<any>) => bulkProgress(host.player)!.steps;

// Each encounter's length and target count as the step row words them: CURRENT, COUNCIL and the health-based PATCHWERK.
const ENCOUNTER_SUMMARIES: Record<string, string> = {
	'bulk_tab.settings.encounters.current': '3:00 ±5s · bulk_tab.settings.encounters.targets',
	Council: '1:30 ±10s · bulk_tab.settings.encounters.targets',
	Patchwerk: 'bulk_tab.settings.encounters.health_summary · bulk_tab.settings.encounters.targets',
};

const step = (
	encounter = 'bulk_tab.settings.encounters.current',
	settings = 'bulk_tab.results.current_settings',
	talents = 'bulk_tab.results.current_talents',
	rotation = 'bulk_tab.results.current_rotation',
) => ({ encounter, encounterSummary: ENCOUNTER_SUMMARIES[encounter], settings, talents, rotation });

const rowsOf = (host: IndividualSimHost<any>) => bulkState(host.player).buildsResults![0].results.chains.flat();

// The sim's rows are the candidates it was sent, each below the reference.
const echoCandidates = async (_context: unknown, _signal: unknown, _settings: unknown, overrides: BulkSimOverrides) => ({
	...CORE_RESULT,
	topGearResults: (overrides.candidates ?? []).map((gear, index) => ({ gear, dpsMetrics: { avg: 900 - index, stdev: 1 } })),
});

describe('runBulkBatch', () => {
	let consoleError: MockInstance;

	beforeEach(() => {
		vi.mocked(runCoreBulkSim).mockReset();
		consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
	});

	it('runs Gear mode as one request with the stored settings and no build or settings override', async () => {
		const host = batchHost({ mode: 'gear', bundles: [ARMS_BUNDLE], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		patchBulkState(host.player, { items: [ItemSpec.create({ id: 5 })] });
		setEnchantPresetChecked(host.player, 'Agility', true);
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(calls()).toHaveLength(1);
		expect(calls()[0][3]).toEqual({ encounter: CURRENT, topResults: 5 });
		expect(calls()[0][2].items).toEqual([ItemSpec.create({ id: 5 })]);
		expect(calls()[0][2].enchants).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [CLOAK_AGILITY.effectId] })]);
		expect(host.sim.getBulkCombinationCount).toHaveBeenCalled();
		expect(steps(host).rows).toEqual([step()]);
		const [entry] = bulkState(host.player).gearResults!;
		expect([entry.isCurrent, entry.results.originalGearResults.settings]).toEqual([true, null]);
		expect(entry.results.originalGearResults.rotation.rotation).toEqual(LIVE_ROTATION);
		expect(entry.results.originalGearResults.rotation.rotation).not.toBe(LIVE_ROTATION);
		expect(bulkState(host.player).bundles).toEqual([ARMS_BUNDLE]);
		expect(bulkState(host.player).encounters).toEqual([CURRENT_ENCOUNTER_REF, COUNCIL_REF]);
	});

	it('sends every gear setup as a candidate with empty settings in Builds mode', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		expect(calls()).toHaveLength(1);
		const [call] = calls();
		expect(call[2]).toEqual(BulkSettings.create());
		// The current gear set with its own enchants, and with the preset's, which it already wears, is the baseline.
		expect(specsOf(call[3].candidates)).toEqual(specsOf([RAID_GEAR, RAID_WITH_AGILITY]));
		expect(call[3].baselineGear).toBeUndefined();
		const backEnchant = (gear: Gear) => gear.getEquippedItem(ItemSlot.ItemSlotBack)?.enchant?.effectId ?? null;
		expect(rowsOf(host).map(row => [row.gearSetup?.gearSet ?? null, backEnchant(row.gear)])).toEqual([
			[null, CLOAK_AGILITY.effectId],
			['Raid', null],
			['Raid', CLOAK_AGILITY.effectId],
		]);
	});

	it('runs each bundle’s gear set with its own enchants with no enchant preset ticked, the live gear as the reference row alone', async () => {
		const host = batchHost({ buildEnchantPresets: [] });
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		expect(calls()).toHaveLength(1);
		expect(specsOf(calls()[0][3].candidates)).toEqual(specsOf([RAID_GEAR]));
		expect(rowsOf(host).map(row => row.gearSetup?.gearSet ?? null)).toEqual([null, 'Raid']);
	});

	it('runs each encounter × bundle group, the Current bundle first, with the bundle’s settings, build override and baseline gear', async () => {
		const host = batchHost({
			bundles: [bundleOf({ gearSet: RAID_SET, talents: ARMS_REF, settings: HEALER_REF })],
			encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF],
		});
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		expect(calls().map(call => [call[3].encounter?.duration, call[3].settings?.race, call[3].build?.talentsString])).toEqual([
			[180, undefined, LIVE],
			[180, Race.RaceTroll, ARMS],
			[90, undefined, LIVE],
			[90, Race.RaceTroll, ARMS],
		]);
		const [reference, arms] = calls();
		expect(reference[3].baselineGear).toBeUndefined();
		expect(reference[3].settings).toBeUndefined();
		expect(reference[3].candidates).toEqual([]);
		expect(arms[3].baselineGear?.asSpec()).toEqual(RAID_GEAR.asSpec());
		expect(specsOf(arms[3].candidates)).toEqual(specsOf([RAID_WITH_AGILITY]));
		expect(arms[3].settings).toEqual(HEALER);
		expect(arms[3].build).toEqual({ talentsString: ARMS, rotation: resolvedFor(ARMS), bonusStats: new Stats().toProto() });
		const results = bulkState(host.player).buildsResults!;
		const settingsOf = (entry: BulkEncounterResults) => [...new Set(entry.results.chains.flat().map(row => row.settings?.name ?? null))].sort();
		expect(results.map(entry => [entry.name, entry.isCurrent, settingsOf(entry)])).toEqual([
			['', true, ['', 'Healer']],
			['Council', false, ['', 'Healer']],
		]);
		expect(results.map(entry => [entry.duration, entry.durationVariation])).toEqual([
			[180, 5],
			[90, 10],
		]);
	});

	it('sims no live gear while the Current bundle is unticked, each other bundle’s request its own first setup as the baseline', async () => {
		const options = { bundles: [bundleOf({ gearSet: RAID_SET, talents: ARMS_REF })], buildEnchantPresets: ['Armor'] };
		const run = async (currentBundleChecked: boolean) => {
			vi.mocked(runCoreBulkSim).mockClear();
			const host = batchHost({ ...options, currentBundleChecked });
			patchBulkState(host.player, { enchantPresets: [ARMOR_PRESET] });
			await runBulkBatch(host);
			return sentBuilds();
		};
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);
		const arms = [ARMS, specsOf([RAID_WITH_ARMOR]), RAID_GEAR.asSpec()];

		expect(await run(true)).toEqual([[LIVE, specsOf([CURRENT_WITH_ARMOR]), undefined], arms]);
		expect(await run(false)).toEqual([arms]);
		// Each request sends its candidates, so the sim never generates any from its baseline gear.
		expect(calls().map(call => Array.isArray(call[3].candidates))).toEqual([true]);
	});

	it('ranks every encounter against its best set while the Current bundle is unticked, with no current row in Compare', async () => {
		const host = batchHost({
			bundles: [bundleOf({ gearSet: RAID_SET, settings: HEALER_REF })],
			currentBundleChecked: false,
			encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF],
		});
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		const results = bulkState(host.player).buildsResults!;
		const reference = ({ results: { originalGearResults, referenceIsCurrent } }: BulkEncounterResults) => [
			referenceIsCurrent,
			originalGearResults.gearSetup?.gearSet,
		];
		expect(results.map(reference)).toEqual([
			[false, 'Raid'],
			[false, 'Raid'],
		]);
		expect(rowsOf(host).map(row => [row.dpsMetrics.avg, row.settings?.name])).toEqual([
			[1000, 'Healer'],
			[900, 'Healer'],
		]);
		const { rows } = compareMatrix(results);
		expect(rows.map(row => row.set?.number)).toEqual([1, 2]);
		expect(rows.map(row => row.cells.map(cell => cell && [cell.rank, cell.gainPct]))).toEqual([
			[
				[1, 0],
				[1, 0],
			],
			[
				[2, -10],
				[2, -10],
			],
		]);
	});

	it('sends no settings for the current settings', async () => {
		const host = batchHost({ encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(sentEncounters()).toEqual([CURRENT, COUNCIL]);
		expect(calls().map(call => call[3].settings)).toEqual([undefined, undefined]);
		expect(bulkState(host.player).buildsResults!.map(entry => entry.results.originalGearResults.settings?.isCurrent)).toEqual([true, true]);
	});

	it('skips an encounter that sims like the ticked current one, whatever it is named, and runs it by name while that is unticked', async () => {
		const run = async (encounters: BulkEncounterRef[]) => {
			vi.mocked(runCoreBulkSim).mockClear();
			const host = batchHost({ encounters });
			writeSaved(host, SAVED_KEY, SavedEncounter, {
				Council: SavedEncounter.create({ encounter: COUNCIL }),
				'Three minutes': SavedEncounter.create({ encounter: CURRENT }),
			});
			await runBulkBatch(host);
			return host;
		};
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		const ticked = await run([CURRENT_ENCOUNTER_REF, eref('Three minutes'), COUNCIL_REF]);
		expect(sentEncounters()).toEqual([CURRENT, COUNCIL]);
		expect(steps(ticked).rows).toEqual([step(), step('Council')]);

		// The saved encounters list Council before "Three minutes", and the batch runs them in list order.
		const unticked = await run([eref('Three minutes'), COUNCIL_REF]);
		expect(sentEncounters()).toEqual([COUNCIL, CURRENT]);
		expect(bulkState(unticked.player).buildsResults!.map(entry => [entry.name, entry.isCurrent])).toEqual([
			['Council', false],
			['Three minutes', false],
		]);
	});

	it('runs a saved talents entry with the talents it holds when the batch starts', async () => {
		const host = batchHost({ bundles: [bundleOf({ talents: BulkTalentRef.create({ name: 'Fury' }) })] });
		writeSavedTalents(host, { Fury: FURY_EDITED });
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(sentTalents()).toEqual([LIVE, FURY_EDITED]);
	});

	it('asks the sim for every candidate of each Builds request by default, and ranks every row', async () => {
		const host = batchHost({ bundles: [RAID_BUNDLE, bundleOf({ gearSet: BUFFED_SET })] });
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		expect(calls().map(call => [call[3].candidates?.length, call[3].topResults])).toEqual([
			[2, 0],
			[1, 0],
		]);
		// Two candidates, the Buffed split's baseline and its candidate, and the reference.
		expect(rowsOf(host)).toHaveLength(5);
	});

	it('sends the Builds top results to every request and ranks that many rows besides the reference', async () => {
		const host = batchHost({ bundles: [RAID_BUNDLE, bundleOf({ gearSet: BUFFED_SET })] });
		setBulkTopResults(host.player, 'builds', 1);
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		expect(calls().map(call => call[3].topResults)).toEqual([1, 1]);
		expect(rowsOf(host)).toHaveLength(2);
	});

	it.each([
		[true, 2],
		[false, 5],
	])('sends Gear mode its stored top results with power tools %s, and 5 without', async (powerTools, sent) => {
		const host = batchHost({ mode: 'gear', powerTools });
		setBulkTopResults(host.player, 'gear', 2);
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(calls()[0][3].topResults).toBe(sent);
	});

	it('runs one request per split inside a run’s single step', async () => {
		const host = batchHost({ bundles: [RAID_BUNDLE, bundleOf({ gearSet: BUFFED_SET })] });
		const running: number[] = [];
		vi.mocked(runCoreBulkSim).mockImplementation(async (...args) => {
			running.push(steps(host).running);
			const result = await echoCandidates(...args);
			// The second request's baseline ranks below the reference, so the rows read in a fixed order.
			return { ...result, referenceDpsMetrics: { avg: running.length === 1 ? 1000 : 950, stdev: 1 } } as never;
		});

		await runBulkBatch(host);

		expect(calls().map(call => call[3].build?.bonusStats)).toEqual([new Stats().toProto(), STRENGTH_BONUS.toProto()]);
		expect(sentBuilds().map(([, candidates, baseline]) => [candidates, baseline])).toEqual([
			[specsOf([RAID_GEAR, RAID_WITH_AGILITY]), undefined],
			[specsOf([RAID_WITH_AGILITY]), RAID_GEAR.asSpec()],
		]);
		expect(running).toEqual([0, 0]);
		expect(steps(host).rows).toHaveLength(1);
		// The reference row once and with no gear setup; the Buffed split's baseline is a row of its own.
		expect(rowsOf(host).map(row => row.gearSetup?.gearSet ?? null)).toEqual([null, 'Buffed', 'Raid', 'Buffed', 'Raid']);
	});

	it('moves the running step’s progress with the iterations its requests sim, each weighted by its estimate, and never back', async () => {
		// Two groups of two splits each, every split three setups: a baseline and two candidates.
		const gearSets = [BulkGearSetRef.create(), RAID_SET, BUFFED_SET, BulkGearSetRef.create({ name: 'Buffed live' })];
		const host = batchHost({
			bundles: [...gearSets.slice(1).map(gearSet => bundleOf({ gearSet })), ...gearSets.map(gearSet => bundleOf({ gearSet, talents: ARMS_REF }))],
		});
		writeSaved(host, SAVED_GEAR_KEY, SavedGearSet, {
			Buffed: BUFFED,
			'Buffed live': SavedGearSet.create({ gear: CURRENT_GEAR.asSpec(), bonusStatsStats: STRENGTH_BONUS.toProto() }),
		});
		const seen: Array<[number, number]> = [];
		const record = () => seen.push([steps(host).running, Math.round(steps(host).progress * 1000) / 1000]);
		const framesAtStart: string[] = [];
		vi.mocked(runCoreBulkSim).mockImplementation(async (...args) => {
			framesAtStart.push(bulkProgress(host.player)!.stage);
			const sims = (args[3].candidates?.length ?? 0) + 1;
			const tick = (bulkStage: BulkSimStage, completedIterations: number, totalIterations: number) => {
				args[0].setSimProgress(ProgressMetrics.create({ bulkStage, completedIterations, totalIterations, totalSims: sims }), {
					currentRound: 1,
					totalRounds: 1,
					useSimCountProgress: true,
				});
				record();
			};
			record();
			// A high stage of 10,000 iterations a sim: the baseline probe, whose total counts every sim plus a spare, then
			// the main pass from the probe's iterations on, then one finalist pass of 6,000 the estimate leaves out.
			tick(BulkSimStage.BulkSimStageHigh, 0, (sims + 1) * 10_000);
			tick(BulkSimStage.BulkSimStageHigh, 10_000, (sims + 1) * 10_000);
			tick(BulkSimStage.BulkSimStageHigh, 10_000, sims * 10_000);
			tick(BulkSimStage.BulkSimStageHigh, 20_000, sims * 10_000);
			tick(BulkSimStage.BulkSimStageHigh, sims * 10_000, sims * 10_000);
			tick(BulkSimStage.BulkSimStageFinalist, 0, 6_000);
			tick(BulkSimStage.BulkSimStageFinalist, 6_000, 6_000);
			return echoCandidates(...args) as never;
		});

		await runBulkBatch(host);

		expect(calls().map(call => call[3].candidates?.length)).toEqual([2, 2, 2, 2]);
		// Two splits of 30,000 estimated iterations each, so each is half of its step.
		const step = (running: number) => [
			[running, 0],
			[running, 0],
			[running, 0.125],
			[running, 0.167],
			[running, 0.333],
			[running, 0.5],
			[running, 0.5],
			[running, 0.5],
			[running, 0.5],
			[running, 0.5],
			[running, 0.625],
			[running, 0.667],
			[running, 0.833],
			[running, 1],
			[running, 1],
			[running, 1],
		];
		expect(seen).toEqual([...step(0), ...step(1)]);
		// A later request starts on the last request's frame, so the bar keeps its readout line.
		expect(framesAtStart).toEqual(['preparing', 'sim', 'sim', 'sim']);
		// Each step names its bundle group and plans both splits. The last frame went out while the second step ran, with
		// the first step's time.
		const { groups, weights, seconds } = steps(host);
		expect(groups).toEqual([0, 1]);
		expect(weights).toEqual([60_000, 60_000]);
		expect(seconds).toHaveLength(1);
	});

	it('gives a setup one Compare row when a bundle without Enchanting sends it without its ring enchants', async () => {
		const host = batchHost({
			enchanter: true,
			bundles: [RAID_BUNDLE, bundleOf({ gearSet: RAID_SET, settings: HEALER_REF })],
			encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF],
			buildEnchantPresets: ['Rings'],
		});
		patchBulkState(host.player, { enchantPresets: [RINGS_PRESET] });
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		const ringEnchant = (gear: Gear | undefined) => gear?.getEquippedItem(ItemSlot.ItemSlotFinger1)?.enchant?.effectId ?? null;
		// Without Enchanting the ring enchant variant equals the set's own and runs as it, the request's baseline.
		expect(calls().map(call => [call[3].candidates?.map(ringEnchant), ringEnchant(call[3].baselineGear)])).toEqual([
			[[null, RING_STAMINA.effectId], null],
			[[], null],
			[[null, RING_STAMINA.effectId], null],
			[[], null],
		]);
		const healerRows = compareMatrix(bulkState(host.player).buildsResults!).rows.filter(row => row.settings?.name === 'Healer');
		expect(healerRows.map(row => [row.gearSetup?.gearSet, ringEnchant(row.set?.gear), row.cells.map(cell => cell !== null)])).toEqual([
			['Raid', null, [true, true]],
		]);
	});

	it('lists one step per encounter and bundle group from the start, encounter outer', async () => {
		const host = batchHost({ bundles: [ARMS_BUNDLE], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		const seen: Array<[number, number]> = [];
		const record = () => {
			const { rows, running } = steps(host);
			seen.push([rows.length, running]);
		};
		// The batch is marked running before it waits for the sim, so this is what the progress dialog finds as it mounts.
		host.sim.waitForInit = async () => record();
		vi.mocked(runCoreBulkSim).mockImplementation(async () => {
			record();
			return CORE_RESULT as never;
		});

		await runBulkBatch(host);

		expect(seen).toEqual([
			[4, 0],
			[4, 0],
			[4, 1],
			[4, 2],
			[4, 3],
		]);
		expect(steps(host).rows).toEqual([step(), step(undefined, undefined, 'Arms'), step('Council'), step('Council', undefined, 'Arms')]);
	});

	it('lists every encounter as a step from the start, skipping a deleted entry, and moves the running one along', async () => {
		const host = batchHost({ encounters: [CURRENT_ENCOUNTER_REF, eref('Patchwerk', true), eref('Gone'), COUNCIL_REF] });
		const seen: Array<[number, number]> = [];
		const record = () => {
			const { rows, running } = steps(host);
			seen.push([rows.length, running]);
		};
		host.sim.waitForInit = async () => record();
		vi.mocked(runCoreBulkSim).mockImplementation(async () => {
			record();
			return CORE_RESULT as never;
		});

		await runBulkBatch(host);

		expect(sentEncounters()).toEqual([CURRENT, PATCHWERK, COUNCIL]);
		expect(seen).toEqual([
			[3, 0],
			[3, 0],
			[3, 1],
			[3, 2],
		]);
		expect(steps(host).rows[1]).toEqual(step('Patchwerk'));
	});

	it('never writes the player, party or raid while running', async () => {
		const opener = { talents: ARMS_REF, rotation: BulkRotationRef.create({ name: 'Opener', preset: true }), settings: HEALER_REF };
		const host = batchHost({
			bundles: [RAID_BUNDLE, bundleOf({ gearSet: BUFFED_SET }), bundleOf({ ...opener, gearSet: RAID_SET }), bundleOf({ ...opener, gearSet: BUFFED_SET })],
			encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF],
		});
		vi.mocked(runCoreBulkSim).mockImplementation(echoCandidates as never);

		await runBulkBatch(host);

		expect(calls()).toHaveLength(8);
		const { player } = host;
		for (const write of [
			player.setGear,
			player.setTalentsString,
			player.setAplRotation,
			player.setBonusStats,
			player.setRace,
			player.setProfessions,
			player.setConsumes,
			player.setBuffs,
			player.getParty()!.setBuffs,
			host.sim.raid.setBuffs,
		]) {
			expect(write).not.toHaveBeenCalled();
		}
		expect(player.setGearAsync).toHaveBeenCalledExactlyOnceWith(CURRENT_GEAR);
	});

	it('does not count Gear-mode combinations in Builds mode', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(host.sim.getBulkCombinationCount).not.toHaveBeenCalled();
		expect(bulkState(host.player).combinationsPending).toBe(false);
	});

	it('runs only the current encounter with power tools off, keeping the stored bundles and encounters for when they are back on', async () => {
		const host = batchHost({ powerTools: false, bundles: [ARMS_BUNDLE], encounters: [CURRENT_ENCOUNTER_REF, eref('Patchwerk', true), COUNCIL_REF] });
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(calls().map(call => call[3])).toEqual([{ encounter: CURRENT, topResults: 5 }]);
		expect(steps(host).rows).toHaveLength(1);
		expect(bulkState(host.player).encounters).toHaveLength(3);
		expect(bulkState(host.player).bundles).toEqual([ARMS_BUNDLE]);

		vi.mocked(runCoreBulkSim).mockClear();
		patchSlice(host.player.sim.store, 'ui', { showPowerTools: true });
		await runBulkBatch(host);

		expect(sentEncounters()).toEqual([CURRENT, CURRENT, PATCHWERK, PATCHWERK, COUNCIL, COUNCIL]);
		expect(sentTalents()).toEqual([LIVE, ARMS, LIVE, ARMS, LIVE, ARMS]);
	});

	it('runs the stored bundles once Gear mode turns to Builds', async () => {
		const host = batchHost({ mode: 'gear', bundles: [ARMS_BUNDLE], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);
		expect(sentEncounters()).toEqual([CURRENT]);

		vi.mocked(runCoreBulkSim).mockClear();
		setBulkMode(host.player, 'builds');
		await runBulkBatch(host);

		expect(sentEncounters()).toEqual([CURRENT, CURRENT, COUNCIL, COUNCIL]);
		expect(sentTalents()).toEqual([LIVE, ARMS, LIVE, ARMS]);
	});

	it('clears and writes only its own mode’s results and started flag, leaving the other mode’s results in place', async () => {
		const host = batchHost();
		const gearResults: BulkEncounterResults[] = [];
		patchBulkState(host.player, { gearResults });
		let whileRunning = bulkState(host.player);
		vi.mocked(runCoreBulkSim).mockImplementation(async () => {
			whileRunning = bulkState(host.player);
			return CORE_RESULT as never;
		});

		await runBulkBatch(host);
		const firstBuildsResults = bulkState(host.player).buildsResults;
		await runBulkBatch(host);

		expect(whileRunning.gearResults).toBe(gearResults);
		expect(whileRunning.buildsResults).toBeNull();
		expect(bulkState(host.player).gearResults).toBe(gearResults);
		expect(bulkState(host.player).buildsResults).toHaveLength(1);
		expect(bulkState(host.player).buildsResults).not.toBe(firstBuildsResults);
		expect([bulkState(host.player).gearStarted, bulkState(host.player).buildsStarted]).toEqual([false, true]);

		const buildsResults = bulkState(host.player).buildsResults;
		setBulkMode(host.player, 'gear');
		await runBulkBatch(host);

		expect(whileRunning.buildsResults).toBe(buildsResults);
		expect(whileRunning.gearResults).toBeNull();
		expect(bulkState(host.player).buildsResults).toBe(buildsResults);
		expect(bulkState(host.player).gearResults).toHaveLength(1);
		expect([bulkState(host.player).gearStarted, bulkState(host.player).buildsStarted]).toEqual([true, true]);
	});

	it('keeps the finished encounters when cancelled', async () => {
		const host = batchHost({ encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		vi.mocked(runCoreBulkSim)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockImplementationOnce(async () => {
				await cancelBulkBatch(host);
				throw new Error('Bulk Sim Aborted');
			});

		await runBulkBatch(host);

		expect(bulkState(host.player).buildsResults!.map(entry => entry.isCurrent)).toEqual([true]);
		expect(bulkState(host.player).isRunning).toBe(false);
	});

	it('keeps the finished encounters when a later one fails', async () => {
		const host = batchHost({ encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		vi.mocked(runCoreBulkSim)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockRejectedValueOnce(new Error('worker crashed'));

		await runBulkBatch(host);

		expect(consoleError).toHaveBeenCalled();
		expect(bulkState(host.player).buildsResults!.map(entry => entry.isCurrent)).toEqual([true]);
	});

	it('leaves no results when the first encounter fails', async () => {
		const host = batchHost({ encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		vi.mocked(runCoreBulkSim).mockRejectedValueOnce(new Error('worker crashed'));

		await runBulkBatch(host);

		expect(consoleError).toHaveBeenCalled();
		expect(bulkState(host.player).buildsResults).toBeNull();
	});

	it('drops an encounter whose requests did not all finish', async () => {
		const host = batchHost({ bundles: [ARMS_BUNDLE], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL_REF] });
		vi.mocked(runCoreBulkSim)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockImplementationOnce(async () => {
				await cancelBulkBatch(host);
				throw new Error('Bulk Sim Aborted');
			});

		await runBulkBatch(host);

		expect(calls()).toHaveLength(4);
		expect(bulkState(host.player).buildsResults!.map(entry => entry.isCurrent)).toEqual([true]);
	});
});
