import type { IndividualSimSettings } from '@generated/proto/ui';

// Whether the encounter is an entry of My encounters belongs to the viewer: an export says nothing of it, and an
// imported encounter is never one, even under the name of one. Changes the envelope it is given and returns it.
export const withoutMyEncounter = (proto: IndividualSimSettings): IndividualSimSettings => {
	if (proto.encounter) proto.encounter.myEncounter = false;
	return proto;
};
