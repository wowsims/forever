import { APLAction, APLListItem, APLRotation, APLRotation_Type as APLRotationType, APLValue, SimpleRotation } from '@generated/proto/apl';
import { Encounter, ItemSlot, ItemType, Profession, Stat, Target, UUID } from '@generated/proto/common';
import { SavedSettings, type UIEnchant } from '@generated/proto/ui';
import { DpsWarrior } from '@generated/proto/warrior';
import type { BulkEncounterSetup, BulkGearSetup, BulkPlayerSettingsSetup, BulkRotationSetup, BulkTalentSetup } from '@sim/bulk/types';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { bulkSimHighStageFloor } from '@sim/wasm/bulk_sim/estimate';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { describe, expect, it, vi } from 'vitest';

import {
	bulkBuildPlan,
	type BulkBuildReference,
	type BulkBuildRequest,
	bulkBuildsCount,
	bulkBuildStep,
	type BulkBundlePlan,
	bulkBundlePlans,
	bulkGearSetSetups,
	bulkPlanRuns,
	bulkRunningPlans,
	type ResolveBuildRotation,
} from './builds';
import type { BulkBundleRun } from './bundles';
import type { BulkGearSetPick } from './gear_sets';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string, options?: { count?: number }) => (options?.count === undefined ? key : `${key}:${options.count}`) },
}));

const RING_STAMINA = makeUIEnchant({ effectId: 7301, type: ItemType.ItemTypeFinger, requiredProfession: Profession.Enchanting });
const RING_AGILITY = makeUIEnchant({ effectId: 7302, type: ItemType.ItemTypeFinger, requiredProfession: Profession.Enchanting });

const gearOf = (mainHand: number, offHand?: number, ringEnchant?: UIEnchant) =>
	new Gear({
		[ItemSlot.ItemSlotMainHand]: makeEquippedItem({ item: { id: mainHand } }),
		...(offHand ? { [ItemSlot.ItemSlotOffHand]: makeEquippedItem({ item: { id: offHand } }) } : {}),
		...(ringEnchant
			? { [ItemSlot.ItemSlotFinger1]: makeEquippedItem({ item: { id: 50, type: ItemType.ItemTypeFinger, unique: false }, enchant: ringEnchant }) }
			: {}),
	});

const CURRENT_GEAR = gearOf(1, 2);
const TWO_HAND_GEAR = gearOf(3);
const OTHER_DUAL_WIELD_GEAR = gearOf(4, 2);

const NO_BONUS = new Stats();
const STRENGTH_BONUS = new Stats().withStat(Stat.StatStrength, 10);
const LIVE_BONUS = new Stats().withStat(Stat.StatAgility, 5);

const setupOf = (gearSet: string, gear: Gear, bonusStats = NO_BONUS): BulkGearSetup => ({ gearSet, gear, bonusStats });

const ALL = 0;

const referenceOf = (overrides: Partial<BulkBuildReference> = {}): BulkBuildReference => ({
	gear: CURRENT_GEAR,
	bonusStats: NO_BONUS,
	isEnchanter: true,
	settings: null,
	...overrides,
});

const rotationOf = (type: APLRotationType, notes = '', uuid = '') =>
	APLRotation.create({
		type,
		priorityList: notes
			? [
					APLListItem.create({
						notes,
						action: uuid ? APLAction.create({ condition: APLValue.create({ uuid: UUID.create({ value: uuid }) }) }) : undefined,
					}),
				]
			: [],
	});

const CURRENT_TALENTS: BulkTalentSetup = { name: '', talentsString: '111', isCurrent: true };
const ARMS_TALENTS: BulkTalentSetup = { name: 'Arms', talentsString: '222', isCurrent: false };
const FURY_TALENTS: BulkTalentSetup = { name: 'Fury', talentsString: '333', isCurrent: false };

const CURRENT_ROTATION: BulkRotationSetup = { name: '', isCurrent: true, rotation: rotationOf(APLRotationType.TypeAuto) };
const AUTO_ROTATION: BulkRotationSetup = { name: 'Auto', isCurrent: false, rotation: rotationOf(APLRotationType.TypeAuto) };
const CUSTOM_ROTATION: BulkRotationSetup = { name: 'Custom', isCurrent: false, rotation: rotationOf(APLRotationType.TypeAPL, 'custom', 'original') };

const simpleRotation = (name: string, specRotationJson: string): BulkRotationSetup => ({
	name,
	isCurrent: false,
	rotation: APLRotation.create({ type: APLRotationType.TypeSimple, simple: SimpleRotation.create({ specRotationJson }) }),
});

const CURRENT_SETTINGS: BulkPlayerSettingsSetup = { name: '', isCurrent: true, settings: null };
const TANK_SETTINGS: BulkPlayerSettingsSetup = { name: 'Tank', isCurrent: false, settings: SavedSettings.create({ professions: [Profession.Enchanting] }) };

const gearSetOf = (name: string, gear: Gear, bonusStats = NO_BONUS): BulkGearSetPick => ({ name, gear, bonusStats });

interface BundleParts {
	gearSet: BulkGearSetPick;
	talents: BulkTalentSetup;
	rotation: BulkRotationSetup;
	settings: BulkPlayerSettingsSetup;
}

// As `bulkBundleRuns` gives them; the build key stands in for the picks by their names, which no two picks here share.
const bundleOf = (
	index: number,
	{ gearSet = gearSetOf('', CURRENT_GEAR), talents = CURRENT_TALENTS, rotation = CURRENT_ROTATION, settings = CURRENT_SETTINGS }: Partial<BundleParts> = {},
): BulkBundleRun => ({
	index,
	gearSet,
	talents,
	rotation,
	settings,
	buildKey: JSON.stringify([talents.name, rotation.name, settings.name]),
});

const planOf = (bundle: BulkBundleRun, setups: BulkGearSetup[]): BulkBundlePlan => ({ bundle, setups });

const currentPlan = (setups: BulkGearSetup[]) => planOf(bundleOf(-1), setups);

// As the player's resolver: a custom APL goes as it is; an Auto or Simple pick generates a new list each time, here by
// the hands the gear fills alone, and the result keeps the pick's type and Simple settings.
const resolve: ResolveBuildRotation = ({ gear, rotation }) => {
	if (rotation.type === APLRotationType.TypeAPL) return rotation;
	const generated = rotationOf(APLRotationType.TypeUnknown, gear.getEquippedItem(ItemSlot.ItemSlotOffHand) ? 'dual wield' : 'two-hand');
	return { ...generated, type: rotation.type, simple: rotation.simple };
};

const DUAL_WIELD_RESOLVED = rotationOf(APLRotationType.TypeAuto, 'dual wield');
const TWO_HAND_RESOLVED = rotationOf(APLRotationType.TypeAuto, 'two-hand');

const encounterOf = (name: string, duration: number): BulkEncounterSetup => ({
	name,
	isCurrent: !name,
	encounter: Encounter.create({ duration, durationVariation: 5, targets: [Target.create()] }),
});

const CURRENT_ENCOUNTER = encounterOf('', 180);
const RAID_ENCOUNTER = encounterOf('Raid boss', 300);

// The setups a request sims, its baseline first.
const sentSetups = ({ baselineSetup, candidates }: BulkBuildRequest) => [...(baselineSetup ? [baselineSetup] : []), ...candidates];

const bulkBuildRuns = (
	plans: readonly BulkBundlePlan[],
	reference: BulkBuildReference,
	encounters: readonly BulkEncounterSetup[],
	resolver: ResolveBuildRotation,
) => bulkPlanRuns(bulkBuildPlan(plans, reference, resolver), encounters);

describe('bulkBuildPlan', () => {
	it('plans nothing without a bundle to run or an encounter', () => {
		expect(bulkBuildRuns([], referenceOf(), [CURRENT_ENCOUNTER], resolve)).toEqual([]);
		expect(bulkBuildRuns([currentPlan([setupOf('', CURRENT_GEAR)])], referenceOf(), [], resolve)).toEqual([]);
	});

	it('runs each encounter with each bundle group, encounter outer, the Current bundle’s group first', () => {
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION }), [
				setupOf('Raid set', OTHER_DUAL_WIELD_GEAR),
			]),
		];

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], resolve);

		expect(runs.map(({ encounter, bundle, group }) => [encounter.name, bundle.index, group])).toEqual([
			['', -1, 0],
			['', 0, 1],
			['Raid boss', -1, 0],
			['Raid boss', 0, 1],
		]);
		expect(runs.map(run => run.requests[0].override.talentsString)).toEqual(['111', '222', '111', '222']);
	});

	it('leads with the reference request: the live gear as its baseline, the live bonus stats and the current rotation resolved for it', () => {
		const currentSetups = [setupOf('', CURRENT_GEAR), setupOf('', gearOf(1, 2, RING_AGILITY))];

		const runs = bulkBuildRuns([currentPlan(currentSetups)], referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(runs[0].requests[0]).toMatchObject({ baselineGear: null, baselineSetup: currentSetups[0] });
		expect(runs[0].requests).toEqual([
			{
				candidates: [currentSetups[1]],
				baselineSetup: currentSetups[0],
				baselineGear: null,
				override: { talentsString: '111', rotation: DUAL_WIELD_RESOLVED, bonusStats: NO_BONUS.toProto() },
			},
		]);
	});

	it('sims a bundle’s first gear setup as its request’s baseline and the rest as candidates', () => {
		const armsSetups = [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR), setupOf('Raid set', gearOf(4, 2, RING_AGILITY))];
		const arms = bundleOf(0, { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION });

		const [, armsRun] = bulkBuildRuns([currentPlan([setupOf('', CURRENT_GEAR)]), planOf(arms, armsSetups)], referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(armsRun.requests).toHaveLength(1);
		expect(armsRun.requests[0].baselineSetup).toBe(armsSetups[0]);
		expect(armsRun.requests[0].baselineGear).toBe(armsSetups[0].gear);
		expect(armsRun.requests[0].candidates).toEqual(armsSetups.slice(1));
		expect(armsRun.requests[0].override).toEqual({ talentsString: '222', rotation: CUSTOM_ROTATION.rotation, bonusStats: NO_BONUS.toProto() });
	});

	it('never sims a gear setup with another group’s talents, rotation or settings', () => {
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR), setupOf('', gearOf(1, 2, RING_AGILITY))]),
			planOf(bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR) }), [setupOf('Two-hand', TWO_HAND_GEAR)]),
			planOf(bundleOf(1, { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION }), [
				setupOf('Raid set', OTHER_DUAL_WIELD_GEAR),
				setupOf('Raid set', gearOf(4, 2, RING_STAMINA)),
			]),
			planOf(bundleOf(2, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR), talents: FURY_TALENTS, rotation: simpleRotation('Simple', '{}') }), [
				setupOf('Two-hand', TWO_HAND_GEAR),
			]),
			planOf(
				bundleOf(3, {
					gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR),
					talents: ARMS_TALENTS,
					rotation: CUSTOM_ROTATION,
					settings: TANK_SETTINGS,
				}),
				[setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)],
			),
		];
		const planHolding = (setup: BulkGearSetup) => plans.find(plan => plan.setups.includes(setup))!;

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], resolve);

		expect(runs).toHaveLength(8);
		for (const run of runs) {
			for (const request of run.requests) {
				expect(request.override.talentsString).toBe(run.bundle.talents.talentsString);
				for (const setup of sentSetups(request)) expect(planHolding(setup).bundle.buildKey).toBe(run.bundle.buildKey);
			}
		}
		expect(runs.map(run => run.bundle.settings.name)).toEqual(['', '', '', 'Tank', '', '', '', 'Tank']);
	});

	it('puts bundles with the same talents, rotation and Settings picks in one group', () => {
		const raidSetup = setupOf('Raid set', OTHER_DUAL_WIELD_GEAR);
		const twoHandSetup = setupOf('Two-hand', TWO_HAND_GEAR);
		const armsCustom = { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION };
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), ...armsCustom }), [raidSetup]),
			planOf(bundleOf(1, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR), talents: FURY_TALENTS, rotation: AUTO_ROTATION }), [
				setupOf('Two-hand', TWO_HAND_GEAR),
			]),
			planOf(bundleOf(2, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR), ...armsCustom }), [twoHandSetup]),
		];

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(runs.map(({ bundle, group }) => [bundle.index, group])).toEqual([
			[-1, 0],
			[0, 1],
			[1, 2],
		]);
		expect(runs[1].requests).toEqual([expect.objectContaining({ baselineSetup: raidSetup, candidates: [twoHandSetup] })]);
	});

	it('runs two bundles that differ only in their Settings entry as two groups', () => {
		const armsCustom = { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION };
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, armsCustom), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)]),
			planOf(bundleOf(1, { ...armsCustom, settings: TANK_SETTINGS }), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)]),
		];

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(bulkBuildPlan(plans, referenceOf(), resolve).sameAs).toEqual(new Map());
		expect(runs.map(({ bundle, group }) => [bundle.index, group, bundle.settings.name])).toEqual([
			[-1, 0, ''],
			[0, 1, ''],
			[1, 2, 'Tank'],
		]);
		expect(runs[1].requests).toEqual(runs[2].requests);
	});

	it('runs the first of two bundles whose Auto and Simple picks resolve alike and leaves the later twin out, naming the first', () => {
		const raid = { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS };
		const autoPlan = planOf(bundleOf(0, { ...raid, rotation: AUTO_ROTATION }), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)]);
		const simplePlan = planOf(bundleOf(1, { ...raid, rotation: simpleRotation('Simple', '{"a":1}') }), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)]);
		// The live gear and talents with a Simple pick, which resolves as the current Auto rotation does.
		const currentTwin = planOf(bundleOf(2, { rotation: simpleRotation('Simple live', '{}') }), [setupOf('', CURRENT_GEAR)]);
		const plans = [currentPlan([setupOf('', CURRENT_GEAR)]), autoPlan, simplePlan, currentTwin];

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], resolve);

		expect(bulkBuildPlan(plans, referenceOf(), resolve).sameAs).toEqual(
			new Map([
				[1, 0],
				[2, -1],
			]),
		);
		expect(runs.map(({ bundle, group }) => [bundle.index, group])).toEqual([
			[-1, 0],
			[0, 1],
			[-1, 0],
			[0, 1],
		]);
		const sent = runs.flatMap(run => run.requests.flatMap(sentSetups));
		expect([...simplePlan.setups, ...currentTwin.setups].some(setup => sent.includes(setup))).toBe(false);
		// Unticked, the Current bundle is not in the plans, so the bundle that resolves like it runs.
		expect(bulkBuildPlan([autoPlan, simplePlan, currentTwin], referenceOf(), resolve).sameAs).toEqual(new Map([[1, 0]]));
	});

	it('runs a bundle that sims like the unticked Current bundle, with its own setup as the baseline', () => {
		const liveSetup = setupOf('', CURRENT_GEAR);
		const currentTwin = planOf(bundleOf(0, { rotation: simpleRotation('Simple live', '{}') }), [liveSetup]);
		const plans = [currentTwin];
		const plan = bulkBuildPlan(plans, referenceOf(), resolve);

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], resolve);

		expect(plan.sameAs).toEqual(new Map());
		expect(runs.map(run => run.requests.map(({ baselineSetup, baselineGear }) => ({ baselineSetup, baselineGear })))).toEqual([
			[{ baselineSetup: liveSetup, baselineGear: CURRENT_GEAR }],
			[{ baselineSetup: liveSetup, baselineGear: CURRENT_GEAR }],
		]);
		expect(bulkBuildsCount(bulkRunningPlans(plans, plan), runs, 2, ALL)).toMatchObject({ bundles: 1, setups: 1, runs: 2, builds: 2 });
	});

	it('leaves out a bundle whose settings entry holds the live settings and that otherwise sims like the Current bundle, naming it', () => {
		const healer: BulkPlayerSettingsSetup = { name: 'Healer', isCurrent: false, settings: SavedSettings.create({ professions: [Profession.Alchemy] }) };
		const simpleLive = simpleRotation('Simple live', '{}');
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, { rotation: simpleLive, settings: TANK_SETTINGS }), [setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(1, { rotation: simpleLive, settings: healer }), [setupOf('', CURRENT_GEAR)]),
		];
		const reference = referenceOf({ settings: TANK_SETTINGS.settings });

		const runs = bulkBuildRuns(plans, reference, [CURRENT_ENCOUNTER], resolve);

		expect(bulkBuildPlan(plans, reference, resolve).sameAs).toEqual(new Map([[0, -1]]));
		expect(runs.map(run => run.bundle.index)).toEqual([-1, 1]);
		// Until the live settings are known, the current settings match no entry.
		expect(bulkBuildPlan(plans, referenceOf(), resolve).sameAs).toEqual(new Map());
	});

	it('names the Current bundle for a settings entry without spec options, which sims the live ones, and keeps one with other spec options', () => {
		const rageSpec = (startingRage: number): SavedSettings['spec'] => ({
			oneofKind: 'dpsWarrior',
			dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage } } }),
		});
		const entry = (name: string, spec?: SavedSettings['spec']): BulkPlayerSettingsSetup => ({
			name,
			isCurrent: false,
			settings: SavedSettings.create({ professions: [Profession.Enchanting], spec }),
		});
		const simpleLive = simpleRotation('Simple live', '{}');
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, { rotation: simpleLive, settings: entry('No spec options') }), [setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(1, { rotation: simpleLive, settings: entry('Other rage', rageSpec(0)) }), [setupOf('', CURRENT_GEAR)]),
		];
		const reference = referenceOf({ settings: SavedSettings.create({ professions: [Profession.Enchanting], spec: rageSpec(50) }) });

		expect(bulkBuildPlan(plans, reference, resolve).sameAs).toEqual(new Map([[0, -1]]));
	});

	it('keeps two bundles whose picks resolve to different lists', () => {
		const raid = { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS };
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, { ...raid, rotation: AUTO_ROTATION }), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)]),
			planOf(bundleOf(1, { ...raid, rotation: CUSTOM_ROTATION }), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)]),
		];

		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(bulkBuildPlan(plans, referenceOf(), resolve).sameAs).toEqual(new Map());
		expect(runs.map(run => [run.bundle.index, run.requests[0].override.rotation])).toEqual([
			[-1, DUAL_WIELD_RESOLVED],
			[0, DUAL_WIELD_RESOLVED],
			[1, CUSTOM_ROTATION.rotation],
		]);
	});

	it('leaves out a bundle whose every gear setup an earlier bundle of its group sends, naming that bundle', () => {
		const staminaGear = gearOf(4, 2, RING_STAMINA);
		const agilityGear = gearOf(4, 2, RING_AGILITY);
		const arms = { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION };
		const plans = [
			currentPlan([setupOf('', CURRENT_GEAR)]),
			planOf(bundleOf(0, { ...arms, gearSet: gearSetOf('Stamina ring', staminaGear) }), [setupOf('Stamina ring', staminaGear)]),
			planOf(bundleOf(1, { ...arms, gearSet: gearSetOf('Agility ring', agilityGear) }), [setupOf('Agility ring', agilityGear)]),
		];

		// Without Enchanting both are sent without their ring enchant, so the Agility ring bundle adds nothing to the group.
		const plan = bulkBuildPlan(plans, referenceOf({ isEnchanter: false }), resolve);

		expect(plan.sameAs).toEqual(new Map([[1, 0]]));
		expect(plan.groups.map(({ bundle, requests }) => [bundle.index, requests.flatMap(sentSetups).map(setup => setup.gearSet)])).toEqual([
			[-1, ['']],
			[0, ['Stamina ring']],
		]);
		expect(bulkBuildPlan(plans, referenceOf(), resolve).sameAs).toEqual(new Map());
	});

	it('names the running bundle for a twin of a bundle that does not run', () => {
		const staminaGear = gearOf(4, 2, RING_STAMINA);
		const agilityGear = gearOf(4, 2, RING_AGILITY);
		const arms = { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION };
		const plans = [
			planOf(bundleOf(0, { ...arms, gearSet: gearSetOf('Stamina ring', staminaGear) }), [setupOf('Stamina ring', staminaGear)]),
			planOf(bundleOf(1, { ...arms, gearSet: gearSetOf('Agility ring', agilityGear) }), [setupOf('Agility ring', agilityGear)]),
			planOf(bundleOf(2, { ...arms, gearSet: gearSetOf('Agility copy', agilityGear) }), [setupOf('Agility copy', agilityGear)]),
		];

		// The copy sims like the Agility ring bundle, which its group's Stamina ring bundle already sends without Enchanting.
		expect(bulkBuildPlan(plans, referenceOf({ isEnchanter: false }), resolve).sameAs).toEqual(
			new Map([
				[1, 0],
				[2, 0],
			]),
		);
	});

	it('runs a bundle wearing the live gear with other talents, its baseline the live gear given as its own setup', () => {
		const liveSetup = setupOf('', CURRENT_GEAR);
		const plans = [currentPlan([setupOf('', CURRENT_GEAR)]), planOf(bundleOf(0, { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION }), [liveSetup])];

		const [, arms] = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(arms.requests).toEqual([
			{
				candidates: [],
				baselineSetup: liveSetup,
				baselineGear: CURRENT_GEAR,
				override: { talentsString: '222', rotation: CUSTOM_ROTATION.rotation, bonusStats: NO_BONUS.toProto() },
			},
		]);
	});

	it('sims no live gear with the Current bundle unticked: every request, the first group’s too, sims its own first setup as its baseline', () => {
		// The current talents, rotation and settings with other gear, which a ticked Current bundle's group would hold.
		const twoHandSetups = [setupOf('Two-hand', TWO_HAND_GEAR), setupOf('Two-hand', gearOf(3, undefined, RING_AGILITY))];
		const twoHand = bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR) });
		const armsSetups = [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR), setupOf('Two-hand', TWO_HAND_GEAR)];
		const arms = bundleOf(1, { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION });

		const runs = bulkBuildRuns(
			[planOf(twoHand, twoHandSetups), planOf(arms, armsSetups)],
			referenceOf({ bonusStats: LIVE_BONUS }),
			[CURRENT_ENCOUNTER],
			resolve,
		);

		expect(runs.map(({ bundle, group }) => [bundle.index, group])).toEqual([
			[0, 0],
			[1, 1],
		]);
		expect(runs.map(run => run.requests)).toEqual([
			[
				{
					candidates: [twoHandSetups[1]],
					baselineSetup: twoHandSetups[0],
					baselineGear: TWO_HAND_GEAR,
					override: { talentsString: '111', rotation: TWO_HAND_RESOLVED, bonusStats: NO_BONUS.toProto() },
				},
			],
			[
				{
					candidates: [armsSetups[1]],
					baselineSetup: armsSetups[0],
					baselineGear: armsSetups[0].gear,
					override: { talentsString: '222', rotation: CUSTOM_ROTATION.rotation, bonusStats: NO_BONUS.toProto() },
				},
			],
		]);
	});

	it('splits a group by bonus stats and by the rotation each gear set resolves to', () => {
		const live = setupOf('', CURRENT_GEAR);
		const twoHand = setupOf('Two-hand', TWO_HAND_GEAR);
		const buffed = setupOf('Buffed', gearOf(5, 6), STRENGTH_BONUS);
		const armsTwoHand = setupOf('Two-hand', TWO_HAND_GEAR);
		const armsBuffed = setupOf('Buffed', gearOf(5, 6), STRENGTH_BONUS);
		const armsCustom = { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION };
		const plans = [
			currentPlan([live]),
			planOf(bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR) }), [twoHand]),
			planOf(bundleOf(1, { gearSet: gearSetOf('Buffed', gearOf(5, 6), STRENGTH_BONUS) }), [buffed]),
			planOf(bundleOf(2, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR), ...armsCustom }), [armsTwoHand]),
			planOf(bundleOf(3, { gearSet: gearSetOf('Buffed', gearOf(5, 6), STRENGTH_BONUS), ...armsCustom }), [armsBuffed]),
		];

		const [current, arms] = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		const shape = ({ baselineSetup, baselineGear, candidates, override }: BulkBuildRequest) => ({
			baselineSetup,
			baselineGear,
			candidates,
			rotation: override.rotation,
			bonusStats: override.bonusStats,
		});
		expect(current.requests.map(shape)).toEqual([
			{ baselineSetup: live, baselineGear: null, candidates: [], rotation: DUAL_WIELD_RESOLVED, bonusStats: NO_BONUS.toProto() },
			{ baselineSetup: twoHand, baselineGear: TWO_HAND_GEAR, candidates: [], rotation: TWO_HAND_RESOLVED, bonusStats: NO_BONUS.toProto() },
			{ baselineSetup: buffed, baselineGear: buffed.gear, candidates: [], rotation: DUAL_WIELD_RESOLVED, bonusStats: STRENGTH_BONUS.toProto() },
		]);
		expect(arms.requests.map(shape)).toEqual([
			{ baselineSetup: armsTwoHand, baselineGear: TWO_HAND_GEAR, candidates: [], rotation: CUSTOM_ROTATION.rotation, bonusStats: NO_BONUS.toProto() },
			{
				baselineSetup: armsBuffed,
				baselineGear: armsBuffed.gear,
				candidates: [],
				rotation: CUSTOM_ROTATION.rotation,
				bonusStats: STRENGTH_BONUS.toProto(),
			},
		]);
	});

	it('sends a group whose settings lack Enchanting its setups without ring enchants', () => {
		const current = gearOf(1, 2, RING_STAMINA);
		const setups = [setupOf('', current), setupOf('', gearOf(1, 2, RING_AGILITY))];
		const enchanter: BulkPlayerSettingsSetup = {
			name: 'Enchanter',
			isCurrent: false,
			settings: SavedSettings.create({ professions: [Profession.Mining, Profession.Enchanting] }),
		};
		const stripped = { ...setups[0], gear: setups[0].gear.withoutEnchanting(), ownGear: setups[0].gear };

		const [live, entry] = bulkBuildRuns(
			[currentPlan(setups), planOf(bundleOf(0, { settings: enchanter }), setups)],
			referenceOf({ gear: current, isEnchanter: false }),
			[CURRENT_ENCOUNTER],
			resolve,
		);

		expect(live.requests.map(({ candidates, baselineSetup }) => ({ candidates, baselineSetup }))).toEqual([{ candidates: [], baselineSetup: stripped }]);
		expect(stripped.gear.getEquippedItem(ItemSlot.ItemSlotFinger1)?.enchant).toBeNull();
		expect(entry.requests.map(({ candidates, baselineSetup, baselineGear }) => ({ candidates, baselineSetup, baselineGear }))).toEqual([
			{ candidates: [setups[1]], baselineSetup: setups[0], baselineGear: current },
		]);
	});

	it('resolves a rotation once per talents, pick and items, whatever the enchants and the encounters', () => {
		const currentSetups = [setupOf('', gearOf(1, 2, RING_STAMINA)), setupOf('', gearOf(1, 2, RING_AGILITY))];
		const twoHand = setupOf('Two-hand', TWO_HAND_GEAR);
		const armsSetups = [setupOf('', gearOf(1, 2, RING_STAMINA)), setupOf('Two-hand', TWO_HAND_GEAR)];
		const plans = [
			currentPlan(currentSetups),
			planOf(bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR) }), [twoHand]),
			planOf(bundleOf(1, { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION }), armsSetups),
		];
		const counted = vi.fn<ResolveBuildRotation>(resolve);

		const runs = bulkBuildRuns(plans, referenceOf({ gear: gearOf(1, 2, RING_STAMINA) }), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], counted);

		const calls = counted.mock.calls.map(([{ talentsString, rotation, gear }]) => [
			talentsString,
			rotation.type,
			gear.getEquippedItem(ItemSlot.ItemSlotMainHand)?.id,
		]);
		expect(calls).toEqual([
			['111', APLRotationType.TypeAuto, 1],
			['111', APLRotationType.TypeAuto, 3],
			['222', APLRotationType.TypeAPL, 1],
			['222', APLRotationType.TypeAPL, 3],
		]);
		expect(runs.map(run => run.requests.map(sentSetups))).toEqual([[currentSetups, [twoHand]], [armsSetups], [currentSetups, [twoHand]], [armsSetups]]);
	});

	it('resolves again only the builds the last plan with the resolver did not resolve', () => {
		const raid = { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS };
		const planWith = (rotation: BulkRotationSetup) => [planOf(bundleOf(0, { ...raid, rotation }), [setupOf('Raid set', OTHER_DUAL_WIELD_GEAR)])];
		const counted = vi.fn<ResolveBuildRotation>(resolve);

		bulkBuildPlan(planWith(AUTO_ROTATION), referenceOf(), counted);
		bulkBuildPlan(planWith(AUTO_ROTATION), referenceOf(), counted);
		bulkBuildPlan(planWith(CUSTOM_ROTATION), referenceOf(), counted);
		bulkBuildPlan(planWith(AUTO_ROTATION), referenceOf(), counted);

		const resolved = counted.mock.calls.map(([{ rotation }]) => rotation.type);
		expect(resolved).toEqual([APLRotationType.TypeAuto, APLRotationType.TypeAPL, APLRotationType.TypeAuto]);
	});

	it('builds a gear set’s setups once however many bundles and plans hold it', () => {
		const getEnchants = vi.fn(() => []);
		const player = makeBulkPlayer({ sim: { db: { getEnchants } } });
		const preBiS = () => gearSetOf('Pre-BiS', OTHER_DUAL_WIELD_GEAR);
		const bundles = [
			bundleOf(0, { gearSet: preBiS(), talents: ARMS_TALENTS }),
			bundleOf(1, { gearSet: preBiS(), talents: FURY_TALENTS }),
			bundleOf(2, { gearSet: gearSetOf('My Pre-BiS copy', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, settings: TANK_SETTINGS }),
		];

		const setupsOf = bulkGearSetSetups(player, []);

		const plans = bulkBundlePlans(bundles, setupsOf);
		const built = getEnchants.mock.calls.length;
		const replanned = bulkBundlePlans([bundles[1], bundleOf(3, { gearSet: preBiS(), settings: TANK_SETTINGS })], setupsOf);

		expect(plans.map(plan => plan.bundle)).toEqual(bundles);
		expect(plans[0].setups).toBe(plans[1].setups);
		expect(plans[2].setups).not.toBe(plans[0].setups);
		expect(plans[2].setups[0].gearSet).toBe('My Pre-BiS copy');
		expect(plans[0].setups.map(setup => setup.gearSet)).toEqual(['Pre-BiS']);
		// A later plan with the same gear set builds nothing again.
		expect(replanned.map(plan => plan.setups)).toEqual([plans[0].setups, plans[0].setups]);
		expect(replanned[1].setups).toBe(plans[0].setups);
		expect(getEnchants.mock.calls.length).toBe(built);
	});

	it('builds a gear set’s setups again once a plan without it has dropped them', () => {
		const getEnchants = vi.fn(() => []);
		const player = makeBulkPlayer({ sim: { db: { getEnchants } } });
		const preBiS = [bundleOf(0, { gearSet: gearSetOf('Pre-BiS', OTHER_DUAL_WIELD_GEAR) })];
		const setupsOf = bulkGearSetSetups(player, []);

		const [first] = bulkBundlePlans(preBiS, setupsOf);
		bulkBundlePlans([bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR) })], setupsOf);
		const built = getEnchants.mock.calls.length;
		const [again] = bulkBundlePlans(preBiS, setupsOf);

		expect(getEnchants.mock.calls.length).toBeGreaterThan(built);
		expect(again.setups).not.toBe(first.setups);
		expect(again.setups).toEqual(first.setups);
	});
});

describe('bulkBuildsCount', () => {
	it('counts the running bundles, the encounters, their gear setups and every request as a run', () => {
		const plans = [
			planOf(bundleOf(0, { gearSet: gearSetOf('Raid set', OTHER_DUAL_WIELD_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION }), [
				setupOf('Raid set', OTHER_DUAL_WIELD_GEAR),
				setupOf('Raid set', gearOf(4, 2, RING_STAMINA)),
			]),
			planOf(bundleOf(1, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR), talents: FURY_TALENTS, rotation: simpleRotation('Simple', '{}') }), [
				setupOf('Two-hand', TWO_HAND_GEAR),
			]),
		];
		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], resolve);

		expect(runs.map(run => run.requests.length)).toEqual([1, 1, 1, 1]);
		expect(bulkBuildsCount(plans, runs, 2, ALL)).toMatchObject({ bundles: 2, encounters: 2, setups: 3, runs: 4, builds: 6 });
	});

	it('counts one build for a lone bundle of one setup with the Current bundle unticked, and one per encounter over more', () => {
		const plans = [planOf(bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR) }), [setupOf('Two-hand', TWO_HAND_GEAR)])];

		expect(bulkBuildsCount(plans, bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve), 1, ALL).builds).toBe(1);
		expect(bulkBuildsCount(plans, bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER, RAID_ENCOUNTER], resolve), 2, ALL).builds).toBe(2);
	});

	it('counts the minimum iterations of every request', () => {
		const plans = [
			currentPlan([
				setupOf('', CURRENT_GEAR),
				setupOf('Raid set', OTHER_DUAL_WIELD_GEAR),
				setupOf('Dungeon', gearOf(5, 6)),
				setupOf('Heroic', gearOf(7, 8)),
			]),
			planOf(bundleOf(0, { gearSet: gearSetOf('Two-hand', TWO_HAND_GEAR), talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION }), [
				setupOf('Two-hand', TWO_HAND_GEAR),
			]),
		];
		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(bulkBuildsCount(plans, runs, 1, ALL)).toEqual({
			bundles: 2,
			encounters: 1,
			setups: 5,
			runs: 2,
			builds: 5,
			totalIterations: 5 * bulkSimHighStageFloor(),
			largestRequestIterations: 4 * bulkSimHighStageFloor(),
			largestRequestCandidates: 3,
		});
	});

	it('counts a request of 28 candidates at the high stage alone with All, and through the culling stage with a top 5', () => {
		const sets = Array.from({ length: 28 }, (_, index) => setupOf(`Set ${index}`, gearOf(100 + index, 200 + index)));
		const plans = [currentPlan([setupOf('', CURRENT_GEAR), ...sets])];
		const runs = bulkBuildRuns(plans, referenceOf(), [CURRENT_ENCOUNTER], resolve);

		expect(bulkBuildsCount(plans, runs, 1, ALL)).toMatchObject({
			largestRequestCandidates: 28,
			largestRequestIterations: 29 * bulkSimHighStageFloor(),
		});
		// The medium stage at its 1,000 floor for the 28 and the baseline, then the high stage for its 25 survivors and the baseline.
		expect(bulkBuildsCount(plans, runs, 1, 5).largestRequestIterations).toBe(29 * 1000 + 26 * bulkSimHighStageFloor());
	});
});

describe('bulkBuildStep', () => {
	it('names a step by its encounter with its length and targets, and its group’s settings, talents and rotation', () => {
		const arms = bundleOf(0, { talents: ARMS_TALENTS, rotation: CUSTOM_ROTATION, settings: TANK_SETTINGS });
		expect(bulkBuildStep({ encounter: RAID_ENCOUNTER, bundle: arms })).toEqual({
			encounter: 'Raid boss',
			encounterSummary: '5:00 ±5s · bulk_tab.settings.encounters.targets:1',
			settings: 'Tank',
			talents: 'Arms',
			rotation: 'Custom',
		});
		expect(bulkBuildStep({ encounter: CURRENT_ENCOUNTER, bundle: bundleOf(-1) })).toEqual({
			encounter: 'bulk_tab.settings.encounters.current',
			encounterSummary: '3:00 ±5s · bulk_tab.settings.encounters.targets:1',
			settings: 'bulk_tab.results.current_settings',
			talents: 'bulk_tab.results.current_talents',
			rotation: 'bulk_tab.results.current_rotation',
		});
	});
});
