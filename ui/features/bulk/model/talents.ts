import { BulkTalentRef, BulkTalentSelection, SavedTalents } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkTalentSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { PlayerSpec } from '@sim/player/player_spec';
import type { PresetTalents } from '@sim/presets/types';
import { bulkState, patchBulkState, storeBulkTalents } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { classTalentsConfig } from '@sim/talents/factory';
import { parseTalentsString, serializeTalentsString } from '@sim/talents/talents_string';

import { checkedRefList, checkedRefsAfter, storedSavedData } from './saved_refs';

export interface BulkTalentEntry {
	ref: BulkTalentRef;
	name: string;
	talentsString: string;
}

// Two strings that differ only in trailing zeroes or dashes are the same setup.
export const sameTalents = (playerSpec: PlayerSpec<any>, a: string, b: string): boolean => {
	const config = classTalentsConfig[playerSpec.classID];
	const normalise = (talentsString: string) => serializeTalentsString(parseTalentsString(config, talentsString));
	return normalise(a) === normalise(b);
};

const entry = (name: string, preset: boolean, talentsString: string): BulkTalentEntry => ({
	ref: BulkTalentRef.create({ name, preset }),
	name,
	talentsString,
});

// Presets first, then the user's saved setups, in the order the Saved Talents panel lists them.
export const bulkTalentEntries = (
	presets: ReadonlyArray<Pick<PresetTalents, 'name' | 'data'>> | undefined,
	saved: ReadonlyArray<{ name: string; data: SavedTalents }>,
): BulkTalentEntry[] => [
	...(presets ?? []).map(({ name, data }) => entry(name, true, data.talentsString)),
	...saved.map(({ name, data }) => entry(name, false, data.talentsString)),
];

export const pickableBulkTalents = (host: IndividualSimHost<any>): BulkTalentEntry[] =>
	bulkTalentEntries(host.individualConfig.presets.talents, storedSavedData(host, host.getSavedTalentsStorageKey(), SavedTalents));

// The rows the Talents group shows, in list order. One equal to the current talents is left out, since the current ones
// run anyway; a checked one equal to an earlier checked row is flagged. A checked ref whose setup no longer exists has no
// row.
export const bulkTalentList = (playerSpec: PlayerSpec<any>, current: string, entries: readonly BulkTalentEntry[], checked: readonly BulkTalentRef[]) =>
	checkedRefList(
		entries,
		checked,
		entry => sameTalents(playerSpec, entry.talentsString, current),
		(a, b) => sameTalents(playerSpec, a.talentsString, b.talentsString),
	);

// The current talents run first, unnamed, then every checked row the list neither leaves out nor flags. Without power
// tools the current ones run alone; the checked ones stay stored for when they are back on.
export const bulkTalentRuns = (
	playerSpec: PlayerSpec<any>,
	current: string,
	entries: readonly BulkTalentEntry[],
	checked: readonly BulkTalentRef[],
	powerTools: boolean,
): BulkTalentSetup[] => [
	{ name: '', talentsString: current, isCurrent: true },
	...(powerTools
		? bulkTalentList(playerSpec, current, entries, checked)
				.filter(item => item.checked && !item.duplicate)
				.map(({ name, talentsString }) => ({ name, talentsString, isCurrent: false }))
		: []),
];

// No `settings` bump: the talents change neither the combination count nor the stored BulkSettings.
export const setBulkTalentChecked = (player: Player<any>, ref: BulkTalentRef, checked: boolean, pickable: readonly BulkTalentRef[]) => {
	const talents = checkedRefsAfter(bulkState(player).talents, ref, checked, pickable);
	patchBulkState(player, { talents });
	storeBulkTalents(player, BulkTalentSelection.create({ talents }));
};

// The current talents run unnamed; every view calls them "Current talents".
export const talentsLabel = ({ isCurrent, name }: Pick<BulkTalentSetup, 'isCurrent' | 'name'>): string =>
	isCurrent ? i18n.t('bulk_tab.results.current_talents') : name;

export const formatTalentsSummary = (talentsString: string): string => (talentsString === '' ? i18n.t('bulk_tab.settings.talents.no_points') : talentsString);
