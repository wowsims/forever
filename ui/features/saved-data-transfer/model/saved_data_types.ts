import { EquipmentSpec, Spec } from '@generated/proto/common';
import type { SavedEncounter, SavedGearSet, SavedTalents } from '@generated/proto/ui';
import { PlayerSpecs } from '@sim/player/specs';
import type { IndividualSimHost } from '@sim/sim_host';
import { classTalentsConfig } from '@sim/talents/factory';
import { talentsStringFits } from '@sim/talents/talents_string';
import { getEnumValues } from '@sim/utils/collections';

import type { SavedDataType } from './transfer_text';

interface SavedDataTypeRules {
	specLocked: boolean;
	// Throws on an entry the panel, the sim or the Builds tab would throw on later.
	check?: (data: any, host: IndividualSimHost<any>) => void;
}

export const SAVED_DATA_TYPE_RULES: Record<SavedDataType, SavedDataTypeRules> = {
	gear: {
		specLocked: false,
		// More items than a slot holds throws here, and the Builds tab looks up every saved set while it renders.
		check: (data: SavedGearSet, { sim }) => sim.db.lookupEquipmentSpec(data.gear ?? EquipmentSpec.create()),
	},
	talents: {
		specLocked: true,
		// The sim panics on a fourth tree or a tree longer than the class has, and the string goes into the Batch CSV as it
		// stands, so only one that fits the class gets in.
		check: (data: SavedTalents, { player }) => {
			if (!talentsStringFits(classTalentsConfig[player.getClass()], data.talentsString)) throw new Error('Not a talents string of this class');
		},
	},
	rotation: { specLocked: true },
	// Locked because the entries hold spec options.
	settings: { specLocked: true },
	encounter: {
		specLocked: false,
		// The encounter picker reads the first target, so it throws on an encounter with none.
		check: (data: SavedEncounter) => {
			if (!data.encounter?.targets.length) throw new Error('No target in the entry');
		},
	},
};

// The export names its spec in text from outside, so only an own member of the enum's name side counts.
export const specFromName = (name: string): Spec | undefined => getEnumValues<Spec>(Spec).find(spec => spec !== Spec.SpecUnknown && Spec[spec] === name);

// A class with one spec has an empty spec name.
export const specFullName = (spec: Spec): string => PlayerSpecs.getFullSpecName(PlayerSpecs.fromProto(spec)).trim();
