import { savedEncounterCodec } from '@features/encounter/hooks/useSavedEncounter';
import { Encounter } from '@generated/proto/common';
import { BulkEncounterRef, BulkEncounterSelection, SavedEncounter } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkEncounterResults } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { PresetEncounter } from '@sim/presets/types';
import { bulkState, patchBulkState, storeBulkEncounters } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { formatDurationSeconds } from '@sim/utils/format';
import { savedDataEntries } from '@ui-kit/hooks/useSavedData';

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

type EncounterSummaryFields = Pick<Encounter, 'duration' | 'durationVariation' | 'targets'>;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.settings.encounters.${key}`, options);

// The sim stamps its api version on every encounter it hands out, while presets and older saves carry another one.
const sameEncounter = (a: Encounter, b: Encounter) => Encounter.equals({ ...a, apiVersion: 0 }, { ...b, apiVersion: 0 });

const sameRef = (a: BulkEncounterRef, b: BulkEncounterRef) => a.name === b.name && a.preset === b.preset;

const isBulkEncounterChecked = (checked: readonly BulkEncounterRef[], ref: BulkEncounterRef) => checked.some(item => sameRef(item, ref));

// As the results store it.
export const encounterLength = ({ duration, durationVariation }: Pick<Encounter, 'duration' | 'durationVariation'>): EncounterLength => ({
	duration,
	durationVariation,
});

// "m:ss ±Ns", or "m:ss" alone when the variation is null.
export const formatEncounterDuration = ({ duration, durationVariation }: EncounterLength): string => {
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

// The record the Saved Encounters panel writes, read as that panel reads it.
const storedSavedEncounters = (host: IndividualSimHost<any>) => {
	let record: unknown;
	try {
		record = JSON.parse(host.sim.env.storage.getItem(host.getSavedEncounterStorageKey()) ?? '{}');
	} catch {
		return [];
	}
	return savedDataEntries(record, savedEncounterCodec);
};

export const pickableBulkEncounters = (host: IndividualSimHost<any>): BulkEncounterEntry[] =>
	bulkEncounterEntries(host.individualConfig.presets.encounters, storedSavedEncounters(host));

interface BulkEncounterListItem extends BulkEncounterEntry {
	checked: boolean;
	duplicate: boolean;
}

// The rows the Encounters group shows, in list order. One equal to the current encounter is left out, since the current
// one runs anyway; a checked one equal to an earlier checked row is flagged. A checked ref whose encounter no longer
// exists has no row.
export const bulkEncounterList = (
	current: Encounter,
	entries: readonly BulkEncounterEntry[],
	checked: readonly BulkEncounterRef[],
): BulkEncounterListItem[] => {
	const picked: Encounter[] = [];
	return entries
		.filter(entry => !sameEncounter(entry.encounter, current))
		.map(entry => {
			const isChecked = isBulkEncounterChecked(checked, entry.ref);
			const duplicate = isChecked && picked.some(earlier => sameEncounter(earlier, entry.encounter));
			if (isChecked) picked.push(entry.encounter);
			return { ...entry, checked: isChecked, duplicate };
		});
};

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

// Rewritten from every pickable ref, rows hidden as equal to the current encounter included, so the stored order
// follows the list, a hidden row keeps its state, and refs to deleted encounters go.
// No `settings` bump: the encounters change neither the combination count nor the stored BulkSettings.
export const setBulkEncounterChecked = (player: Player<any>, ref: BulkEncounterRef, checked: boolean, pickable: readonly BulkEncounterRef[]) => {
	const current = bulkState(player).encounters;
	const encounters = pickable.filter(item => (sameRef(item, ref) ? checked : isBulkEncounterChecked(current, item)));
	patchBulkState(player, { encounters });
	storeBulkEncounters(player, BulkEncounterSelection.create({ encounters }));
};
