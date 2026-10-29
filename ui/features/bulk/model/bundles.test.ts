import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { APLRotation } from '@generated/proto/apl';
import { EquipmentSpec, HealingModel, ItemSlot, ItemSpec, Profession, Race } from '@generated/proto/common';
import {
	BulkBuildBundle,
	BulkGearSetRef,
	BulkRotationRef,
	BulkSettingsRef,
	BulkTalentRef,
	SavedGearSet,
	SavedRotation,
	SavedSettings,
	SavedTalents,
} from '@generated/proto/ui';
import { DpsWarrior } from '@generated/proto/warrior';
import i18n from '@i18n/config';
import type { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { bulkState, loadStoredBulkBundles } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { compareGearSet } from '@testing/storybook/fixtures/bulk-compare';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../assets/locales/en/translation.json';
import {
	addBulkBundle,
	bulkBundleList,
	type BulkBundleParts,
	bulkBundleRuns,
	bundleOptions,
	type CurrentBundle,
	pruneBulkBundles,
	removeBulkBundle,
	replaceBulkBundle,
	setBulkBundleChecked,
	setBulkCurrentBundleChecked,
} from './bundles';
import { bulkGearSetEntries } from './gear_sets';
import { bulkRotationEntries } from './rotations';
import { bulkSettingsEntries } from './settings_entries';
import { bulkTalentEntries } from './talents';

// No i18n mock: the labels are the real copy, and the test setup loads no locale, so this suite loads the English one.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const FURY_TALENTS = '-0505000540501';
const ARMS_TALENTS = '3030500130';
const LIVE_TALENTS = '30305001302';

// One list item whose notes tell rotations apart.
const rotation = (notes: string) => APLRotation.create({ priorityList: [{ notes }] });

const LIVE_ROTATION = rotation('live');
const LIVE_SETTINGS = SavedSettings.create({ race: Race.RaceOrc, professions: [Profession.Alchemy, Profession.Herbalism] });

// The helm id stands for the whole set: the stubbed database turns a spec into the set wearing its first item's id.
const spec = (id: number) => EquipmentSpec.create({ items: [ItemSpec.create({ id })] });

const helmId = (gear: Gear) => gear.getEquippedItem(ItemSlot.ItemSlotHead)?.item.id;

const gearPlayer = () => makeBulkPlayer({ sim: { db: { lookupEquipmentSpec: (equipment: EquipmentSpec) => compareGearSet(equipment.items[0]?.id ?? 0) } } });

const CURRENT: CurrentBundle = {
	gear: compareGearSet(1),
	bonusStats: new Stats(),
	talentsString: LIVE_TALENTS,
	rotation: LIVE_ROTATION,
	settings: LIVE_SETTINGS,
};

// Two grouped presets and two saved sets, "Live copy" wearing the current gear; preset talents and rotations for Fury and
// Arms, the talents "Live" holding the current string, the saved rotation "Live list" holding the current list; preset
// settings for Fury and Arms and a saved "Live settings" entry that sims as the current settings.
const PARTS: BulkBundleParts = {
	gearSets: bulkGearSetEntries(
		[
			{ name: 'Pre-BiS', gear: spec(2), group: 'Fury' },
			{ name: 'Pre-BiS', gear: spec(3), group: 'Arms' },
		],
		[
			{ name: 'Live copy', data: SavedGearSet.create({ gear: spec(1) }) },
			{ name: 'Raid', data: SavedGearSet.create({ gear: spec(4) }) },
		],
		gearPlayer(),
	),
	talents: bulkTalentEntries(
		[
			{ name: 'Fury', data: SavedTalents.create({ talentsString: FURY_TALENTS }) },
			{ name: 'Arms', data: SavedTalents.create({ talentsString: ARMS_TALENTS }) },
			{ name: 'Live', data: SavedTalents.create({ talentsString: LIVE_TALENTS }) },
		],
		[{ name: 'Mine', data: SavedTalents.create({ talentsString: '0505' }) }],
	),
	rotations: bulkRotationEntries(
		[
			{ name: 'Fury APL', rotation: SavedRotation.create({ rotation: rotation('fury') }) },
			{ name: 'Arms APL', rotation: SavedRotation.create({ rotation: rotation('arms') }) },
		],
		[{ name: 'Live list', data: SavedRotation.create({ rotation: LIVE_ROTATION }) }],
	),
	settings: bulkSettingsEntries(
		{
			settings: [
				{ name: 'Pre-BiS - Fury', race: Race.RaceOrc },
				{ name: 'Pre-BiS - Arms', race: Race.RaceHuman },
			],
		} as unknown as SpecPresets,
		[{ name: 'Live settings', data: SavedSettings.create({ ...LIVE_SETTINGS, darkIntentUptime: 50 }) }],
		{} as IndividualSimHost<any>,
	),
};

const gearRef = (name: string, preset = false, group = '') => BulkGearSetRef.create({ name, preset, group });
const talentRef = (name: string, preset = true) => BulkTalentRef.create({ name, preset });
const rotationRef = (name: string, preset = true) => BulkRotationRef.create({ name, preset });
const settingsRef = (name: string, preset = true) => BulkSettingsRef.create({ name, preset });

type BundleRefs = Partial<Pick<BulkBuildBundle, 'gearSet' | 'talents' | 'rotation' | 'settings'>>;

const bundle = (refs: BundleRefs, checked = true) => BulkBuildBundle.create({ ...refs, checked });

const FURY_REFS: BundleRefs = {
	gearSet: gearRef('Pre-BiS', true, 'Fury'),
	talents: talentRef('Fury'),
	rotation: rotationRef('Fury APL'),
	settings: settingsRef('Pre-BiS - Fury'),
};
const FURY = bundle(FURY_REFS);
const ARMS = bundle({
	gearSet: gearRef('Pre-BiS', true, 'Arms'),
	talents: talentRef('Arms'),
	rotation: rotationRef('Arms APL'),
	settings: settingsRef('Pre-BiS - Arms'),
});
const FURY_LABEL = 'Pre-BiS (Fury) · Fury · Fury APL · Pre-BiS - Fury';

const storedOf = (player: ReturnType<typeof makeBulkPlayer>) =>
	bulkState(player).bundles.map(stored => [stored.talents?.name ?? '', stored.settings?.name ?? '', stored.checked]);

describe('bundle pickers and labels', () => {
	it('offers Current first in each picker, then the presets, then the saved entries, a grouped preset gear set named with its group', () => {
		const options = bundleOptions(PARTS);

		expect(options.gearSets.map(option => [option.ref.name, option.ref.preset, option.ref.group ?? '', option.label])).toEqual([
			['', false, '', 'Current gear'],
			['Pre-BiS', true, 'Fury', 'Pre-BiS (Fury)'],
			['Pre-BiS', true, 'Arms', 'Pre-BiS (Arms)'],
			['Live copy', false, '', 'Live copy'],
			['Raid', false, '', 'Raid'],
		]);
		expect(options.talents.map(option => [option.ref.name, option.ref.preset, option.label])).toEqual([
			['', false, 'Current talents'],
			['Fury', true, 'Fury'],
			['Arms', true, 'Arms'],
			['Live', true, 'Live'],
			['Mine', false, 'Mine'],
		]);
		expect(options.rotations.map(option => [option.ref.name, option.ref.preset, option.label])).toEqual([
			['', false, 'Current rotation'],
			['Fury APL', true, 'Fury APL'],
			['Arms APL', true, 'Arms APL'],
			['Live list', false, 'Live list'],
		]);
		expect(options.settings.map(option => [option.ref.name, option.ref.preset, option.label])).toEqual([
			['', false, 'Current settings'],
			['Pre-BiS - Fury', true, 'Pre-BiS - Fury'],
			['Pre-BiS - Arms', true, 'Pre-BiS - Arms'],
			['Live settings', false, 'Live settings'],
		]);
	});

	it('names a bundle from its four parts, a current part reading as current', () => {
		expect(bulkBundleList(CURRENT, PARTS, [FURY, bundle({ gearSet: gearRef('Raid') }), bundle({})], true).map(row => row.label)).toEqual([
			'Pre-BiS (Fury) · Fury · Fury APL · Pre-BiS - Fury',
			'Raid · Current talents · Current rotation · Current settings',
			'Current gear · Current talents · Current rotation · Current settings',
		]);
	});
});

describe('stored bundles', () => {
	it('adds a bundle checked at the end and stores it so it loads back', () => {
		const player = makeBulkPlayer();
		const { settings } = bulkState(player).v;

		addBulkBundle(player, bundle(FURY_REFS, false));
		addBulkBundle(player, ARMS);

		expect(storedOf(player)).toEqual([
			['Fury', 'Pre-BiS - Fury', true],
			['Arms', 'Pre-BiS - Arms', true],
		]);
		expect(loadStoredBulkBundles(player)!.bundles).toEqual(bulkState(player).bundles);
		expect(bulkState(player).v.settings).toBe(settings);
	});

	it('replaces a bundle in place, keeping its checked state, and removes one by its index', () => {
		const player = makeBulkPlayer();
		addBulkBundle(player, FURY);
		addBulkBundle(player, ARMS);
		setBulkBundleChecked(player, 0, false);

		replaceBulkBundle(player, 0, bundle({ ...FURY_REFS, settings: settingsRef('Pre-BiS - Arms') }));

		expect(storedOf(player)).toEqual([
			['Fury', 'Pre-BiS - Arms', false],
			['Arms', 'Pre-BiS - Arms', true],
		]);

		removeBulkBundle(player, 1);

		expect(storedOf(player)).toEqual([['Fury', 'Pre-BiS - Arms', false]]);
		expect(loadStoredBulkBundles(player)!.bundles).toEqual(bulkState(player).bundles);
	});

	it('stores the Current bundle unticked, and ticks it on fresh storage', () => {
		const player = makeBulkPlayer();

		expect(bulkState(player).currentBundleChecked).toBe(true);

		addBulkBundle(player, FURY);
		expect(loadStoredBulkBundles(player)!.currentUnchecked).toBe(false);

		setBulkCurrentBundleChecked(player, false);

		expect(bulkState(player).currentBundleChecked).toBe(false);
		expect(loadStoredBulkBundles(player)!.currentUnchecked).toBe(true);
		expect(loadStoredBulkBundles(player)!.bundles).toEqual([FURY]);
	});
});

describe('bulkBundleList', () => {
	it('flags the later of two bundles with the same parts, naming the earlier, and keeps its checked state', () => {
		const bundles = [bundle(FURY_REFS, false), FURY, ARMS, FURY, bundle(FURY_REFS, false)];

		expect(bulkBundleList(CURRENT, PARTS, bundles, true).map(row => [row.index, row.checked, row.duplicateOf])).toEqual([
			[0, false, null],
			[1, true, null],
			[2, true, null],
			[3, true, FURY_LABEL],
			[4, false, FURY_LABEL],
		]);
		expect(bulkBundleRuns(CURRENT, PARTS, bundles, true).map(run => run.index)).toEqual([-1, 1, 2]);
	});

	it('flags a bundle that sims the same as the ticked Current bundle, whatever its parts are named, and runs it while the Current bundle is unticked', () => {
		const twin = bundle({ gearSet: gearRef('Live copy'), talents: talentRef('Live'), settings: settingsRef('Live settings', false) });
		const arms = bundle({ gearSet: gearRef('Raid'), talents: talentRef('Arms') });

		expect(bulkBundleList(CURRENT, PARTS, [twin], true).map(row => row.duplicateOf)).toEqual(['Current bundle']);
		expect(bulkBundleRuns(CURRENT, PARTS, [twin], true).map(run => run.index)).toEqual([-1]);
		expect(bulkBundleList(CURRENT, PARTS, [twin, twin], false).map(row => row.duplicateOf)).toEqual([
			null,
			'Live copy · Live · Current rotation · Live settings',
		]);
		expect(bulkBundleRuns(CURRENT, PARTS, [twin, arms], false).map(run => run.index)).toEqual([0, 1]);
	});

	it('tells apart two bundles that differ only in their Settings entry', () => {
		const bundles = [FURY, bundle({ ...FURY_REFS, settings: settingsRef('Pre-BiS - Arms') })];

		expect(bulkBundleList(CURRENT, PARTS, bundles, true).map(row => row.duplicateOf)).toEqual([null, null]);
		expect(bulkBundleRuns(CURRENT, PARTS, bundles, true).map(run => run.index)).toEqual([-1, 0, 1]);
	});

	it('tells a Settings entry with other spec options apart from the Current bundle, and one without them sims the live ones', () => {
		const rageSpec = (startingRage: number): SavedSettings['spec'] => ({
			oneofKind: 'dpsWarrior',
			dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage } } }),
		});
		const current: CurrentBundle = { ...CURRENT, settings: SavedSettings.create({ ...LIVE_SETTINGS, spec: rageSpec(50) }) };
		const parts: BulkBundleParts = {
			...PARTS,
			settings: bulkSettingsEntries(
				{} as SpecPresets,
				[
					{ name: 'Other rage', data: SavedSettings.create({ ...LIVE_SETTINGS, spec: rageSpec(0) }) },
					{ name: 'No spec options', data: LIVE_SETTINGS },
				],
				{} as IndividualSimHost<any>,
			),
		};
		const bundles = [bundle({ settings: settingsRef('Other rage', false) }), bundle({ settings: settingsRef('No spec options', false) })];

		expect(bulkBundleList(current, parts, bundles, true).map(row => row.duplicateOf)).toEqual([null, 'Current bundle']);
		expect(bulkBundleRuns(current, parts, bundles, true).map(run => run.index)).toEqual([-1, 0]);
	});

	it('tells the rotations apart as picked: the current rotation and a saved one holding the same list are two bundles', () => {
		const bundles = [
			bundle({ gearSet: gearRef('Raid'), talents: talentRef('Fury') }),
			bundle({ gearSet: gearRef('Raid'), talents: talentRef('Fury'), rotation: rotationRef('Live list', false) }),
		];

		expect(bulkBundleList(CURRENT, PARTS, bundles, true).map(row => row.duplicateOf)).toEqual([null, null]);
		expect(bulkBundleRuns(CURRENT, PARTS, bundles, true).map(run => run.index)).toEqual([-1, 0, 1]);
	});
});

describe('bulkBundleRuns', () => {
	it('runs the Current bundle first, then each checked bundle that duplicates none, with its parts and build key', () => {
		const runs = bulkBundleRuns(CURRENT, PARTS, [FURY, bundle({ ...FURY_REFS }, false), bundle({ gearSet: gearRef('Raid') })], true);

		expect(runs.map(run => run.index)).toEqual([-1, 0, 2]);
		expect(runs[0]).toEqual({
			index: -1,
			gearSet: { name: '', gear: CURRENT.gear, bonusStats: CURRENT.bonusStats },
			talents: { name: '', talentsString: LIVE_TALENTS, isCurrent: true },
			rotation: { name: '', isCurrent: true, rotation: LIVE_ROTATION },
			settings: { name: '', isCurrent: true, settings: null },
			buildKey: runs[0].buildKey,
		});
		expect(runs[1]).toMatchObject({
			gearSet: { name: 'Pre-BiS (Fury)' },
			talents: { name: 'Fury', talentsString: FURY_TALENTS, isCurrent: false },
			rotation: { name: 'Fury APL', isCurrent: false, rotation: rotation('fury') },
			settings: { name: 'Pre-BiS - Fury', isCurrent: false, settings: PARTS.settings[0].settings },
		});
		expect(helmId(runs[1].gearSet.gear)).toBe(2);
		expect(runs[2]).toMatchObject({
			gearSet: { name: 'Raid' },
			talents: { isCurrent: true, talentsString: LIVE_TALENTS },
			rotation: { isCurrent: true, rotation: LIVE_ROTATION },
			settings: { isCurrent: true, settings: null },
			buildKey: runs[0].buildKey,
		});
		expect(helmId(runs[2].gearSet.gear)).toBe(4);
		expect(runs[1].buildKey).not.toBe(runs[0].buildKey);
	});

	it('leaves the unticked Current bundle out, and runs nothing when no other bundle is ticked', () => {
		expect(bulkBundleRuns(CURRENT, PARTS, [ARMS], false).map(run => run.index)).toEqual([0]);
		expect(bulkBundleRuns(CURRENT, PARTS, [{ ...ARMS, checked: false }], false)).toEqual([]);
	});

	it('gives bundles the same build key exactly when their talents, rotation and Settings picks are the same', () => {
		const runs = bulkBundleRuns(
			CURRENT,
			PARTS,
			[
				FURY,
				bundle({ ...FURY_REFS, gearSet: gearRef('Raid') }),
				bundle({ gearSet: gearRef('Raid') }),
				bundle({ ...FURY_REFS, settings: settingsRef('Pre-BiS - Arms') }),
				bundle({ gearSet: gearRef('Pre-BiS', true, 'Arms'), talents: talentRef('Live') }),
			],
			true,
		);
		const keys = runs.map(run => run.buildKey);

		expect(runs.map(run => run.index)).toEqual([-1, 0, 1, 2, 3, 4]);
		expect(keys[1]).toBe(keys[2]);
		expect(keys[0]).toBe(keys[3]);
		expect(new Set([keys[0], keys[1], keys[4], keys[5]]).size).toBe(4);
	});
});

// Each part's entries with one more saved one named "Gone", and a third group of the Pre-BiS preset.
const PARTS_WITH_GONE: BulkBundleParts = {
	gearSets: [
		...PARTS.gearSets,
		{ ref: gearRef('Gone'), name: 'Gone', gear: compareGearSet(8), bonusStats: new Stats() },
		{ ref: gearRef('Pre-BiS', true, 'Prot'), name: 'Pre-BiS', gear: compareGearSet(9), bonusStats: new Stats() },
	],
	talents: [...PARTS.talents, { ref: talentRef('Gone', false), name: 'Gone', talentsString: '1' }],
	rotations: [...PARTS.rotations, { ref: rotationRef('Gone', false), name: 'Gone', rotation: rotation('gone') }],
	settings: [
		...PARTS.settings,
		{ ref: settingsRef('Gone', false), name: 'Gone', settings: SavedSettings.create({ healingModel: HealingModel.create({ hps: 1 }) }) },
	],
};

describe('pruneBulkBundles', () => {
	it.each<[string, BundleRefs]>([
		['gear set', { gearSet: gearRef('Gone') }],
		['talents', { talents: talentRef('Gone', false) }],
		['rotation', { rotation: rotationRef('Gone', false) }],
		['settings', { settings: settingsRef('Gone', false) }],
	])(
		'skips a bundle whose saved %s is missing, and pruning removes it from storage so an entry of that name coming back does not restore it',
		(_, missing) => {
			const player = makeBulkPlayer();
			addBulkBundle(player, FURY);
			addBulkBundle(player, bundle({ ...FURY_REFS, ...missing }));
			addBulkBundle(player, ARMS);
			const stored = bulkState(player).bundles;

			expect(bulkBundleList(CURRENT, PARTS, stored, true).map(row => row.index)).toEqual([0, 2]);
			expect(bulkBundleRuns(CURRENT, PARTS, stored, true).map(run => run.index)).toEqual([-1, 0, 2]);

			pruneBulkBundles(player, PARTS);

			expect(bulkState(player).bundles).toEqual([FURY, ARMS]);
			expect(loadStoredBulkBundles(player)!.bundles).toEqual([FURY, ARMS]);
			expect(bulkBundleList(CURRENT, PARTS_WITH_GONE, bulkState(player).bundles, true).map(row => row.label)).toEqual([
				FURY_LABEL,
				'Pre-BiS (Arms) · Arms · Arms APL · Pre-BiS - Arms',
			]);

			const setItem = vi.spyOn(player.sim.env.storage, 'setItem');
			const pruned = bulkState(player).bundles;
			pruneBulkBundles(player, PARTS);

			expect(setItem).not.toHaveBeenCalled();
			expect(bulkState(player).bundles).toBe(pruned);
		},
	);

	it('keeps a bundle whose preset part is not pickable now stored, with no row and no run, until the preset is pickable again', () => {
		const player = makeBulkPlayer();
		const prot = bundle({ ...FURY_REFS, gearSet: gearRef('Pre-BiS', true, 'Prot') });
		addBulkBundle(player, FURY);
		addBulkBundle(player, prot);
		const stored = bulkState(player).bundles;

		pruneBulkBundles(player, PARTS);

		expect(bulkState(player).bundles).toBe(stored);
		expect(bulkBundleList(CURRENT, PARTS, stored, true).map(row => row.index)).toEqual([0]);
		expect(bulkBundleRuns(CURRENT, PARTS, stored, true).map(run => run.index)).toEqual([-1, 0]);

		expect(bulkBundleList(CURRENT, PARTS_WITH_GONE, stored, true).map(row => [row.index, row.label])).toEqual([
			[0, FURY_LABEL],
			[1, 'Pre-BiS (Prot) · Fury · Fury APL · Pre-BiS - Fury'],
		]);
		expect(bulkBundleRuns(CURRENT, PARTS_WITH_GONE, stored, true).map(run => run.index)).toEqual([-1, 0, 1]);
	});
});
