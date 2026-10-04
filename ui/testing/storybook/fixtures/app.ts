import { SavedTalents } from '@generated/proto/ui';
import type { PresetBuild } from '@sim/presets/types';

import { makeEquipmentSpec } from './protos';

/** A build that sets gear and talents under its own name; `overrides` add or replace categories. */
export const makePresetBuild = (name: string, overrides: Partial<PresetBuild> = {}): PresetBuild => ({
	name,
	gear: { name, gear: makeEquipmentSpec() },
	talents: { name, data: SavedTalents.create({ talentsString: '0505000000000000000000000000000000000000000000000000000000000000' }) },
	...overrides,
});

/** Two phase 1 builds in two groups and one ungrouped phase 2 build. */
export const makePhasedPresetBuilds = (): Array<PresetBuild> => [
	makePresetBuild('Fury P1', { phase: 1, group: 'Fury' }),
	makePresetBuild('Arms P1', { phase: 1, group: 'Arms' }),
	makePresetBuild('Fury P2', { phase: 2 }),
];
