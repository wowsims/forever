import { describe, expect, it } from 'vitest';

import { CLASS_BG, CLASS_BORDER, CLASS_TEXT, FACTION_TEXT, QUALITY_TEXT, RESOURCE_TEXT, SPELL_SCHOOL_BG, SPELL_SCHOOL_TEXT } from './colors';

const allRecords = [CLASS_BG, CLASS_BORDER, CLASS_TEXT, FACTION_TEXT, QUALITY_TEXT, RESOURCE_TEXT, SPELL_SCHOOL_BG, SPELL_SCHOOL_TEXT];

describe('colors', () => {
	it('every value of every record is a non-empty string without interpolation', () => {
		for (const record of allRecords) {
			for (const value of Object.values(record)) {
				expect(typeof value).toBe('string');
				expect((value as string).length).toBeGreaterThan(0);
				expect(value).not.toContain('${');
			}
		}
	});
});
