import { BulkTalentRef, SavedTalents } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkTalentSetup } from '@sim/bulk/types';
import type { PresetTalents } from '@sim/presets/types';
import type { IndividualSimHost } from '@sim/sim_host';

import { storedSavedData } from './saved_refs';

export interface BulkTalentEntry {
	ref: BulkTalentRef;
	name: string;
	talentsString: string;
}

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

// The current talents run unnamed; every view calls them "Current talents".
export const talentsLabel = ({ isCurrent, name }: Pick<BulkTalentSetup, 'isCurrent' | 'name'>): string =>
	isCurrent ? i18n.t('bulk_tab.results.current_talents') : name;
