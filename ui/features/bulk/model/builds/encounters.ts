import { Encounter } from '@generated/proto/common';
import { BulkEncounterRef, BulkEncounterSelection, SavedEncounter } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkEncounterResults, BulkEncounterSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { PresetEncounter } from '@sim/presets/types';
import { bulkState, patchBulkState, storeBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { formatDurationSeconds } from '@sim/utils/format';

import { checkedRefsAfter, keyedOnce, sameRef, storedSavedData } from './saved_refs';

export interface BulkEncounterEntry {
	ref: BulkEncounterRef;
	name: string;
	encounter: Encounter;
}

type EncounterLength = Pick<BulkEncounterResults, 'duration' | 'durationVariation'>;

type EncounterSummaryFields = Pick<Encounter, 'duration' | 'durationVariation' | 'useHealth' | 'targets'>;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.settings.encounters.${key}`, options);

// As the results store it: null for both on a health-based encounter, which has no set length.
export const encounterLength = ({ duration, durationVariation, useHealth }: Pick<Encounter, 'duration' | 'durationVariation' | 'useHealth'>): EncounterLength =>
	useHealth ? { duration: null, durationVariation: null } : { duration, durationVariation };

// "m:ss ±Ns", or "m:ss" alone when the variation is null; a health-based encounter shows how it ends.
export const formatEncounterDuration = ({ duration, durationVariation }: EncounterLength): string => {
	if (duration === null) return t('health_summary');
	const length = formatDurationSeconds(duration, { separatorStyle: 'colon', minimumUnit: 'minutes' });
	return durationVariation === null ? length : `${length} ±${durationVariation}s`;
};

// "3:00 ±5s · 1 target". An encounter with no target still runs one.
export const formatEncounterSummary = (encounter: EncounterSummaryFields): string =>
	`${formatEncounterDuration(encounterLength(encounter))} · ${t('targets', { count: Math.max(1, encounter.targets.length) })}`;

// The current encounter runs unnamed; every view calls it "Current".
export const encounterLabel = ({ isCurrent, name }: Pick<BulkEncounterResults, 'isCurrent' | 'name'>): string =>
	isCurrent ? i18n.t('bulk_tab.results.current') : name;

const entry = (name: string, preset: boolean, encounter: Encounter | undefined): BulkEncounterEntry[] =>
	encounter && encounter.targets.length > 0 ? [{ ref: BulkEncounterRef.create({ name, preset }), name, encounter }] : [];

// Presets first, then the user's saved encounters, in the order the Saved Encounters panel lists them. One with no
// encounter or no target cannot run and is left out.
export const bulkEncounterEntries = (
	presets: readonly PresetEncounter[] | undefined,
	saved: ReadonlyArray<{ name: string; data: SavedEncounter }>,
): BulkEncounterEntry[] => [
	...(presets ?? []).flatMap(({ name, encounter }) => entry(name, true, encounter)),
	...saved.flatMap(({ name, data }) => entry(name, false, data.encounter)),
];

export const pickableBulkEncounters = (host: IndividualSimHost<any>): BulkEncounterEntry[] =>
	bulkEncounterEntries(host.individualConfig.presets.encounters, storedSavedData(host, host.getSavedEncounterStorageKey(), SavedEncounter));

// The proto version, the zone the areas came from and the target names change nothing the sim reads.
const encounterContentKey = keyedOnce((encounter: Encounter) =>
	Encounter.toJsonString({ ...encounter, apiVersion: 0, zoneId: 0, targets: encounter.targets.map(target => ({ ...target, name: '' })) }),
);

// The current encounter's ref: checked when listed, like every other encounter's.
export const CURRENT_ENCOUNTER_REF = BulkEncounterRef.create();

export const currentEncounterChecked = (checked: readonly BulkEncounterRef[]): boolean => checked.some(ref => sameRef(ref, CURRENT_ENCOUNTER_REF));

interface BulkEncounterRow extends BulkEncounterEntry {
	checked: boolean;
	duplicateOf: string | null;
}

// Every entry in list order, each checked by its stored ref. One that sims the same as the current encounter while that
// is checked, or as an earlier checked row, names it, cannot run, and keeps its stored checked state, so it is back once
// the other row changes.
export const bulkEncounterList = (current: Encounter, entries: readonly BulkEncounterEntry[], checked: readonly BulkEncounterRef[]): BulkEncounterRow[] => {
	const picked = new Map<string, string>(currentEncounterChecked(checked) ? [[encounterContentKey(current), t('current')]] : []);
	return entries.map(entry => {
		const content = encounterContentKey(entry.encounter);
		const isChecked = checked.some(ref => sameRef(ref, entry.ref));
		const duplicateOf = picked.get(content) ?? null;
		if (duplicateOf === null && isChecked) picked.set(content, entry.name);
		return { ...entry, checked: isChecked, duplicateOf };
	});
};

// The current encounter runs first while it is checked, unnamed, then every checked row the list shows that duplicates
// none, in list order.
export const bulkEncounterRuns = (current: Encounter, entries: readonly BulkEncounterEntry[], checked: readonly BulkEncounterRef[]): BulkEncounterSetup[] => [
	...(currentEncounterChecked(checked) ? [{ name: '', isCurrent: true, encounter: current }] : []),
	...bulkEncounterList(current, entries, checked)
		.filter(row => row.checked && row.duplicateOf === null)
		.map(({ name, encounter }) => ({ name, isCurrent: false, encounter })),
];

// No `settings` bump: the encounters change neither the Gear mode's combination count nor the stored BulkSettings.
export const setBulkEncounterChecked = (player: Player<any>, ref: BulkEncounterRef, checked: boolean, pickable: readonly BulkEncounterRef[]) => {
	const encounters = checkedRefsAfter(bulkState(player).encounters, ref, checked, [CURRENT_ENCOUNTER_REF, ...pickable]);
	patchBulkState(player, { encounters });
	storeBulkEncounters(player, BulkEncounterSelection.create({ encounters }));
};
