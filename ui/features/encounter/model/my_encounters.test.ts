import i18n from '@i18n/config';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import english from '../../../../assets/locales/en/translation.json';
import { uniqueEncounterName } from './my_encounters';

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('uniqueEncounterName', () => {
	it('keeps a free name, and counts from 2 past the names taken', () => {
		expect(uniqueEncounterName('Ragnaros', ['Onyxia'])).toBe('Ragnaros');
		expect(uniqueEncounterName('Ragnaros', ['Ragnaros'])).toBe('Ragnaros (2)');
		expect(uniqueEncounterName('Ragnaros', ['Ragnaros', 'Ragnaros (2)', 'Ragnaros (4)'])).toBe('Ragnaros (3)');
	});
});
