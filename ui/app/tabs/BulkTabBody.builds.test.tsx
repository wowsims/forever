import { CURRENT_ENCOUNTER_REF } from '@features/bulk/model/builds/encounters';
import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Encounter, EquipmentSpec, ItemSlot, ItemSpec, ItemType, Race, Target } from '@generated/proto/common';
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
} from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import type { PresetGear } from '@sim/presets/types';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import {
	bulkState,
	loadStoredBulkBuildEnchantPresets,
	loadStoredBulkBundles,
	loadStoredBulkEncounters,
	patchBulkState,
	seedBulkSettings,
} from '@sim/settings/bulk_settings';
import { type BulkMode, createSimStore, patchSlice, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { memoryEnv } from '@testing/sim/memory_env';
import { compareGearSet } from '@testing/storybook/fixtures/bulk-compare';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../assets/locales/en/translation.json';
import { BulkTabBody } from './BulkTabBody';

// No i18n mock: the columns show the real copy, and the test setup loads no locale, so this suite loads the English one.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

// Until the sim is ready no gear set is listed; `setReady` lands it.
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

vi.mock('@features/bulk/components/BulkItemSearch', () => ({ BulkItemSearch: () => <div data-testid="bulk-item-search-stub" /> }));
vi.mock('@features/bulk/components/BulkPickerGroups', () => ({ BulkPickerGroups: () => <div data-testid="bulk-picker-groups-stub" /> }));
vi.mock('@features/bulk/components/BulkProgress', () => ({ BulkProgressDialog: () => null }));
vi.mock('@features/bulk/components/BulkSettings', () => ({ BulkSettings: () => <div data-testid="bulk-settings-stub" /> }));
vi.mock('@features/gear/components/SelectorModal', () => ({ SelectorModal: () => null }));
vi.mock('@features/enchant-presets/components/EnchantPresets/EnchantPresetDialog', () => ({
	EnchantPresetDialog: () => <div data-testid="enchant-preset-dialog-stub" />,
}));
vi.mock('@features/import-export', () => ({ BulkGearImporterDialog: () => null }));
// The stub host has none of the raid and player getters the live settings are read from.
vi.mock('@features/settings/hooks/useCurrentSettings', () => ({ useCurrentSettings: () => LIVE_SETTINGS }));

const STORE_KEY = 7;
const KEYS = { gear: 'saved-gear', encounters: 'saved-encounters', settings: 'saved-settings', talents: 'saved-talents', rotations: 'saved-rotations' };

// The stubbed database turns a spec into the set wearing its first item's id as a helm, so the id stands for the set.
const spec = (id: number) => EquipmentSpec.create({ items: [ItemSpec.create({ id })] });

// A second item's enchant the stubbed database puts on a ring the set also wears.
const ringSpec = (id: number, ringEnchant: number) =>
	EquipmentSpec.create({ items: [ItemSpec.create({ id }), ItemSpec.create({ id: 50, enchant: ringEnchant })] });

const gearOfSpec = ({ items: [helm, ring] }: EquipmentSpec) => {
	const gear = compareGearSet(helm?.id ?? 0);
	if (!ring?.enchant) return gear;
	const enchant = makeUIEnchant({ effectId: ring.enchant, type: ItemType.ItemTypeFinger });
	return new Gear({
		...gear.asMap(),
		[ItemSlot.ItemSlotFinger1]: makeEquippedItem({ item: { id: ring.id, type: ItemType.ItemTypeFinger, unique: false }, enchant }),
	});
};

const encounter = (duration: number) => Encounter.create({ duration, targets: [Target.create()] });

const LIVE_SETTINGS = SavedSettings.create({ race: Race.RaceOrc });

const LIVE_ROTATION = APLRotation.create({ type: APLRotationType.TypeAPL, priorityList: [APLListItem.create({ notes: 'live' })] });

const GENERATED = APLRotation.create({ type: APLRotationType.TypeAPL, priorityList: [APLListItem.create({ notes: 'generated' })] });

// As the player's resolver, with one list that every Simple and Auto pick generates; any other rotation goes as it is.
const resolve = ({ rotation }: { rotation: APLRotation }) =>
	rotation.type === APLRotationType.TypeSimple || rotation.type === APLRotationType.TypeAuto ? GENERATED : rotation;

const rotationOf = (type: APLRotationType) => SavedRotation.create({ rotation: APLRotation.create({ type }) });

const enchants = (name: string, effectIds: number[]) =>
	BulkEnchantPreset.create({ name, slots: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotMainHand, effectIds })] });

const GEAR_PRESETS: PresetGear[] = [
	{ name: 'P1 Fury', gear: spec(2) },
	{ name: 'Not for this player', gear: spec(3), enableWhen: () => false },
	{ name: 'Pre-BiS', gear: spec(4) },
];

const PRESETS = {
	gear: GEAR_PRESETS,
	encounters: [{ name: 'Patchwerk', encounter: encounter(300) }],
	settings: [{ name: 'Tank healing', race: Race.RaceHuman }],
	talents: [{ name: 'Arms', data: SavedTalents.create({ talentsString: '3030500130' }) }],
	rotations: [
		{ name: 'Fury APL', rotation: SavedRotation.create({ rotation: APLRotation.create() }) },
		{ name: 'Simple', rotation: rotationOf(APLRotationType.TypeSimple) },
	],
};

const SAVED: Record<keyof typeof KEYS, Record<string, unknown>> = {
	gear: { 'My raid set': SavedGearSet.toJson(SavedGearSet.create({ gear: spec(5) })) },
	encounters: { Council: SavedEncounter.toJson(SavedEncounter.create({ encounter: encounter(90) })) },
	settings: { Healer: SavedSettings.toJson(SavedSettings.create({ race: Race.RaceDwarf })) },
	talents: { 'Deep Cruelty': SavedTalents.toJson(SavedTalents.create({ talentsString: '3030500130-05' })) },
	rotations: {
		'My APL': SavedRotation.toJson(SavedRotation.create({ rotation: APLRotation.create() })),
		Auto: SavedRotation.toJson(rotationOf(APLRotationType.TypeAuto)),
	},
};

const encounterRef = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });

type Part = readonly [name: string, preset: boolean];

// A part left out is the current one.
const bundle = ({ gearSet, talents, rotation, settings }: Partial<Record<'gearSet' | 'talents' | 'rotation' | 'settings', Part>>, checked = true) =>
	BulkBuildBundle.create({
		gearSet: gearSet && BulkGearSetRef.create({ name: gearSet[0], preset: gearSet[1] }),
		talents: talents && BulkTalentRef.create({ name: talents[0], preset: talents[1] }),
		rotation: rotation && BulkRotationRef.create({ name: rotation[0], preset: rotation[1] }),
		settings: settings && BulkSettingsRef.create({ name: settings[0], preset: settings[1] }),
		checked,
	});

const ARMS_FURY = bundle({ talents: ['Arms', true], rotation: ['Fury APL', true] });
const ARMS_FURY_LABEL = 'Current gear · Arms · Fury APL · Current settings';

interface MountOptions {
	powerTools?: boolean;
	mode?: BulkMode;
	presets?: typeof PRESETS | Record<keyof typeof PRESETS, []>;
	saved?: Partial<Record<keyof typeof KEYS, Record<string, unknown>>>;
	bundles?: BulkBuildBundle[];
	currentBundleChecked?: boolean;
	enchantPresets?: BulkEnchantPreset[];
	buildEnchantPresets?: string[];
	encounters?: BulkEncounterRef[];
	currentEncounter?: Encounter;
}

const mount = ({
	powerTools = true,
	mode = 'builds',
	presets = PRESETS,
	saved = SAVED,
	bundles = [ARMS_FURY],
	currentBundleChecked = true,
	enchantPresets = [enchants('Undead', [1898]), enchants('Outdoor', [2504])],
	buildEnchantPresets = [],
	encounters = [],
	currentEncounter,
}: MountOptions = {}) => {
	for (const [kind, key] of Object.entries(KEYS)) localStorage.setItem(key, JSON.stringify(saved[kind as keyof typeof KEYS] ?? {}));
	const store = createSimStore();
	patchSlice(store, 'ui', { showPowerTools: powerTools });
	if (currentEncounter) patchSlice(store, 'encounter', currentEncounter);
	seedKeyed(store, 'players', STORE_KEY, { gear: compareGearSet(1), bonusStats: new Stats(), talentsString: '', v: zeroVersions(PLAYER_FIELDS) } as never);
	const player = {
		sim: {
			store,
			isNative: false,
			env: memoryEnv(),
			db: { lookupEquipmentSpec: gearOfSpec, getEnchants: () => [] },
		},
		storeKey: STORE_KEY,
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		aplRotation: LIVE_ROTATION,
		getResolvedAplRotationForBuild: resolve,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, {
		mode,
		bundles,
		currentBundleChecked,
		enchantPresets,
		buildEnchantPresets,
		encounters: [CURRENT_ENCOUNTER_REF, ...encounters],
	});
	const host = {
		player,
		sim: player.sim,
		individualConfig: { presets },
		getStorageKey: (key: string) => `test-${key}`,
		getSavedGearStorageKey: () => KEYS.gear,
		getSavedEncounterStorageKey: () => KEYS.encounters,
		getSavedSettingsStorageKey: () => KEYS.settings,
		getSavedTalentsStorageKey: () => KEYS.talents,
		getSavedRotationStorageKey: () => KEYS.rotations,
	} as never;
	render(
		<SimHostProvider host={host}>
			<BulkTabBody />
		</SimHostProvider>,
	);
	return player;
};

const column = (name: 'builds' | 'encounters') => screen.getByTestId(`bulk-builds-column-${name}`);
const box = (name: string) => screen.getByRole('checkbox', { name });
// The texts a checkbox is described by: the row's summary, then the duplicate reason when there is one.
const descriptions = (name: string) => (box(name).getAttribute('aria-describedby') ?? '').split(' ').map(id => document.getElementById(id)?.textContent);
const texts = (elements: HTMLElement[]) => elements.map(element => element.textContent);
const listRows = (list: string, row: string) =>
	within(screen.getByTestId(list))
		.getAllByTestId(new RegExp(`^${row}(-current)?$`))
		.map(element => element.querySelector('label')?.textContent);
const bundleRows = () => listRows('bulk-bundles', 'bulk-bundle');
const state = (name: string) => [box(name).getAttribute('aria-checked'), box(name).getAttribute('aria-disabled')];
const storedBundles = (player: Player<any>) => loadStoredBulkBundles(player)?.bundles;

const dialog = () => screen.getByTestId('bulk-bundle-dialog');
const saveButton = () => screen.getByTestId('bulk-bundle-save') as HTMLButtonElement;
const picker = (name: string) => within(dialog()).getByRole('combobox', { name });
const openNewBundle = () => userEvent.click(screen.getByTestId('bulk-bundle-new'));
const editBundle = (label: string) => userEvent.click(within(column('builds')).getByRole('button', { name: `Edit ${label}` }));
const pick = async (name: string, option: string) => {
	await userEvent.click(picker(name));
	await userEvent.click(screen.getByRole('option', { name: option }));
};
const optionsOf = async (name: string) => {
	await userEvent.click(picker(name));
	const options = screen.getAllByRole('option');
	const labels = texts(options);
	await userEvent.click(options[0]);
	return labels;
};

beforeEach(() => {
	localStorage.clear();
	init.ready = true;
});

describe('BulkTabBody in Builds mode', () => {
	it('shows the Builds and Encounters columns in place of the item search and slot pickers in Builds mode', () => {
		mount();

		expect(texts((['builds', 'encounters'] as const).map(name => within(column(name)).getByRole('heading', { level: 5 })))).toEqual([
			'Builds',
			'Encounters',
		]);
		expect(
			within(column('builds')).getByText(
				"Each bundle runs its gear set with its own talents, rotation and settings, and also tries each ticked enchant preset's enchants in their slots.",
			),
		).toBeTruthy();
		expect(within(column('encounters')).getByText('Every ticked bundle runs on each encounter, and each encounter gets its own ranking.')).toBeTruthy();
		expect(screen.queryByTestId('bulk-item-search-stub')).toBeNull();
		expect(screen.queryByTestId('bulk-picker-groups-stub')).toBeNull();
		expect(screen.queryByTestId('bulk-gear-actions')).toBeNull();
		expect(screen.getByTestId('bulk-settings-stub')).toBeTruthy();
	});

	it.each<[string, MountOptions]>([
		['in Gear mode', { mode: 'gear' }],
		['with power tools off', { powerTools: false }],
	])('shows the item search and slot pickers and no columns %s', (_case, options) => {
		mount(options);

		expect(screen.getByTestId('bulk-item-search-stub')).toBeTruthy();
		expect(screen.getByTestId('bulk-picker-groups-stub')).toBeTruthy();
		expect(screen.getByTestId('bulk-gear-actions')).toBeTruthy();
		expect(screen.queryByTestId(/^bulk-builds-column-/)).toBeNull();
	});

	it('shows the Current bundle ticked first and stores it unticked', () => {
		const player = mount();

		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL]);
		expect(state('Current bundle')).toEqual(['true', null]);

		fireEvent.click(box('Current bundle'));

		expect(state('Current bundle')).toEqual(['false', null]);
		expect(bulkState(player).currentBundleChecked).toBe(false);
		expect(loadStoredBulkBundles(player)!.currentUnchecked).toBe(true);
		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL]);
	});

	it('offers Current, then the presets, then the saved entries in each of the dialog’s four pickers, a grouped preset gear set with its group', async () => {
		mount({
			presets: {
				...PRESETS,
				gear: [
					{ name: 'P1 Fury', gear: spec(2) },
					{ name: 'Pre-BiS', group: 'Fury', gear: spec(4) },
				],
			},
		});
		await openNewBundle();

		expect(await optionsOf('Gear set')).toEqual(['Current gear', 'P1 Fury', 'Pre-BiS (Fury)', 'My raid set']);
		expect(await optionsOf('Talents')).toEqual(['Current talents', 'Arms', 'Deep Cruelty']);
		expect(await optionsOf('Rotation')).toEqual(['Current rotation', 'Fury APL', 'Simple', 'My APL', 'Auto']);
		expect(await optionsOf('Settings')).toEqual(['Current settings', 'Tank healing', 'Healer']);
	});

	it('adds a bundle from the dialog, checked, at the end of the list, named from its parts, and stores it', async () => {
		const player = mount();
		await openNewBundle();
		expect(within(dialog()).getByRole('heading', { name: 'New bundle' })).toBeTruthy();

		await pick('Gear set', 'Pre-BiS');
		await pick('Talents', 'Arms');
		await pick('Rotation', 'Fury APL');
		await pick('Settings', 'Tank healing');
		await userEvent.click(saveButton());

		const label = 'Pre-BiS · Arms · Fury APL · Tank healing';
		expect(screen.queryByTestId('bulk-bundle-dialog')).toBeNull();
		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL, label]);
		expect(state(label)).toEqual(['true', null]);
		const added = BulkBuildBundle.create({
			gearSet: BulkGearSetRef.create({ name: 'Pre-BiS', preset: true }),
			talents: BulkTalentRef.create({ name: 'Arms', preset: true }),
			rotation: BulkRotationRef.create({ name: 'Fury APL', preset: true }),
			settings: BulkSettingsRef.create({ name: 'Tank healing', preset: true }),
			checked: true,
		});
		expect(bulkState(player).bundles).toEqual([ARMS_FURY, added]);
		expect(storedBundles(player)).toEqual([ARMS_FURY, added]);
	});

	it('keeps Save off while all four pickers are on Current, in a new bundle and in an edited one', async () => {
		mount();

		await openNewBundle();
		expect(saveButton().disabled).toBe(true);
		await pick('Talents', 'Arms');
		expect(saveButton().disabled).toBe(false);
		await pick('Talents', 'Current talents');
		expect(saveButton().disabled).toBe(true);
		await userEvent.click(within(dialog()).getByRole('button', { name: 'Cancel' }));

		await editBundle(ARMS_FURY_LABEL);
		expect(saveButton().disabled).toBe(false);
		await pick('Talents', 'Current talents');
		await pick('Rotation', 'Current rotation');
		expect(saveButton().disabled).toBe(true);
	});

	it('reopens a bundle’s four picks with its pencil and replaces it on Save, keeping its place and checked state', async () => {
		const unticked = bundle({ gearSet: ['My raid set', false], settings: ['Healer', false] }, false);
		const player = mount({ bundles: [unticked, ARMS_FURY] });
		const label = 'My raid set · Current talents · Current rotation · Healer';

		await editBundle(label);

		expect(within(dialog()).getByRole('heading', { name: 'Edit bundle' })).toBeTruthy();
		expect(['Gear set', 'Talents', 'Rotation', 'Settings'].map(name => picker(name).textContent)).toEqual([
			'My raid set',
			'Current talents',
			'Current rotation',
			'Healer',
		]);

		await pick('Talents', 'Deep Cruelty');
		await userEvent.click(saveButton());

		const edited = 'My raid set · Deep Cruelty · Current rotation · Healer';
		expect(bundleRows()).toEqual(['Current bundle', edited, ARMS_FURY_LABEL]);
		expect(state(edited)).toEqual(['false', null]);
		const replaced = BulkBuildBundle.create({
			gearSet: BulkGearSetRef.create({ name: 'My raid set' }),
			talents: BulkTalentRef.create({ name: 'Deep Cruelty' }),
			rotation: BulkRotationRef.create(),
			settings: BulkSettingsRef.create({ name: 'Healer' }),
			checked: false,
		});
		expect(storedBundles(player)).toEqual([replaced, ARMS_FURY]);
	});

	it('deletes a bundle with its × once confirmed, reachable by keyboard', async () => {
		const player = mount();
		const remove = within(column('builds')).getByRole('button', { name: `Delete ${ARMS_FURY_LABEL}` });

		act(() => box(ARMS_FURY_LABEL).focus());
		await userEvent.tab();
		expect(document.activeElement).toBe(within(column('builds')).getByRole('button', { name: `Edit ${ARMS_FURY_LABEL}` }));
		await userEvent.tab();
		expect(document.activeElement).toBe(remove);
		await userEvent.keyboard('{Enter}');

		expect(screen.getByText(`Delete ${ARMS_FURY_LABEL}?`)).toBeTruthy();
		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL]);

		await userEvent.click(screen.getByRole('button', { name: 'Delete' }));

		expect(bundleRows()).toEqual(['Current bundle']);
		expect(bulkState(player).bundles).toEqual([]);
		expect(storedBundles(player)).toEqual([]);
	});

	it('unchecks a bundle and keeps it in the list', () => {
		const player = mount();

		fireEvent.click(box(ARMS_FURY_LABEL));

		expect(state(ARMS_FURY_LABEL)).toEqual(['false', null]);
		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL]);
		expect(storedBundles(player)).toEqual([{ ...ARMS_FURY, checked: false }]);
	});

	it('disables a bundle that sims the same as an earlier one, unchecked, naming it, and keeps it editable and deletable', () => {
		const player = mount({
			saved: { ...SAVED, gear: { ...SAVED.gear, 'P1 copy': SavedGearSet.toJson(SavedGearSet.create({ gear: spec(2) })) } },
			bundles: [bundle({ gearSet: ['P1 Fury', true] }), bundle({ gearSet: ['P1 copy', false] })],
		});
		const earlier = 'P1 Fury · Current talents · Current rotation · Current settings';
		const later = 'P1 copy · Current talents · Current rotation · Current settings';

		expect([state(earlier), state(later)]).toEqual([
			['true', null],
			['false', 'true'],
		]);
		expect(descriptions(later)).toEqual([`Same as ${earlier}`]);
		const buttons = [`Edit ${later}`, `Delete ${later}`].map(name => within(column('builds')).getByRole('button', { name }) as HTMLButtonElement);
		expect(buttons.map(button => button.disabled)).toEqual([false, false]);

		fireEvent.click(box(earlier));

		expect(state(later)).toEqual(['true', null]);
		expect(bulkState(player).bundles.map(stored => stored.checked)).toEqual([false, true]);
	});

	it('lists the enchant presets under the bundles, and stores a pick without checking the preset for Gear mode', () => {
		const player = mount();

		expect(listRows('bulk-enchant-picks', 'bulk-enchant-pick')).toEqual(['Undead', 'Outdoor']);
		expect(within(column('builds')).getByRole('heading', { level: 6, name: 'Enchant presets' })).toBeTruthy();

		fireEvent.click(box('Outdoor'));

		expect(box('Outdoor').getAttribute('aria-checked')).toBe('true');
		expect(bulkState(player).buildEnchantPresets).toEqual(['Outdoor']);
		expect(loadStoredBulkBuildEnchantPresets(player)?.names).toEqual(['Outdoor']);
		expect(bulkState(player).enchantPresets.map(preset => preset.checked)).toEqual([false, false]);
	});

	it('marks the checkbox of a checked enchant preset equal to an earlier checked one', () => {
		mount({ enchantPresets: [enchants('Undead', [1898]), enchants('Undead copy', [1898])], buildEnchantPresets: ['Undead', 'Undead copy'] });

		expect(['Undead', 'Undead copy'].map(name => box(name).getAttribute('aria-invalid'))).toEqual([null, 'true']);
		expect(descriptions('Undead copy')[1]).toBe('Same enchants as another preset, skipped');
		expect(descriptions('Undead')).toHaveLength(1);
	});

	it('makes the first enchant preset from the Builds column, and drops the button once one is saved', async () => {
		const player = mount({ enchantPresets: [] });

		expect(within(column('builds')).getByText('No enchant presets yet.')).toBeTruthy();
		await userEvent.click(within(column('builds')).getByRole('button', { name: 'New preset' }));
		expect(screen.getByTestId('enchant-preset-dialog-stub')).toBeTruthy();

		act(() => patchBulkState(player, { enchantPresets: [enchants('Undead', [1898])] }));
		expect(listRows('bulk-enchant-picks', 'bulk-enchant-pick')).toEqual(['Undead']);
		expect(within(column('builds')).queryByRole('button', { name: 'New preset' })).toBeNull();
	});

	it('drops a bundle from the list and from storage once its saved gear set is deleted', () => {
		const raid = bundle({ gearSet: ['My raid set', false] });
		const player = mount({ bundles: [raid, ARMS_FURY] });

		expect(bundleRows()).toEqual(['Current bundle', 'My raid set · Current talents · Current rotation · Current settings', ARMS_FURY_LABEL]);

		act(() => {
			localStorage.setItem(KEYS.gear, '{}');
			window.dispatchEvent(new StorageEvent('storage', { key: KEYS.gear }));
		});

		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL]);
		expect(bulkState(player).bundles).toEqual([ARMS_FURY]);
		expect(storedBundles(player)).toEqual([ARMS_FURY]);
	});

	it('keeps every stored bundle until the sim is ready, then drops only one whose saved part is gone', () => {
		init.ready = false;
		const raid = bundle({ gearSet: ['My raid set', false] });
		const gone = bundle({ gearSet: ['Deleted set', false] });
		const player = mount({ bundles: [raid, gone, ARMS_FURY] });

		expect(bundleRows()).toEqual(['Current bundle', ARMS_FURY_LABEL]);
		expect(bulkState(player).bundles).toEqual([raid, gone, ARMS_FURY]);
		expect(storedBundles(player)).toBeUndefined();

		act(setReady);

		expect(bundleRows()).toEqual(['Current bundle', 'My raid set · Current talents · Current rotation · Current settings', ARMS_FURY_LABEL]);
		expect(bulkState(player).bundles).toEqual([raid, ARMS_FURY]);
		expect(storedBundles(player)).toEqual([raid, ARMS_FURY]);
	});

	it('shows the later of two bundles whose rotation picks resolve to the same list disabled, naming the earlier, and keeps it editable and deletable', () => {
		const simple = bundle({ talents: ['Arms', true], rotation: ['Simple', true] });
		const auto = bundle({ talents: ['Arms', true], rotation: ['Auto', false] });
		const player = mount({ bundles: [simple, auto] });
		const earlier = 'Current gear · Arms · Simple · Current settings';
		const later = 'Current gear · Arms · Auto · Current settings';

		expect(bundleRows()).toEqual(['Current bundle', earlier, later]);
		expect([state(earlier), state(later)]).toEqual([
			['true', null],
			['false', 'true'],
		]);
		expect(descriptions(later)).toEqual([`Same as ${earlier}`]);
		const buttons = [`Edit ${later}`, `Delete ${later}`].map(name => within(column('builds')).getByRole('button', { name }) as HTMLButtonElement);
		expect(buttons.map(button => button.disabled)).toEqual([false, false]);
		expect(bulkState(player).bundles).toEqual([simple, auto]);
	});

	it('shows a bundle that sims like the ticked Current bundle disabled, naming it, and enabled once that is unticked, a copy of it naming the bundle it copies', () => {
		const liveCopy = SavedRotation.toJson(SavedRotation.create({ rotation: LIVE_ROTATION }));
		const twin = bundle({ rotation: ['Live copy', false] });
		const copy = bundle({ rotation: ['Live copy 2', false] });
		const player = mount({ saved: { ...SAVED, rotations: { ...SAVED.rotations, 'Live copy': liveCopy, 'Live copy 2': liveCopy } }, bundles: [twin, copy] });
		const twinLabel = 'Current gear · Current talents · Live copy · Current settings';
		const copyLabel = 'Current gear · Current talents · Live copy 2 · Current settings';

		expect(bundleRows()).toEqual(['Current bundle', twinLabel, copyLabel]);
		expect([state(twinLabel), state(copyLabel)]).toEqual([
			['false', 'true'],
			['false', 'true'],
		]);
		expect(descriptions(twinLabel)).toEqual(['Same as Current bundle']);
		expect(descriptions(copyLabel)).toEqual([`Same as ${twinLabel}`]);

		fireEvent.click(box('Current bundle'));

		expect(state('Current bundle')).toEqual(['false', null]);
		expect([state(twinLabel), state(copyLabel)]).toEqual([
			['true', null],
			['false', 'true'],
		]);
		expect(descriptions(copyLabel)).toEqual([`Same as ${twinLabel}`]);
		expect(bulkState(player).bundles).toEqual([twin, copy]);
	});

	it('leaves a bundle on the current gear, talents and rotation with a settings entry holding the live settings enabled while the Current bundle is unticked', () => {
		const settings = { ...SAVED.settings, 'Live settings': SavedSettings.toJson(LIVE_SETTINGS) };
		const live = bundle({ settings: ['Live settings', false] });
		const player = mount({ saved: { ...SAVED, settings }, bundles: [live], currentBundleChecked: false });
		const label = 'Current gear · Current talents · Current rotation · Live settings';

		expect(bundleRows()).toEqual(['Current bundle', label]);
		expect([state('Current bundle'), state(label)]).toEqual([
			['false', null],
			['true', null],
		]);
		expect(bulkState(player).bundles).toEqual([live]);
	});

	it('shows a bundle on a settings entry holding the live settings disabled when it sims like the Current bundle, and one on other settings enabled', () => {
		const rotations = { ...SAVED.rotations, 'Live copy': SavedRotation.toJson(SavedRotation.create({ rotation: LIVE_ROTATION })) };
		const settings = { ...SAVED.settings, 'Live settings': SavedSettings.toJson(LIVE_SETTINGS) };
		const twin = bundle({ rotation: ['Live copy', false], settings: ['Live settings', false] });
		const healer = bundle({ rotation: ['Live copy', false], settings: ['Healer', false] });
		const player = mount({ saved: { ...SAVED, rotations, settings }, bundles: [twin, healer] });
		const twinLabel = 'Current gear · Current talents · Live copy · Live settings';
		const healerLabel = 'Current gear · Current talents · Live copy · Healer';

		expect(bundleRows()).toEqual(['Current bundle', twinLabel, healerLabel]);
		expect([state(twinLabel), state(healerLabel)]).toEqual([
			['false', 'true'],
			['true', null],
		]);
		expect(descriptions(twinLabel)).toEqual(['Same as Current bundle']);
		expect(bulkState(player).bundles).toEqual([twin, healer]);
	});

	it('shows a bundle whose gear is sent like an earlier one of its group disabled, naming it: without Enchanting the ring enchants drop', () => {
		const stamina = bundle({ gearSet: ['Stamina ring', false], talents: ['Arms', true] });
		const agility = bundle({ gearSet: ['Agility ring', false], talents: ['Arms', true] });
		const gear = {
			...SAVED.gear,
			'Stamina ring': SavedGearSet.toJson(SavedGearSet.create({ gear: ringSpec(6, 7301) })),
			'Agility ring': SavedGearSet.toJson(SavedGearSet.create({ gear: ringSpec(6, 7302) })),
		};
		const player = mount({ saved: { ...SAVED, gear }, bundles: [stamina, agility] });
		const earlier = 'Stamina ring · Arms · Current rotation · Current settings';
		const later = 'Agility ring · Arms · Current rotation · Current settings';

		expect([state(earlier), state(later)]).toEqual([
			['true', null],
			['false', 'true'],
		]);
		expect(descriptions(later)).toEqual([`Same as ${earlier}`]);
		expect(bulkState(player).bundles).toEqual([stamina, agility]);
	});

	it('lists the current encounter ticked first, then the preset and saved encounters, each with its length and targets on hover', async () => {
		mount({ currentEncounter: Encounter.create({ duration: 180, durationVariation: 5, targets: [Target.create(), Target.create()] }) });
		const names = ['Current encounter', 'Patchwerk', 'Council'];

		expect(listRows('bulk-encounters', 'bulk-encounter')).toEqual(names);
		expect(texts(within(screen.getByTestId('bulk-encounters')).getAllByRole('heading', { level: 6 }))).toEqual(['Presets', 'Saved']);
		expect(names.map(name => box(name).getAttribute('aria-checked'))).toEqual(['true', 'false', 'false']);
		expect(names.map(name => descriptions(name))).toEqual([['3:00 ±5s · 2 targets'], ['5:00 ±0s · 1 target'], ['1:30 ±0s · 1 target']]);

		const label = within(column('encounters')).getByText('Patchwerk', { selector: 'label' });
		const tooltip = await vi.waitFor(() => {
			fireEvent.mouseEnter(label);
			return screen.getByRole('tooltip');
		});
		expect(tooltip.textContent).toBe('5:00 ±0s · 1 target');
	});

	it('stores a ticked or unticked encounter in list order, the current one included, so it loads back', () => {
		const player = mount();
		const all = [CURRENT_ENCOUNTER_REF, encounterRef('Patchwerk', true), encounterRef('Council')];

		fireEvent.click(box('Council'));
		fireEvent.click(box('Patchwerk'));

		expect(bulkState(player).encounters).toEqual(all);
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual(all);

		fireEvent.click(box('Current encounter'));

		expect(['Current encounter', 'Patchwerk', 'Council'].map(name => box(name).getAttribute('aria-checked'))).toEqual(['false', 'true', 'true']);
		expect(bulkState(player).encounters).toEqual(all.slice(1));
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual(all.slice(1));
	});

	it('disables an encounter that sims like the ticked current one and a later duplicate of an earlier ticked one, naming it, keeping its stored state', () => {
		const player = mount({
			currentEncounter: encounter(300),
			saved: {
				...SAVED,
				encounters: { ...SAVED.encounters, 'Council copy': SavedEncounter.toJson(SavedEncounter.create({ encounter: encounter(90) })) },
			},
			encounters: [encounterRef('Patchwerk', true), encounterRef('Council'), encounterRef('Council copy')],
		});

		expect(listRows('bulk-encounters', 'bulk-encounter')).toEqual(['Current encounter', 'Patchwerk', 'Council', 'Council copy']);
		expect([state('Patchwerk'), state('Council'), state('Council copy')]).toEqual([
			['false', 'true'],
			['true', null],
			['false', 'true'],
		]);
		expect(descriptions('Patchwerk')).toEqual(['5:00 ±0s · 1 target', 'Same as Current encounter']);
		expect(descriptions('Council copy')).toEqual(['1:30 ±0s · 1 target', 'Same as Council']);

		fireEvent.click(box('Council copy'));
		expect(state('Council copy')).toEqual(['false', 'true']);
		expect(bulkState(player).encounters).toEqual([
			CURRENT_ENCOUNTER_REF,
			encounterRef('Patchwerk', true),
			encounterRef('Council'),
			encounterRef('Council copy'),
		]);

		fireEvent.click(box('Council'));

		expect([state('Council'), state('Council copy')]).toEqual([
			['false', null],
			['true', null],
		]);
		expect(descriptions('Council copy')).toEqual(['1:30 ±0s · 1 target']);
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual([CURRENT_ENCOUNTER_REF, encounterRef('Patchwerk', true), encounterRef('Council copy')]);

		fireEvent.click(box('Current encounter'));

		expect(state('Patchwerk')).toEqual(['true', null]);
		expect(descriptions('Patchwerk')).toEqual(['5:00 ±0s · 1 target']);
	});

	it('points to where encounters are saved when there is none to pick', () => {
		mount({ presets: { gear: [], encounters: [], settings: [], talents: [], rotations: [] }, saved: {} });

		expect(listRows('bulk-encounters', 'bulk-encounter')).toEqual(['Current encounter']);
		expect(within(column('encounters')).getByText('No preset or saved encounters yet. Save one in Encounter to run it too.')).toBeTruthy();
	});
});
