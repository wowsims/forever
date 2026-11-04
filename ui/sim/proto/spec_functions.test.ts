import { Spec } from '@generated/proto/common';
import { Hunter_Rotation_Playstyle } from '@generated/proto/hunter';
import { describe, expect, it } from 'vitest';

import { simpleSpecRotationFromJson, specTypeFunctions } from './spec_functions';

describe('simpleSpecRotationFromJson', () => {
	it('keeps a saved hunter rotation that still holds the reserved time_to_weave', () => {
		const rotation = simpleSpecRotationFromJson(specTypeFunctions[Spec.SpecHunter], '{"timeToWeave":250,"playstyle":"Weave","useMulti":true}');
		expect(rotation).toMatchObject({ playstyle: Hunter_Rotation_Playstyle.Weave, useMulti: true });
	});
});
