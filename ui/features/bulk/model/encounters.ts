import { Encounter } from '@generated/proto/common';
import { BulkEncounterRef, BulkEncounterSelection, SavedEncounter } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { PresetEncounter } from '@sim/presets/types';
import { bulkState, patchBulkState, storeBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { formatDurationSeconds } from '@sim/utils/format';

export interface NamedEncounter {
	name: string;
	encounter: Encounter;
}

export interface BulkEncounterEntry extends NamedEncounter {
	ref: BulkEncounterRef;
}

export interface BulkEncounterRun extends NamedEncounter {
	isCurrent: boolean;
}

type EncounterSummaryFields = Pick<Encounter, 'duration' | 'durationVariation' | 'useHealth' | 'targets'>;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.settings.encounters.${key}`, options);

// The sim stamps its api version on every encounter it hands out, while presets and older saves carry another one.
const sameEncounter = (a: Encounter, b: Encounter) => Encounter.equals({ ...a, apiVersion: 0 }, { ...b, apiVersion: 0 });

const sameRef = (a: BulkEncounterRef, b: BulkEncounterRef) => a.name === b.name && a.preset === b.preset;

export const isBulkEncounterChecked = (checked: readonly BulkEncounterRef[], ref: BulkEncounterRef) => checked.some(item => sameRef(item, ref));

export const isDuplicateEncounter = (current: Encounter, picked: readonly Encounter[], index: number): boolean =>
	sameEncounter(current, picked[index]) || picked.slice(0, index).some(earlier => sameEncounter(earlier, picked[index]));

// The current encounter runs first, unnamed; a picked one equal to the current one or to an earlier pick is skipped.
export const bulkEncounterRuns = (current: Encounter, picked: readonly NamedEncounter[]): BulkEncounterRun[] => {
	const encounters = picked.map(entry => entry.encounter);
	return [
		{ name: '', encounter: current, isCurrent: true },
		...picked.filter((_, index) => !isDuplicateEncounter(current, encounters, index)).map(entry => ({ ...entry, isCurrent: false })),
	];
};

export const formatEncounterDuration = ({ duration, durationVariation, useHealth }: EncounterSummaryFields): string =>
	useHealth ? t('health_summary') : `${formatDurationSeconds(duration, { separatorStyle: 'colon', minimumUnit: 'minutes' })} ±${durationVariation}s`;

export const formatEncounterSummary = (encounter: EncounterSummaryFields): string =>
	`${formatEncounterDuration(encounter)} · ${t('targets', { count: Math.max(1, encounter.targets.length) })}`;

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

// The record the Saved Encounters panel writes, name to SavedEncounter JSON; an unreadable entry is left out.
const storedSavedEncounters = (host: IndividualSimHost<any>): Array<{ name: string; data: SavedEncounter }> => {
	let record: unknown;
	try {
		record = JSON.parse(host.sim.env.storage.getItem(host.getSavedEncounterStorageKey()) ?? '{}');
	} catch {
		return [];
	}
	if (record === null || typeof record !== 'object' || Array.isArray(record)) return [];
	return Object.entries(record).flatMap(([name, json]) => {
		try {
			return [{ name, data: SavedEncounter.fromJson(json) }];
		} catch {
			return [];
		}
	});
};

export const pickableBulkEncounters = (host: IndividualSimHost<any>): BulkEncounterEntry[] =>
	bulkEncounterEntries(host.individualConfig.presets.encounters, storedSavedEncounters(host));

// In list order; a checked ref whose encounter no longer exists drops out here.
export const checkedBulkEncounters = (entries: readonly BulkEncounterEntry[], checked: readonly BulkEncounterRef[]): BulkEncounterEntry[] =>
	entries.filter(entry => isBulkEncounterChecked(checked, entry.ref));

// Rewritten from the listed refs, so the stored order follows the list and refs to deleted encounters go.
// No `settings` bump: the encounters change neither the combination count nor the stored BulkSettings.
export const setBulkEncounterChecked = (player: Player<any>, ref: BulkEncounterRef, checked: boolean, listed: readonly BulkEncounterRef[]) => {
	const current = bulkState(player).encounters;
	const encounters = listed.filter(item => (sameRef(item, ref) ? checked : isBulkEncounterChecked(current, item)));
	patchBulkState(player, { encounters });
	storeBulkEncounters(player, BulkEncounterSelection.create({ encounters }));
};
