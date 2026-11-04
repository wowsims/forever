import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, Spec } from '@generated/proto/common';
import { Rogue_Rotation } from '@generated/proto/rogue';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import spec from './spec';

const generateRotation = (simple: Partial<Rogue_Rotation>) =>
	spec.simpleRotation!({} as Player<Spec.SpecRogue>, Rogue_Rotation.create(simple), Cooldowns.create());
const generate = (simple: Partial<Rogue_Rotation>) => APLRotation.toJson(generateRotation(simple));

// Every spell id the rotation names.
const spellIds = (json: unknown): Array<number> => {
	if (Array.isArray(json)) return json.flatMap(spellIds);
	if (json === null || typeof json !== 'object') return [];
	const { spellId } = json as { spellId?: unknown };
	if (typeof spellId === 'number') return [spellId];
	return Object.values(json).flatMap(spellIds);
};

const count = (json: unknown, id: number) => spellIds(json).filter(spellId => spellId === id).length;

const backstabTrainer = 11281;
const backstabBook = 25300;
const eviscerateTrainer = 11300;
const eviscerateBook = 31016;

describe('Rogue simple rotation', () => {
	it('casts the trainer ranks by default', () => {
		const rotation = generate({});
		expect(count(rotation, backstabTrainer)).toBeGreaterThan(0);
		expect(count(rotation, eviscerateTrainer)).toBeGreaterThan(0);
		expect(count(rotation, backstabBook) + count(rotation, eviscerateBook)).toBe(0);
	});

	it('casts the class book ranks when both toggles are on', () => {
		const book = generate({ useBackstabRank9: true, useEviscerateRank9: true });
		expect(count(book, backstabBook)).toBeGreaterThan(0);
		expect(count(book, eviscerateBook)).toBeGreaterThan(0);
		expect(count(book, backstabTrainer) + count(book, eviscerateTrainer)).toBe(0);
	});

	it('sets the Start In Stealth variable from the configurable input', () => {
		const rotation = generateRotation({ startInStealth: true });
		expect(rotation.valueVariables.find(variable => variable.name === 'Start In Stealth')?.value?.value).toEqual({
			oneofKind: 'const',
			const: { val: 'true' },
		});
	});

	it.each([
		['useBackstabRank9', backstabTrainer, backstabBook],
		['useEviscerateRank9', eviscerateTrainer, eviscerateBook],
	] as const)('%s moves only its own spell', (field, trainerId, bookId) => {
		const trainer = generate({});
		const book = generate({ [field]: true });
		expect(spellIds(book).map(id => (id === bookId ? trainerId : id))).toEqual(spellIds(trainer));
	});
});
