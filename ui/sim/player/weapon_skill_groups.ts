import { WeaponSkill } from '@generated/proto/common';

export enum WeaponSkillGroup {
	Melee,
	Ranged,
	Thrown,
	Wands,
	FeralCombat,
}

const WEAPON_SKILL_GROUPS: Record<WeaponSkillGroup, WeaponSkill[]> = {
	[WeaponSkillGroup.Melee]: [
		WeaponSkill.WeaponSkillAxes,
		WeaponSkill.WeaponSkillTwoHandedAxes,
		WeaponSkill.WeaponSkillMaces,
		WeaponSkill.WeaponSkillTwoHandedMaces,
		WeaponSkill.WeaponSkillSwords,
		WeaponSkill.WeaponSkillTwoHandedSwords,
		WeaponSkill.WeaponSkillPolearms,
		WeaponSkill.WeaponSkillStaves,
		WeaponSkill.WeaponSkillDaggers,
		WeaponSkill.WeaponSkillFistWeapons,
	],
	[WeaponSkillGroup.Ranged]: [WeaponSkill.WeaponSkillBows, WeaponSkill.WeaponSkillCrossbows, WeaponSkill.WeaponSkillGuns],
	[WeaponSkillGroup.Thrown]: [WeaponSkill.WeaponSkillThrown],
	[WeaponSkillGroup.Wands]: [WeaponSkill.WeaponSkillWands],
	[WeaponSkillGroup.FeralCombat]: [WeaponSkill.WeaponSkillFeralCombat],
};

export const weaponSkillsInGroups = (groups: ReadonlyArray<WeaponSkillGroup>): WeaponSkill[] => groups.flatMap(group => WEAPON_SKILL_GROUPS[group]);
