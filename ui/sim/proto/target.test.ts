import { CreatureFamily, MobType, Target } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { targetMobType } from './target';

describe('targetMobType', () => {
	it("gives a target of a family and no mob type of its own the family's mob type", () => {
		expect(targetMobType(Target.create({ creatureFamily: CreatureFamily.CreatureFamilyWolf }))).toBe(MobType.MobTypeBeast);
	});

	it('keeps a target its own mob type whatever its family', () => {
		expect(targetMobType(Target.create({ mobType: MobType.MobTypeUndead, creatureFamily: CreatureFamily.CreatureFamilyTroll }))).toBe(
			MobType.MobTypeUndead,
		);
	});
});
