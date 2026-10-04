import { WeaponSkill } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { WeaponSkillGroup, weaponSkillsInGroups } from './weapon_skill_groups';

const SKILLS = Object.values(WeaponSkill).filter((skill): skill is WeaponSkill => typeof skill === 'number' && skill !== WeaponSkill.WeaponSkillUnknown);
const GROUPS = Object.values(WeaponSkillGroup).filter((group): group is WeaponSkillGroup => typeof group === 'number');

describe('weaponSkillsInGroups', () => {
	it.each(SKILLS.map(skill => [WeaponSkill[skill], skill] as const))('puts %s in exactly one group', (_name, skill) => {
		expect(GROUPS.filter(group => weaponSkillsInGroups([group]).includes(skill))).toHaveLength(1);
	});

	it('names no skill outside the weapon skills', () => {
		expect(weaponSkillsInGroups(GROUPS).every(skill => SKILLS.includes(skill))).toBe(true);
	});

	it('takes the skills of every group named, and none for no groups', () => {
		expect(weaponSkillsInGroups([WeaponSkillGroup.Ranged, WeaponSkillGroup.Thrown])).toEqual([
			WeaponSkill.WeaponSkillBows,
			WeaponSkill.WeaponSkillCrossbows,
			WeaponSkill.WeaponSkillGuns,
			WeaponSkill.WeaponSkillThrown,
		]);
		expect(weaponSkillsInGroups([])).toEqual([]);
	});
});
