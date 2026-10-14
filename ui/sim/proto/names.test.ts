import { describe, expect, it } from 'vitest';

import { spellSchoolNames } from './names';

const WIRE_SCHOOLS: Array<[number, string]> = [
	[1, 'common.spell_schools.physical'],
	[2, 'common.spell_schools.holy'],
	[4, 'common.spell_schools.fire'],
	[8, 'common.spell_schools.nature'],
	[16, 'common.spell_schools.frost'],
	[32, 'common.spell_schools.shadow'],
	[64, 'common.spell_schools.arcane'],
];

describe('spellSchoolNames', () => {
	it.each(WIRE_SCHOOLS)('names the bit %i that sim/core/flags.go emits with %s', (mask, name) => {
		expect(spellSchoolNames.get(mask)).toBe(name);
	});
});
