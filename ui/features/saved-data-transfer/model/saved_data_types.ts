import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { EquipmentSpec, Spec } from '@generated/proto/common';
import type { SavedEncounter, SavedGearSet, SavedTalents } from '@generated/proto/ui';
import { PlayerSpecs } from '@sim/player/specs';
import type { IndividualSimHost } from '@sim/sim_host';

import type { SavedDataType } from './transfer_text';

const names = (presets: ReadonlyArray<{ name: string }> | undefined) => (presets ?? []).map(({ name }) => name);

interface SavedDataTypeRules {
	specLocked: boolean;
	presetNames: (presets: SpecPresets) => string[];
	// Throws on an entry the panel or the Builds tab would throw on later.
	check?: (data: any, host: IndividualSimHost<any>) => void;
}

export const SAVED_DATA_TYPE_RULES: Record<SavedDataType, SavedDataTypeRules> = {
	gear: {
		specLocked: false,
		presetNames: presets => names(presets.gear),
		// More items than a slot holds throws here, and the Builds tab looks up every saved set while it renders.
		check: (data: SavedGearSet, { sim }) => sim.db.lookupEquipmentSpec(data.gear ?? EquipmentSpec.create()),
	},
	talents: {
		specLocked: true,
		presetNames: presets => names(presets.talents),
		// The string goes into the Batch CSV as it stands, so only the digits and dashes of a real one get in.
		check: (data: SavedTalents) => {
			if (!/^[\d-]*$/.test(data.talentsString)) throw new Error('Not a talents string');
		},
	},
	rotation: { specLocked: true, presetNames: presets => names(presets.rotations) },
	// Locked because the entries hold spec options.
	settings: { specLocked: true, presetNames: presets => [...names(presets.settings), ...names(presets.itemSwaps)] },
	encounter: {
		specLocked: false,
		presetNames: presets => names(presets.encounters),
		check: (data: SavedEncounter) => {
			if (!data.encounter) throw new Error('No encounter in the entry');
		},
	},
};

// The export names its spec in text from outside, so only an own member of the enum's name side counts.
export const specFromName = (name: string): Spec | undefined => {
	const spec = Object.entries(Spec).find(([key]) => key === name)?.[1];
	return typeof spec === 'number' && spec !== Spec.SpecUnknown ? spec : undefined;
};

// A class with one spec has an empty spec name.
export const specFullName = (spec: Spec): string => PlayerSpecs.getFullSpecName(PlayerSpecs.fromProto(spec)).trim();
