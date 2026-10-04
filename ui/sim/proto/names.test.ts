import { describe, expect, it } from 'vitest';

import { spellSchoolNames } from './names';

const WIRE_SCHOOLS: Array<[number, string]> = [
	[1, 'Physical'],
	[2, 'Holy'],
	[4, 'Fire'],
	[8, 'Nature'],
	[16, 'Frost'],
	[32, 'Shadow'],
	[64, 'Arcane'],
];

describe('spellSchoolNames', () => {
	it.each(WIRE_SCHOOLS)('names the bit %i that sim/core/flags.go emits as %s', (mask, name) => {
		expect(spellSchoolNames.get(mask)).toBe(name);
	});
});
