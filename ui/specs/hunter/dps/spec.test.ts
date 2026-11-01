import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, Spec } from '@generated/proto/common';
import { Hunter_Rotation, Hunter_Rotation_Playstyle as Playstyle } from '@generated/proto/hunter';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import * as Presets from './presets';
import spec from './spec';

const generate = (simple: Partial<Hunter_Rotation>) =>
	APLRotation.toJson(spec.simpleRotation!({} as Player<Spec.SpecHunter>, Hunter_Rotation.create(simple), Cooldowns.create()));

// Every spell id the rotation names.
const spellIds = (json: unknown): Array<number> => {
	if (Array.isArray(json)) return json.flatMap(spellIds);
	if (json === null || typeof json !== 'object') return [];
	const { spellId } = json as { spellId?: unknown };
	if (typeof spellId === 'number') return [spellId];
	return Object.values(json).flatMap(spellIds);
};

const count = (json: unknown, id: number) => spellIds(json).filter(spellId => spellId === id).length;

const serpentStingBook = 25295;
const serpentStingTrainer = 13555;
const hawkBook = 25296;
const hawkTrainer = 14322;

describe.each([
	['Turret', Playstyle.Turret],
	['Weave', Playstyle.Weave],
	['Melee', Playstyle.Melee],
])('%s simple rotation', (_, playstyle) => {
	it('names the spells the priority list names by default, the trainer ranks', () => {
		const rotation = generate({ playstyle });
		expect(spellIds(rotation)).toEqual(spellIds(Presets.PlaystyleAPLs[playstyle]));
		expect(count(rotation, serpentStingTrainer)).toBeGreaterThan(0);
		expect(count(rotation, serpentStingBook) + count(rotation, hawkBook)).toBe(0);
	});

	it('casts the class book ranks when both toggles are on', () => {
		const book = generate({ playstyle, useSerpentStingRank9: true, useAspectOfTheHawkRank7: true });
		expect(count(book, serpentStingBook)).toBeGreaterThan(0);
		expect(count(book, serpentStingTrainer) + count(book, hawkTrainer)).toBe(0);
	});

	it.each([
		['useSerpentStingRank9', serpentStingTrainer, serpentStingBook],
		['useAspectOfTheHawkRank7', hawkTrainer, hawkBook],
	] as const)('%s moves only its own spell', (field, trainerId, bookId) => {
		const trainer = generate({ playstyle });
		const book = generate({ playstyle, [field]: true });
		expect(spellIds(book).map(id => (id === bookId ? trainerId : id))).toEqual(spellIds(trainer));
	});
});
