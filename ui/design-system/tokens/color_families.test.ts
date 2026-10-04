// @vitest-environment node
import { describe, expect, it } from 'vitest';

import { COLOR_FAMILIES, COLOR_TOKENS, familyOf } from './color_families';

describe('COLOR_FAMILIES', () => {
	it('places every colour token in exactly one family', () => {
		const orphans = COLOR_TOKENS.filter(token => familyOf(token) === undefined).map(token => token.name);
		expect(orphans).toEqual([]);
		const segments = COLOR_FAMILIES.flatMap(family => family.segments);
		expect(new Set(segments).size).toBe(segments.length);
	});

	it('has no family without a token', () => {
		const empty = COLOR_FAMILIES.filter(family => !COLOR_TOKENS.some(token => familyOf(token) === family)).map(family => family.name);
		expect(empty).toEqual([]);
	});
});
