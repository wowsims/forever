import { HandType, WeaponSkill } from '@generated/proto/common';
import type { UIItem as Item } from '@generated/proto/ui';

import { ONE_HAND_WEAPON_SKILLS, TWO_HAND_WEAPON_SKILLS } from './classes/capabilities_auto_gen';

export enum WeaponSkillGroup {
	Melee,
	Ranged,
	Thrown,
	Wands,
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
};

export const weaponSkillsInGroups = (groups: ReadonlyArray<WeaponSkillGroup>): WeaponSkill[] => groups.flatMap(group => WEAPON_SKILL_GROUPS[group]);

/** The weapon skill a melee weapon uses, by its type and hand type; undefined for a shield, a held-in-off-hand item or no item. */
export const meleeWeaponSkillOf = (item: Pick<Item, 'weaponType' | 'handType'> | undefined): WeaponSkill | undefined =>
	item && (item.handType === HandType.HandTypeTwoHand ? TWO_HAND_WEAPON_SKILLS : ONE_HAND_WEAPON_SKILLS)[item.weaponType];
