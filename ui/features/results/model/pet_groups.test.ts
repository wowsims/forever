import type { ActionMetrics, UnitMetrics } from '@sim/proto/sim_result';
import { actionMetric, mockActionMetricsMerge } from '@testing/results-metrics';
import { beforeEach, describe, expect, it } from 'vitest';

import { groupActionsWithPets } from './pet_groups';

const action = (name: string) => actionMetric(name, { actionId: { baseName: name, toStringIgnoringTag: () => name, toString: () => name } });

interface Unit {
	name: string;
	actions: Array<ActionMetrics>;
	pets: Array<Unit>;
}

const unit = (name: string, actions: Array<string>, pets: Array<Unit> = []): Unit => ({ name, actions: actions.map(action), pets });

const names = (player: Unit) =>
	groupActionsWithPets(player as unknown as UnitMetrics, source => (source as unknown as Unit).actions).map(group => group.map(metric => metric.name));

describe('groupActionsWithPets', () => {
	beforeEach(() => {
		mockActionMetricsMerge();
	});

	it('groups a player with no pets by action', () => {
		expect(names(unit('Warrior', ['Shield Slam', 'Revenge']))).toEqual([['Shield Slam'], ['Revenge']]);
	});

	it('keeps a pet no spell summons as a group of its own', () => {
		expect(names(unit('Hunter', ['Steady Shot'], [unit('Wolf', ['Claw', 'Bite'])]))).toEqual([['Steady Shot'], ['Claw', 'Bite']]);
	});

	it('leaves out a pet with no actions, whatever comes after it', () => {
		const pets = [unit('Imp', []), unit('Hawk 1', []), unit('Hawk 2', []), unit('Wolf', ['Claw'])];
		expect(names(unit('Hunter', ['Steady Shot'], pets))).toEqual([['Steady Shot'], ['Claw']]);
	});

	it('puts a summoned pet under the spell that summons it, as one row', () => {
		const pets = [unit('Hawk 1', ['Attack']), unit('Hawk 2', []), unit('Wolf', ['Claw'])];
		expect(names(unit('Hunter', ['Steady Shot', 'Summon Hawk'], pets))).toEqual([['Steady Shot'], ['Summon Hawk', 'Attack'], ['Claw']]);
	});

	it('keeps a summoned pet as its own group when its owner shows no row for the spell', () => {
		expect(names(unit('Hunter', ['Steady Shot'], [unit('Hawk 1', ['Attack'])]))).toEqual([['Steady Shot'], ['Attack']]);
	});
});
