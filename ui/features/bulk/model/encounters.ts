import { Encounter } from '@generated/proto/common';
import { BulkEncounterRef, BulkEncounterSelection, SavedEncounter } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkEncounterResults } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { PresetEncounter } from '@sim/presets/types';
import { bulkState, patchBulkState, storeBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { formatDurationSeconds } from '@sim/utils/format';

import { checkedRefList, checkedRefsAfter, storedSavedData } from './saved_refs';

export interface BulkEncounterEntry {
	ref: BulkEncounterRef;
	name: string;
	encounter: Encounter;
}

interface BulkEncounterRun {
	name: string;
	encounter: Encounter;
	isCurrent: boolean;
}

type EncounterLength = Pick<BulkEncounterResults, 'duration' | 'durationVariation'>;

type EncounterSummaryFields = Pick<Encounter, 'duration' | 'durationVariation' | 'useHealth' | 'targets'>;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.settings.encounters.${key}`, options);

// The sim stamps its api version on every encounter it hands out, while presets and older saves carry another one.
const sameEncounter = (a: Encounter, b: Encounter) => Encounter.equals({ ...a, apiVersion: 0 }, { ...b, apiVersion: 0 });

// As the results store it: null for both on a health-based encounter, which has no set length.
export const encounterLength = ({
	duration,
	durationVariation,
	useHealth,
}: Pick<Encounter, 'duration' | 'durationVariation' | 'useHealth'>): EncounterLength =>
	useHealth ? { duration: null, durationVariation: null } : { duration, durationVariation };

// "m:ss ±Ns", or "m:ss" alone when the variation is null; a health-based encounter shows how it ends.
export const formatEncounterDuration = ({ duration, durationVariation }: EncounterLength): string => {
	if (duration === null) return t('health_summary');
	const length = formatDurationSeconds(duration, { separatorStyle: 'colon', minimumUnit: 'minutes' });
	return durationVariation === null ? length : `${length} ±${durationVariation}s`;
};

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

// The rows the Encounters group shows, in list order. One equal to the current encounter is left out, since the current
// one runs anyway; a checked one equal to an earlier checked row is flagged. A checked ref whose encounter no longer
// exists has no row.
export const bulkEncounterList = (current: Encounter, entries: readonly BulkEncounterEntry[], checked: readonly BulkEncounterRef[]) =>
	checkedRefList(
		entries,
		checked,
		entry => sameEncounter(entry.encounter, current),
		(a, b) => sameEncounter(a.encounter, b.encounter),
	);

// The current encounter runs first, unnamed, then every checked row the list neither leaves out nor flags. Without power
// tools the current one runs alone; the checked ones stay stored for when they are back on.
export const bulkEncounterRuns = (
	current: Encounter,
	entries: readonly BulkEncounterEntry[],
	checked: readonly BulkEncounterRef[],
	powerTools: boolean,
): BulkEncounterRun[] => [
	{ name: '', encounter: current, isCurrent: true },
	...(powerTools
		? bulkEncounterList(current, entries, checked)
				.filter(item => item.checked && !item.duplicate)
				.map(({ name, encounter }) => ({ name, encounter, isCurrent: false }))
		: []),
];

// No `settings` bump: the encounters change neither the combination count nor the stored BulkSettings.
export const setBulkEncounterChecked = (player: Player<any>, ref: BulkEncounterRef, checked: boolean, pickable: readonly BulkEncounterRef[]) => {
	const encounters = checkedRefsAfter(bulkState(player).encounters, ref, checked, pickable);
	patchBulkState(player, { encounters });
	storeBulkEncounters(player, BulkEncounterSelection.create({ encounters }));
};
