import { describe, expect, it } from 'vitest';

import { adjustWeaponImbueId, getTalentPoints, getTalentTreePoints } from './utils';

// Talent strings are one dash-separated run of per-row digits per tree, so the number a caller
// wants is the per-tree total. Returning the digits instead type-checks cleanly and quietly breaks
// talent-preset auto-matching.
describe('getTalentTreePoints', () => {
	it('sums each tree, and reads an empty middle tree as zero', () => {
		// A feral druid talent string: the middle tree is empty, spelled as '--'.
		expect(getTalentTreePoints('2500052300030150330125--053500031003001')).toEqual([40, 0, 21]);
	});

	it('sums the digits of each tree rather than listing them', () => {
		expect(getTalentTreePoints('1231321-12313123-0')).toEqual([13, 16, 0]);
	});

	it('pads a two-tree string out to three trees', () => {
		expect(getTalentTreePoints('11-22')).toHaveLength(3);
		expect(getTalentTreePoints('11-22')).toEqual([2, 4, 0]);
	});

	it('totals every tree for getTalentPoints', () => {
		expect(getTalentPoints('2500052300030150330125--053500031003001')).toBe(61);
	});
});

describe('adjustWeaponImbueId', () => {
	const DENSE_SHARPENING_STONE_ID = 12404;
	const DENSE_WEIGHTSTONE_ID = 12643;
	const CONSECRATED_STONE_ID = 28891;

	it('swaps the Dense pair to the family the weapon belongs to', () => {
		expect(adjustWeaponImbueId(DENSE_SHARPENING_STONE_ID, false, true)).toBe(DENSE_WEIGHTSTONE_ID);
		expect(adjustWeaponImbueId(DENSE_WEIGHTSTONE_ID, true, false)).toBe(DENSE_SHARPENING_STONE_ID);
	});

	it('keeps a stone that already matches', () => {
		expect(adjustWeaponImbueId(DENSE_SHARPENING_STONE_ID, true, false)).toBe(DENSE_SHARPENING_STONE_ID);
		expect(adjustWeaponImbueId(DENSE_WEIGHTSTONE_ID, false, true)).toBe(DENSE_WEIGHTSTONE_ID);
	});

	it('clears a stone when neither family fits the slot', () => {
		expect(adjustWeaponImbueId(DENSE_SHARPENING_STONE_ID, false, false)).toBe(0);
		expect(adjustWeaponImbueId(DENSE_WEIGHTSTONE_ID, false, false)).toBe(0);
	});

	it('passes every other imbue through, whatever the weapon is', () => {
		expect(adjustWeaponImbueId(CONSECRATED_STONE_ID, false, false)).toBe(CONSECRATED_STONE_ID);
		expect(adjustWeaponImbueId(0, true, false)).toBe(0);
	});
});
