import { AreaType, CreatureFamily, MobType, Target } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import { areaState, familyState, itemVersusChips, mobTypeState, namesVersus } from './versus_state';

const target = (mobType: MobType, creatureFamily = CreatureFamily.CreatureFamilyUnknown) => Target.create({ mobType, creatureFamily });

describe('mobTypeState', () => {
	it('is on when every target is of the mob type', () => {
		const undead = target(MobType.MobTypeUndead);

		expect(mobTypeState(MobType.MobTypeUndead, [undead, undead])).toEqual({ state: 'on', matched: 2, total: 2 });
	});

	it('is partial when some targets are of the mob type', () => {
		expect(mobTypeState(MobType.MobTypeUndead, [target(MobType.MobTypeUndead), target(MobType.MobTypeHumanoid)])).toEqual({
			state: 'partial',
			matched: 1,
			total: 2,
		});
	});

	it('is off when no target is of the mob type', () => {
		expect(mobTypeState(MobType.MobTypeUndead, [target(MobType.MobTypeMechanical)])).toEqual({ state: 'off', matched: 0, total: 1 });
	});

	it('never matches a target of unknown mob type', () => {
		expect(mobTypeState(MobType.MobTypeUnknown, [target(MobType.MobTypeUnknown)])).toEqual({ state: 'off', matched: 0, total: 1 });
	});

	it('is off for an encounter with no targets', () => {
		expect(mobTypeState(MobType.MobTypeUndead, [])).toEqual({ state: 'off', matched: 0, total: 0 });
	});
});

describe('familyState', () => {
	it('counts the targets of the family, whatever their mob type', () => {
		const targets = [target(MobType.MobTypeBeast, CreatureFamily.CreatureFamilyWolf), target(MobType.MobTypeBeast), target(MobType.MobTypeHumanoid)];

		expect(familyState(CreatureFamily.CreatureFamilyWolf, targets)).toEqual({ state: 'partial', matched: 1, total: 3 });
		expect(familyState(CreatureFamily.CreatureFamilyWolf, targets.slice(0, 1))).toEqual({ state: 'on', matched: 1, total: 1 });
		expect(familyState(CreatureFamily.CreatureFamilyWolf, targets.slice(1))).toEqual({ state: 'off', matched: 0, total: 2 });
	});

	it('never matches a target with no family, and is off with no targets', () => {
		expect(familyState(CreatureFamily.CreatureFamilyUnknown, [target(MobType.MobTypeBeast)])).toEqual({ state: 'off', matched: 0, total: 1 });
		expect(familyState(CreatureFamily.CreatureFamilyWolf, [])).toEqual({ state: 'off', matched: 0, total: 0 });
	});
});

describe('areaState', () => {
	it('is on while the encounter is in the area and off otherwise, never partial', () => {
		expect(areaState(AreaType.AreaTypeSnowy, [AreaType.AreaTypeMountainous, AreaType.AreaTypeSnowy])).toEqual({ state: 'on', matched: 1, total: 1 });
		expect(areaState(AreaType.AreaTypeSnowy, [AreaType.AreaTypeMountainous])).toEqual({ state: 'off', matched: 0, total: 1 });
		expect(areaState(AreaType.AreaTypeSnowy, [])).toEqual({ state: 'off', matched: 0, total: 1 });
	});
});

describe('itemVersusChips', () => {
	it('gives one chip per mob type, then creature family, then area type the item names', () => {
		const item = UIItem.create({
			mobTypes: [MobType.MobTypeUndead, MobType.MobTypeDemon],
			creatureFamilies: [CreatureFamily.CreatureFamilyWolf],
			areaTypes: [AreaType.AreaTypeSnowy],
		});
		const encounter = {
			targets: [target(MobType.MobTypeUndead), target(MobType.MobTypeBeast, CreatureFamily.CreatureFamilyWolf)],
			areaTypes: [AreaType.AreaTypeSnowy],
		};

		expect(itemVersusChips(item, encounter)).toEqual([
			{ kind: 'mobType', value: MobType.MobTypeUndead, state: 'partial', matched: 1, total: 2 },
			{ kind: 'mobType', value: MobType.MobTypeDemon, state: 'off', matched: 0, total: 2 },
			{ kind: 'family', value: CreatureFamily.CreatureFamilyWolf, state: 'partial', matched: 1, total: 2 },
			{ kind: 'area', value: AreaType.AreaTypeSnowy, state: 'on', matched: 1, total: 1 },
		]);
	});

	it('gives no chips, and names nothing, for an item whose effects name no kind of target or area', () => {
		const item = UIItem.create();

		expect(namesVersus(item)).toBe(false);
		expect(itemVersusChips(item, { targets: [target(MobType.MobTypeUndead)], areaTypes: [AreaType.AreaTypeSnowy] })).toEqual([]);
		expect(namesVersus(UIItem.create({ areaTypes: [AreaType.AreaTypeSnowy] }))).toBe(true);
	});
});
