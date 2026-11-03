import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Encounter, type EquipmentSpec, HandType, ItemSlot, ItemType, Target } from '@generated/proto/common';
import {
	BulkBuildBundle,
	BulkEnchantPreset,
	BulkEncounterRef,
	BulkGearSetRef,
	BulkRotationRef,
	BulkTalentRef,
	BulkTopResults,
	SavedEncounter,
	SavedGearSet,
	SavedRotation,
	SavedTalents,
	type UIEnchant,
} from '@generated/proto/ui';
import i18n from '@i18n/config';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { Stats } from '@sim/proto/stats';
import { bulkState, loadStoredBulkTopResults, patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { type BulkMode, createSimStore, patchKeyed, patchSlice, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { memoryEnv } from '@testing/sim/memory_env';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { act, cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { BulkBuildsProvider } from '../../context/BulkBuildsContext';
import { setBulkCurrentBundleChecked } from '../../model/builds/bundles';
import { setBulkEnchantPickChecked } from '../../model/builds/enchant_picks';
import { CURRENT_ENCOUNTER_REF, setBulkEncounterChecked } from '../../model/builds/encounters';
import { BuildBundles } from '../BuildBundles';
import type { BulkSettingsViewProps } from './BulkSettings';

const limits = vi.hoisted(() => ({ iterations: null as number | null }));

// Until the sim is ready the harness's `sim.db` is null, as the real one is; `setReady` lands it.
const init = vi.hoisted(() => ({ ready: true, listeners: new Set<() => void>() }));
vi.mock('@sim/hooks/useSimReady', async () => {
	const { useSyncExternalStore } = await import('react');
	const subscribe = (listener: () => void) => {
		init.listeners.add(listener);
		return () => init.listeners.delete(listener);
	};
	return { useSimReady: () => useSyncExternalStore(subscribe, () => init.ready) };
});
const setReady = () => {
	init.ready = true;
	init.listeners.forEach(listener => listener());
};

vi.mock('@sim/bulk/utils', async () => {
	const actual = await vi.importActual<typeof import('@sim/bulk/utils')>('@sim/bulk/utils');
	return { ...actual, getBulkPlayerCanDualWield: () => true, getBulkFreezeWeaponTypes: () => [] };
});
vi.mock('../../model/run/run', () => ({ runBulkBatch: vi.fn() }));
// The stub host has none of the raid and player getters the live settings are read from.
vi.mock('@features/settings/hooks/useCurrentSettings', () => ({ useCurrentSettings: () => null }));
// Passing the web limit takes some hundred candidates in one run, so the warning case sets a limit of its own.
vi.mock('../../model/run/limits', async () => {
	const actual = await vi.importActual<typeof import('../../model/run/limits')>('../../model/run/limits');
	return { ...actual, bulkIterationsLimit: (isNative: boolean | undefined) => limits.iterations ?? actual.bulkIterationsLimit(isNative) };
});

const { BulkSettings, BulkSettingsView } = await import('./BulkSettings');

// No i18n mock: the card shows the real copy, and the test setup loads no locale, so this suite loads the English one.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const STORE_KEY = 5;
const KEYS = {
	gear: 'test-saved-gear',
	encounters: 'test-saved-encounters',
	settings: 'test-saved-settings',
	talents: 'test-saved-talents',
	rotations: 'test-saved-rotations',
};
const LIVE = '30305001302';
const ARMS = '3030500130';

const CLOAK_AGILITY = makeUIEnchant({ effectId: 849, name: 'Cloak agility', type: ItemType.ItemTypeBack });
const CLOAK_ARMOR = makeUIEnchant({ effectId: 884, name: 'Cloak armor', type: ItemType.ItemTypeBack });
const BACK_ENCHANTS = (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [CLOAK_AGILITY, CLOAK_ARMOR] : []);

const cloak = (enchant?: UIEnchant) => makeEquippedItem({ item: { id: 101, type: ItemType.ItemTypeBack, unique: false }, enchant });
const chest = (id: number) => makeEquippedItem({ item: { id, type: ItemType.ItemTypeChest, unique: false } });
const RING = makeEquippedItem({ item: { id: 102, type: ItemType.ItemTypeFinger, unique: false } });
const TWO_HANDER = makeEquippedItem({ item: { id: 104, handType: HandType.HandTypeTwoHand, unique: false } });

// The current gear already wears the Agility cloak enchant; the sets wear none.
const CURRENT_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(CLOAK_AGILITY), [ItemSlot.ItemSlotFinger1]: RING });
const RAID_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(), [ItemSlot.ItemSlotChest]: chest(103) });
const DUNGEON_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(), [ItemSlot.ItemSlotChest]: chest(105) });
const TWO_HAND_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: cloak(), [ItemSlot.ItemSlotMainHand]: TWO_HANDER });
const KNOWN_GEAR = [CURRENT_GEAR, RAID_GEAR, DUNGEON_GEAR, TWO_HAND_GEAR];

const enchantPreset = (name: string, effectIds: number[]) =>
	BulkEnchantPreset.create({ name, slots: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds })] });

const LIVE_ROTATION = APLRotation.create({ type: APLRotationType.TypeAPL, priorityList: [APLListItem.create({ notes: 'live' })] });

// As the player's resolver: an APL goes as it is, and a Simple pick generates its list by whether the gear holds a weapon.
const resolve = ({ gear, rotation }: { talentsString: string; gear: Gear; rotation: APLRotation }) =>
	rotation.type === APLRotationType.TypeSimple
		? APLRotation.create({ priorityList: [APLListItem.create({ notes: gear.getEquippedItem(ItemSlot.ItemSlotMainHand) ? 'two-hand' : 'no weapon' })] })
		: rotation;

const PRESETS = {
	gear: [
		{ name: 'Raid', gear: RAID_GEAR.asSpec() },
		{ name: 'Two-hand', gear: TWO_HAND_GEAR.asSpec() },
	],
	encounters: [],
	talents: [{ name: 'Arms', data: SavedTalents.create({ talentsString: ARMS }) }],
	rotations: [{ name: 'Simple', rotation: SavedRotation.create({ rotation: APLRotation.create({ type: APLRotationType.TypeSimple }) }) }],
};

const SAVED: Record<keyof typeof KEYS, Record<string, unknown>> = {
	gear: {
		'Raid copy': SavedGearSet.toJson(SavedGearSet.create({ gear: RAID_GEAR.asSpec() })),
		Dungeon: SavedGearSet.toJson(SavedGearSet.create({ gear: DUNGEON_GEAR.asSpec() })),
	},
	encounters: { Council: SavedEncounter.toJson(SavedEncounter.create({ encounter: Encounter.create({ duration: 90, targets: [Target.create()] }) })) },
	settings: {},
	talents: {},
	rotations: { 'Live copy': SavedRotation.toJson(SavedRotation.create({ rotation: LIVE_ROTATION })) },
};

const COUNCIL = BulkEncounterRef.create({ name: 'Council' });

const gearSet = (name: string, preset = false) => BulkGearSetRef.create({ name, preset });

// The gear set alone, its other parts current.
const gearSetBundle = (name: string, preset = false) => BulkBuildBundle.create({ gearSet: gearSet(name, preset), checked: true });

const ARMS_SIMPLE_RAID = BulkBuildBundle.create({
	gearSet: gearSet('Raid', true),
	talents: BulkTalentRef.create({ name: 'Arms', preset: true }),
	rotation: BulkRotationRef.create({ name: 'Simple', preset: true }),
	checked: true,
});

interface MountOptions {
	reforger?: unknown;
	powerTools?: boolean;
	mode?: BulkMode;
	combinations?: number;
	bundles?: BulkBuildBundle[];
	buildEnchantPresets?: string[];
	encounters?: BulkEncounterRef[];
	buildBundles?: boolean;
}

const lookupEquipmentSpec = (spec: EquipmentSpec) => KNOWN_GEAR.find(known => getGearIdentityKey(known.asSpec()) === getGearIdentityKey(spec)) ?? new Gear({});

const mount = ({
	reforger = null,
	powerTools = true,
	mode = 'gear',
	combinations = 0,
	bundles = [],
	buildEnchantPresets = [],
	encounters = [CURRENT_ENCOUNTER_REF],
	buildBundles = false,
}: MountOptions = {}) => {
	for (const [kind, key] of Object.entries(KEYS)) localStorage.setItem(key, JSON.stringify(SAVED[kind as keyof typeof KEYS]));
	const store = createSimStore();
	patchSlice(store, 'encounter', { duration: 180, durationVariation: 5, targets: [Target.create()] });
	patchSlice(store, 'ui', { showPowerTools: powerTools });
	seedKeyed(store, 'players', STORE_KEY, {
		gear: CURRENT_GEAR,
		bonusStats: new Stats(),
		talentsString: LIVE,
		profession1: 0,
		profession2: 0,
		v: zeroVersions(PLAYER_FIELDS),
	} as never);
	// Counted: building a gear set's setups reads its slots' enchants, and planning a build resolves its rotation.
	const getEnchants = vi.fn(BACK_ENCHANTS);
	const resolveBuild = vi.fn(resolve);
	const db = { getEnchants, lookupEquipmentSpec };
	const player = {
		sim: {
			store,
			isNative: false,
			env: memoryEnv(),
			get db() {
				return init.ready ? db : null;
			},
		},
		storeKey: STORE_KEY,
		getGear: () => CURRENT_GEAR,
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		aplRotation: LIVE_ROTATION,
		getResolvedAplRotationForBuild: resolveBuild,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, {
		mode,
		combinations,
		bundles,
		enchantPresets: [enchantPreset('Agility', [CLOAK_AGILITY.effectId]), enchantPreset('Cloaks', [CLOAK_AGILITY.effectId, CLOAK_ARMOR.effectId])],
		buildEnchantPresets,
		encounters,
	});
	const host = {
		player,
		sim: player.sim,
		reforger,
		reforgeOptions: null,
		individualConfig: { presets: PRESETS },
		getStorageKey: (key: string) => `test-${key}`,
		getSavedGearStorageKey: () => KEYS.gear,
		getSavedEncounterStorageKey: () => KEYS.encounters,
		getSavedSettingsStorageKey: () => KEYS.settings,
		getSavedTalentsStorageKey: () => KEYS.talents,
		getSavedRotationStorageKey: () => KEYS.rotations,
	} as never;
	const { container } = render(
		<SimHostProvider host={host}>
			<BulkBuildsProvider>
				<BulkSettings />
				{buildBundles && <BuildBundles />}
			</BulkBuildsProvider>
		</SimHostProvider>,
	);
	const pick = async (id: string, index: number) => {
		const trigger = container.querySelector<HTMLElement>(`#${id}`)!;
		await userEvent.click(trigger);
		const option = screen.getAllByRole('option')[index];
		const label = option.textContent;
		await userEvent.click(option);
		return [trigger, label] as const;
	};
	const openGroup = (group: string) => fireEvent.click(container.querySelector(`[data-testid="bulk-settings-group-${group}-trigger"]`)!);
	const groups = () =>
		[...container.querySelectorAll('[data-testid^="bulk-settings-group-"][data-testid$="-trigger"]')].map(trigger => trigger.getAttribute('data-testid'));
	return { player, container, pick, openGroup, groups, getEnchants, resolveBuild };
};

const text = (testId: string) => screen.getByTestId(testId).textContent;
const runsLabel = () => text('bulk-runs-label');
const runsCount = () => text('bulk-runs-count');
const simulate = () => screen.getByTestId('bulk-settings-btn') as HTMLButtonElement;
const topResultsInput = () => within(screen.getByTestId('bulk-top-results')).getByRole('textbox') as HTMLInputElement;
const allCheckbox = () => within(screen.getByTestId('bulk-top-results')).getByRole('checkbox');
const isAll = () => allCheckbox().getAttribute('aria-checked') === 'true';

beforeEach(() => {
	localStorage.clear();
	limits.iterations = null;
	init.ready = true;
});

describe('BulkSettings', () => {
	// Both selects list None, then the pair's first slot (Finger 1, Main Hand), then its second.
	it.each<[string, (player: Player<any>) => unknown, unknown]>([
		['freeze-ring', player => bulkState(player).frozenItems.get(BulkSimItemSlot.ItemSlotFinger), RING],
		['freeze-weapon', player => bulkState(player).frozenWeaponSlot, ItemSlot.ItemSlotMainHand],
	])('keeps the picked slot in the %s select', async (id, stored, expected) => {
		const { player, pick, openGroup } = mount();
		openGroup('freezes');

		const [select, picked] = await pick(id, 1);

		expect(stored(player)).toBe(expected);
		expect(select.textContent).toBe(picked);
	});

	it('opens no group before anything is stored', () => {
		const { container } = mount();

		expect(container.querySelector('#freeze-ring')).toBeNull();
	});

	it('remembers an opened group in local storage', () => {
		const { openGroup } = mount();

		openGroup('freezes');

		expect(JSON.parse(localStorage.getItem('test-__bulkSettingsGroups')!)).toEqual(['freezes']);
	});

	it('shows the enchant and frozen slot groups and no encounter or talent group in Gear mode, even for a spec with the gear optimizer', () => {
		const { groups } = mount({ reforger: { settings: {} }, bundles: [ARMS_SIMPLE_RAID], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL] });

		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger', 'bulk-settings-group-freezes-trigger']);
	});

	it('shows only the run card in Builds mode', () => {
		const { groups } = mount({ mode: 'builds' });

		expect(screen.getByTestId('bulk-run-card')).toBeTruthy();
		expect(screen.queryByTestId('bulk-settings-group-enchants')).toBeNull();
		expect(groups()).toEqual([]);
	});

	it('shows combinations and iterations only with power tools off, and keeps the stored bundles and encounters as they are', () => {
		const encounters = [CURRENT_ENCOUNTER_REF, COUNCIL];
		const { player, groups } = mount({
			powerTools: false,
			mode: 'builds',
			combinations: 2,
			buildEnchantPresets: ['Cloaks'],
			bundles: [ARMS_SIMPLE_RAID],
			encounters,
		});

		// The two combinations and the baseline at the high stage's 10,000.
		expect(text('bulk-combinations')).toBe('2');
		expect(text('bulk-total-iterations')).toBe('30,000');
		expect(screen.queryByTestId('bulk-gear-setups')).toBeNull();
		expect(screen.queryByTestId('bulk-runs-count')).toBeNull();
		expect(groups()).toEqual(['bulk-settings-group-enchants-trigger', 'bulk-settings-group-freezes-trigger']);

		act(() => patchSlice(player.sim.store, 'ui', { showPowerTools: true }));

		// The current gear with its own Agility or with Armor, and Raid bare or with either.
		expect(text('bulk-gear-setups')).toBe('5');
		expect(groups()).toEqual([]);
		expect(bulkState(player).bundles).toEqual([ARMS_SIMPLE_RAID]);
		expect(bulkState(player).encounters).toEqual(encounters);
	});

	it('counts each bundle’s gear set with its own enchants and with each preset cloak enchant, a duplicate bundle left out', () => {
		// The current gear with its own Agility or with Armor; Raid and Dungeon bare, with Agility or with Armor; the Raid
		// copy bundle holds the Raid set with the same parts and does not run.
		mount({
			mode: 'builds',
			bundles: [gearSetBundle('Raid', true), gearSetBundle('Raid copy'), gearSetBundle('Dungeon')],
			buildEnchantPresets: ['Cloaks'],
		});

		expect(text('bulk-gear-setups')).toBe('8');
	});

	it('counts bundles × encounters and the runs they give', () => {
		// The Current bundle and Arms with a Simple rotation on the Raid set, one request each under every encounter.
		mount({ mode: 'builds', bundles: [ARMS_SIMPLE_RAID], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL] });

		expect(runsLabel()).toBe('2 bundles × 2 encounters');
		expect(runsCount()).toBe('4 runs');
		expect(simulate().disabled).toBe(false);
	});

	it('runs one bundle over two encounters', () => {
		mount({ mode: 'builds', encounters: [CURRENT_ENCOUNTER_REF, COUNCIL] });

		expect(text('bulk-gear-setups')).toBe('1');
		expect(runsLabel()).toBe('1 bundle × 2 encounters');
		expect(runsCount()).toBe('2 runs');
		expect(simulate().disabled).toBe(false);
	});

	it('shows the minimum iterations of every run', () => {
		// The current gear with its own Agility enchant is the baseline; the other four setups are one request's candidates.
		mount({ mode: 'builds', bundles: [gearSetBundle('Raid', true)], buildEnchantPresets: ['Cloaks'] });

		expect(text('bulk-total-iterations')).toBe('50,000');
	});

	it('turns Simulate off while the batch holds only the current build, and on once a ticked preset adds a setup to compare', () => {
		const { player } = mount({ mode: 'builds' });

		expect(text('bulk-gear-setups')).toBe('1');
		expect(text('bulk-runs-count')).toBe('1 run');
		expect(simulate().disabled).toBe(true);
		expect(text('bulk-runs-hint')).toBe('Tick a second build to compare.');

		act(() => setBulkEnchantPickChecked(player, 'Cloaks', true, ['Agility', 'Cloaks']));

		expect(text('bulk-gear-setups')).toBe('2');
		expect(simulate().disabled).toBe(false);
		expect(screen.queryByTestId('bulk-runs-hint')).toBeNull();
	});

	it('turns Simulate off and names each column with nothing ticked', () => {
		const { player } = mount({ mode: 'builds' });
		const hints = () => [...screen.getByTestId('bulk-runs-hint').querySelectorAll('p')].map(line => line.textContent);

		expect(hints()).toEqual([]);
		act(() => setBulkCurrentBundleChecked(player, false));

		expect(simulate().disabled).toBe(true);
		expect(hints()).toEqual(['Tick at least one row under Builds to run.']);

		act(() => setBulkEncounterChecked(player, CURRENT_ENCOUNTER_REF, false, []));

		expect(hints()).toEqual(['Tick at least one row under Builds to run.', 'Tick at least one row under Encounters to run.']);
	});

	it('counts a bundle that sims like the Current bundle in its place once the Current bundle is unticked, over two encounters too', () => {
		// The live gear, talents and settings with a saved rotation holding the live list: a twin of the Current bundle.
		const twin = BulkBuildBundle.create({ rotation: BulkRotationRef.create({ name: 'Live copy' }), checked: true });
		const { player } = mount({ mode: 'builds', bundles: [twin], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL] });

		expect(runsLabel()).toBe('1 bundle × 2 encounters');
		expect(simulate().disabled).toBe(false);

		act(() => setBulkCurrentBundleChecked(player, false));

		expect(runsLabel()).toBe('1 bundle × 2 encounters');
		expect(simulate().disabled).toBe(false);
		expect(screen.queryByTestId('bulk-runs-hint')).toBeNull();
		expect(bulkState(player).bundles).toEqual([twin]);
	});

	it('leaves the unticked Current bundle out of the count, Simulate off with one build left and on over two encounters', () => {
		const { player } = mount({ mode: 'builds', bundles: [ARMS_SIMPLE_RAID] });

		expect(runsLabel()).toBe('2 bundles × 1 encounter');
		act(() => setBulkCurrentBundleChecked(player, false));

		expect(runsLabel()).toBe('1 bundle × 1 encounter');
		expect(runsCount()).toBe('1 run');
		expect(simulate().disabled).toBe(true);

		act(() => setBulkEncounterChecked(player, COUNCIL, true, [COUNCIL]));

		expect(runsLabel()).toBe('1 bundle × 2 encounters');
		expect(runsCount()).toBe('2 runs');
		expect(simulate().disabled).toBe(false);
	});

	it('builds no gear setup again and keeps the resolved rotations as an encounter is ticked or the live talents change', () => {
		const { player, getEnchants, resolveBuild } = mount({ mode: 'builds', bundles: [ARMS_SIMPLE_RAID], buildEnchantPresets: ['Cloaks'] });
		const built = getEnchants.mock.calls.length;
		const resolved = resolveBuild.mock.calls.length;
		expect(built).toBeGreaterThan(0);

		act(() => setBulkEncounterChecked(player, COUNCIL, true, [COUNCIL]));

		expect(runsLabel()).toBe('2 bundles × 2 encounters');
		expect(getEnchants.mock.calls.length).toBe(built);
		expect(resolveBuild.mock.calls.length).toBe(resolved);

		const edited = '30305001301';
		act(() => patchKeyed(player.sim.store, 'players', STORE_KEY, { talentsString: edited }));

		// Only the Current bundle's gear under the new talents resolves; its enchant variants and the Arms bundle keep theirs.
		expect(getEnchants.mock.calls.length).toBe(built);
		expect(resolveBuild.mock.calls.slice(resolved).map(([build]) => build.talentsString)).toEqual([edited]);
	});

	it('keeps Simulate off in Gear mode until the sim is ready', () => {
		init.ready = false;
		const { player } = mount();
		act(() => patchBulkState(player, { combinations: 2 }));

		expect(text('bulk-combinations')).toBe('2');
		expect(simulate().disabled).toBe(true);

		act(setReady);

		expect(simulate().disabled).toBe(false);
	});

	it('lists no bundle holding a gear set and shows spinners for the counts in Builds mode until the sim is ready, keeping the bundle stored', () => {
		init.ready = false;
		const raid = gearSetBundle('Raid', true);
		const { player } = mount({ mode: 'builds', buildBundles: true, bundles: [raid], buildEnchantPresets: ['Cloaks'] });

		expect(screen.queryAllByTestId('bulk-bundle')).toHaveLength(0);
		expect(screen.getByTestId('bulk-gear-setups').querySelector('[data-testid="loader"]')).not.toBeNull();
		expect(screen.getByTestId('bulk-total-iterations').querySelector('[data-testid="loader"]')).not.toBeNull();
		expect(simulate().disabled).toBe(true);
		expect(bulkState(player).bundles).toEqual([raid]);

		act(setReady);

		// The Current bundle runs the current gear with its own Agility or with Armor, the Raid bundle its set bare or with
		// either.
		expect(screen.queryAllByTestId('bulk-bundle')).toHaveLength(1);
		expect(text('bulk-gear-setups')).toBe('5');
		expect(simulate().disabled).toBe(false);
		expect(bulkState(player).bundles).toEqual([raid]);
	});

	it('offers the top results only with power tools on, Gear at 5 and Builds at All', () => {
		mount({ powerTools: false });
		expect(screen.queryByTestId('bulk-top-results')).toBeNull();
		cleanup();

		mount();
		expect(topResultsInput().value).toBe('5');
		expect(isAll()).toBe(false);
		cleanup();

		mount({ mode: 'builds' });
		expect(isAll()).toBe(true);
		expect(topResultsInput().disabled).toBe(true);
	});

	it('estimates the Gear minimum iterations with the top results picked, warns past the limit, and stores the pick', () => {
		const { player } = mount({ combinations: 100 });

		// The medium stage at its 1,000 for the 100 and the baseline, then the high stage's 10,000 for 25 and the baseline.
		expect(text('bulk-total-iterations')).toBe('361,000');
		expect(screen.queryByTestId('warning')).toBeNull();
		expect(simulate().disabled).toBe(false);

		fireEvent.click(allCheckbox());

		// Every combination and the baseline at the high stage's 10,000, past the web limit of 1,000,000.
		expect(text('bulk-total-iterations')).toContain('1,010,000');
		expect(screen.getByTestId('warning').getAttribute('aria-label')).toContain('Lower Top results');
		expect(simulate().disabled).toBe(true);
		expect(loadStoredBulkTopResults(player)).toEqual(BulkTopResults.create({ gear: 0, builds: 0 }));

		fireEvent.click(allCheckbox());
		expect(topResultsInput().value).toBe('5');
		expect(simulate().disabled).toBe(false);
		fireEvent.change(topResultsInput(), { target: { value: '3' } });

		expect(bulkState(player).gearTopResults).toBe(3);
		expect(loadStoredBulkTopResults(player)).toEqual(BulkTopResults.create({ gear: 3, builds: 0 }));
	});

	it('keeps the Builds top results apart from the Gear ones', () => {
		const { player } = mount({ mode: 'builds' });

		fireEvent.click(allCheckbox());

		expect(loadStoredBulkTopResults(player)).toEqual(BulkTopResults.create({ gear: 5, builds: 5 }));
		expect(bulkState(player).buildsTopResults).toBe(5);
	});

	it.each([
		[false, false, false],
		[true, true, true],
	])('with power tools %s gates Simulate on a Gear batch past the iterations limit: off %s, hint %s', (powerTools, simulateOff, hint) => {
		// 100 combinations at the default 5 cost 361,000 iterations, within 5,000 combinations but past this limit.
		limits.iterations = 300_000;
		mount({ powerTools, combinations: 100 });

		const warning = screen.getByTestId('warning').getAttribute('aria-label')!;
		expect(simulate().disabled).toBe(simulateOff);
		expect(warning.includes('Lower Top results')).toBe(hint);
	});

	it('judges the warning and Simulate on the largest request’s iterations in Builds mode', () => {
		// Each encounter runs one request with one candidate: 20,000 iterations each, 40,000 in all.
		const options = { mode: 'builds' as const, buildEnchantPresets: ['Cloaks'], encounters: [CURRENT_ENCOUNTER_REF, COUNCIL] };
		limits.iterations = 30_000;
		mount(options);

		expect(text('bulk-total-iterations')).toBe('40,000');
		expect(screen.queryByTestId('warning')).toBeNull();
		expect(simulate().disabled).toBe(false);
		cleanup();

		limits.iterations = 19_999;
		mount(options);

		expect(screen.getByTestId('warning')).toBeTruthy();
		expect(simulate().disabled).toBe(true);
	});
});

const RING_VIEW = {
	id: 'freeze-ring',
	label: 'Freeze ring',
	options: [
		{ value: -1, label: 'None' },
		{ value: ItemSlot.ItemSlotFinger1, label: 'Finger 1' },
	],
	value: -1,
	onChange: () => {},
};

const renderSettings = (props: Partial<BulkSettingsViewProps> = {}) => {
	const handlers = { onRun: vi.fn(), onOpenGroupsChange: vi.fn(), onFrozenWeaponSlotChange: vi.fn() };
	const { container } = render(
		<BulkSettingsView
			topResults={null}
			combinationsCount={<span data-testid="count" />}
			canRun
			openGroups={['freezes']}
			groups
			enchants={null}
			frozenItems={[RING_VIEW]}
			weaponFreeze={null}
			{...handlers}
			{...props}
		/>,
	);
	return { container, ...handlers };
};

describe('BulkSettingsView', () => {
	it('runs the batch from its button, which is disabled while it cannot run', () => {
		const { onRun } = renderSettings();

		fireEvent.click(screen.getByTestId('bulk-settings-btn'));
		expect(onRun).toHaveBeenCalledOnce();
		expect(screen.getByTestId('count')).toBeTruthy();
	});

	it('disables the run button while the batch cannot run', () => {
		renderSettings({ canRun: false });

		expect((screen.getByTestId('bulk-settings-btn') as HTMLButtonElement).disabled).toBe(true);
	});

	it('renders one freeze select per paired slot', () => {
		const { container } = renderSettings();

		expect(container.querySelector('#freeze-ring')).not.toBeNull();
		expect(container.querySelector('#freeze-weapon')).toBeNull();
	});

	it('offers the weapon freeze and its weapon type filters to a dual wielder', async () => {
		const { container, onFrozenWeaponSlotChange } = renderSettings({
			weaponFreeze: { value: -1, weaponTypes: <span data-testid="weapon-types" /> },
		});

		await userEvent.click(container.querySelector<HTMLElement>('#freeze-weapon')!);
		await userEvent.click(screen.getAllByRole('option')[1]);

		expect(onFrozenWeaponSlotChange).toHaveBeenCalledWith(ItemSlot.ItemSlotMainHand);
		expect(screen.getByTestId('weapon-types')).toBeTruthy();
	});
});
