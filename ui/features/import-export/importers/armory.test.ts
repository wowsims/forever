import { Class, Profession, Race } from '@generated/proto/common';
import type { Database } from '@sim/proto/database';
import { mageTalentsConfig } from '@sim/talents/mage';
import { afterEach, describe, expect, it, vi } from 'vitest';

import type { ArmoryCharacter } from './armory';
import { armoryToImport, fetchArmoryCharacter, parseArmoryParam, parseCharacterName } from './armory';

describe('parseCharacterName', () => {
	it.each(['Erlyn Farthing', 'Erlyn-Farthing', '  Erlyn   Farthing '])('reads %j as a two-part name', input => {
		expect(parseCharacterName(input)).toEqual({ name: 'Erlyn-Farthing' });
	});

	it('takes the realm from the form the game writes in logs', () => {
		expect(parseCharacterName('Erlyn-Farthing-ClassicBetaPvE2')).toEqual({ name: 'Erlyn-Farthing', realm: 'ClassicBetaPvE2' });
	});

	it('keeps a single name', () => {
		expect(parseCharacterName('Kayla')).toEqual({ name: 'Kayla' });
	});

	it.each(['', '   ', 'a-b-c-d'])('rejects %j', input => {
		expect(parseCharacterName(input)).toBeNull();
	});
});

describe('parseArmoryParam', () => {
	it('reads region, realm and a two-part name', () => {
		expect(parseArmoryParam('?armory=US/ClassicBetaPvE2/Erlyn-Farthing')).toEqual({ region: 'us', realm: 'ClassicBetaPvE2', name: 'Erlyn-Farthing' });
	});

	it('normalises a name written with a space', () => {
		expect(parseArmoryParam('?i=rgt&armory=eu/Normal/Erlyn%20Farthing')?.name).toBe('Erlyn-Farthing');
	});

	it.each(['', '?armory=us/Erlyn-Farthing', '?armory=xx/Normal/Erlyn-Farthing', '?armory=us//Erlyn-Farthing'])('ignores %j', search => {
		expect(parseArmoryParam(search)).toBeNull();
	});
});

describe('fetchArmoryCharacter', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('asks the Worker for the character by region, realm and name', async () => {
		const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ name: 'Erlyn' }), { status: 200 }));
		vi.stubGlobal('fetch', fetchMock);

		await fetchArmoryCharacter({ region: 'us', realm: 'ClassicBetaPvE2', name: 'Erlyn-Farthing' });
		expect(fetchMock).toHaveBeenCalledWith('http://localhost:8787/character/us/ClassicBetaPvE2/Erlyn-Farthing');
	});

	it('says when the character does not exist', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 404 })));
		await expect(fetchArmoryCharacter({ region: 'us', realm: 'Normal', name: 'Nobody' })).rejects.toThrow('import.armory.errors.not_found');
	});

	it('says when the Worker is not running', async () => {
		vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
		await expect(fetchArmoryCharacter({ region: 'us', realm: 'Normal', name: 'Erlyn' })).rejects.toThrow('import.armory.errors.unreachable');
	});
});

describe('armoryToImport', () => {
	const [arcane, fire] = mageTalentsConfig;
	const [wandSpecialization, arcaneFocus] = arcane.talents;
	const firstFire = fire.talents[0];

	const db = {
		getItemById: (id: number) => (id == 404 ? undefined : { id }),
		getEnchantsByEffectId: (effectId: number) => (effectId == 2588 ? [{ effectId }] : []),
	} as unknown as Database;

	const character: ArmoryCharacter = {
		name: 'Erlyn',
		realm: 'ClassicBetaPvE2',
		level: 60,
		classId: Class.ClassMage,
		raceName: 'Human',
		items: [
			{ slot: 'TRINKET_2', id: 13 },
			{ slot: 'HEAD', id: 1, enchant: 2588 },
			{ slot: 'TABARD', id: 6948 },
			{ slot: 'CHEST', id: 404 },
			{ slot: 'LEGS', id: 8, enchant: 9999 },
			{ slot: 'TRINKET_1', id: 12 },
		],
		talents: [
			{ spellId: wandSpecialization.spellId, talentId: 0, rank: 9 },
			{ spellId: 0, talentId: arcaneFocus.definitionId!, rank: 3 },
			{ spellId: firstFire.spellId, talentId: 0, rank: 1 },
			{ spellId: 424242, talentId: 0, rank: 1 },
		],
		professions: ['Engineering', 'Basket Weaving'],
	};
	const parsed = armoryToImport(db, character);

	it('takes the class by id and the race by name', () => {
		expect(parsed.charClass).toBe(Class.ClassMage);
		expect(parsed.race).toBe(Race.RaceHuman);
	});

	it('equips known items in slot order, without the tabard', () => {
		expect(parsed.equipmentSpec.items.map(item => item.id)).toEqual([1, 8, 12, 13]);
		expect(parsed.equipmentSpec.items[0].enchant).toBe(2588);
		expect(parsed.equipmentSpec.items[1].enchant).toBe(0);
	});

	it('reports items and enchants the database lacks', () => {
		expect(parsed.missingItems).toEqual([404]);
		expect(parsed.missingEnchants).toEqual([9999]);
	});

	it('places talents by spell id or definition id, with the rank clamped to the talent', () => {
		expect(parsed.talentsStr).toBe(`${wandSpecialization.maxPoints}3-1`);
		expect(parsed.talentsForTheseTrees).toBe(true);
		expect(parsed.unmatchedTalents).toEqual([424242]);
	});

	it('keeps the professions the sim knows', () => {
		expect(parsed.professions).toEqual([Profession.Engineering]);
	});

	it('rejects a race the sim does not have', () => {
		expect(() => armoryToImport(db, { ...character, raceName: 'Vulpera' })).toThrow('import.errors.cannot_parse_race');
	});
});
