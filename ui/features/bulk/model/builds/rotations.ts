import { APLRotation } from '@generated/proto/apl';
import { BulkRotationRef, SavedRotation } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { BulkRotationSetup } from '@sim/bulk/types';
import type { PresetRotation } from '@sim/presets/types';
import type { IndividualSimHost } from '@sim/sim_host';
import { omitDeep } from '@sim/utils/collections';

import { storedSavedData } from './saved_refs';

export interface BulkRotationEntry {
	ref: BulkRotationRef;
	name: string;
	rotation: APLRotation;
}

const entry = (name: string, preset: boolean, rotation: APLRotation): BulkRotationEntry => ({
	ref: BulkRotationRef.create({ name, preset }),
	name,
	rotation,
});

// Every spec preset first, then the user's saved rotations in the order the Saved Rotations panel lists them.
export const bulkRotationEntries = (
	presets: readonly PresetRotation[] | undefined,
	saved: ReadonlyArray<{ name: string; data: SavedRotation }>,
): BulkRotationEntry[] => [
	...(presets ?? []).map(preset => entry(preset.name, true, preset.rotation.rotation ?? APLRotation.create())),
	...saved.map(({ name, data }) => entry(name, false, data.rotation ?? APLRotation.create())),
];

// Read uuid-stripped, as the Saved Rotations panel reads them.
export const pickableBulkRotations = (host: IndividualSimHost<any>): BulkRotationEntry[] =>
	bulkRotationEntries(
		host.individualConfig.presets.rotations,
		storedSavedData(host, host.getSavedRotationStorageKey(), SavedRotation).map(({ name, data }) => ({ name, data: omitDeep(data, ['uuid']) })),
	);

// The current rotation runs unnamed; every view calls it "Current rotation".
export const rotationLabel = ({ isCurrent, name }: Pick<BulkRotationSetup, 'isCurrent' | 'name'>): string =>
	isCurrent ? i18n.t('bulk_tab.results.current_rotation') : name;

// The rotation as picked: two Simple picks that differ only in their settings stay apart.
export const rotationKey = ({ isCurrent, rotation }: Pick<BulkRotationSetup, 'isCurrent' | 'rotation'>): string =>
	isCurrent ? '' : APLRotation.toJsonString(omitDeep(rotation, ['uuid']));
