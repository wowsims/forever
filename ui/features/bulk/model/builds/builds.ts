import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Profession } from '@generated/proto/common';
import type { SavedSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkBuildOverride, BulkEncounterSetup, BulkGearSetup, BulkPlayerSettingsSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { withLiveSpecOptions } from '@sim/proto/spec_functions';
import type { Stats } from '@sim/proto/stats';
import { omitDeep } from '@sim/utils/collections';
import { estimateBulkSimIterations } from '@sim/wasm/bulk_sim/estimate';

import type { BulkBundleRun } from './bundles';
import type { BulkEnchantPick } from './enchant_picks';
import { formatEncounterSummary } from './encounters';
import type { BulkGearSetPick } from './gear_sets';
import { bonusStatsKey, bulkGearSetups, gearSetupsSentWith, sentGearSetupKey } from './gear_setups';
import { rotationLabel } from './rotations';
import { keyedOnce } from './saved_refs';
import { settingsContentKey, settingsLabel } from './settings_entries';
import { talentsLabel } from './talents';

export type ResolveBuildRotation = (build: { talentsString: string; gear: Gear; rotation: APLRotation }) => APLRotation;

export interface BulkBuildReference {
	/** The current gear as equipped. */
	gear: Gear;
	bonusStats: Stats;
	/** By the live professions. */
	isEnchanter: boolean;
	/** The live settings, which the current settings sim; null until the sim is ready. */
	settings: SavedSettings | null;
}

export interface BulkBundlePlan {
	bundle: BulkBundleRun;
	setups: readonly BulkGearSetup[];
}

export interface BulkBuildRequest {
	candidates: readonly BulkGearSetup[];
	/** The setup the baseline stands for: on the reference request the one that wears the live gear, if any. */
	baselineSetup: BulkGearSetup | null;
	/** Null on the reference request, which keeps the live gear; only a ticked Current bundle has one. */
	baselineGear: Gear | null;
	override: BulkBuildOverride;
}

interface BulkBuildGroup {
	/** The group's first bundle, whose talents, rotation and settings the group shares. */
	bundle: BulkBundleRun;
	/** The group's index in the batch, the same under every encounter. */
	group: number;
	requests: readonly BulkBuildRequest[];
}

interface BulkBuildPlan {
	/** Each bundle that does not run, by stored index, with the earlier running bundle it sims the same as; -1 is the Current bundle. */
	sameAs: ReadonlyMap<number, number>;
	groups: readonly BulkBuildGroup[];
}

interface BulkBuildRun extends BulkBuildGroup {
	encounter: BulkEncounterSetup;
}

export interface BulkBuildsCount {
	/** The running bundles. */
	bundles: number;
	encounters: number;
	/** The running bundles' gear setups. */
	setups: number;
	/** The requests of every encounter, after splits. */
	runs: number;
	/** The builds the requests sim: each one's baseline and candidates. */
	builds: number;
	totalIterations: number;
	largestRequestIterations: number;
	largestRequestCandidates: number;
}

interface BuildSplit {
	setups: BulkGearSetup[];
	bonusStats: Stats;
	resolved: APLRotation;
}

interface ResolvedRotation {
	rotation: APLRotation;
	simmedKey: string;
}

interface ResolvedBuild {
	rotation: APLRotation;
	simmedJson: string;
}

// The keys of one plan, each worked out once: a plan reads them per setup, bundle and group.
interface PlanKeys {
	resolve: (talentsString: string, gear: Gear, rotation: APLRotation) => ResolvedRotation;
	identityOf: (setup: BulkGearSetup) => string;
	bonusKeyOf: (bonusStats: Stats) => string;
}

const gearIdentity = (gear: Gear): string => getGearIdentityKey(gear.asSpec());

// What the sim reads of a resolved rotation: the type and Simple settings a resolved rotation keeps from its pick and
// the editor's uuids are left out, so an Auto and a Simple pick that generate the same list compare equal.
const simmedRotationKey = (rotation: APLRotation): string =>
	APLRotation.toJsonString({ ...omitDeep(rotation, ['uuid']), type: APLRotationType.TypeUnknown, simple: undefined });

// The rotation generators read the talents and the main hand's hand type, never an enchant, so gear that holds the same
// items in the same slots resolves alike whatever its enchants. A generator that reads an enchant needs the enchants in
// this key.
const generatorGearKey = keyedOnce((gear: Gear): string =>
	gear
		.asArray()
		.map(item => (item ? `${item.id}:${item.randomSuffix?.id ?? 0}` : ''))
		.join(','),
);

let pickCount = 0;
const pickId = keyedOnce<APLRotation>(() => String(pickCount++));

const setupIdentity = keyedOnce((setup: BulkGearSetup) => gearIdentity(setup.gear));

const setupBonusKey = keyedOnce(bonusStatsKey);

// The builds the last plan with each resolver resolved, by pick, talents and items. Resolving and keying a generated
// rotation take milliseconds each, which every enchant variant of a gear set and every later plan would pay again. A
// plan takes over the builds it resolves again and keeps no other, so a live edit leaves no rotation behind.
const lastResolved = new WeakMap<ResolveBuildRotation, Map<string, ResolvedBuild>>();

const planKeysOf = (resolve: ResolveBuildRotation): PlanKeys => {
	const previous = lastResolved.get(resolve);
	const resolved = new Map<string, ResolvedBuild>();
	lastResolved.set(resolve, resolved);
	// A short stand-in for each distinct key, which is a whole rotation as JSON and would be hashed again per setup.
	const simmedIds = new Map<string, string>();
	const simmedIdOf = (json: string): string => {
		if (!simmedIds.has(json)) simmedIds.set(json, String(simmedIds.size));
		return simmedIds.get(json)!;
	};
	return {
		resolve: (talentsString, gear, rotation) => {
			const key = `${pickId(rotation)}|${talentsString}|${generatorGearKey(gear)}`;
			let build = resolved.get(key) ?? previous?.get(key);
			if (!build) {
				const resolvedRotation = resolve({ talentsString, gear, rotation });
				build = { rotation: resolvedRotation, simmedJson: simmedRotationKey(resolvedRotation) };
			}
			resolved.set(key, build);
			return { rotation: build.rotation, simmedKey: simmedIdOf(build.simmedJson) };
		},
		identityOf: setupIdentity,
		bonusKeyOf: setupBonusKey,
	};
};

// A setup carries its gear set's name, so a preset and a saved copy with the same content keep setups of their own.
const gearSetKey = ({ name, gear, bonusStats }: BulkGearSetPick): string => `${name}|${gearIdentity(gear)}|${bonusStatsKey(bonusStats)}`;

type BulkGearSetSetups = (gearSets: readonly BulkGearSetPick[]) => Array<readonly BulkGearSetup[]>;

// Each of a plan's gear sets' setups under the enchant picks. A gear set the last plan held keeps the setups built for
// it, so another bundle of the plan or the next plan holding it builds none again; one the plan does not hold goes, so
// a live gear edit leaves no setups behind.
export const bulkGearSetSetups = (player: Player<any>, enchantPicks: readonly BulkEnchantPick[]): BulkGearSetSetups => {
	let last = new Map<string, readonly BulkGearSetup[]>();
	return gearSets => {
		const previous = last;
		const kept = new Map<string, readonly BulkGearSetup[]>();
		last = kept;
		return gearSets.map(gearSet => {
			const key = gearSetKey(gearSet);
			const setups = kept.get(key) ?? previous.get(key) ?? bulkGearSetups(player, [gearSet], enchantPicks);
			kept.set(key, setups);
			return setups;
		});
	};
};

// In bundle order, each running bundle with its gear set's setups.
export const bulkBundlePlans = (bundles: readonly BulkBundleRun[], setupsOf: BulkGearSetSetups): BulkBundlePlan[] => {
	const setups = setupsOf(bundles.map(bundle => bundle.gearSet));
	return bundles.map((bundle, index) => ({ bundle, setups: setups[index] }));
};

// By what the settings sim: the current settings by the live ones, so a settings entry that holds them is the same, and
// empty while the live settings are not known.
const simmedSettingsKey = ({ settings }: BulkPlayerSettingsSetup, live: SavedSettings | null): string => {
	const simmed = settings ? withLiveSpecOptions(settings, live) : live;
	return simmed ? settingsContentKey(simmed) : '';
};

// Equal for two bundles that sim the same: the same gear setups, talents and settings, and for each setup the same
// resolved rotation.
const simKey = (keys: PlanKeys, live: SavedSettings | null, { bundle, setups }: BulkBundlePlan): string => {
	const { talentsString } = bundle.talents;
	return JSON.stringify([
		talentsString,
		simmedSettingsKey(bundle.settings, live),
		setups.map(
			setup =>
				`${keys.identityOf(setup)}|${keys.bonusKeyOf(setup.bonusStats)}|${keys.resolve(talentsString, setup.gear, bundle.rotation.rotation).simmedKey}`,
		),
	]);
};

// Each bundle that sims like an earlier running one once the rotations are resolved, with that earlier one. A ticked
// Current bundle comes first and is the earlier one; unticked, it is not in the plans and keys nothing.
const twinsOf = (plans: readonly BulkBundlePlan[], keys: PlanKeys, live: SavedSettings | null): Map<number, number> => {
	const simmed = new Map<string, number>();
	const twins = new Map<number, number>();
	for (const plan of plans) {
		const key = simKey(keys, live, plan);
		const earlier = simmed.get(key);
		if (earlier === undefined) simmed.set(key, plan.bundle.index);
		else twins.set(plan.bundle.index, earlier);
	}
	return twins;
};

const splitKey = (keys: PlanKeys, bonusStats: Stats, resolved: ResolvedRotation): string => `${keys.bonusKeyOf(bonusStats)}|${resolved.simmedKey}`;

// A settings entry sends its own professions; the current settings send the live ones.
const sendsEnchanting = ({ settings }: BulkPlayerSettingsSetup, reference: BulkBuildReference): boolean =>
	settings ? settings.professions.includes(Profession.Enchanting) : reference.isEnchanter;

// One split per bonus stats and rotation the setups resolve to, in the order of each split's first setup.
const buildSplits = (setups: readonly BulkGearSetup[], { talents, rotation }: BulkBundleRun, keys: PlanKeys): Map<string, BuildSplit> => {
	const splits = new Map<string, BuildSplit>();
	for (const setup of setups) {
		const resolved = keys.resolve(talents.talentsString, setup.gear, rotation.rotation);
		const key = splitKey(keys, setup.bonusStats, resolved);
		const split = splits.get(key);
		if (split) split.setups.push(setup);
		else splits.set(key, { setups: [setup], bonusStats: setup.bonusStats, resolved: resolved.rotation });
	}
	return splits;
};

const buildOverride = ({ talents }: BulkBundleRun, split: BuildSplit): BulkBuildOverride => ({
	talentsString: talents.talentsString,
	rotation: split.resolved,
	bonusStats: split.bonusStats.toProto(),
});

// The split's first setup is the baseline the request sims, the rest its candidates.
const ownBaselineRequest = (bundle: BulkBundleRun, split: BuildSplit): BulkBuildRequest => ({
	candidates: split.setups.slice(1),
	baselineSetup: split.setups[0],
	baselineGear: split.setups[0].gear,
	override: buildOverride(bundle, split),
});

// With the live gear as sent, the group leads with the reference request: the split with the live bonus stats and the
// rotation the live gear resolves to, or an empty one, which keeps the live gear as its baseline. Every other split sims
// its first setup as its baseline.
const groupRequests = (
	bundle: BulkBundleRun,
	sent: readonly BulkGearSetup[],
	keys: PlanKeys,
	liveGear: Gear | null,
	liveBonusStats: Stats,
): BulkBuildRequest[] => {
	const splits = buildSplits(sent, bundle, keys);
	if (!liveGear) return [...splits.values()].map(split => ownBaselineRequest(bundle, split));
	const liveIdentity = gearIdentity(liveGear);
	const resolved = keys.resolve(bundle.talents.talentsString, liveGear, bundle.rotation.rotation);
	const lead = splits.get(splitKey(keys, liveBonusStats, resolved)) ?? { setups: [], bonusStats: liveBonusStats, resolved: resolved.rotation };
	return [
		{
			candidates: lead.setups.filter(setup => keys.identityOf(setup) !== liveIdentity),
			baselineSetup: lead.setups.find(setup => keys.identityOf(setup) === liveIdentity) ?? null,
			baselineGear: null,
			override: buildOverride(bundle, lead),
		},
		...[...splits.values()].filter(split => split !== lead).map(split => ownBaselineRequest(bundle, split)),
	];
};

interface PlannedGroup {
	bundle: BulkBundleRun;
	isEnchanter: boolean;
	setups: BulkGearSetup[];
	/** Each sent key with the bundle that sends it first. */
	senders: Map<string, number>;
}

// Bundles with the same talents, rotation and Settings picks form a group, in the order of each group's first bundle,
// and share its talents, rotation and settings, so a gear setup only ever sims with its own bundle's. A ticked Current
// bundle is `plans[0]`, so its group is group 0, which carries the reference request; with it unticked no request sims
// the live gear. A twin does not run, and neither does a bundle whose every gear setup an earlier bundle of its group
// already sends: it is the same as the bundle that sends its first setup. A twin of a bundle that does not run is the
// same as the bundle that one is the same as, so every bundle left out names one that runs.
export const bulkBuildPlan = (plans: readonly BulkBundlePlan[], reference: BulkBuildReference, resolve: ResolveBuildRotation): BulkBuildPlan => {
	const keys = planKeysOf(resolve);
	const sameAs = twinsOf(plans, keys, reference.settings);
	const groups = new Map<string, PlannedGroup>();
	for (const { bundle, setups } of plans) {
		const twin = sameAs.get(bundle.index);
		if (twin !== undefined) {
			sameAs.set(bundle.index, sameAs.get(twin) ?? twin);
			continue;
		}
		const group: PlannedGroup = groups.get(bundle.buildKey) ?? {
			bundle,
			isEnchanter: sendsEnchanting(bundle.settings, reference),
			setups: [],
			senders: new Map(),
		};
		groups.set(bundle.buildKey, group);
		const sentKeys = setups.map(setup => sentGearSetupKey(setup, group.isEnchanter));
		const sender = sentKeys.length ? group.senders.get(sentKeys[0]) : undefined;
		if (sender !== undefined && sentKeys.every(key => group.senders.has(key))) {
			sameAs.set(bundle.index, sender);
			continue;
		}
		for (const key of sentKeys) {
			if (!group.senders.has(key)) group.senders.set(key, bundle.index);
		}
		group.setups.push(...setups);
	}
	return {
		sameAs,
		groups: [...groups.values()].map(({ bundle, isEnchanter, setups }, group) => {
			const liveGear = bundle.index === -1 ? (isEnchanter ? reference.gear : reference.gear.withoutEnchanting()) : null;
			return { bundle, group, requests: groupRequests(bundle, gearSetupsSentWith(setups, isEnchanter), keys, liveGear, reference.bonusStats) };
		}),
	};
};

// Encounter outer, group inner; the requests do not depend on the encounter, so every encounter runs the same ones.
export const bulkPlanRuns = ({ groups }: BulkBuildPlan, encounters: readonly BulkEncounterSetup[]): BulkBuildRun[] =>
	encounters.flatMap(encounter => groups.flatMap(group => (group.requests.length ? [{ encounter, ...group }] : [])));

// Every plan its `sameAs` does not hold.
export const bulkRunningPlans = (plans: readonly BulkBundlePlan[], { sameAs }: BulkBuildPlan): BulkBundlePlan[] =>
	plans.filter(({ bundle }) => !sameAs.has(bundle.index));

/** `plans` are the ones that run, as `bulkRunningPlans` gives them; `topResults` is the batch's, 0 for All. */
export const bulkBuildsCount = (plans: readonly BulkBundlePlan[], runs: readonly BulkBuildRun[], encounters: number, topResults: number): BulkBuildsCount => {
	const requests = runs.flatMap(run => run.requests);
	const iterations = requests.map(request => estimateBulkSimIterations(request.candidates.length, topResults || request.candidates.length));
	return {
		bundles: plans.length,
		encounters,
		setups: plans.reduce((sum, plan) => sum + plan.setups.length, 0),
		runs: requests.length,
		builds: requests.reduce((sum, request) => sum + 1 + request.candidates.length, 0),
		totalIterations: iterations.reduce((sum, count) => sum + count, 0),
		largestRequestIterations: iterations.reduce((largest, count) => Math.max(largest, count), 0),
		largestRequestCandidates: requests.reduce((largest, request) => Math.max(largest, request.candidates.length), 0),
	};
};

export interface BulkBuildStep {
	encounter: string;
	encounterSummary: string;
	settings: string;
	talents: string;
	rotation: string;
}

export const bulkBuildStep = ({
	encounter,
	bundle,
}: {
	encounter: BulkEncounterSetup;
	bundle: Pick<BulkBundleRun, 'talents' | 'rotation' | 'settings'>;
}): BulkBuildStep => ({
	encounter: encounter.isCurrent ? i18n.t('bulk_tab.settings.encounters.current') : encounter.name,
	encounterSummary: formatEncounterSummary(encounter.encounter),
	settings: settingsLabel(bundle.settings),
	talents: talentsLabel(bundle.talents),
	rotation: rotationLabel(bundle.rotation),
});
