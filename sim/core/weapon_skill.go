package core

import (
	"github.com/wowsims/forever/sim/core/proto"
)

// GetWeaponSkill returns the unit's bonus weapon skill for the weapon, on top of its level-based skill.
func GetWeaponSkill(unit *Unit, weapon *Item) float64 {
	if unit.PseudoStats.FeralCombatEnabled {
		return unit.PseudoStats.FeralCombatSkill
	}

	if weapon == nil {
		return 0
	}

	return unit.PseudoStats.WeaponSkills[weaponSkillOf(weapon)]
}

func weaponSkillOf(weapon *Item) proto.WeaponSkill {
	if weapon.HandType == proto.HandType_HandTypeTwoHand {
		return twoHandWeaponSkills[weapon.WeaponType]
	}
	if weapon.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
		return rangedWeaponSkills[weapon.RangedWeaponType]
	}
	return oneHandWeaponSkills[weapon.WeaponType]
}

// Sets each listed skill's bonus over the level-based skill; a skill the list leaves out stays at the maximum.
func (character *Character) applyWeaponSkills(weaponSkills []*proto.WeaponSkillValue) {
	for _, weaponSkill := range weaponSkills {
		if weaponSkill.Skill > proto.WeaponSkill_WeaponSkillUnknown && int(weaponSkill.Skill) < len(character.PseudoStats.WeaponSkills) {
			character.PseudoStats.WeaponSkills[weaponSkill.Skill] = float64(weaponSkill.Value - character.Level*5)
		}
	}
}

func (character *Character) updateAttackTableWeapons() {
	mh, oh, ranged := character.GetMHWeapon(), character.GetOHWeapon(), character.GetRangedWeapon()
	for _, table := range character.AttackTables {
		if table.Defender.Type == EnemyUnit {
			table.setWeapons(mh, oh, ranged)
		}
	}
}
