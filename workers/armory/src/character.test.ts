import { describe, expect, it } from 'vitest';

import { BadRequestError, characterProfilePath, parseCharacterPath, trimCharacter } from './character';
import equipment from './test-fixtures/equipment.json';
import professions from './test-fixtures/professions.json';
import specializations from './test-fixtures/specializations.json';
import summary from './test-fixtures/summary.json';

describe('parseCharacterPath', () => {
	it('reads a two-part name and slugs the realm', () => {
		expect(parseCharacterPath('/character/US/ClassicBetaPvE2/Erlyn-Farthing')).toEqual({
			region: 'us',
			realmSlug: 'classicbetapve2',
			nameParts: ['erlyn', 'farthing'],
		});
	});

	it('takes a space between the name parts as well as a hyphen', () => {
		expect(parseCharacterPath('/character/eu/Normal/Erlyn%20Farthing').nameParts).toEqual(['erlyn', 'farthing']);
	});

	it('takes a single name, as Classic characters have', () => {
		expect(parseCharacterPath('/character/us/Mirage%20Raceway/Kayla').realmSlug).toBe('mirage-raceway');
	});

	it.each([
		['/character/xx/realm/name', 'region'],
		['/character/us/realm', 'path'],
		['/character/us/realm/a-b-c', 'three parts'],
	])('rejects %s (%s)', path => {
		expect(() => parseCharacterPath(path)).toThrow(BadRequestError);
	});
});

describe('characterProfilePath', () => {
	const request = parseCharacterPath('/character/us/ClassicBetaPvE2/Erlyn-Farthing');

	it('joins the name parts with the configured separator', () => {
		expect(characterProfilePath(request, '-')).toBe('/profile/wow/character/classicbetapve2/erlyn-farthing');
		expect(characterProfilePath(request, ' ')).toBe('/profile/wow/character/classicbetapve2/erlyn%20farthing');
		expect(characterProfilePath(request, '')).toBe('/profile/wow/character/classicbetapve2/erlynfarthing');
	});
});

describe('trimCharacter', () => {
	const character = trimCharacter(summary, equipment, specializations, professions);

	it('keeps the identity fields', () => {
		expect(character).toMatchObject({ name: 'Erlyn', realm: 'ClassicBetaPvE2', level: 60, classId: 8, raceName: 'Human' });
	});

	it('keeps each item with its permanent enchant only', () => {
		expect(character.items).toEqual([
			{ slot: 'HEAD', id: 16914, enchant: 2588 },
			{ slot: 'NECK', id: 18814 },
			{ slot: 'MAIN_HAND', id: 19019, enchant: 2504 },
			{ slot: 'TABARD', id: 6948 },
		]);
	});

	it('reads the talents of the active group across all trees', () => {
		expect(character.talents).toEqual([
			{ spellId: 11222, talentId: 135344, rank: 5 },
			{ spellId: 11069, talentId: 2, rank: 3 },
			{ spellId: 0, talentId: 3, rank: 1 },
		]);
	});

	it('names the primary professions', () => {
		expect(character.professions).toEqual(['Engineering', 'Enchanting']);
	});

	it('copes with a character missing professions and talents', () => {
		const bare = trimCharacter(summary, {}, {}, null);
		expect(bare.items).toEqual([]);
		expect(bare.talents).toEqual([]);
		expect(bare.professions).toEqual([]);
	});
});
