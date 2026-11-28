// The fight the sim starts with. This module is a leaf on purpose: state/sim_store.ts
// seeds itself from it and cannot import encounter.ts without a cycle.
export const ENCOUNTER_DEFAULTS = {
	duration: 180,
	durationVariation: 5,
	executeProportion20: 0.2,
	executeProportion35: 0.35,
} as const;
