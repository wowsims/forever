import { VersusEffect, VersusEffectKind } from '@generated/proto/api';
import { AreaType, CreatureFamily, MobType, Target } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import { familyState, itemVersusChips, mobTypeState, namesVersus, versusEffectFactor } from './versus_state';

// VersusChips.test.tsx covers the on, partial and off states through the rendered chips; these cases are the ones a chip cannot show.
const target = (mobType: MobType, creatureFamily = CreatureFamily.CreatureFamilyUnknown) => Target.create({ mobType, creatureFamily });

describe('mobTypeState', () => {
	it('never matches a target of unknown mob type', () => {
		expect(mobTypeState(MobType.MobTypeUnknown, [target(MobType.MobTypeUnknown)])).toEqual({ state: 'off', matched: 0, total: 1 });
	});

	it("counts a target of a family and no mob type of its own as the family's mob type", () => {
		const targets = [
			target(MobType.MobTypeUnknown, CreatureFamily.CreatureFamilyFireElemental),
			target(MobType.MobTypeUnknown, CreatureFamily.CreatureFamilyAquatic),
		];

		expect(mobTypeState(MobType.MobTypeElemental, targets)).toEqual({ state: 'partial', matched: 1, total: 2 });
	});

	it("counts a target of its own mob type as that type alone, whatever its family's", () => {
		expect(mobTypeState(MobType.MobTypeHumanoid, [target(MobType.MobTypeUndead, CreatureFamily.CreatureFamilyTroll)])).toEqual({
			state: 'off',
			matched: 0,
			total: 1,
		});
	});
});

describe('familyState', () => {
	it('never matches a target with no family', () => {
		expect(familyState(CreatureFamily.CreatureFamilyUnknown, [target(MobType.MobTypeBeast)])).toEqual({ state: 'off', matched: 0, total: 1 });
	});
});

describe('namesVersus', () => {
	it('names versus for an item that names only an area', () => {
		const item = UIItem.create({ areaTypes: [AreaType.AreaTypeSnowy] });

		expect(namesVersus(item)).toBe(true);
		expect(itemVersusChips(item, [], [AreaType.AreaTypeSnowy])).toEqual([
			{ kind: 'area', value: AreaType.AreaTypeSnowy, state: 'on', matched: 1, total: 1 },
		]);
	});
});

describe('versusEffectFactor', () => {
	const read = (kind: VersusEffectKind, multiplier: number) => versusEffectFactor(VersusEffect.create({ kind, multiplier }));

	it('gives a key for every kind of effect the sim reports', () => {
		const keys = Object.values(VersusEffectKind)
			.filter((kind): kind is VersusEffectKind => typeof kind === 'number')
			.map(kind => read(kind, 2));

		expect(new Set(keys).size).toBe(keys.length);
		for (const key of keys) expect(key).toMatch(/^sidebar\.character_stats\.versus\.effect_/);
	});
});
