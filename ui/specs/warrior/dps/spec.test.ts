import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, Spec } from '@generated/proto/common';
import { DpsWarrior_Rotation, DpsWarriorSpec } from '@generated/proto/warrior';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import spec from './spec';

const generate = (simple: Partial<DpsWarrior_Rotation>) =>
	spec.simpleRotation!({} as Player<Spec.SpecDpsWarrior>, DpsWarrior_Rotation.create(simple), Cooldowns.create());

const variable = (rotation: APLRotation, name: string) => {
	const value = rotation.valueVariables.find(v => v.name === name)?.value?.value;
	return value?.oneofKind === 'const' ? value.const.val : undefined;
};

describe.each([
	['Fury', DpsWarriorSpec.DpsWarriorSpecFury],
	['Arms', DpsWarriorSpec.DpsWarriorSpecArms],
])('%s simple rotation', (_, warriorSpec) => {
	it('casts the trainer ranks by default', () => {
		const rotation = generate({ spec: warriorSpec });
		expect(variable(rotation, 'Manual: Heroic Strike IX')).toBe('false');
		expect(variable(rotation, 'Manual: Battle Shout VII')).toBe('false');
	});

	it('casts the class book ranks when both toggles are on', () => {
		const rotation = generate({ spec: warriorSpec, useHeroicStrikeRank9: true, useBattleShoutRank7: true });
		expect(variable(rotation, 'Manual: Heroic Strike IX')).toBe('true');
		expect(variable(rotation, 'Manual: Battle Shout VII')).toBe('true');
	});
});
