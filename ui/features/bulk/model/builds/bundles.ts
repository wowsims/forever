import type { APLRotation } from '@generated/proto/apl';
import { type BulkBuildBundle, BulkBuildBundles, type SavedSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkPlayerSettingsSetup, BulkRotationSetup, BulkTalentSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { withLiveSpecOptions } from '@sim/proto/spec_functions';
import type { Stats } from '@sim/proto/stats';
import { bulkState, patchBulkState, storeBulkBundles } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';

import { type BulkGearSetEntry, bulkGearSetName, type BulkGearSetPick, pickableBulkGearSets } from './gear_sets';
import { bonusStatsKey } from './gear_setups';
import { type BulkRotationEntry, pickableBulkRotations, rotationKey, rotationLabel } from './rotations';
import { sameRef, type SavedRef } from './saved_refs';
import { type BulkSettingsEntry, pickableBulkSettings, settingsContentKey, settingsLabel } from './settings_entries';
import { type BulkTalentEntry, pickableBulkTalents, talentsLabel } from './talents';

export interface BulkPart {
	isCurrent: boolean;
	name: string;
}

export const CURRENT_REF: SavedRef = { name: '', preset: false };

export interface BulkBundleParts {
	gearSets: readonly BulkGearSetEntry[];
	talents: readonly BulkTalentEntry[];
	rotations: readonly BulkRotationEntry[];
	settings: readonly BulkSettingsEntry[];
}

export const pickableBulkBundleParts = (host: IndividualSimHost<any>): BulkBundleParts => ({
	gearSets: pickableBulkGearSets(host),
	talents: pickableBulkTalents(host),
	rotations: pickableBulkRotations(host),
	settings: pickableBulkSettings(host),
});

export interface CurrentBundle {
	gear: Gear;
	bonusStats: Stats;
	talentsString: string;
	rotation: APLRotation;
	/** Null until the sim is ready; a current settings part then matches only another current one. */
	settings: SavedSettings | null;
}

export interface BulkBundleRun {
	/** The stored index, -1 for the Current bundle. */
	index: number;
	/** Named '' for the current gear. */
	gearSet: BulkGearSetPick;
	talents: BulkTalentSetup;
	rotation: BulkRotationSetup;
	settings: BulkPlayerSettingsSetup;
	/** Equal for bundles with the same talents, rotation and Settings picks, by ref. */
	buildKey: string;
}

interface BulkBundleRow {
	/** In the stored list. */
	index: number;
	bundle: BulkBuildBundle;
	label: string;
	checked: boolean;
	duplicateOf: string | null;
}

// A bundle's four entries, null for a current part.
interface ResolvedBundle {
	gearSet: BulkGearSetEntry | null;
	talents: BulkTalentEntry | null;
	rotation: BulkRotationEntry | null;
	settings: BulkSettingsEntry | null;
}

const CURRENT_PARTS: ResolvedBundle = { gearSet: null, talents: null, rotation: null, settings: null };

const CURRENT_PART: BulkPart = { isCurrent: true, name: '' };

// Null for a current part; undefined when no entry holds the ref.
const partEntry = <E extends { ref: SavedRef }>(entries: readonly E[], ref: SavedRef | undefined): E | null | undefined =>
	!ref || ref.name === '' ? null : entries.find(entry => sameRef(entry.ref, ref));

// Undefined when a saved part is missing.
const resolveBundle = (parts: BulkBundleParts, bundle: BulkBuildBundle): ResolvedBundle | undefined => {
	const gearSet = partEntry(parts.gearSets, bundle.gearSet);
	const talents = partEntry(parts.talents, bundle.talents);
	const rotation = partEntry(parts.rotations, bundle.rotation);
	const settings = partEntry(parts.settings, bundle.settings);
	if (gearSet === undefined || talents === undefined || rotation === undefined || settings === undefined) return undefined;
	return { gearSet, talents, rotation, settings };
};

const gearSetLabel = ({ isCurrent, name }: BulkPart): string => (isCurrent ? i18n.t('bulk_tab.results.current_gear') : name);

const bundleLabel = (parts: Record<'gearSet' | 'talents' | 'rotation' | 'settings', BulkPart>): string =>
	[gearSetLabel(parts.gearSet), talentsLabel(parts.talents), rotationLabel(parts.rotation), settingsLabel(parts.settings)].join(' · ');

const part = (entry: { name: string } | null): BulkPart => (entry === null ? CURRENT_PART : { isCurrent: false, name: entry.name });

const resolvedLabel = ({ gearSet, talents, rotation, settings }: ResolvedBundle): string =>
	bundleLabel({
		gearSet: gearSet === null ? CURRENT_PART : { isCurrent: false, name: bulkGearSetName(gearSet) },
		talents: part(talents),
		rotation: part(rotation),
		settings: part(settings),
	});

const options = <E extends { ref: SavedRef; name: string }>(
	entries: readonly E[],
	currentLabel: string,
	name: (entry: E) => string = entry => entry.name,
): Array<{ ref: SavedRef; label: string }> => [{ ref: CURRENT_REF, label: currentLabel }, ...entries.map(entry => ({ ref: entry.ref, label: name(entry) }))];

export const bundleOptions = (parts: BulkBundleParts): Record<'gearSets' | 'talents' | 'rotations' | 'settings', Array<{ ref: SavedRef; label: string }>> => ({
	gearSets: options(parts.gearSets, gearSetLabel(CURRENT_PART), bulkGearSetName),
	talents: options(parts.talents, talentsLabel(CURRENT_PART)),
	rotations: options(parts.rotations, rotationLabel(CURRENT_PART)),
	settings: options(parts.settings, settingsLabel(CURRENT_PART)),
});

// Equal for two bundles that sim the same, whatever their parts are named. The rotation is compared as picked, so the
// current rotation and a saved one holding the same list stay apart.
const contentKey = (current: CurrentBundle, { gearSet, talents, rotation, settings }: ResolvedBundle): string => {
	const gear = gearSet ?? current;
	const sentSettings = settings ? withLiveSpecOptions(settings.settings, current.settings) : current.settings;
	return JSON.stringify([
		getGearIdentityKey(gear.gear.asSpec()),
		bonusStatsKey(gear.bonusStats),
		talents?.talentsString ?? current.talentsString,
		rotationKey({ isCurrent: rotation === null, rotation: rotation?.rotation ?? current.rotation }),
		sentSettings ? settingsContentKey(sentSettings) : '',
	]);
};

// The stored bundles whose parts all exist, in stored order, each with the label of the bundle it duplicates: the
// Current bundle while it is ticked, or an earlier checked one with the same content. A duplicate does not run, and its
// stored checked state is left as it is, so it is back once the other bundle changes.
const flaggedBundles = (current: CurrentBundle, parts: BulkBundleParts, bundles: readonly BulkBuildBundle[], currentChecked: boolean) => {
	const picked = new Map<string, string>(currentChecked ? [[contentKey(current, CURRENT_PARTS), i18n.t('bulk_tab.builds.bundles.current')]] : []);
	return bundles.flatMap((bundle, index) => {
		const resolved = resolveBundle(parts, bundle);
		if (!resolved) return [];
		const content = contentKey(current, resolved);
		const label = resolvedLabel(resolved);
		const duplicateOf = picked.get(content) ?? null;
		if (duplicateOf === null && bundle.checked) picked.set(content, label);
		return [{ index, bundle, resolved, label, duplicateOf }];
	});
};

export const bulkBundleList = (current: CurrentBundle, parts: BulkBundleParts, bundles: readonly BulkBuildBundle[], currentChecked: boolean): BulkBundleRow[] =>
	flaggedBundles(current, parts, bundles, currentChecked).map(({ index, bundle, label, duplicateOf }) => ({
		index,
		bundle,
		label,
		checked: bundle.checked,
		duplicateOf,
	}));

const refKey = (entry: { ref: SavedRef } | null) => (entry === null ? [CURRENT_REF.name, CURRENT_REF.preset] : [entry.ref.name, entry.ref.preset]);

const bundleRun = (current: CurrentBundle, index: number, { gearSet, talents, rotation, settings }: ResolvedBundle): BulkBundleRun => ({
	index,
	gearSet:
		gearSet === null
			? { name: '', gear: current.gear, bonusStats: current.bonusStats }
			: { name: bulkGearSetName(gearSet), gear: gearSet.gear, bonusStats: gearSet.bonusStats },
	talents:
		talents === null
			? { name: '', talentsString: current.talentsString, isCurrent: true }
			: { name: talents.name, talentsString: talents.talentsString, isCurrent: false },
	rotation:
		rotation === null ? { name: '', isCurrent: true, rotation: current.rotation } : { name: rotation.name, isCurrent: false, rotation: rotation.rotation },
	settings: settings === null ? { name: '', isCurrent: true, settings: null } : { name: settings.name, isCurrent: false, settings: settings.settings },
	buildKey: JSON.stringify([talents, rotation, settings].map(refKey)),
});

// The Current bundle runs first while it is ticked, then every checked bundle the list shows that duplicates none, in
// list order.
export const bulkBundleRuns = (
	current: CurrentBundle,
	parts: BulkBundleParts,
	bundles: readonly BulkBuildBundle[],
	currentChecked: boolean,
): BulkBundleRun[] => [
	...(currentChecked ? [bundleRun(current, -1, CURRENT_PARTS)] : []),
	...flaggedBundles(current, parts, bundles, currentChecked)
		.filter(({ bundle, duplicateOf }) => bundle.checked && duplicateOf === null)
		.map(({ index, resolved }) => bundleRun(current, index, resolved)),
];

// No `settings` bump: the bundles change neither the Gear mode's combination count nor the stored BulkSettings.
const setBundles = (player: Player<any>, bundles: BulkBuildBundle[], currentBundleChecked = bulkState(player).currentBundleChecked) => {
	patchBulkState(player, { bundles, currentBundleChecked });
	storeBulkBundles(player, BulkBuildBundles.create({ bundles, currentUnchecked: !currentBundleChecked }));
};

export const addBulkBundle = (player: Player<any>, bundle: BulkBuildBundle) => setBundles(player, [...bulkState(player).bundles, { ...bundle, checked: true }]);

export const replaceBulkBundle = (player: Player<any>, index: number, bundle: BulkBuildBundle) =>
	setBundles(
		player,
		bulkState(player).bundles.map((stored, i) => (i === index ? { ...bundle, checked: stored.checked } : stored)),
	);

export const removeBulkBundle = (player: Player<any>, index: number) =>
	setBundles(
		player,
		bulkState(player).bundles.filter((_, i) => i !== index),
	);

export const setBulkBundleChecked = (player: Player<any>, index: number, checked: boolean) =>
	setBundles(
		player,
		bulkState(player).bundles.map((stored, i) => (i === index ? { ...stored, checked } : stored)),
	);

export const setBulkCurrentBundleChecked = (player: Player<any>, checked: boolean) => setBundles(player, [...bulkState(player).bundles], checked);

const savedPartGone = <E extends { ref: SavedRef }>(entries: readonly E[], ref: SavedRef | undefined): boolean =>
	!!ref && !ref.preset && partEntry(entries, ref) === undefined;

// Drops every stored bundle with a saved part that is missing, so an entry of that name coming back does not bring
// the bundle back. A preset part that is not pickable now, such as one its `enableWhen` hides, keeps
// the bundle stored; it has no row and does not run until the preset is pickable again. It writes the store, so it runs
// outside render.
export const pruneBulkBundles = (player: Player<any>, parts: BulkBundleParts): void => {
	const { bundles } = bulkState(player);
	const kept = bundles.filter(
		bundle =>
			!savedPartGone(parts.gearSets, bundle.gearSet) &&
			!savedPartGone(parts.talents, bundle.talents) &&
			!savedPartGone(parts.rotations, bundle.rotation) &&
			!savedPartGone(parts.settings, bundle.settings),
	);
	if (kept.length < bundles.length) setBundles(player, kept);
};
