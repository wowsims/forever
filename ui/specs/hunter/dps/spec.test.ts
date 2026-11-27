import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, CreatureFamily, MobType, Spec, Target } from '@generated/proto/common';
import { Hunter_Rotation, Hunter_Rotation_Playstyle as Playstyle } from '@generated/proto/hunter';
import type { Player } from '@sim/player/player';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import * as Presets from './presets';
import spec from './spec';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('the Improved Tracking warning', () => {
	const warningAgainst = async (target: Target) => {
		const { sim, player, host } = await makeIndividualSim({ spec: Spec.SpecHunter });
		player.setTalentsString('--5');
		sim.encounter.setTargets([target]);
		return spec.warnings![0](host).getContent();
	};

	it("counts a target of a family and no mob type of its own as the family's mob type", async () => {
		expect(await warningAgainst(Target.create({ creatureFamily: CreatureFamily.CreatureFamilyWolf }))).toBe('');
	});

	it('warns against a target of no mob type whose family gives none', async () => {
		expect(await warningAgainst(Target.create({ creatureFamily: CreatureFamily.CreatureFamilyAquatic }))).not.toBe('');
	});

	it('warns against a mechanical target', async () => {
		expect(await warningAgainst(Target.create({ mobType: MobType.MobTypeMechanical }))).not.toBe('');
	});
});

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
	['Turret', { playstyle: Playstyle.Turret }, Presets.PlaystyleAPLs[Playstyle.Turret], true],
	['Turret with hawks', { playstyle: Playstyle.Turret, useHawk: true }, Presets.TurretHawkPlaystyleAPL, true],
	['Weave', { playstyle: Playstyle.Weave }, Presets.PlaystyleAPLs[Playstyle.Weave], true],
	['Melee', { playstyle: Playstyle.Melee }, Presets.PlaystyleAPLs[Playstyle.Melee], false],
] as const)('%s simple rotation', (_, simple, priorityList, castsHawk) => {
	it('names the spells the priority list names by default, the trainer ranks', () => {
		const rotation = generate(simple);
		expect(spellIds(rotation)).toEqual(spellIds(priorityList));
		expect(count(rotation, serpentStingTrainer)).toBeGreaterThan(0);
		expect(count(rotation, hawkTrainer)).toBe(castsHawk ? 1 : 0);
		expect(count(rotation, serpentStingBook) + count(rotation, hawkBook)).toBe(0);
	});

	it('casts the class book ranks when both toggles are on', () => {
		const book = generate({ ...simple, useSerpentStingRank9: true, useAspectOfTheHawkRank7: true });
		expect(count(book, serpentStingBook)).toBeGreaterThan(0);
		expect(count(book, hawkBook)).toBe(castsHawk ? 1 : 0);
		expect(count(book, serpentStingTrainer) + count(book, hawkTrainer)).toBe(0);
	});

	it.each([
		['useSerpentStingRank9', serpentStingTrainer, serpentStingBook],
		['useAspectOfTheHawkRank7', hawkTrainer, hawkBook],
	] as const)('%s moves only its own spell', (field, trainerId, bookId) => {
		const trainer = generate(simple);
		const book = generate({ ...simple, [field]: true });
		expect(spellIds(book).map(id => (id === bookId ? trainerId : id))).toEqual(spellIds(trainer));
	});
});
